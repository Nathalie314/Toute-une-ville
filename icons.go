package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// drawItem dessine une marchandise centrée en (x, y), de « rayon » r.
func drawItem(dst *ebiten.Image, it sim.Item, x, y, r float64) {
	if drawItemNew(dst, it, x, y, r) || drawItemMeat(dst, it, x, y, r) || drawItemExotic(dst, it, x, y, r) || drawItemPort(dst, it, x, y, r) || drawItemDomaine(dst, it, x, y, r) || drawItemPatisserie(dst, it, x, y, r) || drawItemCrepe(dst, it, x, y, r) {
		return
	}
	dark := rgba(0x000000, 0.25)
	switch it {
	case sim.Ble:
		// gerbe de blé
		for i := -2; i <= 2; i++ {
			a := float64(i) * 0.18
			x2, y2 := x+math.Sin(a)*r*0.95, y-math.Cos(a)*r*0.95
			line(dst, x, y+r*0.8, x2, y2, r*0.12, rgb(0xb8932f))
			ellipse(dst, x2, y2, r*0.16, r*0.32, rgb(0xf0c94a))
		}
		rect(dst, x-r*0.35, y+r*0.15, r*0.7, r*0.18, rgb(0xa0522d))
	case sim.Mais:
		ellipse(dst, x+r*0.08, y+r*0.05, r*0.42, r*0.92, rgb(0x5aa83a))
		ellipse(dst, x, y-r*0.05, r*0.32, r*0.78, rgb(0xf6cf3b))
		for k := -2; k <= 2; k++ {
			line(dst, x-r*0.25, y+float64(k)*r*0.25, x+r*0.25, y+float64(k)*r*0.25, r*0.05, rgb(0xd9a91f))
		}
		polygon(dst, []V{{X: x - r*0.45, Y: y + r}, {X: x - r*0.1, Y: y - r*0.1}, {X: x + r*0.05, Y: y + r}}, rgb(0x3f8f2a))
	case sim.Farine:
		// sac blanc, bandeau brun, épi de blé
		roundRect(dst, x-r*0.65, y-r*0.6, r*1.3, r*1.5, r*0.35, rgb(0xfbf7ee))
		roundRect(dst, x-r*0.45, y-r*0.95, r*0.9, r*0.5, r*0.2, rgb(0xe8e0cc))
		line(dst, x-r*0.4, y-r*0.55, x+r*0.4, y-r*0.55, r*0.1, rgb(0x8a5a33))
		line(dst, x, y+r*0.75, x, y-r*0.25, r*0.08, rgb(0xb8932f))
		for k := 0; k < 3; k++ {
			yy := y + r*0.45 - float64(k)*r*0.25
			ellipse(dst, x-r*0.13, yy, r*0.11, r*0.16, rgb(0xd9a521))
			ellipse(dst, x+r*0.13, yy, r*0.11, r*0.16, rgb(0xd9a521))
		}
		ellipse(dst, x, y-r*0.3, r*0.1, r*0.16, rgb(0xd9a521))
	case sim.Pain, sim.BaguetteMais:
		// baguette en biais
		c1, c2 := uint32(0xc98a3d), uint32(0xe3a857)
		if it == sim.BaguetteMais {
			c1, c2 = 0xd9a521, 0xf6cf3b
		}
		polygon(dst, []V{{X: x - r*0.95, Y: y + r*0.55}, {X: x - r*0.7, Y: y + r*0.8}, {X: x + r*0.95, Y: y - r*0.55}, {X: x + r*0.7, Y: y - r*0.8}}, rgb(c1))
		polygon(dst, []V{{X: x - r*0.85, Y: y + r*0.5}, {X: x - r*0.7, Y: y + r*0.65}, {X: x + r*0.85, Y: y - r*0.5}, {X: x + r*0.7, Y: y - r*0.65}}, rgb(c2))
		for k := -1; k <= 1; k++ {
			cx, cy := x+float64(k)*r*0.45, y-float64(k)*r*0.33
			line(dst, cx-r*0.1, cy-r*0.12, cx+r*0.12, cy+r*0.08, r*0.08, rgb(0xfbe3b0))
		}
	case sim.PainMie:
		// un pain de mie carré, deux tranches devant
		roundRect(dst, x-r*0.85, y-r*0.55, r*1.3, r*1.1, r*0.2, rgb(0xc98a3d))
		roundRect(dst, x-r*0.75, y-r*0.45, r*1.1, r*0.25, r*0.1, rgb(0xe3a857))
		for k := 0; k < 2; k++ {
			sx := x + r*0.35 + float64(k)*r*0.25
			roundRect(dst, sx, y-r*0.35, r*0.5, r*0.85, r*0.12, rgb(0xb5793a))
			roundRect(dst, sx+r*0.05, y-r*0.28, r*0.4, r*0.72, r*0.1, rgb(0xfbe9c8))
		}
	case sim.PainMele:
		ellipse(dst, x, y+r*0.15, r*0.95, r*0.6, rgb(0xb07a3a))
		ellipse(dst, x, y+r*0.05, r*0.85, r*0.5, rgb(0xd9a55a))
		line(dst, x-r*0.5, y-r*0.1, x+r*0.5, y-r*0.1, r*0.1, rgb(0xf6cf3b))
		line(dst, x-r*0.4, y+r*0.2, x+r*0.4, y+r*0.2, r*0.1, rgb(0xf6d49a))
	case sim.Brioche:
		circle(dst, x-r*0.4, y+r*0.2, r*0.45, rgb(0xd9893a))
		circle(dst, x+r*0.4, y+r*0.2, r*0.45, rgb(0xd9893a))
		circle(dst, x, y-r*0.3, r*0.45, rgb(0xe8a04a))
		circle(dst, x-r*0.1, y-r*0.45, r*0.12, rgba(0xffffff, 0.6))
	case sim.PainEpices:
		roundRect(dst, x-r*0.85, y-r*0.45, r*1.7, r*0.95, r*0.2, rgb(0x9a5a2a))
		roundRect(dst, x-r*0.85, y-r*0.45, r*1.7, r*0.3, r*0.15, rgb(0xb8732f))
		circle(dst, x-r*0.4, y+r*0.15, r*0.1, rgb(0xf2a516))
		circle(dst, x+r*0.3, y+r*0.2, r*0.1, rgb(0xf2a516))
	case sim.Tarte:
		ellipse(dst, x, y+r*0.2, r*0.98, r*0.5, rgb(0xd9a55a))
		ellipse(dst, x, y+r*0.05, r*0.8, r*0.38, rgb(0xe23b2e))
		for k := 0; k < 5; k++ {
			a := float64(k) * 1.25
			circle(dst, x+math.Cos(a)*r*0.45, y+math.Sin(a)*r*0.2, r*0.13, rgb(0xff6f6f))
		}
	case sim.GateauFruits:
		rect(dst, x-r*0.75, y-r*0.1, r*1.5, r*0.65, rgb(0xf3d9a6))
		ellipse(dst, x, y-r*0.1, r*0.75, r*0.22, rgb(0xfff3dc))
		circle(dst, x-r*0.35, y-r*0.25, r*0.16, rgb(0xe23b2e))
		circle(dst, x+r*0.05, y-r*0.3, r*0.16, rgb(0x7ed957))
		circle(dst, x+r*0.4, y-r*0.22, r*0.16, rgb(0xff9f1a))
	case sim.Crayons:
		for k, c := range []uint32{0xe74c3c, 0x3498db, 0x2ecc71, 0xf1c40f} {
			cx := x - r*0.45 + float64(k)*r*0.3
			rect(dst, cx-r*0.11, y-r*0.6, r*0.22, r*1.2, rgb(c))
			polygon(dst, []V{{X: cx - r*0.11, Y: y - r*0.6}, {X: cx + r*0.11, Y: y - r*0.6}, {X: cx, Y: y - r*0.9}}, rgb(0xf6d2b4))
		}
	case sim.Cahiers:
		roundRect(dst, x-r*0.6, y-r*0.75, r*1.2, r*1.5, r*0.1, rgb(0x3b6ea8))
		roundRect(dst, x-r*0.5, y-r*0.65, r*1.1, r*1.4, r*0.1, rgb(0xe67e22))
		rect(dst, x-r*0.2, y-r*0.3, r*0.6, r*0.25, rgb(0xffffff))
	case sim.Geometrie:
		polygon(dst, []V{{X: x - r*0.8, Y: y + r*0.6}, {X: x + r*0.6, Y: y + r*0.6}, {X: x - r*0.8, Y: y - r*0.8}}, rgba(0x4fc3ff, 0.85))
		line(dst, x-r*0.9, y+r*0.85, x+r*0.9, y-r*0.2, r*0.16, rgb(0xf1c40f))
		line(dst, x+r*0.3, y-r*0.9, x+r*0.1, y+r*0.6, r*0.08, rgb(0x777777))
		line(dst, x+r*0.3, y-r*0.9, x+r*0.6, y+r*0.6, r*0.08, rgb(0x777777))
	case sim.Atlas:
		roundRect(dst, x-r*0.7, y-r*0.75, r*1.4, r*1.5, r*0.1, rgb(0x2e8b57))
		circle(dst, x, y, r*0.45, rgb(0x5aa0e0))
		ellipse(dst, x-r*0.1, y-r*0.1, r*0.2, r*0.15, rgb(0x7ed957))
		ellipse(dst, x+r*0.2, y+r*0.15, r*0.15, r*0.12, rgb(0x7ed957))
	case sim.Calculatrice:
		roundRect(dst, x-r*0.55, y-r*0.85, r*1.1, r*1.7, r*0.15, rgb(0x444a52))
		rect(dst, x-r*0.4, y-r*0.7, r*0.8, r*0.35, rgb(0x9fe0a0))
		for k := 0; k < 9; k++ {
			rect(dst, x-r*0.4+float64(k%3)*r*0.3, y-r*0.2+float64(k/3)*r*0.3, r*0.2, r*0.2, rgb(0xdddddd))
		}
	case sim.Microscope:
		rect(dst, x-r*0.6, y+r*0.6, r*1.2, r*0.2, rgb(0x555555))
		line(dst, x-r*0.2, y+r*0.6, x-r*0.2, y-r*0.2, r*0.16, rgb(0x777777))
		line(dst, x-r*0.2, y-r*0.2, x+r*0.25, y-r*0.75, r*0.22, rgb(0x3b6ea8))
		rect(dst, x-r*0.45, y+r*0.1, r*0.7, r*0.1, rgb(0x999999))
	case sim.Ordinateur:
		roundRect(dst, x-r*0.7, y-r*0.65, r*1.4, r*0.95, r*0.08, rgb(0x2b2b2b))
		rect(dst, x-r*0.6, y-r*0.55, r*1.2, r*0.75, rgb(0x5aa0e0))
		polygon(dst, []V{{X: x - r*0.85, Y: y + r*0.35}, {X: x + r*0.85, Y: y + r*0.35}, {X: x + r, Y: y + r*0.6}, {X: x - r, Y: y + r*0.6}}, rgb(0x999999))
	case sim.Blouse:
		polygon(dst, []V{{X: x - r*0.4, Y: y - r*0.8}, {X: x + r*0.4, Y: y - r*0.8}, {X: x + r*0.7, Y: y + r*0.8}, {X: x - r*0.7, Y: y + r*0.8}}, rgb(0xffffff))
		polygon(dst, []V{{X: x - r*0.4, Y: y - r*0.8}, {X: x - r*0.85, Y: y - r*0.4}, {X: x - r*0.7, Y: y + r*0.1}}, rgb(0xeeeeee))
		polygon(dst, []V{{X: x + r*0.4, Y: y - r*0.8}, {X: x + r*0.85, Y: y - r*0.4}, {X: x + r*0.7, Y: y + r*0.1}}, rgb(0xeeeeee))
		line(dst, x, y-r*0.7, x, y+r*0.8, r*0.05, rgba(0x000000, 0.25))
	case sim.GlaceVanille, sim.GlaceFruits, sim.GlaceChoco, sim.GlaceFruitLait:
		// cornet
		polygon(dst, []V{{X: x - r*0.45, Y: y - r*0.05}, {X: x + r*0.45, Y: y - r*0.05}, {X: x, Y: y + r}}, rgb(0xd9a55a))
		line(dst, x-r*0.25, y+r*0.15, x+r*0.15, y+r*0.5, r*0.06, rgb(0xb07a3a))
		c := uint32(0xf3e0b5) // nature : crème
		if it == sim.GlaceFruits {
			c = 0xff5a6e // sorbet : bien rouge
		}
		if it == sim.GlaceFruitLait {
			c = 0xffc2d4 // lait + fruits : rose pâle
		}
		if it == sim.GlaceChoco {
			c = 0x7a4a2d
		}
		circle(dst, x, y-r*0.35, r*0.5, rgb(c))
		circle(dst, x-r*0.3, y-r*0.15, r*0.25, rgb(c))
		circle(dst, x+r*0.3, y-r*0.15, r*0.25, rgb(c))
		if it == sim.GlaceVanille {
			// une gousse de vanille posée sur la boule
			line(dst, x-r*0.3, y-r*0.7, x+r*0.35, y-r*0.85, r*0.09, rgb(0x3a2418))
		} else {
			circle(dst, x+r*0.1, y-r*0.75, r*0.12, rgb(0xd62839))
		}
		if it == sim.GlaceFruitLait {
			circle(dst, x-r*0.2, y-r*0.4, r*0.09, rgb(0xd62839)) // des morceaux de fruits
			circle(dst, x+r*0.25, y-r*0.25, r*0.09, rgb(0xf08a24))
		}
	case sim.Oeuf:
		ellipse(dst, x-r*0.35, y+r*0.15, r*0.42, r*0.55, rgb(0xf4e3cc))
		ellipse(dst, x+r*0.35, y+r*0.15, r*0.42, r*0.55, rgb(0xfdf6ea))
		ellipse(dst, x, y-r*0.3, r*0.42, r*0.55, rgb(0xf0d8b8))
	case sim.Legume:
		// carotte + tomate
		polygon(dst, []V{{X: x - r*0.75, Y: y - r*0.35}, {X: x - r*0.2, Y: y - r*0.55}, {X: x + r*0.15, Y: y + r*0.95}}, rgb(0xf08a24))
		line(dst, x-r*0.6, y-r*0.5, x-r*0.75, y-r*0.95, r*0.13, rgb(0x3f9b35))
		line(dst, x-r*0.45, y-r*0.55, x-r*0.35, y-r*0.98, r*0.13, rgb(0x3f9b35))
		circle(dst, x+r*0.42, y+r*0.2, r*0.5, rgb(0xe23b2e))
		circle(dst, x+r*0.42, y-r*0.28, r*0.13, rgb(0x3f9b35))
	case sim.Lait:
		roundRect(dst, x-r*0.45, y-r*0.5, r*0.9, r*1.45, r*0.15, rgb(0xffffff))
		polygon(dst, []V{{X: x - r*0.45, Y: y - r*0.5}, {X: x, Y: y - r*0.95}, {X: x + r*0.45, Y: y - r*0.5}}, rgb(0x5aa0e0))
		rect(dst, x-r*0.45, y, r*0.9, r*0.3, rgb(0x5aa0e0))
		roundStroke(dst, x-r*0.45, y-r*0.5, r*0.9, r*1.45, r*0.15, r*0.06, dark)
	case sim.LaitChevre:
		// brique de lait verte, avec une petite tête de chèvre
		roundRect(dst, x-r*0.5, y-r*0.45, r*1.0, r*1.35, r*0.08, rgb(0x5aa84f))
		polygon(dst, []V{{X: x - r*0.5, Y: y - r*0.45}, {X: x, Y: y - r*0.9}, {X: x + r*0.5, Y: y - r*0.45}}, rgb(0xeaf6e4))
		rect(dst, x-r*0.1, y-r*1.0, r*0.2, r*0.2, rgb(0xeaf6e4))
		ellipse(dst, x, y+r*0.25, r*0.32, r*0.26, rgb(0xffffff))
		line(dst, x-r*0.18, y+r*0.02, x-r*0.28, y-r*0.16, r*0.08, rgb(0x8a7a6a)) // cornes
		line(dst, x+r*0.18, y+r*0.02, x+r*0.28, y-r*0.16, r*0.08, rgb(0x8a7a6a))
		roundStroke(dst, x-r*0.5, y-r*0.45, r*1.0, r*1.35, r*0.08, r*0.06, dark)
	case sim.LaitBrebis:
		// pichet en terre cuite, avec sa anse
		circle(dst, x+r*0.62, y+r*0.15, r*0.28, rgb(0xb5653a))
		circle(dst, x+r*0.62, y+r*0.15, r*0.15, rgba(0xffffff, 0))
		polygon(dst, []V{{X: x - r*0.45, Y: y - r*0.55}, {X: x + r*0.45, Y: y - r*0.55}, {X: x + r*0.6, Y: y + r*0.85}, {X: x - r*0.6, Y: y + r*0.85}}, rgb(0xc8763f))
		polygon(dst, []V{{X: x - r*0.45, Y: y - r*0.55}, {X: x - r*0.7, Y: y - r*0.75}, {X: x - r*0.3, Y: y - r*0.62}}, rgb(0xc8763f)) // bec
		ellipse(dst, x, y-r*0.55, r*0.45, r*0.12, rgb(0xfdfaf0))
		rect(dst, x-r*0.55, y+r*0.2, r*1.1, r*0.14, rgb(0xf3e2c0))
	case sim.CuisseCanard:
		// deux cuisses dorées
		for k := 0; k < 2; k++ {
			cx := x - r*0.3 + float64(k)*r*0.6
			line(dst, cx, y+r*0.2, cx+r*0.15, y+r*0.9, r*0.18, rgb(0xf3ead8))
			circle(dst, cx+r*0.2, y+r*0.95, r*0.12, rgb(0xf3ead8))
			ellipse(dst, cx, y-r*0.2, r*0.36, r*0.55, rgb(0xb5651d))
			ellipse(dst, cx-r*0.08, y-r*0.32, r*0.14, r*0.24, rgba(0xffffff, 0.3))
		}
	case sim.CafeVert, sim.CafeGrains:
		c, l := uint32(0x8fbf5a), uint32(0x5a8a32)
		if it == sim.CafeGrains {
			c, l = 0x6a3a1e, 0x3a1e0e
		}
		for k := 0; k < 3; k++ {
			cx := x + float64(k-1)*r*0.48
			cy := y + r*0.2 - float64(k%2)*r*0.45
			ellipse(dst, cx, cy, r*0.3, r*0.4, rgb(c))
			line(dst, cx-r*0.04, cy-r*0.32, cx+r*0.04, cy+r*0.32, r*0.07, rgb(l))
		}
	case sim.CafeMoulu:
		// paquet de café
		roundRect(dst, x-r*0.55, y-r*0.7, r*1.1, r*1.6, r*0.12, rgb(0x4a2a16))
		rect(dst, x-r*0.55, y-r*0.15, r*1.1, r*0.45, rgb(0xd9a55a))
		ellipse(dst, x, y+r*0.07, r*0.18, r*0.12, rgb(0x6a3a1e))
		polygon(dst, []V{{X: x - r*0.55, Y: y - r*0.7}, {X: x, Y: y - r*0.95}, {X: x + r*0.55, Y: y - r*0.7}}, rgb(0x3a1e0e))
	case sim.Beurre:
		// plaquette de beurre dans son papier
		polygon(dst, []V{{X: x - r*0.95, Y: y + r*0.2}, {X: x + r*0.55, Y: y + r*0.55}, {X: x + r*0.95, Y: y + r*0.1}, {X: x - r*0.55, Y: y - r*0.25}}, rgb(0xd8e6f0))
		roundRect(dst, x-r*0.6, y-r*0.5, r*1.2, r*0.7, r*0.1, rgb(0xffe27a))
		rect(dst, x-r*0.6, y-r*0.5, r*1.2, r*0.18, rgb(0xfff0a8))
		roundStroke(dst, x-r*0.6, y-r*0.5, r*1.2, r*0.7, r*0.1, r*0.05, dark)
	case sim.FromageBlanc:
		// faisselle : bol blanc, fromage tout blanc
		ellipse(dst, x, y+r*0.1, r*0.85, r*0.3, rgb(0xfdfdfd))
		polygon(dst, []V{{X: x - r*0.85, Y: y + r*0.1}, {X: x + r*0.85, Y: y + r*0.1}, {X: x + r*0.55, Y: y + r*0.8}, {X: x - r*0.55, Y: y + r*0.8}}, rgb(0x5aa0e0))
		ellipse(dst, x, y+r*0.05, r*0.7, r*0.22, rgb(0xffffff))
		circle(dst, x+r*0.1, y-r*0.15, r*0.25, rgb(0xffffff))
	case sim.FromageChevre:
		// bûche de chèvre, cendrée
		roundRect(dst, x-r*0.9, y-r*0.35, r*1.8, r*0.75, r*0.35, rgb(0xf6f3ea))
		ellipse(dst, x+r*0.75, y+r*0.02, r*0.2, r*0.36, rgb(0xe6e0cf))
		roundStroke(dst, x-r*0.9, y-r*0.35, r*1.8, r*0.75, r*0.35, r*0.05, rgba(0x6a6a5a, 0.5))
		for k := 0; k < 4; k++ {
			circle(dst, x-r*0.6+float64(k)*r*0.35, y-r*0.15+float64(k%2)*r*0.25, r*0.05, rgb(0xa0a090))
		}
	case sim.FromageBrebis:
		// tomme de brebis : croûte brune, pâte crème
		polygon(dst, []V{{X: x - r*0.9, Y: y + r*0.55}, {X: x + r*0.9, Y: y + r*0.55}, {X: x + r*0.9, Y: y - r*0.1}, {X: x - r*0.6, Y: y - r*0.6}}, rgb(0xf3e3b5))
		polygon(dst, []V{{X: x - r*0.6, Y: y - r*0.6}, {X: x + r*0.9, Y: y - r*0.1}, {X: x + r*0.9, Y: y - r*0.28}, {X: x - r*0.55, Y: y - r*0.8}}, rgb(0x9a6a3a))
		line(dst, x+r*0.9, y-r*0.1, x+r*0.9, y+r*0.55, r*0.14, rgb(0x9a6a3a))
	case sim.Yaourt, sim.YaourtChevre, sim.YaourtBrebis:
		c := map[sim.Item]uint32{sim.Yaourt: 0x5aa0e0, sim.YaourtChevre: 0x6aa84f, sim.YaourtBrebis: 0xc0783a}[it]
		polygon(dst, []V{{X: x - r*0.55, Y: y - r*0.45}, {X: x + r*0.55, Y: y - r*0.45}, {X: x + r*0.4, Y: y + r*0.85}, {X: x - r*0.4, Y: y + r*0.85}}, rgb(0xffffff))
		rect(dst, x-r*0.48, y, r*0.96, r*0.35, rgb(c))
		roundRect(dst, x-r*0.65, y-r*0.7, r*1.3, r*0.3, r*0.08, rgb(c))
		roundStroke(dst, x-r*0.65, y-r*0.7, r*1.3, r*0.3, r*0.08, r*0.05, dark)
	case sim.Croissant:
		for k := -2; k <= 2; k++ {
			a := float64(k) * 0.55
			cx, cy := x+math.Sin(a)*r*0.6, y+r*0.25-math.Cos(a)*r*0.45
			rr := r * (0.42 - math.Abs(float64(k))*0.08)
			circle(dst, cx, cy, rr, rgb(0xc98a3d))
			circle(dst, cx-rr*0.15, cy-rr*0.2, rr*0.7, rgb(0xe8b060))
		}
	case sim.PainChoco:
		roundRect(dst, x-r*0.85, y-r*0.45, r*1.7, r*0.9, r*0.3, rgb(0xc98a3d))
		roundRect(dst, x-r*0.75, y-r*0.38, r*1.5, r*0.5, r*0.22, rgb(0xe8b060))
		rect(dst, x-r*0.95, y-r*0.15, r*0.25, r*0.3, rgb(0x5a3220))
		rect(dst, x+r*0.7, y-r*0.15, r*0.25, r*0.3, rgb(0x5a3220))
	case sim.Fleur:
		cols := []uint32{0xff6fa8, 0xffd23f, 0xb07cff}
		for k, c := range cols {
			cx := x + float64(k-1)*r*0.48
			cy := y - r*0.3 + math.Abs(float64(k-1))*r*0.2
			line(dst, cx, cy, x, y+r*0.9, r*0.08, rgb(0x3f9b35))
			for i := 0; i < 5; i++ {
				a := float64(i) * 2 * math.Pi / 5
				circle(dst, cx+math.Cos(a)*r*0.2, cy+math.Sin(a)*r*0.2, r*0.17, rgb(c))
			}
			circle(dst, cx, cy, r*0.12, rgb(0xffffff))
		}
		polygon(dst, []V{{X: x - r*0.4, Y: y + r*0.35}, {X: x + r*0.4, Y: y + r*0.35}, {X: x + r*0.15, Y: y + r}, {X: x - r*0.15, Y: y + r}}, rgb(0xd9b38c))
	case sim.Miel:
		roundRect(dst, x-r*0.6, y-r*0.45, r*1.2, r*1.35, r*0.3, rgb(0xf2a516))
		roundRect(dst, x-r*0.65, y-r*0.75, r*1.3, r*0.35, r*0.1, rgb(0xb5651d))
		roundRect(dst, x-r*0.4, y-r*0.1, r*0.8, r*0.5, r*0.1, rgb(0xfff2c8))
		say(dst, "M", x, y+r*0.15, r*0.45, true, 0, rgb(0xb5651d))
	case sim.Gateau:
		ellipse(dst, x, y+r*0.55, r*0.95, r*0.3, rgb(0xdddddd))
		rect(dst, x-r*0.75, y-r*0.15, r*1.5, r*0.7, rgb(0x7b4a2d))
		ellipse(dst, x, y+r*0.55, r*0.75, r*0.18, rgb(0x6a3d24))
		ellipse(dst, x, y-r*0.15, r*0.75, r*0.22, rgb(0x4a2a1a)) // glaçage au chocolat
		rect(dst, x-r*0.75, y-r*0.15, r*1.5, r*0.18, rgb(0x4a2a1a))
		circle(dst, x, y-r*0.5, r*0.17, rgb(0xd62839))
		line(dst, x, y-r*0.62, x+r*0.12, y-r*0.85, r*0.06, rgb(0x3f9b35))
	case sim.FarineMais:
		// sac orangé, bandeau vert, épi de maïs avec ses feuilles
		roundRect(dst, x-r*0.65, y-r*0.6, r*1.3, r*1.5, r*0.35, rgb(0xf2a83a))
		roundRect(dst, x-r*0.45, y-r*0.95, r*0.9, r*0.5, r*0.2, rgb(0xd98a20))
		line(dst, x-r*0.4, y-r*0.55, x+r*0.4, y-r*0.55, r*0.1, rgb(0x3f9b35))
		polygon(dst, []V{{X: x - r*0.05, Y: y + r*0.75}, {X: x - r*0.42, Y: y + r*0.05}, {X: x - r*0.12, Y: y + r*0.35}}, rgb(0x5cb94a))
		polygon(dst, []V{{X: x + r*0.05, Y: y + r*0.75}, {X: x + r*0.42, Y: y + r*0.05}, {X: x + r*0.12, Y: y + r*0.35}}, rgb(0x5cb94a))
		ellipse(dst, x, y+r*0.2, r*0.2, r*0.42, rgb(0xffe066))
		for k := 0; k < 3; k++ {
			line(dst, x-r*0.15, y-r*0.05+float64(k)*r*0.2, x+r*0.15, y-r*0.05+float64(k)*r*0.2, r*0.04, rgb(0xd9a521))
		}
	case sim.Fruit:
		circle(dst, x-r*0.35, y+r*0.2, r*0.45, rgb(0xe23b2e))
		circle(dst, x+r*0.38, y+r*0.25, r*0.42, rgb(0xff9f1a))
		circle(dst, x, y-r*0.25, r*0.4, rgb(0x7ed957))
		line(dst, x, y-r*0.6, x+r*0.1, y-r*0.85, r*0.08, rgb(0x6b4423))
	case sim.Creme:
		// pot de crème entière : couvercle et bandeau rouges (le lait est bleu)
		roundRect(dst, x-r*0.6, y-r*0.35, r*1.2, r*1.25, r*0.2, rgb(0xffffff))
		rect(dst, x-r*0.6, y-r*0.05, r*1.2, r*0.5, rgb(0xd62839))
		circle(dst, x, y+r*0.2, r*0.17, rgb(0xffffff))
		roundRect(dst, x-r*0.7, y-r*0.7, r*1.4, r*0.4, r*0.12, rgb(0xd62839))
		roundStroke(dst, x-r*0.6, y-r*0.35, r*1.2, r*1.25, r*0.2, r*0.06, rgba(0x000000, 0.25))
		roundStroke(dst, x-r*0.7, y-r*0.7, r*1.4, r*0.4, r*0.12, r*0.06, rgba(0x000000, 0.25))
	case sim.Sucre:
		for k := 0; k < 3; k++ {
			cx := x + float64(k-1)*r*0.5
			cy := y + r*0.2 - float64(k%2)*r*0.45
			roundRect(dst, cx-r*0.3, cy-r*0.3, r*0.6, r*0.6, r*0.08, rgb(0xffffff))
			roundStroke(dst, cx-r*0.3, cy-r*0.3, r*0.6, r*0.6, r*0.08, r*0.05, rgba(0x000000, 0.2))
		}
	case sim.Cacao:
		ellipse(dst, x, y, r*0.55, r*0.9, rgb(0xb5651d))
		line(dst, x, y-r*0.85, x, y+r*0.85, r*0.08, rgb(0x8a4a1a))
		line(dst, x-r*0.3, y-r*0.75, x-r*0.3, y+r*0.75, r*0.06, rgb(0x8a4a1a))
		line(dst, x+r*0.3, y-r*0.75, x+r*0.3, y+r*0.75, r*0.06, rgb(0x8a4a1a))
	case sim.Chocolat:
		roundRect(dst, x-r*0.8, y-r*0.6, r*1.6, r*1.2, r*0.1, rgb(0x5a3220))
		for i := 0; i < 3; i++ {
			for j := 0; j < 2; j++ {
				roundRect(dst, x-r*0.72+float64(i)*r*0.5, y-r*0.52+float64(j)*r*0.55, r*0.42, r*0.46, r*0.06, rgb(0x7a4a2d))
			}
		}
	case sim.Saucisse:
		// deux saucisses attachées
		for k := 0; k < 2; k++ {
			cx := x - r*0.42 + float64(k)*r*0.84
			roundRect(dst, cx-r*0.38, y-r*0.8, r*0.76, r*1.6, r*0.38, rgb(0xa0402a))
			roundRect(dst, cx-r*0.22, y-r*0.6, r*0.18, r*0.9, r*0.09, rgba(0xffffff, 0.35))
		}
		circle(dst, x, y-r*0.75, r*0.12, rgb(0xd9b38c))
	case sim.Fromage:
		// une part de fromage à trous
		polygon(dst, []V{{X: x - r*0.9, Y: y + r*0.55}, {X: x + r*0.9, Y: y + r*0.55}, {X: x + r*0.9, Y: y - r*0.1}, {X: x - r*0.6, Y: y - r*0.6}}, rgb(0xf6c945))
		polygon(dst, []V{{X: x - r*0.9, Y: y + r*0.55}, {X: x - r*0.6, Y: y - r*0.6}, {X: x + r*0.9, Y: y - r*0.1}, {X: x + r*0.9, Y: y - r*0.25}, {X: x - r*0.55, Y: y - r*0.8}, {X: x - r*0.9, Y: y + r*0.3}}, rgb(0xffe27a))
		circle(dst, x-r*0.2, y+r*0.15, r*0.14, rgb(0xd9a521))
		circle(dst, x+r*0.4, y+r*0.25, r*0.1, rgb(0xd9a521))
		circle(dst, x+r*0.2, y-r*0.1, r*0.08, rgb(0xd9a521))
	case sim.Canne:
		for k := -1; k <= 1; k++ {
			cx := x + float64(k)*r*0.35
			line(dst, cx, y+r*0.95, cx+float64(k)*r*0.1, y-r*0.9, r*0.2, rgb(0x9fbf4a))
			for j := -1; j <= 1; j++ {
				line(dst, cx-r*0.12, y+float64(j)*r*0.45, cx+r*0.12, y+float64(j)*r*0.45, r*0.06, rgb(0x6f8f2a))
			}
		}
	}
}

// coin dessine une pièce d'or.
func coin(dst *ebiten.Image, x, y, r float64) {
	circle(dst, x, y+r*0.12, r, rgb(0xb8860b))
	circle(dst, x, y, r, rgb(0xffcc33))
	circle(dst, x, y, r*0.72, rgb(0xffdd66))
	say(dst, "€", x, y, r*1.15, true, 0, rgb(0xb8860b))
}
