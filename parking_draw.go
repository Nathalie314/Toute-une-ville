package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// ---------------------------------------------------------------- parking

var carColors = []uint32{0xd9534f, 0x4a90d9, 0xf1c40f, 0x2ecc71, 0xecf0f1, 0x34495e, 0x9b59b6, 0xe67e22, 0x95a5a6, 0x1abc9c}

// drawParking : trottoir, parking (places, allée, coin des deux-roues) et
// passage piéton devant le portail.
func drawParking(dst *ebiten.Image, p *sim.Place, tl, br V) {
	if len(p.Parks) == 0 || p.ID == "super" || p.ID == "parc" {
		return // (le supermarché a son propre parking)
	}
	const roadY = 1640.0
	x0, x1, gx := p.ParkX()
	// trottoir tout du long
	rect(dst, tl.X-10, roadY+162, br.X-tl.X+20, sim.ParkTop-roadY-162, rgb(0xc9c3b8))
	for x := math.Floor((tl.X-100)/60) * 60; x < br.X+100; x += 60 {
		line(dst, x, roadY+162, x, sim.ParkTop, 1.5, rgba(0xa8a296, 0.6))
	}
	// passage piéton (sans les pointillés de la route)
	rect(dst, gx-76, roadY+4, 152, 142, rgb(0x6b6f74))
	for y := roadY + 12.0; y < roadY+140; y += 26 {
		rect(dst, gx-70, y, 140, 14, rgb(0xf2f2e6))
	}
	// le parking
	roundRect(dst, x0-12, sim.ParkTop-12, x1-x0+24, sim.ParkBot-sim.ParkTop+24, 18, rgb(0x9a9a92))
	roundRect(dst, x0, sim.ParkTop, x1-x0, sim.ParkBot-sim.ParkTop, 12, rgb(0x5f6368))
	// entrées (bateau sur le trottoir) et flèches de sens unique
	for i, ex := range []float64{p.ParkEntry(), p.ParkExit()} {
		rect(dst, ex-50, roadY+150, 100, sim.ParkTop-roadY-150+2, rgb(0x5f6368))
		ay := sim.ParkAisle
		dir := 1.0
		if i == 1 {
			ay = sim.ParkTop + 40
			dir = -1 // sortie : la flèche monte vers la route
			polygon(dst, []V{{X: ex, Y: ay - 22}, {X: ex - 14, Y: ay}, {X: ex + 14, Y: ay}}, rgba(0xffffff, 0.85))
			rect(dst, ex-4, ay, 8, 22, rgba(0xffffff, 0.85))
			continue
		}
		_ = dir
		polygon(dst, []V{{X: ex + 26, Y: ay}, {X: ex + 6, Y: ay - 13}, {X: ex + 6, Y: ay + 13}}, rgba(0xffffff, 0.85))
		rect(dst, ex-20, ay-4, 26, 8, rgba(0xffffff, 0.85))
	}
	// places
	for _, s := range p.Parks[:p.ParksBought()] {
		if s.Bike {
			continue
		}
		top := s.Pos.Y - 55
		line(dst, s.Pos.X-40, top, s.Pos.X-40, top+110, 3, rgba(0xffffff, 0.8))
		line(dst, s.Pos.X+40, top, s.Pos.X+40, top+110, 3, rgba(0xffffff, 0.8))
	}
	// allée piétonne face au portail
	for y := sim.ParkTop + 8; y < sim.ParkBot-8; y += 30 {
		rect(dst, gx-34, y, 68, 16, rgba(0xf2f2e6, 0.9))
	}
	// coin des deux-roues : un arceau par place
	bx0, bx1 := 1e9, -1e9
	for _, s := range p.Parks {
		if s.Bike {
			bx0, bx1 = math.Min(bx0, s.Pos.X), math.Max(bx1, s.Pos.X)
		}
	}
	if bx1 > bx0 {
		roundRect(dst, bx0-22, sim.ParkRow1-55, bx1-bx0+44, 110, 8, rgb(0x6f7378))
		say(dst, "2 roues", (bx0+bx1)/2, sim.ParkRow1+70, 15, true, 0, rgba(0xffffff, 0.75))
	}
}

// drawVehicle : une voiture, un vélo, une moto ou une trottinette.
func drawVehicle(dst *ebiten.Image, v *sim.Vehicle) {
	x, y := v.Pos.X, v.Pos.Y
	side := math.Abs(v.Dir.X) > math.Abs(v.Dir.Y) // vu de profil (sur la route)
	col := carColors[(v.Look/7)%len(carColors)]
	lk := lookFrom(v.Look)
	bump := 0.0
	if v.Moving {
		bump = math.Abs(math.Sin(v.T*18)) * 1.2
	}
	switch v.Kind {
	case sim.VCar:
		if side {
			fx := v.Dir.X // l'avant
			ellipse(dst, x, y+2, 62, 10, rgba(0x000000, 0.25))
			roundRect(dst, x-60, y-38-bump, 120, 30, 12, rgb(col))
			roundRect(dst, x-34-fx*6, y-62-bump, 62, 28, 12, rgb(col))
			// vitres
			roundRect(dst, x-28-fx*6, y-57-bump, 24, 20, 5, rgb(0xbfe3f5))
			roundRect(dst, x+0-fx*6, y-57-bump, 24, 20, 5, rgb(0xbfe3f5))
			if v.Rider {
				hx := x + fx*10 - fx*6
				circle(dst, hx, y-48-bump, 8, rgb(lk.skin))
				ellipse(dst, hx, y-54-bump, 8, 4, rgb(lk.hair))
			}
			// phares
			circle(dst, x+fx*56, y-28-bump, 5, rgb(0xfff3b0))
			circle(dst, x-fx*57, y-28-bump, 4, rgb(0xe74c3c))
			for _, wx := range []float64{-34, 34} {
				circle(dst, x+wx, y-8, 13, rgb(0x2b2b2b))
				circle(dst, x+wx, y-8, 6, rgb(0xb0b0b0))
			}
		} else {
			// de face (ou de dos) : garée, ou qui monte / descend
			ellipse(dst, x, y+28, 40, 10, rgba(0x000000, 0.25))
			rect(dst, x-30, y+10, 12, 20, rgb(0x2b2b2b))
			rect(dst, x+18, y+10, 12, 20, rgb(0x2b2b2b))
			roundRect(dst, x-34, y-16-bump, 68, 40, 12, rgb(col))
			roundRect(dst, x-26, y-44-bump, 52, 32, 12, rgb(col))
			roundRect(dst, x-21, y-39-bump, 42, 20, 5, rgb(0xbfe3f5))
			if v.Rider {
				circle(dst, x-8, y-28-bump, 7, rgb(lk.skin))
				ellipse(dst, x-8, y-33-bump, 7, 3.5, rgb(lk.hair))
			}
			light := uint32(0xfff3b0)
			if v.Dir.Y < 0 {
				light = 0xe74c3c // on voit l'arrière
			}
			circle(dst, x-24, y+2-bump, 5, rgb(light))
			circle(dst, x+24, y+2-bump, 5, rgb(light))
			rect(dst, x-12, y+8-bump, 24, 7, rgb(0xf2f2e6)) // plaque
		}
	default:
		face := v.Dir
		if !side {
			face = V{Y: 1}
		}
		if v.Kind == sim.VMoto {
			lk.hat, lk.hatC = 8, carColors[(v.Look/3)%len(carColors)]
		}
		if side {
			fx := v.Dir.X
			ellipse(dst, x, y+2, 40, 7, rgba(0x000000, 0.22))
			r := map[sim.VehKind]float64{sim.VBike: 15, sim.VMoto: 15, sim.VTrot: 7}[v.Kind]
			wb := map[sim.VehKind]float64{sim.VBike: 30, sim.VMoto: 34, sim.VTrot: 24}[v.Kind]
			// le pilote (assis ou debout)
			if v.Rider {
				lift := 22.0
				if v.Kind == sim.VTrot {
					lift = 10
				}
				drawPerson(dst, V{X: x - fx*4, Y: y - lift - bump}, face, 0, false, lk, nil)
			}
			for _, wx := range []float64{-wb, wb} {
				circle(dst, x+wx, y-r, r, rgb(0x2b2b2b))
				circle(dst, x+wx, y-r, r*0.55, rgb(0xd0d0d0))
			}
			switch v.Kind {
			case sim.VBike:
				line(dst, x-wb, y-r, x, y-r-4, 3, rgb(col))
				line(dst, x, y-r-4, x+wb, y-r, 3, rgb(col))
				line(dst, x-wb*0.4, y-r-20, x, y-r-4, 3, rgb(col))
				line(dst, x+fx*wb, y-r, x+fx*wb*0.8, y-r-26, 3, rgb(col))
				line(dst, x+fx*wb*0.8, y-r-26, x+fx*wb*0.55, y-r-28, 3, rgb(0x333333))
			case sim.VMoto:
				roundRect(dst, x-wb+4, y-r-22, wb*2-8, 18, 8, rgb(col))
				roundRect(dst, x-fx*6-14, y-r-30, 28, 10, 5, rgb(0x333333))
				circle(dst, x+fx*(wb+6), y-r-14, 5, rgb(0xfff3b0))
			case sim.VTrot:
				rect(dst, x-wb, y-r-5, wb*2, 5, rgb(0x333333))
				line(dst, x+fx*wb, y-r-2, x+fx*(wb-4), y-r-58, 3, rgb(0x555555))
				line(dst, x+fx*(wb-12), y-r-58, x+fx*(wb+4), y-r-58, 3, rgb(0x333333))
			}
		} else {
			// garé (ou qui manœuvre) : vu de face, tout fin
			ellipse(dst, x, y+20, 16, 6, rgba(0x000000, 0.22))
			if v.Rider {
				drawPerson(dst, V{X: x, Y: y + 6 - bump}, face, 0, false, lk, nil)
			}
			h := map[sim.VehKind]float64{sim.VBike: 30, sim.VMoto: 32, sim.VTrot: 16}[v.Kind]
			roundRect(dst, x-4, y+20-h, 8, h, 4, rgb(0x2b2b2b))
			switch v.Kind {
			case sim.VBike:
				line(dst, x, y-6, x, y-30, 3, rgb(col))
				line(dst, x-14, y-30, x+14, y-30, 3, rgb(0x333333))
			case sim.VMoto:
				roundRect(dst, x-12, y-22, 24, 26, 8, rgb(col))
				circle(dst, x, y-14, 5, rgb(0xfff3b0))
				line(dst, x-18, y-26, x+18, y-26, 3, rgb(0x333333))
			case sim.VTrot:
				line(dst, x, y+6, x, y-44, 3, rgb(0x555555))
				line(dst, x-12, y-44, x+12, y-44, 3, rgb(0x333333))
			}
		}
	}
}
