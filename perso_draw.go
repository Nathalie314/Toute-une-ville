package main

import (
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// Le personnage du joueur, avec sa tenue et ses équipements, et l'écran « Personnage ».

var shoeCols = []uint32{0x3a2a1a, 0x8a6a4a, 0xffffff, 0xe74c3c, 0x2fa86b, 0xff8c1a, 0x6b4423, 0x8e44ad, 0xf1c40f}

// bagCol : couleur du sac selon son niveau (jamais la case 0, vide).
func bagCol(l int) uint32 { return bagCols[1+(max(1, l)-1)%(len(bagCols)-1)] }

var bagCols = []uint32{0, 0xc0623a, 0x3b6ea8, 0x2fa86b, 0x8e44ad, 0x9c6b3c, 0x5a6878, 0xe67e22, 0x2c3e66}

func (g *game) heroLook() look {
	o := g.sim.OutfitNow()
	lk := look{skin: 0xf6d2b4, hair: 0x6b4423, shirt: o.Colors[0], pants: o.Colors[1], hatC: o.Colors[2], hat: 1}
	switch g.sim.WornIn("tete") {
	case "toque":
		lk.hat = 3
	case "chapeau":
		lk.hat = 9
	}
	if g.sim.Worn("toge") {
		lk.shirt, lk.pants = 0x24243a, 0x24243a
	}
	return lk
}

// drawHero dessine le joueur (en jeu ou dans les menus), avec ce qu'il porte.
func (g *game) drawHero(dst *ebiten.Image, pos, face V, walkT float64, moving bool, carry []sim.Item) {
	o := g.sim.OutfitNow()
	lk := g.heroLook()
	x, y := pos.X, pos.Y
	bag := 0
	if g.sim.Worn("sac") {
		bag = g.sim.GearLevel("sac")
	}
	toge := g.sim.Worn("toge")
	drawBag := func() {
		if bag == 0 {
			return
		}
		s := 0.7 + 0.06*float64(bag)
		bx := x - face.X*14
		roundRect(dst, bx-15*s, y-62-8*s, 30*s, 40*s, 8*s, rgb(bagCol(bag)))
		roundRect(dst, bx-10*s, y-40-4*s, 20*s, 14*s, 5*s, shade(bagCol(bag), 0.8))
	}
	if face.Y >= -0.5 {
		drawBag() // on voit le sac dépasser derrière
	}
	drawPerson(dst, pos, face, walkT, moving, lk, nil)
	if face.Y < -0.5 {
		drawBag() // de dos : le sac est devant nous
	} else if bag > 0 {
		// de face : les bretelles du sac
		c := shade(bagCol(bag), 0.85)
		line(dst, x-12, y-58, x-10, y-30, 5, c)
		line(dst, x+12, y-58, x+10, y-30, 5, c)
	}
	if toge {
		// la toge descend sous les genoux
		polygon(dst, []V{{X: x - 17, Y: y - 40}, {X: x + 17, Y: y - 40}, {X: x + 22, Y: y - 12}, {X: x - 22, Y: y - 12}}, rgb(0x24243a))
		if face.Y >= -0.5 {
			// l'étole dorée
			line(dst, x-8, y-58, x-9, y-16, 5, rgb(0xf1c40f))
			line(dst, x+8, y-58, x+9, y-16, 5, rgb(0xf1c40f))
		}
	}
	sw := 0.0
	if moving {
		sw = math.Sin(walkT * 12)
	}
	// chaussures
	switch g.sim.WornIn("pieds") {
	case "chaussures":
		l := g.sim.GearLevel("chaussures")
		c := rgb(shoeCols[l%len(shoeCols)])
		roundRect(dst, x-14, y-6+sw*3, 14, 9, 4, c)
		roundRect(dst, x, y-6-sw*3, 14, 9, 4, c)
		if l >= 8 { // chaussures ailées
			polygon(dst, []V{{X: x - 14, Y: y - 4}, {X: x - 26, Y: y - 14}, {X: x - 16, Y: y - 10}}, rgb(0xffffff))
			polygon(dst, []V{{X: x + 14, Y: y - 4}, {X: x + 26, Y: y - 14}, {X: x + 16, Y: y - 10}}, rgb(0xffffff))
		}
	case "mocassins":
		c := rgb(0x7a4220)
		roundRect(dst, x-15, y-6+sw*3, 15, 9, 4, c)
		roundRect(dst, x, y-6-sw*3, 15, 9, 4, c)
		circle(dst, x-8, y-4+sw*3, 2, rgb(0xf1c40f))
		circle(dst, x+7, y-4-sw*3, 2, rgb(0xf1c40f))
	}
	if !toge && o.ID == "fermier" {
		line(dst, x-9, y-56, x-9, y-26, 4, rgb(o.Colors[1]))
		line(dst, x+9, y-56, x+9, y-26, 4, rgb(o.Colors[1]))
	}
	// objets en main (quand il ne porte rien)
	if len(carry) == 0 {
		for i, id := range g.sim.Held() {
			hx := x + 26
			if i == 1 {
				hx = x - 26
			}
			drawHandItem(dst, id, hx, y-32)
		}
	}
	for i, it := range carry {
		drawItem(dst, it, x, y-40-float64(i)*14, 14)
	}
}

// drawHandItem : arrosoir, sifflet ou mégaphone.
func drawHandItem(dst *ebiten.Image, id string, x, y float64) {
	switch id {
	case "arrosoir":
		roundRect(dst, x-10, y-8, 20, 18, 4, rgb(0x4a90d9))
		line(dst, x+8, y-2, x+20, y-12, 4, rgb(0x4a90d9))
		circle(dst, x+21, y-13, 4, rgb(0x3a7ac0))
		ring(dst, x-10, y-2, 6, 3, rgb(0x3a7ac0))
	case "sifflet":
		roundRect(dst, x-8, y-5, 16, 10, 5, rgb(0xd0d0d0))
		circle(dst, x-4, y, 5, rgb(0xd0d0d0))
		line(dst, x+2, y-5, x-6, y-22, 2, rgb(0xe74c3c))
	case "megaphone":
		polygon(dst, []V{{X: x - 8, Y: y - 4}, {X: x + 14, Y: y - 12}, {X: x + 14, Y: y + 12}, {X: x - 8, Y: y + 4}}, rgb(0xf5f5f5))
		rect(dst, x+12, y-12, 4, 24, rgb(0xe74c3c))
		rect(dst, x-12, y-4, 6, 12, rgb(0x555555))
	case "livre":
		// un livre ouvert
		polygon(dst, []V{{X: x - 14, Y: y - 8}, {X: x, Y: y - 4}, {X: x, Y: y + 10}, {X: x - 14, Y: y + 6}}, rgb(0x2f6aa8))
		polygon(dst, []V{{X: x + 14, Y: y - 8}, {X: x, Y: y - 4}, {X: x, Y: y + 10}, {X: x + 14, Y: y + 6}}, rgb(0x2f6aa8))
		polygon(dst, []V{{X: x - 12, Y: y - 8}, {X: x, Y: y - 5}, {X: x, Y: y + 7}, {X: x - 12, Y: y + 3}}, rgb(0xfdf8ea))
		polygon(dst, []V{{X: x + 12, Y: y - 8}, {X: x, Y: y - 5}, {X: x, Y: y + 7}, {X: x + 12, Y: y + 3}}, rgb(0xfdf8ea))
		line(dst, x-9, y-3, x-3, y-1, 1, rgb(0x9a9a9a))
		line(dst, x+3, y-1, x+9, y-3, 1, rgb(0x9a9a9a))
	}
}

func (g *game) drawPlayer(dst *ebiten.Image, a *sim.Agent) {
	// rayons d'effet des objets
	for _, id := range append([]string{"toque"}, g.sim.Held()...) {
		if r := g.sim.GearRadius(id); r > 0 && g.sim.Worn(id) {
			col := map[string]uint32{"toque": 0xffffff, "arrosoir": 0x4fc3ff, "sifflet": 0xffd23f}[id]
			ring(dst, a.Pos.X, a.Pos.Y, r, 3, rgba(col, 0.25))
		}
	}
	ring(dst, a.Pos.X, a.Pos.Y, 30, 4, rgba(0x4fc3ff, 0.8))
	if c := g.sim.Cur; c != nil && c.ID == "parc" { // au parc, on roule en trottinette
		g.drawHero(dst, a.Pos.Add(V{Y: -12}), a.Face, 0, false, a.Carry)
		drawScooter(dst, a.Pos, a.Face, 1.15, 0)
		return
	}
	g.drawHero(dst, a.Pos, a.Face, a.WalkT, a.Moving, a.Carry)
}

func (g *game) drawPlayerIcon(dst *ebiten.Image, a *sim.Agent) {
	g.drawHero(dst, a.Pos, a.Face, 0, false, nil)
}

// ---------------------------------------------------------------- écran Personnage
//
// À gauche, le bonhomme et une case par zone (même couleur que sa catégorie) :
// on y voit l'objet porté. À droite, les quatre catégories : un clic sur l'image
// d'un objet le fait porter (halo), « Augmenter » achète le niveau suivant.
// Mains : un objet au début ; la deuxième case s'achète (Ambidextre) en restant
// appuyé dessus.

var slotNames = map[string]string{"tete": "Tête", "main": "Mains", "dos": "Dos", "pieds": "Pieds"}
var slotCols = map[string]uint32{"tete": 0xf6d9b0, "dos": 0xcfe2f6, "main": 0xd6efc6, "pieds": 0xe6d6f3}
var slotInk = map[string]uint32{"tete": 0x9a5a1a, "dos": 0x2f5f8f, "main": 0x3f7a2a, "pieds": 0x6a3f8f}

func (g *game) canBuyGear() bool {
	for _, gr := range sim.Gears {
		l := g.sim.GearLevel(gr.ID)
		if l < gr.Max() && g.sim.Money >= gr.Costs[l] {
			return true
		}
	}
	return false
}

func (g *game) slotHot(slot string) bool {
	for _, gr := range sim.SlotGears(slot) {
		l := g.sim.GearLevel(gr.ID)
		if l < gr.Max() && g.sim.Money >= gr.Costs[l] {
			return true
		}
	}
	return false
}

func (g *game) drawPersoPanel(dst *ebiten.Image) {
	w, h := math.Min(1560, scr.VW-60), 980.0
	x, y := g.panelFrame(dst, "Personnage", w, h)
	// --- à gauche : le bonhomme et ses cases
	lw := 600.0
	roundRect(dst, x+30, y+20, lw-40, h-140, 24, rgb(0xf3e6cc))
	cx, cy := x+30+(lw-40)/2, y+20+(h-140)/2
	scaled(cx, cy, 3.2, func() {
		g.drawHero(dst, V{X: cx, Y: cy + 40}, V{Y: 1}, 0, false, nil)
	})
	const bw, bh = 170.0, 124.0
	box := func(slot string, bx, by float64, id string) {
		roundRect(dst, bx, by+4, bw, bh, 18, rgba(0x000000, 0.12))
		roundRect(dst, bx, by, bw, bh, 18, rgb(slotCols[slot]))
		roundStroke(dst, bx, by, bw, bh, 18, 3, shade(slotCols[slot], 0.8))
		say(dst, slotNames[slot], bx+bw/2, by+18, 18, true, 0, rgb(slotInk[slot]))
		if id == "" {
			say(dst, "Rien", bx+bw/2, by+64, 20, false, 0, rgba(0x5a4a3a, 0.7))
			return
		}
		gearIcon(dst, id, bx+bw/2, by+60, max(1, g.sim.GearLevel(id)))
		sayFit(dst, g.sim.GearName(id), bx+bw/2, by+bh-18, 16, true, 0, bw-14, rgb(0x3b2a10))
	}
	held := g.sim.Held()
	hand := func(i int) string {
		if i < len(held) {
			return held[i]
		}
		return ""
	}
	box("tete", cx-bw/2, y+36, g.sim.WornIn("tete"))
	box("main", x+40, cy-10, hand(0))
	box("dos", x+lw-bw-40, cy-240, g.sim.WornIn("dos"))
	box("pieds", cx-bw/2, y+h-276, g.sim.WornIn("pieds"))
	// deuxième main : à débloquer (Ambidextre), en restant appuyé dessus
	mx, my := x+lw-bw-40, cy+60
	if g.sim.Hands() > 1 {
		box("main", mx, my, hand(1))
	} else {
		u := sim.Ambidextre
		cost := u.Cost(0)
		ok := g.sim.Money >= cost
		roundRect(dst, mx, my+4, bw, bh, 18, rgba(0x000000, 0.12))
		roundRect(dst, mx, my, bw, bh, 18, rgba(slotCols["main"], 0.55))
		dashedRect(dst, mx+3, my+3, bw-6, bh-6, 4, 12, 0, shade(slotCols["main"], 0.6))
		say(dst, "Ambidextre", mx+bw/2, my+22, 19, true, 0, rgb(slotInk["main"]))
		say(dst, "2e main", mx+bw/2, my+46, 15, false, 0, rgb(0x5a4a3a))
		pricePill(dst, money(cost), mx+bw/2, my+74, bw-30)
		pressed := g.hover(mx, my, bw, bh) && (pointerPressed() || g.hand.on)
		if ok && pressed {
			g.ambiT += 1.0 / 60
		} else {
			g.ambiT = 0
		}
		if g.ambiT > 0 {
			f := math.Min(1, g.ambiT/1.2)
			roundRect(dst, mx, my+bh*(1-f), bw, bh*f, 18, rgba(0x2fa86b, 0.4))
		}
		if g.ambiT >= 1.2 {
			g.ambiT = 0
			if g.sim.Buy(u) {
				g.snd.play(sBuy)
				g.toast("Ambidextre : deux objets en main !", "")
			}
		}
		txt := "Reste appuyé pour acheter"
		if !ok {
			txt = "Pas encore assez d'argent"
		}
		sayFit(dst, txt, mx+bw/2, my+bh-14, 13, false, 0, bw-10, rgb(0x5a4a3a))
		if g.hand.on {
			g.addButton(mx, my, bw, bh, func() {}) // la main peut s'y poser (elle reste appuyée)
		}
	}
	say(dst, "Clique sur un objet pour le porter (ou l'enlever).", cx, y+h-138, 17, false, 0, rgb(0x5a4a3a))

	// --- à droite : les quatre catégories
	rx := x + lw
	rw := w - lw - 30
	yy := y + 20
	const lh, hh = 62.0, 42.0
	for _, slot := range sim.Slots {
		gears := sim.SlotGears(slot)
		fh := hh + float64(len(gears))*lh + 8
		roundRect(dst, rx, yy, rw, fh, 18, rgb(slotCols[slot]))
		say(dst, slotNames[slot], rx+24, yy+22, 22, true, -1, rgb(slotInk[slot]))
		hint := "Un seul objet porté à la fois."
		if slot == "main" {
			hint = "Un objet en main (deux avec Ambidextre)."
		}
		say(dst, hint, rx+rw-24, yy+22, 16, false, 1, shade(slotInk[slot], 1.1))
		ly := yy + hh
		for _, gr := range gears {
			gr := gr
			l := g.sim.GearLevel(gr.ID)
			worn := g.sim.Worn(gr.ID)
			// l'image : on clique dessus pour porter l'objet
			ix, iy := rx+16, ly+3
			if worn {
				circle(dst, ix+34, iy+28, 34, rgba(0xffe066, 0.55+0.2*math.Sin(g.t*4)))
				ring(dst, ix+34, iy+28, 32, 4, rgb(0xf1c40f))
			} else {
				roundRect(dst, ix+4, iy, 60, 56, 14, rgba(0xffffff, 0.65))
			}
			scaled(ix+34, iy+28, 0.62, func() { gearIcon(dst, gr.ID, ix+34, iy+28, max(1, l)) })
			if l > 0 {
				g.addButton(ix, iy, 68, 56, func() {
					g.sim.Wear(gr.ID)
					g.snd.play(sClick)
				})
			} else {
				roundRect(dst, ix+4, iy, 60, 56, 14, rgba(0x000000, 0.25))
			}
			// nom et pouvoir actuel
			name := gr.Names[0]
			pow := "Pas encore acheté. Niveau 1 : " + lowerFirst(gr.Desc(1))
			if l > 0 {
				name = gr.Names[l-1] + "  (niv. " + itoa(l) + "/" + itoa(gr.Max()) + ")"
				pow = gr.Desc(l)
				if !worn {
					pow = "Rangé : " + lowerFirst(pow)
				}
			}
			bwid := 240.0
			sayFit(dst, name, rx+100, ly+18, 20, true, -1, rw-bwid-160, rgb(0x3b2a10))
			sayFit(dst, pow, rx+100, ly+42, 15, false, -1, rw-bwid-160, rgb(0x5a4a3a))
			// augmenter
			bx := rx + rw - bwid - 16
			if l < gr.Max() {
				c := gr.Costs[l]
				lbl := "Augmenter · " + money(c)
				if l == 0 {
					lbl = "Acheter · " + money(c)
				}
				g.btn(dst, lbl, bx, ly+8, bwid, 44, 0x2fa86b, g.sim.Money >= c, func() {
					if g.sim.BuyGear(gr) {
						g.snd.play(sBuy)
					}
				})
			} else {
				roundRect(dst, bx, ly+8, bwid, 44, 13, rgba(0x000000, 0.12))
				say(dst, "Niveau max", bx+bwid/2, ly+30, 18, true, 0, rgb(0x5a4a3a))
			}
			ly += lh
		}
		yy += fh + 10
	}
}

// lowerFirst : première lettre en minuscule (« Tu marches… » → « tu marches… »).
func lowerFirst(s string) string {
	for _, p := range []string{"Tu ", "Les ", "Dans ", "En "} {
		if strings.HasPrefix(s, p) {
			return strings.ToLower(p[:1]) + s[1:]
		}
	}
	return s
}

// gearIcon : petite image d'un équipement.
func gearIcon(dst *ebiten.Image, id string, x, y float64, l int) {
	switch id {
	case "chaussures":
		c := rgb(shoeCols[l%len(shoeCols)])
		roundRect(dst, x-26, y-4, 24, 16, 6, c)
		roundRect(dst, x+2, y-4, 24, 16, 6, c)
		rect(dst, x-26, y+8, 24, 5, rgb(0x3a2a1a))
		rect(dst, x+2, y+8, 24, 5, rgb(0x3a2a1a))
	case "sac":
		c := bagCol(l)
		roundRect(dst, x-22, y-26, 44, 56, 12, rgb(c))
		roundRect(dst, x-15, y, 30, 20, 6, shade(c, 0.8))
		line(dst, x-12, y-26, x-12, y-34, 5, rgb(0x6b3a1a))
		line(dst, x+12, y-26, x+12, y-34, 5, rgb(0x6b3a1a))
	case "toque":
		scaled(x, y, 1.6, func() {
			rect(dst, x-11, y-6, 22, 16, rgb(0xffffff))
			circle(dst, x-8, y-8, 9, rgb(0xffffff))
			circle(dst, x+8, y-8, 9, rgb(0xffffff))
			circle(dst, x, y-12, 10, rgb(0xffffff))
			rect(dst, x-11, y+8, 22, 4, rgb(0xdddddd))
		})
	case "chapeau":
		scaled(x, y+6, 1.6, func() { drawMortarboard(dst, x, y+6, 0x24243a) })
	case "toge":
		polygon(dst, []V{{X: x - 16, Y: y - 26}, {X: x + 16, Y: y - 26}, {X: x + 26, Y: y + 28}, {X: x - 26, Y: y + 28}}, rgb(0x24243a))
		line(dst, x-8, y-26, x-10, y+24, 5, rgb(0xf1c40f))
		line(dst, x+8, y-26, x+10, y+24, 5, rgb(0xf1c40f))
		ellipse(dst, x, y-26, 12, 5, rgb(0x3a3a55))
	case "mocassins":
		roundRect(dst, x-28, y-4, 26, 16, 7, rgb(0x7a4220))
		roundRect(dst, x+2, y-4, 26, 16, 7, rgb(0x7a4220))
		circle(dst, x-14, y, 3, rgb(0xf1c40f))
		circle(dst, x+16, y, 3, rgb(0xf1c40f))
	default:
		scaled(x, y, 1.8, func() { drawHandItem(dst, id, x-4, y) })
	}
}

// drawMortarboard : le chapeau carré du diplômé (x, y : haut de la tête).
func drawMortarboard(dst *ebiten.Image, x, y float64, c uint32) {
	roundRect(dst, x-12, y-8, 24, 10, 3, rgb(c))
	polygon(dst, []V{{X: x - 24, Y: y - 12}, {X: x, Y: y - 20}, {X: x + 24, Y: y - 12}, {X: x, Y: y - 4}}, rgb(c))
	line(dst, x, y-12, x+18, y-8, 2, rgb(0xf1c40f))
	line(dst, x+18, y-8, x+19, y+6, 2, rgb(0xf1c40f))
	circle(dst, x+19, y+7, 3, rgb(0xf1c40f))
}
