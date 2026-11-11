package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// drawYard : la dalle de béton de la zone du joueur, l'hélisurface et les titres peints au sol.
func (g *game) drawYard(dst *ebiten.Image, p *sim.Place) {
	y := p.Yard
	if y.W == 0 {
		return
	}
	x0, x1, y0, y1 := y.X, y.X+y.W, y.Y, y.Y+y.H
	roundRect(dst, x0-6, y0-6, x1-x0+12, y1-y0+12, 22, rgb(0x9a9c97))
	roundRect(dst, x0, y0, x1-x0, y1-y0, 18, rgb(0xbfc1bc))
	for x := x0 + 140; x < x1-20; x += 140 {
		line(dst, x, y0+6, x, y1-6, 2, rgba(0x8f918c, 0.6))
	}
	for yy := y0 + 140; yy < y1-20; yy += 140 {
		line(dst, x0+6, yy, x1-6, yy, 2, rgba(0x8f918c, 0.6))
	}
	// liseré jaune peint au sol
	for _, r := range [][4]float64{{x0 + 14, y0 + 14, x1 - x0 - 28, 6}, {x0 + 14, y1 - 20, x1 - x0 - 28, 6}, {x0 + 14, y0 + 14, 6, y1 - y0 - 28}, {x1 - 20, y0 + 14, 6, y1 - y0 - 28}} {
		rect(dst, r[0], r[1], r[2], r[3], rgba(0xf2c14e, 0.85))
	}
	who := "RÉSERVÉ À LA DIRECTION"
	if p.ID == "ferme" {
		who = "RÉSERVÉ AU FERMIER"
	}
	say(dst, who, (x0+x1)/2, y1-30, 16, true, 0, rgba(0x6a6c67, 0.9))
	// titres des rangées d'améliorations
	hasStaff := false
	for _, sp := range p.Spots {
		if sp.ID == "vitesse" || sp.ID == "ouvriers" || sp.Team != "" {
			hasStaff = hasStaff || (sp.Zone.X >= x0 && sp.Zone.X < x1)
		}
	}
	if hasStaff {
		title := "OUVRIERS"
		if sim.TeamZones(p.ID) != nil {
			title = "EMPLOYÉS PAR ESPACE"
		}
		say(dst, title, x0+30, sim.YardRowY()-14, 17, true, -1, rgba(0x6a6c67, 0.95))
	}
	// hélisurface
	c := p.YardPad()
	circle(dst, c.X, c.Y, 118, rgb(0x7d807b))
	circle(dst, c.X, c.Y, 108, rgb(0x5f6368))
	ring(dst, c.X, c.Y, 92, 6, rgb(0xf2f2e6))
	rect(dst, c.X-34, c.Y-40, 14, 80, rgba(0xf2f2e6, 0.9))
	rect(dst, c.X+20, c.Y-40, 14, 80, rgba(0xf2f2e6, 0.9))
	rect(dst, c.X-20, c.Y-7, 40, 14, rgba(0xf2f2e6, 0.9))
}

// drawHelicopter : l'hélicoptère posé (le rotor tourne quand on attend sur un téléporteur).
func (g *game) drawHelicopter(dst *ebiten.Image, p *sim.Place) {
	c := p.YardPad()
	x, y := c.X, c.Y+20
	spin := g.busT > 0
	lift := 0.0
	if spin {
		lift = 10 * math.Min(1, g.busT) // il décolle un peu
	}
	ellipse(dst, x, y+34, 100, 16, rgba(0x000000, 0.25))
	y -= lift
	// patins
	line(dst, x-70, y+32, x+60, y+32, 5, rgb(0x3b3f45))
	line(dst, x-40, y+32, x-34, y+14, 4, rgb(0x3b3f45))
	line(dst, x+30, y+32, x+24, y+14, 4, rgb(0x3b3f45))
	// queue
	polygon(dst, []V{{X: x + 30, Y: y - 18}, {X: x + 150, Y: y - 30}, {X: x + 150, Y: y - 20}, {X: x + 30, Y: y + 4}}, rgb(0xd9534f))
	roundRect(dst, x+138, y-56, 18, 40, 6, rgb(0xd9534f))
	a := g.t * 0.6
	if spin {
		a = g.t * 30
	}
	for k := 0; k < 2; k++ {
		aa := a + float64(k)*math.Pi/2
		line(dst, x+147-math.Cos(aa)*20, y-36-math.Sin(aa)*20, x+147+math.Cos(aa)*20, y-36+math.Sin(aa)*20, 3, rgb(0x3b3f45))
	}
	// cabine
	roundRect(dst, x-70, y-50, 120, 70, 34, rgb(0xd9534f))
	roundRect(dst, x-66, y-44, 50, 40, 18, rgb(0xbfe3f5))
	rect(dst, x-30, y-4, 70, 6, rgb(0xf2f2e6))
	// rotor
	rect(dst, x-14, y-64, 8, 16, rgb(0x3b3f45))
	for k := 0; k < 2; k++ {
		aa := a + float64(k)*math.Pi/2
		dx := math.Cos(aa) * 130
		dy := math.Sin(aa) * 28
		line(dst, x-10-dx, y-66-dy, x-10+dx, y-66+dy, 5, rgba(0x2b2b2b, 0.9))
	}
	circle(dst, x-10, y-66, 7, rgb(0x2b2b2b))
}

// drawVestiaire : le vestiaire (sa porte ouvre la page « Personnage »).
func (g *game) drawVestiaire(dst *ebiten.Image, p *sim.Place, s *sim.Station) {
	b := s.Body
	building(dst, b, 0x8fa9c9, 0x4f6a8a)
	// porte
	d := sim.Rect{X: b.C().X - 34, Y: b.Y + b.H - 86, W: 68, H: 84}
	roundRect(dst, d.X, d.Y, d.W, d.H, 6, rgb(0x5b4636))
	circle(dst, d.X+d.W-14, d.Y+d.H/2, 4, rgb(0xf2c14e))
	// cintre sur la porte
	line(dst, d.C().X-16, d.Y+30, d.C().X, d.Y+18, 3, rgb(0xf2f2e6))
	line(dst, d.C().X+16, d.Y+30, d.C().X, d.Y+18, 3, rgb(0xf2f2e6))
	line(dst, d.C().X-16, d.Y+30, d.C().X+16, d.Y+30, 3, rgb(0xf2f2e6))
	sign(dst, b.C().X, b.Y+40, "Vestiaire")
	// paillasson : on s'y arrête
	z := p.Vestiaire
	hot := g.canBuyGear()
	if hot {
		glow(dst, z.X, z.Y, z.W, z.H, g.t)
	}
	roundRect(dst, z.X, z.Y, z.W, z.H, 14, rgba(0x9b6bd6, 0.25))
	dashedRect(dst, z.X+3, z.Y+3, z.W-6, z.H-6, 4, 12, g.t*25, rgb(0x9b6bd6))
	say(dst, "Personnage", z.C().X, z.C().Y-8, 18, true, 0, rgb(0x5a3a8a))
	say(dst, "touche P", z.C().X, z.C().Y+18, 14, false, 0, rgb(0x5a3a8a))
	if hot {
		badge(dst, z.X+z.W-6, z.Y+6, g.t)
	}
}
