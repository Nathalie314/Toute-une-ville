package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// Le port : la mer à gauche (et au nord), la jetée en bois, les bateaux, les
// parcs et leurs dragues, la criée en bois au centre (sans murs).

var portSecCols = []uint32{0x3a7bd5, 0xd9534f, 0x2a9d8f, 0xf08a24} // coques des bateaux, section par section

func (g *game) portGround(dst *ebiten.Image, p *sim.Place, tl, br V) {
	t := g.t
	x0, x1 := tl.X-20, math.Min(br.X+20, sim.PortCrieeX0)
	// --- la mer (plus foncée au large)
	if x1 > x0 {
		rect(dst, x0, tl.Y-20, x1-x0, br.Y-tl.Y+40, rgb(0x3f93c9))
	}
	if tl.Y < -200 { // le large, plus foncé (sur toute la largeur)
		rect(dst, tl.X-20, tl.Y-20, br.X-tl.X+40, -200-tl.Y+20, rgb(0x2f7fb8))
	}
	// au nord de la criée aussi
	rect(dst, sim.PortCrieeX0-10, tl.Y-20, br.X-sim.PortCrieeX0+30, math.Max(0, sim.PortCrieeY0-20-tl.Y+20), rgb(0x3f93c9))
	// la terre du marais salant, au nord de la criée (bord de pierre côté mer)
	if tl.Y < sim.PortCrieeY0 {
		mx0, mx1 := sim.PortCrieeX0, sim.PortCrieeX1
		rect(dst, mx0-16, sim.PortSalTop-16, mx1-mx0+32, sim.PortCrieeY0-sim.PortSalTop+16, rgb(0x8d8a82))
		rect(dst, mx0, sim.PortSalTop, mx1-mx0, sim.PortCrieeY0-sim.PortSalTop, rgb(0xcbb88f))
		for k := 0; k < 60; k++ { // quelques touffes de salicorne
			fk := float64(k)
			x := mx0 + 20 + math.Mod(fk*233.7, mx1-mx0-40)
			y := sim.PortSalTop + 14 + math.Mod(fk*97.3, sim.PortCrieeY0-sim.PortSalTop-30)
			circle(dst, x, y, 4, rgba(0x6f8f3a, 0.55))
		}
	}
	// des vaguelettes
	for i := 0; i < 420; i++ {
		fi := float64(i)
		x := math.Mod(fi*733.7+t*12, 6000) - 3200
		y := math.Mod(fi*311.3, 3600) - 1900
		if x < tl.X-60 || x > br.X+60 || y < tl.Y-20 || y > br.Y+20 {
			continue
		}
		onWater := x < sim.PortCrieeX0-10 || y < sim.PortCrieeY0-30
		if x > sim.PortCrieeX0-30 && x < sim.PortCrieeX1+20 && y > sim.PortSalTop-30 {
			onWater = false // (la terre du marais salant)
		}
		if !onWater || (x > sim.PortJettyX0-20 && y > sim.PortJettyY0-10 && y < sim.PortJettyY1+10) {
			continue
		}
		a := 0.25 + 0.2*math.Sin(t*1.5+fi)
		line(dst, x, y, x+18, y-4, 2.5, rgba(0xd8f1fb, a))
		line(dst, x+18, y-4, x+36, y, 2.5, rgba(0xd8f1fb, a))
	}
	// --- le quai de la criée (bord de pierre côté mer)
	rect(dst, sim.PortCrieeX0-16, sim.PortCrieeY0-16, 16, sim.PortCrieeY1-sim.PortCrieeY0+130, rgb(0x8d8a82))
	rect(dst, sim.PortCrieeX0-16, sim.PortCrieeY0-16, br.X-sim.PortCrieeX0+40, 16, rgb(0x8d8a82))
	// --- la criée : un grand plancher, des poteaux, pas de murs
	cx0, cy0, cx1, cy1 := sim.PortCrieeX0, sim.PortCrieeY0, sim.PortCrieeX1, sim.PortCrieeY1
	rect(dst, cx0, cy0, cx1-cx0, cy1-cy0, rgb(0xc89a62))
	for y := cy0; y < cy1; y += 34 {
		line(dst, cx0, y, cx1, y, 2, rgba(0x9a7040, 0.55))
	}
	for k := 0; k < 30; k++ { // joints décalés des planches
		y := cy0 + float64(k)*34
		off := math.Mod(float64(k)*137, 260)
		for x := cx0 + off; x < cx1; x += 260 {
			line(dst, x, y, x, y+34, 1.5, rgba(0x9a7040, 0.5))
		}
	}
	for _, x := range []float64{cx0 + 30, (cx0 + cx1) / 2, cx1 - 30} {
		for _, y := range []float64{cy0 + 30, cy1 - 30} {
			circle(dst, x, y, 14, rgb(0x7a5530))
		}
	}
	roundRect(dst, (cx0+cx1)/2-150, cy0-6, 300, 54, 10, rgb(0x2b5a8a))
	say(dst, "CRIÉE", (cx0+cx1)/2, cy0+21, 32, true, 0, rgb(0xfff3d6))
	// --- la jetée
	jx0, jy0, jy1 := sim.PortJettyX0, sim.PortJettyY0, sim.PortJettyY1
	rect(dst, jx0-8, jy0-8, cx0-jx0+8, jy1-jy0+16, rgb(0x7a5530))
	rect(dst, jx0, jy0, cx0-jx0, jy1-jy0, rgb(0xc9a06a))
	for x := jx0; x < cx0; x += 38 {
		line(dst, x, jy0, x, jy1, 2, rgba(0x9a7040, 0.6))
	}
	for x := jx0 + 20; x < cx0; x += 180 { // les pieux
		circle(dst, x, jy0-4, 9, rgb(0x6a4a28))
		circle(dst, x, jy1+4, 9, rgb(0x6a4a28))
	}
	// --- les sections : les appontements des bateaux (au nord) et les noms
	for i := 0; i < 4; i++ {
		sx := sim.PortSecX(i)
		open := p.PortSecOpen(i, false)
		for k := 0; k <= sim.PortBoatMax; k++ {
			fx := sim.PortBerth(i, 0) - 57.5 + float64(k)*115
			if open {
				rect(dst, fx-8, sim.PortBerthY-10, 16, jy0-sim.PortBerthY+10, rgb(0xb88c58))
				circle(dst, fx, sim.PortBerthY-10, 8, rgb(0x6a4a28))
			} else {
				circle(dst, fx, sim.PortBerthY+40, 7, rgba(0x6a4a28, 0.6))
			}
		}
		a := 0.9
		if !open {
			a = 0.4
		}
		roundRect(dst, sx+130, sim.PortBerthY-80, 260, 40, 10, rgba(0x2b5a8a, a))
		sayFit(dst, sim.PortSecName(i, false), sx+260, sim.PortBerthY-60, 18, true, 0, 250, rgba(0xffffff, a))
		a = 0.9
		if !p.PortSecOpen(i, true) {
			a = 0.4
		}
		roundRect(dst, sx+110, sim.PortParcY+4*sim.PortCell+24, 300, 40, 10, rgba(0x2b5a8a, a))
		sayFit(dst, sim.PortSecName(i, true), sx+260, sim.PortParcY+4*sim.PortCell+44, 18, true, 0, 290, rgba(0xffffff, a))
	}
	// --- la route passe sur un pont au-dessus de la mer
	road(dst, tl, br, 1640)
	if tl.X < cx0 {
		bx1 := math.Min(br.X+20, cx0-110)
		for x := math.Floor((tl.X-40)/220) * 220; x < bx1; x += 220 {
			rect(dst, x+90, 1802, 40, 60, rgb(0x8d8a82)) // piles du pont
		}
		rect(dst, tl.X-20, 1620, bx1-tl.X+20, 8, rgb(0x6f6c66)) // garde-corps
		rect(dst, tl.X-20, 1796, bx1-tl.X+20, 8, rgb(0x6f6c66))
	}
}

// drawPortSea : les bateaux et les dragues (au-dessus de l'eau, sous les gens).
func (g *game) drawPortSea(dst *ebiten.Image, p *sim.Place) {
	t := g.t
	for _, b := range p.Boats {
		if b.Pos.Y < sim.PortDecoTop-200 {
			continue // au large : on ne le voit plus
		}
		moving := b.Phase == 1 || b.Phase == 3
		drawBoat(dst, b.Pos.X, b.Pos.Y, portSecCols[b.Sec], b.N, moving, b.Phase == 3, t+float64(b.K))
	}
	for i, d := range p.Dredges() {
		if !p.PortSecOpen(i, true) {
			continue
		}
		drawDredge(dst, d, i == 3, t)
	}
	if m := p.Station("marais"); m != nil && m.Built && p.Salt() != nil {
		drawSaltRake(dst, p.Salt(), t)
	}
}

// drawBoat : un bateau de pêche automatique, la proue vers le nord (centre x, y).
func drawBoat(dst *ebiten.Image, x, y float64, col uint32, n int, moving, back bool, t float64) {
	L := sim.PortBoatL
	top, bot := y-L/2, y+L/2
	bob := math.Sin(t*2) * 1.5
	if moving { // le sillage
		wy := bot
		if back {
			wy = bot // (il recule ? non : il revient la proue vers le nord, le sillage reste derrière)
		}
		for k := 0; k < 4; k++ {
			fk := float64(k)
			line(dst, x-10-fk*6, wy+fk*18, x-22-fk*10, wy+18+fk*18, 3, rgba(0xffffff, 0.55-fk*0.12))
			line(dst, x+10+fk*6, wy+fk*18, x+22+fk*10, wy+18+fk*18, 3, rgba(0xffffff, 0.55-fk*0.12))
		}
	}
	ellipse(dst, x+4, y+6, 34, L/2, rgba(0x000000, 0.18))
	polygon(dst, []V{{X: x, Y: top + bob}, {X: x + 30, Y: top + 50 + bob}, {X: x + 30, Y: bot - 8 + bob}, {X: x + 22, Y: bot + bob}, {X: x - 22, Y: bot + bob}, {X: x - 30, Y: bot - 8 + bob}, {X: x - 30, Y: top + 50 + bob}}, rgb(col))
	polygon(dst, []V{{X: x, Y: top + 12 + bob}, {X: x + 22, Y: top + 54 + bob}, {X: x + 22, Y: bot - 12 + bob}, {X: x - 22, Y: bot - 12 + bob}, {X: x - 22, Y: top + 54 + bob}}, rgb(0xf3ead8))
	// la cabine (automatique : antenne et gyrophare, personne à bord)
	roundRect(dst, x-16, y+10+bob, 32, 40, 6, rgb(0xffffff))
	rect(dst, x-12, y+14+bob, 24, 10, rgb(0x7fc4ea))
	line(dst, x, y+10+bob, x, y-14+bob, 2, rgb(0x555555))
	c := 0xff5a3a
	if math.Mod(t*2, 1) < 0.5 {
		c = 0xffc23a
	}
	circle(dst, x, y-16+bob, 4, rgb(uint32(c)))
	// la pêche sur le pont
	for k := 0; k < min(n, 12); k++ {
		fx := x - 12 + float64(k%4)*8
		fy := top + 64 + float64(k/4)*10 + bob
		ellipse(dst, fx, fy, 5, 2.5, rgb(0xb7c3cc))
	}
}

// drawDredge : la drague d'un parc (une grue jaune, son bras et sa benne ; un
// filet pour le bassin des gambas).
func drawDredge(dst *ebiten.Image, d *sim.Dredge, net bool, t float64) {
	b := d.Base
	roundRect(dst, b.X-34, b.Y-28, 68, 52, 8, rgb(0xe8b52a))
	rect(dst, b.X-34, b.Y+14, 68, 10, rgb(0x3b3b3b))
	circle(dst, b.X, b.Y-6, 12, rgb(0xb8861a))
	tip := d.Tip
	line(dst, b.X, b.Y-6, tip.X, tip.Y, 9, rgb(0xd9a520))
	line(dst, b.X, b.Y-6, tip.X, tip.Y, 3, rgb(0x8a6a10))
	if tip.Dist(b) < 2 {
		return
	}
	if net {
		r := 26.0
		if d.Phase == 2 {
			r = 26 + 6*math.Sin(t*10)
		}
		ellipse(dst, tip.X, tip.Y+8, r, r*0.6, rgba(0x2b2b2b, 0.25))
		for k := -2; k <= 2; k++ {
			line(dst, tip.X+float64(k)*r*0.4, tip.Y+8-r*0.5, tip.X+float64(k)*r*0.4, tip.Y+8+r*0.5, 1.5, rgba(0x333333, 0.7))
			line(dst, tip.X-r, tip.Y+8+float64(k)*r*0.25, tip.X+r, tip.Y+8+float64(k)*r*0.25, 1.5, rgba(0x333333, 0.7))
		}
	} else {
		a := 0.0
		if d.Phase == 2 {
			a = math.Sin(t*12) * 4
		}
		polygon(dst, []V{{X: tip.X - 18, Y: tip.Y - 6 + a}, {X: tip.X + 18, Y: tip.Y - 6 + a}, {X: tip.X + 12, Y: tip.Y + 14 + a}, {X: tip.X - 12, Y: tip.Y + 14 + a}}, rgb(0x6a6a6a))
		for k := -1; k <= 1; k++ {
			line(dst, tip.X+float64(k)*8, tip.Y+14+a, tip.X+float64(k)*8, tip.Y+22+a, 3, rgb(0x4a4a4a))
		}
	}
	if d.Phase == 2 { // des bulles
		for k := 0; k < 4; k++ {
			circle(dst, tip.X-16+float64(k)*11, tip.Y-14-math.Mod(t*30+float64(k)*7, 18), 3, rgba(0xffffff, 0.7))
		}
	}
	if d.N > 0 && d.Phase == 3 {
		say(dst, "×"+itoa(d.N), tip.X+28, tip.Y-14, 16, true, 0, rgb(0xffffff))
	}
}

// drawMarais : le marais salant (six bassins entre des digues d'argile ; l'eau
// s'évapore : du bleu au blanc, des cristaux apparaissent, puis un petit tas de sel).
func (g *game) drawMarais(dst *ebiten.Image, s *sim.Station) {
	for k := 0; k < sim.PortSalMax; k++ {
		b := sim.PortSalBasin(k)
		roundRect(dst, b.X-10, b.Y-10, b.W+20, b.H+20, 14, rgb(0x9c8460)) // la digue
		if !s.Built || k >= len(s.Plants) {
			roundRect(dst, b.X, b.Y, b.W, b.H, 8, rgb(0xb7a27a)) // pas encore en eau : de l'argile sèche
			for j := 0; j < 5; j++ {
				line(dst, b.X+30+float64(j)*80, b.Y+40+float64(j%2)*90, b.X+70+float64(j)*80, b.Y+90+float64(j%3)*40, 2, rgba(0x8a7550, 0.6))
			}
			dashedRect(dst, b.X+6, b.Y+6, b.W-12, b.H-12, 3, 14, 0, rgba(0xffffff, 0.45))
			continue
		}
		gr := math.Max(0, math.Min(1, s.Plants[k].G))
		roundRect(dst, b.X, b.Y, b.W, b.H, 8, lerpCol(0x3f93c9, 0xeef2f5, gr))
		// des reflets sur l'eau qui reste
		if gr < 0.85 {
			a := 0.35 * (1 - gr)
			for j := 0; j < 4; j++ {
				wx := b.X + 40 + math.Mod(float64(j)*97+g.t*10, b.W-100)
				wy := b.Y + 30 + float64(j)*45
				line(dst, wx, wy, wx+30, wy-4, 2, rgba(0xd8f1fb, a))
			}
		}
		// les cristaux de sel
		if gr > 0.3 {
			n := int((gr - 0.3) / 0.7 * 40)
			for j := 0; j < n; j++ {
				hx := math.Abs(math.Sin(float64(j)*12.9898+float64(k)*4.1) * 43758.5453)
				hy := math.Abs(math.Sin(float64(j)*78.233+float64(k)*2.7) * 24634.6345)
				cx := b.X + 14 + (hx-math.Floor(hx))*(b.W-28)
				cy := b.Y + 12 + (hy-math.Floor(hy))*(b.H-24)
				rect(dst, cx, cy, 5, 5, rgb(0xffffff))
				rect(dst, cx+1, cy+1, 3, 3, rgb(0xdfe7ee))
			}
		}
		if gr >= 1 { // prêt : un petit tas (mulon) au milieu, qui scintille
			c := b.C()
			ellipse(dst, c.X+6, c.Y+24, 56, 10, rgba(0x5a6a7a, 0.3)) // son ombre
			polygon(dst, []V{{X: c.X - 52, Y: c.Y + 24}, {X: c.X, Y: c.Y - 34}, {X: c.X + 52, Y: c.Y + 24}}, rgb(0xa9b6c2))
			polygon(dst, []V{{X: c.X - 48, Y: c.Y + 22}, {X: c.X, Y: c.Y - 30}, {X: c.X + 48, Y: c.Y + 22}}, rgb(0xffffff))
			polygon(dst, []V{{X: c.X, Y: c.Y - 30}, {X: c.X + 48, Y: c.Y + 22}, {X: c.X + 10, Y: c.Y + 22}}, rgb(0xc9d3dd))
			a := 0.5 + 0.4*math.Sin(g.t*4+float64(k))
			circle(dst, b.X+b.W-20, b.Y+18, 6, rgba(0xffffff, a))
		}
	}
	a := 0.9
	if !s.Built {
		a = 0.45
	}
	roundRect(dst, sim.PortSalPivot.X-160, sim.PortSalY0-62, 320, 42, 10, rgba(0x2b5a8a, a))
	say(dst, "Marais salant", sim.PortSalPivot.X, sim.PortSalY0-41, 20, true, 0, rgba(0xffffff, a))
}

// drawSaltRake : le râteau mécanique du marais (une grue jaune et un large râteau de bois).
func drawSaltRake(dst *ebiten.Image, d *sim.Dredge, t float64) {
	b := d.Base
	ellipse(dst, b.X+4, b.Y+30, 62, 12, rgba(0x000000, 0.2))
	roundRect(dst, b.X-56, b.Y-36, 112, 66, 10, rgb(0xe8b52a))
	rect(dst, b.X-56, b.Y+18, 112, 14, rgb(0x3b3b3b)) // les chenilles
	for k := 0; k < 7; k++ {
		rect(dst, b.X-52+float64(k)*16, b.Y+21, 8, 8, rgb(0x5a5a5a))
	}
	roundRect(dst, b.X-26, b.Y-30, 52, 26, 6, rgb(0x7fc4ea)) // la vitre de la cabine (automatique : personne)
	circle(dst, b.X, b.Y-6, 15, rgb(0xb8861a))
	tip := d.Tip
	line(dst, b.X, b.Y-6, tip.X, tip.Y, 9, rgb(0xd9a520))
	line(dst, b.X, b.Y-6, tip.X, tip.Y, 3, rgb(0x8a6a10))
	if tip.Dist(b) < 2 {
		return
	}
	sw := 0.0
	if d.Phase == 2 {
		sw = math.Sin(t*10) * 14 // il racle d'un côté à l'autre
	}
	x, y := tip.X+sw, tip.Y
	rect(dst, x-46, y-6, 92, 10, rgb(0x9a6b3e)) // la barre du râteau
	for k := 0; k < 9; k++ {
		line(dst, x-42+float64(k)*10.5, y+4, x-42+float64(k)*10.5, y+14, 3, rgb(0x6e4a26))
	}
	if d.N > 0 {
		polygon(dst, []V{{X: x - 30, Y: y - 6}, {X: x, Y: y - 26}, {X: x + 30, Y: y - 6}}, rgb(0xffffff))
		say(dst, "×"+itoa(d.N), x+48, y-18, 16, true, 0, rgb(0xffffff))
	}
}

// drawSaltHeap : le tas de sel au pied du râteau (il grossit avec le stock).
func drawSaltHeap(dst *ebiten.Image, s *sim.Station) {
	b := s.Body
	ellipse(dst, b.C().X, b.Y+b.H, b.W*0.55, 12, rgba(0x000000, 0.18))
	roundRect(dst, b.X, b.Y+b.H-14, b.W, 14, 4, rgb(0x8a6236)) // une palette de bois
	f := math.Min(1, float64(s.Pile)/math.Max(1, float64(s.PileMax)))
	if s.Pile > 0 {
		h := 12 + 50*f
		w := b.W*0.3 + b.W*0.65*f
		cx, by := b.C().X, b.Y+b.H-12
		polygon(dst, []V{{X: cx - w/2, Y: by}, {X: cx, Y: by - h}, {X: cx + w/2, Y: by}}, rgb(0xf6f8fa))
		polygon(dst, []V{{X: cx, Y: by - h}, {X: cx + w/2, Y: by}, {X: cx + w*0.1, Y: by}}, rgb(0xd5dde5))
	}
}

// drawParc : un parc immergé (les cases achetées, les coquillages qui grossissent).
func (g *game) drawParc(dst *ebiten.Image, s *sim.Station) {
	b := s.Body
	if !s.Built {
		dashedRect(dst, b.X, b.Y, b.W, b.H, 4, 16, 0, rgba(0xffffff, 0.45))
		return
	}
	roundRect(dst, b.X, b.Y, b.W, b.H, 10, rgba(0x1f5f8a, 0.35))
	for k, sl := range s.Slots {
		x, y := sl.X-sim.PortCell/2+6, sl.Y-sim.PortCell/2+6
		w := sim.PortCell - 12
		if k >= len(s.Plants) {
			dashedRect(dst, x, y, w, w, 2, 10, 0, rgba(0xffffff, 0.35))
			continue
		}
		roundRect(dst, x, y, w, w, 8, rgba(0x174a6a, 0.55))
		for _, xx := range []float64{x + w*0.33, x + w*0.66} { // les tables (ou le filet du bassin)
			line(dst, xx, y+4, xx, y+w-4, 2, rgba(0x9ac6e0, 0.5))
		}
		gr := s.Plants[k].G
		if gr < 0 {
			continue
		}
		f := 0.35 + 0.65*gr
		for j := 0; j < 4; j++ {
			fx := sl.X - 18 + float64(j%2)*36
			fy := sl.Y - 16 + float64(j/2)*32
			drawItem(dst, s.Item, fx, fy, 11*f)
		}
		if gr >= 1 {
			circle(dst, sl.X+w/2-12, sl.Y-w/2+12, 6, rgba(0xffffff, 0.6+0.3*math.Sin(g.t*4+float64(k))))
		}
	}
}

// drawBasket : un panier de la jetée (son contenu dépasse).
func drawBasket(dst *ebiten.Image, s *sim.Station) {
	b := s.Body
	ellipse(dst, b.C().X, b.Y+b.H, b.W*0.5, 10, rgba(0x000000, 0.18))
	roundRect(dst, b.X, b.Y+8, b.W, b.H-8, 14, rgb(0xb98a4a))
	for k := 0; k < 3; k++ {
		line(dst, b.X+6, b.Y+20+float64(k)*12, b.X+b.W-6, b.Y+20+float64(k)*12, 2, rgb(0x8a6030))
	}
	for x := b.X + 14; x < b.X+b.W-6; x += 16 {
		line(dst, x, b.Y+10, x, b.Y+b.H-2, 1.5, rgba(0x8a6030, 0.7))
	}
	if s.Pile > 0 {
		n := min(5, 1+s.Pile/10)
		for k := 0; k < n; k++ {
			drawItem(dst, s.Item, b.X+18+float64(k)*(b.W-36)/4, b.Y+6, 12)
		}
	}
}

// drawItemPort : poissons et coquillages.
func drawItemPort(dst *ebiten.Image, it sim.Item, x, y, r float64) bool {
	fish := func(body, belly uint32, l, h float64, stripes bool) {
		ellipse(dst, x-r*0.1, y, r*l, r*h, rgb(body))
		ellipse(dst, x-r*0.1, y+r*h*0.35, r*l*0.8, r*h*0.45, rgb(belly))
		polygon(dst, []V{{X: x + r*l*0.75, Y: y}, {X: x + r*l*1.25, Y: y - r*h*0.9}, {X: x + r*l*1.25, Y: y + r*h*0.9}}, rgb(body))
		circle(dst, x-r*l*0.65, y-r*h*0.2, r*0.12, rgb(0x1b1b1b))
		if stripes {
			for k := -1; k <= 1; k++ {
				line(dst, x+float64(k)*r*0.25, y-r*h*0.8, x+float64(k)*r*0.25-r*0.1, y+r*h*0.2, r*0.07, rgba(0x2b3a4a, 0.6))
			}
		}
	}
	switch it {
	case sim.Sardine:
		fish(0x8fa7bd, 0xe6eef4, 0.8, 0.3, false)
	case sim.Cabillaud:
		fish(0x9a8a6a, 0xf0e8d8, 0.85, 0.42, false)
	case sim.Thon:
		fish(0x2b4a7a, 0xc9d3dc, 0.85, 0.48, true)
	case sim.Saumon:
		ellipse(dst, x, y, r*0.85, r*0.5, rgb(0xf08a5d))
		for k := -1; k <= 1; k++ {
			line(dst, x+float64(k)*r*0.3, y-r*0.45, x+float64(k)*r*0.3+r*0.15, y+r*0.45, r*0.08, rgb(0xfbd3c0))
		}
	case sim.Huitre:
		ellipse(dst, x, y, r*0.85, r*0.6, rgb(0x8a8a7a))
		ellipse(dst, x, y-r*0.05, r*0.6, r*0.4, rgb(0xd9d4c8))
		circle(dst, x, y, r*0.18, rgb(0xf6f2ea))
	case sim.Moule:
		for k := -1; k <= 1; k += 2 {
			ellipse(dst, x+float64(k)*r*0.3, y, r*0.32, r*0.75, rgb(0x2b2b45))
			ellipse(dst, x+float64(k)*r*0.3, y-r*0.15, r*0.18, r*0.45, rgb(0x4a4a6a))
		}
	case sim.SaintJacques:
		polygon(dst, []V{{X: x, Y: y + r*0.7}, {X: x - r*0.9, Y: y - r*0.2}, {X: x - r*0.5, Y: y - r*0.75}, {X: x + r*0.5, Y: y - r*0.75}, {X: x + r*0.9, Y: y - r*0.2}}, rgb(0xf2c49a))
		for k := -2; k <= 2; k++ {
			line(dst, x, y+r*0.65, x+float64(k)*r*0.35, y-r*0.7, r*0.07, rgb(0xd08a5a))
		}
	case sim.Sel:
		// un petit tas de sel blanc (gros cristaux)
		polygon(dst, []V{{X: x - r*0.9, Y: y + r*0.6}, {X: x - r*0.35, Y: y - r*0.35}, {X: x, Y: y - r*0.7}, {X: x + r*0.4, Y: y - r*0.3}, {X: x + r*0.9, Y: y + r*0.6}}, rgb(0xf4f6f8))
		polygon(dst, []V{{X: x, Y: y - r*0.7}, {X: x + r*0.4, Y: y - r*0.3}, {X: x + r*0.9, Y: y + r*0.6}, {X: x + r*0.15, Y: y + r*0.6}}, rgb(0xd5dde5))
		for k := 0; k < 5; k++ {
			circle(dst, x-r*0.5+float64(k)*r*0.25, y+r*0.3-float64(k%2)*r*0.3, r*0.07, rgb(0xb8c4cf))
		}
	case sim.Gambas:
		for k := 0; k < 5; k++ {
			a := math.Pi*0.15 + float64(k)*0.32
			circle(dst, x+math.Cos(a)*r*0.55, y-math.Sin(a)*r*0.45+r*0.2, r*(0.32-float64(k)*0.03), rgb(0xf06a4a))
		}
		line(dst, x+r*0.55, y-r*0.1, x+r*1.1, y-r*0.7, r*0.06, rgb(0xc04a2a))
		line(dst, x+r*0.5, y-r*0.15, x+r*1.0, y-r*0.85, r*0.06, rgb(0xc04a2a))
	default:
		return false
	}
	return true
}
