package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// Dessin de l'école.

var levelFloor = []uint32{0xfff1c9, 0xdff2d0, 0xd8e9f7, 0xe8dcf5}
var levelCol = []uint32{0xf1c40f, 0x2ecc71, 0x3498db, 0x8e44ad}

func (g *game) schoolGround(dst *ebiten.Image, p *sim.Place, tl, br V) {
	road(dst, tl, br, 1640)
	W := sim.SchoolW
	// l'enceinte de l'école
	rect(dst, 80, 110, W-120, 1430, rgb(0x9c9488))
	rect(dst, 96, 126, W-152, 1398, rgb(0xd9d2c4))
	drawKidsYard(dst, g.t)
	drawRecreation(dst, g.t)
	// le couloir entre les deux rangées
	rect(dst, sim.SchoolX0-20, sim.SchoolRow1Y+sim.SchoolRoomH, sim.SchoolCol(8)-sim.SchoolX0+20, sim.SchoolRow2Y-sim.SchoolRow1Y-sim.SchoolRoomH, rgb(0xcfc6b4))
	wc := rgb(0x7a6f60)
	for _, r := range p.Rooms {
		b := r.Body
		if !r.Built() {
			roundRect(dst, b.X, b.Y, b.W, b.H, 10, rgba(0x000000, 0.06))
			dashedRect(dst, b.X+4, b.Y+4, b.W-8, b.H-8, 4, 18, 0, rgba(0x8a8478, 0.7))
			say(dst, r.Name, b.C().X, b.Y+30, 20, true, 0, rgba(0x6a6458, 0.8))
			continue
		}
		rect(dst, b.X, b.Y, b.W, b.H, rgb(levelFloor[r.Level]))
		for x := b.X; x < b.X+b.W; x += 50 {
			line(dst, x, b.Y, x, b.Y+b.H, 1, rgba(0x000000, 0.05))
		}
		if r.Row == 2 {
			drawAmphi(dst, g, p, r)
			continue
		}
		// murs
		rect(dst, b.X-6, b.Y-6, 10, b.H+12, wc)
		rect(dst, b.X+b.W-4, b.Y-6, 10, b.H+12, wc)
		if r.Row == 0 {
			rect(dst, b.X-6, b.Y-6, b.W+12, 10, wc)
			rect(dst, b.X-6, b.Y+b.H-4, b.W/2-50, 10, wc)
			rect(dst, b.X+b.W/2+56, b.Y+b.H-4, b.W/2-50, 10, wc)
		} else {
			rect(dst, b.X-6, b.Y+b.H-4, b.W+12, 10, wc)
			rect(dst, b.X-6, b.Y-6, b.W-100, 10, wc)
			rect(dst, b.X+b.W-10, b.Y-6, 16, 10, wc)
		}
		// tableau noir (ou l'écran du vidéoprojecteur)
		bx := b.X + 160
		roundRect(dst, bx, b.Y+18, 250, 64, 6, rgb(0x6e4322))
		roundRect(dst, bx+6, b.Y+24, 238, 52, 4, rgb(0x2f4a3a))
		say(dst, r.Name, bx+125, b.Y+50, 22, true, 0, rgb(0xf3f3e8))
		if p.HasVideo(r) {
			drawProjector(dst, V{X: bx + 125, Y: b.Y + 50}, V{X: bx + 300, Y: b.Y + 112}, p.Teaching(r), g.t)
		}
		// bureau du prof
		roundRect(dst, b.X+30, b.Y+96, 96, 40, 6, rgb(0x8a5a33))
		// bandeau de couleur du niveau
		rect(dst, b.X+4, b.Y+4, 10, b.H-8, rgb(levelCol[r.Level]))
	}
	// l'accueil
	a := sim.SchoolAccueil()
	rect(dst, a.X, a.Y, a.W, a.H, rgb(0xf3ead8)) // (pas de murs : on y circule librement)
	for k := 0; k < 3; k++ {                     // plantes vertes
		x := a.X + 40 + float64(k)*25
		roundRect(dst, x-10, a.Y+a.H-60, 20, 24, 4, rgb(0xc0623a))
		circle(dst, x, a.Y+a.H-70, 14, rgb(0x4aa043))
	}
	// le local des fournitures
	f := sim.SchoolFourn()
	rect(dst, f.X, f.Y, f.W, f.H, rgb(0xe9dcc0))
	for _, w := range sim.FournWalls() {
		rect(dst, w.X, w.Y, w.W, w.H, wc)
	}
	// (le panneau, entre deux colonnes d'étagères)
	roundRect(dst, f.X+850-130, f.Y+f.H-46, 260, 32, 8, rgb(0x8a5a33))
	say(dst, "LOCAL DES FOURNITURES", f.X+850, f.Y+f.H-30, 16, true, 0, rgb(0xffffff))
	// façade et portail
	gx := p.Gate.X
	roundRect(dst, gx-160, 1490, 320, 44, 10, rgb(0xd4a017))
	say(dst, "ÉCOLE DE CLAIRBOURG", gx, 1512, 22, true, 0, rgb(0xffffff))
	rect(dst, 80, 1530, gx-160-80, 14, rgb(0x6f6a62))
	rect(dst, gx+160, 1530, W-40-gx-160, 14, rgb(0x6f6a62))
}

// drawProjector : le vidéoprojecteur (au plafond) et son faisceau sur le tableau.
func drawProjector(dst *ebiten.Image, screen, at V, on bool, t float64) {
	if on {
		polygon(dst, []V{{X: at.X - 10, Y: at.Y}, {X: screen.X - 120, Y: screen.Y - 30}, {X: screen.X + 120, Y: screen.Y - 30}, {X: at.X + 10, Y: at.Y}}, rgba(0xfff6c8, 0.22))
		roundRect(dst, screen.X-118, screen.Y-28, 236, 56, 4, rgb(0xf6f6f0))
		// une diapositive qui change
		k := int(t/4) % 3
		switch k {
		case 0:
			for i := 0; i < 4; i++ {
				rect(dst, screen.X-90+float64(i)*45, screen.Y+12-float64(i+1)*8, 30, float64(i+1)*8, rgb(0x3498db))
			}
		case 1:
			circle(dst, screen.X, screen.Y, 20, rgb(0x2ecc71))
			polygon(dst, []V{{X: screen.X, Y: screen.Y}, {X: screen.X + 20, Y: screen.Y}, {X: screen.X, Y: screen.Y - 20}}, rgb(0xf1c40f))
		default:
			for i := 0; i < 3; i++ {
				rect(dst, screen.X-80, screen.Y-14+float64(i)*12, 160-float64(i)*40, 5, rgb(0x7a7a7a))
			}
		}
	}
	roundRect(dst, at.X-24, at.Y-12, 48, 24, 6, rgb(0x3b3f45))
	circle(dst, at.X, at.Y-12, 7, rgb(0x9fd3ff))
	line(dst, at.X, at.Y-12, at.X, at.Y-40, 3, rgb(0x555555))
}

// drawAmphi : l'amphithéâtre de l'université (bancs en arc, grand écran).
func drawAmphi(dst *ebiten.Image, g *game, p *sim.Place, r *sim.Room) {
	b := r.Body
	wc := rgb(0x7a6f60)
	rect(dst, b.X-6, b.Y-6, b.W+12, 10, wc)
	rect(dst, b.X-6, b.Y+b.H-4, b.W+12, 10, wc)
	rect(dst, b.X+b.W-4, b.Y-6, 10, b.H+12, wc)
	cy := (sim.SchoolRow1Y + sim.SchoolRoomH + sim.SchoolRow2Y) / 2
	rect(dst, b.X-6, b.Y-6, 10, cy-50-b.Y+6, wc)
	rect(dst, b.X-6, cy+56, 10, b.Y+b.H+6-cy-56, wc)
	c := V{X: b.X + b.W/2, Y: b.Y + 110}
	// gradins en arc
	for k := 0; k < 4; k++ {
		rad := 220.0 + float64(k)*110
		var pts []V
		for a := -0.9; a <= 0.9; a += 0.05 {
			pts = append(pts, V{X: c.X + math.Sin(a)*rad, Y: c.Y + math.Cos(a)*rad*0.9})
		}
		for i := 1; i < len(pts); i++ {
			line(dst, pts[i-1].X, pts[i-1].Y, pts[i].X, pts[i].Y, 14, rgba(0x8a5a33, 0.35+0.1*float64(k)))
		}
	}
	// grand tableau et estrade
	roundRect(dst, c.X-200, b.Y+16, 400, 70, 6, rgb(0x6e4322))
	roundRect(dst, c.X-192, b.Y+22, 384, 58, 4, rgb(0x2f4a3a))
	say(dst, "AMPHITHÉÂTRE", c.X, b.Y+51, 24, true, 0, rgb(0xf3f3e8))
	roundRect(dst, c.X-90, b.Y+96, 180, 30, 6, rgb(0x8a5a33))
	if p.HasVideo(r) {
		drawProjector(dst, V{X: c.X, Y: b.Y + 51}, V{X: c.X + 250, Y: b.Y + 140}, p.Teaching(r), g.t)
	}
	rect(dst, b.X+4, b.Y+4, 10, b.H-8, rgb(levelCol[r.Level]))
}

// drawKidsYard : la cour des petits (parcours de tricycles, jeux, marelle).
func drawKidsYard(dst *ebiten.Image, t float64) {
	x0, x1 := 110.0, 700.0
	roundRect(dst, x0, 140, x1-x0, 1370, 20, rgb(0xb9dd9a))
	say(dst, "COUR DES PETITS", (x0+x1)/2, 168, 18, true, 0, rgba(0x3f6a2a, 0.8))
	// parcours de tricycles : une boucle
	c := V{X: (x0 + x1) / 2, Y: 470}
	for a := 0.0; a < 2*math.Pi; a += 0.04 {
		x, y := c.X+math.Cos(a)*220, c.Y+math.Sin(a)*200
		circle(dst, x, y, 34, rgb(0xd98c5f))
	}
	for a := 0.0; a < 2*math.Pi; a += 0.3 {
		circle(dst, c.X+math.Cos(a)*220, c.Y+math.Sin(a)*200, 3, rgba(0xffffff, 0.7))
	}
	ellipse(dst, c.X, c.Y, 170, 150, rgb(0xb9dd9a))
	say(dst, "Parcours tricycles", c.X, c.Y, 16, true, 0, rgba(0x3f6a2a, 0.8))
	// toboggan et balançoires
	y := 900.0
	rect(dst, x0+60, y-120, 10, 120, rgb(0x8a5a33))
	rect(dst, x0+120, y-120, 10, 120, rgb(0x8a5a33))
	rect(dst, x0+60, y-126, 70, 12, rgb(0xe74c3c))
	polygon(dst, []V{{X: x0 + 130, Y: y - 116}, {X: x0 + 240, Y: y - 10}, {X: x0 + 220, Y: y}, {X: x0 + 125, Y: y - 100}}, rgb(0xf1c40f))
	bx := x1 - 230
	line(dst, bx, y, bx+20, y-140, 6, rgb(0x5d6d7e))
	line(dst, bx+190, y, bx+170, y-140, 6, rgb(0x5d6d7e))
	line(dst, bx+20, y-140, bx+170, y-140, 6, rgb(0x5d6d7e))
	for k := 0; k < 2; k++ {
		sx := bx + 65 + float64(k)*60
		sw := math.Sin(t*2+float64(k)) * 14
		line(dst, sx, y-140, sx+sw, y-50, 2, rgb(0x333333))
		roundRect(dst, sx+sw-16, y-52, 32, 8, 3, rgb(0x3498db))
	}
	say(dst, "Jeux", (x0+x1)/2, y+30, 16, true, 0, rgba(0x3f6a2a, 0.8))
	// marelle
	mx, my := (x0+x1)/2-45, 1080.0
	cells := [][2]float64{{0, 300}, {0, 240}, {-45, 180}, {45, 180}, {0, 120}, {-45, 60}, {45, 60}, {0, 0}}
	for i, c := range cells {
		x, y := mx+c[0], my+c[1]
		rect(dst, x, y, 90, 60, rgba(0xffffff, 0.25))
		roundStroke(dst, x, y, 90, 60, 2, 3, rgb(0xffffff))
		say(dst, itoa(i+1), x+45, y+30, 22, true, 0, rgb(0xffffff))
	}
	roundRect(dst, mx-10, my-60, 110, 50, 25, rgba(0xffffff, 0.25))
	say(dst, "CIEL", mx+45, my-35, 16, true, 0, rgb(0xffffff))
}

// tricycles : leur place sur le parcours (pour les dessiner avec les personnages).
func tricyclePos(i int, t float64) (V, V) {
	c := V{X: 405, Y: 470}
	a := t*0.35 + float64(i)*2.1
	pos := V{X: c.X + math.Cos(a)*220, Y: c.Y + math.Sin(a)*200}
	dir := V{X: -math.Sin(a), Y: math.Cos(a)}
	return pos, dir
}

// drawRecreation : la cour de récréation (basket, ping-pong, pique-nique, bancs, parc).
func drawRecreation(dst *ebiten.Image, t float64) {
	x0, x1 := sim.SchoolX0-20, sim.SchoolW-70
	y0, y1 := 140.0, 500.0
	rect(dst, x0, y0, x1-x0, y1-y0, rgb(0xb8b2a6))
	say(dst, "COUR DE RÉCRÉATION", x0+20, y0+20, 18, true, -1, rgba(0x4a4a4a, 0.7))
	// terrain de basket
	bx := x0 + 40.0
	rect(dst, bx, y0+40, 640, 300, rgb(0xd9773a))
	roundStroke(dst, bx+8, y0+48, 624, 284, 2, 4, rgb(0xffffff))
	line(dst, bx+320, y0+48, bx+320, y0+332, 4, rgb(0xffffff))
	ring(dst, bx+320, y0+190, 50, 4, rgb(0xffffff))
	for _, side := range []float64{0, 1} {
		hx := bx + 30 + side*580
		rect(dst, hx-4+side*8-4, y0+160, 8, 60, rgb(0x555555))
		ring(dst, hx+20-side*40, y0+190, 14, 3, rgb(0xe74c3c))
	}
	// ping-pong
	px := bx + 720
	for k := 0; k < 2; k++ {
		x := px + float64(k)*260
		roundRect(dst, x, y0+120, 200, 110, 6, rgb(0x2f6aa8))
		line(dst, x+100, y0+116, x+100, y0+234, 4, rgb(0xffffff))
		roundStroke(dst, x+3, y0+123, 194, 104, 4, 2, rgb(0xffffff))
		ball := math.Sin(t*3+float64(k)) * 80
		circle(dst, x+100+ball, y0+175, 5, rgb(0xff8a3d))
	}
	// tables de pique-nique
	qx := px + 560
	for k := 0; k < 4; k++ {
		x := qx + float64(k%2)*230
		y := y0 + 70 + float64(k/2)*160
		rect(dst, x, y+20, 170, 50, rgb(0xb07d4a))
		rect(dst, x, y, 170, 14, rgb(0x8a5a33))
		rect(dst, x, y+76, 170, 14, rgb(0x8a5a33))
	}
	// bancs, sous les arbres
	cx := qx + 520
	for k := 0; k < 6; k++ {
		x := cx + float64(k%3)*230
		y := y0 + 100 + float64(k/3)*170
		drawTree(dst, x+80, y-10, 0.6)
		rect(dst, x, y+20, 160, 16, rgb(0x8a5a33))
		rect(dst, x, y, 160, 8, rgb(0x6e4322))
		rect(dst, x+10, y+36, 8, 18, rgb(0x555555))
		rect(dst, x+142, y+36, 8, 18, rgb(0x555555))
	}
	// le joli parc
	gx := cx + 720
	roundRect(dst, gx, y0+20, x1-gx-20, y1-y0-40, 30, rgb(0x8cc56a))
	roundRect(dst, gx+40, y0+160, x1-gx-100, 40, 20, rgb(0xd9c08f)) // allée
	ellipse(dst, gx+(x1-gx)/2, y0+180, 70, 40, rgb(0x9fd3ff))       // bassin
	ellipse(dst, gx+(x1-gx)/2, y0+180, 55, 28, rgb(0x5aa7d6))
	for k := 0; k < 3; k++ {
		h := 10 + 6*math.Sin(t*3+float64(k))
		line(dst, gx+(x1-gx)/2-10+float64(k)*10, y0+178, gx+(x1-gx)/2-10+float64(k)*10, y0+178-h, 3, rgba(0xffffff, 0.8))
	}
	cols := []uint32{0xff6fa8, 0xffd23f, 0xb07cff, 0xff8a3d, 0xffffff}
	for k := 0; k < 40; k++ {
		x := gx + 30 + math.Mod(float64(k)*97, x1-gx-80)
		y := y0 + 50 + math.Mod(float64(k)*53, 90)
		if k%2 == 1 {
			y += 170
		}
		circle(dst, x, y, 6, rgb(cols[k%len(cols)]))
	}
	for k := 0; k < 5; k++ {
		drawTree(dst, gx+60+float64(k)*((x1-gx-120)/4), y0+150+float64(k%2)*170, 0.7)
	}
}

func (g *game) drawSchoolDesk(dst *ebiten.Image, b sim.Rect) {
	ellipse(dst, b.C().X, b.Y+b.H, b.W*0.55, 14, rgba(0x000000, 0.18))
	roundRect(dst, b.X, b.Y, b.W, b.H, 10, rgb(0x6f8fa8))
	roundRect(dst, b.X, b.Y, b.W, 22, 8, rgb(0x8fb0c8))
	say(dst, "ACCUEIL", b.C().X, b.Y+44, 18, true, 0, rgb(0xffffff))
	// registre
	roundRect(dst, b.X+150, b.Y-10, 46, 30, 4, rgb(0xfff8ea))
	line(dst, b.X+173, b.Y-10, b.X+173, b.Y+20, 2, rgb(0xc0623a))
}

// schoolDrawables : tables, élèves et profs, à trier avec le reste.
func (g *game) schoolDrawables(dst *ebiten.Image, p *sim.Place, ds []drawable) []drawable {
	t := g.t
	// des petits sur leurs tricycles
	for i := 0; i < 3; i++ {
		pos, dir := tricyclePos(i, t)
		lk := lookFrom(4242 + i*977)
		ds = append(ds, drawable{pos.Y, func() {
			circle(dst, pos.X-dir.X*14, pos.Y+6, 7, rgb(0x2b2b2b))
			circle(dst, pos.X+dir.X*14, pos.Y+6, 9, rgb(0x2b2b2b))
			line(dst, pos.X-dir.X*14, pos.Y, pos.X+dir.X*14, pos.Y-4, 4, rgb(0xe74c3c))
			scaled(pos.X, pos.Y, 0.6, func() { drawPerson(dst, pos.Add(V{Y: -10}), dir, 0, false, lk, nil) })
		}})
	}
	for _, r := range p.Rooms {
		if !r.Built() {
			continue
		}
		r := r
		for n := 0; n < p.Tables(r); n++ {
			c := r.SeatPos(n)
			ds = append(ds, drawable{c.Y + 20, func() {
				ellipse(dst, c.X, c.Y+22, 46, 8, rgba(0x000000, 0.15))
				roundRect(dst, c.X-44, c.Y-14, 88, 36, 6, rgb(0xb07d4a))
				roundRect(dst, c.X-44, c.Y-14, 88, 10, 5, rgb(0xc8955e))
				rect(dst, c.X-40, c.Y+20, 6, 14, rgb(0x6e4322))
				rect(dst, c.X+34, c.Y+20, 6, 14, rgb(0x6e4322))
				// cahier
				rect(dst, c.X-14, c.Y-10, 24, 16, rgb(0xffffff))
			}})
		}
		if r.Want >= 0 || r.BoostT > 0 {
			ds = append(ds, drawable{r.Body.Y + 2000, func() { g.drawRoomRequest(dst, r) }}) // par-dessus tout
		}
		// formation express : la case, puis l'intervenant (costume, mallette) et son compte à rebours
		if p.ExpressOK(r) {
			z := r.ExpressZone()
			ds = append(ds, drawable{z.Y, func() {
				in := g.sim.Player != nil && z.Has(g.sim.Player.Pos)
				zoneBox(dst, z, in, t, 0xb48cff)
				sayFit(dst, "Formation express", z.C().X, z.Y+18, 15, true, 0, z.W-8, rgb(0x5a3a9a))
				roundRect(dst, z.C().X-58, z.Y+36, 116, 30, 9, rgba(0x2b2140, 0.85))
				coin(dst, z.C().X-42, z.Y+51, 9)
				sayFit(dst, money(p.ExpressCost(r)), z.C().X+8, z.Y+51, 17, true, 0, 84, rgb(0xffd23f))
			}})
		}
		if ex := r.Expert; ex != nil {
			ds = append(ds, drawable{ex.Pos.Y, func() {
				lk := lookFrom(4242 * 31)
				lk.shirt, lk.pants = 0x2b3a5a, 0x2b2b3a
				drawPerson(dst, ex.Pos, ex.Face, ex.WalkT, ex.Moving, lk, nil)
				roundRect(dst, ex.Pos.X+10, ex.Pos.Y-6, 22, 16, 3, rgb(0x6e4322)) // mallette
				if r.ExpertT > 0 {
					k := 1 - r.ExpertT/sim.ExpressTime
					roundRect(dst, ex.Pos.X-60, ex.Pos.Y-110, 120, 40, 10, rgba(0x2b2140, 0.9))
					say(dst, "Formation : "+itoa(int(math.Ceil(r.ExpertT)))+" s", ex.Pos.X, ex.Pos.Y-98, 15, true, 0, rgb(0xffffff))
					rect(dst, ex.Pos.X-50, ex.Pos.Y-84, 100, 6, rgba(0xffffff, 0.3))
					rect(dst, ex.Pos.X-50, ex.Pos.Y-84, 100*k, 6, rgb(0xb48cff))
				}
			}})
		}
		if pr := r.Prof; pr != nil {
			ds = append(ds, drawable{pr.Pos.Y, func() {
				lk := lookFrom(pr.Look * 31)
				lk.shirt, lk.pants = 0x7a5a3a, 0x3b3b4a
				drawPerson(dst, pr.Pos, V{X: 1}, t, false, lk, nil)
			}})
		}
	}
	for _, s := range p.Students {
		s := s
		a := s.A
		k := 0.72
		lv := sim.Primaire
		if s.Room != nil {
			lv = s.Room.Level
			k = []float64{0.72, 0.82, 0.92, 1}[lv]
		}
		ds = append(ds, drawable{a.Pos.Y, func() {
			lk := lookFrom(a.Look)
			scaled(a.Pos.X, a.Pos.Y, k, func() {
				drawPerson(dst, a.Pos, a.Face, a.WalkT, a.Moving, lk, nil)
				// cartable
				if s.State != 2 {
					roundRect(dst, a.Pos.X-14, a.Pos.Y-60, 28, 26, 6, rgb(levelCol[lv]))
				}
			})
			if s.State == 2 {
				// avancement de la formation
				x, y := a.Pos.X, a.Pos.Y-80*k-30
				roundRect(dst, x-30, y, 60, 12, 6, rgba(0x000000, 0.35))
				roundRect(dst, x-28, y+2, 56*math.Min(1, s.Prog), 8, 4, rgb(levelCol[lv]))
				if !p.Teaching(s.Room) {
					say(dst, "z", x+36, y-6+math.Sin(t*2+x)*3, 16, true, 0, rgba(0x5a5a8a, 0.8))
				}
			}
		}})
	}
	return ds
}

// drawRoomRequest : bulle du matériel demandé, ou « ×2 » pendant la vitesse double.
func (g *game) drawRoomRequest(dst *ebiten.Image, r *sim.Room) {
	b := r.Body
	x, y := b.X+300, b.Y+50
	if r.BoostT > 0 {
		k := 0.5 + 0.5*math.Sin(g.t*6)
		roundRect(dst, b.X+164, b.Y+12, 272, 76, 10, rgba(0xffd23f, 0.35+0.3*k))
		roundRect(dst, x+110, y-60, 90, 50, 14, rgb(0xffd23f))
		say(dst, "×2", x+155, y-36, 30, true, 0, rgb(0x6e4322))
		roundRect(dst, x+114, y-14, 82, 8, 4, rgba(0x000000, 0.3))
		roundRect(dst, x+114, y-14, 82*r.BoostT/60, 8, 4, rgb(0x2fa86b))
		return
	}
	bx, by := x+150, y-46
	circle(dst, bx, by, 36, rgba(0xffffff, 0.97))
	circle(dst, bx-30, by+28, 7, rgba(0xffffff, 0.97))
	ring(dst, bx, by, 36, 3, rgba(0xd4a017, 0.6+0.4*math.Sin(g.t*5)))
	drawItem(dst, r.Want, bx, by, 22)
}

// drawSupplyShelf : une étagère du local, avec son matériel.
func drawSupplyShelf(dst *ebiten.Image, s *sim.Station) {
	b := s.Body
	ellipse(dst, b.C().X, b.Y+b.H, b.W*0.55, 10, rgba(0x000000, 0.18))
	roundRect(dst, b.X, b.Y-30, b.W, b.H+30, 6, rgb(0x8a5a33))
	for k := 0; k < 3; k++ {
		rect(dst, b.X+6, b.Y-24+float64(k)*28, b.W-12, 4, rgb(0x6e4322))
	}
	for k := 0; k < 3; k++ {
		drawItem(dst, s.Item, b.X+26+float64(k)*40, b.Y-8, 12)
	}
	sayFit(dst, s.Name, b.C().X, b.Y-44, 15, true, 0, b.W+40, rgb(0x5a3a1a))
}
