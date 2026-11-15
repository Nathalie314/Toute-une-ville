package main

import (
	"math"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// ---------------------------------------------------------------- décor
//
// Le style « nature » de toute la ville : partout où il reste de l'herbe libre
// (autour des lieux, entre les zones), on pose des bosquets d'arbres, des
// sapins, des buissons, des rochers et des fleurs. Les endroits libres sont
// trouvés une fois pour toutes (et recalculés quand le lieu change) : on
// dessine le sol du lieu en tout petit, et on garde les cases restées vides.
// La lumière vient d'en haut à gauche : ombres vers le bas à droite, côté
// gauche des formes plus clair.

type decorKind int

const (
	dTree decorKind = iota
	dPine
	dBush
	dRock
	dFlowers
	dTuft
	dFountain // une fontaine décorative (rare)
)

type decor struct {
	x, y, z float64
	kind    decorKind
	seed    int
}

type decorCache struct {
	key  string
	tall []decor // ont une hauteur : triés avec les personnages
	flat []decor // au sol
}

var decorCaches = map[string]*decorCache{}

const decorCell = 20.0 // taille d'une case du masque (unités du monde)

// decorBounds : la partie du monde décorée (le lieu et une large bordure).
func decorBounds(p *sim.Place) (x0, y0, x1, y1 float64) {
	top := p.Top
	if p.DecoTop < top {
		top = p.DecoTop
	}
	return p.MinX() - 1600, top - 1400, p.MaxX() + 1600, p.H + 1400
}

// decorKey : ce qui change la place libre (bâtiments construits, parking du parc…).
func decorKey(p *sim.Place) string {
	n := 0
	for _, s := range p.Stations {
		if s.Built {
			n++
		}
	}
	k := p.ID + "/" + strconv.Itoa(n) + "/" + strconv.Itoa(len(p.Parks))

	return k
}

func hash2(x, y, s int) float64 {
	h := uint32(x*374761393 + y*668265263 + s*1442695041)
	h = (h ^ (h >> 13)) * 1274126177
	h ^= h >> 16
	return float64(h&0xffffff) / float64(0x1000000)
}

// noise2 : bruit doux (valeurs entre 0 et 1), pour grouper les arbres en bosquets.
func noise2(x, y float64, s int) float64 {
	ix, iy := math.Floor(x), math.Floor(y)
	fx, fy := x-ix, y-iy
	fx, fy = fx*fx*(3-2*fx), fy*fy*(3-2*fy)
	a := hash2(int(ix), int(iy), s)
	b := hash2(int(ix)+1, int(iy), s)
	c := hash2(int(ix), int(iy)+1, s)
	d := hash2(int(ix)+1, int(iy)+1, s)
	return a + (b-a)*fx + (c-a)*fy + (a-b-c+d)*fx*fy
}

// placeDecor : le décor du lieu (calculé au besoin).
func (g *game) placeDecor(p *sim.Place) *decorCache {
	key := decorKey(p)
	if c := decorCaches[p.ID]; c != nil && c.key == key {
		return c
	}
	c := &decorCache{key: key}
	decorCaches[p.ID] = c
	x0, y0, x1, y1 := decorBounds(p)
	w, h := int((x1-x0)/decorCell)+1, int((y1-y0)/decorCell)+1
	img := ebiten.NewImage(w, h)
	defer img.Deallocate()
	img.Fill(rgb(0xff00ff)) // (case restée de cette couleur : herbe libre)
	saved := cam
	cam = view{s: 1 / decorCell, ox: -x0 / decorCell, oy: -y0 / decorCell}
	tl, br := V{X: x0, Y: y0}, V{X: x1, Y: y1}
	g.drawGround(img, p, tl, br)
	block := func(r sim.Rect, m float64) {
		if r.W > 0 && r.H > 0 {
			rect(img, r.X-m, r.Y-m, r.W+2*m, r.H+2*m, rgb(0))
		}
	}
	for _, s := range p.Stations {
		block(s.Body, 40)
		block(s.Zone, 30)
		block(s.OutZone, 30)
		block(s.OutZone2, 30)
	}
	for _, u := range p.Unlocks {
		block(u.Zone, 40)
	}
	for _, sp := range p.Spots {
		block(sp.Zone, 40)
	}
	block(p.Yard, 60)
	if p.Parc != nil {
		for _, r := range p.Parc.Rides {
			block(r.Plot, 30)
			if r.Kind.ID == "bateaux" && p.RideBuilt(r) {
				block(sim.ParcSlot(3), 30) // (la rivière des petits bateaux y monte)
				block(sim.Rect{X: sim.ParcBoulEnd, Y: sim.ParcBoulY - 400, W: 2000, H: 800}, 30)
			}
		}
		// tout le futur parking reste dégagé (il sort de l'herbe au fil des achats)
		x0, x1 := 1e9, -1e9
		for _, s := range p.Parks {
			x0, x1 = math.Min(x0, s.Pos.X), math.Max(x1, s.Pos.X)
		}
		block(sim.Rect{X: x0 - 200, Y: 1640 + 150, W: x1 - x0 + 400, H: p.H - 1640}, 60)
	}
	flush()
	cam = saved
	pix := make([]byte, 4*w*h)
	img.ReadPixels(pix)
	free := func(i, j int) bool {
		if i < 0 || j < 0 || i >= w || j >= h {
			return false
		}
		k := 4 * (j*w + i)
		return pix[k] > 240 && pix[k+1] < 20 && pix[k+2] > 240
	}
	// distance (en cases) à la chose la plus proche, jusqu'à 9
	dist := make([]uint8, w*h)
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			d := uint8(0)
			if free(i, j) {
				d = 10
				for r := 1; r <= 9 && d == 10; r++ {
					for k := -r; k <= r; k++ {
						if !free(i+k, j-r) || !free(i+k, j+r) || !free(i-r, j+k) || !free(i+r, j+k) {
							d = uint8(r)
							break
						}
					}
				}
			}
			dist[j*w+i] = d
		}
	}
	room := func(x, y float64) int {
		i, j := int((x-x0)/decorCell), int((y-y0)/decorCell)
		if i < 0 || j < 0 || i >= w || j >= h {
			return 0
		}
		return int(dist[j*w+i])
	}
	// ce qui a une hauteur ne doit pas cacher le lieu : on garde aussi du vide au-dessus
	// (un arbre dessine sa couronne plus haut que son pied)
	tallRoom := func(x, y, z float64) bool {
		return room(x, y) >= 4 && room(x, y-100*z) >= 4 && room(x+60*z, y-40*z) >= 3 && room(x-60*z, y-40*z) >= 3
	}
	// dans le lieu lui-même, pas de forêt : il doit rester aéré (seulement quelques
	// arbres isolés, des buissons, des fleurs) ; les bosquets sont tout autour
	inside := func(x, y float64) bool {
		top := p.Top
		if p.DecoTop < top {
			top = p.DecoTop
		}
		return p.ID != "parc" && x > p.MinX() && x < p.MaxX() && y > top && y < p.H // (au parc, les bosquets vont bien)
	}
	const step = 64.0
	sd := len(p.ID) * 7
	for _, ch := range p.ID {
		sd = sd*31 + int(ch)
	}
	for gy := math.Floor(y0/step) * step; gy < y1; gy += step {
		for gx := math.Floor(x0/step) * step; gx < x1; gx += step {
			ci, cj := int(gx/step), int(gy/step)
			r := hash2(ci, cj, sd)
			x := gx + (hash2(ci, cj, sd+1)-0.5)*step*0.9
			y := gy + (hash2(ci, cj, sd+2)-0.5)*step*0.9
			if room(x, y) == 0 {
				continue
			}
			forest := noise2(x/900, y/900, sd+3)
			meadow := noise2(x/500, y/500, sd+4)
			z := 0.75 + 0.45*hash2(ci, cj, sd+5)
			seed := int(hash2(ci, cj, sd+6) * 1e6)
			in := inside(x, y)
			// de temps en temps, une fontaine (assez de place tout autour, loin des autres)
			if hash2(ci, cj, sd+9) < 0.012 && room(x, y) >= 9 && room(x-150, y) >= 7 && room(x+150, y) >= 7 && room(x, y-150) >= 7 && room(x, y+80) >= 6 {
				far := true
				for _, o := range c.tall {
					if o.kind == dFountain && math.Hypot(o.x-x, o.y-y) < 2500 {
						far = false
					}
				}
				if far && len(fountainsIn(c)) < 3 {
					c.tall = append(c.tall, decor{x, y, 0.8, dFountain, seed})
					continue
				}
			}
			switch {
			case !in && forest > 0.62 && r < (forest-0.5)*1.6 && tallRoom(x, y, z):
				k := dTree
				if noise2(x/1400, y/1400, sd+7) > 0.55 || hash2(ci, cj, sd+8) < 0.2 {
					k = dPine
				}
				c.tall = append(c.tall, decor{x, y, z, k, seed})
			case r < 0.035 && (!in || r < 0.015) && tallRoom(x, y, z):
				c.tall = append(c.tall, decor{x, y, z, dTree, seed})
			case r < 0.07 && room(x, y) >= 3:
				c.tall = append(c.tall, decor{x, y, z * 0.9, dBush, seed})
			case r < 0.085 && room(x, y) >= 3:
				c.tall = append(c.tall, decor{x, y, z, dRock, seed})
			case meadow > 0.6 && r < 0.4 && room(x, y) >= 2:
				c.flat = append(c.flat, decor{x, y, z, dFlowers, seed})
			case r < 0.22 && room(x, y) >= 1:
				c.flat = append(c.flat, decor{x, y, z, dTuft, seed})
			}
		}
	}
	// rien d'autre autour d'une fontaine
	fs := fountainsIn(c)
	keep := func(ds []decor) []decor {
		out := ds[:0]
		for _, d := range ds {
			ok := true
			for _, f := range fs {
				if d.kind != dFountain && math.Abs(d.x-f.x) < 260 && d.y > f.y-120 && d.y < f.y+260 {
					ok = false
				}
			}
			if ok {
				out = append(out, d)
			}
		}
		return out
	}
	c.tall, c.flat = keep(c.tall), keep(c.flat)
	return c
}

// drawGround : le sol d'un lieu (sans l'herbe ni ce qui a une hauteur).
func (g *game) drawGround(dst *ebiten.Image, p *sim.Place, tl, br V) {
	switch p.ID {
	case "ferme":
		g.farmGround(dst, p, tl, br)
	case "super":
		g.superGround(dst, p, tl, br)
	case "ecole":
		g.schoolGround(dst, p, tl, br)
	case "elevage":
		g.elevageGround(dst, p, tl, br)
	case "exotique":
		g.exoticGround(dst, p, tl, br)
	case "port":
		g.portGround(dst, p, tl, br)
	case "domaine":
		g.domaineGround(dst, p, tl, br)
	case "parc":
		g.parcGround(dst, p, tl, br)
	default:
		g.workshopGround(dst, p, tl, br)
	}
	drawParking(dst, p, tl, br)
	g.drawYard(dst, p)
}

// drawGrass : l'herbe, avec de grandes taches plus claires ou plus foncées
// (le terrain n'est plus tout plat) et des touffes.
func drawGrass(dst *ebiten.Image, p *sim.Place, tl, br V) {
	rect(dst, tl.X-10, tl.Y-10, br.X-tl.X+20, br.Y-tl.Y+20, rgb(0x86c25e))
	const st = 260.0
	for y := math.Floor((tl.Y-st)/st) * st; y < br.Y+st; y += st {
		for x := math.Floor((tl.X-st)/st) * st; x < br.X+st; x += st {
			i, j := int(x/st), int(y/st)
			n := noise2(x/1300, y/1300, 11)
			cx, cy := x+(hash2(i, j, 3)-0.5)*st, y+(hash2(i, j, 4)-0.5)*st
			r := st * (0.6 + 0.5*hash2(i, j, 5))
			switch {
			case n > 0.58:
				ellipse(dst, cx, cy, r*1.3, r, rgba(0x9ad06c, 0.45))
			case n < 0.38:
				ellipse(dst, cx, cy, r*1.3, r, rgba(0x6fae4c, 0.35))
			}
		}
	}
	for i := 0; i < 900; i++ {
		x := math.Mod(float64(i)*733.13, p.W+600) - 300
		y := math.Mod(float64(i)*397.71+float64(i*i)*0.37, p.H+600) - 300
		if x < tl.X-20 || x > br.X+20 || y < tl.Y-20 || y > br.Y+20 {
			continue
		}
		c := rgb(0x77b552)
		if i%3 == 0 {
			c = rgb(0x93cc6b)
		}
		line(dst, x, y, x-4, y-10, 3, c)
		line(dst, x+4, y, x+6, y-9, 3, c)
	}
}

// drawFlatDecor : fleurs et touffes (au sol, sous tout le reste).
func drawFlatDecor(dst *ebiten.Image, c *decorCache, tl, br V) {
	flowerCols := []uint32{0xffffff, 0xffd84a, 0xff7aa8, 0xb48cff, 0xff8a4a}
	for _, d := range c.flat {
		if d.x < tl.X-60 || d.x > br.X+60 || d.y < tl.Y-60 || d.y > br.Y+60 {
			continue
		}
		switch d.kind {
		case dFlowers:
			col := flowerCols[d.seed%len(flowerCols)]
			for k := 0; k < 6; k++ {
				a := float64(k)*1.9 + float64(d.seed%7)
				x, y := d.x+math.Cos(a)*18*d.z, d.y+math.Sin(a)*11*d.z
				line(dst, x, y, x, y+7, 2, rgb(0x4f9a3c))
				circle(dst, x, y, 4.5*d.z, rgb(col))
				circle(dst, x, y, 1.8*d.z, rgb(0xffe36a))
			}
		case dTuft:
			cl := rgb(0x6aa647)
			for k := -2; k <= 2; k++ {
				fk := float64(k)
				line(dst, d.x+fk*4, d.y, d.x+fk*7*d.z, d.y-(14-math.Abs(fk)*3)*d.z, 3, cl)
			}
		}
	}
}

// decorDrawables : arbres, sapins, buissons et rochers, triés avec le reste.
func decorDrawables(dst *ebiten.Image, c *decorCache, ds []drawable, tl, br V) []drawable {
	for _, d := range c.tall {
		d := d
		if d.x < tl.X-150 || d.x > br.X+150 || d.y < tl.Y-40 || d.y > br.Y+260 {
			continue
		}
		ds = append(ds, drawable{x: d.x, y: d.y, fn: func() {
			switch d.kind {
			case dTree:
				drawTree(dst, d.x, d.y, d.z)
			case dPine:
				drawPine(dst, d.x, d.y, d.z)
			case dBush:
				drawBush(dst, d.x, d.y, d.z, d.seed)
			case dRock:
				drawRock(dst, d.x, d.y, d.z, d.seed)
			case dFountain:
				drawFancyFountain(dst, d.seed, V{X: d.x, Y: d.y}, d.z, gameT)
			}
		}})
	}
	return ds
}

// dropShadow : ombre portée vers le bas à droite (la lumière vient d'en haut à gauche).
func dropShadow(dst *ebiten.Image, x, y, rx, ry, h float64) {
	ellipse(dst, x+h*0.35, y+ry*0.2, rx+h*0.25, ry, rgba(0x1d3a12, 0.22))
}

// drawTree : un arbre rond, en volume (côté éclairé, côté à l'ombre).
func drawTree(dst *ebiten.Image, x, y, z float64) {
	if cam.iso { // (debout : seul son pied est projeté)
		upright(x, y, func() { drawTree(dst, x, y, z) })
		return
	}
	dropShadow(dst, x, y, 40*z, 15*z, 70*z)
	rect(dst, x-8*z, y-46*z, 16*z, 48*z, rgb(0x7a4e2c))
	rect(dst, x-8*z, y-46*z, 6*z, 48*z, rgb(0x9a6a40))
	cy := y - 82*z
	circle(dst, x+4*z, cy+6*z, 50*z, rgb(0x2f7a32))
	circle(dst, x-24*z, cy+20*z, 32*z, rgb(0x2f7a32))
	circle(dst, x+26*z, cy+18*z, 30*z, rgb(0x2f7a32))
	circle(dst, x-2*z, cy, 46*z, rgb(0x43993d))
	circle(dst, x-22*z, cy+14*z, 28*z, rgb(0x4aa645))
	circle(dst, x-12*z, cy-14*z, 28*z, rgb(0x5cb94f))
	circle(dst, x-20*z, cy-22*z, 13*z, rgb(0x7fd36a))
}

// drawPine : un sapin à facettes (moitié gauche éclairée).
func drawPine(dst *ebiten.Image, x, y, z float64) {
	if cam.iso { // (debout : seul son pied est projeté)
		upright(x, y, func() { drawPine(dst, x, y, z) })
		return
	}
	dropShadow(dst, x, y, 30*z, 12*z, 110*z)
	rect(dst, x-6*z, y-24*z, 12*z, 26*z, rgb(0x6e4526))
	for k := 0; k < 3; k++ {
		fk := float64(k)
		w := (46 - fk*11) * z
		by := y - (18+fk*38)*z
		ty := by - (62-fk*6)*z
		polygon(dst, []V{{X: x, Y: ty}, {X: x, Y: by + 6*z}, {X: x - w, Y: by}}, rgb(0x4f9e48))
		polygon(dst, []V{{X: x, Y: ty}, {X: x + w, Y: by}, {X: x, Y: by + 6*z}}, rgb(0x2e6e34))
	}
}

// drawBush : un buisson (trois boules), parfois fleuri.
func drawBush(dst *ebiten.Image, x, y, z float64, seed int) {
	if cam.iso { // (debout : seul son pied est projeté)
		upright(x, y, func() { drawBush(dst, x, y, z, seed) })
		return
	}
	dropShadow(dst, x, y, 34*z, 10*z, 22*z)
	circle(dst, x-16*z, y-14*z, 18*z, rgb(0x3a8a3a))
	circle(dst, x+16*z, y-14*z, 17*z, rgb(0x2f7a32))
	circle(dst, x, y-24*z, 22*z, rgb(0x4aa645))
	circle(dst, x-7*z, y-31*z, 10*z, rgb(0x6cc75a))
	if seed%3 == 0 {
		col := []uint32{0xff7aa8, 0xffffff, 0xffd84a}[seed%9/3]
		for k := 0; k < 5; k++ {
			a := float64(k)*1.3 + float64(seed%5)
			circle(dst, x+math.Cos(a)*16*z, y-20*z+math.Sin(a)*10*z, 3.5*z, rgb(col))
		}
	}
}

// drawRock : un rocher à facettes.
func drawRock(dst *ebiten.Image, x, y, z float64, seed int) {
	if cam.iso { // (debout : seul son pied est projeté)
		upright(x, y, func() { drawRock(dst, x, y, z, seed) })
		return
	}
	s := 0.8 + 0.4*float64(seed%5)/4
	w, h := 34*z*s, 26*z*s
	dropShadow(dst, x, y, w, 9*z, 20*z)
	top := V{X: x - w*0.15, Y: y - h}
	polygon(dst, []V{{X: x - w, Y: y}, {X: x - w*0.8, Y: y - h*0.6}, top, {X: x + w*0.1, Y: y - h*0.35}, {X: x - w*0.1, Y: y + 2}}, rgb(0xc9cbc4))
	polygon(dst, []V{top, {X: x + w*0.6, Y: y - h*0.7}, {X: x + w, Y: y - 2}, {X: x - w*0.1, Y: y + 2}, {X: x + w*0.1, Y: y - h*0.35}}, rgb(0x8e918b))
	polygon(dst, []V{{X: x - w*0.8, Y: y - h*0.6}, top, {X: x + w*0.6, Y: y - h*0.7}, {X: x + w*0.1, Y: y - h*0.35}}, rgb(0xe2e3dd))
}

// slab : une dalle en relief (ombre portée, tranche plus foncée en bas, dessus).
func slab(dst *ebiten.Image, x, y, w, h, r, thick float64, top uint32) {
	roundRect(dst, x+thick*0.9, y+thick*1.6, w, h, r, rgba(0x1d3a12, 0.20))
	roundRect(dst, x, y+thick, w, h, r, shade(top, 0.72))
	roundRect(dst, x, y, w, h, r, rgb(top))
}

// stationShadow : l'ombre d'une machine ou d'un bâtiment sur le sol.
func stationShadow(dst *ebiten.Image, s *sim.Station) {
	b := s.Body
	if b.W <= 0 || b.H <= 0 {
		return
	}
	roundRect(dst, b.X+12, b.Y+16, b.W, b.H, 12, rgba(0x000000, 0.10))
}

// mapTrees : des arbres autour des tuiles de la carte de la ville (pas dessus,
// ni sur le titre, la rivière ou les boutons du bas).
func mapTrees(dst *ebiten.Image, bx, by, bw, bh float64) {
	type tr struct {
		x, y, z float64
		pine    bool
	}
	var ts []tr
	for i := 0; i < 400; i++ {
		x := hash2(i, 1, 77) * (scr.VW - 180)
		y := 40 + hash2(i, 2, 77)*(scr.VH-40)
		z := 0.38 + 0.2*hash2(i, 3, 77)
		if x > bx-40*z && x < bx+bw+40*z && y > by && y < by+bh+90*z {
			continue // sur les tuiles
		}
		if y < 200 && (math.Abs(x-scr.VW/2) < 420 || x < 520 || x > scr.VW-460) || y > scr.VH-130 {
			continue // le titre, les boutons
		}
		ok := true
		for _, o := range ts {
			if math.Abs(o.x-x) < 46 && math.Abs(o.y-y) < 40 {
				ok = false
				break
			}
		}
		if ok {
			ts = append(ts, tr{x, y, z, hash2(i, 4, 77) < 0.35})
		}
	}
	for i := 1; i < len(ts); i++ { // tri par profondeur
		for j := i; j > 0 && ts[j].y < ts[j-1].y; j-- {
			ts[j], ts[j-1] = ts[j-1], ts[j]
		}
	}
	for _, t := range ts {
		if t.pine {
			drawPine(dst, t.x, t.y, t.z)
		} else {
			drawTree(dst, t.x, t.y, t.z)
		}
	}
}

func fountainsIn(c *decorCache) []decor {
	var out []decor
	for _, d := range c.tall {
		if d.kind == dFountain {
			out = append(out, d)
		}
	}
	return out
}

// gameT : le temps du jeu (pour animer le décor : jets d'eau…).
var gameT float64
