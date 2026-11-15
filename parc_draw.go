package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// Le parc d'attractions : allées, entrée et caisses, parking, attractions
// (elles grandissent avec leur taille), visiteurs.

const parcRoadY = 1640.0

// parcGround : le sol du parc (pelouse déjà dessinée) : allées, place de
// l'entrée, route, parking.
func (g *game) parcGround(dst *ebiten.Image, p *sim.Place, tl, br V) {
	pave, paveD := rgb(0xe8dcc0), rgb(0xcfbf9c)
	// le grand boulevard et l'avenue
	bt := sim.ParcBoulY - sim.ParcBoulH/2
	roundRect(dst, 80, bt-10, sim.ParcW-160, sim.ParcBoulH+20, 60, paveD)
	roundRect(dst, 90, bt, sim.ParcW-180, sim.ParcBoulH, 56, pave)
	rect(dst, sim.ParcAvX-sim.ParcAvW/2-10, bt+40, sim.ParcAvW+20, sim.ParcPlazaY-bt, paveD)
	rect(dst, sim.ParcAvX-sim.ParcAvW/2, bt+40, sim.ParcAvW, sim.ParcPlazaY-bt, pave)
	// motif de dalles
	for x := math.Floor(math.Max(100, tl.X)/120) * 120; x < math.Min(sim.ParcW-100, br.X); x += 120 {
		line(dst, x, bt+20, x, bt+sim.ParcBoulH-20, 1.5, rgba(0xb59f78, 0.3))
	}
	// place de l'entrée
	roundRect(dst, sim.ParcAvX-1500, sim.ParcPlazaY, 3000, parcRoadY-14-sim.ParcPlazaY, 40, paveD)
	roundRect(dst, sim.ParcAvX-1490, sim.ParcPlazaY+10, 2980, parcRoadY-34-sim.ParcPlazaY, 36, pave)
	// grille du parc, ouverte entre les caisses
	fy := parcRoadY - 26.0
	for x := 0.0; x < sim.ParcW; x += 40 {
		if math.Abs(x-sim.ParcAvX) < 1000 {
			continue
		}
		line(dst, x, fy-40, x, fy, 3, rgb(0x4a5a3a))
	}
	line(dst, 0, fy-40, sim.ParcAvX-1000, fy-40, 4, rgb(0x4a5a3a))
	line(dst, sim.ParcAvX+1000, fy-40, sim.ParcW, fy-40, 4, rgb(0x4a5a3a))
	// la route et le passage piéton
	road(dst, tl, br, parcRoadY)
	gx := sim.ParcAvX
	rect(dst, gx-76, parcRoadY+4, 152, 142, rgb(0x6b6f74))
	for y := parcRoadY + 12.0; y < parcRoadY+140; y += 26 {
		rect(dst, gx-70, y, 140, 14, rgb(0xf2f2e6))
	}
	// le parking : de l'herbe au départ ; chaque agrandissement goudronne ses places
	// (et l'allée qui y mène). Les routes d'accès et de sortie descendent jusqu'à
	// la dernière allée utilisée.
	asph, edge := rgb(0x5f6368), rgb(0x9a9a92)
	rect(dst, tl.X-10, parcRoadY+162, br.X-tl.X+20, 13, rgb(0xc9c3b8))
	total := p.ParcParkTotal()
	deep := 0.0
	aisleOn := map[float64]bool{}
	for _, s := range p.Parks[:total] {
		aisleOn[s.Aisle] = true
		deep = math.Max(deep, s.Aisle)
	}
	ax0, ax1 := p.ParcAccessX()-50, p.ParcExitX()-50 // routes d'accès (gauche) et de sortie (droite)
	for _, x := range []float64{ax0, ax1} {
		rect(dst, x-6, parcRoadY+150, 112, deep+60-parcRoadY-150, edge)
		rect(dst, x, parcRoadY+150, 100, deep+55-parcRoadY-150, asph)
	}
	for a := range aisleOn {
		rect(dst, ax0, a-62, ax1+100-ax0, 124, edge)
		rect(dst, ax0, a-55, ax1+100-ax0, 110, asph)
		for x := math.Floor(math.Max(ax0, tl.X)/120) * 120; x < math.Min(ax1, br.X); x += 120 {
			rect(dst, x, a-3, 50, 6, rgba(0xffffff, 0.5))
		}
	}
	for _, s := range p.Parks[:total] {
		if s.Pos.X < tl.X-60 || s.Pos.X > br.X+60 {
			continue
		}
		t := s.Pos.Y - 58
		rect(dst, s.Pos.X-44, t, 88, 116, asph)
		line(dst, s.Pos.X-42, t+4, s.Pos.X-42, t+112, 3, rgba(0xffffff, 0.85))
		line(dst, s.Pos.X+42, t+4, s.Pos.X+42, t+112, 3, rgba(0xffffff, 0.85))
	}
	// le panneau d'entrée, au-dessus de l'avenue
	ax := sim.ParcAvX
	ay := sim.ParcPlazaY - 220 // (au-dessus des cases des caissières)
	rect(dst, ax-330, ay-130, 26, 150, rgb(0xd9534f))
	rect(dst, ax+304, ay-130, 26, 150, rgb(0xd9534f))
	roundRect(dst, ax-360, ay-190, 720, 90, 30, rgb(0xffd23f))
	roundRect(dst, ax-350, ay-182, 700, 74, 26, rgb(0xd9534f))
	sayFit(dst, "PARC D'ATTRACTIONS", ax, ay-145, 40, true, 0, 660, rgb(0xffffff))
	for i := 0; i < 9; i++ {
		circle(dst, ax-320+float64(i)*80, ay-190, 8, rgb([]uint32{0xff6fa8, 0x4fc3ff, 0x7ed957, 0xffd23f}[i%4]))
	}
	// les terrains des attractions
	for _, r := range p.Parc.Rides {
		g.drawRidePlot(dst, p, r)
	}
}

// parcDrawables : ce qui a une hauteur (caisses, employés, visiteurs, trottinettes).
func (g *game) parcDrawables(dst *ebiten.Image, p *sim.Place, ds []drawable, tl, br V) []drawable {
	t := g.t
	pc := p.Parc
	for _, r := range pc.Regs {
		r := r
		if !p.ParcRegOpen(r) {
			continue
		}
		ds = append(ds, drawable{r.Pos.Y + 30, func() { drawParcBooth(dst, p, r) }})
		if c := r.Cashier; c != nil {
			ds = append(ds, drawable{c.Pos.Y, func() { drawAgent(dst, p.ID, c, t) }})
		}
	}
	// le râtelier des trottinettes
	ds = append(ds, drawable{sim.ParcRackTop + 40, func() { drawTrottRack(dst, p) }})
	for _, r := range pc.Rides {
		if o := r.Operator; o != nil && p.RideBuilt(r) {
			ds = append(ds, drawable{o.Pos.Y, func() { drawAgent(dst, p.ID, o, t) }})
		}
	}
	for _, v := range pc.Visitors {
		v := v
		if v.State == sim.VCarState() || v.State == sim.VRidingState() || v.State == sim.VGoneState() {
			continue
		}
		if v.Pos.X < tl.X-150 || v.Pos.X > br.X+150 || v.Pos.Y < tl.Y-150 || v.Pos.Y > br.Y+200 {
			continue
		}
		ds = append(ds, drawable{v.Pos.Y, func() { drawVisitor(dst, v) }})
	}
	return ds
}

// drawParcBooth : une cabine de caisse (on paie l'entrée).
func drawParcBooth(dst *ebiten.Image, p *sim.Place, r *sim.ParcReg) {
	x, y := r.Pos.X, r.Pos.Y
	ellipse(dst, x, y+30, 70, 14, rgba(0x000000, 0.18))
	roundRect(dst, x-60, y-40, 120, 70, 8, rgb(0xf3ead8))
	rect(dst, x-60, y-40, 120, 14, rgb(0xd9534f))
	roundRect(dst, x-44, y-20, 88, 30, 6, rgb(0x9fd3ff))
	say(dst, "CAISSE "+itoa(r.N), x, y+18, 13, true, 0, rgb(0x6e4322))
	if !p.ParcRegManned(r) && len(r.Queue) > 0 {
		say(dst, "!", x, y-60, 30, true, 0, rgb(0xd9534f))
	}
}

// drawTrottRack : le râtelier des trottinettes, sur la place de l'entrée (visible
// même vide : on sait où elles iront).
func drawTrottRack(dst *ebiten.Image, p *sim.Place) {
	x, y, w := sim.ParcRackX, sim.ParcRackTop, sim.ParcRackW
	total, free := p.ParcTrottTotal(), p.ParcTrottFree()
	// l'abri : un toit et son panneau
	roundRect(dst, x, y+40, w, 300, 18, rgba(0x000000, 0.08))
	rect(dst, x+10, y, 10, 70, rgb(0x5a6a7a))
	rect(dst, x+w-20, y, 10, 70, rgb(0x5a6a7a))
	roundRect(dst, x-10, y-30, w+20, 44, 12, rgb(0x2fa86b))
	say(dst, "TROTTINETTES", x+w/2, y-8, 22, true, 0, rgb(0xffffff))
	// les places (30 visibles) : une trottinette, ou une place vide
	for i := 0; i < 30; i++ {
		cx := x + 30 + float64(i%10)*38
		cy := y + 100 + float64(i/10)*80
		if i < min(free, 30) {
			drawScooter(dst, V{X: cx, Y: cy}, V{X: 0, Y: 1}, 0.9, i)
		} else {
			roundRect(dst, cx-14, cy-4, 28, 8, 4, rgba(0x000000, 0.15))
		}
	}
	say(dst, itoa(free)+" / "+itoa(total), x+w/2, y+340, 22, true, 0, rgb(0x3b2a10))
}

// drawScooter : une trottinette électrique (vue de côté), couleur selon n.
func drawScooter(dst *ebiten.Image, at, face V, k float64, n int) {
	sx := 1.0
	if face.X < -0.2 {
		sx = -1
	}
	col := []uint32{0x4fc3ff, 0xff6fa8, 0x7ed957, 0xffd23f, 0xff8a3d}[(n%5+5)%5]
	x, y := at.X, at.Y
	circle(dst, x-24*k, y+6*k, 8*k, rgb(0x222222))
	circle(dst, x+24*k, y+6*k, 8*k, rgb(0x222222))
	roundRect(dst, x-28*k, y-6*k, 56*k, 10*k, 5*k, rgb(col))
	line(dst, x+22*k*sx, y-2*k, x+18*k*sx, y-46*k, 5*k, rgb(0x3a3a4a))
	line(dst, x+8*k*sx, y-46*k, x+28*k*sx, y-46*k, 5*k, rgb(0x2b2b3a))
}

// drawVisitor : un visiteur (sur sa trottinette s'il en a une et qu'il roule).
func drawVisitor(dst *ebiten.Image, v *sim.Visitor) {
	lk := lookFrom(v.Look)
	if v.OnTrott() {
		drawPerson(dst, v.Pos.Add(V{X: 0, Y: -12}), v.Face, 0, false, lk, nil)
		drawScooter(dst, v.Pos, v.Face, 1, v.Look)
	} else {
		drawPerson(dst, v.Pos, v.Face, v.WalkT, v.Moving, lk, nil)
	}
	// la jauge de bonheur, dans les files
	if v.State == sim.VQueueState() || v.State == sim.VRegQState() {
		c := rgb(0x2fa86b)
		switch {
		case v.Joy < 25:
			c = rgb(0xd9534f)
		case v.Joy < 50:
			c = rgb(0xf0a030)
		}
		circle(dst, v.Pos.X, v.Pos.Y-118, 7, c)
	}
}

// ---------------------------------------------------------------- attractions

// drawRidePlot : le terrain d'une attraction (au sol) et l'attraction elle-même.
func (g *game) drawRidePlot(dst *ebiten.Image, p *sim.Place, r *sim.Ride) {
	pl := r.Plot
	if !p.RideBuilt(r) {
		// un terrain qui attend son attraction
		roundRect(dst, pl.X, pl.Y, pl.W, pl.H, 40, rgba(0x6aa84f, 0.35))
		dashedRect(dst, pl.X+10, pl.Y+10, pl.W-20, pl.H-20, 5, 24, 0, rgba(0xffffff, 0.45))
		say(dst, r.Kind.Name, pl.C().X, pl.Y+80, 40, true, 0, rgba(0xffffff, 0.75))
		return
	}
	level := p.RideLevel(r)
	// le sol de l'attraction : une grande pelouse bordée d'une haie
	roundRect(dst, pl.X, pl.Y, pl.W, pl.H, 40, rgb(0x5f9e45))
	roundRect(dst, pl.X+14, pl.Y+14, pl.W-28, pl.H-28, 32, rgb(0x8cc56a))
	// la file d'attente : des barrières en serpentin
	qb := r.QueueBox()
	roundRect(dst, qb.X, qb.Y, qb.W, qb.H, 16, rgb(0xe8dcc0))
	for row := 0; row < 10; row++ {
		y := r.QueuePos(row * 8).Y
		yy := y + 20*inSign(r)
		x0, x1 := qb.X+20, qb.X+qb.W-20
		if row%2 == 0 {
			x1 -= 50
		} else {
			x0 += 50
		}
		if math.Abs(yy-r.Edge()) < qb.H-10 {
			line(dst, x0, yy, x1, yy, 4, rgb(0xb5865a))
		}
	}
	// nom de l'attraction et cabine de l'employé
	e := r.Edge()
	roundRect(dst, pl.X+20, e+inSign(r)*(-40)-24, 340, 48, 14, rgb(0x3b2a52))
	sayFit(dst, r.Kind.Name, pl.X+190, e+inSign(r)*(-40), 22, true, 0, 320, rgb(0xffe066))
	b := r.Booth()
	roundRect(dst, b.X, b.Y, b.W, b.H, 12, rgb(0xd9534f))
	roundRect(dst, b.X+10, b.Y+10, b.W-20, b.H-40, 8, rgb(0x9fd3ff))
	if !p.RideManned(r) {
		say(dst, "?", b.C().X, b.Y+b.H-18, 20, true, 0, rgb(0xffffff))
	}
	switch r.Kind.ID {
	case "train":
		drawTrainRide(dst, p, r, level, g.t)
	case "fantome":
		drawGhostRide(dst, p, r, level, g.t)
	case "bateaux":
		drawBoatRide(dst, p, r, level, g.t)
	case "volante":
		drawSwingTower(dst, p, r, level, g.t)
	case "grandhuit":
		drawCoaster(dst, p, r, level, g.t)
	case "splash":
		drawSplash(dst, p, r, level, g.t)
	case "vide":
		drawInverted(dst, p, r, level, g.t)
	default:
		drawGenericRide(dst, p, r, level, g.t)
	}
	// taille (en haut à droite)
	tr := r.Track()
	roundRect(dst, tr.X+tr.W-150, tr.Y+6, 140, 40, 14, rgba(0x000000, 0.35))
	say(dst, "Taille "+itoa(level), tr.X+tr.W-80, tr.Y+26, 20, true, 0, rgb(0xffffff))
}

func inSign(r *sim.Ride) float64 {
	if r.Top {
		return -1
	}
	return 1
}

// trackLoop : un circuit fermé dans la zone de l'attraction, qui s'allonge et
// se tortille avec la taille (1 à 10).
func trackLoop(r *sim.Ride, level int, wiggle float64) []V {
	tr := r.Track()
	f := float64(level) / 10
	// le circuit grandit vers la droite et vers le fond
	w := tr.W * (0.28 + 0.7*f)
	h := tr.H * math.Min(0.92, 0.42+0.6*f)
	cx := tr.X + w/2 + 20
	cy := tr.Y + tr.H/2
	if r.Top {
		cy = tr.Y + tr.H - h/2 - 20
	} else {
		cy = tr.Y + h/2 + 20
	}
	k := 2 + level/2
	amp := wiggle * f
	n := 120
	pts := make([]V, 0, n)
	for i := 0; i < n; i++ {
		a := 2 * math.Pi * float64(i) / float64(n)
		m := 1 + amp*math.Sin(float64(k)*a)
		pts = append(pts, V{X: cx + w/2*math.Cos(a)*m*0.95, Y: cy + h/2*math.Sin(a)*(1-0.1*f+0.1*f*math.Cos(2*a))*0.95})
	}
	return pts
}

// along : le point à la fraction s (0 à 1) d'un circuit fermé, et sa direction.
func along(pts []V, s float64) (V, V) {
	total := 0.0
	for i := range pts {
		total += pts[i].Dist(pts[(i+1)%len(pts)])
	}
	d := math.Mod(s, 1) * total
	if d < 0 {
		d += total
	}
	for i := range pts {
		a, b := pts[i], pts[(i+1)%len(pts)]
		l := a.Dist(b)
		if d <= l {
			return a.Add(b.Sub(a).Mul(d / l)), b.Sub(a).Norm()
		}
		d -= l
	}
	return pts[0], V{X: 1}
}

func drawLoop(dst *ebiten.Image, pts []V, w float64, c uint32) {
	for i := range pts {
		a, b := pts[i], pts[(i+1)%len(pts)]
		line(dst, a.X, a.Y, b.X, b.Y, w, rgb(c))
	}
}

// decorTrees : des arbres et des massifs de fleurs, de plus en plus nombreux avec la taille.
func decorTrees(dst *ebiten.Image, r *sim.Ride, level int, pts []V) {
	tr := r.Track()
	n := 3 + 3*level
	for i := 0; i < n; i++ {
		x := tr.X + 40 + math.Mod(float64(i)*373.7+float64(r.Kind.Slot)*91, tr.W-80)
		y := tr.Y + 50 + math.Mod(float64(i)*211.3+float64(r.Kind.Slot)*57, tr.H-100)
		near := false
		for k := 0; k < len(pts); k += 4 {
			if pts[k].Dist(V{X: x, Y: y}) < 70 {
				near = true
				break
			}
		}
		if near {
			continue
		}
		if i%3 == 0 {
			for j := 0; j < 5; j++ {
				circle(dst, x+float64(j%3)*14-14, y+float64(j/3)*12, 8, rgb([]uint32{0xff6fa8, 0xffd23f, 0xff8a3d, 0xb36bd6}[(i+j)%4]))
			}
			continue
		}
		drawTree(dst, x, y, 0.55+0.25*math.Mod(float64(i)*0.37, 1))
	}
}

// drawTrainRide : le petit train. Taille 1 : un tout petit rond près de la gare ;
// puis le circuit s'allonge et se tortille ; un étang et son pont (taille 4), un
// tunnel dans une colline (taille 6), un village miniature (taille 8), une
// grande montagne et une cascade (taille 10).
func drawTrainRide(dst *ebiten.Image, p *sim.Place, r *sim.Ride, level int, t float64) {
	pts := trackLoop(r, level, 0.18)
	tr := r.Track()
	decorTrees(dst, r, level, pts)
	// étang (taille 4+)
	if level >= 4 {
		c := pts[len(pts)/4]
		ellipse(dst, c.X, c.Y+(pts[0].Y-c.Y)*0.5, 160+12*float64(level), 70+6*float64(level), rgb(0x6fb7e6))
	}
	// le ballast puis les rails
	drawLoop(dst, pts, 30, 0xb59f78)
	drawLoop(dst, pts, 14, 0x8a6a4a)
	for i := 0; i < len(pts); i += 2 {
		a, b := pts[i], pts[(i+1)%len(pts)]
		d := b.Sub(a).Norm()
		n := V{X: -d.Y, Y: d.X}
		line(dst, a.X-n.X*13, a.Y-n.Y*13, a.X+n.X*13, a.Y+n.Y*13, 4, rgb(0x6e4322))
	}
	drawLoop(dst, pts, 3, 0xd0d0d0)
	// pont sur l'étang
	if level >= 4 {
		c := pts[len(pts)/4]
		roundRect(dst, c.X-50, c.Y-22, 100, 44, 10, rgb(0xa0522d))
	}
	// la gare, au départ du circuit (le point le plus proche de la cabine)
	st := 0
	b := r.Booth().C()
	for i, q := range pts {
		if q.Dist(b) < pts[st].Dist(b) {
			st = i
		}
	}
	s0 := float64(st) / float64(len(pts))
	g := pts[st]
	building(dst, sim.Rect{X: g.X - 70, Y: g.Y - 120, W: 140, H: 80}, 0xf3ead8, 0xc0623a)
	// tunnel dans une colline (taille 6+)
	var hill V
	if level >= 6 {
		hill = pts[len(pts)*5/8]
	}
	// village miniature (taille 8+)
	if level >= 8 {
		for i := 0; i < 6; i++ {
			x := tr.X + tr.W*0.55 + float64(i%3)*90
			y := tr.Y + tr.H*0.45 + float64(i/3)*90
			building(dst, sim.Rect{X: x, Y: y, W: 60, H: 40}, []uint32{0xf3ead8, 0xffe8b0, 0xd8e8f8}[i%3], []uint32{0xc0623a, 0x5a6a8a, 0x8a5a2b}[i%3])
		}
	}
	// montagne et cascade (taille 10)
	if level >= 10 {
		m := pts[len(pts)*7/8]
		polygon(dst, []V{{X: m.X - 160, Y: m.Y + 40}, {X: m.X, Y: m.Y - 220}, {X: m.X + 160, Y: m.Y + 40}}, rgb(0x8a8a80))
		polygon(dst, []V{{X: m.X - 40, Y: m.Y - 150}, {X: m.X, Y: m.Y - 220}, {X: m.X + 40, Y: m.Y - 150}}, rgb(0xf2f2f2))
		rect(dst, m.X+50, m.Y-120, 16, 160, rgba(0x9fd3ff, 0.9))
	}
	// le train : la locomotive et ses wagons
	units := p.RideUnits(r)
	pos := s0
	if r.State == 1 {
		pos = s0 + r.Prog
	}
	riders := len(r.Riders)
	gap := 52.0 / pathLen(pts)
	for k := 0; k <= units; k++ {
		at, dir := along(pts, pos-float64(k)*gap)
		if k == 0 {
			drawLoco(dst, at, dir)
			continue
		}
		seats := 2
		drawWagon(dst, at, dir, min(seats, max(0, riders-(k-1)*seats)), []uint32{0x4fc3ff, 0xffd23f, 0x7ed957, 0xff6fa8}[k%4])
	}
	if level >= 6 {
		ellipse(dst, hill.X, hill.Y, 120, 80, rgb(0x6a8a4a))
		ellipse(dst, hill.X, hill.Y-14, 90, 56, rgb(0x7ea85a))
		ellipse(dst, hill.X-60, hill.Y+20, 26, 22, rgb(0x2b2b2b))
	}
}

func pathLen(pts []V) float64 {
	t := 0.0
	for i := range pts {
		t += pts[i].Dist(pts[(i+1)%len(pts)])
	}
	return t
}

// drawLoco : la locomotive (rouge, avec sa cheminée).
func drawLoco(dst *ebiten.Image, at, dir V) {
	n := V{X: -dir.Y, Y: dir.X}
	body := []V{at.Add(dir.Mul(26)).Add(n.Mul(16)), at.Add(dir.Mul(26)).Sub(n.Mul(16)), at.Sub(dir.Mul(22)).Sub(n.Mul(16)), at.Sub(dir.Mul(22)).Add(n.Mul(16))}
	polygon(dst, body, rgb(0xd9534f))
	c := at.Add(dir.Mul(12))
	circle(dst, c.X, c.Y-6, 9, rgb(0x2b2b2b))
	cab := at.Sub(dir.Mul(12))
	roundRect(dst, cab.X-14, cab.Y-26, 28, 22, 4, rgb(0x2b4a8a))
}

// drawWagon : un wagon et ses passagers (des têtes).
func drawWagon(dst *ebiten.Image, at, dir V, riders int, col uint32) {
	n := V{X: -dir.Y, Y: dir.X}
	body := []V{at.Add(dir.Mul(20)).Add(n.Mul(15)), at.Add(dir.Mul(20)).Sub(n.Mul(15)), at.Sub(dir.Mul(20)).Sub(n.Mul(15)), at.Sub(dir.Mul(20)).Add(n.Mul(15))}
	polygon(dst, body, rgb(col))
	for i := 0; i < riders; i++ {
		h := at.Add(dir.Mul(float64(i)*16 - 8))
		circle(dst, h.X, h.Y-10, 8, rgb(0xf2c9a0))
	}
}

// drawGenericRide : en attendant leur dessin (étape 2), les autres attractions
// sont une boucle avec leurs wagons ; la tour volante, un mât et ses nacelles.
func drawGenericRide(dst *ebiten.Image, p *sim.Place, r *sim.Ride, level int, t float64) {
	if r.Kind.ID == "volante" {
		tr := r.Track()
		c := tr.C()
		h := 80 + 30*float64(level)
		rad := 60 + 14*float64(level)
		ellipse(dst, c.X, c.Y+20, rad+30, (rad+30)*0.4, rgba(0x000000, 0.15))
		rect(dst, c.X-10, c.Y-h, 20, h+20, rgb(0xd9534f))
		n := 2 * p.RideUnits(r)
		ang := 0.0
		if r.State == 1 {
			ang = t * 2
		}
		for i := 0; i < n; i++ {
			a := ang + 2*math.Pi*float64(i)/float64(n)
			x, y := c.X+rad*math.Cos(a), c.Y-h*0.7+rad*0.35*math.Sin(a)
			line(dst, c.X, c.Y-h*0.75, x, y, 2, rgb(0x888888))
			circle(dst, x, y, 9, rgb(0xffd23f))
		}
		return
	}
	pts := trackLoop(r, level, 0.25)
	decorTrees(dst, r, level, pts)
	col := map[string]uint32{"fantome": 0x3b2a52, "bateaux": 0x6fb7e6, "grandhuit": 0xd9534f, "splash": 0x4f9ee6, "vide": 0x5b7fd6}[r.Kind.ID]
	if r.Kind.Water {
		drawLoop(dst, pts, 44, 0x4f8fc6)
		drawLoop(dst, pts, 34, col)
	} else {
		drawLoop(dst, pts, 18, 0x555555)
		drawLoop(dst, pts, 8, col)
	}
	units := p.RideUnits(r)
	pos := 0.0
	if r.State == 1 {
		pos = r.Prog
	}
	gap := 60.0 / pathLen(pts)
	for k := 0; k < units; k++ {
		at, dir := along(pts, pos-float64(k)*gap)
		drawWagon(dst, at, dir, min(r.Kind.Seats, max(0, len(r.Riders)-k*r.Kind.Seats)), 0xffffff)
	}
}
