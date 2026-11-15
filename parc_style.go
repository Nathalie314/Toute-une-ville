package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// Le style du parc : allées de sable aux bords irréguliers, lampadaires,
// bancs, poubelles et massifs de fleurs le long des allées, haies autour des
// attractions, grande porte d'entrée.

var sandCol, sandRim = uint32(0xf0dcaa), uint32(0xd6bd85)

// sandBand : une allée de sable (rectangle aux bords bosselés).
func sandBand(dst *ebiten.Image, x, y, w, h float64, seed int) {
	bumps := func(c uint32, grow float64) {
		roundRect(dst, x-grow, y-grow, w+2*grow, h+2*grow, 40+grow, rgb(c))
		// des bosses le long des grands côtés
		if w >= h {
			for bx := x + 30; bx < x+w-30; bx += 70 {
				k := hash2(int(bx), seed, 1)
				r := 18 + 22*k + grow
				circle(dst, bx, y+8, r, rgb(c))
				circle(dst, bx+35, y+h-8, 18+22*hash2(int(bx), seed, 2)+grow, rgb(c))
			}
		} else {
			for by := y + 30; by < y+h-30; by += 70 {
				circle(dst, x+8, by, 18+22*hash2(int(by), seed, 3)+grow, rgb(c))
				circle(dst, x+w-8, by+35, 18+22*hash2(int(by), seed, 4)+grow, rgb(c))
			}
		}
	}
	bumps(sandRim, 10)
	bumps(sandCol, 0)
	// quelques cailloux et traces
	for i := 0; i < int(w*h/40000); i++ {
		px := x + 40 + hash2(i, seed, 5)*(w-80)
		py := y + 30 + hash2(i, seed, 6)*(h-60)
		ellipse(dst, px, py, 5, 3, rgba(0xc9ae74, 0.6))
	}
}

// parcLanding : un petit bout d'allée devant l'entrée et la sortie d'une attraction.
func parcLanding(dst *ebiten.Image, r *sim.Ride) {
	e := r.Edge()
	y0, y1 := e, sim.ParcBoulY-sim.ParcBoulH/2+20
	if !r.Top {
		y0, y1 = sim.ParcBoulY+sim.ParcBoulH/2-20, e
	}
	for _, x := range []float64{r.Entrance().X, r.Exit().X} {
		roundRect(dst, x-90, y0-10, 180, y1-y0+20, 30, rgb(sandRim))
		roundRect(dst, x-80, y0-4, 160, y1-y0+8, 26, rgb(sandCol))
	}
}

// parcHedge : la haie qui borde le terrain d'une attraction (sauf côté allée,
// entre l'entrée et la sortie).
func parcHedge(dst *ebiten.Image, r *sim.Ride, openFrom float64) {
	pl := r.Plot
	bush := func(x, y float64, i int) {
		k := 0.8 + 0.4*hash2(i, r.Kind.Slot, 9)
		circle(dst, x+6, y+8, 26*k, rgba(0x1d3a12, 0.25))
		circle(dst, x, y, 26*k, rgb(0x3f8a3a))
		circle(dst, x-7*k, y-7*k, 15*k, rgb(0x5cae4c))
	}
	i := 0
	far := pl.Y
	if !r.Top {
		far = pl.Y + pl.H
	}
	for x := pl.X + 30; x <= pl.X+pl.W-30; x += 44 {
		bush(x, far, i)
		i++
	}
	for y := pl.Y + 30; y <= pl.Y+pl.H-30; y += 44 {
		bush(pl.X, y, i)
		bush(pl.X+pl.W, y, i+1)
		i += 2
	}
	// côté allée : seulement aux deux bouts (l'entrée et la sortie restent ouvertes)
	near := r.Edge()
	for x := pl.X + 30; x <= pl.X+pl.W-30; x += 44 {
		if x > pl.X+20 && x < pl.X+560 || x > r.Exit().X-110 && x < r.Exit().X+110 {
			continue
		}
		if x > openFrom-30 {
			continue // (le terrain continue au-delà du boulevard)
		}
		bush(x, near, i)
		i++
	}
}

type parcProp struct {
	x, y float64
	kind int // 0 lampadaire, 1 banc, 2 poubelle, 3 massif de fleurs
}

// parcProps : le mobilier le long du boulevard, dans les bandes de 100 entre
// les terrains et l'allée (jamais devant une entrée ou une sortie).
var parcProps = func() []parcProp {
	var ps []parcProp
	bt := sim.ParcBoulY - sim.ParcBoulH/2
	bb := sim.ParcBoulY + sim.ParcBoulH/2
	for row := 0; row < 2; row++ {
		y := bt - 52.0
		if row == 1 {
			y = bb + 52
		}
		for i := 0; i < 4; i++ {
			pl := sim.ParcSlot(i + 4*row)
			ent, ex := pl.X+180, pl.X+pl.W-180
			k := 0
			for x := pl.X + 60; x < pl.X+pl.W-40; x += 150 {
				if math.Abs(x-ent) < 140 || math.Abs(x-ex) < 140 {
					continue
				}
				if row == 1 && math.Abs(x-sim.ParcAvX) < sim.ParcAvW/2+60 {
					continue
				}
				nearStand := false // (pas de lampadaire devant un stand)
				for _, k := range sim.StandKinds {
					if (k.Slot < 4) == (row == 0) && math.Abs(sim.ParcSlot(k.Slot).X+k.DX-x) < 150 {
						nearStand = true
					}
				}
				nearStand = nearStand || x > sim.ParcBoulEnd-80 // (le boulevard s'arrête)
				if nearStand {
					continue
				}
				ps = append(ps, parcProp{x, y, []int{0, 1, 3, 1, 0, 2, 3}[k%7]})
				k++
			}
		}
	}
	// le long de l'avenue
	for y := bb + 120; y < sim.ParcPlazaY-260; y += 180 {
		ps = append(ps, parcProp{sim.ParcAvX - sim.ParcAvW/2 - 40, y, 0}, parcProp{sim.ParcAvX + sim.ParcAvW/2 + 40, y, 0})
	}
	return ps
}()

func drawParcProp(dst *ebiten.Image, pr parcProp, t float64) {
	x, y := pr.x, pr.y
	switch pr.kind {
	case 0: // lampadaire
		ellipse(dst, x+22, y+4, 26, 7, rgba(0x1d3a12, 0.22))
		rect(dst, x-11, y-6, 22, 10, rgb(0x2f3b4a))
		rect(dst, x-4, y-96, 8, 92, rgb(0x3b4a5c))
		rect(dst, x-4, y-96, 3, 92, rgb(0x5b6e84))
		circle(dst, x, y-104, 14, rgb(0x2f3b4a))
		circle(dst, x, y-106, 10, rgb(0xfff2b0))
		circle(dst, x-3, y-109, 4, rgb(0xffffff))
		polygon(dst, []V{{X: x - 16, Y: y - 112}, {X: x, Y: y - 126}, {X: x + 16, Y: y - 112}}, rgb(0x2f3b4a))
	case 1: // banc
		ellipse(dst, x+10, y+6, 46, 9, rgba(0x1d3a12, 0.22))
		rect(dst, x-36, y-10, 6, 14, rgb(0x3b3b3b))
		rect(dst, x+30, y-10, 6, 14, rgb(0x3b3b3b))
		roundRect(dst, x-42, y-22, 84, 14, 4, rgb(0xb5733a))
		rect(dst, x-42, y-22, 84, 4, rgb(0xd08a4a))
		roundRect(dst, x-42, y-42, 84, 12, 4, rgb(0x9a5f2e))
	case 2: // poubelle
		ellipse(dst, x+8, y+4, 18, 6, rgba(0x1d3a12, 0.22))
		roundRect(dst, x-13, y-34, 26, 36, 6, rgb(0x2fa86b))
		rect(dst, x-13, y-34, 8, 36, rgb(0x46c283))
		roundRect(dst, x-15, y-40, 30, 8, 4, rgb(0x1f7a4c))
	case 3: // massif de fleurs
		ellipse(dst, x+6, y+6, 50, 18, rgba(0x1d3a12, 0.2))
		ellipse(dst, x, y, 48, 18, rgb(0x8a5a33))
		ellipse(dst, x, y-4, 44, 15, rgb(0x5a3a22))
		cols := []uint32{0xff5d8f, 0xffd23f, 0xffffff, 0xb36bd6, 0xff8a3d}
		for k := 0; k < 11; k++ {
			a := float64(k) * 2.4
			rr := 10 + 22*math.Mod(float64(k)*0.37, 1)
			fx, fy := x+math.Cos(a)*rr, y-6+math.Sin(a)*rr*0.4
			circle(dst, fx, fy-4, 6, rgb(0x3f8a3a))
			circle(dst, fx, fy-8, 5, rgb(cols[(k+int(x))%len(cols)]))
		}
	}
}

// drawParcGate : la grande porte d'entrée (deux tours, une arche, des drapeaux).
func drawParcGate(dst *ebiten.Image, t float64) {
	ax := sim.ParcAvX
	ay := sim.ParcPlazaY - 220
	tower := func(x float64, flag uint32) {
		ellipse(dst, x+40, ay+10, 70, 18, rgba(0x1d3a12, 0.25))
		roundRect(dst, x-38, ay-200, 76, 210, 10, rgb(0xf3ead8))
		rect(dst, x-38, ay-200, 22, 210, rgb(0xfff8ea))
		rect(dst, x+16, ay-200, 22, 210, rgb(0xd9ccb2))
		for k := 0; k < 3; k++ {
			rect(dst, x-38, ay-150+float64(k)*55, 76, 8, rgb(0x6c4bb8))
		}
		roundRect(dst, x-12, ay-180, 24, 34, 12, rgb(0x3b2a52)) // une fenêtre
		// toit pointu en deux pans
		polygon(dst, []V{{X: x - 50, Y: ay - 196}, {X: x, Y: ay - 300}, {X: x, Y: ay - 196}}, rgb(0x8a63d6))
		polygon(dst, []V{{X: x, Y: ay - 300}, {X: x + 50, Y: ay - 196}, {X: x, Y: ay - 196}}, rgb(0x5d3fa0))
		// mât et drapeau qui flotte
		line(dst, x, ay-300, x, ay-360, 4, rgb(0x5a5a5a))
		w := math.Sin(t*4+x) * 6
		polygon(dst, []V{{X: x + 2, Y: ay - 360}, {X: x + 46, Y: ay - 352 + w}, {X: x + 2, Y: ay - 338}}, rgb(flag))
	}
	tower(ax-290, 0xff5d8f)
	tower(ax+290, 0x4fc3ff)
	// l'arche et son enseigne
	roundRect(dst, ax-270, ay-196, 540, 26, 12, rgb(0x5d3fa0))
	roundRect(dst, ax-250, ay-276, 500, 92, 30, rgb(0xffd23f))
	roundRect(dst, ax-242, ay-268, 484, 76, 26, rgb(0xe8484f))
	sayFit(dst, "PARC D'ATTRACTIONS", ax, ay-230, 36, true, 0, 450, rgb(0xffffff))
	// guirlande d'ampoules qui clignotent
	for i := 0; i < 12; i++ {
		c := []uint32{0xff6fa8, 0x4fc3ff, 0x7ed957, 0xffd23f}[(i+int(t*3))%4]
		circle(dst, ax-225+float64(i)*41, ay-282, 7, rgb(c))
	}
}

// parcStyleDrawables : la porte et le mobilier (ils ont une hauteur).
func parcStyleDrawables(dst *ebiten.Image, p *sim.Place, ds []drawable, tl, br V, t float64) []drawable {
	for _, pr := range parcProps {
		pr := pr
		if pr.x < tl.X-100 || pr.x > br.X+100 || pr.y < tl.Y-20 || pr.y > br.Y+160 {
			continue
		}
		ds = append(ds, drawable{pr.y, func() { drawParcProp(dst, pr, t) }})
	}
	ds = append(ds, drawable{sim.ParcPlazaY - 210, func() { drawParcGate(dst, t) }})
	return ds
}
