package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// La pâtisserie : les icônes de ses gâteaux (le gâteau au chocolat est dans icons.go).

// drawItemPatisserie : faux si ce n'est pas une pâtisserie.
func drawItemPatisserie(dst *ebiten.Image, it sim.Item, x, y, r float64) bool {
	switch it {
	case sim.TartePommes:
		// une tarte ronde, des lamelles de pomme en rosace
		ellipse(dst, x, y+r*0.2, r*0.95, r*0.55, rgb(0xc98a3d))
		ellipse(dst, x, y+r*0.1, r*0.8, r*0.42, rgb(0xf6dfa0))
		for k := 0; k < 7; k++ {
			a := float64(k) * 0.9
			ellipse(dst, x+r*0.5*math.Cos(a), y+r*0.1+r*0.25*math.Sin(a), r*0.22, r*0.12, rgb(0xf2c46a))
		}
		circle(dst, x, y+r*0.1, r*0.15, rgb(0xe8a83a))
	case sim.RizLait:
		// une verrine, le riz crémeux et une gousse de vanille
		roundRect(dst, x-r*0.55, y-r*0.5, r*1.1, r*1.25, r*0.15, rgba(0xdff0f4, 0.9))
		roundRect(dst, x-r*0.48, y-r*0.2, r*0.96, r*0.9, r*0.12, rgb(0xfbf3dc))
		for k := 0; k < 6; k++ {
			ellipse(dst, x-r*0.3+float64(k%3)*r*0.3, y+float64(k/3)*r*0.3, r*0.08, r*0.05, rgb(0xffffff))
		}
		line(dst, x-r*0.35, y-r*0.35, x+r*0.4, y-r*0.6, r*0.09, rgb(0x3a2418))
	case sim.Clafoutis:
		// un plat en terre, la pâte dorée et les cerises
		ellipse(dst, x, y+r*0.25, r*0.95, r*0.5, rgb(0xd9d2c8))
		ellipse(dst, x, y+r*0.15, r*0.8, r*0.38, rgb(0xf0c060))
		for k := 0; k < 5; k++ {
			circle(dst, x-r*0.5+float64(k)*r*0.25, y+r*0.1+float64(k%2)*r*0.15, r*0.12, rgb(0x9e1530))
		}
	case sim.Eclair:
		// un éclair allongé, glaçage café
		roundRect(dst, x-r*0.95, y-r*0.15, r*1.9, r*0.6, r*0.3, rgb(0xd9a55a))
		roundRect(dst, x-r*0.9, y-r*0.22, r*1.8, r*0.32, r*0.16, rgb(0x8a5c3a))
		line(dst, x-r*0.6, y-r*0.12, x+r*0.5, y-r*0.12, r*0.06, rgba(0xffffff, 0.4))
	case sim.TarteCitron:
		// une tarte ronde : la pâte dorée, la crème au citron jaune, des pointes de
		// meringue blanche dorées au bout, une rondelle de citron
		ellipse(dst, x, y+r*0.25, r*0.95, r*0.55, rgb(0xc98a3d))
		ellipse(dst, x, y+r*0.15, r*0.82, r*0.44, rgb(0xf5d84a))
		for k := 0; k < 7; k++ {
			a := float64(k)*0.9 + 0.3
			mx, my := x+r*0.5*math.Cos(a), y+r*0.12+r*0.24*math.Sin(a)
			circle(dst, mx, my, r*0.17, rgb(0xfffaf0))
			polygon(dst, []V{{X: mx - r*0.12, Y: my - r*0.05}, {X: mx + r*0.12, Y: my - r*0.05}, {X: mx, Y: my - r*0.3}}, rgb(0xfffaf0))
			circle(dst, mx, my-r*0.27, r*0.05, rgb(0xd9a04a))
		}
		circle(dst, x, y+r*0.05, r*0.2, rgb(0xfffaf0))
		circle(dst, x, y-r*0.05, r*0.14, rgb(0xf5e04a))
		circle(dst, x, y-r*0.05, r*0.09, rgb(0xfff6b0))
	case sim.Fraisier:
		// une part de fraisier : génoise, crème, fraises coupées, dessus rose
		polygon(dst, []V{{X: x - r*0.9, Y: y + r*0.6}, {X: x + r*0.9, Y: y + r*0.6}, {X: x + r*0.9, Y: y - r*0.5}, {X: x - r*0.9, Y: y - r*0.5}}, rgb(0xfff4e6))
		rect(dst, x-r*0.9, y+r*0.4, r*1.8, r*0.2, rgb(0xe8c27a))
		rect(dst, x-r*0.9, y-r*0.5, r*1.8, r*0.2, rgb(0xff9ec4))
		for k := 0; k < 4; k++ {
			ellipse(dst, x-r*0.65+float64(k)*r*0.43, y-r*0.05, r*0.18, r*0.22, rgb(0xe8333f))
			ellipse(dst, x-r*0.65+float64(k)*r*0.43, y-r*0.05, r*0.08, r*0.1, rgb(0xff8a95))
		}
		circle(dst, x, y-r*0.7, r*0.18, rgb(0xe8333f))
		line(dst, x, y-r*0.85, x+r*0.12, y-r*0.98, r*0.06, rgb(0x3f9b35))
	default:
		return false
	}
	return true
}
