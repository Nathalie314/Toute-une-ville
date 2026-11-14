package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// La nouvelle ferme (4.1) : la serre et le verger (une colonne par légume ou
// par arbre), les caisses de graines, les petits marchés et l'étang.

var cropCol = map[sim.Item]uint32{
	sim.Carotte: 0xf08a24, sim.Tomate: 0xe23b2e, sim.Courgette: 0x3f8f3a, sim.Poireau: 0xdfe8c8, sim.Chou: 0x9fe07a, sim.PommeDeTerre: 0xc9a46a,
	sim.Fraise: 0xe8333f, sim.Pomme: 0xd9342b, sim.Poire: 0xc9d24a, sim.Cerise: 0x9e1530, sim.Abricot: 0xf5a03a, sim.Figue: 0x6b3a6e,
}

// serreRect : la serre couvre ses six colonnes (vide tant qu'on ne l'a pas achetée).
func serreRect(p *sim.Place) (sim.Rect, bool) {
	a, b := p.Station("carotte"), p.Station("pdt")
	if a == nil || b == nil {
		return sim.Rect{}, false
	}
	r := sim.Rect{X: a.Body.X - 20, Y: a.Body.Y - 30, W: b.Body.X + b.Body.W + 20 - (a.Body.X - 20), H: a.Body.H + 40}
	return r, a.Built
}

// drawSerreFloor : le sol et les montants de la serre (dessinés sous les plantes).
func drawSerreFloor(dst *ebiten.Image, p *sim.Place) {
	r, ok := serreRect(p)
	if !ok {
		return
	}
	roundRect(dst, r.X-8, r.Y-8, r.W+16, r.H+16, 10, rgb(0x9aa3a6))
	rect(dst, r.X, r.Y, r.W, r.H, rgb(0xd8e6dc))
	for y := r.Y + 30; y < r.Y+r.H; y += 60 {
		line(dst, r.X, y, r.X+r.W, y, 1.5, rgba(0xb7c9bc, 0.8))
	}
	// les vitres : un reflet en biais
	for x := r.X + 40; x < r.X+r.W; x += 150 {
		line(dst, x, r.Y+20, x+60, r.Y+90, 6, rgba(0xffffff, 0.35))
	}
}

// drawSerreRoof : le faîte de la serre, au-dessus de tout.
func drawSerreRoof(dst *ebiten.Image, p *sim.Place) {
	r, ok := serreRect(p)
	if !ok {
		return
	}
	frame := rgb(0x7f8a8e)
	// pignon vitré en haut
	polygon(dst, []V{{X: r.X - 14, Y: r.Y}, {X: r.X + r.W/2, Y: r.Y - 70}, {X: r.X + r.W + 14, Y: r.Y}}, rgba(0xcfe9f2, 0.75))
	line(dst, r.X-14, r.Y, r.X+r.W/2, r.Y-70, 6, frame)
	line(dst, r.X+r.W/2, r.Y-70, r.X+r.W+14, r.Y, 6, frame)
	rect(dst, r.X-14, r.Y-4, r.W+28, 8, frame)
	// montants verticaux, entre les colonnes
	for k := 0; k <= 6; k++ {
		x := r.X + 10 + float64(k)*sim.ColumnW
		if k == 6 {
			x = r.X + r.W - 10
		}
		rect(dst, x-3, r.Y, 6, r.H, rgba(0x7f8a8e, 0.75))
	}
	roundRect(dst, r.X+r.W/2-80, r.Y-58, 160, 38, 8, rgb(0x2f6a3a))
	say(dst, "SERRE", r.X+r.W/2, r.Y-39, 22, true, 0, rgb(0xe8f5e0))
}

// drawColumn : une colonne de la serre ou du verger, ses cases et la prochaine à acheter.
func (g *game) drawColumn(dst *ebiten.Image, s *sim.Station) {
	b := s.Body
	if !s.Built {
		return // (la case d'achat se dessine toute seule)
	}
	serre := s.Serre
	// cases pas encore achetées : un contour discret
	for i := len(s.Plants) + 1; i < len(s.Slots); i++ {
		q := s.Slots[i]
		dashedRect(dst, q.X-50, q.Y-38, 100, 76, 2, 10, 0, rgba(0x6a5a40, 0.25))
	}
	for i := len(s.Plants) - 1; i >= 0; i-- { // (du haut vers le bas : l'arbre de devant cache celui de derrière)
		pl := s.Plants[i]
		q := pl.P
		if serre {
			roundRect(dst, q.X-56, q.Y-40, 112, 80, 10, rgb(0x8b6239))
			roundRect(dst, q.X-52, q.Y-36, 104, 72, 8, rgb(0xa47447))
			for k := -1; k <= 1; k++ {
				line(dst, q.X-44, q.Y+float64(k)*20, q.X+44, q.Y+float64(k)*20, 2, rgba(0x7a5230, 0.5))
			}
		} else {
			ellipse(dst, q.X, q.Y+8, 46, 18, rgb(0x96683e))
		}
		if pl.Empty() {
			// case vide : des petits trous, et le sachet qu'il faut apporter
			for k := -1; k <= 1; k++ {
				circle(dst, q.X+float64(k)*26, q.Y+4, 6, rgb(0x6e4a2a))
			}
			a := 0.55 + 0.45*math.Sin(g.t*4+float64(i))
			drawSeedPacket(dst, s.Item, q.X, q.Y-30, 14, a)
			continue
		}
		k := 1.5 // (plus grand que dans les champs : une case = une plante)
		if !serre {
			yg, gr, tt, it := pl.Young, pl.G, g.t+float64(i)*0.7, s.Item
			scaled(q.X, q.Y+12, 1.15, func() { drawOrchard(dst, it, q.X, q.Y+12, yg, gr, tt) })
			continue
		}
		gr, tt, ii, it := pl.G, g.t+float64(i)*0.7, i, s.Item
		scaled(q.X, q.Y+12, k, func() { drawCrop(dst, it, q.X, q.Y+12, gr, tt, ii) })
	}
	// la prochaine case à acheter
	if len(s.Plants) < len(s.Slots) {
		g.drawPlantBuy(dst, s)
	}
	// petit panneau
	roundRect(dst, b.X+b.W/2-66, b.Y+b.H-6, 132, 34, 8, rgb(0x7a4a26))
	sayFit(dst, s.Name, b.X+b.W/2, b.Y+b.H+11, 18, true, 0, 124, rgb(0xffe7a8))
}

// drawPlantBuy : la case « +1 » de la prochaine pousse (ou case) à acheter.
func (g *game) drawPlantBuy(dst *ebiten.Image, s *sim.Station) {
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

// drawSeedPacket : un sachet de graines (avec l'image du légume).
func drawSeedPacket(dst *ebiten.Image, veg sim.Item, x, y, r, a float64) {
	roundRect(dst, x-r*0.75, y-r, r*1.5, r*2, r*0.2, rgba(0xf6ecd2, a))
	rect(dst, x-r*0.75, y-r, r*1.5, r*0.45, rgba(cropCol[veg], a))
	if a > 0.6 {
		drawItem(dst, veg, x, y+r*0.3, r*0.55)
	}
}

// drawMarketStall : un étal du petit marché, avec son auvent rayé.
func (g *game) drawMarketStall(dst *ebiten.Image, s *sim.Station) {
	b := s.Body
	stripe := uint32(0x2f8a4a) // légumes : vert
	for _, f := range sim.Fruits {
		if f == s.Item {
			stripe = 0xd9342b // fruits : rouge
		}
	}
	ellipse(dst, b.C().X, b.Y+b.H, b.W*0.55, 11, rgba(0x000000, 0.18))
	// poteaux et auvent
	rect(dst, b.X+4, b.Y-58, 6, b.H+58, rgb(0x8a5a33))
	rect(dst, b.X+b.W-10, b.Y-58, 6, b.H+58, rgb(0x8a5a33))
	for k := 0; k < 6; k++ {
		c := uint32(0xffffff)
		if k%2 == 0 {
			c = stripe
		}
		w := (b.W + 16) / 6
		x := b.X - 8 + float64(k)*w
		polygon(dst, []V{{X: x, Y: b.Y - 70}, {X: x + w, Y: b.Y - 70}, {X: x + w, Y: b.Y - 46}, {X: x + w/2, Y: b.Y - 38}, {X: x, Y: b.Y - 46}}, rgb(c))
	}
	// table et cagettes
	roundRect(dst, b.X, b.Y, b.W, b.H, 6, rgb(0x9a6b3e))
	roundRect(dst, b.X+6, b.Y+4, b.W-12, 24, 5, rgb(0xc9935a))
	for i := 0; i < min(s.Stock, 15); i++ {
		c := i % 5
		l := i / 5
		drawItem(dst, s.Item, b.X+18+float64(c)*22, b.Y+14-float64(l)*11, 11)
	}
	price := money(sim.Items[s.Item].Price * g.sim.PriceMult())
	roundRect(dst, b.C().X-36, b.Y+34, 72, 24, 6, rgb(0xfff8e6))
	say(dst, price, b.C().X, b.Y+46, 16, true, 0, rgb(0x4a3a20))
	if s.Stock == 0 && math.Mod(g.t, 1) < 0.6 {
		roundRect(dst, b.C().X-34, b.Y-30, 68, 24, 6, rgb(0xe74c3c))
		say(dst, "VIDE", b.C().X, b.Y-18, 16, true, 0, rgb(0xffffff))
	}
}

// drawPond : un étang naturel (décor), avec ses roseaux et ses nénuphars.
func drawPond(dst *ebiten.Image, b sim.Rect, t float64) {
	cx, cy := b.C().X, b.C().Y
	// berge
	var bank, water []V
	n := 48
	for k := 0; k < n; k++ {
		a := float64(k) / float64(n) * 2 * math.Pi
		w := 1 + 0.06*math.Sin(a*3) + 0.04*math.Cos(a*5)
		bank = append(bank, V{X: cx + math.Cos(a)*b.W/2*w, Y: cy + math.Sin(a)*b.H/2*w})
		water = append(water, V{X: cx + math.Cos(a)*(b.W/2-26)*w, Y: cy + math.Sin(a)*(b.H/2-20)*w})
	}
	polygon(dst, bank, rgb(0xc9b48a))
	polygon(dst, water, rgb(0x5aa7d6))
	// reflets qui bougent
	for k := 0; k < 9; k++ {
		x := cx - b.W*0.35 + float64(k)*b.W*0.09
		y := cy - b.H*0.15 + math.Sin(float64(k)*1.7)*b.H*0.2
		dx := math.Sin(t*0.8+float64(k)) * 12
		line(dst, x-30+dx, y, x+30+dx, y, 3, rgba(0xffffff, 0.35))
	}
	// nénuphars
	for k := 0; k < 7; k++ {
		x := cx - b.W*0.3 + float64(k)*b.W*0.1
		y := cy + math.Cos(float64(k)*2.3)*b.H*0.22
		ellipse(dst, x, y, 22, 12, rgb(0x4f9a3a))
		polygon(dst, []V{{X: x, Y: y}, {X: x + 22, Y: y - 4}, {X: x + 22, Y: y + 4}}, rgb(0x5aa7d6))
		if k%2 == 0 {
			circle(dst, x-6, y-4, 6, rgb(0xffb6d0))
			circle(dst, x-6, y-4, 2.5, rgb(0xffe066))
		}
	}
	// roseaux sur les bords
	for k := 0; k < 14; k++ {
		a := float64(k)/14*2*math.Pi + 0.2
		x := cx + math.Cos(a)*(b.W/2-18)
		y := cy + math.Sin(a)*(b.H/2-12)
		for j := -1; j <= 1; j++ {
			sw := math.Sin(t*1.5+float64(k)) * 3
			line(dst, x+float64(j)*6, y, x+float64(j)*8+sw, y-46, 3, rgb(0x5f8a3a))
			if j == 0 {
				roundRect(dst, x-3+sw, y-50, 7, 18, 3, rgb(0x6b4423))
			}
		}
	}
	// un petit ponton de bois
	px := cx + b.W*0.32
	rect(dst, px-30, cy+b.H*0.3, 60, 70, rgb(0xa47447))
	for k := 0; k < 4; k++ {
		line(dst, px-30, cy+b.H*0.3+float64(k)*18, px+30, cy+b.H*0.3+float64(k)*18, 2, rgb(0x7a5230))
	}
}

// drawCrop : les plantes de la serre et du verger.
func drawCrop(dst *ebiten.Image, it sim.Item, x, y, gr, t float64, i int) {
	sway := 0.0
	ripe := gr >= 1
	if ripe {
		sway = math.Sin(t*2) * 2
	}
	h := 8 + 40*gr
	green, dark := rgb(0x52b246), rgb(0x3f8f3a)
	fc := rgb(cropCol[it])
	tree := func(crown float64, leaf uint32, fruit func(fx, fy float64)) {
		th := 30 + 40*gr
		ellipse(dst, x, y+2, 30*math.Max(0.4, gr), 8, rgba(0x000000, 0.2))
		rect(dst, x-4, y-th, 8, th, rgb(0x8a5a33))
		cr := (10 + 22*gr) * crown
		circle(dst, x+sway*0.3, y-th-cr*0.4, cr, rgb(leaf))
		circle(dst, x-cr*0.6, y-th-cr*0.1, cr*0.7, shade(leaf, 0.85))
		circle(dst, x+cr*0.6, y-th-cr*0.05, cr*0.65, shade(leaf, 1.1))
		if gr > 0.5 && gr < 1 {
			for k := 0; k < 4; k++ {
				circle(dst, x-cr*0.5+float64(k)*cr*0.35, y-th-cr*0.3-float64(k%2)*8, 3, rgb(0xffffff)) // fleurs
			}
		}
		if ripe {
			for k := 0; k < 5; k++ {
				fruit(x-cr*0.6+float64(k)*cr*0.3+sway*0.3, y-th-cr*0.1-float64((k*7)%3)*cr*0.3)
			}
		}
	}
	switch it {
	case sim.Carotte:
		for k := -2; k <= 2; k++ {
			line(dst, x, y, x+float64(k)*7+sway, y-h*0.8, 3, green)
		}
		if gr > 0.4 {
			polygon(dst, []V{{X: x - 8, Y: y - 2}, {X: x + 8, Y: y - 2}, {X: x, Y: y + 8 + 6*gr}}, fc)
		}
	case sim.Tomate:
		line(dst, x+10, y, x+10, y-h-10, 3, rgb(0x9a6b3e)) // tuteur
		line(dst, x, y, x+4+sway, y-h, 4, dark)
		for k := 0; k < 3; k++ {
			ellipse(dst, x-8+float64(k)*8, y-h*0.3-float64(k)*h*0.22, 9, 5, green)
		}
		if gr > 0.5 {
			c := fc
			if !ripe {
				c = rgb(0x8fcf4a)
			}
			circle(dst, x-6, y-h*0.45, 6, c)
			circle(dst, x+6, y-h*0.7, 6, c)
			circle(dst, x-2, y-h*0.9, 5, c)
		}
	case sim.Courgette:
		for k := -1; k <= 1; k++ {
			ellipse(dst, x+float64(k)*14+sway*0.3, y-h*0.35-math.Abs(float64(k))*-4, 13*math.Max(0.4, gr), 9*math.Max(0.4, gr), green)
		}
		if gr > 0.5 {
			circle(dst, x, y-h*0.5, 4, rgb(0xffd23f))
		}
		if ripe {
			roundRect(dst, x-26, y-6, 26, 9, 4, fc)
			roundRect(dst, x+4, y-2, 24, 9, 4, shade(cropCol[it], 0.85))
		}
	case sim.Poireau:
		rect(dst, x-4, y-h*0.5, 8, h*0.5, rgb(0xf3f3e6))
		for k := -1; k <= 1; k++ {
			line(dst, x, y-h*0.45, x+float64(k)*12+sway, y-h*1.1, 5, rgb(0x4a8f3a))
		}
	case sim.Chou: // une salade : des feuilles frisées, ouvertes
		r := 6 + 14*gr
		for k := 0; k < 7; k++ {
			a := float64(k) * 0.9
			ellipse(dst, x+math.Cos(a)*r*0.55, y-r*0.6+math.Sin(a)*r*0.3, r*0.5, r*0.35, rgb(0x5fae3a))
		}
		circle(dst, x, y-r*0.7, r*0.55, fc)
		circle(dst, x+2, y-r*0.75, r*0.3, rgb(0xd8f5a8))
	case sim.PommeDeTerre:
		r := 6 + 12*gr
		circle(dst, x-6, y-r*0.6, r*0.8, green)
		circle(dst, x+7, y-r*0.7, r*0.75, dark)
		if gr > 0.6 {
			circle(dst, x-4, y-r*1.2, 3, rgb(0xffffff))
			circle(dst, x+6, y-r*1.3, 3, rgb(0xffffff))
		}
		if ripe {
			ellipse(dst, x-12, y+4, 7, 5, fc)
			ellipse(dst, x+12, y+5, 6, 5, fc)
		}
	case sim.Fraise:
		r := 6 + 12*gr
		for k := -1; k <= 1; k++ {
			circle(dst, x+float64(k)*12, y-r*0.6, r*0.6, green)
		}
		if gr > 0.5 && !ripe {
			circle(dst, x-6, y-r*0.9, 3, rgb(0xffffff))
		}
		if ripe {
			for k := -1; k <= 1; k++ {
				polygon(dst, []V{{X: x + float64(k)*12 - 5, Y: y - 4}, {X: x + float64(k)*12 + 5, Y: y - 4}, {X: x + float64(k)*12, Y: y + 6}}, fc)
			}
		}
	case sim.Pomme:
		tree(1, 0x4aa043, func(fx, fy float64) { circle(dst, fx, fy, 5, fc) })
	case sim.Poire:
		tree(1, 0x55a84a, func(fx, fy float64) {
			circle(dst, fx, fy+2, 5, fc)
			circle(dst, fx, fy-3, 3.5, fc)
		})
	case sim.Cerise:
		tree(1.05, 0x3f9a45, func(fx, fy float64) {
			line(dst, fx, fy-8, fx-3, fy, 1.5, rgb(0x4a3a20))
			line(dst, fx, fy-8, fx+4, fy, 1.5, rgb(0x4a3a20))
			circle(dst, fx-3, fy, 3.5, fc)
			circle(dst, fx+4, fy, 3.5, fc)
		})
	case sim.Abricot:
		tree(0.95, 0x5aae4a, func(fx, fy float64) { circle(dst, fx, fy, 5, fc) })
	case sim.Figue:
		tree(1.1, 0x2f7a3a, func(fx, fy float64) {
			circle(dst, fx, fy+1, 5, fc)
			polygon(dst, []V{{X: fx - 3, Y: fy - 2}, {X: fx + 3, Y: fy - 2}, {X: fx, Y: fy - 9}}, fc)
		})
	}
	if ripe {
		a := math.Max(0, math.Sin(t*3))
		circle(dst, x+14, y-h-8, 3*a, rgba(0xffffff, 0.8*a))
	}
}

// drawOrchard : une case du verger. D'abord la terre nue, puis l'arbre (ou l'arbuste,
// pour les fraises) petit, moyen, grand ; ensuite les fruits : petits, moyens, gros
// (encore verts), puis gros et de la bonne couleur : on peut cueillir. Après la
// cueillette, seuls les fruits disparaissent.
func drawOrchard(dst *ebiten.Image, it sim.Item, x, y, young, gr, t float64) {
	grow := 1 - young
	if young > 0 && grow < 0.2 {
		// case toute neuve : de la terre retournée
		ellipse(dst, x, y-2, 20, 8, rgb(0x6e4a2a))
		for k := -1; k <= 1; k++ {
			circle(dst, x+float64(k)*9, y-4, 2.5, rgb(0x8b6239))
		}
		return
	}
	size := 1.0 // grand
	if young > 0 {
		size = 0.4 // petit
		if grow >= 0.47 {
			size = 0.7 // moyen
		}
		if grow >= 0.73 {
			size = 1 // grand (pas encore de fruits)
		}
	}
	// fruits : 0 rien (l'arbre pousse encore), puis 3 tailles vertes, puis mûrs
	fr, ripe := 0.0, false
	if young == 0 {
		switch {
		case gr >= 1:
			fr, ripe = 5, true
		case gr >= 2.0/3:
			fr = 5
		case gr >= 1.0/3:
			fr = 3.5
		default:
			fr = 2
		}
	}
	sway := 0.0
	if ripe {
		sway = math.Sin(t*2) * 2
	}
	fc := rgb(cropCol[it])
	if !ripe {
		fc = rgb(0x8fcf4a) // pas encore de la bonne couleur
	}
	for _, e := range sim.ExoFruits {
		if e == it { // ferme exotique : ananas, bananier, cocotier, manguier
			drawExoticTree(dst, it, x, y, size, fr, ripe, t)
			return
		}
	}
	if it == sim.Fraise {
		// un fraisier : un arbuste bas
		r := 18 * size
		ellipse(dst, x, y+2, 26*size, 7, rgba(0x000000, 0.2))
		for k := -1; k <= 1; k++ {
			circle(dst, x+float64(k)*r*0.7, y-r*0.6, r*0.6, rgb(0x52b246))
		}
		circle(dst, x, y-r, r*0.55, rgb(0x3f8f3a))
		if fr > 0 {
			for k := -1; k <= 1; k++ {
				fx, fy := x+float64(k)*r*0.75+sway*0.3, y-r*0.2
				polygon(dst, []V{{X: fx - fr, Y: fy - fr*0.6}, {X: fx + fr, Y: fy - fr*0.6}, {X: fx, Y: fy + fr*1.2}}, fc)
			}
		}
		return
	}
	leaf := map[sim.Item]uint32{sim.Pomme: 0x4aa043, sim.Poire: 0x55a84a, sim.Cerise: 0x3f9a45, sim.Abricot: 0x5aae4a, sim.Figue: 0x2f7a3a}[it]
	th := 44 * size
	ellipse(dst, x, y+2, 30*size, 8, rgba(0x000000, 0.2))
	rect(dst, x-4*size-1, y-th, 8*size+2, th, rgb(0x8a5a33))
	cr := 28 * size
	circle(dst, x+sway*0.3, y-th-cr*0.4, cr, rgb(leaf))
	circle(dst, x-cr*0.6, y-th-cr*0.1, cr*0.7, shade(leaf, 0.85))
	circle(dst, x+cr*0.6, y-th-cr*0.05, cr*0.65, shade(leaf, 1.1))
	if fr == 0 {
		return
	}
	for k := 0; k < 5; k++ {
		fx := x - cr*0.6 + float64(k)*cr*0.3 + sway*0.3
		fy := y - th - cr*0.1 - float64((k*7)%3)*cr*0.3
		switch it {
		case sim.Poire:
			circle(dst, fx, fy+fr*0.4, fr, fc)
			circle(dst, fx, fy-fr*0.6, fr*0.7, fc)
		case sim.Cerise:
			line(dst, fx, fy-fr*1.6, fx-fr*0.6, fy, 1.5, rgb(0x4a3a20))
			line(dst, fx, fy-fr*1.6, fx+fr*0.8, fy, 1.5, rgb(0x4a3a20))
			circle(dst, fx-fr*0.6, fy, fr*0.7, fc)
			circle(dst, fx+fr*0.8, fy, fr*0.7, fc)
		case sim.Figue:
			circle(dst, fx, fy+fr*0.2, fr, fc)
			polygon(dst, []V{{X: fx - fr*0.6, Y: fy - fr*0.4}, {X: fx + fr*0.6, Y: fy - fr*0.4}, {X: fx, Y: fy - fr*1.8}}, fc)
		default: // pommes, abricots
			circle(dst, fx, fy, fr, fc)
		}
	}
	if ripe {
		a := math.Max(0, math.Sin(t*3))
		circle(dst, x+14, y-th-cr-8, 3*a, rgba(0xffffff, 0.8*a))
	}
}

// drawItemNew : les icônes des produits de la 4.1 (faux si ce n'en est pas un).
func drawItemNew(dst *ebiten.Image, it sim.Item, x, y, r float64) bool {
	fc := rgb(cropCol[it])
	switch it {
	case sim.Carotte:
		polygon(dst, []V{{X: x - r*0.35, Y: y - r*0.45}, {X: x + r*0.35, Y: y - r*0.45}, {X: x + r*0.05, Y: y + r}}, fc)
		for k := -1; k <= 1; k++ {
			line(dst, x, y-r*0.45, x+float64(k)*r*0.35, y-r, r*0.12, rgb(0x52b246))
		}
	case sim.Tomate:
		circle(dst, x, y+r*0.1, r*0.8, fc)
		circle(dst, x-r*0.3, y-r*0.15, r*0.18, rgba(0xffffff, 0.4))
		for k := 0; k < 5; k++ {
			a := float64(k) * 2 * math.Pi / 5
			line(dst, x, y-r*0.6, x+math.Cos(a)*r*0.35, y-r*0.6+math.Sin(a)*r*0.2, r*0.1, rgb(0x3f8f3a))
		}
	case sim.Courgette:
		polygon(dst, []V{{X: x - r*0.9, Y: y + r*0.35}, {X: x - r*0.7, Y: y + r*0.6}, {X: x + r*0.9, Y: y - r*0.35}, {X: x + r*0.7, Y: y - r*0.6}}, fc)
		line(dst, x-r*0.6, y+r*0.3, x+r*0.6, y-r*0.35, r*0.08, rgb(0x6fb04a))
		circle(dst, x+r*0.85, y-r*0.55, r*0.12, rgb(0x8a6a3a))
	case sim.Poireau:
		rect(dst, x-r*0.18, y-r*0.1, r*0.36, r*1.05, rgb(0xf3f3e6))
		for k := -1; k <= 1; k++ {
			line(dst, x, y, x+float64(k)*r*0.45, y-r, r*0.2, rgb(0x4a8f3a))
		}
	case sim.Chou: // une salade
		for k := 0; k < 7; k++ {
			a := float64(k) * 0.9
			ellipse(dst, x+math.Cos(a)*r*0.5, y+math.Sin(a)*r*0.3, r*0.45, r*0.32, rgb(0x5fae3a))
		}
		circle(dst, x, y-r*0.05, r*0.5, fc)
		circle(dst, x+r*0.1, y-r*0.12, r*0.25, rgb(0xd8f5a8))
	case sim.PommeDeTerre:
		ellipse(dst, x-r*0.3, y+r*0.2, r*0.5, r*0.38, fc)
		ellipse(dst, x+r*0.35, y-r*0.15, r*0.45, r*0.34, shade(cropCol[it], 1.08))
		circle(dst, x-r*0.4, y+r*0.1, r*0.06, rgb(0x8a6a3a))
		circle(dst, x+r*0.3, y-r*0.2, r*0.06, rgb(0x8a6a3a))
	case sim.Fraise:
		polygon(dst, []V{{X: x - r*0.7, Y: y - r*0.35}, {X: x + r*0.7, Y: y - r*0.35}, {X: x + r*0.45, Y: y + r*0.4}, {X: x, Y: y + r}, {X: x - r*0.45, Y: y + r*0.4}}, fc)
		for k := 0; k < 5; k++ {
			circle(dst, x-r*0.35+float64(k%3)*r*0.35, y+float64(k/3)*r*0.35, r*0.06, rgb(0xffe066))
		}
		polygon(dst, []V{{X: x - r*0.6, Y: y - r*0.35}, {X: x, Y: y - r*0.8}, {X: x + r*0.6, Y: y - r*0.35}, {X: x, Y: y - r*0.5}}, rgb(0x3f9b35))
	case sim.Pomme:
		circle(dst, x-r*0.25, y+r*0.1, r*0.6, fc)
		circle(dst, x+r*0.25, y+r*0.1, r*0.6, fc)
		line(dst, x, y-r*0.4, x+r*0.1, y-r*0.85, r*0.1, rgb(0x6b4423))
		ellipse(dst, x+r*0.35, y-r*0.65, r*0.25, r*0.12, rgb(0x52b246))
	case sim.Poire:
		circle(dst, x, y+r*0.35, r*0.6, fc)
		circle(dst, x, y-r*0.2, r*0.38, fc)
		line(dst, x, y-r*0.5, x+r*0.1, y-r*0.9, r*0.1, rgb(0x6b4423))
	case sim.Cerise:
		line(dst, x+r*0.1, y-r*0.9, x-r*0.35, y+r*0.2, r*0.08, rgb(0x4a6a2a))
		line(dst, x+r*0.1, y-r*0.9, x+r*0.4, y+r*0.25, r*0.08, rgb(0x4a6a2a))
		circle(dst, x-r*0.4, y+r*0.45, r*0.4, fc)
		circle(dst, x+r*0.42, y+r*0.5, r*0.4, fc)
		circle(dst, x-r*0.5, y+r*0.35, r*0.1, rgba(0xffffff, 0.5))
	case sim.Abricot:
		circle(dst, x, y+r*0.1, r*0.75, fc)
		line(dst, x+r*0.05, y-r*0.6, x-r*0.05, y+r*0.8, r*0.06, shade(cropCol[it], 0.85))
		ellipse(dst, x+r*0.3, y-r*0.65, r*0.25, r*0.12, rgb(0x52b246))
	case sim.Figue:
		circle(dst, x, y+r*0.3, r*0.62, fc)
		polygon(dst, []V{{X: x - r*0.45, Y: y}, {X: x + r*0.45, Y: y}, {X: x, Y: y - r*0.8}}, fc)
		circle(dst, x, y+r*0.3, r*0.2, rgba(0xd9768f, 0.8))
	case sim.GrainesCarotte, sim.GrainesTomate, sim.GrainesCourgette, sim.GrainesPoireau, sim.GrainesChou, sim.PlantsPdt:
		for v, sd := range sim.SeedOf {
			if sd == it {
				drawSeedPacket(dst, v, x, y, r*0.85, 1)
			}
		}
	case sim.SucreRoux, sim.SucreMorceaux:
		c := uint32(0xffffff)
		if it == sim.SucreRoux {
			c = 0xc98a4a
		}
		for k := 0; k < 3; k++ {
			cx := x + float64(k-1)*r*0.5
			cy := y + r*0.2 - float64(k%2)*r*0.45
			roundRect(dst, cx-r*0.3, cy-r*0.3, r*0.6, r*0.6, r*0.08, rgb(c))
			roundStroke(dst, cx-r*0.3, cy-r*0.3, r*0.6, r*0.6, r*0.08, r*0.05, rgba(0x000000, 0.2))
		}
	case sim.SucrePoudre:
		roundRect(dst, x-r*0.6, y-r*0.7, r*1.2, r*1.6, r*0.2, rgb(0x5aa0e0))
		roundRect(dst, x-r*0.45, y-r*0.3, r*0.9, r*0.7, r*0.1, rgb(0xffffff))
		ellipse(dst, x, y+r*0.05, r*0.3, r*0.15, rgb(0xeeeeee))
	case sim.SorbetFraise, sim.SorbetCerise, sim.SorbetMangue, sim.SorbetAnanas:
		// coupe et boules de sorbet
		c := cropCol[sim.Fraise]
		switch it {
		case sim.SorbetCerise:
			c = cropCol[sim.Cerise]
		case sim.SorbetMangue:
			c = 0xffa53a
		case sim.SorbetAnanas:
			c = 0xffe25a
		}
		polygon(dst, []V{{X: x - r*0.7, Y: y + r*0.1}, {X: x + r*0.7, Y: y + r*0.1}, {X: x + r*0.3, Y: y + r*0.7}, {X: x - r*0.3, Y: y + r*0.7}}, rgb(0xd8eef7))
		rect(dst, x-r*0.08, y+r*0.7, r*0.16, r*0.25, rgb(0xd8eef7))
		circle(dst, x-r*0.3, y-r*0.05, r*0.38, rgb(c))
		circle(dst, x+r*0.3, y-r*0.05, r*0.38, shade(c, 1.15))
		circle(dst, x, y-r*0.45, r*0.36, rgb(c))
	case sim.GlaceFraise, sim.GlaceCerise, sim.GlacePoireChoco, sim.GlaceCoco, sim.GlaceCafe:
		// cornet
		polygon(dst, []V{{X: x - r*0.45, Y: y}, {X: x + r*0.45, Y: y}, {X: x, Y: y + r}}, rgb(0xd9a55a))
		line(dst, x-r*0.25, y+r*0.1, x+r*0.15, y+r*0.6, r*0.05, rgb(0xb07d3a))
		top, bot := uint32(0xffb3c7), uint32(0xff8fb0)
		switch it {
		case sim.GlaceCerise:
			top, bot = 0xe86a8a, 0xc0345a
		case sim.GlacePoireChoco:
			top, bot = 0x6b3e26, 0xe8f0b0
		case sim.GlaceCoco:
			top, bot = 0xfbfaf4, 0xf0ebdc
		case sim.GlaceCafe:
			top, bot = 0xa8774f, 0x8a5c3a
		}
		circle(dst, x, y-r*0.15, r*0.48, rgb(bot))
		circle(dst, x, y-r*0.6, r*0.4, rgb(top))
		if it == sim.GlaceCerise {
			circle(dst, x+r*0.1, y-r*1.0, r*0.15, rgb(0x9e1530))
		}
		if it == sim.GlaceCafe { // un grain de café
			ellipse(dst, x+r*0.1, y-r*0.98, r*0.15, r*0.1, rgb(0x3a2418))
		}
	default:
		return false
	}
	return true
}
