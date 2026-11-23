package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
	"ville/tr"
)

// Les voleurs (dessin) : cinq personnes en noir, visage découvert (la vidéosurveillance pourra
// les reconnaître), un pied-de-biche à la main et un sac de butin sur l'épaule.

// theftPlace : le lieu attaqué en ce moment ("" : aucun) ; les employés y lèvent les bras.
var theftPlace string

var robberSkins = []uint32{0xf1c9a5, 0xc68863, 0x8d5a3b, 0xe0ac85, 0x5c3a24}

func drawRobber(dst *ebiten.Image, r *sim.Robber, t float64) {
	lk := look{skin: robberSkins[r.Look%len(robberSkins)], hair: 0x2b2118, shirt: 0x24252b, pants: 0x16171b, hat: 2, hatC: 0x111114}
	// le sac de butin, sur l'épaule (derrière lui)
	sx, sy := r.Pos.X-20, r.Pos.Y-62
	circle(dst, sx, sy, 17, rgb(0x9c7a4a))
	circle(dst, sx-2, sy-3, 14, rgb(0xb8925a))
	say(dst, "€", sx-1, sy-1, 15, true, 0, rgb(0x5a3d1a))
	drawPerson(dst, r.Pos, r.Face, r.WalkT, r.Moving, lk, nil)
	// le pied-de-biche
	hx, hy := r.Pos.X+18, r.Pos.Y-36
	sw := 0.0
	if r.Moving {
		sw = math.Sin(t*8+float64(r.Look)) * 4
	}
	line(dst, hx, hy+sw, hx+10, hy-30+sw, 5, rgb(0x6a6f76))
	line(dst, hx+10, hy-30+sw, hx+2, hy-38+sw, 5, rgb(0x6a6f76))
}

// drawTheftHUD : la carte du vol en cours (à la place de celle de la commande).
func (g *game) drawTheftHUD(dst *ebiten.Image, x, y float64) {
	th := g.sim.Theft
	w, h := 420.0, 104.0
	k := 0.5 + 0.5*math.Sin(g.t*5)
	roundRect(dst, x, y, w, h, 16, rgba(0x5a1010, 0.85))
	roundRect(dst, x+10, y+6, w-20, 26, 8, rgba(0xd9342b, 0.8+0.2*k))
	title := tr.F("Vol en cours : {nom}", "nom", tr.T(busNames[th.Place]))
	if th.Fire {
		title = tr.F("Incendie : {nom}", "nom", tr.T(busNames[th.Place]))
	}
	sayFit(dst, title, x+w/2, y+19, 16, true, 0, w-36, rgb(0xffffff))
	m, sec := int(th.Left)/60, int(th.Left)%60
	ts := itoa(m) + ":" + map[bool]string{true: "0", false: ""}[sec < 10] + itoa(sec)
	sub := tr.T("Tout est arrêté. Les voleurs partiront dans :")
	if th.Fire {
		sub = tr.T("Tout le monde fuit. Le feu sera éteint dans :")
	}
	sayFit(dst, sub, x+20, y+52, 15, false, -1, w-110, rgb(0xffd9d9))
	say(dst, ts, x+w-16, y+52, 20, true, 1, rgb(0xffffff))
	switch {
	case th.Fire:
		sayFit(dst, tr.F("Les pompiers sauveront {n} élément(s).", "n", itoa(g.sim.SaveCapacity())), x+20, y+82, 15, true, -1, w-40, rgb(0xffe066))
	case th.Spotted:
		sayFit(dst, "Repérés par la vidéosurveillance : la police les attend !", x+20, y+82, 15, true, -1, w-40, rgb(0x9fe0ff))
	case th.Stolen > 0:
		sayFit(dst, tr.F("Volé dans les caisses : {prix}", "prix", money(th.Stolen)), x+20, y+82, 15, true, -1, w-40, rgb(0xffe066))
	}
}
