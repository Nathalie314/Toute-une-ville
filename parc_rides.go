package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// Les attractions du parc (sauf le petit train, dans parc_draw.go). Chacune
// grandit avec sa taille (1 à 10) : le circuit s'allonge, le décor s'enrichit.

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

// ridersIn : passagers dans le wagon n° k.
func ridersIn(r *sim.Ride, k, seats int) int {
	return min(seats, max(0, len(r.Riders)-k*seats))
}

func centroid(pts []V) V {
	c := V{}
	for _, q := range pts {
		c = c.Add(q)
	}
	return c.Mul(1 / float64(len(pts)))
}

func heads(dst *ebiten.Image, at, dir V, n int, spacing float64) {
	for i := 0; i < n; i++ {
		h := at.Add(dir.Mul((float64(i) - float64(n-1)/2) * spacing))
		circle(dst, h.X, h.Y-10, 8, rgb(0xf2c9a0))
	}
}

func box(at, dir V, l, w float64) []V {
	n := V{X: -dir.Y, Y: dir.X}
	return []V{at.Add(dir.Mul(l)).Add(n.Mul(w)), at.Add(dir.Mul(l)).Sub(n.Mul(w)), at.Sub(dir.Mul(l)).Sub(n.Mul(w)), at.Sub(dir.Mul(l)).Add(n.Mul(w))}
}

// scatter : des positions fixes (selon la graine) dans la zone, loin du circuit.
func scatter(r *sim.Ride, pts []V, n, seed int, keep float64) []V {
	tr := r.Track()
	var out []V
	for i := 0; i < n*3 && len(out) < n; i++ {
		x := tr.X + 40 + math.Mod(float64(i+seed)*373.7+float64(r.Kind.Slot)*91, tr.W-80)
		y := tr.Y + 50 + math.Mod(float64(i+seed)*211.3+float64(r.Kind.Slot)*57, tr.H-100)
		ok := true
		for k := 0; k < len(pts); k += 3 {
			if pts[k].Dist(V{X: x, Y: y}) < keep {
				ok = false
				break
			}
		}
		if ok {
			out = append(out, V{X: x, Y: y})
		}
	}
	return out
}

// ---------------------------------------------------------------- train fantôme

// ghostTrack : le circuit du train fantôme, différent du petit train : des allers
// et retours serrés (plus nombreux avec la taille), qui reviennent au départ.
func ghostTrack(r *sim.Ride, level int) []V {
	tr := r.Track()
	f := float64(level) / 10
	w := tr.W * (0.32 + 0.64*f)
	h := tr.H * math.Min(0.9, 0.5+0.45*f)
	x0 := tr.X + 30
	y0, y1 := tr.Y+tr.H-30-h, tr.Y+tr.H-30
	if !r.Top {
		y0, y1 = tr.Y+30, tr.Y+30+h
	}
	n := 1 + level/3 // nombre d'allers et retours (4 au plus)
	var raw []V
	step := w / float64(2*n+1)
	for i := 0; i < n; i++ {
		x := x0 + float64(2*i)*step
		raw = append(raw, V{X: x, Y: y1}, V{X: x, Y: y0 + 90}, V{X: x + step, Y: y0 + 90}, V{X: x + step, Y: y1 - 90})
	}
	raw = append(raw, V{X: x0 + float64(2*n)*step, Y: y1 - 90}, V{X: x0 + w, Y: y1 - 90}, V{X: x0 + w, Y: y0}, V{X: x0 - 20, Y: y0}, V{X: x0 - 20, Y: y1})
	return resample(smoothClosed(raw, 3), 180)
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
		roundRect(dst, at.X-16, at.Y-26, 32, 30, 10, rgb(0x8a8a8a))
	case 1:
		roundRect(dst, at.X-14, at.Y-34, 28, 38, 6, rgb(0xc8a050))
	case 2:
		polygon(dst, []V{{X: at.X - 12, Y: at.Y - 36}, {X: at.X + 12, Y: at.Y - 36}, {X: at.X + 16, Y: at.Y + 4}, {X: at.X - 16, Y: at.Y + 4}}, rgb(0x4a2a1a))
	default:
		ellipse(dst, at.X, at.Y, 22, 10, rgb(0x6a4a2a))
	}
	if up < 0.05 {
		return
	}
	h := 70 * up // il se dresse
	y := at.Y - 10 - h
	body := map[int]uint32{0: 0xf2f2e6, 1: 0xe8e0c8, 2: 0x2b2b3a, 3: 0x5a7a4a}[kind]
	face := map[int]uint32{0: 0xf2f2e6, 1: 0xe8e0c8, 2: 0xf0e8f0, 3: 0x8ab06a}[kind]
	rect(dst, at.X-10, y+14, 20, h*0.8, rgb(body))
	if kind == 2 { // la cape du vampire, doublée de rouge
		polygon(dst, []V{{X: at.X - 10, Y: y + 16}, {X: at.X - 26, Y: y + 16 + h*0.75}, {X: at.X - 8, Y: y + 16 + h*0.7}}, rgb(0xc0392b))
		polygon(dst, []V{{X: at.X + 10, Y: y + 16}, {X: at.X + 26, Y: y + 16 + h*0.75}, {X: at.X + 8, Y: y + 16 + h*0.7}}, rgb(0xc0392b))
	}
	if kind == 1 { // les bandelettes de la momie
		for k := 0.0; k < h*0.8; k += 8 {
			line(dst, at.X-10, y+16+k, at.X+10, y+20+k, 2, rgb(0xb8a888))
		}
	}
	if kind == 0 { // les côtes du squelette
		for k := 0.0; k < h*0.5; k += 9 {
			line(dst, at.X-8, y+20+k, at.X+8, y+20+k, 3, rgb(0x5a5a5a))
		}
	}
	// les bras levés (ils s'agitent)
	wave := 6 * math.Sin(t*10+at.X)
	line(dst, at.X-10, y+20, at.X-24, y+2+wave, 5, rgb(face))
	line(dst, at.X+10, y+20, at.X+24, y+2-wave, 5, rgb(face))
	circle(dst, at.X, y+6, 12, rgb(face))
	circle(dst, at.X-4, y+4, 3, rgb(map[int]uint32{0: 0x1b1b1b, 1: 0x1b1b1b, 2: 0xd9534f, 3: 0xd9534f}[kind]))
	circle(dst, at.X+4, y+4, 3, rgb(map[int]uint32{0: 0x1b1b1b, 1: 0x1b1b1b, 2: 0xd9534f, 3: 0xd9534f}[kind]))
	if kind == 2 {
		polygon(dst, []V{{X: at.X - 12, Y: y - 4}, {X: at.X, Y: y + 2}, {X: at.X + 12, Y: y - 4}, {X: at.X, Y: y - 10}}, rgb(0x1b1b1b))
	}
}

// drawWitch : une sorcière qui touille son chaudron.
func drawWitch(dst *ebiten.Image, at V, t float64) {
	ellipse(dst, at.X+40, at.Y+22, 44, 12, rgba(0x000000, 0.25))
	// le feu sous le chaudron
	for k := 0; k < 3; k++ {
		fl := 10 + 6*math.Sin(t*9+float64(k))
		polygon(dst, []V{{X: at.X + 28 + float64(k)*12, Y: at.Y + 20}, {X: at.X + 34 + float64(k)*12, Y: at.Y + 20 - fl}, {X: at.X + 40 + float64(k)*12, Y: at.Y + 20}}, rgb(0xff8a3d))
	}
	ellipse(dst, at.X+40, at.Y, 38, 26, rgb(0x2b2b2b))
	ellipse(dst, at.X+40, at.Y-12, 32, 10, rgb(0x6ad04a))
	for k := 0; k < 4; k++ { // les bulles
		b := math.Mod(t*1.5+float64(k)*0.25, 1)
		circle(dst, at.X+28+float64(k)*8, at.Y-14-b*40, 4+2*b, rgba(0x8af06a, 1-b))
	}
	// la sorcière : robe noire, visage vert, chapeau pointu ; la louche tourne
	a := t * 3
	sx, sy := at.X+40+math.Cos(a)*16, at.Y-12+math.Sin(a)*5
	polygon(dst, []V{{X: at.X - 26, Y: at.Y + 20}, {X: at.X - 10, Y: at.Y - 50}, {X: at.X + 6, Y: at.Y + 20}}, rgb(0x2b1b3b))
	line(dst, at.X-4, at.Y-30, sx, sy-30, 4, rgb(0x6aa04a))
	line(dst, sx, sy-36, sx, sy, 4, rgb(0x6e4322))
	circle(dst, at.X-10, at.Y-58, 12, rgb(0x6aa04a))
	circle(dst, at.X-4, at.Y-58, 3, rgb(0x1b1b1b))
	polygon(dst, []V{{X: at.X - 30, Y: at.Y - 66}, {X: at.X - 10, Y: at.Y - 110}, {X: at.X + 10, Y: at.Y - 66}}, rgb(0x1b1b1b))
	rect(dst, at.X-34, at.Y-68, 48, 6, rgb(0x1b1b1b))
}

// drawGhostRide : le train fantôme. Son circuit zigzague (de plus en plus avec la
// taille) ; il disparaît sous des tunnels sombres ; des squelettes, des momies,
// des vampires et des zombies surgissent à son passage ; une sorcière touille son
// chaudron près des rails. Tombes, citrouilles, arbres morts, fantômes, chauves-souris.
func drawGhostRide(dst *ebiten.Image, p *sim.Place, r *sim.Ride, level int, t float64) {
	pts := ghostTrack(r, level)
	c := centroid(pts)
	tr := r.Track()
	f := float64(level)
	ellipse(dst, c.X, c.Y, math.Min(tr.W/2, 260+60*f), math.Min(tr.H/2, 180+30*f), rgba(0x3b3b2b, 0.25))
	for i, q := range scatter(r, pts, 2+2*level, 11, 60) {
		switch i % 3 {
		case 0:
			circle(dst, q.X, q.Y, 14, rgb(0xff8a3d))
			rect(dst, q.X-2, q.Y-18, 4, 6, rgb(0x2f6a3a))
			circle(dst, q.X-5, q.Y-2, 2.5, rgb(0x3b2a10))
			circle(dst, q.X+5, q.Y-2, 2.5, rgb(0x3b2a10))
		default:
			line(dst, q.X, q.Y, q.X, q.Y-60, 6, rgb(0x4a3a2a))
			line(dst, q.X, q.Y-40, q.X-20, q.Y-62, 4, rgb(0x4a3a2a))
			line(dst, q.X, q.Y-48, q.X+18, q.Y-70, 4, rgb(0x4a3a2a))
		}
	}
	drawLoop(dst, pts, 26, 0x2b2b2b)
	drawLoop(dst, pts, 6, 0x7a5a9a)
	s0 := rideStart(r, pts)
	// les tunnels : un, puis un de plus tous les 3 niveaux (jamais à la gare)
	var tunnels [][2]float64
	nt := 1 + level/3
	for i := 0; i < nt; i++ {
		a := math.Mod(s0+0.12+float64(i)*0.8/float64(nt), 1)
		tunnels = append(tunnels, [2]float64{a, math.Min(0.999, a+0.045+0.004*f)})
	}
	pos := ridePos(r, s0)
	units := p.RideUnits(r)
	gap := 48.0 / pathLen(pts)
	var cars []V
	for k := 0; k < units; k++ {
		s := pos - float64(k)*gap
		at, dir := along(pts, s)
		cars = append(cars, at)
		if inTunnel(s, tunnels) {
			continue // sous le tunnel
		}
		polygon(dst, box(at, dir, 20, 15), rgb(0x2b2b3a))
		circle(dst, at.X+dir.X*14, at.Y+dir.Y*14-6, 7, rgb(0xd9d9d9)) // la tête de mort à l'avant
		heads(dst, at, dir, ridersIn(r, k, 2), 14)
	}
	// les tunnels par-dessus les rails : une voûte sombre et ses deux bouches
	for _, tn := range tunnels {
		var seg []V
		for s := tn[0]; s <= tn[1]; s += 0.004 {
			q, _ := along(pts, s)
			seg = append(seg, q)
		}
		for i := 0; i+1 < len(seg); i++ {
			line(dst, seg[i].X, seg[i].Y, seg[i+1].X, seg[i+1].Y, 48, rgb(0x3b2b4b))
			line(dst, seg[i].X, seg[i].Y-5, seg[i+1].X, seg[i+1].Y-5, 32, rgb(0x4a3a5a))
		}
		for _, e := range []V{seg[0], seg[len(seg)-1]} {
			circle(dst, e.X, e.Y, 26, rgb(0x2b1b3b))
			circle(dst, e.X, e.Y+4, 17, rgb(0x0b0b10))
		}
	}
	// les monstres, au bord des rails : ils surgissent quand un wagon arrive
	nm := min(14, 2+level)
	for i := 0; i < nm; i++ {
		s := math.Mod(s0+0.06+float64(i)*0.9/float64(nm), 1)
		if inTunnel(s, tunnels) {
			s = math.Mod(s+0.1, 1)
		}
		q, dir := along(pts, s)
		side := V{X: -dir.Y, Y: dir.X}.Mul(float64(1 - 2*(i%2)))
		at := q.Add(side.Mul(48))
		near := 1e9
		for _, cq := range cars {
			near = math.Min(near, cq.Dist(at))
		}
		up := math.Max(0, math.Min(1, (170-near)/80))
		drawMonster(dst, i%4, at, up, t)
	}
	// la sorcière et son chaudron, près des rails
	wq, wd := along(pts, s0+0.5)
	drawWitch(dst, wq.Add(V{X: -wd.Y, Y: wd.X}.Mul(80)).Add(V{X: -20}), t)
	// fantômes et chauves-souris
	for i := 0; i < level/2; i++ {
		a := t*0.8 + float64(i)*1.7
		x, y := c.X+math.Cos(a)*tr.W*0.25, c.Y-60+math.Sin(a*1.3)*tr.H*0.2
		ellipse(dst, x, y, 18, 24, rgba(0xffffff, 0.75))
		circle(dst, x-6, y-6, 3, rgb(0x2b2b2b))
		circle(dst, x+6, y-6, 3, rgb(0x2b2b2b))
	}
	for i := 0; i < (level+1)/2; i++ {
		a := t*1.6 + float64(i)*2.3
		x, y := c.X+math.Cos(a)*tr.W*0.3, c.Y-tr.H*0.25-30*math.Sin(a*2)
		wing := 12 * math.Abs(math.Sin(t*12+float64(i)))
		polygon(dst, []V{{X: x, Y: y}, {X: x - 22, Y: y - wing}, {X: x - 10, Y: y + 4}}, rgb(0x2b2b2b))
		polygon(dst, []V{{X: x, Y: y}, {X: x + 22, Y: y - wing}, {X: x + 10, Y: y + 4}}, rgb(0x2b2b2b))
	}
}

// ---------------------------------------------------------------- petits bateaux

// drawBoatRide : une rivière qui serpente dans un jardin : des ponts (un tous les
// trois niveaux), un kiosque (4), une fontaine (5), des cygnes (6), des saules.
func drawBoatRide(dst *ebiten.Image, p *sim.Place, r *sim.Ride, level int, t float64) {
	pts := trackLoop(r, level, 0.3)
	c := centroid(pts)
	for i, q := range scatter(r, pts, 4+3*level, 5, 90) {
		if i%2 == 0 {
			for j := 0; j < 6; j++ {
				circle(dst, q.X+float64(j%3)*14-14, q.Y+float64(j/3)*12, 8, rgb([]uint32{0xff6fa8, 0xffd23f, 0xb36bd6, 0xffffff}[(i+j)%4]))
			}
		} else {
			drawTree(dst, q.X, q.Y, 0.6+0.2*math.Mod(float64(i)*0.37, 1))
		}
	}
	drawLoop(dst, pts, 84, 0x5f9e45) // la berge
	drawLoop(dst, pts, 70, 0x4f8fc6)
	drawLoop(dst, pts, 56, 0x6fb7e6)
	for i := 0; i < len(pts); i += 6 { // des vaguelettes
		q := pts[(i+int(t*4))%len(pts)]
		line(dst, q.X-8, q.Y, q.X+8, q.Y, 2, rgba(0xffffff, 0.5))
	}
	if level >= 4 { // le kiosque
		roundRect(dst, c.X-60, c.Y-40, 120, 80, 40, rgb(0xf3ead8))
		polygon(dst, []V{{X: c.X - 80, Y: c.Y - 40}, {X: c.X, Y: c.Y - 110}, {X: c.X + 80, Y: c.Y - 40}}, rgb(0x2fa86b))
	}
	if level >= 5 { // la fontaine
		f := V{X: c.X, Y: c.Y + 120}
		ellipse(dst, f.X, f.Y, 70, 32, rgb(0xb0b0a8))
		ellipse(dst, f.X, f.Y, 58, 24, rgb(0x6fb7e6))
		for j := 0; j < 6; j++ {
			a := float64(j) * math.Pi / 3
			h := 30 + 6*math.Sin(t*5+float64(j))
			line(dst, f.X, f.Y, f.X+math.Cos(a)*30, f.Y-h, 3, rgba(0xffffff, 0.8))
		}
	}
	s0 := rideStart(r, pts)
	pos := ridePos(r, s0)
	units := p.RideUnits(r)
	gap := 80.0 / pathLen(pts)
	for k := 0; k < units; k++ {
		at, dir := along(pts, pos-float64(k)*gap)
		polygon(dst, box(at, dir, 32, 18), rgb(0xffffff))
		polygon(dst, box(at.Add(dir.Mul(-6)), dir, 16, 13), rgb([]uint32{0xff6fa8, 0xffd23f, 0x4fc3ff, 0x7ed957, 0xff8a3d}[k%5]))
		heads(dst, at, dir, ridersIn(r, k, 4), 13)
	}
	// les ponts par-dessus la rivière
	for b := 1; b <= level/3; b++ {
		q := pts[(len(pts)*b/(level/3+1)+len(pts)/8)%len(pts)]
		_, dir := along(pts, float64((len(pts)*b/(level/3+1)+len(pts)/8)%len(pts))/float64(len(pts)))
		n := V{X: -dir.Y, Y: dir.X}
		polygon(dst, box(q, n, 52, 22), rgb(0xa0522d))
		polygon(dst, box(q, n, 52, 14), rgb(0xc0824d))
	}
	for i := 0; i < level/3; i++ { // les cygnes
		at, dir := along(pts, float64(i)*0.31+t*0.01)
		ellipse(dst, at.X, at.Y, 14, 9, rgb(0xffffff))
		line(dst, at.X+dir.X*8, at.Y, at.X+dir.X*12, at.Y-16, 4, rgb(0xffffff))
		circle(dst, at.X+dir.X*13, at.Y-18, 4, rgb(0xffffff))
	}
}

// ---------------------------------------------------------------- tour volante

// drawSwingTower : la tour de chaises volantes. Le mât monte avec la taille ;
// pendant le tour, la couronne grimpe le long du mât et les nacelles s'envolent.
func drawSwingTower(dst *ebiten.Image, p *sim.Place, r *sim.Ride, level int, t float64) {
	tr := r.Track()
	f := float64(level)
	base := V{X: tr.X + tr.W*0.45, Y: tr.Y + tr.H*0.72}
	if r.Top {
		base.Y = tr.Y + tr.H*0.8
	}
	h := 140 + 62*f      // hauteur du mât
	cr := 70 + 15*f      // rayon de la couronne
	up, spin := 0.0, 0.0 // montée (0 à 1), vitesse
	if r.State == 1 {
		pr := r.Prog
		up = math.Min(1, math.Min(pr/0.15, (1-pr)/0.12))
		spin = up
	}
	ang := t * 1.8 * spin
	hc := 50 + (h-80)*up // hauteur de la couronne
	// la plate-forme et son ombre
	ellipse(dst, base.X, base.Y, cr+80, (cr+80)*0.38, rgb(0xb0b0a8))
	ellipse(dst, base.X, base.Y, cr+66, (cr+66)*0.38, rgb(0xe8dcc0))
	for i := 0; i < 12; i++ {
		a := float64(i) * math.Pi / 6
		circle(dst, base.X+math.Cos(a)*(cr+72), base.Y+math.Sin(a)*(cr+72)*0.38, 4, rgb([]uint32{0xff6fa8, 0xffd23f, 0x4fc3ff}[i%3]))
	}
	n := 2 * p.RideUnits(r)
	riders := len(r.Riders)
	cy := base.Y - hc
	swing := cr * (0.25 + 0.75*up)
	seat := func(i int, front bool) {
		a := ang + 2*math.Pi*float64(i)/float64(n)
		if (math.Sin(a) >= 0) != front {
			return
		}
		ox, oy := math.Cos(a), math.Sin(a)*0.38
		top := V{X: base.X + ox*cr*0.9, Y: cy + oy*cr*0.9}
		s := V{X: base.X + ox*(cr+swing), Y: cy + oy*(cr+swing) + 70 - 40*up}
		line(dst, top.X, top.Y, s.X, s.Y-10, 2, rgb(0x9a9a9a))
		roundRect(dst, s.X-10, s.Y-10, 20, 16, 5, rgb([]uint32{0xd9534f, 0xffd23f}[i%2]))
		if i < riders {
			circle(dst, s.X, s.Y-16, 7, rgb(0xf2c9a0))
			line(dst, s.X-3, s.Y+6, s.X-5, s.Y+18, 3, rgb(0x2b4a8a))
			line(dst, s.X+3, s.Y+6, s.X+5, s.Y+18, 3, rgb(0x2b4a8a))
		}
	}
	for i := 0; i < n; i++ {
		seat(i, false)
	}
	// le mât (rouge et blanc, des lumières)
	w := 16 + 1.6*f
	for y := 0.0; y < h; y += 40 {
		col := rgb(0xd9534f)
		if int(y/40)%2 == 1 {
			col = rgb(0xf2f2f2)
		}
		rect(dst, base.X-w/2, base.Y-y-40, w, 40, col)
	}
	roundRect(dst, base.X-w, base.Y-h-30, 2*w, 30, 10, rgb(0x2b2b3a))
	circle(dst, base.X, base.Y-h-38, 8, rgba(0xffd23f, 0.6+0.4*math.Sin(t*4)))
	// la couronne
	ellipse(dst, base.X, cy, cr+8, (cr+8)*0.38, rgb(0xffd23f))
	ellipse(dst, base.X, cy, cr-6, (cr-6)*0.38, rgb(0xd9534f))
	for i := 0; i < n; i++ {
		seat(i, true)
	}
	// des drapeaux autour (taille 5+)
	for i := 0; i < level-4; i++ {
		x := tr.X + 60 + float64(i)*(tr.W-120)/6
		y := base.Y + cr*0.38 + 120
		line(dst, x, y, x, y-80, 3, rgb(0x6e4322))
		polygon(dst, []V{{X: x, Y: y - 80}, {X: x + 34, Y: y - 70 + 4*math.Sin(t*3+float64(i))}, {X: x, Y: y - 60}}, rgb([]uint32{0xff6fa8, 0x4fc3ff, 0x7ed957}[i%3]))
	}
}

// ---------------------------------------------------------------- grand huit

// drawCoaster : un grand huit : rails rouges sur des pylônes blancs, une montée à
// chaîne, un looping tous les trois niveaux.
func drawCoaster(dst *ebiten.Image, p *sim.Place, r *sim.Ride, level int, t float64) {
	pts := trackLoop(r, level, 0.38)
	decorTrees(dst, r, level/2, pts)
	for i := 0; i < len(pts); i += 5 { // pylônes
		q := pts[i]
		line(dst, q.X, q.Y, q.X, q.Y+46, 5, rgb(0xe8e8e8))
		rect(dst, q.X-8, q.Y+44, 16, 6, rgb(0xb0b0a8))
	}
	drawLoop(dst, pts, 16, 0x7a1f1f)
	drawLoop(dst, pts, 8, 0xd9534f)
	for i := 0; i < len(pts); i += 2 {
		a, b := pts[i], pts[(i+1)%len(pts)]
		d := b.Sub(a).Norm()
		n := V{X: -d.Y, Y: d.X}
		line(dst, a.X-n.X*9, a.Y-n.Y*9, a.X+n.X*9, a.Y+n.Y*9, 2, rgb(0x5a1a1a))
	}
	s0 := rideStart(r, pts)
	// la montée à chaîne, juste après le départ
	for k := 0; k < 10; k++ {
		at, _ := along(pts, s0+0.02+float64(k)*0.006)
		circle(dst, at.X, at.Y, 3, rgb(0xffd23f))
	}
	// les loopings (un cercle debout sur le circuit)
	for i := 1; i <= level/3; i++ {
		q, dir := along(pts, s0+0.25*float64(i))
		rad := 50 + 4*float64(level)
		cc := q.Add(V{Y: -rad})
		_ = dir
		ring(dst, cc.X, cc.Y, rad, 12, rgb(0x7a1f1f))
		ring(dst, cc.X, cc.Y, rad, 6, rgb(0xd9534f))
		line(dst, cc.X-rad*0.7, cc.Y+rad*0.7, cc.X-rad*0.7, q.Y+40, 5, rgb(0xe8e8e8))
		line(dst, cc.X+rad*0.7, cc.Y+rad*0.7, cc.X+rad*0.7, q.Y+40, 5, rgb(0xe8e8e8))
	}
	pos := ridePos(r, s0)
	units := p.RideUnits(r)
	gap := 46.0 / pathLen(pts)
	for k := 0; k < units; k++ {
		at, dir := along(pts, pos-float64(k)*gap)
		polygon(dst, box(at, dir, 20, 16), rgb([]uint32{0x4fc3ff, 0xffd23f}[k%2]))
		heads(dst, at, dir, ridersIn(r, k, 4), 10)
	}
}

// ---------------------------------------------------------------- grand splash

// drawSplash : un canal en bois sur l'eau ; une chute dans les rochers tous les trois
// niveaux (et des éclaboussures quand un rondin la descend) ; des sapins.
func drawSplash(dst *ebiten.Image, p *sim.Place, r *sim.Ride, level int, t float64) {
	pts := trackLoop(r, level, 0.22)
	for i, q := range scatter(r, pts, 3+2*level, 17, 80) {
		k := 0.7 + 0.3*math.Mod(float64(i)*0.41, 1)
		polygon(dst, []V{{X: q.X - 26*k, Y: q.Y}, {X: q.X, Y: q.Y - 70*k}, {X: q.X + 26*k, Y: q.Y}}, rgb(0x2f6a3a))
		polygon(dst, []V{{X: q.X - 20*k, Y: q.Y - 30*k}, {X: q.X, Y: q.Y - 90*k}, {X: q.X + 20*k, Y: q.Y - 30*k}}, rgb(0x3f8a4a))
	}
	s0 := rideStart(r, pts)
	drops := 1 + level/3
	// les chutes : des rochers et un grand bassin au pied
	for i := 1; i <= drops; i++ {
		s := s0 + float64(i)/float64(drops+1)
		top, _ := along(pts, s-0.03)
		bot, _ := along(pts, s)
		polygon(dst, []V{{X: top.X - 90, Y: top.Y + 30}, {X: top.X - 30, Y: top.Y - 70 - 6*float64(level)}, {X: top.X + 40, Y: top.Y - 50}, {X: top.X + 90, Y: top.Y + 30}}, rgb(0x8a7a6a))
		ellipse(dst, bot.X, bot.Y+10, 90, 40, rgb(0x9fd3ff))
	}
	drawLoop(dst, pts, 50, 0x8a5a2b)
	drawLoop(dst, pts, 38, 0x4fa3e6)
	for i := 0; i < len(pts); i += 7 {
		q := pts[(i+int(t*6))%len(pts)]
		line(dst, q.X-6, q.Y, q.X+6, q.Y, 2, rgba(0xffffff, 0.6))
	}
	pos := ridePos(r, s0)
	units := p.RideUnits(r)
	gap := 80.0 / pathLen(pts)
	for k := 0; k < units; k++ {
		s := pos - float64(k)*gap
		at, dir := along(pts, s)
		polygon(dst, box(at, dir, 30, 15), rgb(0x8a5a2b))
		polygon(dst, box(at, dir, 24, 10), rgb(0xa0703b))
		heads(dst, at, dir, ridersIn(r, k, 4), 12)
		// au pied d'une chute : ça éclabousse !
		if r.State == 1 {
			for i := 1; i <= drops; i++ {
				d := math.Mod(s-(s0+float64(i)/float64(drops+1))+10, 1)
				if d < 0.02 {
					for j := 0; j < 8; j++ {
						a := float64(j)*math.Pi/4 + t*3
						circle(dst, at.X+math.Cos(a)*(30+d*900), at.Y+math.Sin(a)*(14+d*400)-20, 6, rgba(0xffffff, 0.8))
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
	for i := 0; i < len(pts); i += 6 { // les colonnes et leur bras
		q := pts[i]
		rect(dst, q.X-9, q.Y-20, 18, 80, rgb(0x5a6a7a))
		rect(dst, q.X-26, q.Y+58, 52, 10, rgb(0x8a9aaa))
	}
	drawLoop(dst, pts, 18, 0x1f3a7a)
	drawLoop(dst, pts, 8, 0xffd23f)
	s0 := rideStart(r, pts)
	// les vrilles : une spirale autour des rails
	for i := 1; i <= level/2; i++ {
		s := s0 + float64(i)/float64(level/2+1)
		for k := 0; k < 24; k++ {
			at, dir := along(pts, s+float64(k)*0.002)
			n := V{X: -dir.Y, Y: dir.X}
			a := float64(k) * math.Pi / 6
			q := at.Add(n.Mul(18 * math.Sin(a)))
			circle(dst, q.X, q.Y-10*math.Cos(a), 4, rgb(0xffd23f))
		}
	}
	pos := ridePos(r, s0)
	units := p.RideUnits(r)
	gap := 40.0 / pathLen(pts)
	for k := 0; k < units; k++ {
		at, dir := along(pts, pos-float64(k)*gap)
		polygon(dst, box(at, dir, 16, 10), rgb(0x2b2b3a)) // le chariot accroché aux rails
		n := ridersIn(r, k, 2)
		for i := 0; i < 2; i++ {
			sx := at.Add(dir.Mul(float64(i)*14 - 7))
			roundRect(dst, sx.X-8, sx.Y+6, 16, 16, 5, rgb(0x4fc3ff))
			if i < n {
				circle(dst, sx.X, sx.Y+2, 7, rgb(0xf2c9a0))
				sw := 4 * math.Sin(t*8+float64(k+i)) // les jambes balancent
				line(dst, sx.X-3, sx.Y+22, sx.X-3+sw, sx.Y+40, 3, rgb(0x2b4a8a))
				line(dst, sx.X+3, sx.Y+22, sx.X+3-sw, sx.Y+40, 3, rgb(0x2b4a8a))
			}
		}
	}
}
