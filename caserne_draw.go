package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
	"ville/tr"
)

// La caserne de pompiers (le commandant en haut au centre, la prévention à gauche, le garage des
// camions à droite) ; et, dans un lieu qui brûle, les flammes, la fumée, les camions qui arrosent.

const fireRed = 0xd62d20

// firemanLook : un pompier (veste rouge, casque jaune).
func firemanLook(n int) look {
	lk := lookFrom(900 + n*31)
	lk.shirt, lk.pants, lk.hat, lk.hatC = fireRed, 0x2b2b33, 2, 0xf5c518
	lk.skirt = 0
	return lk
}

func (g *game) caserneGround(dst *ebiten.Image, p *sim.Place, tl, br V) {
	gs := g.sim
	road(dst, tl, br, 1640)
	bx, by, bw, bh := sim.CasBldX, sim.CasBldY, sim.CasBldW, sim.CasBldH
	door := V{X: sim.CasHall().C().X, Y: by + bh}
	rect(dst, door.X-90, door.Y, 180, 1624-door.Y, rgb(0xd6d2c8))
	gar := sim.CasGarage()
	rect(dst, gar.X, by+bh, gar.W, 1624-by-bh, rgb(0x6a6e73)) // (la sortie des camions, jusqu'à la route)
	slab(dst, bx, by, bw, bh, 8, 16, 0xa0453a)
	c, h := sim.CasChief(), sim.CasHall()
	rect(dst, c.X, c.Y, c.W, c.H, rgb(0xc9a477))
	for y := c.Y; y < c.Y+c.H; y += 36 {
		rect(dst, c.X, y, c.W, 2, rgba(0x8a6a44, 0.35))
	}
	rect(dst, h.X, h.Y, h.W, h.H, rgb(0xe8ecf0))
	say(dst, "COMMANDANT", c.C().X, c.Y+24, 20, true, 0, rgba(0x4a3a22, 0.75))
	say(dst, "HALL", h.C().X, h.Y+24, 20, true, 0, rgba(0x2b3a5a, 0.7))
	// la prévention : un couloir, six bureaux
	cy := sim.CasCorridorY()
	o1 := sim.CasOffice(1)
	left := sim.CasOffice(sim.CasOffN)
	rect(dst, left.X-10, cy+20, o1.X+o1.W-left.X+sim.CasGap+10, sim.CasCorH-40, rgb(0xd9d6ce))
	say(dst, "PRÉVENTION", (left.X+o1.X+o1.W)/2, by+28, 20, true, 0, rgba(0xffffff, 0.85))
	for n := 1; n <= sim.CasOffN; n++ {
		r := sim.CasOffice(n)
		col := rgb(0x9aa8bd)
		if gs.Levels["caserne.bureau_"+itoa(n)] == 0 {
			col = rgba(0x9aa8bd, 0.35)
		}
		rect(dst, r.X, r.Y, r.W, r.H, col)
	}
	// le garage
	rect(dst, gar.X, gar.Y, gar.W, gar.H, rgb(0x8a8e93))
	say(dst, "ACTION", gar.C().X, by+28, 20, true, 0, rgba(0xffffff, 0.85))
	for k := 1; k <= sim.CasTruckN; k++ {
		b := sim.CasBay(k)
		rect(dst, b.X-4, b.Y, 6, b.H, rgb(0xf5c518))
		rect(dst, b.X+b.W-2, b.Y, 6, b.H, rgb(0xf5c518))
		rect(dst, b.X, by+bh-14, b.W, 14, rgb(0x5a5e63)) // (la porte du garage)
	}
	// les murs (avec leurs portes, sur le couloir et sur le garage)
	wall := rgb(0x4a505a)
	for _, x := range []float64{c.X - sim.CasGap/2, c.X + c.W + sim.CasGap/2} {
		rect(dst, x-14, by+40, 28, cy+20-by-40, wall)
		rect(dst, x-14, cy+sim.CasCorH-20, 28, by+bh-40-(cy+sim.CasCorH-20), wall)
	}
	rect(dst, c.X, c.Y+c.H+6, c.W/2-70, 10, wall)
	rect(dst, c.X+c.W/2+70, c.Y+c.H+6, c.W/2-70, 10, wall)
	// les cases du hall : la prévention (les PPI), les camions
	if gs.HasCommander() {
		for i, r := range []sim.Rect{sim.CasPrevZone(), sim.CasTruckZone()} {
			in := gs.Cur == p && r.Has(gs.Player.Pos)
			zoneBox(dst, r, in, g.t, 0x9fd3ff)
			say(dst, []string{"PRÉVENTION", "CAMIONS"}[i], r.C().X, r.C().Y-14, 18, true, 0, rgb(0x2b3a5a))
			say(dst, "Reste ici", r.C().X, r.C().Y+18, 15, false, 0, rgba(0x2b3a5a, 0.85))
		}
	}
	tw := textWidth("CASERNE DE POMPIERS", 24, true) + 60
	roundRect(dst, door.X-tw/2, door.Y+26, tw, 42, 10, rgb(fireRed))
	say(dst, "CASERNE DE POMPIERS", door.X, door.Y+47, 24, true, 0, rgb(0xffffff))
}

func (g *game) caserneDrawables(dst *ebiten.Image, p *sim.Place, ds []drawable) []drawable {
	gs := g.sim
	t := g.t
	add := func(x, y float64, fn func()) { ds = append(ds, drawable{x: x, y: y, fn: fn}) }
	// le commandant
	c := sim.CasChief()
	dx, dy := c.C().X-60, c.Y+170
	add(dx, dy+70, func() {
		roundRect(dst, dx-150, dy, 300, 70, 8, rgb(0x6e4322))
		roundRect(dst, dx-110, dy+16, 70, 44, 4, rgb(0xf3ead8))
	})
	if gs.HasCommander() {
		add(dx, dy-5, func() {
			lk := firemanLook(1)
			lk.shirt, lk.vest = 0x1c2a44, fireRed
			drawPerson(dst, V{X: dx, Y: dy - 12}, V{Y: 1}, 0, false, lk, nil)
		})
	}
	// les bureaux de prévention : un ordinateur, un téléphone, une imprimante ; leur chargé
	for n := 1; n <= sim.CasOffN; n++ {
		r := sim.CasOffice(n)
		if gs.Levels["caserne.bureau_"+itoa(n)] == 0 {
			continue
		}
		x, y := r.C().X, r.Y+120
		add(x, y+60, func() {
			roundRect(dst, x-120, y, 240, 60, 6, rgb(0x6e4322))
			roundRect(dst, x-30, y-30, 60, 40, 4, rgb(0x2b2b2b)) // l'ordinateur
			rect(dst, x-25, y-26, 50, 30, rgb(0x5b9bd5))
			roundRect(dst, x-100, y+12, 40, 26, 6, rgb(0x2b2b2b)) // le téléphone
			roundRect(dst, x+60, y+8, 50, 36, 4, rgb(0xd8d8d8))   // l'imprimante
			rect(dst, x+66, y+4, 38, 8, rgb(0xffffff))
		})
		if gs.Levels["caserne.charge_"+itoa(n)] > 0 {
			n := n
			add(x, y-5, func() {
				lk := lookFrom(1300 + n*17)
				lk.shirt = 0xe8833a
				drawPerson(dst, V{X: x, Y: y - 12}, V{Y: 1}, 0, false, lk, nil)
			})
		}
	}
	// les camions au garage (partis au feu : la place est vide) ; leur pompier à côté
	away := gs.Theft != nil && gs.Theft.Fire
	for k := 1; k <= sim.CasTruckN; k++ {
		if gs.Levels["caserne.camion_"+itoa(k)] == 0 {
			continue
		}
		b := sim.CasBay(k)
		cc := b.C()
		if away && gs.TruckReady(k) {
			add(cc.X, cc.Y, func() { say(dst, "Au feu !", cc.X, cc.Y, 22, true, 0, rgb(0xffffff)) })
			continue
		}
		add(cc.X, cc.Y+150, func() { drawFireTruckTop(dst, cc.X, cc.Y, false, t) })
		if gs.Levels["caserne.pompier_"+itoa(k)] > 0 {
			v := V{X: b.X + b.W - 20, Y: b.Y + b.H + 60}
			k := k
			add(v.X, v.Y, func() { drawPerson(dst, v, V{Y: 1}, 0, false, firemanLook(k), nil) })
		}
	}
	return ds
}

// drawFireTruckTop : un camion de pompiers vu de dessus (l'avant en bas : il sort vers la route).
func drawFireTruckTop(dst *ebiten.Image, x, y float64, siren bool, t float64) {
	ellipse(dst, x+6, y+10, 80, 160, rgba(0x000000, 0.22))
	roundRect(dst, x-70, y-150, 140, 300, 18, rgb(fireRed))
	roundRect(dst, x-56, y+90, 112, 40, 8, rgb(0x2b3a4a)) // le pare-brise
	rect(dst, x-14, y-140, 28, 220, rgb(0xb0b0b0))        // l'échelle
	for k := 0; k < 8; k++ {
		rect(dst, x-14, y-130+float64(k)*26, 28, 4, rgb(0x7a7a7a))
	}
	on := siren && math.Mod(t*3, 1) < 0.5
	c := uint32(0x2b6cff)
	if on {
		c = 0x9fc0ff
	}
	roundRect(dst, x-40, y+70, 80, 14, 4, rgb(c))
}

// drawFireTruckSide : le camion de profil, sur la route (fx : 1 vers la droite).
func drawFireTruckSide(dst *ebiten.Image, x, y, fx, t float64) {
	ellipse(dst, x, y+2, 110, 12, rgba(0x000000, 0.25))
	roundRect(dst, x-108, y-90, 216, 80, 10, rgb(fireRed))
	roundRect(dst, x+fx*70-26, y-84, 50, 34, 6, rgb(0xbfe3f5))
	rect(dst, x-100, y-102, 150, 10, rgb(0xb0b0b0)) // l'échelle
	on := math.Mod(t*3, 1) < 0.5
	c := uint32(0x2b6cff)
	if on {
		c = 0x9fc0ff
	}
	roundRect(dst, x+fx*70-14, y-100, 28, 10, 3, rgb(c))
	for _, wx := range []float64{-70, 0, 70} {
		circle(dst, x+wx, y-8, 15, rgb(0x2b2b2b))
		circle(dst, x+wx, y-8, 6, rgb(0xb0b0b0))
	}
}

// fireDrawables : un lieu qui brûle : des flammes sur ses machines, de la fumée ; les camions
// des pompiers arrivent et arrosent.
func (g *game) fireDrawables(dst *ebiten.Image, p *sim.Place, ds []drawable) []drawable {
	th := g.sim.Theft
	if th == nil || !th.Fire || th.Place != p.ID {
		return ds
	}
	t := g.t
	el := th.Total - th.Left
	k := 0
	for _, s := range p.Stations {
		if !s.Built || s.Kind == sim.Field || (p.Yard.W > 0 && p.Yard.Has(s.Body.C())) {
			continue // (la zone du joueur ne brûle pas)
		}
		k++
		c := s.Body.C()
		phase := float64(k) * 1.7
		ds = append(ds, drawable{x: c.X, y: s.Body.Y + s.Body.H + 5, fn: func() { drawFlames(dst, c.X, s.Body.Y+20, 1+0.3*math.Sin(phase), t+phase) }})
	}
	// les camions : ils arrivent au bout de 8 secondes, devant le portail, et arrosent
	n := g.sim.TrucksReady()
	if n > 0 && el > 8 {
		for i := 0; i < n; i++ {
			x := p.Gate.X - float64(n-1)*130 + float64(i)*260
			y := 1770.0
			ds = append(ds, drawable{x: x, y: y + 20, fn: func() {
				drawFireTruckSide(dst, x, y, 1, t)
				for j := 0; j < 6; j++ { // l'eau
					f := math.Mod(t*1.4+float64(j)/6, 1)
					circle(dst, x+20+f*20, y-110-f*260, 6+f*6, rgba(0x6fb7e6, 0.8-0.6*f))
				}
			}})
		}
	}
	return ds
}

// drawFlames : des flammes qui dansent.
func drawFlames(dst *ebiten.Image, x, y, k, t float64) {
	for i := 0; i < 5; i++ {
		f := math.Mod(t*1.8+float64(i)*0.23, 1)
		sx := x + math.Sin(t*5+float64(i)*2)*12*k
		h := (40 + 30*math.Sin(t*7+float64(i))) * k
		ellipse(dst, sx+float64(i-2)*16*k, y-h*0.5, 16*k, h, rgba(0xff8c1a, 0.85))
		ellipse(dst, sx+float64(i-2)*16*k, y-h*0.35, 9*k, h*0.6, rgba(0xffe066, 0.9))
		circle(dst, sx, y-h-f*120, (10+f*24)*k, rgba(0x4a4a4a, 0.45*(1-f))) // la fumée
	}
}

// caserneHUD : les cases du hall ouvrent le menu (la prévention, les camions).
func (g *game) caserneHUD(dst *ebiten.Image) {
	p := g.sim.Cur
	if p == nil || p.Cas == nil || !g.sim.HasCommander() {
		return
	}
	pl := g.sim.Player
	a, b := sim.CasPrevZone().Has(pl.Pos), sim.CasTruckZone().Has(pl.Pos)
	if (a || b) && !pl.Moving && !g.casZoneIn && g.panel == panelNone {
		g.casZoneIn = true
		g.casTab = 0
		if b {
			g.casTab = 1
		}
		g.openPanel(panelCaserne)
	}
	if !a && !b {
		g.casZoneIn = false
	}
}

// drawCasernePanel : deux onglets : la prévention (répartir les PPI), les camions (améliorations).
func (g *game) drawCasernePanel(dst *ebiten.Image) {
	gs := g.sim
	w, h := math.Min(1400, scr.VW-60), 900.0
	x, y := g.panelFrame(dst, "Caserne de pompiers", w, h)
	for i, lbl := range []string{"Prévention", "Camions"} {
		i := i
		col := uint32(0x8a9ab8)
		if g.casTab == i {
			col = fireRed
		}
		g.btn(dst, lbl, x+w/2-370+float64(i)*380, y+14, 360, 60, col, true, func() { g.casTab = i })
	}
	y += 90
	if g.casTab == 0 {
		txt := tr.F("PPI à répartir : {n}.", "n", itoa(gs.PPI()))
		if n := gs.PrevOfficers(); n > 0 {
			txt += " " + tr.F("Prochain PPI dans {duree} ({n} chargé(s) de prévention).", "duree", clock(gs.PPINext()), "n", itoa(n))
		}
		tr.Seen(txt)
		sayFit(dst, txt, x+w/2, y, 20, true, 0, w-60, rgb(0x2b3a5a))
		say(dst, "Chaque lieu commence à 100 % de risque : 10 PPI le baissent de 1 % (10 % au moins).", x+w/2, y+34, 18, false, 0, rgb(0x5a4a3a))
		i := 0
		for _, id := range sim.HotPlaces() {
			q := gs.Place(id)
			if q == nil || !q.Owned {
				continue
			}
			id := id
			ry := y + 80 + float64(i)*96
			i++
			roundRect(dst, x+30, ry, w-60, 84, 16, rgb(0xf1e6d2))
			say(dst, mapNames[id], x+60, ry+26, 22, true, -1, rgb(0x2b3a5a))
			risk := int(math.Round(100 * gs.FireRisk(id)))
			sub := tr.F("Risque d'incendie : {p} % ({z} zone(s) chaude(s), {n} PPI)", "p", itoa(risk), "z", itoa(gs.HotZones(id)), "n", itoa(gs.PPIOf(id)))
			tr.Seen(sub)
			say(dst, sub, x+60, ry+58, 17, false, -1, rgb(0x5a4a3a))
			for j, n := range []int{10, 100} {
				n := n
				ok := gs.PPI() > 0 && gs.PPIOf(id) < sim.PPIMax
				g.btn(dst, tr.F("+{n} PPI", "n", itoa(n)), x+w-30-float64(2-j)*190, ry+14, 170, 56, 0x2fa86b, ok, func() {
					if gs.SpendPPI(id, n) {
						g.snd.play(sBuy)
					}
				})
			}
		}
		if i == 0 {
			say(dst, "Aucun lieu à risque pour l'instant.", x+w/2, y+160, 20, false, 0, rgb(0x5a4a3a))
		}
		return
	}
	txt := tr.F("Ensemble, les camions sauvent {n} élément(s) d'un lieu en feu (le plus grand en compte {m}).", "n", itoa(gs.SaveCapacity()), "m", itoa(gs.MaxFireElems()))
	tr.Seen(txt)
	sayFit(dst, txt, x+w/2, y, 20, true, 0, w-60, rgb(0x2b3a5a))
	say(dst, "D'abord la caissière, puis les employés, puis les objets du plus cher au moins cher.", x+w/2, y+34, 18, false, 0, rgb(0x5a4a3a))
	for k := 1; k <= sim.CasTruckN; k++ {
		k := k
		ry := y + 80 + float64(k-1)*110
		roundRect(dst, x+30, ry, w-60, 96, 16, rgb(0xf1e6d2))
		say(dst, tr.F("Camion n° {n}", "n", itoa(k)), x+60, ry+28, 22, true, -1, rgb(0x2b3a5a))
		var sub string
		switch {
		case gs.Levels["caserne.camion_"+itoa(k)] == 0:
			sub = "Pas encore acheté (au garage)"
		case gs.Levels["caserne.pompier_"+itoa(k)] == 0:
			sub = "Il lui faut son pompier (case sous le camion)"
		default:
			sub = tr.F("Il sauve {n} élément(s).", "n", itoa(gs.TruckSaves(k)))
		}
		tr.Seen(sub)
		say(dst, sub, x+60, ry+64, 18, false, -1, rgb(0x5a4a3a))
		if gs.Levels["caserne.camion_"+itoa(k)] == 0 {
			continue
		}
		if gs.Levels["caserne.camion_niv_"+itoa(k)] >= gs.TruckLvlMax() {
			say(dst, "Maximum", x+w-150, ry+48, 20, true, 0, rgb(0x2fa86b))
			continue
		}
		g.btn(dst, tr.F("+1 élément : {prix}", "prix", ptsPrice("ouvrier", sim.TruckUpCost)), x+w-430, ry+18, 380, 60, 0x2fa86b, gs.Points("ouvrier") >= sim.TruckUpCost, func() {
			if gs.BuyTruckUp(k) {
				g.snd.play(sBuy)
			}
		})
	}
}
