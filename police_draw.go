package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
	"ville/tr"
)

// Le commissariat : le bâtiment (bureau du commissaire, hall, bureaux), le parking des voitures
// de police, les lieutenants qui y errent ; et, dans les lieux protégés, la patrouille qui passe.

const (
	polBlue  = 0x2b4f8a
	polNavy  = 0x1c2a44
	polWhite = 0xf4f6f8
)

// copLook : la tenue des policiers (chemise bleue, pantalon marine, casquette, insigne).
func copLook(n int) look {
	lk := lookFrom(n)
	lk.shirt, lk.pants = 0x4a78c2, polNavy
	lk.hat, lk.hatC = 2, polNavy
	lk.badge = true
	lk.skirt = 0
	return lk
}

// chiefLook : le commissaire (veste marine, pas de casquette).
func chiefLook() look {
	lk := lookFrom(4242)
	lk.shirt, lk.pants, lk.vest = polNavy, polNavy, 0x3a4a6a
	lk.badge = true
	lk.hat = 0
	lk.hair = 0x9a9a9a
	return lk
}

func (g *game) policeGround(dst *ebiten.Image, p *sim.Place, tl, br V) {
	gx := p.Gate.X
	road(dst, tl, br, 1640)
	// l'allée, de la porte à la route ; de l'autre côté, l'entrée du parking
	path := rgb(0xd6d2c8)
	rect(dst, gx-90, sim.PolBldY+sim.PolBldH, 180, 1624-sim.PolBldY-sim.PolBldH, path)
	b1, bn := sim.PolBay(1), sim.PolBay(sim.PolCars())
	px0, px1 := b1.X-sim.PolParkM, bn.X+bn.W+sim.PolParkM
	rect(dst, (px0+px1)/2-110, 1790, 220, sim.PolParkY-1790, rgb(0x55595e))
	// le parking des voitures de police
	slab(dst, px0, sim.PolParkY, px1-px0, sim.PolParkH, 14, 6, 0x5d6166)
	for k := 1; k <= sim.PolCars(); k++ {
		b := sim.PolBay(k)
		rect(dst, b.X-4, b.Y, 8, b.H, rgb(0xf2f2e6))
		rect(dst, b.X+b.W-4, b.Y, 8, b.H, rgb(0xf2f2e6))
		say(dst, "POLICE", b.C().X, b.Y+b.H-22, 20, true, 0, rgba(0xf2f2e6, 0.8))
		if g.sim.PolOnPatrol(k) { // la voiture est partie : où ?
			where := ""
			for s := 0; s < g.sim.PolSlots(k); s++ {
				if id := g.sim.PolPatrol(k, s); id != "" {
					if where != "" {
						where += ", "
					}
					where += tr.T(g.sim.Place(id).Name)
				}
			}
			tr.Seen(where)
			say(dst, "En patrouille", b.C().X, b.C().Y-18, 20, true, 0, rgba(0xffffff, 0.85))
			sayFit(dst, where, b.C().X, b.C().Y+14, 18, false, 0, b.W-20, rgba(0xffffff, 0.85))
		}
	}
	// la case des patrouilles (le menu des voitures)
	if g.sim.HasChief() {
		r := sim.PolPatrolZone()
		in := g.sim.Cur == p && r.Has(g.sim.Player.Pos)
		zoneBox(dst, r, in, g.t, 0x9fd3ff)
		say(dst, "PATROUILLES", r.C().X, r.C().Y-14, 22, true, 0, rgb(0xffffff))
		say(dst, "Reste ici ou clique", r.C().X, r.C().Y+20, 15, false, 0, rgba(0xffffff, 0.85))
	}
	// le bâtiment, vu de dessus : au centre le bureau du commissaire (au sommet) et le hall ;
	// à gauche les bureaux des capitaines, à droite la garde à vue
	bx, by, bw, bh := sim.PolBldX, sim.PolBldY, sim.PolBldW, sim.PolBldH
	slab(dst, bx, by, bw, bh, 8, 16, 0x7f8794)
	off, hall, wing, cells := sim.PolOffice(), sim.PolHall(), sim.PolWing(), sim.PolCells()
	rect(dst, off.X, off.Y, off.W, off.H, rgb(0xc9a477)) // parquet
	for y := off.Y; y < off.Y+off.H; y += 36 {
		rect(dst, off.X, y, off.W, 2, rgba(0x8a6a44, 0.35))
	}
	rect(dst, hall.X, hall.Y, hall.W, hall.H, rgb(0xe8ecf0)) // carrelage
	for x := hall.X; x < hall.X+hall.W; x += 60 {
		for y := hall.Y; y < hall.Y+hall.H; y += 60 {
			if int((x-hall.X)/60+(y-hall.Y)/60)%2 == 0 {
				rect(dst, x, y, math.Min(60, hall.X+hall.W-x), math.Min(60, hall.Y+hall.H-y), rgb(0xd3dbe4))
			}
		}
	}
	// les couloirs, entre les deux rangées (ils partent des portes de côté du hall)
	cy := sim.PolCorridorY()
	for _, r := range []sim.Rect{{X: wing.X - 10, Y: cy, W: wing.W + sim.PolGap + 10, H: sim.PolCorH}, {X: hall.X + hall.W, Y: cy, W: cells.W + sim.PolGap + 10, H: sim.PolCorH}} {
		rect(dst, r.X, r.Y+20, r.W, r.H-40, rgb(0xd9d6ce))
		for x := r.X; x < r.X+r.W; x += 80 {
			rect(dst, x, r.Y+r.H/2-2, 40, 4, rgba(0xb8b4aa, 0.8))
		}
	}
	for n := 1; n <= sim.PolOffices(); n++ { // les bureaux : moquette (grisée tant qu'ils ne sont pas construits)
		r := sim.PolCapOffice(n)
		c := rgb(0x9aa8bd)
		if !g.sim.PolOfficeOwned(n) {
			c = rgba(0x9aa8bd, 0.35)
		}
		rect(dst, r.X, r.Y, r.W, r.H, c)
	}
	for n := 1; n <= sim.PolCellN; n++ { // les cellules : du béton
		r := sim.PolCell(n)
		c := rgb(0xb8b4aa)
		if g.sim.Levels["commissariat.cellule_"+itoa(n)] == 0 {
			c = rgba(0xb8b4aa, 0.35)
		}
		rect(dst, r.X, r.Y, r.W, r.H, c)
	}
	// les murs (avec leurs portes)
	wall := rgb(0x4a505a)
	for _, x := range []float64{off.X - sim.PolGap/2, off.X + off.W + sim.PolGap/2} { // (une porte sur chaque couloir)
		rect(dst, x-14, by+40, 28, cy+20-by-40, wall)
		rect(dst, x-14, cy+sim.PolCorH-20, 28, by+bh-40-(cy+sim.PolCorH-20), wall)
	}
	rect(dst, off.X, off.Y+off.H+6, off.W/2-70, 10, wall)
	rect(dst, off.X+off.W/2+70, off.Y+off.H+6, off.W/2-70, 10, wall)
	say(dst, "BUREAU DU COMMISSAIRE", off.C().X, off.Y+24, 20, true, 0, rgba(0x4a3a22, 0.75))
	say(dst, "HALL", hall.C().X, hall.Y+24, 20, true, 0, rgba(0x2b3a5a, 0.7))
	say(dst, "BUREAUX DES CAPITAINES", wing.C().X, by+28, 20, true, 0, rgba(0xffffff, 0.85))
	say(dst, "GARDE À VUE", cells.C().X, by+28, 20, true, 0, rgba(0xffffff, 0.85))
	// la case de la vidéosurveillance (le menu des capitaines)
	if g.sim.HasChief() {
		r := sim.PolVideoZone()
		in := g.sim.Cur == p && r.Has(g.sim.Player.Pos)
		zoneBox(dst, r, in, g.t, 0x9fd3ff)
		say(dst, "VIDÉOSURVEILLANCE", r.C().X, r.C().Y-14, 18, true, 0, rgb(0x2b3a5a))
		say(dst, "Reste ici", r.C().X, r.C().Y+18, 15, false, 0, rgba(0x2b3a5a, 0.85))
	}
	// la façade : la porte, et l'enseigne bleue
	dx := sim.PolDoor().X
	rect(dst, dx-80, by+bh-6, 160, 22, rgb(0xd6d2c8))
	tw := textWidth("COMMISSARIAT", 26, true) + 60
	roundRect(dst, dx-tw/2, by+bh+26, tw, 42, 10, rgb(polBlue))
	say(dst, "COMMISSARIAT", dx, by+bh+47, 26, true, 0, rgb(0xffffff))
}

// captainLook : un capitaine (chemise bleu marine, casquette, insigne).
func captainLook(n int) look {
	lk := copLook(900 + 37*n)
	lk.shirt = 0x2b4f8a
	lk.vest = 0x1c2a44
	return lk
}

// policeDrawables : le mobilier, le commissaire, les capitaines, la garde à vue, les voitures
// garées, les lieutenants.
func (g *game) policeDrawables(dst *ebiten.Image, p *sim.Place, ds []drawable) []drawable {
	t := g.t
	off, hall := sim.PolOffice(), sim.PolHall()
	// le bureau du commissaire, au fond de la pièce (sa case est à droite)
	dx, dy := off.C().X-60, off.Y+170
	ds = append(ds, drawable{x: dx, y: dy + 70, fn: func() {
		roundRect(dst, dx-150, dy, 300, 70, 8, rgb(0x6e4322))
		rect(dst, dx-140, dy+6, 280, 8, rgb(0x8a5a33))
		roundRect(dst, dx-110, dy+16, 70, 44, 4, rgb(0xf3ead8)) // des dossiers
		roundRect(dst, dx+40, dy+14, 60, 40, 4, rgb(0x2b2b2b))  // un écran
		rect(dst, dx+46, dy+18, 48, 28, rgb(0x5b9bd5))
	}})
	ds = append(ds, drawable{x: dx, y: dy - 5, fn: func() {
		roundRect(dst, dx-32, dy-80, 64, 70, 14, rgb(0x2b2b33))
		if g.sim.HasChief() {
			drawPerson(dst, V{X: dx, Y: dy - 12}, V{Y: 1}, 0, false, chiefLook(), nil)
		}
		fx := off.X + 40 // le drapeau
		rect(dst, fx, off.Y+60, 5, 150, rgb(0x9a9a9a))
		rect(dst, fx+5, off.Y+64, 26, 46, rgb(0x2045a8))
		rect(dst, fx+31, off.Y+64, 26, 46, rgb(0xffffff))
		rect(dst, fx+57, off.Y+64, 26, 46, rgb(0xd9342b))
	}})
	// le hall : le comptoir d'accueil, deux plantes vertes
	hx := hall.C().X
	ay := sim.PolCorridorY() + sim.PolCorH - hall.Y + 30 // (sous le passage d'un couloir à l'autre)
	ds = append(ds, drawable{x: hx, y: hall.Y + ay + 70, fn: func() {
		roundRect(dst, hx-200, hall.Y+ay, 400, 70, 10, rgb(0x7a8ca8))
		rect(dst, hx-190, hall.Y+ay+6, 380, 12, rgb(0x9fb0c8))
		say(dst, "ACCUEIL", hx, hall.Y+ay+45, 20, true, 0, rgb(0xffffff))
	}})
	for _, x := range []float64{hall.X + 50, hall.X + hall.W - 50} {
		x := x
		ds = append(ds, drawable{x: x, y: hall.Y + hall.H - 30, fn: func() {
			roundRect(dst, x-20, hall.Y+hall.H-70, 40, 40, 6, rgb(0xa0522d))
			circle(dst, x, hall.Y+hall.H-90, 30, rgb(0x3f8f3a))
			circle(dst, x-12, hall.Y+hall.H-104, 18, rgb(0x52a84a))
		}})
	}
	// les bureaux des capitaines : un bureau, trois écrans ; le capitaine ; ce qu'il surveille
	for n := 1; n <= sim.PolOffices(); n++ {
		n, r := n, sim.PolCapOffice(n)
		if !g.sim.PolOfficeOwned(n) {
			continue
		}
		cx, cy := r.C().X, r.Y+120
		ds = append(ds, drawable{x: cx, y: cy + 60, fn: func() {
			roundRect(dst, cx-120, cy, 240, 60, 6, rgb(0x6e4322))
			for k := -1; k <= 1; k++ { // les écrans de la vidéosurveillance
				sx := cx + float64(k)*74
				roundRect(dst, sx-32, cy-34, 64, 44, 4, rgb(0x2b2b2b))
				c := uint32(0x3a5a4a)
				if g.sim.PolWatching(n) && g.sim.PolCamera(n) != "" {
					c = 0x5fae8a
					if math.Mod(t*0.7+float64(k+n), 3) < 0.15 {
						c = 0x9fe0c0
					}
				}
				rect(dst, sx-27, cy-30, 54, 34, rgb(c))
			}
		}})
		ds = append(ds, drawable{x: cx, y: cy - 5, fn: func() {
			roundRect(dst, cx-30, cy-70, 60, 62, 14, rgb(0x2b2b33))
			if g.sim.PolCaptain(n) {
				drawPerson(dst, V{X: cx, Y: cy - 10}, V{Y: 1}, 0, false, captainLook(n), nil)
			}
			lbl := tr.F("Bureau n° {n}", "n", itoa(n))
			if id := g.sim.PolCamera(n); id != "" && g.sim.PolCaptain(n) {
				lbl = tr.F("{lieu} : {p} %", "lieu", tr.T(mapNames[id]), "p", itoa(int(math.Round(100*g.sim.CatchChance(n)))))
			}
			tr.Seen(lbl)
			sayFit(dst, lbl, cx, r.Y+22, 16, true, 0, r.W-20, rgb(0xffffff))
		}})
	}
	// la garde à vue : des barreaux ; dedans, les voleurs arrêtés
	left := g.sim.InCustody()
	for n := 1; n <= sim.PolCellN; n++ {
		n, r := n, sim.PolCell(n)
		if g.sim.Levels["commissariat.cellule_"+itoa(n)] == 0 {
			continue
		}
		in := min(left, sim.PolCellCap)
		left -= in
		for k := 0; k < in; k++ {
			y0 := r.Y + 110 // (au fond de la cellule, loin des barreaux)
			if n > 5 {
				y0 = r.Y + 210
			}
			rb := &sim.Robber{Pos: V{X: r.X + 30 + float64(k%3)*55, Y: y0 + float64(k/3)*80}, Face: V{Y: 1}, Look: 31*n + 7*k}
			ds = append(ds, drawable{x: rb.Pos.X, y: rb.Pos.Y, fn: func() { drawRobber(dst, rb, t) }})
		}
		top := n <= 5 // (les barreaux du côté du couloir)
		ds = append(ds, drawable{x: r.C().X, y: r.Y + r.H, fn: func() {
			by := r.Y + r.H - 120
			if !top {
				by = r.Y
			}
			for x := r.X + 8; x < r.X+r.W; x += 22 {
				rect(dst, x, by, 5, 120, rgb(0x3a3f47))
			}
			rect(dst, r.X, by-4, r.W, 8, rgb(0x3a3f47))
		}})
	}
	// les voitures garées (achetées, pas en patrouille)
	for k := 1; k <= sim.PolCars(); k++ {
		if !g.sim.PolCarOwned(k) || g.sim.PolOnPatrol(k) {
			continue
		}
		b := sim.PolBay(k)
		c := b.C()
		ds = append(ds, drawable{x: c.X, y: c.Y + 90, fn: func() { drawPoliceCarTop(dst, c.X, c.Y, false, t) }})
	}
	// les lieutenants à pied : ils errent sur le parking en sifflotant
	for i, c := range p.Pol.Cops {
		if c == nil || c.InCar {
			continue
		}
		c, i := c, i
		ds = append(ds, drawable{x: c.Pos.X, y: c.Pos.Y, fn: func() {
			drawPerson(dst, c.Pos, c.Face, c.WalkT, c.Moving, copLook(c.Look), nil)
			if !c.Moving {
				drawNotes(dst, c.Pos.X, c.Pos.Y, t+float64(i))
			}
		}})
	}
	return ds
}

// drawPoliceCarTop : une voiture de police vue de dessus (garée, l'avant en haut ; down : vers le bas).
func drawPoliceCarTop(dst *ebiten.Image, x, y float64, down bool, t float64) {
	f := -1.0
	if down {
		f = 1
	}
	ellipse(dst, x+6, y+10, 70, 110, rgba(0x000000, 0.22))
	roundRect(dst, x-58, y-100, 116, 200, 26, rgb(polWhite))
	roundRect(dst, x-58, y-28, 116, 56, 6, rgb(polBlue)) // (la bande bleue)
	// le pare-brise (devant) et la lunette (derrière)
	roundRect(dst, x-44, y+f*62-14, 88, 28, 8, rgb(0x2b3a4a))
	roundRect(dst, x-40, y-f*66-10, 80, 20, 8, rgb(0x2b3a4a))
	// la rampe de gyrophares, qui clignote
	on := math.Mod(t*3, 1) < 0.5
	l, r := uint32(0xd9342b), uint32(0x2b6cff)
	if on {
		l, r = r, l
	}
	roundRect(dst, x-36, y-8, 36, 16, 4, rgb(l))
	roundRect(dst, x, y-8, 36, 16, 4, rgb(r))
	// les phares
	circle(dst, x-40, y+f*96, 7, rgb(0xfff3b0))
	circle(dst, x+40, y+f*96, 7, rgb(0xfff3b0))
}

// drawPoliceCarSide : la voiture de police vue de profil, sur la route (fx : 1 vers la droite).
func drawPoliceCarSide(dst *ebiten.Image, x, y, fx, t float64) {
	ellipse(dst, x, y+2, 66, 10, rgba(0x000000, 0.25))
	roundRect(dst, x-64, y-40, 128, 32, 12, rgb(polWhite))
	rect(dst, x-64, y-30, 128, 12, rgb(polBlue))
	roundRect(dst, x-36-fx*6, y-66, 66, 30, 12, rgb(polWhite))
	roundRect(dst, x-30-fx*6, y-61, 26, 20, 5, rgb(0xbfe3f5))
	roundRect(dst, x+0-fx*6, y-61, 26, 20, 5, rgb(0xbfe3f5))
	on := math.Mod(t*3, 1) < 0.5
	l, r := uint32(0xd9342b), uint32(0x2b6cff)
	if on {
		l, r = r, l
	}
	roundRect(dst, x-22-fx*6, y-76, 20, 11, 3, rgb(l))
	roundRect(dst, x-fx*6, y-76, 20, 11, 3, rgb(r))
	circle(dst, x+fx*60, y-26, 5, rgb(0xfff3b0))
	circle(dst, x-fx*61, y-26, 4, rgb(0xe74c3c))
	for _, wx := range []float64{-36, 36} {
		circle(dst, x+wx, y-8, 13, rgb(0x2b2b2b))
		circle(dst, x+wx, y-8, 6, rgb(0xb0b0b0))
	}
}

// patrolAt : captures d'écran (CAPPATROL=0.5) : où en est la voiture de patrouille (−1 : l'heure).
var patrolAt = -1.0

// patrolDrawables : dans un lieu protégé, la voiture de police passe de temps en temps.
func (g *game) patrolDrawables(dst *ebiten.Image, p *sim.Place, ds []drawable) []drawable {
	if p.Pol != nil || !g.sim.Patrolled(p.ID) {
		return ds
	}
	span := p.MaxX() - p.MinX() + 400
	if p.ID == "port" {
		span = p.H - p.Top + 400
	}
	cross := span / 450    // (elle roule à 450 par seconde)
	period := cross + 20.0 // (et repasse 20 s après)
	h := 0.0
	for _, c := range p.ID {
		h += float64(c) * 1.7
	}
	ph := math.Mod(g.t+h, period)
	if patrolAt >= 0 { // (captures d'écran : la voiture à cet endroit de son passage)
		ph = patrolAt * cross
	}
	if ph > cross {
		return ds
	}
	k := ph / cross
	n := int(math.Floor((g.t + h) / period)) // (un passage sur deux dans l'autre sens)
	t := g.t
	if p.ID == "port" { // la route du port est verticale
		y0, y1 := p.Top-200, p.H+200
		if n%2 == 0 {
			y := y0 + (y1-y0)*k
			ds = append(ds, drawable{x: sim.PortLaneS, y: y + 100, fn: func() { drawPoliceCarTop(dst, sim.PortLaneS, y, true, t) }})
		} else {
			y := y1 - (y1-y0)*k
			ds = append(ds, drawable{x: sim.PortLaneN, y: y + 100, fn: func() { drawPoliceCarTop(dst, sim.PortLaneN, y, false, t) }})
		}
		return ds
	}
	x0, x1 := p.MinX()-200, p.MaxX()+200
	if n%2 == 0 { // vers la droite, sur la voie du bas
		x := x0 + (x1-x0)*k
		ds = append(ds, drawable{x: x, y: 1750 + 20, fn: func() { drawPoliceCarSide(dst, x, 1770, 1, t) }})
	} else {
		x := x1 - (x1-x0)*k
		ds = append(ds, drawable{x: x, y: 1680 + 20, fn: func() { drawPoliceCarSide(dst, x, 1700, -1, t) }})
	}
	return ds
}

// ptsPrice : un prix en points d'école (« 10 pts Université »).
func ptsPrice(t string, n int) string {
	for i, pt := range sim.PointTypes {
		if pt == t {
			return tr.F("{n} pts {niveau}", "n", itoa(n), "niveau", tr.T(sim.LevelName(i)))
		}
	}
	return itoa(n)
}
