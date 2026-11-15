package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Les fontaines décoratives (petits bateaux, et çà et là dans la ville). Vues du
// dessus avec un peu de relief : la pierre claire, l'eau bleue, des jets qui bougent.
// kind : 0 à étages, 1 à la boule, 2 à colonnes (la grande), 3 bassin carré et ses
// jets, 4 bassin octogonal, 5 jets qui sortent des dalles, 6 mur d'eau.
const nFountainKinds = 7

var (
	stoneL  = uint32(0xe9eaee)
	stoneM  = uint32(0xc9ccd4)
	stoneD  = uint32(0x9fa4b0)
	waterD  = uint32(0x5c9fe0)
	waterL  = uint32(0x9fd0ff)
	foamCol = uint32(0xeaf6ff)
)

// jet : un jet d'eau qui monte de (x, y) et retombe en gerbe.
func jet(dst *ebiten.Image, x, y, h, spread, t, ph float64) {
	hh := h * (0.9 + 0.1*math.Sin(t*6+ph))
	line(dst, x, y, x, y-hh, 6, rgba(waterL, 0.9))
	line(dst, x, y, x, y-hh, 2.5, rgba(0xffffff, 0.9))
	for _, sd := range []float64{-1, 1} {
		for j := 1; j <= 5; j++ {
			u := float64(j) / 5
			px := x + sd*spread*u
			py := y - hh + hh*0.55*u*u
			circle(dst, px, py, 4.5-2*u, rgba(foamCol, 0.9))
		}
	}
	circle(dst, x, y-hh, 6, rgba(0xffffff, 0.95))
}

// pool : un bassin rond (bord de pierre, eau, reflets, vaguelettes).
func pool(dst *ebiten.Image, x, y, rx, ry, t float64) {
	ellipse(dst, x+14, y+16, rx+10, ry+8, rgba(0x1d3a12, 0.22))
	ellipse(dst, x, y+10, rx, ry, rgb(stoneD))
	ellipse(dst, x, y, rx, ry, rgb(stoneL))
	ellipse(dst, x, y+2, rx-16, ry-12, rgb(waterD))
	ellipse(dst, x-rx*0.15, y-ry*0.1, rx*0.55, ry*0.4, rgba(waterL, 0.55))
	for k := 0; k < 3; k++ {
		f := math.Mod(t*0.5+float64(k)/3, 1)
		ellipse(dst, x, y+2, (rx-20)*f, (ry-14)*f, rgba(0xffffff, 0.25*(1-f)))
	}
}

func drawFancyFountain(dst *ebiten.Image, kind int, q V, k, t float64) {
	scaled(q.X, q.Y, k, func() { fancyFountain(dst, kind%nFountainKinds, q.X, q.Y, t) })
}

func fancyFountain(dst *ebiten.Image, kind int, x, y, t float64) {
	switch kind {
	case 0: // à étages : trois vasques l'une sur l'autre, l'eau déborde de chacune
		pool(dst, x, y, 130, 58, t)
		for i, r := range []float64{70, 46, 26} {
			cy := y - 30 - float64(i)*44
			rect(dst, x-9, cy, 18, 44, rgb(stoneM))
			ellipse(dst, x, cy+8, r, r*0.36, rgb(stoneD))
			ellipse(dst, x, cy, r, r*0.36, rgb(stoneL))
			ellipse(dst, x, cy, r-8, r*0.36-5, rgb(waterL))
			for j := 0; j < 8; j++ { // le rideau d'eau qui retombe
				a := float64(j) / 8 * 2 * math.Pi
				if math.Sin(a) < -0.2 {
					continue
				}
				ex, ey := x+math.Cos(a)*r, cy+math.Sin(a)*r*0.36
				line(dst, ex, ey, ex*1.0+math.Cos(a)*6, ey+30+6*math.Sin(t*5+float64(j)), 3, rgba(foamCol, 0.85))
			}
		}
		jet(dst, x, y-126, 50, 20, t, 0)
	case 1: // à la boule : une grande sphère, l'eau coule tout autour
		pool(dst, x, y, 140, 62, t)
		rect(dst, x-26, y-40, 52, 40, rgb(stoneM))
		ellipse(dst, x, y-40, 40, 14, rgb(stoneL))
		circle(dst, x, y-92, 56, rgb(0xdfe3ea))
		circle(dst, x-14, y-104, 34, rgb(0xf4f6fa))
		circle(dst, x-24, y-116, 12, rgb(0xffffff))
		for j := 0; j < 10; j++ {
			a := float64(j)/10*math.Pi + 0.1
			ex := x + math.Cos(a)*56
			line(dst, ex, y-92+math.Sin(a)*10, ex+math.Cos(a)*24, y-30, 4, rgba(waterL, 0.8))
		}
	case 2: // la grande fontaine à colonnes, au milieu d'un large bassin
		pool(dst, x, y, 190, 84, t)
		ellipse(dst, x, y-10, 92, 40, rgb(stoneD))
		for j := 0; j < 10; j++ {
			a := float64(j) / 10 * 2 * math.Pi
			if math.Sin(a) < 0 {
				continue
			}
			cx := x + math.Cos(a)*80
			cy := y - 10 + math.Sin(a)*34
			rect(dst, cx-8, cy-70, 16, 70, rgb(stoneL))
			rect(dst, cx-8, cy-70, 5, 70, rgb(0xffffff))
		}
		ellipse(dst, x, y-84, 96, 40, rgb(stoneM))
		ellipse(dst, x, y-90, 92, 36, rgb(stoneL))
		ellipse(dst, x, y-90, 76, 28, rgb(waterL))
		for j := 0; j < 7; j++ {
			jet(dst, x-60+float64(j)*20, y-92+math.Abs(float64(j)-3)*4, 50+16*math.Sin(float64(j)), 16, t, float64(j))
		}
		for j := 0; j < 12; j++ { // les gerbes tout autour, qui retombent dans le grand bassin
			a := float64(j) / 12 * 2 * math.Pi
			if math.Sin(a) < -0.1 {
				continue
			}
			bx, by := x+math.Cos(a)*130, y+math.Sin(a)*56
			jet(dst, bx, by, 40, 12, t, float64(j)*0.7)
		}
	case 3: // un bassin carré, des jets en rangées
		rect(dst, x-130+14, y-70+18, 260, 140, rgba(0x1d3a12, 0.22))
		rect(dst, x-130, y-60, 260, 140, rgb(stoneD))
		rect(dst, x-130, y-70, 260, 140, rgb(stoneL))
		rect(dst, x-114, y-56, 228, 112, rgb(waterD))
		rect(dst, x-114, y-56, 228, 40, rgba(waterL, 0.4))
		for i := 0; i < 3; i++ {
			for j := 0; j < 2; j++ {
				jet(dst, x-70+float64(i)*70, y-20+float64(j)*50, 60, 16, t, float64(i*2+j))
			}
		}
	case 4: // un bassin octogonal, une petite fontaine au milieu
		var oct, in []V
		for j := 0; j < 8; j++ {
			a := (float64(j) + 0.5) / 8 * 2 * math.Pi
			oct = append(oct, V{X: x + math.Cos(a)*140, Y: y + math.Sin(a)*64})
			in = append(in, V{X: x + math.Cos(a)*118, Y: y + 2 + math.Sin(a)*50})
		}
		var sh []V
		for _, o := range oct {
			sh = append(sh, o.Add(V{X: 14, Y: 18}))
		}
		polygon(dst, sh, rgba(0x1d3a12, 0.22))
		var low []V
		for _, o := range oct {
			low = append(low, o.Add(V{Y: 10}))
		}
		polygon(dst, low, rgb(stoneD))
		polygon(dst, oct, rgb(stoneL))
		polygon(dst, in, rgb(waterD))
		ellipse(dst, x-20, y-6, 60, 20, rgba(waterL, 0.5))
		rect(dst, x-10, y-50, 20, 50, rgb(stoneM))
		ellipse(dst, x, y-50, 40, 14, rgb(stoneL))
		ellipse(dst, x, y-52, 32, 10, rgb(waterL))
		jet(dst, x, y-54, 60, 26, t, 0)
		for j := 0; j < 4; j++ {
			a := float64(j)/4*2*math.Pi + 0.4
			jet(dst, x+math.Cos(a)*80, y+math.Sin(a)*30, 34, 10, t, float64(j))
		}
	case 5: // une place dallée : des jets sortent directement du sol
		rect(dst, x-160, y-80, 320, 160, rgb(0xd4d6dc))
		for i := 0; i < 8; i++ {
			for j := 0; j < 4; j++ {
				if (i+j)%2 == 0 {
					rect(dst, x-160+float64(i)*40, y-80+float64(j)*40, 40, 40, rgb(0xc6c9d0))
				}
			}
		}
		for i := 0; i < 4; i++ {
			for j := 0; j < 2; j++ {
				px, py := x-120+float64(i)*80, y-40+float64(j)*80
				on := math.Sin(t*1.3+float64(i+j*2)) > -0.3 // ils s'allument tour à tour
				ellipse(dst, px, py, 12, 5, rgb(stoneD))
				if on {
					ellipse(dst, px, py+2, 26, 10, rgba(waterL, 0.5))
					jet(dst, px, py, 70+20*math.Sin(t*2+float64(i)), 14, t, float64(i+j))
				}
			}
		}
	case 6: // un mur d'eau : la nappe tombe du haut d'un muret dans un bassin
		rect(dst, x-120+14, y-20+18, 240, 80, rgba(0x1d3a12, 0.22))
		rect(dst, x-120, y-10, 240, 70, rgb(stoneD))
		rect(dst, x-120, y-20, 240, 70, rgb(stoneL))
		rect(dst, x-104, y-8, 208, 46, rgb(waterD))
		rect(dst, x-90, y-130, 180, 120, rgb(stoneM))
		rect(dst, x-90, y-130, 40, 120, rgb(stoneL))
		rect(dst, x-96, y-140, 192, 16, rgb(stoneL))
		rect(dst, x-80, y-124, 160, 116, rgba(waterL, 0.85))
		for j := 0; j < 8; j++ {
			px := x - 72 + float64(j)*20
			f := math.Mod(t*1.6+float64(j)*0.3, 1)
			line(dst, px, y-124+f*90, px, y-104+f*90, 3, rgba(0xffffff, 0.9))
		}
		for j := 0; j < 6; j++ {
			circle(dst, x-70+float64(j)*28, y-6+3*math.Sin(t*6+float64(j)), 8, rgba(foamCol, 0.85))
		}
	}
}
