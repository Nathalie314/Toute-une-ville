package main

import (
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
		ellipse(dst, x, y, r*0.85, r*0.62, rgb(0xc8782a))
		ellipse(dst, x-r*0.2, y-r*0.2, r*0.35, r*0.18, rgba(0xffe0a0, 0.6))
		line(dst, x-r*0.5, y+r*0.3, x-r*0.9, y+r*0.8, r*0.24, rgb(0xb5651d))
		line(dst, x+r*0.5, y+r*0.3, x+r*0.9, y+r*0.8, r*0.24, rgb(0xb5651d))
		circle(dst, x-r*0.9, y+r*0.85, r*0.14, bone)
		circle(dst, x+r*0.9, y+r*0.85, r*0.14, bone)
	case sim.CuissePoulet:
		ellipse(dst, x-r*0.15, y-r*0.15, r*0.6, r*0.5, rgb(0xe0a070))
		line(dst, x+r*0.2, y+r*0.2, x+r*0.75, y+r*0.75, r*0.18, bone)
		circle(dst, x+r*0.8, y+r*0.8, r*0.16, bone)
	case sim.Boeuf, sim.VeauCarcasse, sim.MoutonCarcasse, sim.AgneauCarcasse, sim.Porc:
		// carcasse suspendue à son crochet
		c := map[sim.Item]uint32{sim.Boeuf: 0xb03030, sim.VeauCarcasse: 0xd9706a, sim.MoutonCarcasse: 0xb84a40, sim.AgneauCarcasse: 0xe0807a, sim.Porc: 0xf0a0a8}[it]
		line(dst, x, y-r, x, y-r*0.6, r*0.1, rgb(0x6a6a6a))
		ellipse(dst, x, y, r*0.55, r*0.85, rgb(c))
		ellipse(dst, x, y, r*0.3, r*0.65, fat)
		ellipse(dst, x, y, r*0.18, r*0.5, rgb(c))
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
	case sim.Steak, sim.RotiBoeuf, sim.MorceauxVeau, sim.Gigot, sim.CoteAgneau, sim.Cotelette:
		c := map[sim.Item]uint32{sim.Steak: 0xb03030, sim.RotiBoeuf: 0xa83a30, sim.MorceauxVeau: 0xd9706a, sim.Gigot: 0xa84a38, sim.CoteAgneau: 0xd06a60, sim.Cotelette: 0xf0a0a8}[it]
		switch it {
		case sim.RotiBoeuf:
			roundRect(dst, x-r*0.85, y-r*0.5, r*1.7, r, r*0.45, rgb(c))
			for i := -1; i <= 1; i++ {
				line(dst, x+float64(i)*r*0.45, y-r*0.55, x+float64(i)*r*0.45, y+r*0.55, r*0.08, bone)
			}
		case sim.Gigot:
			ellipse(dst, x-r*0.15, y+r*0.1, r*0.65, r*0.5, rgb(c))
			line(dst, x+r*0.35, y-r*0.2, x+r*0.85, y-r*0.75, r*0.18, bone)
			circle(dst, x+r*0.9, y-r*0.8, r*0.15, bone)
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
		circle(dst, x, y, r*0.75, rgb(0x9a3a30))
		for i := 0; i < 7; i++ {
			a := float64(i) * 0.9
			circle(dst, x+math.Cos(a)*r*0.45, y+math.Sin(a)*r*0.45, r*0.1, rgb(0xc85a4a))
		}
	case sim.Jambon:
		polygon(dst, []V{{X: x - r*0.9, Y: y + r*0.2}, {X: x - r*0.2, Y: y - r*0.75}, {X: x + r*0.6, Y: y - r*0.4}, {X: x + r*0.8, Y: y + r*0.5}, {X: x - r*0.4, Y: y + r*0.75}}, rgb(0xe88a98))
		ellipse(dst, x-r*0.05, y, r*0.4, r*0.3, rgb(0xf4b6c0))
		line(dst, x-r*0.9, y+r*0.2, x-r*1.05, y+r*0.35, r*0.16, bone)
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
