package main

import (
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// L'huilerie : le sol (un espace par culture), le tournesol, le colza et le lin,
// les oliviers, noisetiers et noyers, les presses et les bouteilles d'huile.

func init() {
	cropCol[sim.Olive] = 0x3a3a24
	cropCol[sim.Noisette] = 0xa8662e
	cropCol[sim.Noix] = 0x8a6a40
}

// huilZoneFloor : la couleur du sol de chaque espace.
var huilZoneFloor = map[string]uint32{"tournesol": 0xf3e3a0, "colza": 0xe9eba0, "olivier": 0xd8d9b0, "lin": 0xd6dcef, "noisetier": 0xe2cfae, "noyer": 0xd9c6a6}

// huilerieGround : chemins, espaces (leur nom), réserves, pistes des drones, route.
func (g *game) huilerieGround(dst *ebiten.Image, p *sim.Place, tl, br V) {
	path := rgb(0xd9c08f)
	for _, z := range sim.TeamZones(p.ID) {
		a := 0.8
		if !p.ZoneOpen(z.ID) {
			a = 0.3
		}
		roundRect(dst, z.X0, -380, z.X1-z.X0, 1600, 30, rgba(huilZoneFloor[z.ID], a)) // (jusque sous les réserves)
		roundRect(dst, (z.X0+z.X1)/2-110, 1000, 220, 34, 10, rgba(0x5a5a22, 0.85))
		say(dst, strings.ToUpper(z.Name), (z.X0+z.X1)/2, 1017, 18, true, 0, rgb(0xfff3c4))
	}
	// le chemin entre les cultures et les presses, et celui devant les réserves
	roundRect(dst, 300, 470, p.W-500, 100, 50, path)
	roundRect(dst, 100, 1060, p.W-200, 30, 14, path)
	// pistes des drones, jusqu'au quai
	roundRect(dst, 380, 1320, p.W-760, 140, 20, rgba(0x8a8f96, 0.35))
	road(dst, tl, br, sim.RoadY)
	gx := p.Gate.X
	rect(dst, gx-70, 1500, 140, 140, path)
	roundRect(dst, gx-140, 1475, 280, 46, 8, rgb(0x6a6a22))
	say(dst, "L'HUILERIE", gx, 1498, 24, true, 0, rgb(0xfff3c4))
}

// drawHuilPlant : le tournesol, le colza et le lin (faux si ce n'en est pas un).
func drawHuilPlant(dst *ebiten.Image, it sim.Item, x, y, gr, t float64) bool {
	ripe := gr >= 1
	sway := 0.0
	if ripe {
		sway = math.Sin(t*2+x*0.01) * 2
	}
	switch it {
	case sim.Tournesol:
		// une grande tige, deux feuilles, la fleur jaune au cœur brun (tournée vers le soleil)
		h := 18 + 52*gr
		line(dst, x, y, x+sway, y-h, 3, rgb(0x4a8f3a))
		ellipse(dst, x-8, y-h*0.45, 9, 4, rgb(0x5aa83a))
		ellipse(dst, x+8, y-h*0.6, 9, 4, rgb(0x5aa83a))
		if gr > 0.35 {
			r := 4 + 10*math.Min(1, (gr-0.35)/0.65)
			for k := 0; k < 10; k++ {
				a := float64(k) * math.Pi / 5
				ellipse(dst, x+sway+math.Cos(a)*r, y-h+math.Sin(a)*r, r*0.55, r*0.55, rgb(0xf5c518))
			}
			circle(dst, x+sway, y-h, r*0.7, rgb(0x6e4322))
			if ripe {
				circle(dst, x+sway-2, y-h-2, r*0.25, rgb(0x8a5a33))
			}
		}
	case sim.Colza:
		// une touffe verte, puis plein de petites fleurs jaune vif
		h := 12 + 34*gr
		for k := -1; k <= 1; k++ {
			line(dst, x, y, x+float64(k)*8+sway, y-h, 2.5, rgb(0x4f9a3a))
		}
		ellipse(dst, x, y-6, 12, 5, rgb(0x5aa83a))
		if gr > 0.4 {
			c := lerpCol(0xc9e07a, 0xf7e01a, (gr-0.4)/0.6)
			for k := 0; k < 7; k++ {
				circle(dst, x-9+float64(k%4)*6+sway, y-h-2+float64(k/4)*6, 3.5, c)
			}
		}
	case sim.Lin:
		// des tiges fines, des petites fleurs bleues
		h := 14 + 36*gr
		for k := -2; k <= 2; k++ {
			line(dst, x+float64(k)*3, y, x+float64(k)*5+sway, y-h, 1.6, rgb(0x6aa84a))
		}
		if gr > 0.4 {
			for k := 0; k < 5; k++ {
				circle(dst, x-10+float64(k)*5+sway, y-h-float64(k%2)*4, 3.5, rgb(0x7aa6e8))
				circle(dst, x-10+float64(k)*5+sway, y-h-float64(k%2)*4, 1.2, rgb(0xf3f3e6))
			}
		}
	default:
		return false
	}
	return true
}

// drawHuilTree : l'olivier, le noisetier et le noyer (faux si ce n'en est pas un).
func drawHuilTree(dst *ebiten.Image, it sim.Item, x, y, size, fr float64, ripe bool, t float64) bool {
	fc := rgb(cropCol[it])
	if !ripe {
		fc = rgb(0x8fcf4a)
	}
	switch it {
	case sim.Olive:
		// un olivier : tronc noueux et tordu, feuillage gris-vert, olives noires
		th := 36 * size
		ellipse(dst, x, y+2, 30*size, 8, rgba(0x000000, 0.2))
		line(dst, x-3*size, y, x+5*size, y-th*0.5, 9*size, rgb(0x7a6a52))
		line(dst, x+5*size, y-th*0.5, x-2*size, y-th, 8*size, rgb(0x7a6a52))
		cr := 26 * size
		for k := 0; k < 5; k++ {
			a := float64(k) * 1.26
			circle(dst, x+math.Cos(a)*cr*0.55, y-th-cr*0.35+math.Sin(a)*cr*0.35, cr*0.55, rgb([]uint32{0x8a9a6a, 0x7a8a5a, 0x9aaa7a}[k%3]))
		}
		if fr > 0 {
			if !ripe {
				fc = rgb(0x8aa04a)
			}
			for k := 0; k < 6; k++ {
				ellipse(dst, x-cr*0.6+float64(k)*cr*0.24, y-th-cr*0.1-float64(k%2)*cr*0.35, fr*0.6, fr*0.8, fc)
			}
		}
	case sim.Noisette:
		// un noisetier : un buisson à plusieurs tiges, feuilles rondes, noisettes en grappes
		th := 26 * size
		ellipse(dst, x, y+2, 32*size, 8, rgba(0x000000, 0.2))
		for k := -1; k <= 1; k++ {
			line(dst, x, y, x+float64(k)*10*size, y-th, 4*size, rgb(0x8a6a4a))
		}
		cr := 26 * size
		circle(dst, x, y-th-cr*0.4, cr, rgb(0x5a9a3a))
		circle(dst, x-cr*0.6, y-th-cr*0.05, cr*0.65, rgb(0x4f8a35))
		circle(dst, x+cr*0.6, y-th-cr*0.1, cr*0.65, rgb(0x66a842))
		if fr > 0 {
			for k := 0; k < 4; k++ {
				fx, fy := x-cr*0.5+float64(k)*cr*0.33, y-th-cr*0.15-float64(k%2)*cr*0.4
				ellipse(dst, fx, fy-fr*0.6, fr*0.8, fr*0.5, rgb(0x7ab84a)) // la collerette
				circle(dst, fx, fy, fr*0.75, fc)
			}
		}
	case sim.Noix:
		// un noyer : grand tronc, large couronne, noix vertes puis brunes
		th := 50 * size
		ellipse(dst, x, y+2, 34*size, 9, rgba(0x000000, 0.2))
		rect(dst, x-5*size-1, y-th, 10*size+2, th, rgb(0x6a5040))
		cr := 30 * size
		circle(dst, x, y-th-cr*0.5, cr, rgb(0x3f7a35))
		circle(dst, x-cr*0.7, y-th-cr*0.15, cr*0.7, rgb(0x356a2e))
		circle(dst, x+cr*0.7, y-th-cr*0.2, cr*0.7, rgb(0x4a8a3c))
		if fr > 0 {
			for k := 0; k < 5; k++ {
				circle(dst, x-cr*0.65+float64(k)*cr*0.32, y-th-cr*0.1-float64(k%2)*cr*0.45, fr*0.9, fc)
			}
		}
	default:
		return false
	}
	return true
}

// drawHuilMachine : une presse (ou le moulin à huile d'olive) sous son auvent, une cuve,
// l'huile qui coule dans la bouteille.
func drawHuilMachine(dst *ebiten.Image, s *sim.Station, t float64) {
	b := s.Body
	ellipse(dst, b.C().X, b.Y+b.H, b.W*0.55, 16, rgba(0x000000, 0.18))
	roof := rgb(0x6a6a22)
	if s.ID == "moulin_olive" {
		roof = rgb(0x8a6a3a)
	}
	// l'abri : murs crème, toit
	roundRect(dst, b.X, b.Y+30, b.W, b.H-30, 8, rgb(0xf0e6cc))
	polygon(dst, []V{{X: b.X - 12, Y: b.Y + 40}, {X: b.C().X, Y: b.Y - 10}, {X: b.X + b.W + 12, Y: b.Y + 40}}, roof)
	work := 0.0
	if s.Working {
		work = 1
	}
	if s.ID == "moulin_olive" {
		// la meule de pierre qui tourne dans son bac
		cx, cy := b.X+80, b.Y+b.H-50
		ellipse(dst, cx, cy+10, 56, 22, rgb(0x8a8a82))
		ellipse(dst, cx, cy+4, 48, 16, rgb(0x3a3a24))
		a := t * 2 * work
		ellipse(dst, cx+math.Cos(a)*18, cy-6+math.Sin(a)*5, 22, 26, rgb(0xb8b4a8))
		ellipse(dst, cx+math.Cos(a)*18, cy-6+math.Sin(a)*5, 14, 18, rgb(0xa09c90))
	} else {
		// la presse : une vis et un plateau au-dessus d'un bac de graines
		cx := b.X + 80
		roundRect(dst, cx-50, b.Y+b.H-50, 100, 36, 6, rgb(0x8a5a33))
		item := s.Inputs[0]
		for k := 0; k < 5; k++ {
			drawItem(dst, item, cx-36+float64(k)*18, b.Y+b.H-40, 8)
		}
		py := b.Y + 60 + math.Abs(math.Sin(t*1.6))*14*work
		rect(dst, cx-4, b.Y+40, 8, py-b.Y-40, rgb(0x5a5a5a))
		roundRect(dst, cx-44, py, 88, 12, 4, rgb(0x6e6e6e))
	}
	// la cuve d'huile et sa bouteille
	ox := b.X + b.W - 70
	roundRect(dst, ox-30, b.Y+60, 60, 70, 10, rgb(0xb8c0c6))
	rect(dst, ox-2, b.Y+130, 4, 12, rgb(0x6e6e6e))
	if s.Working {
		ph := math.Mod(t*2, 1)
		circle(dst, ox, b.Y+142+ph*10, 2.5, rgb(oilCol(s.Item)))
	}
	drawItem(dst, s.Item, ox, b.Y+b.H-24, 18)
}

// oilCol : la couleur d'une huile.
func oilCol(it sim.Item) uint32 {
	switch it {
	case sim.HuileTournesol:
		return 0xf5d23a
	case sim.HuileColza:
		return 0xe8c62a
	case sim.HuileOlive:
		return 0x9aa22a
	case sim.HuileLin:
		return 0xd9a02a
	case sim.HuileNoisette:
		return 0xc98a2a
	case sim.HuileNoix:
		return 0xb5782a
	}
	return 0xe8c62a
}

// drawItemHuilerie : les icônes de l'huilerie (faux si ce n'en est pas une).
func drawItemHuilerie(dst *ebiten.Image, it sim.Item, x, y, r float64) bool {
	switch it {
	case sim.Tournesol:
		for k := 0; k < 5; k++ { // des graines rayées noir et blanc
			cx, cy := x+float64(k%3-1)*r*0.42, y+float64(k/3)*r*0.4-r*0.15
			ellipse(dst, cx, cy, r*0.18, r*0.32, rgb(0x2b2b2b))
			line(dst, cx, cy-r*0.25, cx, cy+r*0.25, r*0.05, rgb(0xe8e8e0))
		}
	case sim.Colza:
		ellipse(dst, x, y+r*0.3, r*0.7, r*0.3, rgb(0x8a6a3a)) // un petit tas de graines noires
		for k := 0; k < 9; k++ {
			circle(dst, x-r*0.45+float64(k%5)*r*0.22, y+r*0.15-float64(k/5)*r*0.2, r*0.11, rgb(0x2a2018))
		}
		circle(dst, x+r*0.5, y-r*0.5, r*0.2, rgb(0xf7e01a))
	case sim.Lin:
		for k := 0; k < 6; k++ { // des graines plates et brillantes
			cx, cy := x+float64(k%3-1)*r*0.4, y+float64(k/3)*r*0.38-r*0.15
			ellipse(dst, cx, cy, r*0.2, r*0.13, rgb(0x9a5a2a))
			ellipse(dst, cx-r*0.05, cy-r*0.04, r*0.07, r*0.04, rgba(0xffffff, 0.5))
		}
		circle(dst, x+r*0.55, y-r*0.55, r*0.18, rgb(0x7aa6e8))
	case sim.HuileTournesol, sim.HuileColza, sim.HuileOlive, sim.HuileLin, sim.HuileNoisette, sim.HuileNoix:
		// une bouteille d'huile : verre clair, l'huile dorée, une étiquette avec sa matière
		c := oilCol(it)
		roundRect(dst, x-r*0.34, y-r*0.35, r*0.68, r*1.25, r*0.16, rgba(0xf4f4ea, 0.9))
		roundRect(dst, x-r*0.3, y-r*0.15, r*0.6, r*1.02, r*0.13, rgb(c))
		rect(dst, x-r*0.1, y-r*0.82, r*0.2, r*0.5, rgba(0xf4f4ea, 0.9))
		rect(dst, x-r*0.13, y-r*0.95, r*0.26, r*0.16, rgb(0x2b5a2a))
		roundRect(dst, x-r*0.3, y+r*0.12, r*0.6, r*0.42, r*0.05, rgb(0xfff8e6))
		raw := map[sim.Item]sim.Item{sim.HuileTournesol: sim.Tournesol, sim.HuileColza: sim.Colza, sim.HuileOlive: sim.Olive,
			sim.HuileLin: sim.Lin, sim.HuileNoisette: sim.Noisette, sim.HuileNoix: sim.Noix}[it]
		drawItem(dst, raw, x, y+r*0.33, r*0.2)
		line(dst, x-r*0.2, y-r*0.2, x-r*0.2, y+r*0.75, r*0.06, rgba(0xffffff, 0.4))
	default:
		return false
	}
	return true
}
