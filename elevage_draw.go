package main

import (
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// Étape 6 : les dessins de la ferme d'élevage, de la boucherie, du grenier à foin
// et les icônes de la viande.

var zoneFloor = map[string]uint32{"volaille": 0xe9d9a6, "bovin": 0xa8cf78, "ovin": 0xbcd99a, "porc": 0xc9a77a}

// elevageGround : la ferme d'élevage (chemins, espaces, frigos, route).
func (g *game) elevageGround(dst *ebiten.Image, p *sim.Place, tl, br V) {
	path := rgb(0xd9c08f)
	// les espaces (un sol par espace) et leur nom
	for _, z := range sim.TeamZones(p.ID) {
		open := p.ZoneOpen(z.ID)
		c := zoneFloor[z.ID]
		a := 0.85
		if !open {
			a = 0.35
		}
		roundRect(dst, z.X0, 360, z.X1-z.X0, 700, 30, rgba(c, a))
		roundRect(dst, (z.X0+z.X1)/2-110, 366, 220, 34, 10, rgba(0x5a3a22, 0.8))
		say(dst, strings.ToUpper(z.Name), (z.X0+z.X1)/2, 383, 18, true, 0, rgb(0xffeab0))
	}
	// la réserve en haut, les chemins
	roundRect(dst, 120, 140, p.W-300, 215, 24, rgba(0xb9b3a8, 0.9))
	say(dst, "RÉSERVE (camions de ma ferme)", 1150, 160, 18, true, 0, rgba(0x4a4a4a, 0.75))
	roundRect(dst, 100, 1000, p.W-200, 30, 14, path)
	// la chambre froide : le sol carrelé sous les frigos
	roundRect(dst, 140, 1030, p.W-750, 210, 20, rgb(0xdfe9ef))
	for x := 150.0; x < p.W-620; x += 48 {
		line(dst, x, 1034, x, 1236, 1.5, rgba(0xa9bcc8, 0.6))
	}
	say(dst, "CHAMBRE FROIDE", p.W-750, 1214, 16, true, 0, rgba(0x2b4a66, 0.7))
	// pistes des drones
	roundRect(dst, 140, 1320, p.W-750, 140, 20, rgba(0x8a8f96, 0.35))
	road(dst, tl, br, 1640)
	// la route du haut : les camions de ma ferme y roulent jusqu'à la réserve
	rect(dst, tl.X-10, 10, br.X-tl.X+20, 100, rgb(0x6b6f74))
	for x := math.Floor(tl.X/120) * 120; x < br.X+120; x += 120 {
		rect(dst, x, 58, 60, 4, rgba(0xf2f2e6, 0.8))
	}
	// panneau d'entrée
	roundRect(dst, 1100, 1475, 300, 46, 8, rgb(0x7a4a26))
	say(dst, "FERME D'ÉLEVAGE", 1250, 1498, 24, true, 0, rgb(0xffe7a8))
	g.backTruck(dst, p)
}

// boucherieGround : la boutique (sol carrelé blanc et rouge, espaces).
func (g *game) boucherieGround(dst *ebiten.Image, p *sim.Place) {
	for _, z := range sim.TeamZones(p.ID) {
		c := map[string]uint32{"volaille": 0xfbeede, "bovin": 0xf6dcdc, "ovin": 0xeef3e2, "porc": 0xfbe2ea}[z.ID]
		a := 0.9
		if !p.ZoneOpen(z.ID) {
			a = 0.35
		}
		_, _ = c, a // (le carrelage de la boutique partout, comme sous la caissière : seulement le nom de l'espace)
		roundRect(dst, (z.X0+z.X1)/2-150, 376, 300, 34, 10, rgba(0x8a2a2a, 0.9))
		say(dst, strings.ToUpper(z.Name), (z.X0+z.X1)/2, 393, 19, true, 0, rgb(0xffffff))
	}
}

// drawStage6 : les bâtiments de l'étape 6 (faux : pas l'un d'eux).
func (g *game) drawStage6(dst *ebiten.Image, p *sim.Place, s *sim.Station) bool {
	b := s.Body
	t := g.t
	switch {
	case s.ID == "pre_bovin" || s.ID == "pre_ovin":
		drawMeadow(dst, s, t)
	case strings.HasPrefix(s.ID, "abattoir_"):
		drawAbattoir(dst, s, t)
	case strings.HasPrefix(s.ID, "frigo_"):
		drawFridge(dst, g, s)
	case s.ID == "grenier":
		building(dst, b, 0xc98a4a, 0x7a3a22)
		roundRect(dst, b.C().X-50, b.Y+b.H-100, 100, 90, 8, rgb(0x5a3a22))
		for k := 0; k < 3; k++ {
			drawItem(dst, sim.Foin, b.C().X-30+float64(k)*30, b.Y+b.H-30, 16)
		}
		drawItem(dst, sim.Foin, b.C().X, b.Y+b.H-60, 16)
		if s.Working {
			for k := 0; k < 4; k++ {
				ph := math.Mod(t*0.8+float64(k)/4, 1)
				circle(dst, b.X+30+float64(k)*60, b.Y+40-ph*40, 3, rgba(0xe8c86a, 1-ph))
			}
		}
	case s.ID == "grange":
		// une grange ouverte, des bottes de foin empilées
		ellipse(dst, b.C().X, b.Y+b.H, b.W*0.55, 12, rgba(0x000000, 0.18))
		polygon(dst, []V{{X: b.X - 10, Y: b.Y + 10}, {X: b.C().X, Y: b.Y - 50}, {X: b.X + b.W + 10, Y: b.Y + 10}}, rgb(0x8a3a22))
		rect(dst, b.X, b.Y+6, 8, b.H-6, rgb(0x6e4322))
		rect(dst, b.X+b.W-8, b.Y+6, 8, b.H-6, rgb(0x6e4322))
		for i := 0; i < min(s.Stock, 12); i++ {
			drawItem(dst, sim.Foin, b.X+28+float64(i%6)*29, b.Y+b.H-14-float64(i/6)*18, 13)
		}
		roundRect(dst, b.C().X-40, b.Y-26, 80, 24, 6, rgba(0x2b2140, 0.85))
		say(dst, itoa(s.Stock)+"/"+itoa(s.StockMax), b.C().X, b.Y-14, 15, true, 0, rgb(0xffffff))
		sign(dst, b.C().X, b.Y+b.H+22, "Grange à foin")
	case s.ID == "silo":
		// deux silos ronds en tôle, et des sacs de blé devant
		ellipse(dst, b.C().X, b.Y+b.H, b.W*0.55, 12, rgba(0x000000, 0.18))
		for k, x := range []float64{b.X + 20, b.X + 110} {
			h := 150.0 - float64(k)*25
			rect(dst, x, b.Y+b.H-h, 70, h, rgb(0xc9ced4))
			for y := b.Y + b.H - h + 14; y < b.Y+b.H; y += 18 {
				line(dst, x, y, x+70, y, 1.5, rgba(0x8a9096, 0.7))
			}
			ellipse(dst, x+35, b.Y+b.H-h, 35, 14, rgb(0xa8aeb5))
			polygon(dst, []V{{X: x, Y: b.Y + b.H - h}, {X: x + 35, Y: b.Y + b.H - h - 34}, {X: x + 70, Y: b.Y + b.H - h}}, rgb(0x9aa1a8))
		}
		for i := 0; i < min(s.Stock, 10); i++ {
			drawItem(dst, sim.Ble, b.X+20+float64(i%5)*38, b.Y+b.H-10-float64(i/5)*16, 13)
		}
		roundRect(dst, b.C().X-40, b.Y-130, 80, 24, 6, rgba(0x2b2140, 0.85))
		say(dst, itoa(s.Stock)+"/"+itoa(s.StockMax), b.C().X, b.Y-118, 15, true, 0, rgb(0xffffff))
		sign(dst, b.C().X, b.Y+b.H+22, "Silo à blé")
	default:
		return false
	}
	return true
}

// drawMeadow : un grand pré clôturé (vaches, taureau, veaux ; brebis, moutons, agneaux).
func drawMeadow(dst *ebiten.Image, s *sim.Station, t float64) {
	b := s.Body
	bov := s.ID == "pre_bovin"
	ellipse(dst, b.C().X, b.Y+b.H, b.W*0.55, 16, rgba(0x000000, 0.15))
	roundRect(dst, b.X, b.Y+20, b.W, b.H-20, 10, rgb(0x7fbf4f))
	for i := 0; i < 30; i++ {
		x := b.X + 12 + math.Mod(float64(i)*97.3, b.W-24)
		y := b.Y + 34 + math.Mod(float64(i)*53.7, b.H-50)
		line(dst, x, y, x-3, y-7, 2, rgb(0x6aa83f))
	}
	// l'abri, à gauche
	hb := sim.Rect{X: b.X + 6, Y: b.Y - 10, W: 90, H: 110}
	if bov {
		building(dst, hb, 0xb8352c, 0x5a4a44)
	} else {
		building(dst, hb, 0xe8dcc0, 0x7a5a3a)
	}
	// clôture
	fc := rgb(0xa47447)
	for x := b.X; x <= b.X+b.W; x += 30 {
		rect(dst, x-3, b.Y+12, 6, 22, fc)
		rect(dst, x-3, b.Y+b.H-22, 6, 22, fc)
	}
	rect(dst, b.X, b.Y+18, b.W, 4, fc)
	rect(dst, b.X, b.Y+b.H-16, b.W, 4, fc)
	// les bêtes : deux adultes, le mâle, deux petits
	move := func(k int) (float64, float64) {
		x := b.X + 150 + float64(k%3)*90 + math.Sin(t*0.4+float64(k)*1.9)*18
		y := b.Y + 80 + float64(k/3)*60 + math.Cos(t*0.3+float64(k))*6
		return x, y
	}
	for k := 0; k < 5; k++ {
		x, y := move(k)
		big := k < 3
		sc := 0.8
		if !big {
			sc = 0.55
		}
		scaled(x, y, sc, func() {
			switch {
			case bov && k == 2: // le taureau
				drawBull(dst, x, y, t+float64(k))
			case bov:
				drawCow(dst, x, y, t+float64(k))
			case k == 2: // le bélier
				drawSheep(dst, x, y, t+float64(k))
				line(dst, x+18, y-12, x+28, y-18, 4, rgb(0xc9b48a))
			default:
				drawSheep(dst, x, y, t+float64(k))
			}
		})
	}
}

// drawBull : un taureau (brun, cornes).
func drawBull(dst *ebiten.Image, x, y, t float64) {
	bob := math.Sin(t*1.2) * 2
	ellipse(dst, x, y+26, 46, 10, rgba(0x000000, 0.2))
	for _, dx := range []float64{-28, -14, 16, 30} {
		rect(dst, x+dx-4, y+4, 8, 22, rgb(0x5a3a22))
	}
	ellipse(dst, x, y, 46, 26, rgb(0x7a4a2a))
	ellipse(dst, x+44, y-12+bob, 17, 15, rgb(0x6a3a1e))
	ellipse(dst, x+53, y-6+bob, 9, 7, rgb(0xd9a0a0))
	circle(dst, x+44, y-16+bob, 2.2, rgb(0x111111))
	line(dst, x+36, y-24+bob, x+28, y-34+bob, 5, rgb(0xf3ead8))
	line(dst, x+50, y-24+bob, x+58, y-34+bob, 5, rgb(0xf3ead8))
	ring(dst, x+56, y-4+bob, 4, 1.5, rgb(0xd9c08f))
}

// drawAbattoir : bâtiment carrelé blanc, toit rouge, porte à bandes.
func drawAbattoir(dst *ebiten.Image, s *sim.Station, t float64) {
	b := s.Body
	ellipse(dst, b.C().X, b.Y+b.H, b.W*0.55, 14, rgba(0x000000, 0.2))
	roundRect(dst, b.X, b.Y+20, b.W, b.H-20, 8, rgb(0xeef2f4))
	for x := b.X + 10; x < b.X+b.W-6; x += 24 {
		line(dst, x, b.Y+24, x, b.Y+b.H-4, 1, rgba(0xb8c4cc, 0.7))
	}
	rect(dst, b.X-8, b.Y+6, b.W+16, 22, rgb(0xa83232))
	// porte à lanières
	dx := b.C().X - 40
	roundRect(dst, dx, b.Y+b.H-78, 80, 74, 6, rgb(0x5a6470))
	for k := 0; k < 6; k++ {
		sw := 0.0
		if s.Working {
			sw = math.Sin(t*5+float64(k)) * 2
		}
		rect(dst, dx+4+float64(k)*12.5+sw, b.Y+b.H-74, 10, 68, rgba(0xcfe8f2, 0.75))
	}
	// ce qui entre et ce qui sort
	drawItem(dst, s.Inputs[0], b.X+30, b.Y+b.H-40, 16)
	drawItem(dst, s.Item, b.X+b.W-30, b.Y+b.H-40, 16)
}

// drawFridge : un frigo de la ferme d'élevage (porte vitrée, carcasses suspendues).
func drawFridge(dst *ebiten.Image, g *game, s *sim.Station) {
	b := s.Body
	ellipse(dst, b.C().X, b.Y+b.H, b.W*0.55, 12, rgba(0x000000, 0.18))
	roundRect(dst, b.X, b.Y, b.W, b.H, 10, rgb(0xdfe6ea))
	roundRect(dst, b.X+8, b.Y+6, b.W-16, b.H-14, 8, rgb(0xa9d6ee))
	line(dst, b.X+12, b.Y+12, b.X+b.W-12, b.Y+12, 3, rgb(0x8a949c))
	for i := 0; i < min(s.Stock, 7); i++ {
		x := b.X + 24 + float64(i)*20
		line(dst, x, b.Y+12, x, b.Y+20, 1.5, rgb(0x6a6a6a))
		drawItem(dst, s.Item, x, b.Y+32, 11)
	}
	roundRect(dst, b.X+b.W-46, b.Y+b.H-24, 50, 22, 7, rgba(0x2b2140, 0.85))
	say(dst, itoa(s.Stock), b.X+b.W-21, b.Y+b.H-13, 15, true, 0, rgb(0xffffff))
	say(dst, "❄", b.X+18, b.Y+b.H-14, 16, true, 0, rgb(0x2b6cd6))
}

// drawFridgeTruck : le camion frigorifique (caisse blanche, flocon).
func drawFridgeTruck(dst *ebiten.Image, cx, cy float64, busy bool) {
	ellipse(dst, cx, cy+38, 120, 10, rgba(0x000000, 0.25))
	roundRect(dst, cx-124, cy-48, 166, 80, 6, rgb(0xfafcff))
	rect(dst, cx-124, cy-48, 166, 10, rgb(0x2b6cd6))
	roundRect(dst, cx-112, cy-62, 40, 16, 4, rgb(0x8a949c)) // groupe froid
	sayFit(dst, "FRIGORIFIQUE", cx-41, cy-16, 16, true, 0, 150, rgb(0x2b6cd6))
	label := "ÉLEVAGE"
	if truckFrom == "boucherie" {
		label = "BOUCHERIE"
	}
	sayFit(dst, label, cx-41, cy+8, 15, true, 0, 150, rgb(0x8a2a2a))
	roundRect(dst, cx+42, cy-22, 70, 54, 10, rgb(0x2b6cd6))
	rect(dst, cx+68, cy-14, 34, 20, rgb(0x9fd3ff))
	for _, wx := range []float64{-85, -10, 80} {
		circle(dst, cx+wx, cy+34, 15, rgb(0x2b2b2b))
		circle(dst, cx+wx, cy+34, 6, rgb(0x999999))
	}
	if busy {
		circle(dst, cx+118, cy+10, 6, rgb(0xffd23f))
	}
}

// drawItemMeat : les icônes de l'étape 6 (faux si ce n'en est pas une).
func drawItemMeat(dst *ebiten.Image, it sim.Item, x, y, r float64) bool {
	k := r / 22
	meat, fat, bone := rgb(0xc0392b), rgb(0xf6d6c8), rgb(0xf3ead8)
	switch it {
	case sim.Foin:
		roundRect(dst, x-r*0.9, y-r*0.6, r*1.8, r*1.2, r*0.25, rgb(0xe8c86a))
		for i := -2; i <= 2; i++ {
			line(dst, x-r*0.8, y+float64(i)*r*0.2, x+r*0.8, y+float64(i)*r*0.2+r*0.05, r*0.06, rgb(0xc9a040))
		}
		line(dst, x-r*0.35, y-r*0.6, x-r*0.35, y+r*0.6, r*0.1, rgb(0x8a5a33))
		line(dst, x+r*0.35, y-r*0.6, x+r*0.35, y+r*0.6, r*0.1, rgb(0x8a5a33))
	case sim.Poulet:
		scaled(x, y, k*1.4, func() { drawChicken(dst, x-6, y+2, 0) })
	case sim.Canard:
		scaled(x, y, k*1.4, func() { drawDuck(dst, x-6, y+2, 0, true) })
	case sim.Vache, sim.Veau:
		f := 0.5
		if it == sim.Veau {
			f = 0.38
		}
		scaled(x, y, k*f*1.4, func() { drawCow(dst, x-14, y, 0) })
	case sim.Mouton, sim.Agneau:
		f := 0.85
		if it == sim.Agneau {
			f = 0.6
		}
		scaled(x, y, k*f*1.2, func() { drawSheep(dst, x-6, y, 0) })
	case sim.Cochon:
		scaled(x, y, k*1.1, func() { drawPig(dst, x-6, y, 0) })
	case sim.PouletAbattu, sim.CanardAbattu, sim.PouletCru:
		// volaille plumée, rose, cuisses en l'air ; sa tête dit si c'est un poulet
		// (bec pointu, crête rouge) ou un canard (bec rond et plat, pas de crête)
		duck := it == sim.CanardAbattu
		c, dk := rgb(0xf6c8b8), rgb(0xe8a898)
		if duck {
			c, dk = rgb(0xefc0a6), rgb(0xd99c86)
		}
		// le cou et la tête, à gauche
		line(dst, x-r*0.55, y-r*0.05, x-r*0.9, y-r*0.45, r*0.2, c)
		hx, hy := x-r*0.95, y-r*0.55
		circle(dst, hx, hy, r*0.24, c)
		circle(dst, hx-r*0.04, hy-r*0.04, r*0.04, rgb(0x2a2a2a))
		if duck {
			ellipse(dst, hx-r*0.3, hy+r*0.04, r*0.2, r*0.09, rgb(0xf0a030)) // bec de canard, rond et plat
		} else {
			polygon(dst, []V{{X: hx - r*0.18, Y: hy - r*0.06}, {X: hx - r*0.18, Y: hy + r*0.1}, {X: hx - r*0.42, Y: hy + r*0.03}}, rgb(0xf0b030)) // bec pointu
			for k := 0; k < 3; k++ {
				circle(dst, hx-r*0.08+float64(k)*r*0.1, hy-r*0.24-float64(k%2)*r*0.04, r*0.08, rgb(0xd9342b)) // la crête
			}
			ellipse(dst, hx-r*0.12, hy+r*0.2, r*0.05, r*0.09, rgb(0xd9342b)) // le barbillon
		}
		// le corps, dodu
		ellipse(dst, x, y+r*0.05, r*0.75, r*0.58, c)
		ellipse(dst, x+r*0.1, y+r*0.25, r*0.55, r*0.32, dk)
		ellipse(dst, x-r*0.15, y-r*0.15, r*0.35, r*0.16, rgba(0xffffff, 0.35)) // reflet
		// une aile repliée
		ellipse(dst, x-r*0.05, y+r*0.1, r*0.28, r*0.16, dk)
		// les cuisses, vers le haut à droite, l'os au bout
		ellipse(dst, x+r*0.55, y-r*0.3, r*0.3, r*0.22, c)
		line(dst, x+r*0.7, y-r*0.4, x+r*0.92, y-r*0.6, r*0.1, rgb(0xf3ead8))
		circle(dst, x+r*0.95, y-r*0.64, r*0.08, rgb(0xf3ead8))
		ellipse(dst, x+r*0.45, y-r*0.1, r*0.28, r*0.2, c)
		line(dst, x+r*0.62, y-r*0.15, x+r*0.88, y-r*0.28, r*0.1, rgb(0xf3ead8))
		circle(dst, x+r*0.91, y-r*0.3, r*0.08, rgb(0xf3ead8))
	case sim.PilonPoulet:
		// deux pilons crus : une goutte charnue au bout d'un os
		for k := 0; k < 2; k++ {
			dx := float64(k)*r*0.45 - r*0.2
			dy := float64(k) * r * 0.2
			line(dst, x+dx+r*0.1, y+dy, x+dx+r*0.5, y+dy+r*0.55, r*0.12, bone)
			circle(dst, x+dx+r*0.52, y+dy+r*0.6, r*0.11, bone)
			ellipse(dst, x+dx-r*0.08, y+dy-r*0.2, r*0.32, r*0.38, rgb(0xf0b8a0))
			ellipse(dst, x+dx-r*0.14, y+dy-r*0.3, r*0.12, r*0.1, rgba(0xffffff, 0.4))
		}
	case sim.PouletRoti:
		// un poulet doré couché sur le flanc, vu de côté : le croupion à gauche, l'aile repliée
		// devant, les deux cuisses à droite avec leurs os blancs pointés vers le haut ; un trait
		// brun autour de chaque morceau, des reflets et une ombre dessous
		ol, gold, deep, shine := rgb(0x7a3612), rgb(0xe08a2c), rgb(0xc46c1c), rgba(0xfff0c8, 0.55)
		oe := func(cx, cy, rx, ry float64, c color.Color) { // une ellipse cernée de brun
			ellipse(dst, x+cx*r, y+cy*r, rx*r+r*0.04, ry*r+r*0.04, ol)
			ellipse(dst, x+cx*r, y+cy*r, rx*r, ry*r, c)
		}
		os := func(x0, y0, x1, y1 float64) { // un os : la tige et ses deux nœuds
			line(dst, x+x0*r, y+y0*r, x+x1*r, y+y1*r, r*0.13, ol)
			circle(dst, x+(x1-0.04)*r, y+(y1-0.03)*r, r*0.09, ol)
			circle(dst, x+(x1+0.04)*r, y+(y1+0.03)*r, r*0.09, ol)
			line(dst, x+x0*r, y+y0*r, x+x1*r, y+y1*r, r*0.07, bone)
			circle(dst, x+(x1-0.04)*r, y+(y1-0.03)*r, r*0.055, bone)
			circle(dst, x+(x1+0.04)*r, y+(y1+0.03)*r, r*0.055, bone)
		}
		ellipse(dst, x, y+r*0.5, r*0.85, r*0.12, rgba(0x6a4a20, 0.3)) // l'ombre
		os(0.3, -0.25, 0.45, -0.62)                                   // la cuisse du fond
		oe(0.3, -0.18, 0.2, 0.2, deep)
		tail := []V{{X: -0.9, Y: 0.12}, {X: -0.72, Y: -0.42}, {X: -0.62, Y: -0.12}, {X: -0.5, Y: -0.3}, {X: -0.4, Y: 0.1}}
		var tp, ti []V
		for _, q := range tail {
			tp = append(tp, V{X: x + q.X*r, Y: y + q.Y*r})
			ti = append(ti, V{X: x + (q.X*0.85-0.08)*r, Y: y + (q.Y*0.8+0.02)*r})
		}
		polygon(dst, tp, ol) // le croupion
		polygon(dst, ti, deep)
		belly := []V{{X: 0.3, Y: 0.1}, {X: 0.92, Y: 0.32}, {X: 0.55, Y: 0.48}, {X: 0.2, Y: 0.45}}
		var bp []V
		for _, q := range belly {
			bp = append(bp, V{X: x + q.X*r, Y: y + q.Y*r})
		}
		polygon(dst, bp, ol)
		oe(0, 0.05, 0.7, 0.42, gold) // le corps
		ellipse(dst, x+r*0.05, y+r*0.25, r*0.6, r*0.2, deep)
		ellipse(dst, x+r*0.02, y+r*0.12, r*0.64, r*0.24, gold)
		ellipse(dst, x-r*0.05, y-r*0.22, r*0.18, r*0.06, shine)
		ellipse(dst, x-r*0.3, y-r*0.14, r*0.05, r*0.04, shine)
		oe(-0.3, 0.26, 0.34, 0.15, gold) // l'aile repliée
		ellipse(dst, x-r*0.3, y+r*0.32, r*0.3, r*0.07, deep)
		os(0.55, -0.02, 0.76, -0.42) // la cuisse de devant
		oe(0.42, 0.14, 0.26, 0.26, gold)
		ellipse(dst, x+r*0.46, y+r*0.24, r*0.18, r*0.12, deep)
		ellipse(dst, x+r*0.38, y+r*0.02, r*0.08, r*0.05, shine)
	case sim.CuissePoulet:
		// une cuisse dorée en équerre : le haut de cuisse (large), le genou, le pilon qui
		// part vers la droite jusqu'à son petit bout ; peau rôtie, quelques grains d'épices
		shape := []V{{X: -0.18, Y: -0.72}, {X: 0.05, Y: -0.62}, {X: 0.3, Y: -0.42}, {X: 0.52, Y: -0.18}, {X: 0.36, Y: -0.02},
			{X: 0.05, Y: 0.07}, {X: -0.3, Y: 0.13}, {X: 0.05, Y: 0.3}, {X: 0.4, Y: 0.46}, {X: 0.62, Y: 0.5}, {X: 0.62, Y: 0.66},
			{X: 0.3, Y: 0.66}, {X: -0.12, Y: 0.62}, {X: -0.45, Y: 0.5}, {X: -0.68, Y: 0.25}, {X: -0.66, Y: -0.05},
			{X: -0.52, Y: -0.3}, {X: -0.36, Y: -0.55}}
		{ // (des bords arrondis, une seule fois : la pointe du haut de cuisse reste nette)
			var sm []V
			for i, a := range shape {
				b := shape[(i+1)%len(shape)]
				sm = append(sm, a.Lerp(b, 0.25), a.Lerp(b, 0.75))
			}
			shape = sm
		}
		pts := func(k, dx, dy float64) []V {
			var ps []V
			for _, q := range shape {
				ps = append(ps, V{X: x + (q.X*k+dx)*r, Y: y + (q.Y*k+dy)*r})
			}
			return ps
		}
		polygon(dst, pts(1, 0.02, 0.06), rgb(0xc9541e)) // le dessous, plus cuit
		polygon(dst, pts(1, 0, 0), rgb(0xe8803a))
		// le petit bout du pilon (deux bosses)
		for _, q := range []V{{X: 0.68, Y: 0.52}, {X: 0.69, Y: 0.65}} {
			circle(dst, x+q.X*r, y+q.Y*r+r*0.03, r*0.1, rgb(0xc9541e))
			circle(dst, x+q.X*r, y+q.Y*r, r*0.1, rgb(0xe8803a))
		}
		polygon(dst, pts(0.62, -0.08, -0.02), rgb(0xf29a5c)) // le dessus, plus clair
		ellipse(dst, x+r*0.05, y-r*0.38, r*0.1, r*0.05, rgba(0xffffff, 0.35))
		for _, q := range []V{{X: -0.3, Y: -0.42}, {X: -0.05, Y: -0.5}, {X: 0.25, Y: -0.32}, {X: -0.5, Y: 0.05}, {X: -0.2, Y: 0.42}, {X: 0.15, Y: 0.48}, {X: 0.45, Y: 0.56}, {X: 0.05, Y: -0.15}} {
			circle(dst, x+q.X*r, y+q.Y*r, r*0.04, rgb(0xc8441a))
		}
	case sim.Boeuf, sim.VeauCarcasse, sim.MoutonCarcasse, sim.AgneauCarcasse:
		// une demi-carcasse pendue par la patte arrière à son crochet : la viande rouge rosé,
		// des marbrures et des traînées de gras blanc, plus foncée sur les bords (le veau, plus
		// clair ; le mouton, plus petit, plus sombre et plus gras ; l'agneau, petit et rosé)
		base, dark, light := uint32(0xd9504c), uint32(0xa82a2e), uint32(0xf2a59a)
		switch it {
		case sim.VeauCarcasse:
			base, dark, light = 0xec8a82, 0xcc5a5a, 0xf8c4bc
		case sim.MoutonCarcasse:
			base, dark, light = 0xb8433c, 0x8a2a2a, 0xf3e2da
			r *= 0.88
		case sim.AgneauCarcasse:
			base, dark, light = 0xe68078, 0xc4585a, 0xf8dcd6
			r *= 0.76
		}
		line(dst, x+r*0.03, y-r*1.25, x+r*0.03, y-r*0.98, r*0.05, rgb(0x6a6a6a)) // le crochet
		ellipse(dst, x+r*0.08, y-r*0.98, r*0.07, r*0.05, rgb(0x6a6a6a))
		shape := []V{{X: -0.05, Y: -0.95}, {X: 0.1, Y: -0.98}, {X: 0.16, Y: -0.85}, {X: 0.13, Y: -0.68}, {X: 0.28, Y: -0.5},
			{X: 0.38, Y: -0.2}, {X: 0.42, Y: 0.15}, {X: 0.38, Y: 0.45}, {X: 0.3, Y: 0.68}, {X: 0.38, Y: 0.86}, {X: 0.22, Y: 0.95},
			{X: -0.05, Y: 0.88}, {X: -0.3, Y: 0.78}, {X: -0.38, Y: 0.52}, {X: -0.35, Y: 0.2}, {X: -0.32, Y: -0.15},
			{X: -0.25, Y: -0.45}, {X: -0.12, Y: -0.68}, {X: -0.1, Y: -0.86}}
		for it := 0; it < 2; it++ { // (des bords arrondis)
			var sm []V
			for i, a := range shape {
				b := shape[(i+1)%len(shape)]
				sm = append(sm, a.Lerp(b, 0.25), a.Lerp(b, 0.75))
			}
			shape = sm
		}
		pts := func(k, dx float64) []V {
			var ps []V
			for _, q := range shape {
				ps = append(ps, V{X: x + (q.X*k+dx)*r, Y: y + q.Y*k*r})
			}
			return ps
		}
		polygon(dst, pts(1, 0), rgb(dark))
		polygon(dst, pts(0.9, -0.02), rgb(base))
		polygon(dst, pts(0.55, -0.1), rgb(light)) // l'intérieur, plus clair (le gras sous la viande)
		// des traînées de gras blanc et quelques marbrures foncées
		fatC := rgb(0xf8ece6)
		line(dst, x-r*0.12, y-r*0.6, x-r*0.02, y-r*0.15, r*0.06, fatC)
		line(dst, x-r*0.02, y-r*0.15, x-r*0.14, y+r*0.3, r*0.05, fatC)
		line(dst, x+r*0.12, y+r*0.2, x+r*0.22, y+r*0.62, r*0.05, fatC)
		ellipse(dst, x+r*0.2, y-r*0.35, r*0.08, r*0.14, rgb(dark))
		ellipse(dst, x-r*0.22, y+r*0.55, r*0.1, r*0.07, rgb(dark))
		ellipse(dst, x+r*0.25, y+r*0.78, r*0.1, r*0.06, rgb(dark))
		circle(dst, x+r*0.05, y-r*0.9, r*0.06, rgb(0xf3ead8)) // l'os de la patte
	case sim.Porc:
		// un porc entier pendu par ses deux pattes arrière à une barre : rose, ouvert au milieu
		// (l'échine rouge foncé et les côtes blanches), les pattes avant en bas
		pink, shade2, cut := rgb(0xf2b2b8), rgb(0xd98e98), rgb(0xe8344a)
		line(dst, x-r*0.35, y-r*1.12, x+r*0.4, y-r*0.95, r*0.08, rgb(0xb8bec4)) // la barre
		line(dst, x+r*0.02, y-r*1.04, x+r*0.02, y-r*0.88, r*0.05, rgb(0x9aa0a6))
		// les pattes arrière, en haut, coupées net
		for _, lx := range []float64{-0.18, 0.12} {
			line(dst, x+lx*r, y-r*0.9, x+(lx+0.04)*r, y-r*0.55, r*0.17, pink)
			ellipse(dst, x+lx*r, y-r*0.92, r*0.1, r*0.07, cut)
			circle(dst, x+lx*r, y-r*0.92, r*0.035, rgb(0xffffff))
		}
		// le corps
		shape := []V{{X: -0.22, Y: -0.62}, {X: 0.18, Y: -0.66}, {X: 0.4, Y: -0.5}, {X: 0.42, Y: -0.3}, {X: 0.3, Y: -0.12},
			{X: 0.3, Y: 0.3}, {X: 0.24, Y: 0.6}, {X: 0.12, Y: 0.7}, {X: -0.16, Y: 0.68}, {X: -0.26, Y: 0.5},
			{X: -0.3, Y: 0.1}, {X: -0.3, Y: -0.3}}
		for k := 0; k < 2; k++ { // (des bords arrondis)
			var sm []V
			for i, a := range shape {
				b := shape[(i+1)%len(shape)]
				sm = append(sm, a.Lerp(b, 0.25), a.Lerp(b, 0.75))
			}
			shape = sm
		}
		var ps []V
		for _, q := range shape {
			ps = append(ps, V{X: x + q.X*r, Y: y + q.Y*r})
		}
		polygon(dst, ps, pink)
		ellipse(dst, x+r*0.24, y-r*0.38, r*0.18, r*0.2, shade2)  // le jambon
		ellipse(dst, x+r*0.12, y+r*0.55, r*0.14, r*0.12, shade2) // l'épaule
		// l'échine, ouverte au milieu, et les côtes
		line(dst, x-r*0.04, y-r*0.6, x-r*0.12, y+r*0.55, r*0.1, rgb(0xa82a2e))
		for k := 0; k < 6; k++ {
			yy := y + r*0.05 + float64(k)*r*0.08
			line(dst, x-r*0.12, yy, x-r*0.02, yy-r*0.04, r*0.025, rgb(0xffffff))
		}
		for k := 0; k < 4; k++ { // les plis de la peau sur le flanc
			yy := y + r*0.12 + float64(k)*r*0.07
			line(dst, x+r*0.06, yy, x+r*0.26, yy+r*0.02, r*0.015, shade2)
		}
		// les pattes avant, en bas
		for _, lx := range []float64{-0.1, 0.08} {
			line(dst, x+lx*r, y+r*0.62, x+(lx-0.05)*r, y+r*0.95, r*0.12, pink)
			ellipse(dst, x+(lx-0.05)*r, y+r*0.97, r*0.06, r*0.03, cut)
		}
	case sim.Aiguillettes:
		for i := -1; i <= 1; i++ {
			roundRect(dst, x-r*0.8, y+float64(i)*r*0.45-r*0.15, r*1.6, r*0.3, r*0.15, rgb(0xb84a40))
		}
	case sim.Magret:
		ellipse(dst, x, y, r*0.9, r*0.5, meat)
		ellipse(dst, x, y-r*0.25, r*0.85, r*0.22, fat)
		for i := -1; i <= 1; i++ {
			line(dst, x+float64(i)*r*0.35-r*0.1, y-r*0.4, x+float64(i)*r*0.35+r*0.1, y-r*0.1, r*0.06, rgb(0xd9a080))
		}
	case sim.FoieGras:
		roundRect(dst, x-r*0.85, y-r*0.55, r*1.7, r*1.1, r*0.2, rgb(0x3a3a5a)) // terrine
		ellipse(dst, x, y-r*0.15, r*0.6, r*0.35, rgb(0xf0c8a0))
		ellipse(dst, x-r*0.2, y-r*0.2, r*0.2, r*0.1, rgba(0xffffff, 0.5))
	case sim.Steak:
		// une côte de bœuf vue de dessus : le bord cuit (brun), le gras (crème), la viande
		// rouge orangé coupée par l'os en T, la moelle au bout
		shape := []V{{X: -0.95, Y: 0.05}, {X: -0.85, Y: -0.25}, {X: -0.55, Y: -0.36}, {X: -0.1, Y: -0.36}, {X: 0.3, Y: -0.46},
			{X: 0.68, Y: -0.64}, {X: 0.9, Y: -0.57}, {X: 0.96, Y: -0.36}, {X: 0.9, Y: -0.05}, {X: 0.7, Y: 0.18},
			{X: 0.4, Y: 0.32}, {X: 0, Y: 0.38}, {X: -0.5, Y: 0.4}, {X: -0.85, Y: 0.3}}
		for it := 0; it < 3; it++ { // (des bords arrondis : on coupe les coins, trois fois)
			var sm []V
			for i, a := range shape {
				b := shape[(i+1)%len(shape)]
				sm = append(sm, a.Lerp(b, 0.25), a.Lerp(b, 0.75))
			}
			shape = sm
		}
		at := func(k, dy float64) []V {
			var ps []V
			for _, q := range shape {
				ps = append(ps, V{X: x + q.X*r*k, Y: y - r*0.08 + q.Y*r*k + dy})
			}
			return ps
		}
		polygon(dst, at(1, r*0.28), rgb(0x9a4a1a)) // l'épaisseur
		polygon(dst, at(1, 0), rgb(0xf3e3b0))      // le gras
		polygon(dst, at(0.84, 0), rgb(0xf0642a))   // la viande
		ellipse(dst, x-r*0.35, y-r*0.05, r*0.25, r*0.12, rgba(0xffa070, 0.45))
		// l'os en T et la moelle
		cream := rgb(0xf3e3b0)
		line(dst, x-r*0.55, y-r*0.36, x+r*0.05, y-r*0.06, r*0.08, cream)
		line(dst, x+r*0.05, y-r*0.06, x+r*0.62, y-r*0.5, r*0.08, cream)
		line(dst, x+r*0.05, y-r*0.06, x+r*0.05, y+r*0.3, r*0.08, cream)
		circle(dst, x+r*0.62, y-r*0.5, r*0.15, cream)
		circle(dst, x+r*0.62, y-r*0.5, r*0.1, rgb(0xeeeeea))
	case sim.RotiBoeuf, sim.MorceauxVeau, sim.Gigot, sim.CoteAgneau, sim.Cotelette:
		c := map[sim.Item]uint32{sim.Steak: 0xb03030, sim.RotiBoeuf: 0xa83a30, sim.MorceauxVeau: 0xd9706a, sim.Gigot: 0xa84a38, sim.CoteAgneau: 0xd06a60, sim.Cotelette: 0xf0a0a8}[it]
		switch it {
		case sim.RotiBoeuf:
			// un rôti ficelé : un gros boudin de viande rouge sombre, plus foncé à droite,
			// serré par des ficelles (en travers et en long) qui le font bomber
			roundRect(dst, x-r*0.9, y-r*0.52, r*1.8, r*1.04, r*0.4, rgb(0x8e1e2a))
			roundRect(dst, x-r*0.9, y-r*0.52, r*1.55, r*0.98, r*0.4, rgb(0xb02434))
			for k := 0; k < 4; k++ { // les bourrelets entre les ficelles
				cx := x - r*0.62 + float64(k)*r*0.42
				ellipse(dst, cx, y-r*0.08, r*0.19, r*0.42, rgb(0xbf2c3c))
			}
			for k := 0; k < 3; k++ { // quelques reflets
				ellipse(dst, x-r*0.55+float64(k)*r*0.4, y-r*0.28, r*0.09, r*0.03, rgba(0xffffff, 0.18))
			}
			str := rgb(0xf3e6c8)
			for k := 0; k < 5; k++ { // les ficelles en travers, qui s'arrondissent sur les bords
				sx := x - r*0.64 + float64(k)*r*0.32
				var prev V
				for i := 0; i <= 10; i++ {
					u := -1 + float64(i)/5
					q := V{X: sx + (1-u*u)*r*0.05*(float64(k)-2)/2, Y: y + u*r*0.48}
					if i > 0 {
						line(dst, prev.X, prev.Y, q.X, q.Y, r*0.05, str)
					}
					prev = q
				}
			}
			for _, yy := range []float64{-0.24, 0.26} { // les ficelles en long (un peu bombées)
				var prev V
				for i := 0; i <= 10; i++ {
					u := -1 + float64(i)/5
					q := V{X: x + u*r*0.86, Y: y + yy*r - (1-u*u)*r*0.05}
					if i > 0 {
						line(dst, prev.X, prev.Y, q.X, q.Y, r*0.045, str)
					}
					prev = q
				}
			}
			for k := 0; k < 5; k++ { // les nœuds, en haut
				u := (-0.64 + float64(k)*0.32) / 0.86
				circle(dst, x-r*0.64+float64(k)*r*0.32, y-r*0.24-(1-u*u)*r*0.05, r*0.045, rgb(0xe0d0a8))
			}
		case sim.Gigot:
			// un gigot rôti, doré orangé, coupé à gauche (la tranche rouge, son bord rose et l'os
			// rond au milieu), le manche et son os à droite ; une tranche posée devant
			crust, crustD, rim, red, pale := rgb(0xe8642a), rgb(0xc44a1a), rgb(0xf6c4bc), rgb(0xdc3e4a), rgb(0xf8dcd4)
			knob := rgb(0xf6d6c8)
			line(dst, x+r*0.6, y+r*0.16, x+r*0.8, y+r*0.16, r*0.11, knob) // l'os du manche
			circle(dst, x+r*0.82, y+r*0.1, r*0.075, knob)
			circle(dst, x+r*0.82, y+r*0.23, r*0.075, knob)
			shape := []V{{X: -0.32, Y: -0.53}, {X: 0.05, Y: -0.56}, {X: 0.32, Y: -0.42}, {X: 0.5, Y: -0.15}, {X: 0.72, Y: 0.04},
				{X: 0.8, Y: 0.16}, {X: 0.72, Y: 0.29}, {X: 0.45, Y: 0.33}, {X: 0.1, Y: 0.43}, {X: -0.32, Y: 0.43}}
			for k := 0; k < 2; k++ { // (des bords arrondis)
				var sm []V
				for i, a := range shape {
					b := shape[(i+1)%len(shape)]
					sm = append(sm, a.Lerp(b, 0.25), a.Lerp(b, 0.75))
				}
				shape = sm
			}
			at := func(k, dy float64) []V {
				var ps []V
				for _, q := range shape {
					ps = append(ps, V{X: x + q.X*r, Y: y + (q.Y*k+dy)*r})
				}
				return ps
			}
			polygon(dst, at(1, 0.03), crustD)
			polygon(dst, at(0.94, -0.01), crust)
			ellipse(dst, x+r*0.12, y-r*0.36, r*0.2, r*0.07, rgba(0xffb080, 0.6)) // un reflet
			// la tranche du gigot
			ellipse(dst, x-r*0.3, y-r*0.05, r*0.3, r*0.5, crustD)
			ellipse(dst, x-r*0.31, y-r*0.05, r*0.25, r*0.45, rim)
			ellipse(dst, x-r*0.31, y-r*0.04, r*0.2, r*0.38, red)
			line(dst, x-r*0.31, y-r*0.4, x-r*0.28, y-r*0.06, r*0.025, rim) // les veines de gras
			line(dst, x-r*0.28, y-r*0.06, x-r*0.46, y+r*0.2, r*0.025, rim)
			line(dst, x-r*0.28, y-r*0.06, x-r*0.14, y+r*0.24, r*0.025, rim)
			ellipse(dst, x-r*0.28, y-r*0.04, r*0.08, r*0.09, pale) // l'os
			ellipse(dst, x-r*0.28, y-r*0.04, r*0.04, r*0.045, rgb(0xffffff))
			// la tranche posée devant
			sx, sy := x-r*0.5, y+r*0.5
			ellipse(dst, sx, sy+r*0.07, r*0.42, r*0.19, crustD)
			ellipse(dst, sx, sy, r*0.42, r*0.19, crust)
			ellipse(dst, sx, sy, r*0.37, r*0.155, rim)
			ellipse(dst, sx, sy, r*0.31, r*0.12, red)
			line(dst, sx-r*0.28, sy-r*0.02, sx+r*0.28, sy+r*0.01, r*0.02, rim)
			ellipse(dst, sx+r*0.02, sy, r*0.07, r*0.04, pale)
			ellipse(dst, sx+r*0.02, sy, r*0.035, r*0.02, rgb(0xffffff))
		case sim.CoteAgneau:
			// une côtelette d'agneau : un médaillon rose de gras, épais, la noix de viande rouge
			// découpée par des veines de gras, puis un manche qui s'affine jusqu'à l'os
			smooth := func(sh []V, n int) []V {
				for k := 0; k < n; k++ {
					var sm []V
					for i, a := range sh {
						b := sh[(i+1)%len(sh)]
						sm = append(sm, a.Lerp(b, 0.25), a.Lerp(b, 0.75))
					}
					sh = sm
				}
				return sh
			}
			at := func(sh []V, dy float64) []V {
				var ps []V
				for _, q := range sh {
					ps = append(ps, V{X: x + q.X*r, Y: y + (q.Y+dy)*r})
				}
				return ps
			}
			// l'os, derrière, qui descend vers la droite
			line(dst, x+r*0.5, y-r*0.02, x+r*0.9, y+r*0.3, r*0.14, rgb(0xe8cfa8))
			line(dst, x+r*0.5, y-r*0.05, x+r*0.9, y+r*0.26, r*0.07, rgb(0xf6e6c8))
			ellipse(dst, x+r*0.9, y+r*0.3, r*0.08, r*0.08, rgb(0xe8cfa8))
			body := smooth([]V{{X: -0.92, Y: -0.12}, {X: -0.82, Y: -0.42}, {X: -0.3, Y: -0.52}, {X: 0.18, Y: -0.5},
				{X: 0.36, Y: -0.3}, {X: 0.62, Y: -0.12}, {X: 0.62, Y: 0.04}, {X: 0.35, Y: 0.04}, {X: 0.12, Y: 0.2},
				{X: -0.3, Y: 0.28}, {X: -0.82, Y: 0.2}}, 2)
			polygon(dst, at(body, 0.18), rgb(0xe3a2a6)) // l'épaisseur
			polygon(dst, at(body, 0), rgb(0xf6cccb))    // le gras du dessus
			meat := smooth([]V{{X: -0.8, Y: -0.14}, {X: -0.7, Y: -0.38}, {X: -0.3, Y: -0.45}, {X: 0.1, Y: -0.43},
				{X: 0.22, Y: -0.28}, {X: 0.06, Y: -0.06}, {X: -0.3, Y: 0.12}, {X: -0.72, Y: 0.1}}, 2)
			polygon(dst, at(meat, 0), rgb(0xe8403a))
			ellipse(dst, x-r*0.45, y-r*0.28, r*0.16, r*0.06, rgba(0xff9a80, 0.45))
			vein := rgba(0xf8e0d0, 0.9) // les veines de gras
			line(dst, x-r*0.2, y-r*0.44, x-r*0.12, y-r*0.2, r*0.03, vein)
			line(dst, x-r*0.12, y-r*0.2, x+r*0.12, y-r*0.3, r*0.03, vein)
			line(dst, x-r*0.12, y-r*0.2, x-r*0.2, y+r*0.05, r*0.03, vein)
			line(dst, x-r*0.75, y-r*0.06, x-r*0.4, y-r*0.1, r*0.03, vein)
			line(dst, x-r*0.4, y-r*0.1, x-r*0.2, y+r*0.05, r*0.03, vein)
			ellipse(dst, x+r*0.05, y+r*0.14, r*0.3, r*0.04, rgba(0xffffff, 0.3)) // un reflet sur le bord
		case sim.MorceauxVeau:
			for i := 0; i < 3; i++ {
				roundRect(dst, x-r*0.8+float64(i)*r*0.55, y-r*0.3+float64(i%2)*r*0.2, r*0.5, r*0.45, r*0.12, rgb(c))
			}
		default: // steak, côtes, côtelettes
			ellipse(dst, x, y, r*0.85, r*0.6, rgb(c))
			ellipse(dst, x+r*0.4, y-r*0.15, r*0.25, r*0.2, fat)
			if it != sim.Steak {
				line(dst, x-r*0.3, y, x-r*0.95, y-r*0.5, r*0.16, bone)
			}
		}
	case sim.SteakHache:
		// un steak haché cru : un palet rose, épais, des sillons de viande hachée en biais
		ellipse(dst, x, y+r*0.18, r*0.88, r*0.58, rgb(0xb8304e)) // l'épaisseur
		ellipse(dst, x, y, r*0.88, r*0.58, rgb(0xe0566f))
		// les sillons : des cordes parallèles du dessus, en biais, ondulées
		ea, eb := 0.8, 0.5  // (un peu dans le bord)
		dx, dy := 0.6, -0.8 // la direction d'un sillon
		nx, ny := 0.8, 0.6  // l'écart entre deux sillons
		for k := -4; k <= 4; k++ {
			c := float64(k) * 0.17
			// p(t) = c·n + t·d coupe l'ellipse (x/ea)² + (y/eb)² = 1 : on résout en t
			A := dx*dx/(ea*ea) + dy*dy/(eb*eb)
			B := 2 * (c*nx*dx/(ea*ea) + c*ny*dy/(eb*eb))
			C := c*c*nx*nx/(ea*ea) + c*c*ny*ny/(eb*eb) - 1
			D := B*B - 4*A*C
			if D <= 0 {
				continue
			}
			t0, t1 := (-B-math.Sqrt(D))/(2*A), (-B+math.Sqrt(D))/(2*A)
			var prev V
			for i := 0; i <= 24; i++ {
				t := t0 + (t1-t0)*float64(i)/24
				w := math.Sin(t*8+float64(k)) * 0.02
				q := V{X: x + (c*nx+t*dx+w*nx)*r, Y: y + (c*ny+t*dy+w*ny)*r}
				if i > 0 {
					line(dst, prev.X+r*0.03, prev.Y+r*0.03, q.X+r*0.03, q.Y+r*0.03, r*0.05, rgb(0xb8304e))
					line(dst, prev.X, prev.Y, q.X, q.Y, r*0.07, rgb(0xf28aa0))
				}
				prev = q
			}
		}
		for _, q := range []V{{X: -0.4, Y: 0.05}, {X: 0.2, Y: -0.2}, {X: 0.45, Y: 0.15}, {X: -0.1, Y: 0.25}} {
			circle(dst, x+q.X*r, y+q.Y*r, r*0.035, rgba(0xffffff, 0.6))
		}
	case sim.Jambon:
		// deux tranches de jambon posées l'une sur l'autre : roses, en forme de haricot, la couenne
		// rouge foncé sur l'épaisseur, un liseré clair et des marbrures plus pâles dessus
		shape := []V{{X: -0.9, Y: 0.18}, {X: -0.84, Y: -0.1}, {X: -0.6, Y: -0.28}, {X: -0.38, Y: -0.4}, {X: -0.22, Y: -0.6},
			{X: 0.1, Y: -0.68}, {X: 0.48, Y: -0.64}, {X: 0.6, Y: -0.44}, {X: 0.48, Y: -0.24}, {X: 0.14, Y: -0.1},
			{X: -0.06, Y: 0.18}, {X: -0.36, Y: 0.3}, {X: -0.74, Y: 0.34}}
		for k := 0; k < 3; k++ { // (des bords arrondis)
			var sm []V
			for i, a := range shape {
				b := shape[(i+1)%len(shape)]
				sm = append(sm, a.Lerp(b, 0.25), a.Lerp(b, 0.75))
			}
			shape = sm
		}
		cx, cy := 0.0, 0.0
		for _, q := range shape {
			cx += q.X / float64(len(shape))
			cy += q.Y / float64(len(shape))
		}
		at := func(dx, dy, k float64) []V {
			var ps []V
			for _, q := range shape {
				ps = append(ps, V{X: x + (cx+(q.X-cx)*k+dx)*r, Y: y + (cy+(q.Y-cy)*k+dy)*r})
			}
			return ps
		}
		slice := func(dx, dy float64) {
			polygon(dst, at(dx, dy+0.1, 1), rgb(0x9e1424))           // la couenne, sur l'épaisseur
			polygon(dst, at(dx, dy+0.05, 1), rgb(0xc42a38))          // son bord
			polygon(dst, at(dx, dy, 1), rgb(0xfbd4d8))               // le liseré clair
			polygon(dst, at(dx+0.005, dy+0.01, 0.93), rgb(0xf4a6ae)) // la chair
			polygon(dst, at(dx-0.02, dy-0.03, 0.62), rgb(0xf8bcc2))  // plus clair au milieu
			ellipse(dst, x+(cx+dx-0.2)*r, y+(cy+dy-0.12)*r, r*0.22, r*0.05, rgba(0xffffff, 0.35))
			ellipse(dst, x+(cx+dx+0.12)*r, y+(cy+dy+0.04)*r, r*0.18, r*0.04, rgba(0xffffff, 0.25))
		}
		slice(0.18, 0.2)                                       // la tranche du dessous
		polygon(dst, at(-0.09, -0.02, 1), rgba(0x8a1020, 0.3)) // l'ombre de celle du dessus
		slice(-0.17, -0.18)
	case sim.Pate:
		roundRect(dst, x-r*0.85, y-r*0.45, r*1.7, r*0.9, r*0.2, rgb(0xc89a6a))
		roundRect(dst, x-r*0.85, y-r*0.45, r*1.7, r*0.25, r*0.12, rgb(0xa8743a))
		circle(dst, x-r*0.3, y+r*0.1, r*0.1, rgb(0x8a5a33))
		circle(dst, x+r*0.25, y+r*0.2, r*0.08, rgb(0x8a5a33))
	default:
		return false
	}
	return true
}

var _ = sim.Foin
