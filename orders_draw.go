package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
	"ville/tr"
)

// Commandes spéciales et objectif affiché en haut de l'écran.

func (g *game) drawOrderBox(dst *ebiten.Image, s *sim.Station) {
	b := s.Body
	// tableau d'affichage
	rect(dst, b.X+20, b.Y-120, 8, 120, rgb(0x6e4322))
	rect(dst, b.X+b.W-28, b.Y-120, 8, 120, rgb(0x6e4322))
	roundRect(dst, b.X, b.Y-170, b.W, 90, 8, rgb(0x8a5a33))
	roundRect(dst, b.X+6, b.Y-164, b.W-12, 78, 6, rgb(0xf3e6cc))
	o := g.sim.Order
	if o != nil && o.Place() != g.sim.Cur.ID {
		o = nil // la commande est à livrer ailleurs
	}
	if o == nil {
		say(dst, "COMMANDES", b.C().X, b.Y-140, 18, true, 0, rgb(0x8a5a33))
		say(dst, "Pas de commande", b.C().X, b.Y-110, 15, false, 0, rgb(0x8a7a6a))
	} else {
		sayFit(dst, o.Who, b.C().X, b.Y-154, 14, true, 0, b.W-16, rgb(0x8a5a33))
		// les lignes, sur une ou deux rangées
		per := max(1, int((b.W-12)/52))
		rows := (len(o.Lines) + per - 1) / per
		for i, l := range o.Lines {
			k := min(per, len(o.Lines)-i/per*per)
			x := b.C().X + (float64(i%per)-float64(k-1)/2)*52
			y := b.Y - 122
			if rows > 1 {
				y = b.Y - 138 + float64(i/per)*24
			}
			drawItem(dst, l.It, x-14, y, 10)
			c := rgb(0x3b2a10)
			if l.Got >= l.N {
				c = rgb(0x2a8a3a)
			}
			say(dst, itoa(l.Got)+"/"+itoa(l.N), x+12, y, 12, true, 0, c)
		}
	}
	// la caisse
	ellipse(dst, b.C().X, b.Y+b.H, b.W*0.55, 12, rgba(0x000000, 0.2))
	roundRect(dst, b.X, b.Y, b.W, b.H, 8, rgb(0xb07d4a))
	for k := 0; k < 3; k++ {
		line(dst, b.X+6, b.Y+14+float64(k)*20, b.X+b.W-6, b.Y+14+float64(k)*20, 3, rgb(0x8a5a33))
	}
	say(dst, "COMMANDES", b.C().X, b.Y+b.H/2, 18, true, 0, rgb(0xfff8ea))
}

// drawOrderHUD : la commande en cours, sous l'argent (visible partout).
func (g *game) drawOrderHUD(dst *ebiten.Image, x, y float64) {
	o := g.sim.Order
	if o == nil {
		g.drawBoostHUD(dst, x, y)
		return
	}
	at := o.Place()
	if cur := g.sim.Cur; cur == nil || cur.ID != at {
		// ailleurs : une ligne discrète, qui dit où aller (et le cadeau, pour donner envie)
		lineW := 80.0
		w, h := math.Max(380, 26+float64(len(o.Lines))*lineW+60), 74.0
		if o.Gift != "" {
			w = math.Max(w, textWidth("Cadeau : "+o.GiftText(), 15, true)+70)
			h = 104
		}
		roundRect(dst, x, y, w, h, 16, rgba(0x2b2140, 0.75))
		if o.Gift != "" {
			drawGift(dst, x+24, y+86, 9)
			sayFit(dst, "Cadeau : "+o.GiftText(), x+40, y+87, 15, true, -1, w-50, rgb(0xffe066))
		}
		m, sec := int(o.Left)/60, int(o.Left)%60
		ts := itoa(m) + ":" + map[bool]string{true: "0", false: ""}[sec < 10] + itoa(sec)
		pc := busCols[at]
		roundRect(dst, x+10, y+6, w-20, 24, 8, rgb(pc))
		sayFit(dst, "Commande à livrer : "+busNames[at], x+w/2, y+18, 15, true, 0, w-36, rgb(0xffffff))
		for i, l := range o.Lines {
			lx := x + 26 + float64(i)*lineW
			drawItem(dst, l.It, lx, y+51, 11)
			say(dst, itoa(l.Got)+"/"+itoa(l.N), lx+16, y+51, 15, true, -1, rgb(0xffffff))
		}
		col := rgb(0xd8d0f0)
		if o.Left < 60 {
			col = rgb(0xff8a8a)
		}
		say(dst, ts, x+w-14, y+51, 17, true, 1, col)
		return
	}
	lineW := 118.0
	if len(o.Lines) > 3 {
		lineW = 96
	}
	w, h := math.Max(500, 30+float64(len(o.Lines))*lineW), 142.0
	if o.Gift != "" {
		w = math.Max(w, textWidth("Cadeau : "+o.GiftText(), 17, true)+70)
		h = 168
	}
	roundRect(dst, x, y, w, h, 18, rgba(0x2b2140, 0.82))
	m := int(o.Left) / 60
	sec := int(o.Left) % 60
	ts := itoa(m) + ":" + map[bool]string{true: "0", false: ""}[sec < 10] + itoa(sec)
	col := rgb(0xd8d0f0)
	if o.Left < 60 {
		col = rgb(0xff8a8a)
	}
	sayFit(dst, "Commande : "+o.Who, x+16, y+22, 18, true, -1, w-90, rgb(0xffe066))
	say(dst, ts, x+w-16, y+22, 20, true, 1, col)
	// gain, perte, et livraison express
	if o.Gift != "" {
		drawGift(dst, x+26, y+h-62, 10)
		sayFit(dst, "Cadeau : "+o.GiftText(), x+44, y+h-60, 17, true, -1, w-60, rgb(0xffe066))
		if fl := o.FastLeft(); fl > 0 {
			sayFit(dst, "Livrée avant "+clock(fl)+" : "+o.GiftExpress()+" !", x+16, y+h-30, 15, true, -1, w-140, rgb(0x7ed957))
		}
		say(dst, "Perte : 1", x+w-44, y+h-30, 15, true, 1, rgb(0xff8a8a))
		star(dst, x+w-28, y+h-31, 8, rgb(0xffd23f))
	} else {
		gain := "Gain : " + money(o.Reward)
		if fl := o.FastLeft(); fl > 0 {
			gain = "Gain : " + money(2*o.Reward) + " si livrée avant " + clock(fl) + " (sinon " + money(o.Reward) + ")"
		}
		sayFit(dst, gain, x+16, y+h-38, 15, true, -1, w-140, rgb(0x7ed957))
		say(dst, "Perte : 1", x+w-44, y+h-38, 15, true, 1, rgb(0xff8a8a))
		star(dst, x+w-28, y+h-39, 8, rgb(0xffd23f))
	}
	for i, l := range o.Lines {
		lx := x + 40 + float64(i)*lineW
		drawItem(dst, l.It, lx, y+60, 14)
		c := rgb(0xffffff)
		if l.Got >= l.N {
			c = rgb(0x7ed957)
		}
		say(dst, itoa(l.Got)+"/"+itoa(l.N), lx+22, y+60, 18, true, -1, c)
		sayFit(dst, sim.Items[l.It].Name, lx+20, y+86, 14, false, 0, lineW-6, rgb(0xd8d0f0))
	}
	// temps restant
	roundRect(dst, x+16, y+h-10, (w-32)*math.Max(0, o.Left/o.Total), 4, 2, rgba(0xffe066, 0.8))
}

// nextGoal : le prochain lieu à acheter (le moins cher).
func (g *game) nextGoal() *sim.Place {
	var best *sim.Place
	for _, p := range g.sim.Places {
		if !p.Owned && (best == nil || p.Price < best.Price) {
			best = p
		}
	}
	return best
}

func (g *game) drawGoal(dst *ebiten.Image, x, y float64) {
	p := g.nextGoal()
	if p == nil {
		return
	}
	right := 790.0 // (les boutons de droite : 790 de large, 862 avec la porte « Quitter »)
	if !isWeb {
		right += 72
	}
	w := math.Min(400, scr.VW-right-x-20)
	if w < 220 {
		return
	}
	h := 92.0
	roundRect(dst, x, y, w, h, 22, rgba(0x2b2140, 0.82))
	sayFit(dst, tr.F("Objectif : {lieu}", "lieu", tr.T(p.Name)), x+w/2, y+26, 20, true, 0, w-24, rgb(0xd8d0f0))
	price := g.sim.PriceOf(p)
	f := math.Min(1, g.sim.Money/price)
	roundRect(dst, x+16, y+48, w-32, 26, 13, rgba(0x000000, 0.4))
	c := uint32(0xd4a017)
	if f >= 1 {
		c = 0x2fa86b
	}
	roundRect(dst, x+18, y+50, (w-36)*f, 22, 11, rgb(c))
	txt := money(math.Min(g.sim.Money, price)) + " / " + money(price)
	if f >= 1 {
		txt = "Disponible sur la carte !"
	}
	sayFit(dst, txt, x+w/2, y+61, 16, true, 0, w-40, rgb(0xffffff))
}

// clock : 75 → « 1:15 ».
func clock(sec float64) string {
	m, s := int(sec)/60, int(sec)%60
	z := ""
	if s < 10 {
		z = "0"
	}
	return itoa(m) + ":" + z + itoa(s)
}

// drawGift : un petit paquet cadeau (centre x, y ; demi-taille r).
func drawGift(dst *ebiten.Image, x, y, r float64) {
	rect(dst, x-r, y-r*0.6, 2*r, r*1.6, rgb(0xe0457b))
	rect(dst, x-r*1.1, y-r*0.9, 2.2*r, r*0.5, rgb(0xf06a95))
	rect(dst, x-r*0.18, y-r*0.9, r*0.36, r*1.9, rgb(0xffd23f))
	circle(dst, x-r*0.35, y-r*1.05, r*0.32, rgb(0xffd23f))
	circle(dst, x+r*0.35, y-r*1.05, r*0.32, rgb(0xffd23f))
}

// drawBoostHUD : le cadeau en cours (à la place de la commande : un seul à la fois).
func (g *game) drawBoostHUD(dst *ebiten.Image, x, y float64) {
	b := g.sim.Boost
	if b == nil {
		return
	}
	txt := tr.F("Cadeau : {quoi}", "quoi", tr.T(b.Text()))
	col := uint32(0xffe066)
	if b.Malus() { // une pénalité (commande ratée) : en rouge
		txt = tr.F("Pénalité : {quoi}", "quoi", tr.T(b.Text()))
		col = 0xff8a7a
	}
	w, h := math.Max(360, textWidth(txt, 18, true)+150), 62.0
	roundRect(dst, x, y, w, h, 16, rgba(0x2b2140, 0.82))
	// un léger scintillement
	pulse := 0.5 + 0.5*math.Sin(g.t*4)
	roundRect(dst, x+2, y+2, w-4, h-4, 14, rgba(col, 0.08+0.08*pulse))
	if b.Malus() {
		circle(dst, x+30, y+28, 13, rgb(0xd9534f))
		say(dst, "!", x+30, y+29, 20, true, 0, rgb(0xffffff))
	} else {
		drawGift(dst, x+30, y+28, 12)
	}
	sayFit(dst, txt, x+54, y+26, 18, true, -1, w-140, rgb(col))
	say(dst, clock(b.Left), x+w-16, y+26, 20, true, 1, rgb(0xffffff))
	roundRect(dst, x+16, y+h-12, (w-32)*math.Max(0, b.Left/b.Total), 5, 2, rgba(col, 0.85))
}
