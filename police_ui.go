package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
	"ville/tr"
)

// Le menu des patrouilles (le parking du commissariat) et le petit tutoriel du commissariat.

// mapNames : le nom des lieux comme sur la carte.
var mapNames = map[string]string{"ferme": "Ma ferme", "super": "Supermarché", "cremerie": "Crèmerie", "exotique": "Ferme exotique",
	"boulangerie": "Boulangerie", "glacier": "Glacier", "port": "Port", "elevage": "Ferme d'élevage", "boucherie": "Boucherie",
	"domaine": "Le domaine", "creperie": "Crêperie", "patisserie": "Pâtisserie", "parc": "Parc d'attractions", "huilerie": "L'huilerie"}

// policeHUD : au commissariat, le tutoriel la première fois ; la case et le parking ouvrent le menu.
func (g *game) policeHUD(dst *ebiten.Image) {
	p := g.sim.Cur
	if p == nil || p.Pol == nil {
		return
	}
	if !g.sim.PolTutoSeen() && g.panel == panelNone {
		g.sim.SetPolTutoSeen()
		g.polTuto = 0
		g.openPanel(panelPoliceTuto)
		return
	}
	if !g.sim.HasChief() {
		return
	}
	pl := g.sim.Player
	in := sim.PolPatrolZone().Has(pl.Pos)
	vin := sim.PolVideoZone().Has(pl.Pos)
	if (in || vin) && !pl.Moving && !g.polZoneIn && g.panel == panelNone {
		g.polZoneIn = true
		g.polTab = 0
		if vin {
			g.polTab = 1
		}
		g.openPanel(panelPolice)
	}
	if !in && !vin {
		g.polZoneIn = false
	}
	// un clic sur le parking ouvre aussi le menu
	b1, bn := sim.PolBay(1), sim.PolBay(sim.PolCars())
	x0, y0 := g.worldCam.P(b1.X-sim.PolParkM, sim.PolParkY)
	x1, y1 := g.worldCam.P(bn.X+bn.W+sim.PolParkM, sim.PolParkY+sim.PolParkH)
	s := cam.s
	if g.panel == panelNone {
		g.addButton(float64(x0)/s, float64(y0)/s, float64(x1-x0)/s, float64(y1-y0)/s, func() { g.polTab = 0; g.openPanel(panelPolice) })
		g.btns[len(g.btns)-1].danger = true // (pas en restant dessus : seulement d'un clic)
	}
}

// drawPolicePanel : à gauche les voitures, à droite les lieux de la rangée du haut.
func (g *game) drawPolicePanel(dst *ebiten.Image) {
	gs := g.sim
	n := sim.PolCars()
	w, h := math.Min(1500, scr.VW-60), 960.0
	x, y := g.panelFrame(dst, "Commissariat", w, h)
	// deux onglets : les voitures (lieutenants), la vidéosurveillance (capitaines)
	for i, lbl := range []string{"Patrouilles", "Vidéosurveillance"} {
		i := i
		col := uint32(0x8a9ab8)
		if g.polTab == i {
			col = 0x2b6cd6
		}
		g.btn(dst, lbl, x+w/2-370+float64(i)*380, y+14, 360, 60, col, true, func() { g.polTab = i })
	}
	y += 70
	h -= 70
	if g.polTab == 1 {
		g.drawVideoTab(dst, x, y, w, h)
		return
	}
	say(dst, "Choisis une voiture à gauche, puis le lieu où elle patrouille : les voleurs n'y viendront plus.", x+w/2, y+34, 20, false, 0, rgb(0x5a4a3a))
	if g.polSel < 1 || g.polSel > n {
		g.polSel, g.polSlot = 1, 0
	}
	lw := 560.0
	lx, ly := x+30, y+80
	ch := math.Min(230, (h-200)/float64(n)-10)
	for k := 1; k <= n; k++ {
		k := k
		cy := ly + float64(k-1)*(ch+10)
		sel := g.polSel == k
		bg := uint32(0xf1e6d2)
		if sel {
			bg = 0xd9e6f7
		}
		roundRect(dst, lx, cy, lw, ch, 18, rgb(bg))
		scaled(lx+60, cy+ch/2, 0.42, func() { drawPoliceCarTop(dst, lx+60, cy+ch/2, false, g.t) })
		say(dst, tr.F("Voiture n° {n}", "n", itoa(k)), lx+120, cy+30, 24, true, -1, rgb(0x2b3a5a))
		status := ""
		switch {
		case !gs.PolCarOwned(k):
			status = "Pas encore achetée (au parking)"
		case !gs.PolLieutenant(k):
			status = tr.F("Il lui faut son lieutenant (case sous la voiture n° {n})", "n", itoa(k))
		case gs.PolOnPatrol(k):
			status = "En patrouille"
		default:
			status = "Au parking : choisis un lieu"
		}
		tr.Seen(status)
		sayFit(dst, status, lx+120, cy+62, 18, false, -1, lw-140, rgb(0x5a4a3a))
		if !gs.PolCarOwned(k) {
			continue
		}
		// ses cases : un lieu (deux avec les mécaniciens)
		for s := 0; s < 2; s++ {
			s := s
			bx, by, bw := lx+120+float64(s)*215, cy+86, 205.0
			if s >= gs.PolSlots(k) {
				if s == 1 { // les mécaniciens : un deuxième lieu
					c := sim.DoubleCost(k)
					ok := gs.Points("ouvrier") >= c
					g.btn(dst, ptsPrice("ouvrier", c), bx, by, bw, 56, 0x2fa86b, ok, func() {
						if gs.BuyDouble(k) {
							g.snd.play(sBuy)
							g.polSel, g.polSlot = k, 1
						}
					})
					sayFit(dst, "Mécaniciens : un 2e lieu", bx+bw/2, by+76, 15, false, 0, bw, rgb(0x5a4a3a))
				}
				continue
			}
			lbl := "Aucun lieu"
			if id := gs.PolPatrol(k, s); id != "" {
				lbl = mapNames[id]
			}
			col := uint32(0x8a9ab8)
			if sel && g.polSlot == s {
				col = 0x2b6cd6
			}
			g.btn(dst, lbl, bx, by, bw, 56, col, true, func() { g.polSel, g.polSlot = k, s })
		}
	}
	// à droite : les lieux de la rangée du haut (débloqués)
	rx := lx + lw + 40
	rw := x + w - 30 - rx
	say(dst, tr.F("Voiture n° {n} : où patrouille-t-elle ?", "n", itoa(g.polSel)), rx+rw/2, ly+20, 22, true, 0, rgb(0x2b3a5a))
	var ids []string
	for _, id := range sim.TopRow() {
		if q := gs.Place(id); q != nil && q.Owned {
			ids = append(ids, id)
		}
	}
	cols := 3
	bw := (rw - float64(cols-1)*14) / float64(cols)
	can := gs.PolCarOwned(g.polSel) && g.polSlot < gs.PolSlots(g.polSel)
	for i, id := range ids {
		id := id
		bx := rx + float64(i%cols)*(bw+14)
		by := ly + 60 + float64(i/cols)*80
		col := uint32(0x5b7fd6)
		k := gs.PatrolOf(id)
		mine := gs.PolPatrol(g.polSel, g.polSlot) == id
		switch {
		case mine:
			col = 0x2fa86b
		case k > 0 || g.patrolledBy(id) > 0:
			col = 0x9a8ab8
		}
		g.btn(dst, mapNames[id], bx, by, bw, 64, col, can, func() {
			if mine {
				gs.SetPatrol(g.polSel, g.polSlot, "") // (re-cliquer : la voiture rentre)
			} else {
				gs.SetPatrol(g.polSel, g.polSlot, id)
			}
		})
		if o := g.patrolledBy(id); o > 0 && !mine {
			sayFit(dst, tr.F("voiture n° {n}", "n", itoa(o)), bx+bw/2, by+76, 14, false, 0, bw, rgb(0x6a5a8a))
		}
	}
	if can && gs.PolPatrol(g.polSel, g.polSlot) != "" {
		g.btn(dst, "Rentrer au parking", rx+rw/2-170, y+h-200, 340, 64, 0x8a7a6a, true, func() { gs.SetPatrol(g.polSel, g.polSlot, "") })
	}
	if !gs.PolDriven(g.polSel) && gs.PolCarOwned(g.polSel) {
		say(dst, "Sans son lieutenant, la voiture ne patrouille pas (le lieu choisi reste sans protection).", x+w/2, y+h-120, 18, false, 0, rgb(0xb03a3a))
	}
}

// patrolledBy : la voiture à qui ce lieu est confié (même sans lieutenant ; 0 : aucune).
func (g *game) patrolledBy(id string) int { return g.sim.PatrolAssigned(id) }

// polTutoPages : le petit tutoriel du commissariat (facultatif, proposé à la première visite).
var polTutoPages = []string{
	"Le commissariat protège la ville des voleurs. Il est offert, mais il ne fonctionne qu'avec un commissaire : il ne s'achète pas avec de l'argent, il coûte 10 points d'université (l'école). Sa case est dans son bureau.",
	"Avec le commissaire, la première voiture de police est offerte. Les suivantes s'achètent au parking. Une voiture ne roule qu'avec son lieutenant : il se forme avec des points du collège, sur la case sous sa voiture (10 points pour la première voiture, puis 50, puis 250…). Un lieutenant sans patrouille erre sur le parking en sifflotant.",
	"Sur le parking, la case « Patrouilles » (ou un clic sur le parking) ouvre le menu des voitures : choisis le lieu où chacune patrouille. Ce lieu ne peut plus être attaqué. Avec des mécaniciens (points du primaire), une voiture protège un deuxième lieu.",
	"Dans les bureaux, à gauche, les capitaines (points du lycée) surveillent chacun un lieu par vidéosurveillance (case du hall). S'ils repèrent les voleurs, ceux-ci sont arrêtés en partant et passent 2 heures en garde à vue, à droite, s'il y a une cellule libre.",
}

func (g *game) drawPoliceTuto(dst *ebiten.Image) {
	w, h := math.Min(1100, scr.VW-60), 520.0
	x, y := g.panelFrame(dst, "Le commissariat", w, h)
	if g.polTuto == 0 {
		say(dst, "Bienvenue au commissariat ! Veux-tu une petite explication ?", x+w/2, y+120, 26, true, 0, rgb(0x2b3a5a))
		g.btn(dst, "Oui, explique-moi", x+w/2-330, y+260, 300, 76, 0x2fa86b, true, func() { g.polTuto = 1 })
		g.btn(dst, "Non merci", x+w/2+30, y+260, 300, 76, 0x8a7a6a, true, func() { g.closePanel() })
		return
	}
	i := g.polTuto - 1
	for j, l := range wrap(polTutoPages[i], 26, w-120, false) {
		say(dst, l, x+60, y+50+float64(j)*40, 26, false, -1, rgb(0x3b2a10))
	}
	say(dst, tr.F("{n} / {m}", "n", itoa(i+1), "m", itoa(len(polTutoPages))), x+w/2, y+h-150, 20, false, 0, rgb(0x8a7a6a))
	if i+1 < len(polTutoPages) {
		g.btn(dst, "Suivant", x+w/2-150, y+h-130, 300, 70, 0x2fa86b, true, func() { g.polTuto++ })
	} else {
		g.btn(dst, "C'est compris !", x+w/2-150, y+h-130, 300, 70, 0x2fa86b, true, func() { g.closePanel() })
	}
}

// drawVideoTab : à gauche les bureaux des capitaines, à droite les lieux ; en bas, la garde à vue.
func (g *game) drawVideoTab(dst *ebiten.Image, x, y, w, h float64) {
	gs := g.sim
	n := sim.PolOffices()
	say(dst, "Choisis un capitaine à gauche, puis le lieu qu'il surveille : il peut faire arrêter les voleurs.", x+w/2, y+34, 20, false, 0, rgb(0x5a4a3a))
	if g.vidSel < 1 || g.vidSel > n {
		g.vidSel = 1
	}
	lw := 640.0
	lx, ly := x+30, y+70
	ch := math.Min(130, (h-230)/float64(n)-8)
	for k := 1; k <= n; k++ {
		k := k
		cy := ly + float64(k-1)*(ch+8)
		bg := uint32(0xf1e6d2)
		if g.vidSel == k {
			bg = 0xd9e6f7
		}
		roundRect(dst, lx, cy, lw, ch, 16, rgb(bg))
		say(dst, tr.F("Bureau n° {n}", "n", itoa(k)), lx+20, cy+24, 21, true, -1, rgb(0x2b3a5a))
		switch {
		case !gs.PolOfficeOwned(k):
			sayFit(dst, "Pas encore construit (bureaux des capitaines)", lx+20, cy+ch/2+16, 17, false, -1, lw-40, rgb(0x5a4a3a))
			continue
		case !gs.PolCaptain(k):
			sayFit(dst, "Il lui faut son capitaine (case dans son bureau)", lx+20, cy+ch/2+16, 17, false, -1, lw-40, rgb(0x5a4a3a))
			continue
		}
		lbl := "Aucun lieu"
		if id := gs.PolCamera(k); id != "" {
			lbl = mapNames[id]
		}
		col := uint32(0x8a9ab8)
		if g.vidSel == k {
			col = 0x2b6cd6
		}
		g.btn(dst, lbl, lx+200, cy+6, 220, 50, col, true, func() { g.vidSel = k })
		p := int(math.Round(100 * gs.CatchChance(k)))
		say(dst, tr.F("Chance d'arrêter les voleurs : {p} %", "p", itoa(p)), lx+20, cy+ch-26, 17, false, -1, rgb(0x5a4a3a))
		if p < 100 { // des techniciens : 10 points du primaire pour +1 %, dans CE bureau
			for i, pts := range []int{10, 100} {
				pts := pts
				ok := gs.Points("ouvrier") >= pts
				g.btn(dst, tr.F("+{p} % : {prix}", "p", itoa(pts/10), "prix", ptsPrice("ouvrier", pts)), lx+lw-210, cy+6+float64(i)*(ch/2-2), 200, ch/2-12, 0x2fa86b, ok, func() {
					if gs.BuyVideo(k, pts) {
						g.snd.play(sBuy)
					}
				})
			}
		}
	}
	// à droite : les lieux
	rx := lx + lw + 40
	rw := x + w - 30 - rx
	say(dst, tr.F("Bureau n° {n} : quel lieu surveille-t-il ?", "n", itoa(g.vidSel)), rx+rw/2, ly+10, 22, true, 0, rgb(0x2b3a5a))
	var ids []string
	for _, id := range sim.TopRow() {
		if q := gs.Place(id); q != nil && q.Owned {
			ids = append(ids, id)
		}
	}
	cols := 2
	bw := (rw - float64(cols-1)*14) / float64(cols)
	can := gs.PolWatching(g.vidSel)
	for i, id := range ids {
		id := id
		bx := rx + float64(i%cols)*(bw+14)
		by := ly + 50 + float64(i/cols)*74
		col := uint32(0x5b7fd6)
		mine := gs.PolCamera(g.vidSel) == id
		other := 0
		for c := 1; c <= sim.PolOffices(); c++ {
			if c != g.vidSel && gs.PolCamera(c) == id {
				other = c
			}
		}
		switch {
		case mine:
			col = 0x2fa86b
		case other > 0 || g.patrolledBy(id) > 0:
			col = 0x9a8ab8
		}
		car := g.patrolledBy(id) > 0 // (déjà protégé par une voiture : pas de capitaine)
		g.btn(dst, mapNames[id], bx, by, bw, 58, col, can && !car, func() {
			if mine {
				gs.SetCamera(g.vidSel, "")
			} else {
				gs.SetCamera(g.vidSel, id)
			}
		})
		switch {
		case other > 0 && !mine:
			sayFit(dst, tr.F("bureau n° {n}", "n", itoa(other)), bx+bw/2, by+68, 13, false, 0, bw, rgb(0x6a5a8a))
		case g.patrolledBy(id) > 0:
			sayFit(dst, tr.F("voiture n° {n}", "n", itoa(g.patrolledBy(id))), bx+bw/2, by+68, 13, false, 0, bw, rgb(0x6a5a8a))
		}
	}
	// la garde à vue
	places := gs.PolCellsOwned() * sim.PolCellCap
	txt := tr.F("Garde à vue : {n} voleur(s) sur {m} place(s). Voleurs en liberté : {v}.", "n", itoa(gs.InCustody()), "m", itoa(places), "v", itoa(gs.Thieves()))
	if places == 0 {
		txt = tr.F("Aucune cellule de garde à vue : les voleurs repérés restent libres. Voleurs en liberté : {v}.", "v", itoa(gs.Thieves()))
	}
	tr.Seen(txt)
	sayFit(dst, txt, x+w/2, y+h-130, 20, true, 0, w-60, rgb(0x2b3a5a))
}
