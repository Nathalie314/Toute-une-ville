package main

import (
	"math"
	"sort"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// ---------------------------------------------------------------- caméra

// worldView : transformation du monde vers les pixels (k : sur-échantillonnage).
func (g *game) worldView(k float64) view {
	z := scr.S * k * g.zoom
	return view{s: z, ox: scr.W*k/2 - g.camPos.X*z, oy: scr.H*k/2 - g.camPos.Y*z}
}

func (g *game) updateCamera(p *sim.Place, dt float64) {
	t := g.sim.Player.Pos
	k := 1 - math.Exp(-dt*6)
	g.camPos = g.camPos.Add(t.Sub(g.camPos).Mul(k))
	// on ne sort pas (trop) du terrain
	hw := scr.VW / 2 / g.zoom
	hh := scr.VH / 2 / g.zoom
	clamp := func(v, lo, hi float64) float64 {
		if lo > hi {
			return (lo + hi) / 2
		}
		return math.Max(lo, math.Min(hi, v))
	}
	g.camPos.X = clamp(g.camPos.X, p.MinX()+hw-60, p.MaxX()-hw+60)
	// en bas, on laisse de la marge : l'aide affichée en bas de l'écran ne cache pas les zones de la route
	// en haut aussi : les encadrés (argent, commande…) ne cachent pas le terrain
	top := p.Top
	if p.DecoTop < top {
		top = p.DecoTop + 200 // (un décor au nord : la caméra monte le voir)
	}
	g.camPos.Y = clamp(g.camPos.Y, top+hh-200, p.H+200-hh)
}

// ---------------------------------------------------------------- monde

type drawable struct {
	y  float64
	fn func()
}

func (g *game) drawWorld(dst *ebiten.Image, p *sim.Place) {
	staffPlace = p.ID
	cam = g.worldView(scr.K)
	t := g.t
	// --- herbe et décor au sol
	tl := cam.toLogical(0, 0)
	br := cam.toLogical(scr.W*scr.K, scr.H*scr.K)
	dec := g.placeDecor(p)
	drawGrass(dst, p, tl, br)
	drawFlatDecor(dst, dec, tl, br)
	g.drawGround(dst, p, tl, br)
	if p.ID == "domaine" {
		drawAlcoholNotice(dst, p)
	}

	// --- champs (au sol)
	for _, s := range p.Stations {
		if s.Kind == sim.Field {
			g.drawField(dst, s)
		}
	}
	if p.ID == "port" {
		g.drawPortSea(dst, p) // les bateaux et les dragues
	}
	// --- ombres des machines et des bâtiments
	for _, s := range p.Stations {
		if s.Built && s.Kind != sim.Field {
			stationShadow(dst, s)
		}
	}
	// --- zones au sol
	for _, s := range p.Stations {
		if s.Built {
			g.drawZones(dst, p, s)
		}
	}
	for _, u := range p.Unlocks {
		if !u.Done && p.Visible(u) {
			g.drawUnlock(dst, u)
		} else if p.ShowLocked(u) {
			r := u.Zone
			roundRect(dst, r.X, r.Y, r.W, r.H, 18, rgba(0x000000, 0.12))
			dashedRect(dst, r.X+4, r.Y+4, r.W-8, r.H-8, 4, 14, 0, rgba(0x8a8a8a, 0.8))
			say(dst, u.Name, r.C().X, r.Y+32, 19, true, 0, rgb(0x7a7a7a))
			say(dst, u.Locked, r.C().X, r.Y+r.H-36, 16, false, 0, rgb(0x7a7a7a))
		}
	}
	for _, s := range p.Stations {
		if s.Built && s.Upgradable() {
			g.drawLevelZone(dst, p, s)
		}
		if s.Built && s.Stockable() {
			g.drawStockZone(dst, p, s)
		}
	}
	g.drawBus(dst, p)
	for _, sp := range p.Spots {
		if p.SpotVisible(sp) {
			g.drawSpot(dst, p, sp)
		} else if why, title := p.DroneLock(sp), "Drone"; why != "" || p.HireLock(sp) != "" {
			// pas encore achetable : on dit pourquoi
			if why == "" {
				why, title = p.HireLock(sp), "Ouvrier"
			}
			r := sp.Zone
			roundRect(dst, r.X, r.Y, r.W, r.H, 14, rgba(0x000000, 0.10))
			dashedRect(dst, r.X+3, r.Y+3, r.W-6, r.H-6, 4, 12, 0, rgba(0x8a8a8a, 0.8))
			say(dst, title, r.C().X, r.Y+28, 17, true, 0, rgb(0x7a7a7a))
			for i, ln := range strings.Split(why, "|") {
				sayFit(dst, ln, r.C().X, r.Y+56+float64(i)*20, 14, false, 0, r.W-8, rgb(0x6a6a6a))
			}
		}
	}
	g.drawFlowers(dst, p) // (cadeau de la fleuriste)
	// --- tout ce qui a une hauteur, trié par profondeur
	var ds []drawable
	for _, s := range p.Stations {
		s := s
		if !s.Built || s.Kind == sim.Field {
			continue
		}
		ds = append(ds, drawable{s.Body.Y + s.Body.H, func() { g.drawStation(dst, p, s) }})
	}
	for _, w := range p.Workers {
		w := w
		ds = append(ds, drawable{w.Pos.Y, func() { drawAgent(dst, p.ID, w, t) }})
	}
	for _, r := range p.Registers() {
		if c := r.Cashier; c != nil {
			ds = append(ds, drawable{c.Pos.Y, func() { drawAgent(dst, p.ID, c, t) }})
		}
	}
	for _, c := range p.Customers {
		c := c
		if c.InCar() {
			continue // encore dans sa voiture
		}
		if c.Pos.X < tl.X-150 || c.Pos.X > br.X+150 || c.Pos.Y < tl.Y-150 || c.Pos.Y > br.Y+200 {
			continue // hors de l'écran (il peut y en avoir des centaines)
		}
		ds = append(ds, drawable{c.Pos.Y, func() { drawCustomer(dst, p, c, t) }})
	}
	for _, v := range p.Vehicles {
		v := v
		if v.Pos.X < tl.X-200 || v.Pos.X > br.X+200 || v.Hidden {
			continue
		}
		ds = append(ds, drawable{v.Pos.Y + 20, func() { drawVehicle(dst, v) }})
	}
	if g.sim.Cur == p {
		pl := g.sim.Player
		ds = append(ds, drawable{pl.Pos.Y, func() { g.drawPlayer(dst, pl) }})
	}
	if p.Yard.W > 0 {
		ds = append(ds, drawable{p.YardPad().Y + 50, func() { g.drawHelicopter(dst, p) }})
	}
	if r, ok := serreRect(p); ok && p.ID == "ferme" {
		ds = append(ds, drawable{r.Y + 1, func() { drawSerreRoof(dst, p) }})
	}
	if p.ID == "ecole" {
		ds = g.schoolDrawables(dst, p, ds)
	}
	if p.ID == "parc" {
		ds = g.parcDrawables(dst, p, ds, tl, br)
	}
	ds = decorDrawables(dst, dec, ds, tl, br)
	// arbres décoratifs autour (à la ferme)
	for _, tr := range trees {
		if p.ID != "ferme" {
			break
		}
		tr := tr
		ds = append(ds, drawable{tr.Y, func() { drawTree(dst, tr.X, tr.Y, tr.Z) }})
	}
	sort.SliceStable(ds, func(i, j int) bool { return ds[i].y < ds[j].y })
	for _, d := range ds {
		d.fn()
	}
	// ce que demandent les machines : par-dessus les personnages (la caissière
	// cachait le miel du pain d'épices)
	for _, s := range p.Stations {
		if s.Built && s.Kind == sim.Machine {
			machineInputs(dst, s)
		}
	}
	// les drones volent au-dessus de tout
	for _, s := range p.Stations {
		for _, d := range s.Drones {
			lift := 0.0
			if d.Phase != 0 {
				lift = 40
			}
			ellipse(dst, d.Pos.X, d.Pos.Y+10, 26, 9, rgba(0x000000, 0.18))
			y := d.Pos.Y - 20 - lift + math.Sin(t*3+d.Pos.X)*3
			if d.N > 0 && d.Phase == 2 {
				line(dst, d.Pos.X, y+8, d.Pos.X, y+26, 2, rgb(0x444444))
				drawItem(dst, d.It, d.Pos.X, y+38, 14)
				if d.N > 1 {
					sayOutlined(dst, "×"+itoa(d.N), d.Pos.X+26, y+40, 15, 0, rgb(0xffffff), rgb(0x2b2b2b))
				}
			}
			drawDroneShape(dst, d.Pos.X, y, 1, t)
		}
	}
	if p.ID == "super" && g.sim.Cur == p {
		g.drawRayonBubble(dst, p) // pourquoi le rayon devant lequel on passe est vide
	}
	g.drawPetals(dst, p, tl, br)
	g.drawParticles(dst)
	g.drawGuide(dst, p)
}

type tree struct{ X, Y, Z float64 }

var trees = func() []tree {
	var ts []tree
	right := sim.FarmW + sim.FarmX0 + 40
	for i := -4; i < 14; i++ {
		ts = append(ts, tree{-40 - (sim.YardW - 560) + float64(i%2)*30, 120 + float64(i)*105, 0.8 + 0.3*math.Mod(float64(i)*0.37, 1)}) // (à gauche de la zone du joueur)
		ts = append(ts, tree{right - float64(i%2)*30, 140 + float64(i)*100, 0.8 + 0.3*math.Mod(float64(i)*0.61, 1)})
	}
	// en haut, au-dessus de la serre, du verger et de l'étang
	for i := 0; i < 40; i++ {
		ts = append(ts, tree{80 + float64(i)*170, sim.FarmTop + 70 - float64(i%2)*40, 0.9 + 0.3*math.Mod(float64(i)*0.53, 1)})
	}
	return ts
}()

// ---------------------------------------------------------------- champs

func (g *game) drawField(dst *ebiten.Image, s *sim.Station) {
	if s.Dredged {
		g.drawParc(dst, s)
		return
	}
	if s.Column {
		g.drawColumn(dst, s)
		return
	}
	b := s.Body
	if !s.Built {
		// terrain en friche
		roundRect(dst, b.X, b.Y, b.W, b.H, 20, rgba(0x6f9b4a, 0.55))
		return
	}
	roundRect(dst, b.X-6, b.Y-6, b.W+12, b.H+12, 24, rgb(0x8b6239))
	roundRect(dst, b.X, b.Y, b.W, b.H, 20, rgb(0xa47447))
	if s.Item == sim.Riz {
		// la rizière : de l'eau, avec des reflets
		roundRect(dst, b.X+8, b.Y+8, b.W-16, b.H-16, 16, rgb(0x6fb3c9))
		for k := 0; k < 6; k++ {
			x := b.X + 40 + math.Mod(float64(k)*137+g.t*8, b.W-80)
			line(dst, x, b.Y+40+float64(k%3)*90, x+30, b.Y+40+float64(k%3)*90, 2, rgba(0xffffff, 0.5))
		}
		// des canards qui barbotent entre les plants (pour faire joli)
		for k := 0; k < 4; k++ {
			ph := g.t*0.05 + float64(k)*0.27
			f := math.Mod(ph, 2)
			right := f < 1
			if !right {
				f = 2 - f
			}
			x := b.X + 50 + f*(b.W-100)
			y := b.Y + 105 + float64(k%2)*90 + math.Sin(g.t*1.3+float64(k))*4
			drawDuck(dst, x, y, g.t+float64(k), right)
		}
	} else {
		for r := 0; r < 3; r++ {
			y := b.Y + 60 + float64(r)*90
			roundRect(dst, b.X+18, y-14, b.W-36, 28, 14, rgb(0x8e623a))
		}
	}
	// emplacements encore vides : petites buttes de terre
	for i := len(s.Plants); i < len(s.Slots); i++ {
		q := s.Slots[i]
		ellipse(dst, q.X, q.Y, 22, 9, rgb(0x96683e))
	}
	for i, pl := range s.Plants {
		drawPlant(dst, s.Item, pl.P.X, pl.P.Y, pl.G, g.t+float64(i)*0.7, i)
	}
	// prochaine pousse à acheter
	if len(s.Plants) < len(s.Slots) {
		z := s.PlantZone()
		cost := s.PlantCost()
		c := uint32(0xffd23f)
		if g.sim.Money < 1 {
			c = 0xe0e0e0
		}
		roundRect(dst, z.X, z.Y, z.W, z.H, 12, rgba(0x000000, 0.25))
		if s.PlantPaid > 0 {
			f := s.PlantPaid / cost
			roundRect(dst, z.X, z.Y+z.H*(1-f), z.W, z.H*f, 12, rgba(0xffd23f, 0.5))
		}
		dashedRect(dst, z.X+3, z.Y+3, z.W-6, z.H-6, 4, 9, g.t*25, rgb(c))
		say(dst, "+1", z.C().X, z.Y+22, 20, true, 0, rgb(0xffffff))
		sayOutlined(dst, money(cost-s.PlantPaid), z.C().X, z.Y+z.H-20, 18, 0, rgb(0xffe066), rgb(0x3b2a10))
	}
	// petit panneau
	roundRect(dst, b.X+b.W/2-90, b.Y+b.H-6, 180, 36, 8, rgb(0x7a4a26))
	sayFit(dst, s.Name, b.X+b.W/2, b.Y+b.H+12, 20, true, 0, 170, rgb(0xffe7a8))
}

func drawPlant(dst *ebiten.Image, it sim.Item, x, y, gr, t float64, i int) {
	if drawExoticPlant(dst, it, x, y, gr, t) || drawDomainePlant(dst, it, x, y, gr, t) {
		return
	}
	sway := 0.0
	if gr >= 1 {
		sway = math.Sin(t*2) * 3
	}
	h := 8 + 42*gr
	ripe := gr >= 1
	ellipse(dst, x, y+4, 18*math.Max(0.4, gr), 6, rgba(0x000000, 0.18))
	switch it {
	case sim.Ble:
		c := lerpCol(0x7cc24a, 0xe9c24a, gr)
		for k := -2; k <= 2; k++ {
			tx := x + float64(k)*6 + sway
			line(dst, x+float64(k)*3, y, tx, y-h, 3, c)
			if gr > 0.6 {
				ellipse(dst, tx, y-h, 4, 9, lerpCol(0x9ccf5a, 0xf5d35a, gr))
			}
		}
	case sim.Mais:
		line(dst, x, y, x+sway, y-h*1.25, 6, rgb(0x4f9a34))
		line(dst, x, y-h*0.5, x-16, y-h*0.75, 4, rgb(0x5aa83a))
		line(dst, x, y-h*0.7, x+16+sway, y-h*0.95, 4, rgb(0x5aa83a))
		if ripe {
			ellipse(dst, x+7+sway*0.5, y-h*0.85, 7, 14, rgb(0xf6cf3b))
		}
	case sim.Legume:
		r := 6 + 14*gr
		circle(dst, x-6, y-r*0.8, r, rgb(0x3f9b35))
		circle(dst, x+7, y-r*0.9, r*0.85, rgb(0x52b246))
		if ripe {
			circle(dst, x-8+sway*0.3, y-r*0.7, 7, rgb(0xe23b2e))
			circle(dst, x+9, y-r*1.1, 6, rgb(0xe23b2e))
			polygon(dst, []V{{X: x - 2, Y: y - 4}, {X: x + 6, Y: y - 4}, {X: x + 2, Y: y + 8}}, rgb(0xf08a24))
		}
	case sim.Fleur:
		line(dst, x, y, x+sway, y-h, 3, rgb(0x3f9b35))
		ellipse(dst, x-6, y-h*0.4, 7, 3, rgb(0x52b246))
		if gr > 0.5 {
			cols := []uint32{0xff6fa8, 0xffd23f, 0xb07cff, 0xff8a3d, 0xffffff}
			c := rgb(cols[i%len(cols)])
			rr := 4 + 6*gr
			for k := 0; k < 5; k++ {
				a := float64(k)*2*math.Pi/5 + t*0.2
				circle(dst, x+sway+math.Cos(a)*rr, y-h+math.Sin(a)*rr, rr*0.75, c)
			}
			circle(dst, x+sway, y-h, rr*0.55, rgb(0xffd23f))
		}
	case sim.Canne:
		for k := -1; k <= 1; k++ {
			line(dst, x+float64(k)*7, y, x+float64(k)*9+sway, y-h*1.3, 5, lerpCol(0x6fb04a, 0xb5c94f, gr))
		}
	case sim.Fruit:
		// petit arbre fruitier
		rect(dst, x-3, y-h*0.7, 6, h*0.7, rgb(0x8a5a33))
		cr := 8 + 14*gr
		circle(dst, x+sway*0.3, y-h*0.8, cr, rgb(0x4aa043))
		circle(dst, x-cr*0.5, y-h*0.65, cr*0.7, rgb(0x3f8f3a))
		if ripe {
			cols := []uint32{0xe23b2e, 0xff9f1a, 0xffd23f}
			c := rgb(cols[i%3])
			circle(dst, x-6, y-h*0.8, 4, c)
			circle(dst, x+7, y-h*0.9, 4, c)
			circle(dst, x+2, y-h*0.65, 4, c)
		}
	case sim.CafeVert:
		// caféier : arbuste aux cerises rouges
		rect(dst, x-2, y-h*0.6, 4, h*0.6, rgb(0x6b4423))
		cr := 7 + 11*gr
		circle(dst, x+sway*0.3, y-h*0.75, cr, rgb(0x2f6a35))
		circle(dst, x-cr*0.6, y-h*0.55, cr*0.65, rgb(0x3a7a3a))
		circle(dst, x+cr*0.6, y-h*0.6, cr*0.6, rgb(0x3a7a3a))
		if ripe {
			for k := 0; k < 4; k++ {
				circle(dst, x-8+float64(k)*5, y-h*0.6-float64(k%2)*6, 3, rgb(0xc0392b))
			}
		}
	case sim.Cacao:
		rect(dst, x-3, y-h*0.8, 6, h*0.8, rgb(0x6b4423))
		cr := 8 + 12*gr
		circle(dst, x+sway*0.3, y-h*0.9, cr, rgb(0x2f7a35))
		if ripe {
			ellipse(dst, x-5, y-h*0.5, 3, 6, rgb(0xd9893a))
			ellipse(dst, x+6, y-h*0.55, 3, 6, rgb(0xb5651d))
		}
	}
	if ripe {
		// petit éclat pour dire « prêt »
		a := math.Max(0, math.Sin(t*3))
		circle(dst, x+12, y-h-6, 3*a, rgba(0xffffff, 0.8*a))
	}
}

func lerpCol(a, b uint32, t float64) colorRGBA {
	t = math.Max(0, math.Min(1, t))
	ca, cb := rgb(a), rgb(b)
	l := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*t) }
	return colorRGBA{l(ca.R, cb.R), l(ca.G, cb.G), l(ca.B, cb.B), 255}
}

// ---------------------------------------------------------------- zones

func zoneBox(dst *ebiten.Image, r sim.Rect, active bool, t float64, c uint32) {
	a := 0.18
	if active {
		a = 0.32 + 0.08*math.Sin(t*6)
	}
	roundRect(dst, r.X, r.Y, r.W, r.H, 16, rgba(c, a))
	roundStroke(dst, r.X+3, r.Y+3, r.W-6, r.H-6, 14, 4, rgba(c, 0.85))
}

// pileGrid dessine une pile d'objets dans une zone.
func pileGrid(dst *ebiten.Image, it sim.Item, n int, r sim.Rect) {
	cols := 4
	step := 26.0
	x0 := r.C().X - step*float64(cols-1)/2
	shown := min(n, 12)
	for i := 0; i < shown; i++ {
		c := i % cols
		l := i / cols
		drawItem(dst, it, x0+float64(c)*step, r.Y+r.H-24-float64(l)*16, 13)
	}
	if n > shown {
		// au-delà de 12, on affiche le nombre
		roundRect(dst, r.X+r.W-46, r.Y+4, 42, 24, 8, rgba(0x2b2140, 0.85))
		say(dst, itoa(n), r.X+r.W-25, r.Y+16, 16, true, 0, rgb(0xffffff))
	}
}

func (g *game) drawZones(dst *ebiten.Image, p *sim.Place, s *sim.Station) {
	pl := g.sim.Player
	in := func(r sim.Rect) bool { return g.sim.Cur == p && r.Has(pl.Pos) }
	switch s.Kind {
	case sim.Animal:
		zoneBox(dst, s.Zone, in(s.Zone), g.t, 0xffffff)
		pileGrid(dst, s.Item, s.Pile, s.Zone)
		if s.Pile == 0 {
			drawItem(dst, s.Item, s.Zone.C().X, s.Zone.C().Y, 20)
			rect(dst, s.Zone.X+8, s.Zone.Y+8, s.Zone.W-16, s.Zone.H-16, rgba(0xa47447, 0.35))
		}
	case sim.Machine:
		zoneBox(dst, s.Zone, in(s.Zone), g.t, 0x9fd3ff)
		zoneBox(dst, s.OutZone, in(s.OutZone), g.t, 0xffffff)
		// (les entrées, icône + quantité, sont dessinées par-dessus les personnages : machineInputs)
		pileGrid(dst, s.Item, s.Pile, s.OutZone)
		if s.Pile == 0 {
			say(dst, "SORTIE", s.OutZone.C().X, s.OutZone.C().Y, 15, true, 0, rgba(0x6b5a40, 0.7))
		}
		if s.Two { // deuxième sortie (les petits)
			zoneBox(dst, s.OutZone2, in(s.OutZone2), g.t, 0xffffff)
			pileGrid(dst, s.Item2, s.Pile2, s.OutZone2)
			if s.Pile2 == 0 {
				drawItem(dst, s.Item2, s.OutZone2.C().X, s.OutZone2.C().Y-8, 16)
				say(dst, "SORTIE", s.OutZone2.C().X, s.OutZone2.C().Y+22, 15, true, 0, rgba(0x6b5a40, 0.7))
			}
		}
	case sim.Shelf:
		zoneBox(dst, s.Zone, in(s.Zone), g.t, 0xfff3c4)
		drawItem(dst, s.Item, s.Zone.C().X, s.Zone.C().Y, 18)
	case sim.Store:
		zoneBox(dst, s.Zone, in(s.Zone), g.t, 0xffd6ec)
		drawItem(dst, s.Item, s.Zone.C().X-34, s.Zone.C().Y, 18)
		say(dst, itoa(s.Stock)+"/"+itoa(s.StockMax), s.Zone.C().X+22, s.Zone.C().Y, 22, true, 0, rgb(0x6a2d50))
		// la base prévue pour le drone (il viendra avec le fleuriste)
		pad := s.StorePad()
		dashedRect(dst, pad.X, pad.Y, pad.W, pad.H, 3, 10, 0, rgba(0x8a3d6a, 0.5))
		say(dst, "Drone : avec le fleuriste", pad.C().X, pad.C().Y, 13, false, 0, rgba(0x6a2d50, 0.8))
	case sim.Pallet:
		zoneBox(dst, s.Zone, in(s.Zone), g.t, 0xffffff)
		if sim.IsSeed(s.Item) {
			drawItem(dst, s.Item, s.Zone.X+22, s.Zone.C().Y, 18)
			n := p.Empties(s.Item)
			txt, c := "à planter : "+itoa(n), rgb(0x2f6a3a)
			if n == 0 {
				txt, c = "rien à planter", rgba(0x5a4a3a, 0.7)
			}
			sayFit(dst, txt, s.Zone.C().X+14, s.Zone.C().Y, 15, true, 0, s.Zone.W-44, c)
			break
		}
		if p.ID == "super" { // (le dessin est déjà sur la palette, juste au-dessus)
			say(dst, itoa(s.Pile), s.Zone.C().X, s.Zone.C().Y, 26, true, 0, rgb(0x3b3b3b))
			break
		}
		drawItem(dst, s.Item, s.Zone.C().X-30, s.Zone.C().Y, 18)
		say(dst, itoa(s.Pile), s.Zone.C().X+28, s.Zone.C().Y, 26, true, 0, rgb(0x3b3b3b))
	case sim.Dock:
		if g.sim.DockActive(s) {
			zoneBox(dst, s.Zone, in(s.Zone), g.t, 0x9fd3ff)
			say(dst, "QUAI DU CAMION", s.Zone.C().X, s.Zone.Y+22, 17, true, 0, rgb(0x2b4a66))
			dest := "→ " + placeShort(s.To)
			if m := sim.DockMult(g.sim.Place(s.To)); m > 1 {
				dest += " (1 objet = " + itoa(m) + ")"
			}
			say(dst, dest, s.Zone.C().X, s.Zone.Y+44, 15, false, 0, rgb(0x2b4a66))
		}
	case sim.Register:
		if s.Cashier == nil {
			zoneBox(dst, s.Zone, in(s.Zone), g.t, 0xffe066)
			lbl := "CAISSE"
			if p.ID == "ecole" {
				lbl = "ACCUEIL"
			}
			say(dst, lbl, s.Zone.C().X, s.Zone.C().Y, 22, true, 0, rgb(0x8a6a10))
		} else if p.PlayerHelps(s) {
			// le joueur à côté de la caissière : ils encaissent à deux
			sayOutlined(dst, "Tu l'aides : ça va plus vite !", s.Spot.X, s.Spot.Y-70, 17, 0, rgb(0xffffff), rgb(0x2f8a4a))
		}
	case sim.Trash:
		if s.Zone.W > 0 {
			zoneBox(dst, s.Zone, in(s.Zone), g.t, 0xc8c8c8)
		}
	case sim.OrderBox:
		if o := g.sim.Order; o != nil && o.Place() == p.ID {
			zoneBox(dst, s.Zone, in(s.Zone), g.t, 0xffd23f)
			say(dst, "LIVRER ICI", s.Zone.C().X, s.Zone.C().Y, 18, true, 0, rgb(0x8a6a10))
		}
	}
}

func (g *game) drawUnlock(dst *ebiten.Image, u *sim.Unlock) {
	r := u.Zone
	afford := g.sim.Money >= 1
	c := uint32(0xffd23f)
	if !afford {
		c = 0xd8d8d8
	}
	roundRect(dst, r.X, r.Y, r.W, r.H, 18, rgba(0x000000, 0.22))
	frac := u.Paid / u.Cost
	if frac > 0 {
		roundRect(dst, r.X, r.Y+r.H*(1-frac), r.W, r.H*frac, 18, rgba(0xffd23f, 0.45))
	}
	dashedRect(dst, r.X+4, r.Y+4, r.W-8, r.H-8, 5, 16, g.t*30, rgb(c))
	sayOutlined(dst, u.Name, r.C().X, r.Y+30, 21, 0, rgb(0xffffff), rgb(0x3b2a10))
	left := u.Cost - u.Paid
	coin(dst, r.C().X-textWidth(money(left), 28, true)/2-16, r.Y+r.H-38, 13)
	sayOutlined(dst, money(left), r.C().X+12, r.Y+r.H-38, 28, 0, rgb(0xffe066), rgb(0x3b2a10))
}

// ---------------------------------------------------------------- bâtiments

func building(dst *ebiten.Image, b sim.Rect, wall, roof uint32) {
	// ombre portée vers le bas à droite
	roundRect(dst, b.X+16, b.Y+40, b.W+6, b.H-14, 10, rgba(0x1d3a12, 0.18))
	roundRect(dst, b.X, b.Y+30, b.W, b.H-30, 8, rgb(wall))
	rect(dst, b.X, b.Y+30, 10, b.H-44, shade(wall, 1.12)) // (le coin éclairé)
	rect(dst, b.X, b.Y+b.H-14, b.W, 14, shade(wall, 0.8))
	// toit : versant éclairé à gauche, versant à l'ombre à droite
	m := b.X + b.W/2
	polygon(dst, []V{{X: b.X - 16, Y: b.Y + 50}, {X: b.X + 30, Y: b.Y - 30}, {X: m, Y: b.Y - 30}, {X: m, Y: b.Y + 50}}, shade(roof, 1.1))
	polygon(dst, []V{{X: m, Y: b.Y - 30}, {X: b.X + b.W - 30, Y: b.Y - 30}, {X: b.X + b.W + 16, Y: b.Y + 50}, {X: m, Y: b.Y + 50}}, rgb(roof))
	rect(dst, b.X+30, b.Y-34, b.W-60, 8, shade(roof, 1.25)) // (le faîtage)
	polygon(dst, []V{{X: b.X - 16, Y: b.Y + 50}, {X: b.X + b.W + 16, Y: b.Y + 50}, {X: b.X + b.W + 10, Y: b.Y + 60}, {X: b.X - 10, Y: b.Y + 60}}, shade(roof, 0.75))
}

func sign(dst *ebiten.Image, x, y float64, s string) {
	w := math.Max(110, textWidth(s, 20, true)+28)
	roundRect(dst, x-w/2, y-18, w, 36, 9, rgb(0x6e4322))
	roundRect(dst, x-w/2+3, y-15, w-6, 30, 7, rgb(0x8a5a33))
	say(dst, s, x, y, 20, true, 0, rgb(0xffeab0))
}

func (g *game) drawStation(dst *ebiten.Image, p *sim.Place, s *sim.Station) {
	if strings.HasPrefix(s.ID, "mur_") {
		return // les murs des salles sont dessinés avec le sol
	}
	if strings.HasPrefix(s.ID, "barriere_") { // (tous les espaces du supermarché, même les nouveaux)
		drawZoneBarrier(dst, s)
		return
	}
	if p.ID == "domaine" && s.Kind == sim.Machine {
		drawDomaineMachine(dst, s, g.t)
		sign(dst, s.Body.C().X, s.Body.Y-14, s.Name)
		return
	}
	if g.drawStage6(dst, p, s) {
		if s.Kind == sim.Machine {
			sign(dst, s.Body.C().X, s.Body.Y+48, s.Name)
		}
		return
	}
	b := s.Body
	t := g.t
	switch s.ID {
	case "moulin", "moulin_mais", "moulin_sucre":
		building(dst, b, 0xd8cbb5, 0x8a5a44)
		// ailes du moulin
		cx, cy := b.C().X, b.Y-10
		a0 := 0.0
		if s.Working {
			a0 = -t * 2.5 // (sens des aiguilles d'une montre inversé)
		} else {
			a0 = float64(int(t*0)) + 0.4
		}
		for k := 0; k < 4; k++ {
			a := a0 + float64(k)*math.Pi/2
			ex, ey := cx+math.Cos(a)*110, cy+math.Sin(a)*110
			line(dst, cx, cy, ex, ey, 8, rgb(0x6e4322))
			px, py := -math.Sin(a)*22, math.Cos(a)*22
			polygon(dst, []V{{X: cx + math.Cos(a)*30, Y: cy + math.Sin(a)*30}, {X: ex, Y: ey}, {X: ex + px, Y: ey + py}, {X: cx + math.Cos(a)*30 + px, Y: cy + math.Sin(a)*30 + py}}, rgba(0xf5efe0, 0.92))
		}
		circle(dst, cx, cy, 12, rgb(0x4a2f18))
		roundRect(dst, b.C().X-26, b.Y+b.H-70, 52, 60, 20, rgb(0x6e4322))
	case "four":
		building(dst, b, 0xc4553b, 0x5b4a44)
		rect(dst, b.X+b.W-70, b.Y-70, 36, 70, rgb(0x7a3a2a))
		if s.Working {
			for k := 0; k < 3; k++ {
				ph := math.Mod(t*0.8+float64(k)/3, 1)
				circle(dst, b.X+b.W-52+math.Sin(ph*6)*8, b.Y-80-ph*90, 12+ph*16, rgba(0xdddddd, 0.6*(1-ph)))
			}
		}
		// bouche du four
		roundRect(dst, b.C().X-50, b.Y+b.H-80, 100, 64, 30, rgb(0x3a2018))
		if s.Working {
			roundRect(dst, b.C().X-38, b.Y+b.H-60, 76, 40, 18, rgba(0xff8a2a, 0.6+0.3*math.Sin(t*9)))
		}
	case "porcherie":
		hb := sim.Rect{X: b.X, Y: b.Y, W: 120, H: b.H}
		building(dst, hb, 0xe8a0a8, 0x8a4a3a)
		roundRect(dst, hb.C().X-24, hb.Y+hb.H-60, 48, 50, 8, rgb(0x7a3a2a))
		// enclos boueux
		ex := b.X + 130
		rect(dst, ex, b.Y+60, b.W-130, b.H-60, rgb(0x9a7a52))
		ellipse(dst, ex+60, b.Y+125, 40, 16, rgb(0x7a5a3a))
		for x := ex; x <= b.X+b.W; x += 26 {
			rect(dst, x-3, b.Y+52, 6, 30, rgb(0xc9a36a))
			rect(dst, x-3, b.Y+b.H-24, 6, 26, rgb(0xc9a36a))
		}
		for k := 0; k < 2; k++ {
			drawPig(dst, ex+40+float64(k)*60+math.Sin(t*0.8+float64(k)*2)*8, b.Y+110+float64(k)*14, t+float64(k))
		}
	case "mare":
		// cabane des canards et leur mare
		hb := sim.Rect{X: b.X, Y: b.Y + 30, W: 100, H: b.H - 30}
		building(dst, hb, 0xd9c08f, 0x3f8a6a)
		roundRect(dst, hb.C().X-16, hb.Y+hb.H-44, 32, 36, 14, rgb(0x5a3a22))
		px, py := b.X+180, b.Y+b.H-58
		ellipse(dst, px, py+6, 82, 42, rgb(0x7a9a52))
		ellipse(dst, px, py, 76, 36, rgb(0x5aa0e0))
		ellipse(dst, px-20, py-10, 30, 8, rgba(0xffffff, 0.25))
		for k := 0; k < 3; k++ {
			a := t*0.5 + float64(k)*2.1
			drawDuck(dst, px+math.Cos(a)*44, py+math.Sin(a)*16, t+float64(k), math.Sin(a) > 0)
		}
	case "chevrerie", "bergerie":
		goat := s.ID == "chevrerie"
		hb := sim.Rect{X: b.X, Y: b.Y, W: 120, H: b.H}
		wall, roof := uint32(0xc9a06a), uint32(0x6a4a2a)
		if !goat {
			wall, roof = 0xe8dcc0, 0x7a5a3a
		}
		building(dst, hb, wall, roof)
		roundRect(dst, hb.C().X-24, hb.Y+hb.H-60, 48, 50, 8, rgb(0x5a3a22))
		ex := b.X + 130
		rect(dst, ex, b.Y+60, b.W-130, b.H-60, rgb(0x9cc46a))
		for x := ex; x <= b.X+b.W; x += 26 {
			rect(dst, x-3, b.Y+52, 6, 30, rgb(0xa47447))
			rect(dst, x-3, b.Y+b.H-24, 6, 26, rgb(0xa47447))
		}
		for k := 0; k < 2; k++ {
			x := ex + 40 + float64(k)*62 + math.Sin(t*0.7+float64(k)*2)*6
			y := b.Y + 112 + float64(k)*12
			if goat {
				drawGoat(dst, x, y, t+float64(k))
			} else {
				drawSheep(dst, x, y, t+float64(k))
			}
		}
	case "torrefacteur":
		rect(dst, b.X+30, b.Y-40, 32, 60, rgb(0x4a2e1e)) // cheminée (à gauche : à droite, la case du champ)
		building(dst, b, 0x8a5a3a, 0x3a2418)
		if s.Working {
			for k := 0; k < 3; k++ {
				ph := math.Mod(t*0.7+float64(k)/3, 1)
				circle(dst, b.X+46+math.Sin(ph*6)*8, b.Y-50-ph*70, 10+ph*14, rgba(0x8a7a6a, 0.55*(1-ph)))
			}
		}
		// le tambour qui tourne
		circle(dst, b.C().X, b.Y+b.H-52, 40, rgb(0x5a5a5a))
		circle(dst, b.C().X, b.Y+b.H-52, 32, rgb(0x7a7a7a))
		a := 0.0
		if s.Working {
			a = t * 3
		}
		line(dst, b.C().X, b.Y+b.H-52, b.C().X+math.Cos(a)*30, b.Y+b.H-52+math.Sin(a)*30, 5, rgb(0x3a3a3a))
		drawItem(dst, sim.CafeGrains, b.X+40, b.Y+b.H-40, 18)
	case "moulin_cafe":
		building(dst, b, 0xe8d6b8, 0x6a3a1e)
		// un grand moulin à café, avec sa manivelle
		cx, cy := b.C().X, b.Y+b.H-60
		roundRect(dst, cx-34, cy-10, 68, 60, 6, rgb(0x8a5a33))
		polygon(dst, []V{{X: cx - 26, Y: cy - 10}, {X: cx + 26, Y: cy - 10}, {X: cx + 14, Y: cy - 40}, {X: cx - 14, Y: cy - 40}}, rgb(0xb0b0b0))
		a := 0.6
		if s.Working {
			a = t * 4
		}
		line(dst, cx, cy-44, cx+math.Cos(a)*30, cy-44+math.Sin(a)*10, 4, rgb(0x3a3a3a))
		circle(dst, cx+math.Cos(a)*30, cy-44+math.Sin(a)*10, 5, rgb(0x6a3a1e))
		rect(dst, cx-10, cy+20, 20, 12, rgb(0x5a3220))
		drawItem(dst, sim.CafeMoulu, b.X+40, b.Y+b.H-40, 18)
	case "patisserie":
		building(dst, b, 0xffd6e4, 0xe0567f)
		for k := 0; k < 7; k++ {
			c := uint32(0xffffff)
			if k%2 == 0 {
				c = 0xe0567f
			}
			x := b.X + float64(k)*b.W/7
			polygon(dst, []V{{X: x, Y: b.Y + 70}, {X: x + b.W/7, Y: b.Y + 70}, {X: x + b.W/7, Y: b.Y + 96}, {X: x + b.W/14, Y: b.Y + 104}, {X: x, Y: b.Y + 96}}, rgb(c))
		}
		roundRect(dst, b.C().X-30, b.Y+b.H-62, 60, 52, 10, rgb(0x9a5a6a))
		drawItem(dst, sim.Gateau, b.X+50, b.Y+130, 22)
	case "poulailler":
		hb := sim.Rect{X: b.X, Y: b.Y, W: 120, H: b.H}
		building(dst, hb, 0xc99a5e, 0xa0412d)
		circle(dst, hb.C().X, hb.Y+110, 16, rgb(0x3a2a1a))
		// enclos
		ex := b.X + 130
		rect(dst, ex, b.Y+60, b.W-130, b.H-60, rgb(0xd7c189))
		for x := ex; x <= b.X+b.W; x += 26 {
			rect(dst, x-3, b.Y+52, 6, 30, rgb(0xf3ead8))
			rect(dst, x-3, b.Y+b.H-24, 6, 26, rgb(0xf3ead8))
		}
		for k := 0; k < 3; k++ {
			cx := ex + 30 + float64(k)*38 + math.Sin(t*1.3+float64(k)*2)*10
			cy := b.Y + 110 + math.Cos(t*0.9+float64(k))*16
			drawChicken(dst, cx, cy, t+float64(k))
		}
	case "etable":
		building(dst, b, 0xb8352c, 0x5a4a44)
		roundRect(dst, b.C().X-48, b.Y+b.H-104, 96, 94, 8, rgb(0xf3ead8))
		line(dst, b.C().X-44, b.Y+b.H-100, b.C().X+44, b.Y+b.H-14, 8, rgb(0xb8352c))
		line(dst, b.C().X+44, b.Y+b.H-100, b.C().X-44, b.Y+b.H-14, 8, rgb(0xb8352c))
		drawCow(dst, b.X+58, b.Y+b.H-42, t)
	case "ruche":
		ellipse(dst, b.C().X, b.Y+b.H, b.W*0.55, 16, rgba(0x000000, 0.15))
		rect(dst, b.X, b.Y+60, b.W, b.H-60, rgb(0x9bc76a))
		for k := 0; k < 6; k++ { // (six ruches)
			x := b.X + 30 + float64(k)*76
			y := b.Y + 80
			rect(dst, x, y, 60, 70, rgb(0xf2c14e))
			rect(dst, x, y+22, 60, 6, rgb(0xd09a2a))
			rect(dst, x, y+46, 60, 6, rgb(0xd09a2a))
			rect(dst, x-6, y-10, 72, 14, rgb(0xb5651d))
			rect(dst, x+24, y+56, 12, 8, rgb(0x3a2a1a))
		}
		for k := 0; k < 8; k++ {
			bx := b.C().X + math.Sin(t*2.1+float64(k)*1.7)*130
			by := b.Y + 60 + math.Cos(t*2.7+float64(k)*2.3)*40
			ellipse(dst, bx, by, 6, 4, rgb(0xffd23f))
			line(dst, bx-2, by-4, bx-2, by+4, 2, rgb(0x2a2a2a))
			ellipse(dst, bx, by-5, 4, 3, rgba(0xffffff, 0.7))
		}
	case "local":
		roundRect(dst, b.X, b.Y, b.W, b.H, 10, rgb(0x8a94a0))
		roundRect(dst, b.X+10, b.Y+10, b.W-20, 30, 6, rgb(0x6a7480))
		say(dst, "PERSONNEL", b.C().X, b.Y+26, 18, true, 0, rgb(0xffffff))
		roundRect(dst, b.C().X-30, b.Y+b.H-90, 60, 86, 6, rgb(0x5a6470))
		circle(dst, b.C().X+18, b.Y+b.H-46, 4, rgb(0xffd23f))
	case "accueil":
		g.drawSchoolDesk(dst, b)
	case "caisse", "caisse2", "caisse3", "caisse4", "caisse5", "caisse6", "caisse7", "caisse8", "caisse9", "caisse10", "caisse11", "caisse12": // (les 12 caisses du supermarché)
		ellipse(dst, b.C().X, b.Y+b.H, b.W*0.55, 14, rgba(0x000000, 0.18))
		roundRect(dst, b.X, b.Y, b.W, b.H, 10, rgb(0x8a5a33))
		roundRect(dst, b.X, b.Y, b.W, 22, 8, rgb(0xb07d4a))
		roundRect(dst, b.X+120, b.Y-28, 64, 42, 6, rgb(0x444a52))
		rect(dst, b.X+128, b.Y-22, 48, 14, rgb(0x9fe0a0))
		say(dst, "CAISSE", b.X+60, b.Y+44, 18, true, 0, rgb(0xffe7a8))
	case "commande":
		g.drawOrderBox(dst, s)
	case "etang":
		drawPond(dst, b, g.t)
	case "reserve_fleurs":
		drawFlowerStore(dst, s)
	case "vestiaire":
		g.drawVestiaire(dst, p, s)
	case "poubelle":
		// local à compost : bac en lattes de bois sous un petit toit
		ellipse(dst, b.C().X, b.Y+b.H, b.W*0.6, 12, rgba(0x000000, 0.2))
		rect(dst, b.X+4, b.Y-28, 6, b.H+28, rgb(0x6e4322))
		rect(dst, b.X+b.W-10, b.Y-28, 6, b.H+28, rgb(0x6e4322))
		// tas de compost
		ellipse(dst, b.C().X, b.Y+26, b.W*0.42, 20, rgb(0x5a3a22))
		circle(dst, b.C().X-14, b.Y+18, 5, rgb(0x7ed957))
		circle(dst, b.C().X+10, b.Y+14, 4, rgb(0xf08a24))
		circle(dst, b.C().X+2, b.Y+24, 4, rgb(0xe8c26a))
		// lattes
		for k := 0; k < 3; k++ {
			y := b.Y + 30 + float64(k)*17
			roundRect(dst, b.X, y, b.W, 12, 3, rgb(0xa0703f))
			line(dst, b.X+3, y+3, b.X+b.W-3, y+3, 2, rgb(0xb98550))
		}
		// toit
		polygon(dst, []V{{X: b.X - 12, Y: b.Y - 18}, {X: b.C().X, Y: b.Y - 44}, {X: b.X + b.W + 12, Y: b.Y - 18}, {X: b.X + b.W + 12, Y: b.Y - 10}, {X: b.X - 12, Y: b.Y - 10}}, rgb(0x4f7a3a))
		sign(dst, b.C().X, b.Y+b.H+24, "Compost")
	default:
		if p.ID == "ecole" && s.Kind == sim.Animal {
			drawSupplyShelf(dst, s)
			return
		}
		switch s.Kind {
		case sim.Shelf:
			if p.ID == "super" {
				g.drawRayon(dst, p, s)
			} else if s.Market {
				g.drawMarketStall(dst, s)
			} else {
				g.drawShelf(dst, s)
			}
		case sim.Pallet:
			if p.ID == "port" {
				drawBasket(dst, s)
			} else {
				drawPallet(dst, s)
			}
		case sim.Dock:
			g.drawDock(dst, s)
		case sim.Machine:
			drawWorkshopMachine(dst, s, g.t)
		}
	}
	if s.Kind == sim.Machine || s.Kind == sim.Animal {
		sign(dst, b.C().X, b.Y+48, s.Name) // (plus de niveaux : seuls les points de l'école les accélèrent)
	}
}

func (g *game) drawShelf(dst *ebiten.Image, s *sim.Station) {
	b := s.Body
	ellipse(dst, b.C().X, b.Y+b.H, b.W*0.55, 12, rgba(0x000000, 0.18))
	roundRect(dst, b.X, b.Y, b.W, b.H, 8, rgb(0x9a6b3e))
	roundRect(dst, b.X, b.Y, b.W, 26, 8, rgb(0xc9935a))
	// marchandise posée dessus
	for i := 0; i < min(s.Stock, 18); i++ {
		c := i % 6
		l := i / 6
		drawItem(dst, s.Item, b.X+18+float64(c)*23, b.Y+12-float64(l)*12, 12)
	}
	if s.Stock > 18 {
		roundRect(dst, b.X+b.W-30, b.Y+30, 40, 22, 7, rgba(0x2b2140, 0.85))
		say(dst, itoa(s.Stock), b.X+b.W-10, b.Y+41, 15, true, 0, rgb(0xffffff))
	}
	// étiquette de prix
	price := money(sim.Items[s.Item].Price * g.sim.PriceMult())
	roundRect(dst, b.C().X-40, b.Y+34, 80, 26, 6, rgb(0xfff8e6))
	say(dst, price, b.C().X, b.Y+47, 17, true, 0, rgb(0x4a3a20))
	if s.Stock == 0 && math.Mod(g.t, 1) < 0.6 {
		roundRect(dst, b.C().X-34, b.Y-30, 68, 26, 6, rgb(0xe74c3c))
		say(dst, "VIDE", b.C().X, b.Y-17, 17, true, 0, rgb(0xffffff))
	}
}

func drawChicken(dst *ebiten.Image, x, y, t float64) {
	ellipse(dst, x, y+10, 12, 4, rgba(0x000000, 0.2))
	ellipse(dst, x, y, 13, 10, rgb(0xffffff))
	circle(dst, x+10, y-8, 7, rgb(0xffffff))
	polygon(dst, []V{{X: x + 16, Y: y - 9}, {X: x + 22, Y: y - 7}, {X: x + 16, Y: y - 5}}, rgb(0xf5a623))
	circle(dst, x+10, y-15, 3.5, rgb(0xe23b2e))
	circle(dst, x+12, y-9, 1.6, rgb(0x222222))
}

func drawDuck(dst *ebiten.Image, x, y, t float64, right bool) {
	d := 1.0
	if !right {
		d = -1
	}
	ellipse(dst, x, y+6, 14, 4, rgba(0x2a5a8a, 0.35))
	ellipse(dst, x, y, 14, 9, rgb(0xffffff))
	circle(dst, x+d*10, y-9, 7, rgb(0x2f8a4a))
	polygon(dst, []V{{X: x + d*15, Y: y - 10}, {X: x + d*23, Y: y - 8}, {X: x + d*15, Y: y - 6}}, rgb(0xf5a623))
	circle(dst, x+d*12, y-11, 1.6, rgb(0x222222))
	rect(dst, x+d*4-3, y-4, 6, 3, rgb(0xffffff))
}

func drawGoat(dst *ebiten.Image, x, y, t float64) {
	bob := math.Sin(t*1.5) * 1.5
	ellipse(dst, x, y+18, 26, 6, rgba(0x000000, 0.2))
	for _, dx := range []float64{-16, -8, 8, 16} {
		rect(dst, x+dx-3, y+4, 6, 14, rgb(0xd8cbb5))
	}
	ellipse(dst, x, y, 24, 13, rgb(0xf3ead8))
	ellipse(dst, x+24, y-10+bob, 9, 8, rgb(0xf3ead8))
	line(dst, x+22, y-17+bob, x+18, y-27+bob, 3, rgb(0x8a7a6a))
	line(dst, x+27, y-17+bob, x+29, y-27+bob, 3, rgb(0x8a7a6a))
	line(dst, x+26, y-3+bob, x+27, y+4+bob, 3, rgb(0xd8cbb5)) // barbiche
	circle(dst, x+26, y-12+bob, 1.6, rgb(0x222222))
}

func drawSheep(dst *ebiten.Image, x, y, t float64) {
	bob := math.Sin(t*1.2) * 1.5
	ellipse(dst, x, y+18, 26, 6, rgba(0x000000, 0.2))
	for _, dx := range []float64{-14, -6, 8, 15} {
		rect(dst, x+dx-3, y+6, 6, 12, rgb(0x3a3a3a))
	}
	for k := 0; k < 5; k++ {
		circle(dst, x-14+float64(k)*7, y-4+math.Sin(float64(k))*3, 10, rgb(0xf6f3ea))
	}
	ellipse(dst, x, y+2, 20, 10, rgb(0xf6f3ea))
	ellipse(dst, x+22, y-6+bob, 8, 9, rgb(0x3a3a3a))
	circle(dst, x+24, y-8+bob, 1.6, rgb(0xffffff))
}

func drawCow(dst *ebiten.Image, x, y, t float64) {
	bob := math.Sin(t*1.2) * 2
	ellipse(dst, x, y+26, 46, 10, rgba(0x000000, 0.2))
	for _, dx := range []float64{-28, -14, 16, 30} {
		rect(dst, x+dx-4, y+4, 8, 22, rgb(0xeeeeee))
	}
	ellipse(dst, x, y, 44, 24, rgb(0xffffff))
	circle(dst, x-14, y-6, 10, rgb(0x2a2a2a))
	circle(dst, x+16, y+6, 8, rgb(0x2a2a2a))
	ellipse(dst, x+44, y-12+bob, 16, 14, rgb(0xffffff))
	ellipse(dst, x+52, y-6+bob, 9, 7, rgb(0xf3b7b7))
	circle(dst, x+44, y-16+bob, 2.2, rgb(0x222222))
	line(dst, x+36, y-24+bob, x+30, y-32+bob, 4, rgb(0xd9c08f))
}

// ---------------------------------------------------------------- personnages

type look struct {
	skin, hair, shirt, pants uint32
	hat                      int // 0 rien, 1 chapeau de paille, 2 casquette, 3 toque, 4 masque de héros
	hatC                     uint32
	long                     bool
	// ennui
	sit    bool // assis sur une botte de foin
	shut   bool // yeux fermés
	mouth  int  // 0 rien, 1 grande bouche (bâillement), 2 petit « o » (siffle)
	armsUp bool // s'étire
	crate  bool // assis sur une caisse (ailleurs qu'à la ferme)
	// tenue de travail
	apron     uint32 // tablier (0 : aucun)
	stripe    uint32 // rayures du tablier
	flour     bool   // taches de farine
	vest      uint32 // gilet sans manches
	badge     bool   // badge sur la poitrine
	whistle   bool   // sifflet au cou
	boots     uint32 // bottes en caoutchouc (0 : chaussures)
	mariniere bool   // chemise rayée bleu et blanc
}

// staffLook : la tenue du personnel selon le lieu.
func staffLook(place string, kind sim.AgentKind, n int) look {
	if kind == sim.CashierAgent {
		lk := lookFrom(31337)
		lk.shirt, lk.long, lk.hair = 0xd9534f, true, 0xd9a441
		switch place {
		case "boulangerie":
			lk.shirt, lk.apron, lk.flour = 0xc0392b, 0xfbe9d0, true
			lk.hat = 6
		case "glacier":
			lk.shirt, lk.apron, lk.stripe = 0x48c9b0, 0xffffff, 0xff8fb8
			lk.hat, lk.hatC = 7, 0xff8fb8
		case "cremerie":
			lk.shirt, lk.apron = 0x7fb3d5, 0xffffff
			lk.hat = 6
		case "boucherie":
			lk.shirt, lk.apron, lk.stripe = 0xb03a3a, 0xffffff, 0xb03a3a
		case "super":
			lk.badge = true
		case "port": // marinière, tablier ciré
			lk.shirt, lk.mariniere, lk.apron = 0xffffff, true, 0x1f5fa8
		case "domaine":
			lk.shirt, lk.apron = 0x7a2a4a, 0x8a5a33
		case "patisserie":
			lk.shirt, lk.apron = 0xe0567f, 0xffffff
			lk.hat = 6
		}
		return lk
	}
	lk := lookFrom(n*7919 + 13)
	switch place {
	case "boulangerie": // veste blanche, tablier fariné, charlotte
		lk.shirt, lk.pants = 0xf7f3ea, 0x4a4a4a
		lk.apron, lk.flour = 0xe2cca0, true
		lk.hat = 6
	case "cremerie": // blouse blanche, tablier rayé bleu, charlotte
		lk.shirt, lk.pants = 0xffffff, 0x34495e
		lk.apron, lk.stripe = 0x7fb3d5, 0xffffff
		lk.hat = 6
	case "glacier": // tablier rayé rose et calot
		lk.shirt, lk.pants = 0xff9ec7, 0x5d6d7e
		lk.apron, lk.stripe = 0xffffff, 0xff8fb8
		lk.hat, lk.hatC = 7, 0xff8fb8
	case "super": // polo bleu, gilet orange, badge
		lk.shirt, lk.pants = 0x2f6db5, 0x34495e
		lk.vest, lk.badge = 0xf39c12, true
		lk.hat, lk.hatC = 2, 0x2f6db5
	case "exotique": // chemise à fleurs, chapeau de paille
		lk.shirt, lk.pants = 0x1fa37a, 0xe9d7a8
		lk.hat, lk.hatC = 2, 0xe8c86a
	case "patisserie": // veste blanche, tablier rose, toque
		lk.shirt, lk.pants = 0xffffff, 0x4a4a4a
		lk.apron, lk.stripe = 0xffc2d4, 0xffffff
		lk.hat = 6
	case "domaine": // vigneron : chemise bordeaux, tablier de cuir, chapeau
		lk.shirt, lk.pants = 0x7a2a4a, 0x3b3b4a
		lk.apron = 0x8a5a33
		lk.hat, lk.hatC = 2, 0x6e4322
	case "elevage": // salopette, bottes, casquette
		lk.shirt, lk.pants = 0x8a5a2b, 0x3b5998
		lk.hat, lk.hatC = 2, 0x8a5a2b
	case "boucherie": // veste blanche, tablier rayé rouge, calot
		lk.shirt, lk.pants = 0xffffff, 0x2b2b2b
		lk.apron, lk.stripe = 0xffffff, 0xb03a3a
		lk.hat, lk.hatC = 7, 0xb03a3a
	case "port": // poissonnier : marinière, grand tablier ciré bleu, bottes jaunes, bonnet
		lk.shirt, lk.pants, lk.mariniere = 0xffffff, 0x1f3a5f, true
		lk.apron, lk.boots = 0x1f5fa8, 0xf2c230
		lk.hat, lk.hatC = 2, 0x1f3a5f
	case "ecole": // surveillant : gilet et sifflet
		lk.shirt, lk.pants = 0x8e6c8a, 0x3b3b4a
		lk.whistle = true
	default: // ferme
		lk.shirt, lk.pants = 0x2e8b57, 0x3b5998
		lk.hat, lk.hatC = 2, 0x2e8b57
	}
	return lk
}

// staffPlace : le lieu affiché (pour les icônes d'embauche).
var staffPlace = "ferme"

var skins = []uint32{0xf6d2b4, 0xe8b996, 0xc68a62, 0x8d5a3b, 0xf1c9a5}
var hairs = []uint32{0x3a2a1a, 0x6b4423, 0xd9a441, 0x1e1e1e, 0xa0522d, 0xcfcfcf}
var shirts = []uint32{0x4a90d9, 0xe67e22, 0x9b59b6, 0x1abc9c, 0xe74c3c, 0xf1c40f, 0x2ecc71, 0xff7eb6, 0x34495e}

func lookFrom(n int) look {
	return look{
		skin:  skins[n%len(skins)],
		hair:  hairs[(n/7)%len(hairs)],
		shirt: shirts[(n/43)%len(shirts)],
		pants: []uint32{0x34495e, 0x5d4037, 0x2c3e50, 0x7f8c8d}[(n/389)%4],
		long:  (n/1013)%2 == 0,
	}
}

func drawPerson(dst *ebiten.Image, pos, face V, walkT float64, moving bool, lk look, carry []sim.Item) {
	x, y := pos.X, pos.Y
	sw := 0.0
	if moving {
		sw = math.Sin(walkT * 12)
	}
	bob := math.Abs(sw) * 2
	if lk.sit {
		ellipse(dst, x, y, 34, 9, rgba(0x000000, 0.22))
		if lk.crate {
			// une caisse en bois
			roundRect(dst, x-30, y-34, 60, 32, 4, rgb(0xb5835a))
			rect(dst, x-30, y-24, 60, 3, rgb(0x8a5f3c))
			rect(dst, x-30, y-13, 60, 3, rgb(0x8a5f3c))
			rect(dst, x-30, y-34, 5, 32, rgb(0x8a5f3c))
			rect(dst, x+25, y-34, 5, 32, rgb(0x8a5f3c))
		} else {
			// botte de foin derrière les jambes
			roundRect(dst, x-32, y-34, 64, 32, 7, rgb(0xe3c25a))
			rect(dst, x-32, y-26, 64, 3, rgb(0xc9a43c))
			rect(dst, x-32, y-14, 64, 3, rgb(0xc9a43c))
			rect(dst, x-20, y-34, 4, 32, rgb(0xa5822a))
			rect(dst, x+16, y-34, 4, 32, rgb(0xa5822a))
		}
		// cuisses vers l'avant, jambes qui pendent
		roundRect(dst, x-13, y-30, 26, 10, 4, rgb(lk.pants))
		rect(dst, x-11, y-22, 9, 14, rgb(lk.pants))
		rect(dst, x+2, y-22, 9, 14, rgb(lk.pants))
		roundRect(dst, x-13, y-10, 12, 7, 3, rgb(0x3a2a1a))
		roundRect(dst, x+1, y-10, 12, 7, 3, rgb(0x3a2a1a))
		y += 6 // le haut du corps descend un peu
	} else {
		ellipse(dst, x, y, 22, 8, rgba(0x000000, 0.22))
		// jambes
		rect(dst, x-11, y-24, 9, 22+sw*3, rgb(lk.pants))
		rect(dst, x+2, y-24, 9, 22-sw*3, rgb(lk.pants))
		roundRect(dst, x-13, y-5+sw*3, 12, 7, 3, rgb(0x3a2a1a))
		roundRect(dst, x+1, y-5-sw*3, 12, 7, 3, rgb(0x3a2a1a))
	}
	if lk.boots != 0 && !lk.sit { // bottes en caoutchouc
		rect(dst, x-12, y-16, 11, 14+sw*3, rgb(lk.boots))
		rect(dst, x+1, y-16, 11, 14-sw*3, rgb(lk.boots))
	}
	// corps
	roundRect(dst, x-17, y-58-bob, 34, 38, 12, rgb(lk.shirt))
	if lk.mariniere {
		for i := 0; i < 5; i++ {
			rect(dst, x-16, y-54-bob+float64(i)*7, 32, 3, rgb(0x1f3a7a))
		}
	}
	back := face.Y < -0.5
	if lk.vest != 0 {
		roundRect(dst, x-17, y-56-bob, 11, 34, 5, rgb(lk.vest))
		roundRect(dst, x+6, y-56-bob, 11, 34, 5, rgb(lk.vest))
		if back {
			roundRect(dst, x-17, y-56-bob, 34, 34, 8, rgb(lk.vest))
		}
		rect(dst, x-17, y-40-bob, 11, 3, rgb(0xfdfdfd))
		rect(dst, x+6, y-40-bob, 11, 3, rgb(0xfdfdfd))
	}
	if lk.apron != 0 {
		if back {
			// seulement le nœud dans le dos
			rect(dst, x-17, y-38-bob, 34, 3, rgb(lk.apron))
			circle(dst, x-3, y-37-bob, 3.5, rgb(lk.apron))
			circle(dst, x+3, y-37-bob, 3.5, rgb(lk.apron))
		} else {
			line(dst, x-9, y-58-bob, x-6, y-52-bob, 2, rgb(lk.apron))
			line(dst, x+9, y-58-bob, x+6, y-52-bob, 2, rgb(lk.apron))
			roundRect(dst, x-10, y-53-bob, 20, 16, 3, rgb(lk.apron))
			roundRect(dst, x-15, y-40-bob, 30, 26, 5, rgb(lk.apron))
			if lk.stripe != 0 {
				for i := 0; i < 4; i++ {
					rect(dst, x-11+float64(i)*7, y-53-bob, 3, 37, rgb(lk.stripe))
				}
			}
			rect(dst, x-8, y-32-bob, 16, 7, shade(lk.apron, 0.9)) // poche
			if lk.flour {
				fl := rgba(0xffffff, 0.95)
				circle(dst, x-6, y-46-bob, 3, fl)
				circle(dst, x+5, y-36-bob, 4, fl)
				circle(dst, x+8, y-38-bob, 2.2, fl)
				circle(dst, x-9, y-22-bob, 3.2, fl)
				circle(dst, x+2, y-20-bob, 2.4, fl)
				ellipse(dst, x-2, y-28-bob, 5, 2.5, fl)
			}
		}
	}
	if lk.badge && !back {
		rect(dst, x+4, y-50-bob, 9, 6, rgb(0xffffff))
		rect(dst, x+4, y-50-bob, 9, 2, rgb(0xd9534f))
	}
	if lk.whistle && !back {
		line(dst, x-7, y-58-bob, x, y-44-bob, 2, rgb(0xe74c3c))
		line(dst, x+7, y-58-bob, x, y-44-bob, 2, rgb(0xe74c3c))
		roundRect(dst, x-4, y-46-bob, 9, 6, 3, rgb(0xbdc3c7))
	}
	// bras
	if lk.armsUp {
		// s'étire : les bras au-dessus de la tête
		rect(dst, x-25, y-90-bob, 8, 36, rgb(lk.shirt))
		rect(dst, x+17, y-90-bob, 8, 36, rgb(lk.shirt))
		circle(dst, x-21, y-93-bob, 5, rgb(lk.skin))
		circle(dst, x+21, y-93-bob, 5, rgb(lk.skin))
	} else if len(carry) > 0 {
		rect(dst, x-21, y-50-bob, 8, 18, rgb(lk.shirt))
		rect(dst, x+13, y-50-bob, 8, 18, rgb(lk.shirt))
		circle(dst, x-14, y-34-bob, 5, rgb(lk.skin))
		circle(dst, x+14, y-34-bob, 5, rgb(lk.skin))
	} else {
		rect(dst, x-23, y-54-bob-sw*2, 8, 24, rgb(lk.shirt))
		rect(dst, x+15, y-54-bob+sw*2, 8, 24, rgb(lk.shirt))
		circle(dst, x-19, y-29-bob-sw*2, 5, rgb(lk.skin))
		circle(dst, x+19, y-29-bob+sw*2, 5, rgb(lk.skin))
	}
	// tête
	hy := y - 74 - bob
	if lk.long {
		roundRect(dst, x-16, hy-8, 32, 34, 12, rgb(lk.hair))
	}
	circle(dst, x, hy, 16, rgb(lk.skin))
	// cheveux
	if face.Y < -0.5 {
		circle(dst, x, hy, 16, rgb(lk.hair))
	} else {
		polygon(dst, []V{{X: x - 16, Y: hy - 2}, {X: x - 14, Y: hy - 12}, {X: x, Y: hy - 18}, {X: x + 14, Y: hy - 12}, {X: x + 16, Y: hy - 2}, {X: x + 8, Y: hy - 8}, {X: x - 8, Y: hy - 8}}, rgb(lk.hair))
		ex := face.X * 5
		if lk.shut {
			rect(dst, x-9+ex, hy+1, 6, 2, rgb(0x2a2a2a))
			rect(dst, x+3+ex, hy+1, 6, 2, rgb(0x2a2a2a))
		} else {
			circle(dst, x-6+ex, hy+2, 2.4, rgb(0x2a2a2a))
			circle(dst, x+6+ex, hy+2, 2.4, rgb(0x2a2a2a))
		}
		switch lk.mouth {
		case 1:
			ellipse(dst, x+ex, hy+10, 5, 4, rgb(0x6b1f1f))
		case 2:
			circle(dst, x+ex+2, hy+9, 2.6, rgb(0x6b1f1f))
		}
	}
	switch lk.hat {
	case 1: // chapeau de paille
		ellipse(dst, x, hy-10, 28, 9, rgb(lk.hatC))
		roundRect(dst, x-13, hy-26, 26, 18, 8, rgb(lk.hatC))
		rect(dst, x-13, hy-14, 26, 4, rgb(0xc0392b))
	case 2: // casquette
		roundRect(dst, x-15, hy-20, 30, 14, 7, rgb(lk.hatC))
		ellipse(dst, x+face.X*10, hy-8+math.Max(0, face.Y)*4, 14, 5, shade(lk.hatC, 0.8))
	case 3: // toque
		rect(dst, x-11, hy-28, 22, 18, rgb(0xffffff))
		circle(dst, x-8, hy-30, 9, rgb(0xffffff))
		circle(dst, x+8, hy-30, 9, rgb(0xffffff))
		circle(dst, x, hy-34, 10, rgb(0xffffff))
	case 4: // masque + cape
		if face.Y >= -0.5 {
			rect(dst, x-15, hy-3, 30, 8, rgb(lk.hatC))
			circle(dst, x-6+face.X*5, hy+1, 2.6, rgb(0xffffff))
			circle(dst, x+6+face.X*5, hy+1, 2.6, rgb(0xffffff))
		}
	case 6: // charlotte : bonnet blanc bouffant, avec son élastique froncé
		ellipse(dst, x, hy-13, 19, 12, rgb(0xffffff))
		for i := -4; i <= 4; i++ {
			circle(dst, x+float64(i)*4, hy-6+math.Abs(float64(i))*0.4, 2.8, rgb(0xeeeeee))
		}
		circle(dst, x-6, hy-17, 5, rgb(0xf7f7f7))
		circle(dst, x+7, hy-15, 4, rgb(0xf7f7f7))
	case 7: // calot de glacier
		polygon(dst, []V{{X: x - 15, Y: hy - 8}, {X: x + 15, Y: hy - 8}, {X: x + 11, Y: hy - 24}, {X: x - 11, Y: hy - 24}}, rgb(0xffffff))
		rect(dst, x-15, hy-13, 30, 5, rgb(lk.hatC))
	case 8: // casque de moto
		circle(dst, x, hy-3, 19, rgb(lk.hatC))
		if face.Y >= -0.5 {
			roundRect(dst, x-13+face.X*5, hy-4, 26, 12, 5, rgba(0x223344, 0.85))
		}
	case 9: // chapeau de diplômé
		drawMortarboard(dst, x, hy-6, 0x24243a)
	case 5: // haut-de-forme
		ellipse(dst, x, hy-12, 22, 6, rgb(lk.hatC))
		rect(dst, x-13, hy-38, 26, 26, rgb(lk.hatC))
	}
	// pile d'objets portés
	for i, it := range carry {
		drawItem(dst, it, x, y-40-bob-float64(i)*14, 14)
	}
}

func drawAgent(dst *ebiten.Image, place string, a *sim.Agent, t float64) {
	lk := staffLook(place, a.Kind, a.Look)
	lk.crate = place != "ferme"
	if a.Kind == sim.WorkerAgent && a.Settled && !a.Moving {
		drawBored(dst, a, lk, t)
		return
	}
	if a.Kind == sim.CashierAgent && a.Bored > sim.CashierBoredAfter {
		// pas de client : elle sifflote derrière sa caisse
		lk.mouth = 2
		drawPerson(dst, a.Pos, a.Face, 0, false, lk, nil)
		drawNotes(dst, a.Pos.X, a.Pos.Y, t+float64(a.Look)*1.3)
		return
	}
	drawPerson(dst, a.Pos, a.Face, a.WalkT, a.Moving, lk, a.Carry)
}

// drawBored : un ouvrier qui n'a rien à faire (c'est bon signe : la ferme
// produit plus que ce que la ville consomme).
func drawBored(dst *ebiten.Image, a *sim.Agent, lk look, t float64) {
	ph := t + float64(a.Look)*1.7 // chacun à son rythme
	x, y := a.Pos.X, a.Pos.Y
	switch a.Mood {
	case 0: // assis sur une botte de foin ; au bout d'un moment, il s'endort
		lk.sit = true
		lk.shut = a.Bored > 40
		drawPerson(dst, a.Pos, a.Face, 0, false, lk, nil)
		if lk.shut {
			for i := 0; i < 3; i++ {
				k := math.Mod(ph*0.5+float64(i)/3, 1)
				c := rgba(0x34495e, 1-k)
				say(dst, "z", x+14+k*22, y-92-k*40, 14+k*10, true, 0, c)
			}
		}
	case 1: // bâille et s'étire de temps en temps
		k := math.Mod(ph, 5)
		if k < 1.4 {
			lk.mouth, lk.shut, lk.armsUp = 1, true, true
		}
		drawPerson(dst, a.Pos, a.Face, 0, false, lk, nil)
	default: // sifflote : des notes s'envolent
		lk.mouth = 2
		drawPerson(dst, a.Pos, a.Face, 0, false, lk, nil)
		drawNotes(dst, x, y, ph)
	}
}

// drawNotes : des notes de musique qui s'envolent (quelqu'un sifflote).
func drawNotes(dst *ebiten.Image, x, y, ph float64) {
	for i := 0; i < 2; i++ {
		k := math.Mod(ph*0.6+float64(i)/2, 1)
		nx := x + 16 + k*26 + math.Sin(k*9)*4
		ny := y - 70 - k*50
		c := rgba(0x8e44ad, math.Min(1, 1.4-1.4*k))
		ellipse(dst, nx, ny, 6, 4.5, c)
		rect(dst, nx+4, ny-18, 2.5, 18, c)
		if i == 0 {
			rect(dst, nx+4, ny-18, 9, 3, c)
		}
	}
}

func drawCustomer(dst *ebiten.Image, p *sim.Place, c *sim.Customer, t float64) {
	lk := lookFrom(c.Look)
	if c.Cart {
		// avec un caddie : les articles vont dedans
		behind := math.Abs(c.Face.X) <= math.Abs(c.Face.Y) && c.Face.Y < 0
		if behind {
			drawCart(dst, c.Pos, c.Face, c.Carry, c.Paid > 0)
		}
		drawPerson(dst, c.Pos, c.Face, c.WalkT, c.Moving, lk, nil)
		if !behind {
			drawCart(dst, c.Pos, c.Face, c.Carry, c.Paid > 0)
		}
	} else {
		drawPerson(dst, c.Pos, c.Face, c.WalkT, c.Moving, lk, c.Carry)
	}
	if c.CheckN > 0 && c.Paid == 0 && c.CheckT > 0 {
		// passage en caisse : une barre qui se remplit, objet par objet
		x, y := c.Pos.X, c.Pos.Y-120
		roundRect(dst, x-34, y, 68, 12, 6, rgba(0x000000, 0.35))
		roundRect(dst, x-32, y+2, 64*(1-c.CheckT/c.CheckN), 8, 4, rgb(0x2fa86b))
	}
	hx, hy := c.Pos.X+26, c.Pos.Y-112
	switch {
	case c.Mood == 2:
		face(dst, hx, hy, 0xe74c3c, 2)
	case c.Mood == 1:
		face(dst, hx, hy, 0xf5c518, 1)
	case c.Shopping() && c.Wanted() >= 0:
		// bulle : ce qu'il vient chercher
		circle(dst, hx, hy, 22, rgba(0xffffff, 0.95))
		circle(dst, hx-16, hy+20, 5, rgba(0xffffff, 0.95))
		drawItem(dst, sim.Item(c.Wanted()), hx, hy, 14)
	case c.Paid > 0:
		face(dst, hx, hy, 0x7ed957, 0)
	}
}

// drawCart : le caddie poussé devant soi (articles dedans, ou sacs une fois payé).
func drawCart(dst *ebiten.Image, pos, face V, items []sim.Item, bags bool) {
	wire, dark := rgb(0xb8c2cc), rgb(0x6c7a89)
	if math.Abs(face.X) > math.Abs(face.Y) {
		// de profil
		fx := 1.0
		if face.X < 0 {
			fx = -1
		}
		x, y := pos.X+fx*46, pos.Y
		ellipse(dst, x, y+2, 28, 6, rgba(0x000000, 0.18))
		polygon(dst, []V{{X: x - 28, Y: y - 56}, {X: x + 28, Y: y - 56}, {X: x + 22, Y: y - 24}, {X: x - 20, Y: y - 24}}, wire)
		for k := 0; k < 4; k++ {
			xx := x - 18 + float64(k)*12
			line(dst, xx, y-54, xx, y-26, 1.5, dark)
		}
		line(dst, x-24, y-44, x+24, y-44, 1.5, dark)
		line(dst, x-20, y-24, x+18, y-8, 3, dark)
		line(dst, x-fx*28, y-56, x-fx*40, y-60, 4, rgb(0xe74c3c)) // poignée
		circle(dst, x-16, y-5, 5, rgb(0x2b2b2b))
		circle(dst, x+18, y-5, 5, rgb(0x2b2b2b))
		cartLoad(dst, x, y-56, items, bags)
		return
	}
	// de face (vers le bas) ou de dos (vers le haut)
	x, y := pos.X, pos.Y+34
	if face.Y < 0 {
		y = pos.Y - 18
	}
	ellipse(dst, x, y+4, 30, 6, rgba(0x000000, 0.18))
	circle(dst, x-20, y-2, 4.5, rgb(0x2b2b2b))
	circle(dst, x+20, y-2, 4.5, rgb(0x2b2b2b))
	line(dst, x-20, y-4, x-22, y-22, 3, dark)
	line(dst, x+20, y-4, x+22, y-22, 3, dark)
	roundRect(dst, x-28, y-54, 56, 34, 4, wire)
	for k := 0; k < 5; k++ {
		line(dst, x-22+float64(k)*11, y-52, x-22+float64(k)*11, y-22, 1.5, dark)
	}
	line(dst, x-26, y-38, x+26, y-38, 1.5, dark)
	cartLoad(dst, x, y-54, items, bags)
	if face.Y < 0 {
		rect(dst, x-28, y-58, 56, 5, rgb(0xe74c3c)) // la poignée, de notre côté
	}
}

// cartLoad : ce qui dépasse du caddie.
func cartLoad(dst *ebiten.Image, x, top float64, items []sim.Item, bags bool) {
	if bags {
		roundRect(dst, x-20, top-14, 18, 22, 3, rgb(0xc8a36a))
		roundRect(dst, x+2, top-18, 18, 26, 3, rgb(0xd9b77e))
		return
	}
	for i, it := range items {
		if i >= 6 {
			break
		}
		drawItem(dst, it, x-16+float64(i%3)*16, top+4-float64(i/3)*11, 10)
	}
}

// drawCartShelter : l'abri à caddies (ceux qui restent, emboîtés).
func drawCartShelter(dst *ebiten.Image, p *sim.Place) {
	s, ok := p.CartSpot()
	if !ok {
		return
	}
	x0 := s.X + 30
	roundRect(dst, x0, s.Y-70, 150, 84, 8, rgba(0x2fa86b, 0.25))
	rect(dst, x0, s.Y-74, 150, 6, rgb(0x2fa86b))
	n := p.CartsFree()
	for i := min(n, 12) - 1; i >= 0; i-- {
		x := x0 + 30 + float64(i)*8
		polygon(dst, []V{{X: x - 22, Y: s.Y - 52}, {X: x + 22, Y: s.Y - 52}, {X: x + 17, Y: s.Y - 24}, {X: x - 16, Y: s.Y - 24}}, rgb(0xb8c2cc))
		line(dst, x-22, s.Y-52, x-30, s.Y-56, 4, rgb(0xe74c3c))
		circle(dst, x-12, s.Y-5, 4, rgb(0x2b2b2b))
		circle(dst, x+14, s.Y-5, 4, rgb(0x2b2b2b))
		line(dst, x-16, s.Y-24, x+14, s.Y-8, 2.5, rgb(0x6c7a89))
	}
	say(dst, "Caddies : "+itoa(n), x0+75, s.Y-90, 16, true, 0, rgb(0x1e6e46))
}

// face : petit émoticône (0 content, 1 impatient, 2 fâché).
func face(dst *ebiten.Image, x, y float64, col uint32, mood int) {
	circle(dst, x, y, 17, rgb(col))
	circle(dst, x-6, y-3, 2.6, rgb(0x2a2a2a))
	circle(dst, x+6, y-3, 2.6, rgb(0x2a2a2a))
	switch mood {
	case 0:
		var pts []V
		for k := 0; k <= 8; k++ {
			a := math.Pi * float64(k) / 8
			pts = append(pts, V{X: x + math.Cos(a)*8, Y: y + 4 + math.Sin(a)*5})
		}
		polygon(dst, pts, rgb(0x2a2a2a))
	case 1:
		line(dst, x-6, y+7, x+6, y+7, 3, rgb(0x2a2a2a))
		ellipse(dst, x+13, y-8, 3, 5, rgb(0x6ec6ff))
	case 2:
		line(dst, x-9, y-9, x-3, y-6, 3, rgb(0x2a2a2a))
		line(dst, x+9, y-9, x+3, y-6, 3, rgb(0x2a2a2a))
		var pts []V
		for k := 0; k <= 8; k++ {
			a := math.Pi + math.Pi*float64(k)/8
			pts = append(pts, V{X: x + math.Cos(a)*7, Y: y + 10 + math.Sin(a)*4})
		}
		polygon(dst, pts, rgb(0x2a2a2a))
	}
}

// drawLevelZone : case d'amélioration à côté d'un champ ou d'un bâtiment.
func (g *game) drawLevelZone(dst *ebiten.Image, p *sim.Place, s *sim.Station) {
	u := p.StationUpgrade(s)
	l := g.sim.Levels[u.Key]
	r := s.LevelZone()
	if l >= u.Max {
		roundRect(dst, r.X+20, r.Y+20, r.W-40, r.H-40, 12, rgba(0x000000, 0.15))
		say(dst, "MAX", r.C().X, r.C().Y, 20, true, 0, rgb(0xffe066))
		return
	}
	cost := u.Cost(l)
	afford := g.sim.Money >= cost-s.LvlPaid
	c := uint32(0x9fd3ff)
	if afford {
		c = 0x6fe36f
	}
	roundRect(dst, r.X, r.Y, r.W, r.H, 14, rgba(0x000000, 0.25))
	if s.LvlPaid > 0 {
		f := s.LvlPaid / cost
		roundRect(dst, r.X, r.Y+r.H*(1-f), r.W, r.H*f, 14, rgba(0x6fe36f, 0.5))
	}
	dashedRect(dst, r.X+3, r.Y+3, r.W-6, r.H-6, 4, 10, g.t*25, rgb(c))
	// flèche vers le haut
	cx, top := r.C().X, r.Y+12
	bob := 0.0
	if afford {
		bob = 3 * math.Sin(g.t*6)
	}
	polygon(dst, []V{{X: cx, Y: top - bob}, {X: cx + 15, Y: top + 15 - bob}, {X: cx - 15, Y: top + 15 - bob}}, rgb(c))
	rect(dst, cx-6, top+14-bob, 12, 12, rgb(c))
	sayOutlined(dst, "Niveau "+itoa(l+2), cx, r.Y+r.H/2+4, 20, 0, rgb(0xffffff), rgb(0x1e2a12))
	pricePill(dst, money(cost-s.LvlPaid), cx, r.Y+r.H-4, r.W+20)
}

// drawSpot : case au sol pour embaucher ou améliorer (personnel, ventes).
func (g *game) drawSpot(dst *ebiten.Image, p *sim.Place, sp *sim.Spot) {
	u := p.SpotUpgrade(sp)
	l := g.sim.Levels[u.Key]
	r := sp.Zone
	// décor : panneaux au bord de la route
	decor := sp.ID
	if p.ID != "ferme" {
		decor = ""
	}
	switch decor {
	case "pancarte":
		px := r.X - 50
		rect(dst, px-4, 1500, 8, 80, rgb(0x6e4322))
		w := 90 + 12*float64(l)
		roundRect(dst, px-w/2, 1470-4*float64(l), w, 48+4*float64(l), 6, rgb(0xf3ead8))
		sayFit(dst, "MA FERME", px, 1494-2*float64(l), 18+float64(l), true, 0, w-10, rgb(0x6e4322))
	case "bio":
		if l > 0 {
			px := r.X + r.W + 50
			rect(dst, px-4, 1510, 8, 70, rgb(0x6e4322))
			circle(dst, px, 1500, 30, rgb(0x2fa86b))
			say(dst, "BIO", px, 1500, 20, true, 0, rgb(0xffffff))
		}
	}
	if sp.Shelf != nil && (l >= u.Max || l >= p.DroneMax(sp.Shelf)) {
		// base du drone
		c := r.C()
		ellipse(dst, c.X, c.Y+8, 46, 26, rgba(0x000000, 0.12))
		ellipse(dst, c.X, c.Y, 44, 24, rgb(0x8a8f96))
		ellipse(dst, c.X, c.Y, 34, 17, rgb(0x6f747b))
		say(dst, "H", c.X, c.Y, 22, true, 0, rgb(0xffe066))
		return
	}
	if l >= u.Max {
		if !strings.HasPrefix(sp.ID, "caissiere") && sp.ID != "secretaire" {
			roundRect(dst, r.X+25, r.Y+30, r.W-50, r.H-60, 12, rgba(0x000000, 0.15))
			say(dst, "MAX", r.C().X, r.C().Y, 20, true, 0, rgb(0xffe066))
		}
		return
	}
	cost := u.Cost(l)
	afford := g.sim.Money >= cost-sp.Paid
	c := uint32(0x9fd3ff)
	if afford {
		c = 0x6fe36f
	}
	roundRect(dst, r.X, r.Y, r.W, r.H, 14, rgba(0x000000, 0.28))
	if sp.Paid > 0 {
		f := sp.Paid / cost
		roundRect(dst, r.X, r.Y+r.H*(1-f), r.W, r.H*f, 14, rgba(0x6fe36f, 0.5))
	}
	dashedRect(dst, r.X+3, r.Y+3, r.W-6, r.H-6, 4, 10, g.t*25, rgb(c))
	icon := sp.ID
	if sp.Team != "" {
		icon = "ouvriers"
	}
	if strings.HasPrefix(sp.ID, "drone") {
		icon = "drone"
	}
	spotIcon(dst, icon, r.C().X, r.Y+30)
	title := sp.Label
	switch {
	case sp.ID == "ouvriers":
		title = sp.Label + " +1"
	case sp.Team != "":
		title = "Ouvrier +1"
		for _, t := range sim.FarmTeams {
			if t.ID == sp.Team && p.ID == "ferme" {
				title = "Ouvrier " + t.Name + " +1" // (« Ouvrier du maïs +1 »)
			}
		}
		if p.ID == "super" {
			title = "Magasinier +1"
		}
		if sim.TeamZones(p.ID) != nil {
			title = sp.Label + " +1" // (« Bovins +1 »)
		}
	case sp.Shelf != nil && l > 0:
		title = "Drone +1" // (ferme : un 2e, un 3e drone pour l'étal)
	case sp.Room != nil && strings.HasPrefix(sp.ID, "tables_"):
		title = "Table +1"
	case u.Max > 1:
		title += " niv. " + itoa(l+1)
	}
	sayFit(dst, title, r.C().X, r.Y+r.H-46, 19, true, 0, r.W-8, rgb(0xffffff))
	pricePill(dst, money(cost-sp.Paid), r.C().X, r.Y+r.H-4, r.W+20)
}

func spotIcon(dst *ebiten.Image, id string, x, y float64) {
	switch id {
	case "caissiere":
		lk := staffLook(staffPlace, sim.CashierAgent, 0)
		scaled(x, y, 0.42, func() { drawPerson(dst, V{X: x, Y: y + 46}, V{Y: 1}, 0, false, lk, nil) })
	case "ouvriers":
		lk := staffLook(staffPlace, sim.WorkerAgent, 0)
		scaled(x, y, 0.42, func() { drawPerson(dst, V{X: x, Y: y + 46}, V{Y: 1}, 0, false, lk, nil) })
	case "vitesse":
		roundRect(dst, x-16, y-14, 13, 24, 4, rgb(0x6b4423))
		roundRect(dst, x+3, y-14, 13, 24, 4, rgb(0x6b4423))
		roundRect(dst, x-19, y+4, 17, 9, 4, rgb(0x4a2f18))
		roundRect(dst, x+2, y+4, 17, 9, 4, rgb(0x4a2f18))
	case "capacite":
		polygon(dst, []V{{X: x - 22, Y: y - 10}, {X: x + 18, Y: y - 10}, {X: x + 12, Y: y + 8}, {X: x - 16, Y: y + 8}}, rgb(0x6f8fa8))
		circle(dst, x+14, y+13, 7, rgb(0x333333))
		line(dst, x-22, y-10, x-30, y+12, 3, rgb(0x8a5a33))
	case "pancarte":
		rect(dst, x-3, y, 6, 20, rgb(0x6e4322))
		roundRect(dst, x-24, y-16, 48, 22, 4, rgb(0xf3ead8))
	case "bio":
		circle(dst, x, y, 18, rgb(0x2fa86b))
		say(dst, "BIO", x, y, 13, true, 0, rgb(0xffffff))
	case "etals":
		stockIcon(dst, x, y+4)
	case "entrepot":
		// une palette chargée de cartons
		rect(dst, x-22, y+10, 44, 6, rgb(0xa47447))
		roundRect(dst, x-20, y-8, 18, 18, 3, rgb(0xc8a26a))
		roundRect(dst, x+2, y-8, 18, 18, 3, rgb(0xb88f55))
		roundRect(dst, x-9, y-24, 18, 16, 3, rgb(0xd9b77e))
	case "drone":
		drawDroneShape(dst, x, y, 0.8, 0)
	}
}

func road(dst *ebiten.Image, tl, br V, ry float64) {
	x0, w := tl.X-10, br.X-tl.X+20
	// trottoirs avec leur bordure (le trottoir du haut a une petite tranche)
	rect(dst, x0, ry-16, w, 16, rgb(0xc4c3bb))
	rect(dst, x0, ry-4, w, 4, rgb(0x93928a))
	rect(dst, x0, ry, w, 150, rgb(0x62666b))
	rect(dst, x0, ry, w, 10, rgba(0x000000, 0.10)) // (l'ombre de la bordure)
	rect(dst, x0, ry+150, w, 12, rgb(0xc4c3bb))
	rect(dst, x0, ry+150, w, 3, rgb(0xe0dfd8))
	for x := math.Floor((tl.X-100)/160) * 160; x < br.X+100; x += 160 {
		rect(dst, x, ry+71, 80, 8, rgb(0xf2f2e6))
	}
	// joints du trottoir
	for x := math.Floor((tl.X-100)/80) * 80; x < br.X+100; x += 80 {
		line(dst, x, ry-16, x, ry-4, 1.5, rgba(0x9c9b93, 0.6))
	}
}

func (g *game) farmGround(dst *ebiten.Image, p *sim.Place, tl, br V) {
	// --- chemins de terre
	path := rgb(0xd9c08f)
	roundRect(dst, sim.FarmX0+80, 470, p.W-sim.FarmX0-140, 110, 50, path) // le chemin du haut commence au champ de maïs
	drawSerreFloor(dst, p)
	roundRect(dst, 60, 880, p.W-120, 150, 60, path)
	roundRect(dst, 1180, 470, 140, 560, 60, path)
	// --- sol de la boutique
	roundRect(dst, 110, 1030, p.W-230, 480, 30, rgb(0xcfb894))
	roundRect(dst, 120, 1040, p.W-250, 460, 26, rgb(0xe8dcc0))
	for x := 120.0; x < p.W-140; x += 70 {
		line(dst, x, 1042, x, 1498, 1.5, rgba(0xb59f78, 0.35))
	}
	// --- route
	road(dst, tl, br, 1640)
	// allée du portail (entre la caisse et l'étal du maïs)
	gx := sim.FarmGateX
	rect(dst, gx-70, 1500, 140, 140, rgb(0xcfb894))
	// --- clôture de la boutique (avec le portail)
	fence := func(x1, x2 float64) {
		for x := x1; x <= x2; x += 40 {
			rect(dst, x-5, 1532, 10, 40, rgb(0x9a6b3e))
			circle(dst, x, 1532, 6, rgb(0xb07d4a))
		}
		rect(dst, x1, 1542, x2-x1, 7, rgb(0xb07d4a))
		rect(dst, x1, 1558, x2-x1, 7, rgb(0xb07d4a))
	}
	fence(60, gx-80)
	fence(gx+80, p.W-60)
	// panneau d'entrée
	roundRect(dst, gx-100, 1475, 200, 46, 8, rgb(0x7a4a26))
	say(dst, "MA FERME", gx, 1498, 26, true, 0, rgb(0xffe7a8))

}

// superGround : le bâtiment du supermarché (sol, murs, réserve).
func (g *game) superGround(dst *ebiten.Image, p *sim.Place, tl, br V) {
	drawSuperGround(dst, g, p, tl, br)
	g.backTruck(dst, p)
}

// backTruck : le prochain camion attendu à l'arrière d'un lieu.
func (g *game) backTruck(dst *ebiten.Image, p *sim.Place) {
	best, every := 1e9, 25.0
	var from, fromID string
	for _, q := range g.sim.Places {
		for _, d := range q.Stations {
			if d.Kind == sim.Dock && d.To == p.ID && g.sim.DockActive(d) && q.Owned {
				if l := g.sim.TruckLeft(d); l < best {
					best, from, every, fromID = l, q.Name, g.sim.DockEvery(d), q.ID
				}
			}
		}
	}
	if from == "" {
		return
	}
	e := every - best
	truckFrom = fromID
	defer func() { truckFrom = "ferme" }()
	stop := 260.0
	if p.ID == "super" {
		stop = superTruckStop(fromID) // il s'arrête derrière son espace
	}
	switch {
	case best < 3: // il arrive par la gauche
		f := (3 - best) / 3
		drawTruck(dst, stop-1400*(1-f)*(1-f), 60, true)
	case e < 2: // il décharge
		drawTruck(dst, stop, 60, true)
	case e < 4.5: // il repart vers la droite
		f := (e - 2) / 2.5
		drawTruck(dst, stop+3600*f*f, 60, false)
	}
	sayOutlined(dst, "Prochaine livraison ("+strings.ToLower(from)+") : "+itoa(int(best))+" s", 420, 60, 20, -1, rgb(0xffffff), rgb(0x2b2b2b))
}

// workshopGround : boulangerie et glacier (sol, réserve, vitrine).
func (g *game) workshopGround(dst *ebiten.Image, p *sim.Place, tl, br V) {
	floor, tile := uint32(0xf3e3c8), uint32(0xe2cfae)
	title, sc := "BOULANGERIE", uint32(0xc0623a)
	switch p.ID {
	case "glacier":
		floor, tile, title, sc = 0xe6f4fb, 0xcfe6f2, "GLACIER", 0x5aa0e0
	case "cremerie":
		floor, tile, title, sc = 0xf6fbff, 0xdbe9f4, "CRÈMERIE", 0x5b8fb9
	case "boucherie":
		floor, tile, title, sc = 0xfdfbf8, 0xf2d6d6, "BOUCHERIE-CHARCUTERIE", 0xb03a3a
	case "patisserie":
		floor, tile, title, sc = 0xfff5f8, 0xf9dbe6, "PÂTISSERIE", 0xe0567f
	}
	rect(dst, tl.X-10, 10, br.X-tl.X+20, 100, rgb(0x6b6f74))
	road(dst, tl, br, 1640)
	// (le bâtiment commence à x = 20 : les cases de la caissière, des vitrines et de la caisse sont dedans)
	slab(dst, 20, 130, p.W-100, 1400, 6, 16, 0x8a8478)
	rect(dst, 36, 146, p.W-132, 270, rgb(0xb9b3a8)) // réserve
	rect(dst, 36, 426, p.W-132, 1088, rgb(floor))
	for x := 36.0; x < p.W-96; x += 70 {
		for y := 426.0; y < 1514; y += 70 {
			if int(x/70+y/70)%2 == 0 {
				// (les derniers carreaux sont coupés au bord du sol : rien ne déborde sur l'herbe)
				rect(dst, x, y, math.Min(70, p.W-96-x), math.Min(70, 1514-y), rgb(tile))
			}
		}
	}
	rect(dst, 36, 416, 680, 14, rgb(0x6f6a62))
	rect(dst, 900, 416, p.W-1016, 14, rgb(0x6f6a62))
	say(dst, "RÉSERVE", 800, 160, 20, true, 0, rgba(0x4a4a4a, 0.7))
	if p.ID == "boucherie" {
		g.boucherieGround(dst, p)
	}
	// devanture
	rect(dst, 20, 1514, 1160, 16, rgb(0x6f6a62))
	rect(dst, 1320, 1514, p.W-1400, 16, rgb(0x6f6a62))
	tw := math.Max(280, textWidth(title, 24, true)+40)
	roundRect(dst, 1250-tw/2, 1518, tw, 34, 8, rgb(sc)) // (sur le mur du bas : les bases des drones sont juste au-dessus)
	say(dst, title, 1250, 1535, 22, true, 0, rgb(0xffffff))
	g.backTruck(dst, p)
}

// drawTruck : camion vu de côté (cx, cy : centre).
// truckFrom : lieu d'où part le camion qu'on dessine (son nom est sur le flanc).
var truckFrom = "ferme"

func drawTruck(dst *ebiten.Image, cx, cy float64, busy bool) {
	if truckFrom == "elevage" {
		drawFridgeTruck(dst, cx, cy, busy) // la viande voyage au froid
		return
	}
	ellipse(dst, cx, cy+38, 120, 10, rgba(0x000000, 0.25))
	roundRect(dst, cx-120, cy-42, 160, 74, 8, rgb(0xf3ead8))
	label := map[string]string{"ferme": "MA FERME", "boulangerie": "BOULANGERIE", "cremerie": "CRÈMERIE", "glacier": "GLACIER", "boucherie": "BOUCHERIE", "exotique": "EXOTIQUE", "port": "LE PORT", "elevage": "ÉLEVAGE", "domaine": "LE DOMAINE", "patisserie": "PÂTISSERIE"}[truckFrom]
	if label == "" {
		label = "MA FERME"
	}
	sayFit(dst, label, cx-40, cy-6, 20, true, 0, 150, rgb(0x2fa86b))
	roundRect(dst, cx+40, cy-22, 70, 54, 10, rgb(0x2fa86b))
	rect(dst, cx+66, cy-14, 34, 20, rgb(0x9fd3ff))
	for _, wx := range []float64{-85, -10, 80} {
		circle(dst, cx+wx, cy+34, 15, rgb(0x2b2b2b))
		circle(dst, cx+wx, cy+34, 6, rgb(0x999999))
	}
	if busy {
		circle(dst, cx+118, cy+10, 6, rgb(0xffd23f))
	}
}

// drawRayon : rayonnage métallique du supermarché.
func (g *game) drawRayon(dst *ebiten.Image, p *sim.Place, s *sim.Station) {
	b := s.Body
	ellipse(dst, b.C().X, b.Y+b.H, b.W*0.55, 12, rgba(0x000000, 0.18))
	roundRect(dst, b.X, b.Y-30, b.W, b.H+30, 6, rgb(0x7f93a8))
	rect(dst, b.X+6, b.Y-24, b.W-12, 30, rgb(0xdfe6ee))
	rect(dst, b.X+6, b.Y+10, b.W-12, 28, rgb(0xdfe6ee))
	if s.Stock > 16 {
		roundRect(dst, b.X+b.W-30, b.Y+36, 40, 22, 7, rgba(0x2b2140, 0.85))
		say(dst, itoa(s.Stock), b.X+b.W-10, b.Y+47, 15, true, 0, rgb(0xffffff))
	}
	for i := 0; i < min(s.Stock, 16); i++ {
		row := i / 8
		x := b.X + 18 + float64(i%8)*20
		y := b.Y + 28 - float64(row)*34
		drawItem(dst, s.Item, x, y, 11)
	}
	price := money(sim.Items[s.Item].Price * g.sim.PriceMultOf(p))
	roundRect(dst, b.C().X-40, b.Y+42, 80, 22, 5, rgb(0xffe066))
	say(dst, price, b.C().X, b.Y+53, 16, true, 0, rgb(0x4a3a20))
	if s.Stock == 0 && math.Mod(g.t, 1) < 0.6 {
		roundRect(dst, b.C().X-34, b.Y-58, 68, 26, 6, rgb(0xe74c3c))
		say(dst, "VIDE", b.C().X, b.Y-45, 17, true, 0, rgb(0xffffff))
	}
	// presque vide : ce qui bloque, en une icône
	if bl := g.sim.RayonBlock(s); bl != sim.BlockNone {
		drawBlockBadge(dst, bl, g.sim.RayonMissing(s), b.X+b.W-8, b.Y-48, g.t)
	}
}

// machineInputs : entrées d'une machine (icône + quantité), dans sa zone de dépôt.
func machineInputs(dst *ebiten.Image, s *sim.Station) {
	n := len(s.Inputs)
	for k, it := range s.Inputs {
		x := s.Zone.X + s.Zone.W*float64(k+1)/float64(n+1)
		drawItem(dst, it, x, s.Zone.Y+30, 15)
		y, size := s.Zone.Y+66, 16.0
		if n >= 3 {
			y, size = s.Zone.Y+58+float64(k%2)*20, 14 // en quinconce, sinon les nombres se touchent
		}
		sayOutlined(dst, itoa(s.Buf[it])+"/"+itoa(s.BufMax), x, y, size, 0, rgb(0x2b4a66), rgb(0xffffff))
	}
}

// drawFlowerStore : un abri en bois, ouvert devant, avec des seaux de fleurs.
func drawFlowerStore(dst *ebiten.Image, s *sim.Station) {
	b := s.Body
	ellipse(dst, b.C().X, b.Y+b.H, b.W*0.55, 12, rgba(0x000000, 0.2))
	roundRect(dst, b.X, b.Y+12, b.W, b.H-12, 6, rgb(0x8a5a33))
	roundRect(dst, b.X+8, b.Y+22, b.W-16, b.H-30, 4, rgb(0x5e3b20))
	// seaux : un par tranche de 4 fleurs, sur deux étagères
	n := (s.Stock + 3) / 4
	cols := []uint32{0xff6fa8, 0xffd23f, 0xb07cff, 0xff8a3d, 0xffffff}
	for k := 0; k < min(n, 10); k++ {
		x := b.X + 22 + float64(k%5)*29
		y := b.Y + 68 + float64(k/5)*36
		roundRect(dst, x-9, y, 18, 14, 3, rgb(0x9aa3ad))
		for j := 0; j < 3; j++ {
			circle(dst, x-6+float64(j)*6, y-4-float64(j%2)*4, 5, rgb(cols[(k+j)%len(cols)]))
		}
		line(dst, x-6, y, x-4, y-4, 2, rgb(0x3f8a3a))
	}
	line(dst, b.X+8, b.Y+87, b.X+b.W-8, b.Y+87, 3, rgb(0x8a5a33))
	// toit
	polygon(dst, []V{{X: b.X - 12, Y: b.Y + 16}, {X: b.C().X, Y: b.Y - 20}, {X: b.X + b.W + 12, Y: b.Y + 16}}, rgb(0xc0507a))
	rect(dst, b.X-12, b.Y+12, b.W+24, 8, rgb(0x9a3d62))
	sign(dst, b.C().X, b.Y+30, "Réserve de fleurs")
}

// drawPallet : palette en bois avec des cartons.
func drawPallet(dst *ebiten.Image, s *sim.Station) {
	b := s.Body
	ellipse(dst, b.C().X, b.Y+b.H, b.W*0.55, 10, rgba(0x000000, 0.2))
	rect(dst, b.X, b.Y+b.H-16, b.W, 16, rgb(0xb98550))
	for k := 0; k < 4; k++ {
		rect(dst, b.X+8+float64(k)*44, b.Y+b.H-12, 18, 8, rgb(0x8a5a33))
	}
	n := (s.Pile + 4) / 5 // un carton pour 5 produits
	for i := 0; i < n && i < 8; i++ {
		x := b.X + 10 + float64(i%4)*36
		y := b.Y + b.H - 52 - float64(i/4)*34
		rect(dst, x, y, 34, 34, rgb(0xc9a36a))
		rect(dst, x, y+14, 34, 5, rgb(0xa8834a))
	}
	roundRect(dst, b.C().X-26, b.Y-24, 52, 40, 8, rgb(0xffffff))
	drawItem(dst, s.Item, b.C().X, b.Y-4, 15)
}

// drawStockZone : case pour agrandir le stock d'un bâtiment.
func (g *game) drawStockZone(dst *ebiten.Image, p *sim.Place, s *sim.Station) {
	u := p.StockUpgrade(s)
	l := g.sim.Levels[u.Key]
	r := s.StockZone()
	if l >= u.Max {
		roundRect(dst, r.X+20, r.Y+18, r.W-40, r.H-36, 12, rgba(0x000000, 0.15))
		say(dst, "MAX", r.C().X, r.C().Y, 18, true, 0, rgb(0xffe066))
		return
	}
	cost := u.Cost(l)
	afford := g.sim.Money >= cost-s.StkPaid
	c := uint32(0x9fd3ff)
	if afford {
		c = 0x6fe36f
	}
	roundRect(dst, r.X, r.Y, r.W, r.H, 14, rgba(0x000000, 0.25))
	if s.StkPaid > 0 {
		f := s.StkPaid / cost
		roundRect(dst, r.X, r.Y+r.H*(1-f), r.W, r.H*f, 14, rgba(0x6fe36f, 0.5))
	}
	dashedRect(dst, r.X+3, r.Y+3, r.W-6, r.H-6, 4, 10, g.t*25, rgb(c))
	stockIcon(dst, r.C().X, r.Y+22)
	sayOutlined(dst, "Stock +", r.C().X, r.Y+r.H/2+6, 19, 0, rgb(0xffffff), rgb(0x1e2a12))
	pricePill(dst, money(cost-s.StkPaid), r.C().X, r.Y+r.H-4, r.W+20)
}

// stockIcon : deux cartons empilés.
func stockIcon(dst *ebiten.Image, x, y float64) {
	rect(dst, x-16, y-2, 18, 16, rgb(0xc9a36a))
	rect(dst, x+2, y-2, 18, 16, rgb(0xb98a50))
	rect(dst, x-7, y-16, 18, 14, rgb(0xd8b47a))
}

func drawPig(dst *ebiten.Image, x, y, t float64) {
	ellipse(dst, x, y+14, 24, 5, rgba(0x000000, 0.2))
	for _, dx := range []float64{-14, -6, 8, 15} {
		rect(dst, x+dx-3, y+4, 6, 10, rgb(0xe88a98))
	}
	ellipse(dst, x, y, 24, 15, rgb(0xf4a6b4))
	circle(dst, x+20, y-6, 11, rgb(0xf4a6b4))
	ellipse(dst, x+29, y-4, 5, 4, rgb(0xe07a8c))
	circle(dst, x+20, y-9, 1.8, rgb(0x2a2a2a))
	polygon(dst, []V{{X: x + 14, Y: y - 14}, {X: x + 19, Y: y - 22}, {X: x + 21, Y: y - 13}}, rgb(0xe07a8c))
	line(dst, x-24, y-4, x-30, y-10+math.Sin(t*4)*2, 2.5, rgb(0xe07a8c))
}

func placeShort(id string) string {
	switch id {
	case "super":
		return "supermarché"
	case "boulangerie":
		return "boulangerie"
	case "cremerie":
		return "crèmerie"
	case "glacier":
		return "glacier"
	case "elevage":
		return "ferme d'élevage"
	case "exotique":
		return "ferme exotique"
	case "boucherie":
		return "boucherie"
	case "port":
		return "port"
	case "domaine":
		return "domaine"
	case "patisserie":
		return "pâtisserie"
	}
	return id
}

// drawDock : quai d'expédition et son camion.
func (g *game) drawDock(dst *ebiten.Image, s *sim.Station) {
	if !g.sim.DockActive(s) {
		return
	}
	if s.PlaceID() == "domaine" {
		g.drawDockV(dst, s) // (la route des camions est verticale)
		return
	}
	b := s.Body
	roundRect(dst, b.X, b.Y, b.W, b.H, 6, rgb(0x9a6b3e))
	k := 0
	for it := sim.Item(0); it < sim.NItems; it++ {
		for n := 0; n < s.Buf[it] && k < 24; n++ {
			drawItem(dst, it, b.X+16+float64(k%12)*17, b.Y+14-float64(k/12)*12, 11)
			k++
		}
	}
	left := g.sim.TruckLeft(s)
	every := g.sim.DockEvery(s)
	truckFrom = s.PlaceID()
	defer func() { truckFrom = "ferme" }()
	e := every - left // temps depuis le dernier départ
	x0, y0 := b.C().X, b.Y+90
	truck := func(x float64) {
		scaled(x, y0, 0.75, func() { drawTruck(dst, x, y0, true) })
	}
	// le camion qui vient de partir s'éloigne vers la droite
	if e < 2.5 {
		f := e / 2.5
		truck(x0 + 2200*f*f)
	}
	// le suivant arrive par la gauche, puis attend au quai
	switch {
	case e >= 4.5:
		scaled(x0, y0, 0.75, func() { drawTruck(dst, x0, y0, left < 2) })
		sayOutlined(dst, "Départ dans "+itoa(int(left))+" s", x0, b.Y+140, 16, 0, rgb(0xffffff), rgb(0x2b2b2b))
	case e >= 1.5:
		f := (e - 1.5) / 3
		truck(x0 - 2200*(1-f)*(1-f))
	}
}

// drawWorkshopMachine : fours, turbines et autres machines sans dessin particulier.
func drawWorkshopMachine(dst *ebiten.Image, s *sim.Station, t float64) {
	b := s.Body
	wall, roof := uint32(0xf0d6a8), uint32(0x9a5a3a)
	switch s.Item {
	case sim.GlaceVanille, sim.GlaceFruits, sim.GlaceChoco, sim.GlaceFruitLait:
		wall, roof = 0xdff3ff, 0x5aa0e0
	case sim.Sucre:
		wall, roof = 0xf3f3f3, 0x7a9a5a
	case sim.Chocolat:
		wall, roof = 0xc9a07a, 0x5a3a2a
	case sim.FarineMais:
		wall, roof = 0xf6e3a6, 0xc98a2a
	case sim.Creme, sim.Beurre, sim.FromageBlanc, sim.Fromage, sim.FromageChevre, sim.FromageBrebis:
		wall, roof = 0xf6fbff, 0x5b8fb9
	case sim.Yaourt, sim.YaourtChevre, sim.YaourtBrebis:
		wall, roof = 0xf6fbff, 0x7fb3d5
	case sim.Croissant, sim.PainChoco:
		wall, roof = 0xf6e3c0, 0xb5651d
	}
	if s.PlaceID() == "boucherie" {
		wall, roof = 0xfdf4f2, 0xb03a3a // carrelage blanc, auvent rouge
	}
	if s.PlaceID() == "patisserie" {
		wall, roof = 0xfff0f5, 0xe0567f // rose bonbon
	}
	if s.PlaceID() == "glacier" {
		wall, roof = 0xf0d6a8, 0x9a5a3a // (toutes les fabriques de glace pareilles)
	}
	building(dst, b, wall, roof)
	// une grande vitrine qui montre ce qu'on fabrique
	roundRect(dst, b.C().X-42, b.Y+b.H-82, 84, 70, 12, rgb(0x5a3a2a))
	roundRect(dst, b.C().X-36, b.Y+b.H-76, 72, 58, 10, rgb(0xfff8ea))
	drawItem(dst, s.Item, b.C().X, b.Y+b.H-47, 20)
	if s.Working {
		for k := 0; k < 2; k++ {
			ph := math.Mod(t*0.7+float64(k)/2, 1)
			circle(dst, b.X+b.W-40+math.Sin(ph*6)*6, b.Y-10-ph*60, 10+ph*12, rgba(0xeeeeee, 0.6*(1-ph)))
		}
	}
}

// drawDroneShape : un petit drone à quatre hélices.
func drawDroneShape(dst *ebiten.Image, x, y, k, t float64) {
	for _, o := range []V{{X: -22, Y: -10}, {X: 22, Y: -10}, {X: -22, Y: 10}, {X: 22, Y: 10}} {
		line(dst, x, y, x+o.X*k, y+o.Y*k, 4*k, rgb(0x4a4f57))
		w := 14 * k * (0.6 + 0.4*math.Abs(math.Sin(t*30+o.X)))
		ellipse(dst, x+o.X*k, y+o.Y*k-4*k, w, 4*k, rgba(0xdddddd, 0.8))
	}
	roundRect(dst, x-12*k, y-8*k, 24*k, 16*k, 6*k, rgb(0xf2f2f2))
	circle(dst, x, y, 4*k, rgb(0x2fa86b))
}

// pricePill : un prix bien lisible, en jaune sur une pastille sombre, à cheval sur le bas d'une case.
func pricePill(dst *ebiten.Image, txt string, cx, cy, maxW float64) {
	size := 24.0
	w := textWidth(txt, size, true) + 24
	if w > maxW {
		size *= maxW / w
		w = maxW
	}
	h := size + 12
	roundRect(dst, cx-w/2, cy-h/2+3, w, h, h/2, rgba(0x000000, 0.35))
	roundRect(dst, cx-w/2, cy-h/2, w, h, h/2, rgb(0x2b2140))
	say(dst, txt, cx, cy, size, true, 0, rgb(0xffe066))
}
