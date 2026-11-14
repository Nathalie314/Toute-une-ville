package main

import (
	"hash/fnv"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// Le cadeau de la fleuriste : pendant 5 min, des fleurs poussent partout dans
// le lieu (pour faire joli), et des pétales volent dans l'air.

type flowerPos struct {
	P     V
	Kind  int     // couleur
	Delay float64 // apparition échelonnée
	Pot   bool    // en pot (à l'intérieur)
}

var flowerCache = map[string][]flowerPos{}

var flowerCols = []uint32{0xff6fa8, 0xffd23f, 0xe8413c, 0xb36ae2, 0xffffff, 0xff9a3c, 0x6aa8ff}

// flowerArea : où les fleurs peuvent pousser dans un lieu (et si c'est en pot).
func flowerArea(p *sim.Place) (sim.Rect, bool) {
	switch p.ID {
	case "super":
		return sim.Rect{X: sim.SuperBX0 + 30, Y: 470, W: sim.SuperBX1 - sim.SuperBX0 - 60, H: 1060}, true
	case "ecole":
		return sim.Rect{X: 110, Y: 140, W: sim.SchoolW - 180, H: 1370}, true
	case "ferme", "exotique", "elevage", "domaine":
		return sim.Rect{X: 20, Y: 140, W: p.W - 40, H: 1470}, false
	}
	return sim.Rect{X: 40, Y: 430, W: p.W - 160, H: 1090}, true // boutiques : dans le magasin
}

// flowersOf : les emplacements des fleurs d'un lieu (calculés une fois, à
// l'écart des bâtiments, des zones, des cases et des routes).
func flowersOf(p *sim.Place) []flowerPos {
	key := p.ID
	if f, ok := flowerCache[key]; ok {
		return f
	}
	area, pot := flowerArea(p)
	var block []sim.Rect
	for _, s := range p.Stations {
		if s.Kind == sim.Field {
			block = append(block, s.Body.Grow(10)) // (pas dans les champs)
		}
		for _, r := range []sim.Rect{s.Body, s.Zone, s.OutZone, s.OutZone2} {
			if r.W > 0 {
				block = append(block, r.Grow(18))
			}
		}
	}
	for _, sp := range p.Spots {
		block = append(block, sp.Zone.Grow(14))
	}
	for _, u := range p.Unlocks {
		block = append(block, u.Zone.Grow(14))
	}
	if p.Yard.W > 0 {
		block = append(block, p.Yard.Grow(20))
	}
	for _, s := range p.Parks {
		block = append(block, sim.Rect{X: s.Pos.X - 70, Y: s.Pos.Y - 70, W: 140, H: 140})
	}
	var potIn sim.Rect // (à la ferme : en pot dans la boutique)
	if p.ID == "ferme" {
		block = append(block, sim.Rect{X: sim.FarmX0 + 70, Y: 460, W: p.W - sim.FarmX0 - 120, H: 130},
			sim.Rect{X: 50, Y: 870, W: p.W - 100, H: 170}, sim.Rect{X: 1170, Y: 460, W: 160, H: 580}) // chemins
		potIn = sim.Rect{X: 110, Y: 1030, W: p.W - 230, H: 480}
	}
	h := fnv.New32a()
	h.Write([]byte(p.ID))
	rng := rand.New(rand.NewSource(int64(h.Sum32())))
	var out []flowerPos
	n := int(area.W * area.H / 9000)
	for try := 0; try < n*3 && len(out) < 600; try++ {
		v := V{X: area.X + rng.Float64()*area.W, Y: area.Y + rng.Float64()*area.H}
		ok := true
		for _, r := range block {
			if r.Has(v) {
				ok = false
				break
			}
		}
		for _, f := range out {
			if math.Abs(f.P.X-v.X) < 70 && math.Abs(f.P.Y-v.Y) < 55 {
				ok = false
				break
			}
		}
		if ok {
			out = append(out, flowerPos{P: v, Kind: rng.Intn(len(flowerCols)), Delay: rng.Float64() * 2.5, Pot: pot || potIn.Has(v)})
		}
	}
	flowerCache[key] = out
	return out
}

// flowerScale : taille des fleurs (elles poussent au début, se fanent à la fin).
func flowerScale(b *sim.Boost, delay float64) float64 {
	el := b.Total - b.Left
	f := math.Min(1, math.Max(0, (el-delay)/0.8))
	f = math.Min(f, math.Max(0, (b.Left-delay*0.4)/2))
	return f
}

// drawFlowers : au sol (les gens marchent devant).
func (g *game) drawFlowers(dst *ebiten.Image, p *sim.Place) {
	b := g.sim.Boost
	if b == nil || b.Kind != "fleurs" || b.Place != p.ID {
		return
	}
	for i, f := range flowersOf(p) {
		s := flowerScale(b, f.Delay)
		if s <= 0 {
			continue
		}
		if b.Mul() > 2 {
			s *= 1.3 // grand embellissement (livraison express)
		}
		drawFlowerClump(dst, f, s, g.t, i)
	}
}

func drawFlowerClump(dst *ebiten.Image, f flowerPos, s, t float64, i int) {
	x, y := f.P.X, f.P.Y
	c := rgb(flowerCols[f.Kind])
	c2 := rgb(flowerCols[(f.Kind+2)%len(flowerCols)])
	if f.Pot {
		ellipse(dst, x, y+4, 20*s, 6*s, rgba(0x000000, 0.18))
		polygon(dst, []V{{x - 16*s, y - 22*s}, {x + 16*s, y - 22*s}, {x + 11*s, y + 2}, {x - 11*s, y + 2}}, rgb(0xc8693a))
		rect(dst, x-18*s, y-26*s, 36*s, 7*s, rgb(0xe07f4c))
		y -= 24 * s
	} else {
		ellipse(dst, x, y+2, 22*s, 7*s, rgba(0x000000, 0.12))
	}
	sway := math.Sin(t*1.8+float64(i)) * 3 * s
	for k, d := range []V{{-12, -26}, {0, -38}, {12, -28}} {
		bx, by := x+d.X*s+sway, y+d.Y*s
		line(dst, x+d.X*0.3*s, y, bx, by, 2.5*s, rgb(0x4f9a3a))
		if k == 1 {
			ellipse(dst, x+6*s, y-12*s, 6*s, 3*s, rgb(0x5fb048)) // une feuille
		}
		col := c
		if k == 1 {
			col = c2
		}
		pr := 6.5 * s
		for a := 0; a < 5; a++ {
			ang := float64(a)*2*math.Pi/5 + t*0.3
			circle(dst, bx+math.Cos(ang)*pr*0.9, by+math.Sin(ang)*pr*0.9, pr*0.75, col)
		}
		circle(dst, bx, by, pr*0.55, rgb(0xffe066))
	}
}

// drawPetals : des pétales qui volent devant tout pendant l'embellissement.
func (g *game) drawPetals(dst *ebiten.Image, p *sim.Place, tl, br V) {
	b := g.sim.Boost
	if b == nil || b.Kind != "fleurs" || b.Place != p.ID {
		return
	}
	a := flowerScale(b, 0)
	if a <= 0 {
		return
	}
	w, h := br.X-tl.X+200, br.Y-tl.Y+200
	n := 70
	if b.Mul() > 2 {
		n = 150 // grand embellissement
	}
	for i := 0; i < n; i++ {
		fi := float64(i)
		x := tl.X - 100 + math.Mod(fi*317.7+g.t*(40+math.Mod(fi*13, 30))+math.Sin(g.t+fi)*30, w)
		y := tl.Y - 100 + math.Mod(fi*211.3+g.t*(70+math.Mod(fi*7, 40)), h)
		r := (7 + math.Mod(fi*5, 6)) * a
		ang := g.t*2 + fi
		ellipse(dst, x, y, r, r*0.5*(0.6+0.4*math.Abs(math.Sin(ang))), rgba(flowerCols[i%len(flowerCols)], 0.85))
	}
}
