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

// drawGhostRide : une maison hantée qui grandit (des tours aux tailles 3, 6 et 9),
// un cimetière, des citrouilles, des arbres morts, des fantômes et des chauves-souris.
// Les wagons disparaissent dans la maison.
func drawGhostRide(dst *ebiten.Image, p *sim.Place, r *sim.Ride, level int, t float64) {
	pts := trackLoop(r, level, 0.15)
	f := float64(level)
	// le sol devient sombre
	c := centroid(pts)
	tr := r.Track()
	ellipse(dst, c.X, c.Y, math.Min(tr.W/2, 260+60*f), math.Min(tr.H/2, 170+30*f), rgba(0x3b3b2b, 0.25))
	for i, q := range scatter(r, pts, 3+3*level, 11, 70) {
		switch i % 3 {
		case 0: // une tombe
			roundRect(dst, q.X-14, q.Y-30, 28, 34, 10, rgb(0x8a8a8a))
			line(dst, q.X, q.Y-24, q.X, q.Y-8, 3, rgb(0x5a5a5a))
			line(dst, q.X-6, q.Y-19, q.X+6, q.Y-19, 3, rgb(0x5a5a5a))
		case 1: // une citrouille
			circle(dst, q.X, q.Y, 14, rgb(0xff8a3d))
			rect(dst, q.X-2, q.Y-18, 4, 6, rgb(0x2f6a3a))
			circle(dst, q.X-5, q.Y-2, 2.5, rgb(0x3b2a10))
			circle(dst, q.X+5, q.Y-2, 2.5, rgb(0x3b2a10))
		default: // un arbre mort
			line(dst, q.X, q.Y, q.X, q.Y-60, 6, rgb(0x4a3a2a))
			line(dst, q.X, q.Y-40, q.X-20, q.Y-62, 4, rgb(0x4a3a2a))
			line(dst, q.X, q.Y-48, q.X+18, q.Y-70, 4, rgb(0x4a3a2a))
		}
	}
	drawLoop(dst, pts, 24, 0x2b2b2b)
	drawLoop(dst, pts, 6, 0x7a5a9a)
	// la maison, au centre du circuit
	bw := math.Min(160+55*f, (pts[0].X-c.X)*1.5)
	bh := 110 + 26*f
	h := sim.Rect{X: c.X - bw/2, Y: c.Y - bh/2, W: bw, H: bh}
	s0 := rideStart(r, pts)
	pos := ridePos(r, s0)
	units := p.RideUnits(r)
	gap := 48.0 / pathLen(pts)
	for k := 0; k < units; k++ { // les wagons dehors (dans la maison, on ne les voit pas)
		at, dir := along(pts, pos-float64(k)*gap)
		if h.Grow(10).Has(at) {
			continue
		}
		polygon(dst, box(at, dir, 20, 15), rgb(0x2b2b3a))
		circle(dst, at.X+dir.X*14, at.Y+dir.Y*14-6, 7, rgb(0xd9d9d9)) // la tête de mort à l'avant
		heads(dst, at, dir, ridersIn(r, k, 2), 14)
	}
	ellipse(dst, h.C().X, h.Y+h.H, bw*0.55, 22, rgba(0x000000, 0.25))
	rect(dst, h.X, h.Y-30, h.W, h.H+30, rgb(0x4a3a5a))
	polygon(dst, []V{{X: h.X - 20, Y: h.Y - 30}, {X: h.C().X, Y: h.Y - 30 - 60 - 6*f}, {X: h.X + h.W + 20, Y: h.Y - 30}}, rgb(0x2b1b3b))
	for i := 0; i < int(2+f/2); i++ { // fenêtres qui luisent
		x := h.X + 30 + float64(i)*(h.W-60)/math.Max(1, 1+f/2)
		glowc := rgba(0xffd23f, 0.55+0.45*math.Sin(t*3+float64(i)))
		roundRect(dst, x, h.Y+10, 24, 30, 6, glowc)
	}
	roundRect(dst, h.C().X-30, h.Y+h.H-60, 60, 60, 20, rgb(0x1b1b1b)) // la porte
	tower := func(x float64, k float64) {
		rect(dst, x-26*k, h.Y-60*k, 52*k, h.H+60*k, rgb(0x3b2b4b))
		polygon(dst, []V{{X: x - 36*k, Y: h.Y - 60*k}, {X: x, Y: h.Y - 150*k}, {X: x + 36*k, Y: h.Y - 60*k}}, rgb(0x2b1b3b))
		roundRect(dst, x-8*k, h.Y-40*k, 16*k, 22*k, 5, rgba(0xffd23f, 0.8))
	}
	if level >= 3 {
		tower(h.X+20, 1)
	}
	if level >= 6 {
		tower(h.X+h.W-20, 1.15)
	}
	if level >= 9 {
		tower(h.C().X, 1.5)
	}
	// les fantômes et les chauves-souris
	for i := 0; i < level/2; i++ {
		a := t*0.8 + float64(i)*1.7
		x, y := c.X+math.Cos(a)*(bw*0.7), c.Y-bh-40+math.Sin(a*1.3)*30
		ellipse(dst, x, y, 18, 24, rgba(0xffffff, 0.75))
		circle(dst, x-6, y-6, 3, rgb(0x2b2b2b))
		circle(dst, x+6, y-6, 3, rgb(0x2b2b2b))
	}
	for i := 0; i < (level+1)/2; i++ {
		a := t*1.6 + float64(i)*2.3
		x, y := c.X+math.Cos(a)*(bw*0.5+40), h.Y-120-30*math.Sin(a*2)
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
