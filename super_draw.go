package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// Le supermarché (4.1) : la zone du joueur à gauche, l'entrepôt et le magasin
// découpés en espaces, la colonne des caddies, le sas, le parking et le parking à étages.

var zoneCols = map[string]uint32{
	"ferme": 0xf3e3c4, "legumes": 0xdff0d2, "fruits": 0xfbe0dc, "cremerie": 0xe3eefa, "boulangerie": 0xf6e6cf,
	"glace": 0xdff3f8, "epicerie": 0xe2f4e4, "boucherie": 0xf5d9d9, "poisson": 0xdcecf7, "patisserie": 0xf9e1ee, "alcool": 0xe9e0f2, "huiles": 0xf6f0c8,
}

// superTruckStop : où s'arrête le camion d'un lieu (derrière son espace).
func superTruckStop(from string) float64 {
	id := map[string]string{"ferme": "ferme", "cremerie": "cremerie", "boulangerie": "boulangerie", "glacier": "glace", "boucherie": "boucherie", "exotique": "epicerie", "port": "poisson", "domaine": "alcool", "patisserie": "patisserie", "huilerie": "huiles"}[from]
	for _, z := range sim.SuperZones {
		if z.ID == id {
			return z.X + z.W/2
		}
	}
	return 260
}

func drawSuperGround(dst *ebiten.Image, g *game, p *sim.Place, tl, br V) {
	// route du haut (camions et voitures) et route du bas (voitures)
	rect(dst, tl.X-10, 10, br.X-tl.X+20, 100, rgb(0x6b6f74))
	for x := math.Floor(tl.X/120) * 120; x < br.X+120; x += 120 {
		rect(dst, x, 58, 60, 4, rgba(0xf2f2e6, 0.8))
	}
	road(dst, tl, br, sim.RoadY)
	// passage piéton : de la zone du joueur (sous la route) jusqu'aux caddies
	cxp := sim.SuperCartX + 150
	rect(dst, cxp-90, sim.SuperFloorY1+16, 180, sim.RoadY-sim.SuperFloorY1-16, rgb(0xb0b0a8))
	for k := 0; k < 7; k++ {
		rect(dst, cxp-80, 1648+float64(k)*20, 160, 11, rgba(0xf2f2e6, 0.95))
	}
	rect(dst, cxp-90, 1800, 180, sim.SuperYardY-1800, rgb(0xb0b0a8))
	x0, x1 := sim.SuperBX0, sim.SuperBX1
	// --- le bâtiment
	slab(dst, x0, 140, x1-x0, sim.SuperFloorY1+16-140, 6, 16, 0x8a8478)
	rect(dst, x0+16, 156, x1-x0-32, 284, rgb(0xb9b3a8))                  // entrepôt
	rect(dst, x0+16, 456, x1-x0-32, sim.SuperFloorY1-456, rgb(0xeef0f2)) // magasin
	for y := 456.0; y < sim.SuperFloorY1; y += 80 {
		line(dst, x0+16, y, x1-16, y, 1.5, rgba(0xc8ccd2, 0.7))
	}
	for x := x0 + 16; x < x1-16; x += 80 {
		line(dst, x, 456, x, sim.SuperFloorY1, 1.5, rgba(0xc8ccd2, 0.7))
	}
	// les espaces : un sol teinté, leur nom, leur coin d'entrepôt
	for _, z := range sim.SuperZones {
		open := z.Open(g.sim)
		c := zoneCols[z.ID]
		rect(dst, z.X-10, 456, z.W+20, z.Bot-456, rgba(c, 0.85))
		rect(dst, z.X-10, 160, z.W+20, 276, rgba(c, 0.35))
		// panneau suspendu
		roundRect(dst, z.X+z.W/2-120, 466, 240, 40, 10, rgb(shade32(c, 0.55)))
		sayFit(dst, z.Name, z.X+z.W/2, 486, 20, true, 0, 228, rgb(0xffffff))
		if !open {
			// coin d'entrepôt vide
			for k := 0.0; k < z.W; k += 40 {
				line(dst, z.X+k, 190, z.X+k+30, 430, 3, rgba(0x6f6a62, 0.25))
			}
		}
	}
	// cloison entrepôt / magasin : un passage par espace
	wall := rgb(0x6f6a62)
	prev := x0 + 16
	for _, z := range sim.SuperZones {
		cx := z.X + z.W/2
		rect(dst, prev, 440, cx-80-prev, 16, wall)
		prev = cx + 80
	}
	rect(dst, prev, 440, x1-16-prev, 16, wall)
	// murs (le mur droit s'ouvre sur le sas)
	rect(dst, x0, 140, 16, sim.SuperFloorY1+16-140, wall)
	rect(dst, x1-16, 140, 16, 1320, wall)
	rect(dst, x0, sim.SuperFloorY1, x1-x0, 16, wall)
	// --- la colonne des caddies et le sas d'entrée
	cx := sim.SuperCartX
	rect(dst, cx, 140, 300, sim.SuperFloorY1+16-140, rgb(0xc9c3b8))
	roundRect(dst, cx+10, 590, 280, 700, 12, rgba(0x2fa86b, 0.18))
	drawCartColumn(dst, p)
	// sas vitré
	rect(dst, cx, 1390, 300, 230, rgb(0xd8eef7))
	rect(dst, cx, 1390, 300, 8, rgb(0x6f6a62))
	rect(dst, cx+292, 1390, 8, 230, rgba(0x9fd3ff, 0.9))
	rect(dst, x1-16, 1460, 16, 144, rgba(0x9fd3ff, 0.9))
	roundRect(dst, cx+30, 1560, 240, 40, 8, rgb(0x2fa86b))
	say(dst, "SUPERMARCHÉ", cx+150, 1580, 22, true, 0, rgb(0xffffff))
	say(dst, "ENTRÉE • SORTIE", cx+150, 1430, 17, true, 0, rgba(0x2b4a66, 0.8))
	// --- route d'accès au parking
	ax := sim.SuperAccessX()
	rect(dst, ax-70, 110, 140, 1530, rgb(0x5f6368))
	// route de sortie (à droite des places au sol), jusqu'à la route du haut
	ex := sim.SuperExitX()
	rect(dst, ex-80, 110, 160, 1290+90-110, rgb(0x5f6368))
	// sens unique : flèches peintes (on monte l'accès, on va à droite dans les
	// allées, on monte la sortie)
	arrowUp := func(x, y float64) {
		polygon(dst, []V{{X: x, Y: y - 26}, {X: x - 16, Y: y - 4}, {X: x + 16, Y: y - 4}}, rgba(0xf2f2e6, 0.85))
		rect(dst, x-5, y-6, 10, 28, rgba(0xf2f2e6, 0.85))
	}
	for y := 260.0; y < 1600; y += 220 {
		arrowUp(ax, y)
	}
	for y := 260.0; y < 1360; y += 220 {
		arrowUp(ex, y)
	}
	roundRect(dst, ax-62, 1560, 124, 30, 8, rgb(0x2fa86b))
	say(dst, "ENTRÉE", ax, 1575, 16, true, 0, rgb(0xffffff))
	roundRect(dst, ex-62, 120, 124, 30, 8, rgb(0xd9534f))
	say(dst, "SORTIE", ex, 135, 16, true, 0, rgb(0xffffff))
	// --- le parking au sol
	px := sim.SuperParkX
	pw := 21*85 + 40.0
	roundRect(dst, px-12, 138, pw+24, 1384, 18, rgb(0x9a9a92))
	roundRect(dst, px, 150, pw, 1360, 12, rgb(0x5f6368))
	n := p.ParkTotal()
	for i, s := range p.Parks {
		if s.Etage {
			break
		}
		top := s.Pos.Y - 50
		if s.Pos.Y < s.Aisle {
			top = s.Pos.Y - 55
		}
		if i < n {
			line(dst, s.Pos.X-42, top, s.Pos.X-42, top+105, 3, rgba(0xffffff, 0.8))
			line(dst, s.Pos.X+42, top, s.Pos.X+42, top+105, 3, rgba(0xffffff, 0.8))
		} else {
			rect(dst, s.Pos.X-40, top, 80, 105, rgba(0x3b3f45, 0.55)) // pas encore achetée
		}
	}
	for _, a := range sim.SuperAisles {
		for x := px + 30; x < px+pw-30; x += 120 {
			rect(dst, x, a-3, 50, 6, rgba(0xffe066, 0.5))
		}
		// l'allée part de la route d'accès et rejoint la route de sortie
		rect(dst, ax+60, a-42, px-ax-50, 84, rgb(0x5f6368))
		rect(dst, px+pw-20, a-42, ex-80-(px+pw)+40, 84, rgb(0x5f6368))
		// les allées : vers la droite seulement
		for x := px + 150; x < px+pw-60; x += 420 {
			polygon(dst, []V{{X: x + 26, Y: a}, {X: x + 4, Y: a - 15}, {X: x + 4, Y: a + 15}}, rgba(0xf2f2e6, 0.8))
			rect(dst, x-22, a-5, 28, 10, rgba(0xf2f2e6, 0.8))
		}
	}
	say(dst, "P  "+itoa(p.ParkUsed())+" / "+itoa(n), px+pw/2, 128, 22, true, 0, rgb(0x2b4a66))
	// --- le parking à étages
	if e := p.Etages(); e > 0 {
		ex := sim.SuperEtageX()
		d := sim.SuperEtageDoor()
		h := 260 + 140*float64(e)
		y0 := d.Y + 120 - h
		rect(dst, ex, y0, 380, h, rgb(0x9aa0a6))
		for k := 0; k <= e+1; k++ {
			y := y0 + float64(k)*140
			rect(dst, ex, y, 380, 18, rgb(0x7f858b))
			for w := ex + 20; w < ex+360; w += 60 {
				rect(dst, w, y+40, 44, 60, rgba(0x3b3f45, 0.6))
			}
		}
		rect(dst, ex-6, d.Y-50, 70, 100, rgb(0x3b3f45)) // l'entrée des voitures
		roundRect(dst, ex+120, y0-60, 140, 60, 10, rgb(0x2b6cd6))
		say(dst, "P", ex+190, y0-30, 40, true, 0, rgb(0xffffff))
		say(dst, itoa(e)+" étage(s)", ex+190, y0+30, 18, true, 0, rgb(0x2b2b2b))
	}
}

// drawCartColumn : les caddies rangés, emboîtés par rangées de 30.
func drawCartColumn(dst *ebiten.Image, p *sim.Place) {
	s, ok := p.CartSpot()
	if !ok {
		return
	}
	x0 := sim.SuperCartX + 24
	n := p.CartsFree()
	// rangées de 12 caddies emboîtés (25 rangées au plus)
	for i := min(n, sim.SuperMaxPark) - 1; i >= 0; i-- {
		row, col := i/12, i%12
		x := x0 + 24 + float64(col)*19
		y := 618 + float64(row)*27
		polygon(dst, []V{{X: x - 14, Y: y - 14}, {X: x + 14, Y: y - 14}, {X: x + 11, Y: y + 2}, {X: x - 10, Y: y + 2}}, rgb(0xb8c2cc))
		roundStroke(dst, x-14, y-14, 28, 16, 2, 1.5, rgb(0x6c7a89))
		line(dst, x-14, y-14, x-19, y-17, 2.5, rgb(0xe74c3c))
		circle(dst, x-8, y+7, 2.5, rgb(0x2b2b2b))
		circle(dst, x+9, y+7, 2.5, rgb(0x2b2b2b))
	}
	// (comme le parking : ceux qui sont pris / le total ; avant, c'était ceux qui restaient)
	say(dst, "Caddies pris : "+itoa(p.CartsTotal()-n)+" / "+itoa(p.CartsTotal()), sim.SuperCartX+150, s.Y+30, 18, true, 0, rgb(0x1e6e46))
}

// drawZoneBarrier : un espace pas encore ouvert (barrières et panneau).
func drawZoneBarrier(dst *ebiten.Image, s *sim.Station) {
	b := s.Body
	rect(dst, b.X, b.Y, b.W, b.H, rgba(0x6f6a62, 0.35))
	for x := b.X + 10; x < b.X+b.W-10; x += 60 {
		rect(dst, x, b.Y+b.H-40, 50, 14, rgb(0xe74c3c))
		rect(dst, x+16, b.Y+b.H-40, 16, 14, rgb(0xffffff))
		rect(dst, x+4, b.Y+b.H-26, 4, 26, rgb(0x555555))
	}
	txt := "Bientôt"
	for _, z := range sim.SuperZones {
		if "barriere_"+z.ID == s.ID && !z.Soon {
			txt = map[string]string{"legumes": "Il faut la serre (ma ferme)", "fruits": "Il faut le verger (ma ferme)",
				"cremerie": "Il faut la crèmerie", "boulangerie": "Il faut la boulangerie", "glace": "Il faut le glacier", "boucherie": "Il faut la boucherie", "epicerie": "Il faut la ferme exotique", "alcool": "Il faut le domaine", "patisserie": "Il faut la pâtisserie", "huiles": "Il faut l'huilerie"}[z.ID]
		}
	}
	roundRect(dst, b.C().X-170, b.C().Y-34, 340, 68, 12, rgba(0x2b2140, 0.85))
	say(dst, "FERMÉ", b.C().X, b.C().Y-12, 22, true, 0, rgb(0xffd23f))
	sayFit(dst, txt, b.C().X, b.C().Y+14, 17, false, 0, 320, rgb(0xffffff))
}

// shade32 : une couleur plus sombre (en entier).
func shade32(c uint32, f float64) uint32 {
	r := uint32(math.Min(255, float64(c>>16&0xff)*f))
	g := uint32(math.Min(255, float64(c>>8&0xff)*f))
	b := uint32(math.Min(255, float64(c&0xff)*f))
	return r<<16 | g<<8 | b
}

// drawBlockBadge : pastille au-dessus d'un rayon presque vide : le maillon qui bloque.
func drawBlockBadge(dst *ebiten.Image, b sim.Block, in sim.Item, x, y, t float64) {
	col := map[sim.Block]uint32{sim.BlockStaff: 0x3b7dd8, sim.BlockTruck: 0xe8862a, sim.BlockDrone: 0x8e5cc4, sim.BlockProd: 0xc0392b, sim.BlockLock: 0x8a8a8a}[b]
	ellipse(dst, x, y+28, 20, 5, rgba(0x000000, 0.18))
	circle(dst, x, y, 27, rgb(col))
	circle(dst, x, y, 22, rgb(0xffffff))
	switch b {
	case sim.BlockStaff: // un magasinier et son carton
		circle(dst, x-4, y-10, 6, rgb(0xf1c27d))
		roundRect(dst, x-12, y-3, 16, 16, 5, rgb(0x3b7dd8))
		rect(dst, x+2, y-2, 12, 11, rgb(0xc9a36a))
		line(dst, x+2, y+3, x+14, y+3, 1.5, rgb(0x9a7440))
	case sim.BlockTruck: // un petit camion
		roundRect(dst, x-17, y-10, 22, 16, 3, rgb(0xf3ead8))
		roundRect(dst, x+5, y-4, 11, 10, 3, rgb(0xe8862a))
		rect(dst, x+9, y-2, 5, 4, rgb(0x9fd3ff))
		rect(dst, x-17, y+5, 33, 2, rgb(0x6b6b6b))
		for _, wx := range []float64{-11, 1, 11} {
			circle(dst, x+wx, y+9, 4, rgb(0x2b2b2b))
		}
	case sim.BlockDrone:
		drawDroneShape(dst, x, y+1, 0.62, t)
	case sim.BlockProd: // une usine qui fume
		rect(dst, x+6, y-15, 6, 14, rgb(0x7a4a3a))
		circle(dst, x+10, y-19+math.Sin(t*2)*1.5, 3.5, rgba(0x9a9a9a, 0.8))
		rect(dst, x-16, y-1, 32, 14, rgb(0xc0392b))
		for _, tx := range []float64{-16, -6, 4} { // le toit en dents de scie
			polygon(dst, []V{{X: x + tx, Y: y - 1}, {X: x + tx, Y: y - 9}, {X: x + tx + 10, Y: y - 1}}, rgb(0xc0392b))
		}
		rect(dst, x-12, y+4, 5, 5, rgb(0xffe066))
		rect(dst, x-3, y+4, 5, 5, rgb(0xffe066))
		rect(dst, x+6, y+4, 5, 5, rgb(0xffe066))
		if in >= 0 { // l'ingrédient qui manque, dans un petit rond
			circle(dst, x+22, y+20, 13, rgb(0xc0392b))
			circle(dst, x+22, y+20, 10.5, rgb(0xffffff))
			drawItem(dst, in, x+22, y+20, 7)
		}
	case sim.BlockLock: // un cadenas
		ring(dst, x, y-5, 7, 3, rgb(0x6b6b6b))
		roundRect(dst, x-10, y-3, 20, 16, 3, rgb(0x8a8a8a))
		circle(dst, x, y+4, 2.5, rgb(0x3a3a3a))
	}
}

// drawRayonBubble : devant un rayon presque vide, une bulle dit pourquoi.
func (g *game) drawRayonBubble(dst *ebiten.Image, p *sim.Place) {
	pl := g.sim.Player.Pos
	var best *sim.Station
	bd := 190.0
	for _, s := range p.Stations {
		if s.Kind != sim.Shelf || g.sim.RayonBlock(s) == sim.BlockNone {
			continue
		}
		if d := s.Body.C().Sub(pl).Len(); d < bd {
			best, bd = s, d
		}
	}
	if best == nil {
		return
	}
	l1, l2 := g.sim.RayonWhy(best)
	k := math.Max(0.6, math.Min(2.4, 1/g.zoom)) // (même taille à l'écran, quel que soit le zoom)
	w := math.Max(textWidth(l1, 20*k, true), textWidth(l2, 18*k, false)) + 36*k
	h := 78 * k
	b := best.Body
	cx := b.C().X
	bot := b.Y - 84
	x := cx - w/2
	roundRect(dst, x+3*k, bot-h+4*k, w, h, 14*k, rgba(0x000000, 0.25))
	polygon(dst, []V{{X: cx - 12*k, Y: bot - 1}, {X: cx + 12*k, Y: bot - 1}, {X: cx, Y: bot + 14*k}}, rgb(0xfffdf5))
	roundRect(dst, x, bot-h, w, h, 14*k, rgb(0xfffdf5))
	say(dst, l1, cx, bot-h+25*k, 20*k, true, 0, rgb(0x2b2140))
	say(dst, l2, cx, bot-h+54*k, 18*k, false, 0, rgb(0x5a4a70))
}
