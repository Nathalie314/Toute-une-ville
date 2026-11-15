package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// Les attractions du parc. Chacune grandit avec sa taille (1 à 10) : le circuit
// s'allonge, le décor s'enrichit. Les wagons, bateaux et nacelles sont à la
// taille des visiteurs, qu'on voit assis à bord.

// rideStart : la place de départ (le point du circuit le plus proche de la cabine).
func rideStart(r *sim.Ride, pts []V) float64 {
	b := r.Booth().C()
	st := 0
	for i, q := range pts {
		if q.Dist(b) < pts[st].Dist(b) {
			st = i
		}
	}
	return float64(st) / float64(len(pts))
}

// ridePos : où en est le premier wagon (au départ tant que le tour n'a pas commencé).
func ridePos(r *sim.Ride, s0 float64) float64 {
	if r.State == 1 {
		return s0 + r.Prog
	}
	return s0
}

// riderLooks : l'allure des passagers du wagon n° k (seats places).
func riderLooks(riders []*sim.Visitor, k, seats int) []int {
	var out []int
	for i := k * seats; i < min(len(riders), (k+1)*seats); i++ {
		out = append(out, riders[i].Look)
	}
	return out
}

func centroid(pts []V) V {
	c := V{}
	for _, q := range pts {
		c = c.Add(q)
	}
	return c.Mul(1 / float64(len(pts)))
}

func box(at, dir V, l, w float64) []V {
	n := V{X: -dir.Y, Y: dir.X}
	return []V{at.Add(dir.Mul(l)).Add(n.Mul(w)), at.Add(dir.Mul(l)).Sub(n.Mul(w)), at.Sub(dir.Mul(l)).Sub(n.Mul(w)), at.Sub(dir.Mul(l)).Add(n.Mul(w))}
}

// seated : un passager assis (le haut du corps), à la taille des visiteurs.
// arms : 0 bras le long du corps, 1 bras levés (grand huit), 2 mains sur la barre.
func seated(dst *ebiten.Image, x, y float64, lookN, arms int, t float64) {
	lk := lookFrom(lookN)
	switch arms {
	case 1:
		w := 3 * math.Sin(t*9+float64(lookN))
		line(dst, x-13, y-22, x-22+w, y-58, 8, rgb(lk.shirt))
		line(dst, x+13, y-22, x+22-w, y-58, 8, rgb(lk.shirt))
		circle(dst, x-22+w, y-61, 5, rgb(lk.skin))
		circle(dst, x+22-w, y-61, 5, rgb(lk.skin))
	case 2:
		line(dst, x-13, y-20, x-14, y-4, 8, rgb(lk.shirt))
		line(dst, x+13, y-20, x+14, y-4, 8, rgb(lk.shirt))
	}
	if lk.long {
		roundRect(dst, x-14, y-50, 28, 30, 10, rgb(lk.hair))
	}
	roundRect(dst, x-15, y-28, 30, 28, 10, rgb(lk.shirt))
	rect(dst, x-15, y-28, 8, 26, shade(lk.shirt, 1.12))
	circle(dst, x, y-40, 14, rgb(lk.skin))
	polygon(dst, []V{{X: x - 14, Y: y - 42}, {X: x - 12, Y: y - 51}, {X: x, Y: y - 56}, {X: x + 12, Y: y - 51}, {X: x + 14, Y: y - 42}, {X: x + 7, Y: y - 47}, {X: x - 7, Y: y - 47}}, rgb(lk.hair))
	circle(dst, x-5, y-38, 2.2, rgb(0x2a2a2a))
	circle(dst, x+5, y-38, 2.2, rgb(0x2a2a2a))
	if arms == 1 { // il crie de joie
		ellipse(dst, x, y-31, 4, 3.5, rgb(0x6b1f1f))
	}
}

// ridersIn : dessine les passagers d'un véhicule (rangées de deux, dans le sens de la marche).
func ridersIn(dst *ebiten.Image, at, dir V, looks []int, perRow, rowsCap int, pitch float64, arms int, t float64) {
	n := V{X: -dir.Y, Y: dir.X}
	type s struct {
		q V
		l int
	}
	var ss []s
	for i, l := range looks {
		row, col := i/perRow, i%perRow
		off := float64(col) - float64(perRow-1)/2
		q := at.Add(dir.Mul(pitch * (float64(rowsCap-1)/2 - float64(row)))).Add(n.Mul(off * 30))
		ss = append(ss, s{q, l})
	}
	for i := 1; i < len(ss); i++ { // de l'arrière vers l'avant de l'écran
		for j := i; j > 0 && ss[j].q.Y < ss[j-1].q.Y; j-- {
			ss[j], ss[j-1] = ss[j-1], ss[j]
		}
	}
	for _, e := range ss {
		seated(dst, e.q.X, e.q.Y+10, e.l, arms, t)
	}
}

// vehicle : la caisse d'un véhicule vue de dessus, en relief (ombre, bord clair).
func vehicle(dst *ebiten.Image, at, dir V, l, w float64, col uint32) {
	polygon(dst, box(at.Add(V{X: 8, Y: 12}), dir, l, w), rgba(0x1d3a12, 0.25))
	polygon(dst, box(at.Add(V{Y: 6}), dir, l, w), shade(col, 0.7))
	polygon(dst, box(at, dir, l, w), rgb(col))
	polygon(dst, box(at, dir, l-8, w-8), shade(col, 0.82))
}

// scatter : des positions fixes (selon la graine) dans la zone, loin du circuit.
func scatter(r *sim.Ride, pts []V, n, seed int, keep float64) []V {
	return scatterRect(r.Track(), r.Kind.Slot, pts, n, seed, keep)
}

// scatterRect : comme scatter, dans un rectangle donné.
func scatterRect(tr sim.Rect, slot int, pts []V, n, seed int, keep float64) []V {
	var out []V
	for i := 0; i < n*4 && len(out) < n; i++ {
		x := tr.X + 50 + math.Mod(float64(i+seed)*373.7+float64(slot)*91, tr.W-100)
		y := tr.Y + 70 + math.Mod(float64(i+seed)*211.3+float64(slot)*57, tr.H-110)
		q := V{X: x, Y: y}
		if trackDist(pts, q) < keep {
			continue
		}
		ok := true
		for _, o := range out {
			if o.Dist(q) < 70 {
				ok = false
				break
			}
		}
		if ok {
			out = append(out, q)
		}
	}
	return out
}

// trackDist : distance d'un point au circuit.
func trackDist(pts []V, q V) float64 {
	d := 1e9
	for i := range pts {
		a, b := pts[i], pts[(i+1)%len(pts)]
		ab := b.Sub(a)
		l2 := ab.X*ab.X + ab.Y*ab.Y
		f := 0.0
		if l2 > 0 {
			f = math.Max(0, math.Min(1, ((q.X-a.X)*ab.X+(q.Y-a.Y)*ab.Y)/l2))
		}
		d = math.Min(d, q.Dist(a.Add(ab.Mul(f))))
	}
	return d
}

// ---------------------------------------------------------------- petit train

// drawTrainRide : le petit train. Taille 1 : une petite boucle près de la gare ;
// puis le circuit s'allonge et se tortille ; un étang et son pont (taille 4), un
// tunnel dans une colline (taille 6), un village (taille 8), une montagne
// rocheuse et sa cascade (taille 10).
func drawTrainRide(dst *ebiten.Image, p *sim.Place, r *sim.Ride, level int, t float64) {
	pts := trackLoop(r, level, 0.18)
	tr := r.Track()
	c := centroid(pts)
	// étang (taille 4+), au milieu de la boucle
	if level >= 4 {
		rx, ry := 180+18*float64(level), 90+9*float64(level)
		ellipse(dst, c.X, c.Y+10, rx+14, ry+10, rgb(0xd9c48f))
		ellipse(dst, c.X, c.Y+10, rx, ry, rgb(0x5fa8dc))
		ellipse(dst, c.X-rx*0.3, c.Y-ry*0.2, rx*0.35, ry*0.25, rgba(0xffffff, 0.25))
		for i := 0; i < 4; i++ { // nénuphars
			a := float64(i) * 1.7
			circle(dst, c.X+math.Cos(a)*rx*0.6, c.Y+10+math.Sin(a)*ry*0.5, 12, rgb(0x5fae4c))
		}
	}
	// le village (taille 8+) : de vraies petites maisons, autour de l'étang
	var houses []V
	if level >= 8 {
		for _, q := range scatter(r, pts, 12, 41, 150) {
			if math.Abs(q.X-c.X) < 180+18*float64(level)+90 && math.Abs(q.Y-c.Y) < 90+9*float64(level)+120 {
				continue // pas dans l'étang
			}
			if len(houses) < 5 {
				houses = append(houses, q)
			}
		}
		for i, q := range houses {
			building(dst, sim.Rect{X: q.X - 60, Y: q.Y - 90, W: 120, H: 100}, []uint32{0xf3ead8, 0xffe8b0, 0xd8e8f8, 0xf6d6c8}[i%4], []uint32{0xc0623a, 0x5a6a8a, 0x8a5a2b, 0x2f8a6a}[i%4])
			roundRect(dst, q.X-14, q.Y-30, 28, 40, 6, rgb(0x6e4322)) // la porte
		}
	}
	mountain := V{X: tr.X + tr.W - 260, Y: tr.Y + 340}
	if math.Min(trackDist(pts, mountain), math.Min(trackDist(pts, mountain.Add(V{X: -260})), trackDist(pts, mountain.Add(V{X: 260})))) < 160 {
		mountain = V{X: tr.X + 260, Y: tr.Y + 340}
	}
	decorTrees(dst, r, level, pts, func(q V) bool {
		if level >= 4 && math.Abs(q.X-c.X) < 180+18*float64(level)+60 && math.Abs(q.Y-c.Y) < 90+9*float64(level)+110 {
			return true // l'étang
		}
		for _, h := range houses {
			if q.Dist(h) < 150 {
				return true
			}
		}
		return level >= 10 && math.Abs(q.X-mountain.X) < 340 && q.Y < mountain.Y+120 && q.Y > mountain.Y-400
	})
	// le ballast puis les rails
	drawLoop(dst, pts, 70, 0xc9b48a)
	drawLoop(dst, pts, 54, 0xa98f68)
	for i := range pts {
		a, b := pts[i], pts[(i+1)%len(pts)]
		d := b.Sub(a).Norm()
		n := V{X: -d.Y, Y: d.X}
		line(dst, a.X-n.X*30, a.Y-n.Y*30, a.X+n.X*30, a.Y+n.Y*30, 8, rgb(0x6e4322))
	}
	railPair(dst, pts, 18, 0xd8d8d8)
	// pont sur l'étang
	if level >= 4 {
		q, dir := along(pts, 0.25)
		polygon(dst, box(q, dir, 70, 46), rgb(0x8a5a2b))
		polygon(dst, box(q, dir, 64, 38), rgb(0xa0703b))
	}
	s0 := rideStart(r, pts)
	g := pts[int(s0*float64(len(pts)))%len(pts)]
	// la gare
	building(dst, sim.Rect{X: g.X - 130, Y: g.Y - 230, W: 260, H: 150}, 0xf3ead8, 0xc0623a)
	roundRect(dst, g.X-60, g.Y-110, 120, 30, 8, rgb(0x3b2a52))
	sayFit(dst, "GARE", g.X, g.Y-95, 20, true, 0, 110, rgb(0xffe066))
	// le train : la locomotive et ses wagons
	units := p.RideUnits(r)
	pos := ridePos(r, s0)
	gap := 128.0 / pathLen(pts)
	type car struct {
		at, dir V
		k       int
	}
	var cars []car
	for k := 0; k <= units; k++ {
		at, dir := along(pts, pos-float64(k)*gap)
		cars = append(cars, car{at, dir, k})
	}
	for i := 1; i < len(cars); i++ {
		for j := i; j > 0 && cars[j].at.Y < cars[j-1].at.Y; j-- {
			cars[j], cars[j-1] = cars[j-1], cars[j]
		}
	}
	for _, cr := range cars {
		if cr.k == 0 {
			drawLoco(dst, cr.at, cr.dir, t, r.State == 1)
			continue
		}
		vehicle(dst, cr.at, cr.dir, 54, 32, []uint32{0x4fc3ff, 0xffd23f, 0x7ed957, 0xff6fa8}[cr.k%4])
		ridersIn(dst, cr.at, cr.dir, riderLooks(r.Riders, cr.k-1, 2), 1, 2, 40, 0, t)
	}
	// tunnel dans une colline (taille 6+), par-dessus les rails
	if level >= 6 {
		h, hd := along(pts, 0.62)
		ellipse(dst, h.X+30, h.Y+30, 230, 120, rgba(0x1d3a12, 0.25))
		ellipse(dst, h.X, h.Y, 220, 130, rgb(0x5f9e45))
		ellipse(dst, h.X-30, h.Y-30, 160, 90, rgb(0x76b85a))
		ellipse(dst, h.X-60, h.Y-50, 70, 40, rgb(0x8fcc6e))
		for i := 0; i < 4; i++ {
			drawBush(dst, h.X-120+float64(i)*80, h.Y-40+float64(i%2)*40, 0.8, i)
		}
		for _, sgn := range []float64{-1, 1} {
			e := h.Add(hd.Mul(sgn * 200))
			ellipse(dst, e.X, e.Y, 50, 44, rgb(0x6a5a4a))
			ellipse(dst, e.X, e.Y+6, 38, 34, rgb(0x15120f))
		}
	}
	// une montagne rocheuse, ses neiges et sa cascade (taille 10)
	if level >= 10 {
		drawMountain(dst, mountain, t)
	}
}

// railPair : deux rails parallèles le long du circuit.
func railPair(dst *ebiten.Image, pts []V, half float64, c uint32) {
	for i := range pts {
		a, b := pts[i], pts[(i+1)%len(pts)]
		d := b.Sub(a).Norm()
		n := V{X: -d.Y, Y: d.X}
		for _, sg := range []float64{-1, 1} {
			line(dst, a.X+n.X*half*sg, a.Y+n.Y*half*sg, b.X+n.X*half*sg, b.Y+n.Y*half*sg, 5, rgb(c))
		}
	}
}

// drawMountain : une montagne à facettes (plusieurs sommets, faces éclairées et à
// l'ombre, neige, rochers au pied) et une cascade qui tombe dans un bassin.
func drawMountain(dst *ebiten.Image, m V, t float64) {
	ellipse(dst, m.X+60, m.Y+40, 300, 70, rgba(0x1d3a12, 0.25))
	peaks := []struct{ dx, h, w float64 }{{-150, 230, 150}, {120, 260, 160}, {0, 360, 190}}
	for _, pk := range peaks {
		top := V{X: m.X + pk.dx, Y: m.Y - pk.h}
		l, rr := V{X: m.X + pk.dx - pk.w, Y: m.Y + 30}, V{X: m.X + pk.dx + pk.w, Y: m.Y + 30}
		mid := V{X: m.X + pk.dx + pk.w*0.15, Y: m.Y + 30}
		ridge := V{X: m.X + pk.dx - pk.w*0.1, Y: m.Y - pk.h*0.45}
		polygon(dst, []V{l, top, ridge, mid}, rgb(0xa8a59a))
		polygon(dst, []V{top, rr, mid, ridge}, rgb(0x76736a))
		polygon(dst, []V{l, ridge, mid}, rgb(0x96938a))
		// la neige
		sl := top.Add(l.Sub(top).Mul(0.28))
		sr := top.Add(rr.Sub(top).Mul(0.28))
		sm := top.Add(ridge.Sub(top).Mul(0.5))
		polygon(dst, []V{top, sl, V{X: sl.X + 18, Y: sl.Y - 10}, sm}, rgb(0xffffff))
		polygon(dst, []V{top, sm, V{X: sr.X - 14, Y: sr.Y - 6}, sr}, rgb(0xdfe6ee))
	}
	// rochers au pied
	for i := 0; i < 6; i++ {
		drawRock(dst, m.X-260+float64(i)*100, m.Y+50+float64(i%2)*20, 1.2, i)
	}
	// la cascade et son bassin
	fx := m.X + 60
	ellipse(dst, fx, m.Y+80, 90, 30, rgb(0x5fa8dc))
	rect(dst, fx-16, m.Y-170, 32, 250, rgba(0x9fd3ff, 0.95))
	for k := 0; k < 5; k++ {
		y := m.Y - 170 + math.Mod(t*160+float64(k)*50, 250)
		line(dst, fx-12, y, fx-12, y+20, 3, rgba(0xffffff, 0.9))
		line(dst, fx+8, y+25, fx+8, y+42, 3, rgba(0xffffff, 0.8))
	}
	for k := 0; k < 6; k++ {
		a := t*3 + float64(k)
		circle(dst, fx+math.Cos(a)*40, m.Y+76+math.Sin(a)*8, 7, rgba(0xffffff, 0.7))
	}
}

// drawLoco : la locomotive (rouge, avec sa cheminée qui fume).
func drawLoco(dst *ebiten.Image, at, dir V, t float64, running bool) {
	vehicle(dst, at, dir, 62, 34, 0xd9534f)
	c := at.Add(dir.Mul(30))
	circle(dst, c.X, c.Y-8, 16, rgb(0x2b2b2b))
	circle(dst, c.X, c.Y-12, 10, rgb(0x4a4a4a))
	if running {
		for k := 0; k < 3; k++ {
			f := math.Mod(t*1.5+float64(k)/3, 1)
			circle(dst, c.X-dir.X*60*f, c.Y-30-60*f, 10+14*f, rgba(0xffffff, 0.7*(1-f)))
		}
	}
	cab := at.Sub(dir.Mul(26))
	roundRect(dst, cab.X-30, cab.Y-54, 60, 50, 8, rgb(0x2b4a8a))
	roundRect(dst, cab.X-22, cab.Y-46, 44, 18, 4, rgb(0x9fd3ff))
}

// ---------------------------------------------------------------- train fantôme

// ghostTrack : le circuit du train fantôme : un serpentin (de plus en plus long
// avec la taille), puis un retour par le fond. Les voies restent bien écartées.
func ghostTrack(r *sim.Ride, level int) []V {
	tr := r.Track()
	f := float64(level) / 10
	w := tr.W * (0.36 + 0.6*f)
	h := tr.H * math.Min(0.9, 0.55+0.4*f)
	x0 := tr.X + 70
	near, far := tr.Y+70, tr.Y+70+h // côté boulevard, côté fond
	if r.Top {
		near, far = tr.Y+tr.H-70, tr.Y+tr.H-70-h
	}
	sg := 1.0
	if near < far {
		sg = -1
	}
	// le serpentin entre « a » et « b » ; le retour par la voie du fond
	ret, a, b := far, far+300*sg, near
	cols := 2 * (1 + level/3)
	step := (w - 360) / float64(cols)
	raw := []V{{X: x0, Y: ret}, {X: x0, Y: b}}
	y := b
	for i := 0; i < cols; i++ {
		x := x0 + 260 + float64(i)*step
		raw = append(raw, V{X: x, Y: y})
		if y == b {
			y = a
		} else {
			y = b
		}
		raw = append(raw, V{X: x, Y: y})
	}
	xe := x0 + w
	raw = append(raw, V{X: xe, Y: y}, V{X: xe, Y: ret})
	return resample(smoothClosed(raw, 4), 260)
}

// smoothClosed : arrondit les coins d'un circuit fermé (Chaikin).
func smoothClosed(pts []V, iters int) []V {
	for k := 0; k < iters; k++ {
		var out []V
		for i := range pts {
			a, b := pts[i], pts[(i+1)%len(pts)]
			out = append(out, a.Mul(0.75).Add(b.Mul(0.25)), a.Mul(0.25).Add(b.Mul(0.75)))
		}
		pts = out
	}
	return pts
}

// resample : n points régulièrement espacés le long d'un circuit fermé.
func resample(pts []V, n int) []V {
	out := make([]V, 0, n)
	for i := 0; i < n; i++ {
		q, _ := along(pts, float64(i)/float64(n))
		out = append(out, q)
	}
	return out
}

// inTunnel : la fraction s du circuit est-elle sous un tunnel ?
func inTunnel(s float64, tunnels [][2]float64) bool {
	s = math.Mod(s+10, 1)
	for _, tn := range tunnels {
		if s >= tn[0] && s <= tn[1] {
			return true
		}
	}
	return false
}

// drawMonster : un monstre (0 squelette, 1 momie, 2 vampire, 3 zombie) ; up : 0 caché, 1 surgi.
func drawMonster(dst *ebiten.Image, kind int, at V, up, t float64) {
	// sa cachette : une tombe, un sarcophage, un cercueil, un tas de terre
	switch kind {
	case 0:
		roundRect(dst, at.X-22, at.Y-40, 44, 44, 14, rgb(0x8a8a8a))
		rect(dst, at.X-3, at.Y-34, 6, 20, rgb(0x6a6a6a))
		rect(dst, at.X-10, at.Y-28, 20, 6, rgb(0x6a6a6a))
	case 1:
		roundRect(dst, at.X-20, at.Y-50, 40, 54, 8, rgb(0xc8a050))
		roundRect(dst, at.X-12, at.Y-44, 24, 20, 8, rgb(0x2f6a8a))
	case 2:
		polygon(dst, []V{{X: at.X - 16, Y: at.Y - 52}, {X: at.X + 16, Y: at.Y - 52}, {X: at.X + 22, Y: at.Y + 6}, {X: at.X - 22, Y: at.Y + 6}}, rgb(0x4a2a1a))
	default:
		ellipse(dst, at.X, at.Y, 32, 14, rgb(0x6a4a2a))
	}
	if up < 0.05 {
		return
	}
	h := 90 * up // il se dresse
	y := at.Y - 14 - h
	body := map[int]uint32{0: 0xf2f2e6, 1: 0xe8e0c8, 2: 0x2b2b3a, 3: 0x5a7a4a}[kind]
	face := map[int]uint32{0: 0xf2f2e6, 1: 0xe8e0c8, 2: 0xf0e8f0, 3: 0x8ab06a}[kind]
	rect(dst, at.X-14, y+18, 28, h*0.8, rgb(body))
	if kind == 2 { // la cape du vampire, doublée de rouge
		polygon(dst, []V{{X: at.X - 14, Y: y + 20}, {X: at.X - 34, Y: y + 20 + h*0.75}, {X: at.X - 10, Y: y + 20 + h*0.7}}, rgb(0xc0392b))
		polygon(dst, []V{{X: at.X + 14, Y: y + 20}, {X: at.X + 34, Y: y + 20 + h*0.75}, {X: at.X + 10, Y: y + 20 + h*0.7}}, rgb(0xc0392b))
	}
	if kind == 1 { // les bandelettes de la momie
		for k := 0.0; k < h*0.8; k += 9 {
			line(dst, at.X-14, y+20+k, at.X+14, y+25+k, 2.5, rgb(0xb8a888))
		}
	}
	if kind == 0 { // les côtes du squelette
		for k := 0.0; k < h*0.5; k += 10 {
			line(dst, at.X-10, y+26+k, at.X+10, y+26+k, 3.5, rgb(0x5a5a5a))
		}
	}
	// les bras levés (ils s'agitent)
	wave := 8 * math.Sin(t*10+at.X)
	line(dst, at.X-14, y+26, at.X-32, y+2+wave, 7, rgb(face))
	line(dst, at.X+14, y+26, at.X+32, y+2-wave, 7, rgb(face))
	circle(dst, at.X, y+8, 16, rgb(face))
	ey := map[int]uint32{0: 0x1b1b1b, 1: 0x1b1b1b, 2: 0xd9534f, 3: 0xd9534f}[kind]
	circle(dst, at.X-5, y+5, 3.5, rgb(ey))
	circle(dst, at.X+5, y+5, 3.5, rgb(ey))
	if kind == 2 {
		polygon(dst, []V{{X: at.X - 16, Y: y - 4}, {X: at.X, Y: y + 4}, {X: at.X + 16, Y: y - 4}, {X: at.X, Y: y - 12}}, rgb(0x1b1b1b))
	}
}

// drawWitch : une sorcière qui touille son chaudron (at : pied de la sorcière ;
// le chaudron est à sa droite).
func drawWitch(dst *ebiten.Image, at V, t float64) {
	ellipse(dst, at.X+50, at.Y+28, 70, 16, rgba(0x000000, 0.25))
	cx := at.X + 70
	// le feu sous le chaudron
	for k := 0; k < 4; k++ {
		fl := 14 + 8*math.Sin(t*9+float64(k))
		polygon(dst, []V{{X: cx - 30 + float64(k)*16, Y: at.Y + 26}, {X: cx - 22 + float64(k)*16, Y: at.Y + 26 - fl}, {X: cx - 14 + float64(k)*16, Y: at.Y + 26}}, rgb(0xff8a3d))
	}
	ellipse(dst, cx, at.Y, 50, 34, rgb(0x2b2b2b))
	ellipse(dst, cx, at.Y-16, 42, 13, rgb(0x6ad04a))
	for k := 0; k < 5; k++ { // les bulles
		b := math.Mod(t*1.5+float64(k)*0.2, 1)
		circle(dst, cx-24+float64(k)*12, at.Y-18-b*60, 5+3*b, rgba(0x8af06a, 1-b))
	}
	// la sorcière : robe noire, visage vert, chapeau pointu ; la louche tourne
	a := t * 3
	sx, sy := cx+math.Cos(a)*20, at.Y-16+math.Sin(a)*6
	polygon(dst, []V{{X: at.X - 30, Y: at.Y + 26}, {X: at.X - 4, Y: at.Y - 64}, {X: at.X + 22, Y: at.Y + 26}}, rgb(0x2b1b3b))
	line(dst, at.X+4, at.Y-40, sx, sy-40, 6, rgb(0x6aa04a))
	line(dst, sx, sy-48, sx, sy, 5, rgb(0x6e4322))
	circle(dst, at.X-4, at.Y-76, 16, rgb(0x6aa04a))
	circle(dst, at.X+3, at.Y-76, 3.5, rgb(0x1b1b1b))
	polygon(dst, []V{{X: at.X - 30, Y: at.Y - 86}, {X: at.X - 4, Y: at.Y - 150}, {X: at.X + 22, Y: at.Y - 86}}, rgb(0x1b1b1b))
	rect(dst, at.X-38, at.Y-90, 68, 8, rgb(0x1b1b1b))
}

// drawGhostRide : le train fantôme. Son circuit serpente (de plus en plus avec la
// taille) ; il disparaît sous des tunnels sombres ; des squelettes, des momies,
// des vampires et des zombies surgissent à son passage ; une sorcière touille son
// chaudron. Tout le décor reste à l'écart des rails.
func drawGhostRide(dst *ebiten.Image, p *sim.Place, r *sim.Ride, level int, t float64) {
	pts := ghostTrack(r, level)
	c := centroid(pts)
	tr := r.Track()
	f := float64(level)
	ellipse(dst, c.X, c.Y, math.Min(tr.W/2, 500+90*f), math.Min(tr.H/2, 300+40*f), rgba(0x3b3b2b, 0.25))
	// la sorcière : l'endroit libre le plus loin des rails
	witch := tr.C()
	best := 0.0
	for gx := tr.X + 120; gx < tr.X+tr.W-200; gx += 60 {
		for gy := tr.Y + 180; gy < tr.Y+tr.H-60; gy += 60 {
			q := V{X: gx, Y: gy}
			d := math.Min(trackDist(pts, q), trackDist(pts, q.Add(V{X: 70})))
			if d > best {
				best, witch = d, q
			}
		}
	}
	// citrouilles et arbres morts, loin des rails et de la sorcière
	for i, q := range scatter(r, pts, 3+2*level, 11, 110) {
		if q.Dist(witch.Add(V{X: 40})) < 170 {
			continue
		}
		switch i % 3 {
		case 0:
			circle(dst, q.X, q.Y, 20, rgb(0xff8a3d))
			ellipse(dst, q.X-8, q.Y, 7, 18, rgb(0xf07a2d))
			rect(dst, q.X-3, q.Y-26, 6, 9, rgb(0x2f6a3a))
			circle(dst, q.X-7, q.Y-3, 3.5, rgb(0x3b2a10))
			circle(dst, q.X+7, q.Y-3, 3.5, rgb(0x3b2a10))
		default:
			line(dst, q.X, q.Y, q.X, q.Y-90, 9, rgb(0x4a3a2a))
			line(dst, q.X, q.Y-60, q.X-30, q.Y-92, 6, rgb(0x4a3a2a))
			line(dst, q.X, q.Y-72, q.X+26, q.Y-104, 6, rgb(0x4a3a2a))
		}
	}
	drawLoop(dst, pts, 64, 0x2b2b2b)
	railPair(dst, pts, 18, 0x7a5a9a)
	s0 := rideStart(r, pts)
	// les tunnels : un, puis un de plus tous les 3 niveaux (jamais à la gare)
	var tunnels [][2]float64
	nt := 1 + level/3
	for i := 0; i < nt; i++ {
		a := math.Mod(s0+0.1+float64(i)*0.8/float64(nt), 1)
		tunnels = append(tunnels, [2]float64{a, math.Min(0.999, a+0.05)})
	}
	pos := ridePos(r, s0)
	units := p.RideUnits(r)
	gap := 100.0 / pathLen(pts)
	var cars []V
	for k := 0; k < units; k++ {
		s := pos - float64(k)*gap
		at, dir := along(pts, s)
		cars = append(cars, at)
		if inTunnel(s, tunnels) {
			continue // sous le tunnel
		}
		vehicle(dst, at, dir, 44, 32, 0x7a4bc8)
		front := at.Add(dir.Mul(40))
		circle(dst, front.X, front.Y-12, 13, rgb(0xe8e8e8)) // la tête de mort à l'avant
		circle(dst, front.X-4, front.Y-14, 3, rgb(0x1b1b1b))
		circle(dst, front.X+4, front.Y-14, 3, rgb(0x1b1b1b))
		ridersIn(dst, at, dir, riderLooks(r.Riders, k, 2), 1, 2, 40, 0, t)
	}
	// les tunnels par-dessus les rails : une voûte sombre et ses deux bouches
	for _, tn := range tunnels {
		var seg []V
		for s := tn[0]; s <= tn[1]; s += 0.003 {
			q, _ := along(pts, s)
			seg = append(seg, q)
		}
		for i := 0; i+1 < len(seg); i++ {
			line(dst, seg[i].X, seg[i].Y, seg[i+1].X, seg[i+1].Y, 110, rgb(0x3b2b4b))
			circle(dst, seg[i].X, seg[i].Y, 55, rgb(0x3b2b4b))
		}
		for i := 0; i+1 < len(seg); i++ {
			line(dst, seg[i].X, seg[i].Y-10, seg[i+1].X, seg[i+1].Y-10, 76, rgb(0x4a3a5a))
		}
		for _, e := range []V{seg[0], seg[len(seg)-1]} {
			circle(dst, e.X, e.Y, 52, rgb(0x2b1b3b))
			circle(dst, e.X, e.Y+8, 36, rgb(0x0b0b10))
		}
	}
	// les monstres, au bord des rails (jamais dessus, jamais devant la sorcière) :
	// ils surgissent quand un wagon arrive
	nm := min(12, 2+level)
	kinds := []int{0, 1, 3, 2}
	placed := 0
	for i := 0; i < 60 && placed < nm; i++ {
		s := math.Mod(s0+0.06+float64(i)*0.9/float64(nm)+float64(i/nm)*0.013, 1)
		if inTunnel(s, tunnels) {
			continue
		}
		q, dir := along(pts, s)
		side := V{X: -dir.Y, Y: dir.X}.Mul(float64(1 - 2*(i%2)))
		at := q.Add(side.Mul(95))
		if trackDist(pts, at) < 85 || trackDist(pts, at.Add(V{Y: -100})) < 50 {
			at = q.Sub(side.Mul(95))
			if trackDist(pts, at) < 85 || trackDist(pts, at.Add(V{Y: -100})) < 50 {
				continue
			}
		}
		if at.Dist(witch.Add(V{X: 40})) < 220 || at.Y < tr.Y+60 {
			continue
		}
		near := 1e9
		for _, cq := range cars {
			near = math.Min(near, cq.Dist(at))
		}
		up := math.Max(0, math.Min(1, (300-near)/120))
		drawMonster(dst, kinds[placed%4], at, up, t)
		placed++
	}
	drawWitch(dst, witch, t)
	// fantômes et chauves-souris
	for i := 0; i < level/2; i++ {
		a := t*0.8 + float64(i)*1.7
		x, y := c.X+math.Cos(a)*tr.W*0.3, c.Y-80+math.Sin(a*1.3)*tr.H*0.22
		ellipse(dst, x, y, 26, 34, rgba(0xffffff, 0.75))
		circle(dst, x-8, y-8, 4, rgb(0x2b2b2b))
		circle(dst, x+8, y-8, 4, rgb(0x2b2b2b))
	}
	for i := 0; i < (level+1)/2; i++ {
		a := t*1.6 + float64(i)*2.3
		x, y := c.X+math.Cos(a)*tr.W*0.32, c.Y-tr.H*0.25-40*math.Sin(a*2)
		wing := 16 * math.Abs(math.Sin(t*12+float64(i)))
		polygon(dst, []V{{X: x, Y: y}, {X: x - 30, Y: y - wing}, {X: x - 14, Y: y + 5}}, rgb(0x2b2b2b))
		polygon(dst, []V{{X: x, Y: y}, {X: x + 30, Y: y - wing}, {X: x + 14, Y: y + 5}}, rgb(0x2b2b2b))
	}
}

// ---------------------------------------------------------------- bateaux et rondins

// boatAt : où est un bateau (rondin) autonome sur son circuit (fraction de tour).
// Il part du quai (s0), fait le tour et débarque juste avant le quai ; ceux qui
// attendent se rangent derrière.
func boatAt(b *sim.RideBoat, s0, gapL float64) float64 {
	switch b.State {
	case 2:
		return s0 + b.Prog*(1-gapL)
	case 3:
		return s0 - gapL
	case 1:
		return s0
	}
	return s0 - gapL*float64(b.Line+1)
}

// drawBoatRide : une rivière aux courbes douces dans un jardin : un kiosque (4),
// des statues, des parterres de fleurs et des fontaines (de plus en plus avec la
// taille), des cygnes, des arbres.
// Chaque bateau part seul, dès qu'il est plein.
func drawBoatRide(dst *ebiten.Image, p *sim.Place, r *sim.Ride, level int, t float64) {
	pts := boatTrack(r, level)
	bg := boatGeom(r, level)
	c := V{X: (bg.xl + bg.xr) / 2, Y: (bg.yb0+bg.yBot)/2 + 60} // (le kiosque : au milieu de la boucle du bas)
	// le décor : statues, parterres de fleurs, fontaines (de plus en plus avec la taille),
	// loin de la rivière et du kiosque
	var spots []V
	cand := scatter(r, pts, 60, 23, 190)
	cand = append(cand, scatterRect(sim.ParcSlot(3), r.Kind.Slot+3, pts, 40, 29, 190)...)
	for _, q := range cand {
		if level >= 4 && math.Abs(q.X-c.X) < 220 && math.Abs(q.Y-c.Y) < 200 {
			continue
		}
		ok := true
		for _, o := range spots {
			if o.Dist(q) < 220 {
				ok = false
				break
			}
		}
		if ok {
			spots = append(spots, q)
		}
	}
	nStat, nBed, nFont := 1+level/2, 2+level, level/3
	decorTrees(dst, r, level/2+1, pts, func(q V) bool {
		if level >= 4 && math.Abs(q.X-c.X) < 220 && math.Abs(q.Y-c.Y) < 220 {
			return true
		}
		for i, o := range spots {
			if i < 3+2*level && o.Dist(q) < 170 {
				return true
			}
		}
		return false
	})
	// (dans l'ordre : fontaine, statue, parterres… pour que chaque sorte ait sa place)
	var kinds []int
	for nStat+nBed+nFont > 0 {
		if nFont > 0 {
			kinds, nFont = append(kinds, 2), nFont-1
		}
		if nStat > 0 {
			kinds, nStat = append(kinds, 1), nStat-1
		}
		for k := 0; k < 2 && nBed > 0; k++ {
			kinds, nBed = append(kinds, 0), nBed-1
		}
	}
	for i, q := range spots {
		if i >= len(kinds) {
			break
		}
		switch kinds[i] {
		case 0:
			drawFlowerBed(dst, q, i)
		case 1:
			drawStatue(dst, q, i)
		case 2:
			drawFountain(dst, q, 0.8, t)
		}
	}
	if level >= 4 { // le kiosque, au milieu
		ellipse(dst, c.X+20, c.Y+40, 130, 40, rgba(0x1d3a12, 0.25))
		roundRect(dst, c.X-110, c.Y-60, 220, 110, 50, rgb(0xf3ead8))
		for i := 0; i < 6; i++ {
			rect(dst, c.X-96+float64(i)*38, c.Y-60, 10, 100, rgb(0xffffff))
		}
		polygon(dst, []V{{X: c.X - 140, Y: c.Y - 50}, {X: c.X, Y: c.Y - 190}, {X: c.X, Y: c.Y - 50}}, rgb(0x46c283))
		polygon(dst, []V{{X: c.X, Y: c.Y - 190}, {X: c.X + 140, Y: c.Y - 50}, {X: c.X, Y: c.Y - 50}}, rgb(0x2fa86b))
	}
	drawLoop(dst, pts, 230, 0xd9c48f) // la berge
	drawLoop(dst, pts, 210, 0x4f8fc6)
	drawLoop(dst, pts, 180, 0x6fb7e6)
	for i := 0; i < len(pts); i += 4 { // des vaguelettes
		q := pts[(i+int(t*4))%len(pts)]
		line(dst, q.X-14, q.Y+20*math.Sin(float64(i)), q.X+14, q.Y+20*math.Sin(float64(i)), 3, rgba(0xffffff, 0.5))
	}
	s0 := rideStart(r, pts)
	// le quai
	dq, dd := along(pts, s0)
	nn := V{X: -dd.Y, Y: dd.X}
	if dq.Add(nn.Mul(140)).Dist(r.Booth().C()) > dq.Sub(nn.Mul(140)).Dist(r.Booth().C()) {
		nn = nn.Mul(-1) // (le quai, sur la berge du côté de la cabine)
	}
	pq := dq.Add(nn.Mul(140))
	polygon(dst, box(pq.Add(V{X: 6, Y: 10}), dd, 130, 34), rgba(0x1d3a12, 0.25))
	polygon(dst, box(pq, dd, 130, 34), rgb(0x8a5a2b))
	polygon(dst, box(pq, dd, 124, 28), rgb(0xb07d4a))
	for k := -2; k <= 2; k++ {
		e := pq.Add(dd.Mul(float64(k) * 50))
		line(dst, e.X-nn.X*28, e.Y-nn.Y*28, e.X+nn.X*28, e.Y+nn.Y*28, 2, rgba(0x6e4322, 0.6))
	}
	gapL := 170.0 / pathLen(pts)
	for k, b := range r.Boats {
		at, dir := along(pts, boatAt(b, s0, gapL))
		col := []uint32{0xff6fa8, 0xffd23f, 0x4fc3ff, 0x7ed957, 0xff8a3d}[k%5]
		vehicle(dst, at, dir, 70, 44, 0xffffff)
		bn := V{X: -dir.Y, Y: dir.X}
		front := at.Add(dir.Mul(70))
		polygon(dst, []V{front.Add(dir.Mul(30)), front.Add(bn.Mul(44)), front.Sub(bn.Mul(44))}, rgb(0xffffff))
		ridersIn(dst, at, dir, riderLooks(b.Riders, 0, 4), 2, 2, 50, 0, t)
		// l'auvent coloré, à l'arrière
		polygon(dst, box(at.Sub(dir.Mul(50)), dir, 16, 46), rgb(col))
	}
	for i := 0; i < 1+level/3; i++ { // les cygnes
		at, dir := along(pts, float64(i)*0.31+t*0.006)
		at = at.Add(V{X: -dir.Y, Y: dir.X}.Mul(60))
		ellipse(dst, at.X, at.Y, 26, 16, rgb(0xffffff))
		line(dst, at.X+dir.X*14, at.Y, at.X+dir.X*20, at.Y-30, 7, rgb(0xffffff))
		circle(dst, at.X+dir.X*22, at.Y-32, 7, rgb(0xffffff))
		circle(dst, at.X+dir.X*28, at.Y-31, 3, rgb(0xff8a3d))
	}
	// le boulevard passe sur la rivière : deux ponts (les bateaux passent dessous)
	g := boatGeom(r, level)
	drawBoulevardBridge(dst, g.xa)
	drawBoulevardBridge(dst, g.xb)
}

// ---------------------------------------------------------------- tour volante

// drawSwingTower : la tour de chaises volantes. Le mât monte avec la taille ;
// pendant le tour, la couronne grimpe le long du mât et les nacelles s'envolent.
// Chaque passager est assis dans un siège à dossier, attaché par une barre de
// sécurité, et le siège pend à deux chaînes.
func drawSwingTower(dst *ebiten.Image, p *sim.Place, r *sim.Ride, level int, t float64) {
	tr := r.Track()
	f := float64(level)
	base := V{X: tr.X + tr.W*0.45, Y: tr.Y + tr.H*0.78}
	if r.Top {
		base.Y = tr.Y + tr.H*0.84
	}
	h := 260 + 90*f      // hauteur du mât
	cr := 130 + 22*f     // rayon de la couronne
	up, spin := 0.0, 0.0 // montée (0 à 1), vitesse
	if r.State == 1 {
		pr := r.Prog
		up = math.Min(1, math.Min(pr/0.15, (1-pr)/0.12))
		spin = up
	}
	ang := t * 1.6 * spin
	hc := 120 + (h-200)*up // hauteur de la couronne
	// la plate-forme
	ellipse(dst, base.X+20, base.Y+16, cr+140, (cr+140)*0.38, rgba(0x1d3a12, 0.25))
	ellipse(dst, base.X, base.Y, cr+140, (cr+140)*0.38, rgb(0xb0b0a8))
	ellipse(dst, base.X, base.Y-6, cr+124, (cr+124)*0.38, rgb(0xe8dcc0))
	for i := 0; i < 16; i++ {
		a := float64(i) * math.Pi / 8
		circle(dst, base.X+math.Cos(a)*(cr+132), base.Y-6+math.Sin(a)*(cr+132)*0.38, 6, rgb([]uint32{0xff6fa8, 0xffd23f, 0x4fc3ff}[i%3]))
	}
	n := 2 * p.RideUnits(r)
	cy := base.Y - hc
	swing := cr * (0.2 + 0.8*up)
	seat := func(i int, front bool) {
		a := ang + 2*math.Pi*float64(i)/float64(n)
		if (math.Sin(a) >= 0) != front {
			return
		}
		ox, oy := math.Cos(a), math.Sin(a)*0.38
		top := V{X: base.X + ox*cr*0.9, Y: cy + oy*cr*0.9}
		s := V{X: base.X + ox*(cr+swing), Y: cy + oy*(cr+swing) + 150 - 70*up}
		col := []uint32{0xd9534f, 0xffd23f, 0x4fc3ff}[i%3]
		// les deux chaînes, accrochées aux côtés du siège
		line(dst, top.X-10, top.Y, s.X-22, s.Y-40, 2.5, rgb(0x9a9a9a))
		line(dst, top.X+10, top.Y, s.X+22, s.Y-40, 2.5, rgb(0x9a9a9a))
		// le dossier
		roundRect(dst, s.X-24, s.Y-46, 48, 40, 8, shade(col, 0.8))
		if i < len(r.Riders) {
			lk := lookFrom(r.Riders[i].Look)
			seated(dst, s.X, s.Y-4, r.Riders[i].Look, 0, t)
			// les jambes qui pendent
			sw := 5 * math.Sin(t*5+float64(i))
			line(dst, s.X-7, s.Y+4, s.X-8+sw, s.Y+36, 8, rgb(lk.pants))
			line(dst, s.X+7, s.Y+4, s.X+8-sw, s.Y+36, 8, rgb(lk.pants))
		}
		// l'assise et la barre de sécurité
		roundRect(dst, s.X-26, s.Y-8, 52, 16, 6, rgb(col))
		line(dst, s.X-24, s.Y-20, s.X+24, s.Y-20, 5, rgb(0x3b3b3b))
		line(dst, s.X-22, s.Y-40, s.X-24, s.Y-6, 3, rgb(0x7a7a7a))
		line(dst, s.X+22, s.Y-40, s.X+24, s.Y-6, 3, rgb(0x7a7a7a))
	}
	for i := 0; i < n; i++ {
		seat(i, false)
	}
	// le mât (rouge et blanc, des lumières)
	w := 30 + 2.5*f
	for y := 0.0; y < h; y += 60 {
		col := uint32(0xd9534f)
		if int(y/60)%2 == 1 {
			col = 0xf2f2f2
		}
		rect(dst, base.X-w/2, base.Y-y-60, w, 60, rgb(col))
		rect(dst, base.X-w/2, base.Y-y-60, w*0.3, 60, shade(col, 1.12))
	}
	roundRect(dst, base.X-w, base.Y-h-50, 2*w, 50, 14, rgb(0x2b2b3a))
	circle(dst, base.X, base.Y-h-62, 13, rgba(0xffd23f, 0.6+0.4*math.Sin(t*4)))
	// la couronne
	ellipse(dst, base.X, cy+8, cr+14, (cr+14)*0.38, shade(0xffd23f, 0.75))
	ellipse(dst, base.X, cy, cr+14, (cr+14)*0.38, rgb(0xffd23f))
	ellipse(dst, base.X, cy, cr-10, (cr-10)*0.38, rgb(0xd9534f))
	for i := 0; i < n; i++ {
		seat(i, true)
	}
	// des drapeaux autour (taille 5+)
	for i := 0; i < level-4; i++ {
		x := tr.X + 80 + float64(i)*(tr.W-160)/6
		y := base.Y + cr*0.38 + 190
		line(dst, x, y, x, y-130, 5, rgb(0x6e4322))
		polygon(dst, []V{{X: x, Y: y - 130}, {X: x + 54, Y: y - 115 + 6*math.Sin(t*3+float64(i))}, {X: x, Y: y - 100}}, rgb([]uint32{0xff6fa8, 0x4fc3ff, 0x7ed957}[i%3]))
	}
}

// ---------------------------------------------------------------- grand huit

// coasterPath : le circuit du grand huit en relief. Pour chaque point : sa place au
// sol (g) et sa hauteur (h). Une montée à chaîne au départ, puis des descentes et
// des bosses ; les loopings font partie du circuit (le train les emprunte).
type cPt struct {
	g    V
	h    float64
	loop bool // dans un looping (pas de pilier)
	lift bool // dans la montée (à chaîne)
}

func coasterPath(r *sim.Ride, level int) []cPt {
	ground := trackLoop(r, level, 0.38)
	n := len(ground)
	f := float64(level)
	H := 170 + 24*f // hauteur du sommet de la montée
	// le circuit en hauteur doit rester dans son terrain : on tasse le sol vers le bas
	tr := r.Track()
	bot := tr.Y + tr.H - 30
	for i, q := range ground {
		ground[i].Y = bot - (bot-q.Y)*(tr.H-60-H-140)/(tr.H-60)
	}
	s0 := rideStart(r, ground)
	i0 := int(s0 * float64(n))
	var raw []cPt
	hills := 2 + level/3
	for k := 0; k < n; k++ {
		u := float64(k) / float64(n)
		var h float64
		lift := false
		switch {
		case u < 0.04:
			h = 30
		case u < 0.22: // la montée
			h = 30 + (H-30)*(u-0.04)/0.18
			lift = true
		default:
			v := (u - 0.22) / 0.78
			env := 1 - v
			h = 30 + (H-30)*env*(0.5+0.5*math.Cos(2*math.Pi*float64(hills)*v))
		}
		raw = append(raw, cPt{g: ground[(i0+k)%n], h: h, lift: lift})
	}
	// les loopings : un tous les trois niveaux, là où le circuit va de gauche à droite
	nl := level / 3
	var out []cPt
	done := 0
	for k := 0; k < len(raw); k++ {
		u := float64(k) / float64(len(raw))
		if done < nl && u > 0.3+0.6*float64(done)/float64(max(1, nl)) && k+6 < len(raw) {
			a, b := raw[k], raw[k+6]
			d := b.g.Sub(a.g)
			if math.Abs(d.X) > 2.5*math.Abs(d.Y) && a.h < H*0.55 {
				dir := d.Norm()
				rad := 70 + 9*f
				for j := 0; j <= 28; j++ {
					th := 2 * math.Pi * float64(j) / 28
					fr := float64(j) / 28
					g := a.g.Add(d.Mul(fr)).Add(dir.Mul(rad * 0.85 * math.Sin(th)))
					out = append(out, cPt{g: g, h: a.h + (b.h-a.h)*fr + rad*(1-math.Cos(th)), loop: j > 0 && j < 28})
				}
				k += 6
				done++
				continue
			}
		}
		out = append(out, raw[k])
	}
	return out
}

// coasterTimes : le temps passé sur chaque morceau (lent dans la montée, de plus
// en plus vite en descendant). Renvoie les temps cumulés (le dernier : le total).
func coasterTimes(c []cPt) []float64 {
	top := 0.0
	for _, q := range c {
		top = math.Max(top, q.h)
	}
	ts := make([]float64, len(c)+1)
	for i := range c {
		a, b := c[i], c[(i+1)%len(c)]
		l := a.g.Dist(b.g) + math.Abs(a.h-b.h)
		v := 60.0 // montée à chaîne
		if !a.lift {
			v = 240 + 22*math.Sqrt(math.Max(0, top+40-a.h))
		}
		ts[i+1] = ts[i] + l/v
	}
	return ts
}

func cScreen(q cPt) V { return V{X: q.g.X, Y: q.g.Y - q.h} }

// drawCoaster : le grand huit. Le circuit est en hauteur, sur des piliers de
// hauteurs différentes ; une montée lente, des descentes qui vont très vite,
// des loopings (le train passe dedans). Pendant le tour, le train fait deux fois
// le circuit ; les passagers lèvent les bras.
func drawCoaster(dst *ebiten.Image, p *sim.Place, r *sim.Ride, level int, t float64) {
	c := coasterPath(r, level)
	ground := make([]V, len(c))
	for i, q := range c {
		ground[i] = q.g
	}
	decorTrees(dst, r, level/2+1, ground)
	n := len(c)
	// l'ombre au sol (décalée selon la hauteur)
	for i := range c {
		a, b := c[i], c[(i+1)%n]
		sa, sb := a.g.Add(V{X: a.h * 0.25, Y: 10}), b.g.Add(V{X: b.h * 0.25, Y: 10})
		line(dst, sa.X, sa.Y, sb.X, sb.Y, 26, rgba(0x1d3a12, 0.18))
	}
	// les piliers (plus hauts là où le circuit est haut)
	for i := 0; i < n; i += 3 {
		q := c[i]
		if q.loop || q.h < 35 {
			continue
		}
		ellipse(dst, q.g.X+14, q.g.Y+6, 22, 8, rgba(0x1d3a12, 0.25))
		rect(dst, q.g.X-9, q.g.Y-q.h, 18, q.h, rgb(0xf5a623))
		rect(dst, q.g.X-9, q.g.Y-q.h, 6, q.h, rgb(0xffc65c))
		rect(dst, q.g.X-13, q.g.Y-8, 26, 10, rgb(0xc77d12))
	}
	// les pieds des loopings
	for i := 1; i < n; i++ {
		if c[i].loop && !c[i-1].loop {
			q := c[i]
			rad := 70 + 9*float64(level)
			for _, sg := range []float64{-1, 1} {
				x := q.g.X + sg*rad*0.6
				line(dst, x, q.g.Y-q.h-rad*0.3, x, q.g.Y+4, 12, rgb(0xf5a623))
			}
		}
	}
	// les rails, en hauteur (plus clairs en haut)
	for i := range c {
		a, b := cScreen(c[i]), cScreen(c[(i+1)%n])
		line(dst, a.X, a.Y, b.X, b.Y, 26, rgb(0x23408f))
		circle(dst, a.X, a.Y, 13, rgb(0x23408f))
	}
	for i := range c {
		a, b := cScreen(c[i]), cScreen(c[(i+1)%n])
		line(dst, a.X, a.Y, b.X, b.Y, 14, shade(0x3b6fe0, 1+math.Min(0.35, c[i].h/900)))
		if c[i].lift && i%2 == 0 {
			circle(dst, a.X, a.Y, 4, rgb(0xffd23f)) // la chaîne de la montée
		}
	}
	// le train : la tête suit le temps (lente en montée, rapide en descente)
	ts := coasterTimes(c)
	total := ts[n]
	headT := 0.0
	if r.State == 1 {
		headT = math.Mod(r.Prog*2, 1) * total // deux tours pendant la durée du tour
	}
	hi := 0
	for hi < n-1 && ts[hi+1] < headT {
		hi++
	}
	fr := 0.0
	if ts[hi+1] > ts[hi] {
		fr = (headT - ts[hi]) / (ts[hi+1] - ts[hi])
	}
	// position à l'écran, puis les voitures à la queue leu leu le long des rails
	scr := make([]V, n)
	for i, q := range c {
		scr[i] = cScreen(q)
	}
	cum := make([]float64, n+1)
	for i := 0; i < n; i++ {
		cum[i+1] = cum[i] + scr[i].Dist(scr[(i+1)%n])
	}
	at := func(d float64) (V, V, float64) {
		d = math.Mod(d, cum[n])
		if d < 0 {
			d += cum[n]
		}
		k := 0
		for k < n-1 && cum[k+1] < d {
			k++
		}
		a, b := scr[k], scr[(k+1)%n]
		l := cum[k+1] - cum[k]
		w := 0.0
		if l > 0 {
			w = (d - cum[k]) / l
		}
		return a.Add(b.Sub(a).Mul(w)), b.Sub(a).Norm(), c[k].h
	}
	head := cum[hi] + fr*(cum[hi+1]-cum[hi])
	units := p.RideUnits(r)
	for k := units - 1; k >= 0; k-- {
		q, dir, h := at(head - float64(k)*112)
		vehicle(dst, q, dir, 50+math.Min(10, h/40), 34+math.Min(8, h/50), []uint32{0xff5d8f, 0xffd23f, 0x4fc3ff, 0x7ed957, 0xff8a3d}[k%5])
		arms := 0
		if r.State == 1 {
			arms = 1
		}
		ridersIn(dst, q, dir, riderLooks(r.Riders, k, 4), 2, 2, 44, arms, t)
	}
}

// ---------------------------------------------------------------- grand splash

// drawSplash : un canal en bois sur l'eau ; une chute dans les rochers tous les trois
// niveaux (et de grandes éclaboussures quand un rondin la descend) ; des sapins.
// Chaque rondin part seul, dès qu'il est plein.
func drawSplash(dst *ebiten.Image, p *sim.Place, r *sim.Ride, level int, t float64) {
	pts := trackLoop(r, level, 0.22)
	for i, q := range scatter(r, pts, 4+3*level, 17, 150) {
		drawPine(dst, q.X, q.Y, 0.9+0.4*math.Mod(float64(i)*0.41, 1))
	}
	s0 := rideStart(r, pts)
	drops := 1 + level/3
	dropAt := func(i int) float64 { return s0 + float64(i)/float64(drops+1) }
	// les chutes : de gros rochers et un grand bassin au pied
	for i := 1; i <= drops; i++ {
		top, _ := along(pts, dropAt(i)-0.035)
		bot, _ := along(pts, dropAt(i))
		ellipse(dst, bot.X, bot.Y+20, 200, 80, rgb(0x9fd3ff))
		hh := 120 + 10*float64(level)
		polygon(dst, []V{{X: top.X - 170, Y: top.Y + 50}, {X: top.X - 60, Y: top.Y - hh}, {X: top.X + 10, Y: top.Y - hh*0.6}, {X: top.X - 20, Y: top.Y + 50}}, rgb(0xa89a8a))
		polygon(dst, []V{{X: top.X - 20, Y: top.Y + 50}, {X: top.X + 10, Y: top.Y - hh*0.6}, {X: top.X + 80, Y: top.Y - hh*0.8}, {X: top.X + 170, Y: top.Y + 50}}, rgb(0x7a6c5e))
	}
	drawLoop(dst, pts, 120, 0x6e4322)
	drawLoop(dst, pts, 100, 0x8a5a2b)
	drawLoop(dst, pts, 80, 0x4fa3e6)
	for i := 0; i < len(pts); i += 5 {
		q := pts[(i+int(t*8))%len(pts)]
		line(dst, q.X-10, q.Y, q.X+10, q.Y, 3, rgba(0xffffff, 0.6))
	}
	gapL := 210.0 / pathLen(pts)
	for _, b := range r.Boats {
		s := boatAt(b, s0, gapL)
		at, dir := along(pts, s)
		// le rondin : un tronc creusé, quatre passagers l'un derrière l'autre
		vehicle(dst, at, dir, 92, 30, 0x8a5a2b)
		front := at.Add(dir.Mul(92))
		circle(dst, front.X, front.Y-2, 22, rgb(0xa0703b))
		circle(dst, front.X, front.Y-2, 13, rgb(0xc89a5a))
		ridersIn(dst, at.Add(dir.Mul(-10)), dir, riderLooks(b.Riders, 0, 4), 1, 4, 38, 0, t)
		// au pied d'une chute : ça éclabousse !
		if b.State == 2 {
			for i := 1; i <= drops; i++ {
				d := math.Mod(s-dropAt(i)+10, 1)
				if d < 0.03 {
					for j := 0; j < 12; j++ {
						a := float64(j)*math.Pi/6 + t*3
						circle(dst, at.X+math.Cos(a)*(60+d*3000), at.Y+math.Sin(a)*(26+d*1200)-40, 12, rgba(0xffffff, 0.85))
					}
				}
			}
		}
	}
}

// ---------------------------------------------------------------- les pieds dans le vide

// drawInverted : un grand huit suspendu : rails sur de grosses colonnes, des vrilles
// (une tous les deux niveaux), les passagers suspendus sous les rails, jambes dans le vide.
func drawInverted(dst *ebiten.Image, p *sim.Place, r *sim.Ride, level int, t float64) {
	pts := trackLoop(r, level, 0.32)
	decorTrees(dst, r, level/2, pts)
	const lift = 150.0   // les rails sont haut au-dessus du sol
	for i := range pts { // l'ombre du circuit au sol
		a, b := pts[i].Add(V{X: 40, Y: lift}), pts[(i+1)%len(pts)].Add(V{X: 40, Y: lift})
		line(dst, a.X, a.Y, b.X, b.Y, 30, rgba(0x1d3a12, 0.2))
	}
	for i := 0; i < len(pts); i += 5 { // les colonnes et leur bras
		q := pts[i]
		ellipse(dst, q.X+16, q.Y+lift, 30, 10, rgba(0x1d3a12, 0.25))
		rect(dst, q.X-14, q.Y-30, 28, lift+30, rgb(0x6c4bb8))
		rect(dst, q.X-14, q.Y-30, 9, lift+30, rgb(0x8f72d8))
		rect(dst, q.X-20, q.Y+lift-10, 40, 12, rgb(0x4a3290))
	}
	drawLoop(dst, pts, 30, 0x1f3a7a)
	drawLoop(dst, pts, 14, 0xffd23f)
	s0 := rideStart(r, pts)
	// les vrilles : une spirale autour des rails
	for i := 1; i <= level/2; i++ {
		s := s0 + float64(i)/float64(level/2+1)
		for k := 0; k < 30; k++ {
			at, dir := along(pts, s+float64(k)*0.002)
			n := V{X: -dir.Y, Y: dir.X}
			a := float64(k) * math.Pi / 6
			q := at.Add(n.Mul(32 * math.Sin(a)))
			circle(dst, q.X, q.Y-18*math.Cos(a), 7, rgb(0xffd23f))
		}
	}
	pos := ridePos(r, s0)
	if r.State == 1 {
		pos = s0 + math.Mod(r.Prog*1.5, 1) // il va vite : un tour et demi
	}
	units := p.RideUnits(r)
	gap := 90.0 / pathLen(pts)
	for k := 0; k < units; k++ {
		at, dir := along(pts, pos-float64(k)*gap)
		polygon(dst, box(at, dir, 36, 22), rgb(0x2b2b3a)) // le chariot accroché aux rails
		looks := riderLooks(r.Riders, k, 2)
		for i := 0; i < 2; i++ {
			sx := at.Add(dir.Mul(float64(i)*44 - 22))
			roundRect(dst, sx.X-22, sx.Y-6, 44, 52, 10, rgb(0x4fc3ff)) // le siège sous les rails
			if i < len(looks) {
				lk := lookFrom(looks[i])
				seated(dst, sx.X, sx.Y+40, looks[i], 2, t)
				sw := 7 * math.Sin(t*8+float64(k+i)) // les jambes balancent dans le vide
				line(dst, sx.X-7, sx.Y+42, sx.X-7+sw, sx.Y+86, 9, rgb(lk.pants))
				line(dst, sx.X+7, sx.Y+42, sx.X+7-sw, sx.Y+86, 9, rgb(lk.pants))
				roundRect(dst, sx.X-14+sw, sx.Y+82, 14, 8, 3, rgb(0x3a2a1a))
				roundRect(dst, sx.X-sw, sx.Y+82, 14, 8, 3, rgb(0x3a2a1a))
				line(dst, sx.X-18, sx.Y+14, sx.X+18, sx.Y+14, 6, rgb(0x3b3b3b)) // l'arceau de sécurité
			}
		}
	}
}

// boatGeom : la forme de la rivière des petits bateaux à cette taille : une boucle
// dans son terrain, et un grand méandre qui passe sous le boulevard (deux ponts),
// monte dans le terrain d'en face et redescend ; il grandit avec la taille.
type boatShape struct{ xa, xb, yTop, yb0, yBot, xl, xr float64 }

func boatGeom(r *sim.Ride, level int) boatShape {
	pl := r.Plot
	p3 := sim.ParcSlot(3)
	f := float64(level) / 10
	g := boatShape{}
	g.xa = pl.X + 1300
	g.xb = g.xa + 600 + 500*f
	g.yb0 = pl.Y + 190
	g.yTop = p3.Y + p3.H - 210 - 1000*f
	g.yBot = g.yb0 + 500 + 700*f
	g.xl = pl.X + 700
	g.xr = math.Max(g.xb+350, math.Min(pl.X+pl.W-150, g.xl+1300+900*f))
	return g
}

// boatCrossings : où la rivière passe sous le boulevard (rien si l'attraction n'est pas achetée).
func boatCrossings(p *sim.Place) []float64 {
	for _, r := range p.Parc.Rides {
		if r.Kind.ID == "bateaux" && p.RideBuilt(r) {
			g := boatGeom(r, p.RideLevel(r))
			return []float64{g.xa, g.xb}
		}
	}
	return nil
}

// boatTrack : la rivière des petits bateaux (courbes douces).
func boatTrack(r *sim.Ride, level int) []V {
	g := boatGeom(r, level)
	raw := []V{
		{X: g.xl, Y: g.yb0 + 150}, {X: g.xa - 160, Y: g.yb0}, {X: g.xa, Y: g.yb0 - 200},
		{X: g.xa, Y: g.yTop + 260}, {X: g.xa + 170, Y: g.yTop}, {X: g.xb - 170, Y: g.yTop}, {X: g.xb, Y: g.yTop + 260},
		{X: g.xb, Y: g.yb0 - 200}, {X: g.xb + 160, Y: g.yb0}, {X: g.xr, Y: g.yb0 + 150},
		{X: g.xr, Y: g.yBot - 160}, {X: g.xr - 160, Y: g.yBot}, {X: g.xl + 160, Y: g.yBot}, {X: g.xl, Y: g.yBot - 160},
	}
	return resample(smoothClosed(raw, 4), 260)
}

// drawBoatGarden : le terrain d'en face (rangée du haut), quand la rivière y monte :
// une pelouse, une haie ouverte là où passe la rivière.
func drawBoatGarden(dst *ebiten.Image, gaps []float64) {
	pl := sim.ParcSlot(3)
	roundRect(dst, pl.X, pl.Y, pl.W, pl.H, 40, rgb(0x8cc56a))
	for i := 0; i < 7; i++ {
		ellipse(dst, pl.X+240+(pl.W-480)*hash2(i, 3, 1), pl.Y+160+(pl.H-320)*hash2(i, 3, 2), 220, 140, rgba(0x9fd47a, 0.5))
	}
	bush := func(x, y float64, i int) {
		k := 0.8 + 0.4*hash2(i, 3, 9)
		circle(dst, x+6, y+8, 26*k, rgba(0x1d3a12, 0.25))
		circle(dst, x, y, 26*k, rgb(0x3f8a3a))
		circle(dst, x-7*k, y-7*k, 15*k, rgb(0x5cae4c))
	}
	i := 0
	for x := pl.X + 30; x <= pl.X+pl.W-30; x += 44 {
		bush(x, pl.Y, i)
		open := false
		for _, gx := range gaps {
			open = open || math.Abs(x-gx) < 170
		}
		if !open {
			bush(x, pl.Y+pl.H, i+1)
		}
		i += 2
	}
	for y := pl.Y + 30; y <= pl.Y+pl.H-30; y += 44 {
		bush(pl.X, y, i)
		bush(pl.X+pl.W, y, i+1)
		i += 2
	}
}

// drawBoulevardBridge : le boulevard passe sur la rivière : un pont de bois, ses garde-corps.
func drawBoulevardBridge(dst *ebiten.Image, x float64) {
	bt := sim.ParcBoulY - sim.ParcBoulH/2
	bb := sim.ParcBoulY + sim.ParcBoulH/2
	rect(dst, x-170, bt+14, 340, bb-bt-28+18, rgba(0x1d3a12, 0.25))
	rect(dst, x-160, bt, 320, bb-bt, rgb(0x9a6a3e))
	for xx := x - 156; xx < x+156; xx += 26 {
		rect(dst, xx, bt, 22, bb-bt, rgb(0xb98a56))
	}
	for _, y := range []float64{bt + 6, bb - 10} { // garde-corps
		rect(dst, x-170, y-4, 340, 10, rgb(0x6e4322))
		for xx := x - 160; xx <= x+160; xx += 40 {
			rect(dst, xx-4, y-14, 8, 20, rgb(0x5a3a22))
		}
	}
}

// drawFlowerBed : un parterre de fleurs (bordure de buis, fleurs en rangs).
func drawFlowerBed(dst *ebiten.Image, q V, i int) {
	cols := [][2]uint32{{0xff5d8f, 0xffd23f}, {0xb36bd6, 0xffffff}, {0xff8a3d, 0xffd23f}, {0x4fc3ff, 0xffffff}}[i%4]
	ellipse(dst, q.X+8, q.Y+10, 100, 46, rgba(0x1d3a12, 0.2))
	ellipse(dst, q.X, q.Y, 100, 46, rgb(0x3f8a3a))
	ellipse(dst, q.X, q.Y-4, 88, 38, rgb(0x6e4a2a))
	for k := 0; k < 26; k++ {
		a := float64(k) * 2.4
		rr := 0.25 + 0.7*math.Mod(float64(k)*0.37, 1)
		x, y := q.X+math.Cos(a)*80*rr, q.Y-4+math.Sin(a)*32*rr
		circle(dst, x, y-3, 7, rgb(0x4f9e48))
		circle(dst, x, y-7, 6, rgb(cols[k%2]))
		circle(dst, x, y-7, 2.2, rgb(0xffe36a))
	}
}

// drawStatue : une statue de pierre sur son socle (un personnage, ou un vase).
func drawStatue(dst *ebiten.Image, q V, i int) {
	ellipse(dst, q.X+20, q.Y+6, 56, 14, rgba(0x1d3a12, 0.25))
	rect(dst, q.X-36, q.Y-50, 72, 52, rgb(0xb8b4aa))
	rect(dst, q.X-36, q.Y-50, 20, 52, rgb(0xd6d2c8))
	rect(dst, q.X-42, q.Y-58, 84, 12, rgb(0xc8c4ba))
	st, li := rgb(0xc9c6bd), rgb(0xe4e1d9)
	if i%2 == 0 { // un personnage, un bras levé
		roundRect(dst, q.X-16, q.Y-130, 32, 74, 12, st)
		rect(dst, q.X-16, q.Y-130, 10, 70, li)
		circle(dst, q.X, q.Y-146, 16, st)
		circle(dst, q.X-5, q.Y-150, 6, li)
		line(dst, q.X+14, q.Y-120, q.X+34, q.Y-170, 9, st)
		line(dst, q.X-14, q.Y-120, q.X-24, q.Y-80, 9, st)
	} else { // un grand vase
		ellipse(dst, q.X, q.Y-96, 34, 40, st)
		ellipse(dst, q.X-10, q.Y-104, 12, 22, li)
		rect(dst, q.X-14, q.Y-70, 28, 14, st)
		ellipse(dst, q.X, q.Y-134, 30, 8, li)
		for k := 0; k < 5; k++ {
			circle(dst, q.X-20+float64(k)*10, q.Y-140-float64(k%2)*8, 8, rgb(0x5fae4c))
		}
	}
}

// drawFountain : une fontaine (bassin de pierre, jets d'eau qui bougent).
func drawFountain(dst *ebiten.Image, fo V, k, t float64) {
	ellipse(dst, fo.X+12, fo.Y+12, 120*k, 50*k, rgba(0x1d3a12, 0.2))
	ellipse(dst, fo.X, fo.Y, 120*k, 50*k, rgb(0xb0b0a8))
	ellipse(dst, fo.X, fo.Y-4, 102*k, 40*k, rgb(0x6fb7e6))
	rect(dst, fo.X-10*k, fo.Y-60*k, 20*k, 56*k, rgb(0xc8c4ba))
	ellipse(dst, fo.X, fo.Y-60*k, 34*k, 12*k, rgb(0xb0b0a8))
	for j := 0; j < 8; j++ {
		a := float64(j) * math.Pi / 4
		h := (40 + 10*math.Sin(t*5+float64(j))) * k
		line(dst, fo.X, fo.Y-62*k, fo.X+math.Cos(a)*60*k, fo.Y-62*k-h*0.4+math.Sin(a)*20*k+h*0.6, 4, rgba(0xffffff, 0.8))
	}
	for j := 0; j < 5; j++ {
		ph := math.Mod(t*1.2+float64(j)*0.2, 1)
		circle(dst, fo.X+math.Cos(float64(j)*1.3)*70*k*ph, fo.Y-4+math.Sin(float64(j)*1.3)*28*k*ph, 4*k, rgba(0xffffff, 0.6*(1-ph)))
	}
}
