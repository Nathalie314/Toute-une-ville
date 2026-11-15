package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// Les stands du parc : un petit kiosque surmonté, en géant, de ce qu'il vend.

// kiosk : le comptoir et son auvent rayé.
func kiosk(dst *ebiten.Image, x, y float64, body, stripe uint32, name string) {
	ellipse(dst, x+26, y+6, 96, 18, rgba(0x1d3a12, 0.25))
	roundRect(dst, x-72, y-74, 144, 74, 10, rgb(body))
	rect(dst, x-72, y-74, 18, 74, shade(body, 1.15))
	rect(dst, x-72, y-12, 144, 12, shade(body, 0.78))
	roundRect(dst, x-58, y-66, 116, 28, 6, rgb(0x3b2a52)) // le guichet
	roundRect(dst, x-66, y-32, 132, 22, 6, rgb(0x3b2a52))
	// auvent
	for i := 0; i < 8; i++ {
		c := rgb(stripe)
		if i%2 == 1 {
			c = rgb(0xffffff)
		}
		polygon(dst, []V{{X: x - 84 + float64(i)*21, Y: y - 100}, {X: x - 63 + float64(i)*21, Y: y - 100}, {X: x - 63 + float64(i)*21, Y: y - 72}, {X: x - 84 + float64(i)*21, Y: y - 72}}, c)
		circle(dst, x-73.5+float64(i)*21, y-72, 10.5, c)
	}
	rect(dst, x-84, y-104, 168, 6, shade(stripe, 0.8))
	sayFit(dst, name, x, y-21, 15, true, 0, 124, rgb(0xffffff))
}

func drawStand(dst *ebiten.Image, s *sim.Stand, t float64) {
	// le food truck et la boutique de souvenirs : à la taille du parc (agrandi)
	if k := map[string]float64{"foodtruck": 1.7, "souvenirs": 1.8}[s.Kind.ID]; k > 0 {
		scaled(s.Pos.X, s.Pos.Y, k, func() { drawStandBody(dst, s, t) })
		return
	}
	drawStandBody(dst, s, t)
}

func drawStandBody(dst *ebiten.Image, s *sim.Stand, t float64) {
	x, y := s.Pos.X, s.Pos.Y
	pop := 0.0
	if s.Sold < 0.4 { // il vient de vendre : le stand saute un peu
		pop = math.Sin(s.Sold/0.4*math.Pi) * 8
	}
	top := y - 104 - pop
	switch s.Kind.ID {
	case "ballons":
		kiosk(dst, x, y, 0x4fc3ff, 0xff5d8f, s.Kind.Name)
		cols := []uint32{0xff5d8f, 0xffd23f, 0x7ed957, 0x4fc3ff, 0xb36bd6, 0xff8a3d, 0xffffff}
		for i := 0; i < 7; i++ {
			a := -2.4 + float64(i)*0.8
			bx := x + math.Cos(a)*58 + math.Sin(t*1.5+float64(i))*4
			by := top - 70 + math.Sin(a)*40 + math.Cos(t*1.3+float64(i))*3
			line(dst, x, top, bx, by+22, 1.5, rgb(0x5a5a5a))
			ellipse(dst, bx, by, 20, 25, rgb(cols[i]))
			ellipse(dst, bx-6, by-8, 5, 8, rgba(0xffffff, 0.55))
		}
	case "glaces":
		kiosk(dst, x, y, 0xfff0f5, 0x4fc3ff, s.Kind.Name)
		// un cornet géant, trois boules
		polygon(dst, []V{{X: x - 34, Y: top - 60}, {X: x + 34, Y: top - 60}, {X: x, Y: top + 4}}, rgb(0xd9a35a))
		polygon(dst, []V{{X: x, Y: top - 60}, {X: x + 34, Y: top - 60}, {X: x, Y: top + 4}}, rgb(0xc08a42))
		for k := 0; k < 4; k++ {
			line(dst, x-28+float64(k)*14, top-58, x-8+float64(k)*10, top-14, 2, rgba(0x8a5a2b, 0.6))
		}
		circle(dst, x-20, top-74, 28, rgb(0xffb3c7))
		circle(dst, x+20, top-74, 28, rgb(0xfff2c2))
		circle(dst, x, top-108, 28, rgb(0x8a5a33))
		circle(dst, x-8, top-118, 8, rgba(0xffffff, 0.5))
		circle(dst, x, top-140, 9, rgb(0xd9534f)) // la cerise
	case "popcorn":
		kiosk(dst, x, y, 0xffd23f, 0xd9534f, s.Kind.Name)
		polygon(dst, []V{{X: x - 42, Y: top - 90}, {X: x + 42, Y: top - 90}, {X: x + 30, Y: top}, {X: x - 30, Y: top}}, rgb(0xffffff))
		for k := 0; k < 4; k++ {
			fk := float64(k)
			polygon(dst, []V{{X: x - 42 + fk*21, Y: top - 90}, {X: x - 31 + fk*21, Y: top - 90}, {X: x - 25 + fk*15, Y: top}, {X: x - 30 + fk*15, Y: top}}, rgb(0xd9534f))
		}
		for k := 0; k < 9; k++ {
			a := float64(k) * 0.7
			circle(dst, x-36+float64(k)*9, top-96-math.Abs(math.Sin(a))*16, 14, rgb(0xfff6d8))
			circle(dst, x-36+float64(k)*9+3, top-100-math.Abs(math.Sin(a))*16, 5, rgb(0xffe08a))
		}
	case "barbe":
		kiosk(dst, x, y, 0xb36bd6, 0xff8fc8, s.Kind.Name)
		line(dst, x, top, x, top-60, 6, rgb(0xf3ead8))
		for k := 0; k < 7; k++ {
			a := float64(k) * 0.9
			circle(dst, x+math.Cos(a)*26, top-92+math.Sin(a)*22, 30, rgb(0xff9fd2))
		}
		circle(dst, x-12, top-104, 22, rgb(0xffc4e6))
	case "photo":
		kiosk(dst, x, y, 0x3b4a5c, 0xffd23f, s.Kind.Name)
		// un gros appareil photo
		roundRect(dst, x-56, top-86, 112, 74, 14, rgb(0x2b2b3a))
		roundRect(dst, x-56, top-86, 112, 20, 10, rgb(0x4a4a5a))
		roundRect(dst, x-20, top-102, 44, 20, 6, rgb(0x2b2b3a))
		circle(dst, x, top-48, 30, rgb(0x5a5a6a))
		circle(dst, x, top-48, 22, rgb(0x1a2a4a))
		circle(dst, x-7, top-55, 7, rgba(0xffffff, 0.6))
		flash := 0.0
		if s.Sold < 0.3 {
			flash = 1 - s.Sold/0.3
		}
		roundRect(dst, x+28, top-80, 18, 12, 3, rgb(0xfff8c0))
		if flash > 0 {
			circle(dst, x+37, top-74, 40*flash, rgba(0xffffff, 0.8*flash))
		}
	case "foodtruck":
		// un camion : son côté s'ouvre en comptoir
		ellipse(dst, x+30, y+6, 130, 20, rgba(0x1d3a12, 0.25))
		roundRect(dst, x-120, y-120, 200, 108, 16, rgb(0xff8a3d))
		rect(dst, x-120, y-120, 200, 14, rgb(0xffb070))
		roundRect(dst, x+70, y-92, 60, 80, 12, rgb(0xff8a3d))
		roundRect(dst, x+84, y-84, 38, 30, 6, rgb(0x9fd3ff))
		roundRect(dst, x-104, y-100, 150, 46, 6, rgb(0x3b2a52))
		polygon(dst, []V{{X: x - 112, Y: y - 104}, {X: x + 54, Y: y - 104}, {X: x + 64, Y: y - 134}, {X: x - 122, Y: y - 134}}, rgb(0xffd23f))
		for _, wx := range []float64{-80, 70} {
			circle(dst, x+wx, y-10, 20, rgb(0x2b2b2b))
			circle(dst, x+wx, y-10, 9, rgb(0xb0b0b0))
		}
		sayFit(dst, s.Kind.Name, x-30, y-32, 16, true, 0, 150, rgb(0xffffff))
		// le burger géant sur le toit
		bx, by := x-30, y-160-pop
		ellipse(dst, bx, by-8, 50, 26, rgb(0xe0a050))
		rect(dst, bx-50, by-2, 100, 10, rgb(0x6aa84f))
		rect(dst, bx-46, by+6, 92, 14, rgb(0x6e3a1e))
		ellipse(dst, bx, by+26, 48, 12, rgb(0xd99040))
		for k := 0; k < 5; k++ {
			circle(dst, bx-28+float64(k)*14, by-20+float64(k%2)*6, 3, rgb(0xfff2c2))
		}
	case "canards":
		// un bassin rond où tournent des canards, sous un chapiteau
		ellipse(dst, x+24, y+4, 110, 22, rgba(0x1d3a12, 0.25))
		ellipse(dst, x, y-30, 100, 42, rgb(0xd9534f))
		ellipse(dst, x, y-36, 88, 34, rgb(0x6fc3f0))
		for k := 0; k < 8; k++ {
			a := t*0.8 + float64(k)*math.Pi/4
			dx, dy := x+math.Cos(a)*66, y-36+math.Sin(a)*22
			ellipse(dst, dx, dy, 11, 7, rgb(0xffd23f))
			circle(dst, dx+6*math.Cos(a+math.Pi/2), dy-6, 6, rgb(0xffd23f))
		}
		line(dst, x-90, y-36, x-90, y-150, 5, rgb(0xf3ead8))
		line(dst, x+90, y-36, x+90, y-150, 5, rgb(0xf3ead8))
		for i := 0; i < 8; i++ {
			c := rgb(0xd9534f)
			if i%2 == 1 {
				c = rgb(0xffffff)
			}
			polygon(dst, []V{{X: x - 100 + float64(i)*25, Y: y - 150}, {X: x - 75 + float64(i)*25, Y: y - 150}, {X: x, Y: y - 200 - pop}}, c)
		}
		roundRect(dst, x-70, y+8, 140, 26, 8, rgb(0x3b2a52))
		sayFit(dst, s.Kind.Name, x, y+21, 15, true, 0, 130, rgb(0xffffff))
	case "souvenirs":
		ellipse(dst, x+30, y+6, 120, 20, rgba(0x1d3a12, 0.25))
		roundRect(dst, x-100, y-110, 200, 110, 10, rgb(0xf3ead8))
		rect(dst, x-100, y-110, 22, 110, rgb(0xfff8ea))
		rect(dst, x-100, y-14, 200, 14, rgb(0xd9ccb2))
		roundRect(dst, x-80, y-90, 70, 56, 6, rgb(0x9fd3ff))
		roundRect(dst, x+16, y-90, 64, 90, 6, rgb(0x6c4bb8))
		circle(dst, x+68, y-44, 4, rgb(0xffd23f))
		polygon(dst, []V{{X: x - 116, Y: y - 104}, {X: x + 116, Y: y - 104}, {X: x + 84, Y: y - 150}, {X: x - 84, Y: y - 150}}, rgb(0x8a63d6))
		polygon(dst, []V{{X: x, Y: y - 104}, {X: x + 116, Y: y - 104}, {X: x + 84, Y: y - 150}, {X: x, Y: y - 150}}, rgb(0x6c4bb8))
		sayFit(dst, s.Kind.Name, x, y-127, 15, true, 0, 190, rgb(0xffffff))
		// un paquet cadeau géant
		gx, gy := x, y-150-pop
		rect(dst, gx-36, gy-62, 72, 60, rgb(0xff5d8f))
		rect(dst, gx-36, gy-62, 20, 60, rgb(0xff86ab))
		rect(dst, gx-7, gy-62, 14, 60, rgb(0xffd23f))
		ellipse(dst, gx-16, gy-70, 16, 10, rgb(0xffd23f))
		ellipse(dst, gx+16, gy-70, 16, 10, rgb(0xffd23f))
	}
}

// drawHeld : ce qu'un visiteur a acheté, dans sa main.
func drawHeld(dst *ebiten.Image, v *sim.Visitor, t float64) {
	if v.Holding <= 0 || v.Holding > len(sim.StandKinds) {
		return
	}
	x, y := v.Pos.X+16, v.Pos.Y-52
	if v.OnTrott() {
		y -= 12
	}
	switch sim.StandKinds[v.Holding-1].ID {
	case "ballons":
		cols := []uint32{0xff5d8f, 0xffd23f, 0x7ed957, 0x4fc3ff, 0xb36bd6}
		bx := x + 6 + math.Sin(t*2+float64(v.Look))*4
		line(dst, x, y, bx, y-74, 1.5, rgb(0x5a5a5a))
		ellipse(dst, bx, y-88, 15, 19, rgb(cols[v.Look%len(cols)]))
		ellipse(dst, bx-5, y-94, 4, 6, rgba(0xffffff, 0.55))
	case "glaces":
		polygon(dst, []V{{X: x - 7, Y: y - 10}, {X: x + 7, Y: y - 10}, {X: x, Y: y + 6}}, rgb(0xd9a35a))
		circle(dst, x, y-14, 8, rgb([]uint32{0xffb3c7, 0xfff2c2, 0x8a5a33}[v.Look%3]))
	case "popcorn":
		polygon(dst, []V{{X: x - 9, Y: y - 14}, {X: x + 9, Y: y - 14}, {X: x + 7, Y: y + 4}, {X: x - 7, Y: y + 4}}, rgb(0xd9534f))
		circle(dst, x-4, y-16, 5, rgb(0xfff6d8))
		circle(dst, x+4, y-17, 5, rgb(0xfff6d8))
	case "barbe":
		line(dst, x, y+6, x, y-12, 2, rgb(0xf3ead8))
		circle(dst, x, y-20, 11, rgb(0xff9fd2))
	case "photo":
		rect(dst, x-10, y-14, 20, 16, rgb(0xffffff))
		rect(dst, x-8, y-12, 16, 10, rgb(0x6fb7e6))
	case "foodtruck":
		ellipse(dst, x, y-6, 10, 6, rgb(0xe0a050))
		rect(dst, x-10, y-3, 20, 4, rgb(0x6e3a1e))
	case "canards":
		ellipse(dst, x, y-6, 9, 6, rgb(0xffd23f))
		circle(dst, x+5, y-12, 5, rgb(0xffd23f))
	case "souvenirs":
		roundRect(dst, x-9, y-16, 18, 18, 3, rgb(0x8a63d6))
		line(dst, x-5, y-16, x-3, y-22, 2, rgb(0x5d3fa0))
		line(dst, x+5, y-16, x+3, y-22, 2, rgb(0x5d3fa0))
	}
}

// standDrawables : les stands ouverts (et la case d'achat des autres reste au sol).
func standDrawables(dst *ebiten.Image, p *sim.Place, ds []drawable, tl, br V, t float64) []drawable {
	for _, s := range p.Parc.Stands {
		s := s
		if !p.StandBuilt(s) || s.Pos.X < tl.X-200 || s.Pos.X > br.X+200 {
			continue
		}
		ds = append(ds, drawable{s.Pos.Y, func() { drawStand(dst, s, t) }})
	}
	return ds
}
