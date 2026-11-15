package main

import (
	"bytes"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"

	"ville/sim"
	"ville/tr"
)

// Tout est dessiné en coordonnées « virtuelles » puis converti en pixels :
// l'écran fait toujours 1080 unités de haut (la largeur dépend de l'écran),
// ce qui garde le jeu net à n'importe quelle résolution.
// Pour le monde, la transformation ajoute la caméra (position + zoom).

type view struct {
	s, ox, oy float64
	iso       bool // vue isométrique (essai) : le sol est projeté en losanges
}

// isoK : la projection isométrique (rotation de 45°, hauteur divisée par 2).
const isoK = 0.70710678

func isoProj(x, y float64) (float64, float64) { return (x - y) * isoK, (x + y) * isoK * 0.5 }
func isoInv(sx, sy float64) (float64, float64) {
	a, b := sx/isoK, 2*sy/isoK // a = x − y, b = x + y
	return (a + b) / 2, (b - a) / 2
}

// P : un point du monde, en pixels.
func (v view) P(x, y float64) (float32, float32) {
	if v.iso {
		x, y = isoProj(x, y)
	}
	return float32(v.ox + x*v.s), float32(v.oy + y*v.s)
}

// upright : ce qui est debout (personnages, arbres, bâtiments…) est dessiné droit,
// posé sur son point au sol (x, y) : en vue isométrique, seul ce point est projeté.
func upright(x, y float64, fn func()) {
	if !cam.iso {
		fn()
		return
	}
	saved := cam
	px, py := cam.P(x, y)
	cam = view{s: saved.s, ox: float64(px) - x*saved.s, oy: float64(py) - y*saved.s}
	fn()
	cam = saved
}

// isoLocal : une petite maquette vue de biais (isométrique) au milieu de la vue de
// dessus : le point c reste à sa place à l'écran, tout le reste est projeté autour.
func isoLocal(c V, fn func()) {
	if cam.iso {
		fn()
		return
	}
	saved := cam
	nx, ny := cam.P(c.X, c.Y)
	ix, iy := isoProj(c.X, c.Y)
	cam = view{s: saved.s, ox: float64(nx) - ix*saved.s, oy: float64(ny) - iy*saved.s, iso: true}
	fn()
	cam = saved
}

// lift : le point du monde qui s'affiche h unités au-dessus du point au sol g
// (rails en hauteur, gerbes d'eau…).
func lift(g V, h float64) V {
	if !cam.iso {
		return V{X: g.X, Y: g.Y - h}
	}
	sx, sy := isoProj(g.X, g.Y)
	x, y := isoInv(sx, sy-h)
	return V{X: x, Y: y}
}

var cam view

func (v view) X(x float64) float32 { return float32(v.ox + x*v.s) }
func (v view) Y(y float64) float32 { return float32(v.oy + y*v.s) }
func (v view) L(l float64) float32 { return float32(l * v.s) }

func (v view) toLogical(px, py float64) sim.V {
	if v.iso {
		x, y := isoInv((px-v.ox)/v.s, (py-v.oy)/v.s)
		return sim.V{X: x, Y: y}
	}
	return sim.V{X: (px - v.ox) / v.s, Y: (py - v.oy) / v.s}
}

type V = sim.V

type colorRGBA = color.RGBA
type imageRect = image.Rectangle
type imagePt = image.Point

func itoa(n int) string { return group(int64(n)) }

func rgb(hex uint32) color.RGBA {
	return color.RGBA{uint8(hex >> 16), uint8(hex >> 8), uint8(hex), 255}
}

func rgba(hex uint32, a float64) color.RGBA {
	c := rgb(hex)
	return color.RGBA{uint8(float64(c.R) * a), uint8(float64(c.G) * a), uint8(float64(c.B) * a), uint8(255 * a)}
}

func shade(hex uint32, k float64) color.RGBA {
	c := rgb(hex)
	f := func(x uint8) uint8 {
		v := float64(x) * k
		if k > 1 {
			v = float64(x) + (255-float64(x))*(k-1)
		}
		return uint8(math.Max(0, math.Min(255, v)))
	}
	return color.RGBA{f(c.R), f(c.G), f(c.B), 255}
}

// ---------------------------------------------------------------- géométrie
//
// Toutes les formes sont découpées en triangles et accumulées dans un lot,
// envoyé à la carte graphique en une seule fois (bien plus rapide que des
// milliers de petits dessins).

var batch struct {
	dst *ebiten.Image
	vs  []ebiten.Vertex
	is  []uint16
}

func flush() {
	if batch.dst != nil && len(batch.is) > 0 {
		batch.dst.DrawTriangles(batch.vs, batch.is, whiteSub, &ebiten.DrawTrianglesOptions{})
	}
	batch.vs = batch.vs[:0]
	batch.is = batch.is[:0]
}

// prep prépare l'ajout de n sommets ; renvoie l'indice du premier.
func prep(dst *ebiten.Image, n int) uint16 {
	if dst != batch.dst || len(batch.vs)+n > 65000 {
		flush()
		batch.dst = dst
	}
	return uint16(len(batch.vs))
}

func vert(x, y float32, c color.Color) ebiten.Vertex {
	// les couleurs des sommets ne sont pas « prémultipliées » : on divise par l'alpha
	r, g, b, a := c.RGBA()
	fa := float32(a) / 0xffff
	k := float32(1)
	if a > 0 {
		k = 1 / fa
	}
	return ebiten.Vertex{DstX: x, DstY: y, SrcX: 1, SrcY: 1,
		ColorR: float32(r) / 0xffff * k, ColorG: float32(g) / 0xffff * k, ColorB: float32(b) / 0xffff * k, ColorA: fa}
}

// fan remplit un polygone (en pixels) vu depuis son centre.
func fan(dst *ebiten.Image, xs, ys []float32, c color.Color) {
	n := len(xs)
	if n < 3 {
		return
	}
	var cx, cy float32
	for i := range xs {
		cx += xs[i]
		cy += ys[i]
	}
	cx /= float32(n)
	cy /= float32(n)
	b := prep(dst, n+1)
	batch.vs = append(batch.vs, vert(cx, cy, c))
	for i := range xs {
		batch.vs = append(batch.vs, vert(xs[i], ys[i], c))
	}
	for i := 0; i < n; i++ {
		batch.is = append(batch.is, b, b+1+uint16(i), b+1+uint16((i+1)%n))
	}
}

// strip remplit l'anneau entre deux contours de même longueur.
func strip(dst *ebiten.Image, ox, oy, ix, iy []float32, c color.Color) {
	n := len(ox)
	b := prep(dst, 2*n)
	for i := 0; i < n; i++ {
		batch.vs = append(batch.vs, vert(ox[i], oy[i], c), vert(ix[i], iy[i], c))
	}
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		o1, i1, o2, i2 := b+uint16(2*i), b+uint16(2*i+1), b+uint16(2*j), b+uint16(2*j+1)
		batch.is = append(batch.is, o1, i1, o2, o2, i1, i2)
	}
}

func segs(rpx float32) int {
	n := int(rpx * 0.5)
	if n < 8 {
		n = 8
	}
	if n > 48 {
		n = 48
	}
	return n
}

var tmpX, tmpY, tmpX2, tmpY2 []float32

func ellipsePts(cx, cy, rx, ry float32, n int) ([]float32, []float32) {
	tmpX, tmpY = tmpX[:0], tmpY[:0]
	for i := 0; i < n; i++ {
		a := float64(i) / float64(n) * 2 * math.Pi
		tmpX = append(tmpX, cx+rx*float32(math.Cos(a)))
		tmpY = append(tmpY, cy+ry*float32(math.Sin(a)))
	}
	return tmpX, tmpY
}

func circle(dst *ebiten.Image, x, y, r float64, c color.Color) {
	R := cam.L(r)
	if R <= 0.2 {
		return
	}
	X, Y := cam.P(x, y)
	if cam.iso { // un cercle au sol : une ellipse aplatie
		xs, ys := ellipsePts(X, Y, R, R/2, segs(R))
		fan(dst, xs, ys, c)
		return
	}
	xs, ys := ellipsePts(X, Y, R, R, segs(R))
	fan(dst, xs, ys, c)
}

func ellipse(dst *ebiten.Image, cx, cy, rx, ry float64, c color.Color) {
	RX, RY := cam.L(rx), cam.L(ry)
	if cam.iso { // une ellipse au sol, projetée point par point
		n := segs(max(RX, RY))
		tmpX, tmpY = tmpX[:0], tmpY[:0]
		for i := 0; i < n; i++ {
			a := float64(i) / float64(n) * 2 * math.Pi
			X, Y := cam.P(cx+rx*math.Cos(a), cy+ry*math.Sin(a))
			tmpX, tmpY = append(tmpX, X), append(tmpY, Y)
		}
		fan(dst, tmpX, tmpY, c)
		return
	}
	xs, ys := ellipsePts(cam.X(cx), cam.Y(cy), RX, RY, segs(max(RX, RY)))
	fan(dst, xs, ys, c)
}

func ring(dst *ebiten.Image, x, y, r, w float64, c color.Color) {
	R, W := cam.L(r), cam.L(w)
	n := segs(R)
	X, Y := cam.P(x, y)
	sq := float32(1)
	if cam.iso {
		sq = 0.5
	}
	ox, oy := ellipsePts(X, Y, R+W/2, (R+W/2)*sq, n)
	tmpX2 = append(tmpX2[:0], ox...)
	tmpY2 = append(tmpY2[:0], oy...)
	ix, iy := ellipsePts(X, Y, max(0, R-W/2), max(0, R-W/2)*sq, n)
	strip(dst, tmpX2, tmpY2, ix, iy, c)
}

func quad(dst *ebiten.Image, x1, y1, x2, y2, x3, y3, x4, y4 float32, c color.Color) {
	b := prep(dst, 4)
	batch.vs = append(batch.vs, vert(x1, y1, c), vert(x2, y2, c), vert(x3, y3, c), vert(x4, y4, c))
	batch.is = append(batch.is, b, b+1, b+2, b, b+2, b+3)
}

func line(dst *ebiten.Image, x1, y1, x2, y2, w float64, c color.Color) {
	X1, Y1 := cam.P(x1, y1)
	X2, Y2 := cam.P(x2, y2)
	dx, dy := X2-X1, Y2-Y1
	l := float32(math.Hypot(float64(dx), float64(dy)))
	if l < 1e-4 {
		return
	}
	hw := cam.L(w) / 2
	nx, ny := -dy/l*hw, dx/l*hw
	quad(dst, X1+nx, Y1+ny, X2+nx, Y2+ny, X2-nx, Y2-ny, X1-nx, Y1-ny, c)
}

func rect(dst *ebiten.Image, x, y, w, h float64, c color.Color) {
	if cam.iso {
		x1, y1 := cam.P(x, y)
		x2, y2 := cam.P(x+w, y)
		x3, y3 := cam.P(x+w, y+h)
		x4, y4 := cam.P(x, y+h)
		quad(dst, x1, y1, x2, y2, x3, y3, x4, y4, c)
		return
	}
	X, Y, W, H := cam.X(x), cam.Y(y), cam.L(w), cam.L(h)
	quad(dst, X, Y, X+W, Y, X+W, Y+H, X, Y+H, c)
}

func strokeRect(dst *ebiten.Image, x, y, w, h, lw float64, c color.Color) {
	line(dst, x, y, x+w, y, lw, c)
	line(dst, x+w, y, x+w, y+h, lw, c)
	line(dst, x+w, y+h, x, y+h, lw, c)
	line(dst, x, y+h, x, y, lw, c)
}

var whiteImg = func() *ebiten.Image {
	img := ebiten.NewImage(3, 3)
	img.Fill(color.White)
	return img
}()
var whiteSub = whiteImg.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)

// roundPts : contour d'un rectangle arrondi, en pixels.
func roundPts(x, y, w, h, r float64, n int, xs, ys []float32) ([]float32, []float32) {
	r = math.Max(0, math.Min(r, math.Min(w, h)/2))
	if cam.iso { // le contour au sol, projeté point par point
		if n <= 0 {
			n = segs(cam.L(r))/4 + 2
		}
		corner := func(cx, cy, a0 float64) {
			for i := 0; i <= n; i++ {
				a := a0 + float64(i)/float64(n)*math.Pi/2
				X, Y := cam.P(cx+r*math.Cos(a), cy+r*math.Sin(a))
				xs, ys = append(xs, X), append(ys, Y)
			}
		}
		corner(x+w-r, y+r, -math.Pi/2)
		corner(x+w-r, y+h-r, 0)
		corner(x+r, y+h-r, math.Pi/2)
		corner(x+r, y+r, math.Pi)
		return xs, ys
	}
	X, Y, W, H, R := cam.X(x), cam.Y(y), cam.L(w), cam.L(h), cam.L(r)
	if n <= 0 {
		n = segs(R)/4 + 2
	}
	corner := func(cx, cy float32, a0 float64) {
		for i := 0; i <= n; i++ {
			a := a0 + float64(i)/float64(n)*math.Pi/2
			xs = append(xs, cx+R*float32(math.Cos(a)))
			ys = append(ys, cy+R*float32(math.Sin(a)))
		}
	}
	corner(X+W-R, Y+R, -math.Pi/2)
	corner(X+W-R, Y+H-R, 0)
	corner(X+R, Y+H-R, math.Pi/2)
	corner(X+R, Y+R, math.Pi)
	return xs, ys
}

func roundRect(dst *ebiten.Image, x, y, w, h, r float64, c color.Color) {
	if w <= 0 || h <= 0 {
		return
	}
	tmpX, tmpY = roundPts(x, y, w, h, r, 0, tmpX[:0], tmpY[:0])
	fan(dst, tmpX, tmpY, c)
}

func roundStroke(dst *ebiten.Image, x, y, w, h, r, lw float64, c color.Color) {
	d := lw / 2
	n := segs(cam.L(r+d))/4 + 2
	tmpX2, tmpY2 = roundPts(x-d, y-d, w+lw, h+lw, r+d, n, tmpX2[:0], tmpY2[:0])
	tmpX, tmpY = roundPts(x+d, y+d, w-lw, h-lw, math.Max(0, r-d), n, tmpX[:0], tmpY[:0])
	strip(dst, tmpX2, tmpY2, tmpX, tmpY, c)
}

func polygon(dst *ebiten.Image, pts []V, c color.Color) {
	tmpX, tmpY = tmpX[:0], tmpY[:0]
	for _, q := range pts {
		X, Y := cam.P(q.X, q.Y)
		tmpX = append(tmpX, X)
		tmpY = append(tmpY, Y)
	}
	fan(dst, tmpX, tmpY, c)
}

// dashedRect : contour en pointillés (zones à débloquer).
func dashedRect(dst *ebiten.Image, x, y, w, h, lw, dash float64, phase float64, c color.Color) {
	per := 2 * (w + h)
	pt := func(d float64) (float64, float64) {
		d = math.Mod(d+per, per)
		switch {
		case d < w:
			return x + d, y
		case d < w+h:
			return x + w, y + d - w
		case d < 2*w+h:
			return x + w - (d - w - h), y + h
		default:
			return x, y + h - (d - 2*w - h)
		}
	}
	edges := []float64{0, w, w + h, 2*w + h, per}
	for d := math.Mod(phase, dash*2) - dash*2; d < per; d += dash * 2 {
		a, b := math.Max(0, d), math.Min(d+dash, per)
		if b <= a {
			continue
		}
		// on coupe le tiret aux coins
		for k := 0; k < 4; k++ {
			lo, hi := math.Max(a, edges[k]), math.Min(b, edges[k+1])
			if hi > lo {
				x1, y1 := pt(lo)
				x2, y2 := pt(hi - 1e-9)
				line(dst, x1, y1, x2, y2, lw, c)
			}
		}
	}
}

func star(dst *ebiten.Image, cx, cy, r float64, c color.Color) {
	var pts []V
	for i := 0; i < 10; i++ {
		a := -math.Pi/2 + float64(i)*math.Pi/5
		rr := r
		if i%2 == 1 {
			rr = r * 0.45
		}
		pts = append(pts, V{X: cx + math.Cos(a)*rr, Y: cy + math.Sin(a)*rr})
	}
	polygon(dst, pts, c)
}

// ---------------------------------------------------------------- texte

var (
	fontBold *text.GoTextFaceSource
	fontReg  *text.GoTextFaceSource
)

func init() {
	fontBold, _ = text.NewGoTextFaceSource(bytes.NewReader(gobold.TTF))
	fontReg, _ = text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
}

// say écrit un texte centré (align 0), aligné à gauche (-1) ou à droite (1).
// Le texte est traduit dans la langue choisie (tr.T) : un texte écrit en
// français dans le code s'affiche en anglais s'il est dans langues/en.txt.
func say(dst *ebiten.Image, s string, x, y, size float64, bold bool, align int, c color.Color) {
	s = tr.T(s)
	flush()
	src := fontReg
	if bold {
		src = fontBold
	}
	f := &text.GoTextFace{Source: src, Size: size * cam.s}
	op := &text.DrawOptions{}
	X, Y := cam.P(x, y)
	op.GeoM.Translate(float64(X), float64(Y))
	switch align {
	case 0:
		op.PrimaryAlign = text.AlignCenter
	case 1:
		op.PrimaryAlign = text.AlignEnd
	}
	op.SecondaryAlign = text.AlignCenter
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, s, f, op)
}

func textWidth(s string, size float64, bold bool) float64 {
	return textWidthRaw(tr.T(s), size, bold)
}

// textWidthRaw : la largeur d'un texte déjà traduit.
func textWidthRaw(s string, size float64, bold bool) float64 {
	src := fontReg
	if bold {
		src = fontBold
	}
	w, _ := text.Measure(s, &text.GoTextFace{Source: src, Size: size}, 0)
	return w
}

// sayFit écrit un texte en réduisant sa taille s'il dépasse maxW.
func sayFit(dst *ebiten.Image, s string, x, y, size float64, bold bool, align int, maxW float64, c color.Color) {
	if w := textWidth(s, size, bold); w > maxW {
		size *= maxW / w
	}
	say(dst, s, x, y, size, bold, align, c)
}

// sayOutlined : texte avec un contour (lisible sur n'importe quel fond).
func sayOutlined(dst *ebiten.Image, s string, x, y, size float64, align int, c, outline color.Color) {
	d := size * 0.08
	for i := 0; i < 8; i++ {
		a := float64(i) * math.Pi / 4
		say(dst, s, x+math.Cos(a)*d, y+math.Sin(a)*d, size, true, align, outline)
	}
	say(dst, s, x, y, size, true, align, c)
}

// wrap coupe un texte en lignes de largeur maxW.
func wrap(s string, size, maxW float64, bold bool) []string {
	s = tr.T(s)
	var lines []string
	cur := ""
	word := ""
	flush := func() {
		if word == "" {
			return
		}
		try := word
		if cur != "" {
			try = cur + " " + word
		}
		if textWidthRaw(try, size, bold) > maxW && cur != "" {
			lines = append(lines, cur)
			cur = word
		} else {
			cur = try
		}
		word = ""
	}
	for _, r := range s {
		if r == ' ' {
			flush()
		} else if r == '\n' {
			flush()
			lines = append(lines, cur)
			cur = ""
		} else {
			word += string(r)
		}
	}
	flush()
	if cur != "" {
		lines = append(lines, cur)
	}
	for _, l := range lines {
		tr.Seen(l) // (une ligne coupée n'est pas un texte à traduire)
	}
	return lines
}

// ---------------------------------------------------------------- nombres

// money met en forme une somme : 1 234 €, 12,5 k€ n'est pas utilisé pour
// rester lisible ; au-delà du million : 1,25 M€.
func money(v float64) string {
	v = math.Floor(v + 1e-6)
	// à partir du million : notation scientifique, un chiffre après la virgule
	// (1,2 × 10⁶ €)
	if v >= 1e6 {
		e := int(math.Floor(math.Log10(v)))
		m := math.Round(v/math.Pow(10, float64(e))*10) / 10
		if m >= 10 { // 9,96 arrondi donne 10,0 → 1,0 × 10^(e+1)
			m /= 10
			e++
		}
		t := int(math.Round(m * 10)) // toujours un chiffre après la virgule, même 0
		if tr.English() {
			return "$" + string(rune('0'+t/10)) + "." + string(rune('0'+t%10)) + " × 10" + sup(e)
		}
		return string(rune('0'+t/10)) + "," + string(rune('0'+t%10)) + " × 10" + sup(e) + " €"
	}
	if tr.English() {
		return "$" + group(int64(v))
	}
	return group(int64(v)) + " €"
}

// rateText : un gain par seconde (+1,5 € / s ; en anglais +$1.5/s).
func rateText(v float64) string {
	if tr.English() {
		return "+$" + ftoa(v, 1) + "/s"
	}
	return "+" + ftoa(v, 1) + " € / s"
}

// sup écrit un entier positif en exposant (⁶, ¹²…).
func sup(n int) string {
	digits := []rune("⁰¹²³⁴⁵⁶⁷⁸⁹")
	s := ""
	for {
		s = string(digits[n%10]) + s
		n /= 10
		if n == 0 {
			return s
		}
	}
}

func group(n int64) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := ""
	if n == 0 {
		s = "0"
	}
	k := 0
	for n > 0 {
		if k > 0 && k%3 == 0 {
			if tr.English() {
				s = "," + s
			} else {
				s = " " + s
			}
		}
		s = string(rune('0'+n%10)) + s
		n /= 10
		k++
	}
	if neg {
		s = "-" + s
	}
	return s
}

func ftoa(v float64, dec int) string {
	p := math.Pow(10, float64(dec))
	n := int64(math.Round(v * p))
	ip := n / int64(p)
	fp := n % int64(p)
	s := group(ip)
	if dec > 0 {
		f := ""
		for i := 0; i < dec; i++ {
			f = string(rune('0'+fp%10)) + f
			fp /= 10
		}
		// on enlève les zéros inutiles
		for len(f) > 0 && f[len(f)-1] == '0' {
			f = f[:len(f)-1]
		}
		if f != "" {
			if tr.English() {
				s += "." + f
			} else {
				s += "," + f
			}
		}
	}
	return s
}
