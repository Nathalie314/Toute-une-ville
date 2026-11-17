package main

import (
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// La ferme exotique : le sol, les arbres (ananas, bananier, cocotier, manguier),
// les plantes des champs (riz, vanille, thé, cannelle) et les icônes des produits.

func init() {
	cropCol[sim.Ananas] = 0xf2b632
	cropCol[sim.Banane] = 0xf6d53b
	cropCol[sim.NoixCoco] = 0x7a4a26
	cropCol[sim.Mangue] = 0xf08a24
}

// exoticGround : sable et herbe tropicale, chemins, boutique, route.
func (g *game) exoticGround(dst *ebiten.Image, p *sim.Place, tl, br V) {
	g.exoticWild(dst, p, tl, br) // la cascade, la rivière, la forêt
	// sol : herbe plus vive, sable sous les chemins
	path := rgb(0xe9d7a8)
	roundRect(dst, 380, 470, p.W-520, 110, 50, path)
	roundRect(dst, 60, 880, p.W-120, 150, 60, path)
	roundRect(dst, 1180, 470, 140, 560, 60, path)
	// les espaces : leur nom au-dessus des bâtiments
	for _, z := range sim.TeamZones(p.ID) {
		a := 0.85
		if !p.ZoneOpen(z.ID) {
			a = 0.35
		}
		cx := (z.X0 + z.X1) / 2
		// (au milieu de ses champs, juste sous leur nom)
		lo, hi := math.Inf(1), math.Inf(-1)
		for _, s := range p.Stations {
			if s.Team == z.ID && s.Kind == sim.Field {
				lo, hi = math.Min(lo, s.Body.X), math.Max(hi, s.Body.X+s.Body.W)
			}
		}
		if lo < hi {
			cx = (lo + hi) / 2
		}
		roundRect(dst, cx-90, 486, 180, 30, 10, rgba(0x1f6e4a, a))
		say(dst, strings.ToUpper(z.Name), cx, 501, 17, true, 0, rgba(0xffffff, a))
	}
	// la boutique : un plancher de bois clair (une paillote)
	roundRect(dst, 110, 1030, p.W-230, 480, 30, rgb(0xc9a46a))
	roundRect(dst, 120, 1040, p.W-250, 460, 26, rgb(0xf1e2bd))
	for x := 120.0; x < p.W-140; x += 70 {
		line(dst, x, 1042, x, 1498, 1.5, rgba(0xc9a46a, 0.45))
	}
	road(dst, tl, br, 1640)
	rect(dst, 1180, 1500, 140, 140, rgb(0xe9d7a8))
	// clôture en bambou
	fence := func(x1, x2 float64) {
		for x := x1; x <= x2; x += 30 {
			rect(dst, x-4, 1530, 8, 42, rgb(0xb5a24a))
			rect(dst, x-4, 1545, 8, 3, rgb(0x8a7a30))
		}
		rect(dst, x1, 1542, x2-x1, 6, rgb(0x9a8a3a))
	}
	fence(60, 1170)
	fence(1330, p.W-60)
	roundRect(dst, 1110, 1475, 280, 46, 8, rgb(0x1f8a5a))
	say(dst, "FERME EXOTIQUE", 1250, 1498, 24, true, 0, rgb(0xfff3c4))
	// quelques palmiers de décor, en haut à gauche
	for k := 0; k < 3; k++ {
		x := 120 + float64(k)*90
		scaled(x, 420, 0.9, func() { drawPalm(dst, x, 420, 1, 0, g.t+float64(k)) })
	}
}

// drawPalm : un cocotier (tronc courbé, palmes, noix de coco de taille fr).
func drawPalm(dst *ebiten.Image, x, y, size, fr float64, t float64) {
	sway := math.Sin(t*0.8) * 3 * size
	h := 80 * size
	ellipse(dst, x, y+2, 26*size, 7, rgba(0x000000, 0.2))
	// tronc en anneaux, un peu courbé
	for k := 0; k < 8; k++ {
		f := float64(k) / 8
		cx := x + math.Sin(f*1.4)*10*size + sway*f
		circle(dst, cx, y-f*h, 6*size+1, rgb(0x9a7448))
		line(dst, cx-5*size, y-f*h, cx+5*size, y-f*h, 1.2, rgb(0x7a5a33))
	}
	top := V{X: x + math.Sin(1.4)*10*size + sway, Y: y - h}
	// palmes
	for k := 0; k < 6; k++ {
		a := -math.Pi/2 + (float64(k)-2.5)*0.55
		ex := top.X + math.Cos(a)*44*size
		ey := top.Y + math.Sin(a)*22*size + 18*size
		line(dst, top.X, top.Y, (top.X+ex)/2, top.Y+math.Sin(a)*16*size-6*size, 5*size+1, rgb(0x3f9b35))
		line(dst, (top.X+ex)/2, top.Y+math.Sin(a)*16*size-6*size, ex, ey, 4*size+1, rgb(0x4aa843))
	}
	if fr > 0 {
		c := rgb(cropCol[sim.NoixCoco])
		if fr < 5 {
			c = rgb(0x8fcf4a)
		}
		for k := -1; k <= 1; k++ {
			circle(dst, top.X+float64(k)*fr*1.6, top.Y+fr*1.3, fr, c)
		}
	}
}

// drawExoticTree : une case de la ferme exotique (comme drawOrchard au verger).
// size : taille de la plante (0,4 petit, 0,7 moyen, 1 grand) ; fr : taille des
// fruits (0 : pas encore) ; ripe : de la bonne couleur.
func drawExoticTree(dst *ebiten.Image, it sim.Item, x, y, size, fr float64, ripe bool, t float64) {
	fc := rgb(cropCol[it])
	if !ripe {
		fc = rgb(0x8fcf4a)
	}
	switch it {
	case sim.Ananas:
		// une grosse touffe de feuilles au sol, l'ananas pousse au milieu
		ellipse(dst, x, y+2, 30*size, 8, rgba(0x000000, 0.2))
		for k := 0; k < 7; k++ {
			a := math.Pi + float64(k)*math.Pi/6
			line(dst, x, y-4, x+math.Cos(a)*34*size, y-4+math.Sin(a)*26*size, 6*size+1, rgb(0x3f8f5a))
		}
		if fr > 0 {
			r := fr * 1.6
			ellipse(dst, x, y-10-r, r*0.8, r*1.15, fc)
			for k := -1; k <= 1; k++ {
				line(dst, x, y-10-r*2.1, x+float64(k)*r*0.5, y-10-r*2.8, 2.5, rgb(0x3f8f3a))
			}
			if ripe {
				for k := 0; k < 3; k++ {
					line(dst, x-r*0.6, y-10-r*1.6+float64(k)*r*0.6, x+r*0.6, y-10-r*1.2+float64(k)*r*0.6, 1.2, rgb(0xb5781a))
				}
			}
		}
	case sim.Banane:
		// un bananier : tronc lisse, grandes feuilles, un régime de bananes
		h := 56 * size
		ellipse(dst, x, y+2, 26*size, 7, rgba(0x000000, 0.2))
		rect(dst, x-5*size-1, y-h, 10*size+2, h, rgb(0x8a9a4a))
		for k := 0; k < 5; k++ {
			a := -math.Pi/2 + (float64(k)-2)*0.6
			ex, ey := x+math.Cos(a)*40*size, y-h+math.Sin(a)*30*size+14*size
			line(dst, x, y-h, ex, ey, 9*size+1, rgb(0x4aa843))
		}
		if fr > 0 {
			for k := 0; k < 4; k++ {
				bx := x + 6*size + float64(k%2)*fr*1.2
				by := y - h + 10 + float64(k/2)*fr*1.3
				roundRect(dst, bx-fr*0.4, by-fr, fr*0.8, fr*2.2, fr*0.4, fc)
			}
		}
	case sim.NoixCoco:
		drawPalm(dst, x, y, size, fr, t)
	default: // manguier
		th := 44 * size
		ellipse(dst, x, y+2, 30*size, 8, rgba(0x000000, 0.2))
		rect(dst, x-4*size-1, y-th, 8*size+2, th, rgb(0x7a5a33))
		cr := 30 * size
		circle(dst, x, y-th-cr*0.4, cr, rgb(0x2f8a3a))
		circle(dst, x-cr*0.6, y-th-cr*0.1, cr*0.7, rgb(0x267a32))
		circle(dst, x+cr*0.6, y-th-cr*0.05, cr*0.65, rgb(0x3a9a45))
		if fr > 0 {
			for k := 0; k < 4; k++ {
				fx := x - cr*0.5 + float64(k)*cr*0.33
				fy := y - th - cr*0.05 - float64((k*5)%3)*cr*0.28
				line(dst, fx, fy-fr*1.4, fx, fy-fr*0.6, 1.2, rgb(0x5a4a20))
				ellipse(dst, fx, fy, fr*0.75, fr, fc)
			}
		}
	}
}

// drawExoticPlant : une pousse des champs de la ferme exotique (faux si ce n'en est pas une).
func drawExoticPlant(dst *ebiten.Image, it sim.Item, x, y, gr, t float64) bool {
	sway := 0.0
	ripe := gr >= 1
	if ripe {
		sway = math.Sin(t*2) * 2
	}
	h := 8 + 40*gr
	switch it {
	case sim.Riz:
		// des touffes vertes qui jaunissent, les épis penchés
		c := lerpCol(0x6fc04a, 0xe2c650, gr)
		for k := -2; k <= 2; k++ {
			tx := x + float64(k)*5 + sway
			line(dst, x+float64(k)*2, y, tx, y-h*0.9, 2, c)
			if gr > 0.6 {
				line(dst, tx, y-h*0.9, tx+6, y-h*0.75, 3, lerpCol(0x9ccf5a, 0xf0d060, gr))
			}
		}
	case sim.Vanille:
		// une liane sur un tuteur, des gousses vertes puis brunes
		line(dst, x, y, x, y-h-14, 3, rgb(0x9a6b3e))
		for k := 0; k < 4; k++ {
			ellipse(dst, x+float64(k%2*2-1)*8, y-h*0.25-float64(k)*h*0.22, 8, 4, rgb(0x3f9b35))
		}
		if gr > 0.5 {
			c := rgb(0x7fb04a)
			if ripe {
				c = rgb(0x4a2a12)
			}
			for k := 0; k < 3; k++ {
				line(dst, x+6+float64(k)*3, y-h*0.7, x+8+float64(k)*3, y-h*0.3, 2.5, c)
			}
		}
	case sim.FeuillesThe:
		// un théier : un buisson taillé bien rond
		r := 6 + 12*gr
		circle(dst, x, y-r*0.7, r, rgb(0x2f7a35))
		circle(dst, x-r*0.5, y-r*0.5, r*0.6, rgb(0x3a8f3e))
		if ripe {
			for k := 0; k < 4; k++ {
				circle(dst, x-r*0.6+float64(k)*r*0.4, y-r*1.4, 3, rgb(0x8fdc5a)) // jeunes pousses à cueillir
			}
		}
	case sim.Cannelle:
		// un cannellier : tronc à écorce rousse
		rect(dst, x-3, y-h*0.8, 6, h*0.8, rgb(0xb5651d))
		cr := 7 + 11*gr
		circle(dst, x+sway*0.3, y-h*0.9, cr, rgb(0x4a8f3a))
		circle(dst, x-cr*0.5, y-h*0.75, cr*0.6, rgb(0x3f7a32))
		if ripe {
			line(dst, x-10, y-4, x+10, y-8, 4, rgb(0xa0521d)) // un bâton prêt
		}
	default:
		return false
	}
	if ripe {
		a := math.Max(0, math.Sin(t*3))
		circle(dst, x+12, y-h-6, 3*a, rgba(0xffffff, 0.8*a))
	}
	return true
}

// drawItemExotic : les icônes des produits exotiques (faux si ce n'en est pas un).
func drawItemExotic(dst *ebiten.Image, it sim.Item, x, y, r float64) bool {
	switch it {
	case sim.Ananas:
		ellipse(dst, x, y+r*0.25, r*0.55, r*0.75, rgb(0xf2b632))
		for k := 0; k < 3; k++ {
			line(dst, x-r*0.45, y-r*0.1+float64(k)*r*0.35, x+r*0.45, y+r*0.15+float64(k)*r*0.35, r*0.06, rgb(0xb5781a))
		}
		for k := -1; k <= 1; k++ {
			line(dst, x, y-r*0.45, x+float64(k)*r*0.35, y-r, r*0.14, rgb(0x3f8f3a))
		}
	case sim.Banane:
		for k := 0; k < 3; k++ {
			dx := float64(k-1) * r * 0.32
			polygon(dst, []V{{X: x + dx - r*0.6, Y: y - r*0.5}, {X: x + dx - r*0.4, Y: y - r*0.6}, {X: x + dx + r*0.2, Y: y + r*0.5}, {X: x + dx + r*0.6, Y: y + r*0.6}, {X: x + dx, Y: y + r*0.75}}, rgb(0xf6d53b))
		}
		circle(dst, x-r*0.55, y-r*0.6, r*0.12, rgb(0x6a4a20))
	case sim.NoixCoco:
		circle(dst, x, y, r*0.8, rgb(0x7a4a26))
		circle(dst, x+r*0.25, y-r*0.15, r*0.45, rgb(0xf6f0e0))
		circle(dst, x+r*0.25, y-r*0.15, r*0.3, rgb(0xffffff))
	case sim.Mangue:
		ellipse(dst, x, y, r*0.65, r*0.85, rgb(0xf08a24))
		ellipse(dst, x+r*0.15, y-r*0.3, r*0.35, r*0.4, rgb(0xe8452b))
		line(dst, x, y-r*0.85, x+r*0.2, y-r, r*0.1, rgb(0x5a4a20))
	case sim.Riz:
		// un sachet de riz
		roundRect(dst, x-r*0.6, y-r*0.75, r*1.2, r*1.6, r*0.15, rgb(0xf6f0e0))
		rect(dst, x-r*0.6, y-r*0.75, r*1.2, r*0.35, rgb(0xd9342b))
		for k := 0; k < 5; k++ {
			ellipse(dst, x-r*0.3+float64(k%3)*r*0.3, y+r*0.1+float64(k/3)*r*0.3, r*0.1, r*0.05, rgb(0xd8d0b8))
		}
	case sim.Vanille:
		for k := -1; k <= 1; k++ {
			line(dst, x+float64(k)*r*0.25-r*0.2, y+r*0.8, x+float64(k)*r*0.25+r*0.2, y-r*0.8, r*0.18, rgb(0x3a2412))
		}
		ellipse(dst, x+r*0.5, y-r*0.6, r*0.25, r*0.15, rgb(0xf6efc8)) // une fleur
	case sim.FeuillesThe:
		for k := -1; k <= 1; k++ {
			ellipse(dst, x+float64(k)*r*0.4, y-math.Abs(float64(k))*r*0.2, r*0.3, r*0.6, rgb(0x3f9b35))
		}
	case sim.The:
		// une boîte de thé
		roundRect(dst, x-r*0.6, y-r*0.7, r*1.2, r*1.5, r*0.15, rgb(0x2f7a35))
		rect(dst, x-r*0.6, y-r*0.7, r*1.2, r*0.3, rgb(0xd4a017))
		ellipse(dst, x, y+r*0.15, r*0.25, r*0.4, rgb(0x8fdc5a))
	case sim.Cannelle:
		for k := -1; k <= 1; k++ {
			roundRect(dst, x+float64(k)*r*0.32-r*0.13, y-r*0.85, r*0.26, r*1.7, r*0.12, rgb(0xa0521d))
			line(dst, x+float64(k)*r*0.32, y-r*0.8, x+float64(k)*r*0.32, y+r*0.8, r*0.05, rgb(0x6e3a14))
		}
	default:
		return false
	}
	return true
}

// ---------------------------------------------------------------- le nord sauvage
// Au nord : une cascade en haut à gauche (au-dessus des arbres fruitiers), puis
// une rivière rapide vers la droite, avec des rochers ; à l'ouest, au nord et à
// l'est, la forêt amazonienne marque la fin de la ferme.

const (
	exoRiverY  = -540.0 // haut de la rivière
	exoRiverH  = 115.0
	exoFallX   = 420.0 // la cascade
	exoFallW   = 190.0
	exoForestY = -720.0 // lisière nord (la forêt est au-dessus)
)

func (g *game) exoticWild(dst *ebiten.Image, p *sim.Place, tl, br V) {
	t := g.t
	x0, x1 := p.MinX()-200, p.MaxX()+200
	// --- la forêt du nord (fond sombre, puis les arbres)
	rect(dst, x0, sim.ExoDecoTop-400, x1-x0, exoForestY-sim.ExoDecoTop+400, rgb(0x1e5a2c))
	// --- la falaise de la cascade
	fx := exoFallX
	polygon(dst, []V{{fx - 170, exoRiverY + 30}, {fx - 120, exoForestY - 120}, {fx + exoFallW + 120, exoForestY - 140}, {fx + exoFallW + 190, exoRiverY + 30}}, rgb(0x6f6a62))
	for k := 0; k < 9; k++ { // des blocs de roche
		bx := fx - 120 + float64(k%3)*(exoFallW+200)/2.4 + math.Mod(float64(k)*37, 40)
		by := exoForestY - 90 + float64(k/3)*70
		ellipse(dst, bx, by, 60, 34, rgb(0x857f75))
		ellipse(dst, bx-10, by-8, 36, 16, rgb(0x9a948a))
	}
	// --- la rivière (berges, eau, courant)
	ry, rh := exoRiverY, exoRiverH
	rect(dst, x0, ry-16, x1-x0, rh+32, rgb(0x8a6f4a)) // berges de terre
	rect(dst, x0, ry-6, x1-x0, rh+12, rgb(0x3f8fb0))
	rect(dst, x0, ry+10, x1-x0, rh-20, rgb(0x4aa3c8))
	vx0, vx1 := math.Max(x0, tl.X-200), math.Min(x1, br.X+200)
	for i := 0; i < 260; i++ { // le courant : des traits blancs qui filent vers la droite
		fi := float64(i)
		span := x1 - x0
		x := x0 + math.Mod(fi*977.3+t*(260+math.Mod(fi*31, 120)), span)
		if x < vx0 || x > vx1 {
			continue
		}
		y := ry + 14 + math.Mod(fi*53.7, rh-28)
		l := 30 + math.Mod(fi*17, 40)
		line(dst, x, y, x+l, y, 3, rgba(0xe8f6fb, 0.55))
	}
	// des rochers de temps en temps, l'écume autour
	for x := 900.0; x < x1; x += 520 {
		k := math.Mod(x*0.37, 3)
		y := ry + 30 + k*25
		if x < vx0-100 || x > vx1+100 {
			continue
		}
		ellipse(dst, x-34, y+4, 30, 12, rgba(0xffffff, 0.5+0.2*math.Sin(t*6+x)))
		ellipse(dst, x, y+6, 34+k*6, 18, rgb(0x5f5a52))
		ellipse(dst, x-4, y, 30+k*6, 16, rgb(0x7d776d))
		ellipse(dst, x-10, y-6, 14, 6, rgb(0x9d978d))
	}
	// --- la cascade : l'eau tombe de la falaise dans un bassin d'écume
	top := exoForestY - 110
	rect(dst, fx, top, exoFallW, ry+20-top, rgb(0x5fb3d6))
	for c := 0; c < 14; c++ {
		cx := fx + 8 + float64(c)*(exoFallW-16)/13
		for k := 0; k < 6; k++ {
			y := top + math.Mod(float64(k)*70+float64(c)*23+t*420, ry+20-top)
			line(dst, cx, y, cx, y+34, 4, rgba(0xeaf7fc, 0.75))
		}
	}
	line(dst, fx-2, top, fx+exoFallW+2, top, 8, rgb(0x8fd0ea)) // le bord
	for k := 0; k < 12; k++ {                                  // l'écume en bas
		a := t*3 + float64(k)
		ex := fx - 20 + float64(k)*(exoFallW+40)/11
		circle(dst, ex, ry+18+math.Sin(a)*5, 22+math.Sin(a*1.3)*5, rgba(0xffffff, 0.8))
	}
	ellipse(dst, fx+exoFallW/2, ry+10, exoFallW*0.75, 20, rgba(0xffffff, 0.25+0.1*math.Sin(t*2))) // la brume
	// --- les arbres : lisière nord, et les côtés ouest et est
	for x := x0; x < x1; x += 70 {
		k := math.Mod(x*0.913, 1)
		if x > fx-170 && x < fx+exoFallW+170 { // (la falaise de la cascade)
			continue
		}
		jungleTree(dst, x+k*30, exoForestY-30-k*60, 0.9+0.5*k, t, x)
	}
	for x := x0; x < x1; x += 110 { // un second rang, plus haut
		k := math.Mod(x*0.577, 1)
		jungleTree(dst, x+k*40, exoForestY-180-k*60, 1.1+0.4*k, t, x+7)
	}
	for _, side := range []float64{p.MinX() + 30, p.MaxX() - 20} {
		for y := exoForestY + 60; y < 1560; y += 95 {
			k := math.Mod(y*0.731+side, 1)
			jungleTree(dst, side+(k-0.5)*60, y, 0.9+0.4*k, t, y+side)
		}
	}
	// des fougères au bord de la rivière
	for x := x0; x < x1; x += 130 {
		k := math.Mod(x*0.29, 1)
		if x < vx0-100 || x > vx1+100 {
			continue
		}
		fern(dst, x+k*60, ry+rh+40+k*30, 0.8+0.4*k)
	}
}

// jungleTree : un grand arbre de la forêt tropicale (tronc clair, houppier en
// grosses boules de vert, parfois une liane ou des fleurs).
func jungleTree(dst *ebiten.Image, x, y, s, t, seed float64) {
	k := math.Mod(math.Abs(seed)*0.618, 1)
	h := (120 + 60*k) * s
	ellipse(dst, x, y+4, 50*s, 12*s, rgba(0x000000, 0.2))
	rect(dst, x-8*s, y-h, 16*s, h, rgb(0x8a7458))
	line(dst, x-8*s, y, x-22*s, y+6, 5*s, rgb(0x7a6448)) // racines
	line(dst, x+8*s, y, x+22*s, y+6, 5*s, rgb(0x7a6448))
	greens := []uint32{0x1f6b2e, 0x2a8a3a, 0x358f3f, 0x24753a}
	sway := math.Sin(t*0.5+seed) * 2
	for i := 0; i < 6; i++ {
		fi := float64(i)
		a := fi*1.05 + k*3
		cx := x + math.Cos(a)*34*s + sway
		cy := y - h - 10*s + math.Sin(a)*18*s
		circle(dst, cx, cy, (30+math.Mod(fi*7+k*20, 12))*s, rgb(greens[(i+int(k*4))%len(greens)]))
	}
	circle(dst, x+sway, y-h-30*s, 32*s, rgb(0x3a9a46))
	circle(dst, x-10*s+sway, y-h-40*s, 14*s, rgb(0x56b55a)) // la lumière sur le dessus
	if k < 0.3 {                                            // une liane
		for j := 0; j < 5; j++ {
			fj := float64(j)
			line(dst, x+20*s+math.Sin(fj+t)*2, y-h+fj*16*s, x+22*s+math.Sin(fj+1+t)*2, y-h+(fj+1)*16*s, 2.5, rgb(0x2f7a2a))
		}
	} else if k > 0.85 { // des fleurs
		for j := 0; j < 3; j++ {
			circle(dst, x+(float64(j)-1)*22*s, y-h-14*s+float64(j%2)*12*s, 5*s, rgb(0xff5a8a))
		}
	}
}

// fern : une fougère.
func fern(dst *ebiten.Image, x, y, s float64) {
	for i := 0; i < 7; i++ {
		a := -math.Pi + 0.2 + float64(i)*(math.Pi-0.4)/6
		line(dst, x, y, x+math.Cos(a)*34*s, y+math.Sin(a)*26*s, 4*s, rgb(0x3d9a3a))
	}
}

// drawRiverDock : un quai de la ferme exotique, au bord de la rivière : un ponton de
// bois ; un bateau à moteur y accoste, se charge, puis part vers l'aval (à droite) ;
// le suivant arrive de l'amont (à gauche).
func (g *game) drawRiverDock(dst *ebiten.Image, s *sim.Station) {
	b := s.Body
	// le ponton (planches) et ses pilotis
	roundRect(dst, b.X, b.Y-30, b.W, b.H+30, 6, rgb(0x7a5a36))
	for x := b.X + 6; x < b.X+b.W-6; x += 22 {
		rect(dst, x, b.Y-28, 18, b.H+26, rgb(0xa47a4a))
	}
	k := 0
	for it := sim.Item(0); it < sim.NItems; it++ {
		for n := 0; n < s.Buf[it] && k < 24; n++ {
			drawItem(dst, it, b.X+16+float64(k%12)*17, b.Y+14-float64(k/12)*12, 11)
			k++
		}
	}
	left := g.sim.TruckLeft(s)
	every := g.sim.DockEvery(s)
	e := every - left // temps depuis le dernier départ
	x0, y0 := b.C().X, exoRiverY+exoRiverH/2
	if e < 2.5 { // le bateau qui vient de partir descend la rivière
		f := e / 2.5
		drawCargoBoat(dst, x0+2600*f*f, y0, true, g.t)
	}
	switch {
	case e >= 4.5:
		drawCargoBoat(dst, x0, y0, left < 2, g.t)
		sayOutlined(dst, "Départ dans "+itoa(int(left))+" s", x0, y0-90, 16, 0, rgb(0xffffff), rgb(0x2b2b2b))
	case e >= 1.5: // le suivant remonte vers le quai
		f := (e - 1.5) / 3
		drawCargoBoat(dst, x0-2600*(1-f)*(1-f), y0, true, g.t)
	}
}

// drawCargoBoat : un petit bateau à moteur chargé de caisses (vu de côté, sur l'eau).
func drawCargoBoat(dst *ebiten.Image, x, y float64, busy bool, t float64) {
	bob := 2 * math.Sin(t*3+x/90)
	y += bob
	ellipse(dst, x+10, y+26, 130, 12, rgba(0x1f5f80, 0.45)) // reflet
	polygon(dst, []V{{X: x - 120, Y: y - 6}, {X: x + 130, Y: y - 6}, {X: x + 100, Y: y + 26}, {X: x - 100, Y: y + 26}}, rgb(0x1f8a5a))
	rect(dst, x-118, y-6, 246, 8, rgb(0xfff3c4))
	// la cabine, à l'arrière
	roundRect(dst, x-100, y-56, 64, 50, 8, rgb(0xf3ead8))
	rect(dst, x-92, y-48, 48, 18, rgb(0x9fd3ff))
	polygon(dst, []V{{X: x - 106, Y: y - 56}, {X: x - 30, Y: y - 56}, {X: x - 36, Y: y - 66}, {X: x - 100, Y: y - 66}}, rgb(0xc0623a))
	// les caisses
	for i := 0; i < 3; i++ {
		cx := x - 20 + float64(i)*42
		rect(dst, cx, y-36, 36, 30, rgb(0xb5835a))
		rect(dst, cx, y-36, 36, 6, rgb(0xd09a6a))
		line(dst, cx, y-36, cx+36, y-6, 2, rgba(0x6e4322, 0.6))
	}
	say(dst, "EXOTIQUE", x+10, y+12, 14, true, 0, rgb(0xffffff))
	if busy { // l'écume du moteur
		for k := 0; k < 4; k++ {
			f := math.Mod(t*2+float64(k)/4, 1)
			circle(dst, x-125-f*40, y+12+math.Sin(float64(k))*6, 6*(1-f)+2, rgba(0xffffff, 0.8*(1-f)))
		}
	}
}

// ---------------------------------------------------------------- un pays chaud

// drawTropicalPalm : un grand palmier (décor), tronc courbé, palmes, parfois des noix.
func drawTropicalPalm(dst *ebiten.Image, x, y, z float64, seed int) {
	if cam.iso {
		upright(x, y, func() { drawTropicalPalm(dst, x, y, z, seed) })
		return
	}
	dropShadow(dst, x, y, 40*z, 14*z, 120*z)
	lean := float64(seed%7-3) * 6 * z
	h := (150 + float64(seed%40)) * z
	for k := 0; k < 12; k++ { // le tronc, en anneaux
		f := float64(k) / 12
		cx := x + lean*f*f
		circle(dst, cx, y-f*h, 9*z, rgb(0x9a7448))
		line(dst, cx-8*z, y-f*h, cx+8*z, y-f*h, 2, rgb(0x7a5a33))
	}
	top := V{X: x + lean, Y: y - h}
	greens := []uint32{0x2f8a35, 0x3fa040, 0x4fb84a}
	for k := 0; k < 8; k++ { // les palmes, qui retombent
		a := float64(k)*math.Pi/4 + float64(seed%5)*0.2
		ex, ey := top.X+math.Cos(a)*80*z, top.Y+math.Sin(a)*34*z+28*z
		mx, my := (top.X+ex)/2, (top.Y+ey)/2-18*z
		line(dst, top.X, top.Y, mx, my, 9*z, rgb(greens[k%3]))
		line(dst, mx, my, ex, ey, 7*z, rgb(greens[(k+1)%3]))
	}
	if seed%3 == 0 { // des noix de coco
		for k := -1; k <= 1; k++ {
			circle(dst, top.X+float64(k)*9*z, top.Y+10*z, 6*z, rgb(0x6e4a22))
		}
	}
}

// drawBananaPlant : un bananier (grandes feuilles qui retombent).
func drawBananaPlant(dst *ebiten.Image, x, y, z float64, seed int) {
	if cam.iso {
		upright(x, y, func() { drawBananaPlant(dst, x, y, z, seed) })
		return
	}
	dropShadow(dst, x, y, 36*z, 12*z, 70*z)
	rect(dst, x-7*z, y-70*z, 14*z, 70*z, rgb(0x7a9a3a))
	for k := 0; k < 6; k++ {
		a := -math.Pi/2 + (float64(k)-2.5)*0.6
		ex, ey := x+math.Cos(a)*70*z, y-70*z+math.Sin(a)*40*z+30*z
		polygon(dst, []V{{X: x, Y: y - 72*z}, {X: (x+ex)/2 - 10*z, Y: (y-72*z+ey)/2 - 16*z}, {X: ex, Y: ey}, {X: (x+ex)/2 + 8*z, Y: (y-72*z+ey)/2 + 4*z}}, rgb([]uint32{0x4fb84a, 0x3fa040}[k%2]))
	}
	if seed%2 == 0 { // un régime de bananes
		for k := 0; k < 4; k++ {
			ellipse(dst, x+10*z, y-50*z+float64(k)*6*z, 8*z, 4*z, rgb(0xf0d040))
		}
	}
}

// drawHibiscus : un buisson tropical aux grandes fleurs rouges ou roses.
func drawHibiscus(dst *ebiten.Image, x, y, z float64, seed int) {
	if cam.iso {
		upright(x, y, func() { drawHibiscus(dst, x, y, z, seed) })
		return
	}
	dropShadow(dst, x, y, 36*z, 10*z, 24*z)
	circle(dst, x-16*z, y-14*z, 20*z, rgb(0x2f7a32))
	circle(dst, x+16*z, y-14*z, 19*z, rgb(0x2a6e2e))
	circle(dst, x, y-26*z, 24*z, rgb(0x3f9b3a))
	col := []uint32{0xe8304a, 0xff6fa8, 0xff8a3d, 0xffd23f}[seed%4]
	for k := 0; k < 5; k++ {
		a := float64(k)*1.3 + float64(seed%5)
		fx, fy := x+math.Cos(a)*18*z, y-22*z+math.Sin(a)*11*z
		for j := 0; j < 5; j++ {
			b := float64(j) * 2 * math.Pi / 5
			circle(dst, fx+math.Cos(b)*4*z, fy+math.Sin(b)*4*z, 3.5*z, rgb(col))
		}
		circle(dst, fx, fy, 2*z, rgb(0xffe36a))
	}
}

// exoticAnimals : la faune : des perroquets qui volent, des singes dans les arbres de
// la forêt, un toucan, des flamants roses au bord de la rivière.
func exoticAnimals(dst *ebiten.Image, p *sim.Place, ds []drawable, tl, br V, t float64) []drawable {
	x0, x1 := p.MinX(), p.MaxX()
	add := func(x, y float64, fn func()) {
		if x > tl.X-200 && x < br.X+200 && y > tl.Y-200 && y < br.Y+200 {
			ds = append(ds, drawable{x: x, y: y, fn: fn})
		}
	}
	// les perroquets (des aras), qui traversent le ciel en bande
	for i := 0; i < 5; i++ {
		fi := float64(i)
		span := x1 - x0 + 1200
		x := x0 - 600 + math.Mod(t*(140+fi*12)+fi*900, span)
		y := exoForestY + 220 + 70*math.Sin(t*0.7+fi*2) + fi*40
		col := []uint32{0xe8304a, 0x2f8ae0, 0x3fb84a, 0xffd23f, 0xe8304a}[i]
		add(x, y+400, func() { drawParrot(dst, x, y, col, t+fi) })
	}
	// des singes, accrochés aux arbres de la lisière
	for i := 0; i < 4; i++ {
		fi := float64(i)
		x := x0 + 700 + fi*(x1-x0-1400)/3
		y := exoForestY + 20
		add(x, y, func() { drawMonkey(dst, x, y, t+fi*1.3) })
	}
	// des flamants roses, au bord de la rivière (à l'écart des quais)
	for i := 0; i < 3; i++ {
		fi := float64(i)
		x := x1 - 900 + fi*130 // (loin des quais)
		y := exoRiverY + exoRiverH + 40 + fi*14
		add(x, y, func() { drawFlamingo(dst, x, y, t+fi) })
	}
	// un toucan, posé sur un rocher près de la cascade
	tx, ty := exoFallX+exoFallW+260, exoRiverY-30
	add(tx, ty, func() { drawToucan(dst, tx, ty, t) })
	return ds
}

func drawParrot(dst *ebiten.Image, x, y float64, col uint32, t float64) {
	w := 14 * math.Sin(t*10)
	ellipse(dst, x+30, y+260, 16, 5, rgba(0x000000, 0.15)) // (son ombre, loin en dessous)
	polygon(dst, []V{{X: x - 4, Y: y}, {X: x - 26, Y: y - 6 - w}, {X: x - 10, Y: y + 4}}, rgb(col))
	polygon(dst, []V{{X: x + 4, Y: y}, {X: x + 26, Y: y - 6 - w}, {X: x + 10, Y: y + 4}}, rgb(col))
	ellipse(dst, x, y, 8, 14, rgb(col))
	line(dst, x, y+10, x-2, y+26, 4, rgb(0x2f8ae0))
	circle(dst, x, y-12, 6, rgb(col))
	polygon(dst, []V{{X: x + 4, Y: y - 13}, {X: x + 11, Y: y - 10}, {X: x + 4, Y: y - 8}}, rgb(0x2b2b2b))
}

func drawMonkey(dst *ebiten.Image, x, y, t float64) {
	sw := 8 * math.Sin(t*2)
	line(dst, x, y-120, x+sw, y-40, 3, rgb(0x2f7a2a)) // la liane
	cx, cy := x+sw, y-40
	circle(dst, cx, cy, 14, rgb(0x7a4e2c))
	circle(dst, cx, cy-18, 11, rgb(0x7a4e2c))
	ellipse(dst, cx, cy-16, 7, 6, rgb(0xd9b48a))
	circle(dst, cx-3, cy-19, 1.8, rgb(0x2b2b2b))
	circle(dst, cx+3, cy-19, 1.8, rgb(0x2b2b2b))
	line(dst, cx-8, cy-8, cx-2, cy-38, 4, rgb(0x7a4e2c))                   // le bras sur la liane
	line(dst, cx+10, cy+6, cx+20, cy+22+4*math.Sin(t*3), 3, rgb(0x7a4e2c)) // la queue
}

func drawFlamingo(dst *ebiten.Image, x, y, t float64) {
	ellipse(dst, x+6, y+2, 14, 4, rgba(0x000000, 0.15))
	line(dst, x, y, x, y-40, 2.5, rgb(0xe07a90))
	line(dst, x+3, y-20, x+8, y-30, 2, rgb(0xe07a90))
	ellipse(dst, x, y-48, 16, 10, rgb(0xff8fb0))
	bend := 4 * math.Sin(t*0.8)
	line(dst, x+10, y-50, x+14+bend, y-74, 3.5, rgb(0xff8fb0))
	circle(dst, x+14+bend, y-77, 5, rgb(0xff8fb0))
	line(dst, x+17+bend, y-77, x+22+bend, y-72, 3, rgb(0x2b2b2b))
}

func drawToucan(dst *ebiten.Image, x, y, t float64) {
	ellipse(dst, x, y+6, 30, 10, rgb(0x857f75)) // le rocher
	ellipse(dst, x, y-14, 12, 16, rgb(0x1b1b1b))
	ellipse(dst, x+2, y-18, 6, 8, rgb(0xffffff))
	circle(dst, x+2, y-30, 9, rgb(0x1b1b1b))
	polygon(dst, []V{{X: x + 8, Y: y - 34}, {X: x + 30, Y: y - 30 + 2*math.Sin(t)}, {X: x + 8, Y: y - 25}}, rgb(0xff9a1a))
	circle(dst, x+4, y-32, 2, rgb(0xffffff))
}
