package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
	"ville/tr"
)

// Le tutoriel : au premier lancement (et après « Recommencer à zéro »), une fenêtre
// de bienvenue propose « Jouer » ou « Tutoriel ». Le tutoriel avance pas à pas : chaque
// étape se valide toute seule dès qu'elle est faite (même faite en avance), une flèche
// montre où aller, et les conseils habituels se taisent. Sur la carte, au début, seule
// la ferme se clique. « Passer le tutoriel » l'arrête ; « ? » sur la carte le relance.

const (
	tutoWelcome  = iota + 1 // la fenêtre de bienvenue (Jouer / Tutoriel)
	tutoMap                 // la carte : allons à la ferme
	tutoHarvest             // récolter un maïs
	tutoShelf               // le poser sur l'étal
	tutoServe               // servir le premier client
	tutoBuy                 // acheter une pousse
	tutoWait                // attendre qu'elles mûrissent
	tutoHarvest2            // les récolter
	tutoShelf2              // les poser
	tutoServe2              // encaisser le client suivant
	tutoSchool              // l'hélicoptère vers l'école
	tutoStudent             // accueillir un élève
	tutoMairie              // l'icône de la mairie
	tutoPoints              // la page des points (et le veilleur)
	tutoBack                // retour à la ferme
	tutoVest                // le vestiaire
	tutoWard                // la page « Personnage »
	tutoBravo               // fin
)

// tutoLock : le début du tutoriel, sur la carte (on ne peut que cliquer sur la ferme).
func (g *game) tutoLock() bool { return g.tuto == tutoMap && g.scene == sceneMap }

// newStart : une nouvelle partie commence sur la carte, avec la fenêtre de bienvenue.
func (g *game) newStart() {
	g.goMap()
	g.tuto = tutoWelcome
}

// setTuto : passe à une étape (et note où en sont les compteurs).
func (g *game) setTuto(n int) {
	g.tuto, g.tutoT, g.tutoReady = n, 0, true
	farm := g.sim.Place("ferme")
	switch n {
	case tutoHarvest:
		g.tutoBase = g.sim.Stats.Harvested
	case tutoServe, tutoServe2:
		g.tutoBase = farm.Served
	case tutoBuy:
		g.tutoBase = len(farm.Station("mais").Plants)
	case tutoStudent:
		g.tutoBase = g.sim.Place("ecole").Welcomed
	case tutoPoints: // la page s'ouvre sur les compétences de la ferme
		g.ptsTab, g.ptsPlace, g.ptsScroll = 0, "ferme", 0
	}
	if n > tutoWelcome && n != tutoBravo {
		g.snd.play(sClick)
	}
}

func (g *game) ripeMais() int {
	n := 0
	for _, pl := range g.sim.Place("ferme").Station("mais").Plants {
		if pl.G >= 1 {
			n++
		}
	}
	return n
}

// updateTuto : l'étape en cours est-elle faite ?
func (g *game) updateTuto(dt float64) {
	if g.tuto <= tutoWelcome {
		return
	}
	if !g.tutoReady {
		g.setTuto(g.tuto) // (partie rechargée en plein tutoriel)
	}
	g.tutoT += dt
	gs := g.sim
	farm := gs.Place("ferme")
	mais := farm.Station("mais")
	at := func(id string) bool { return g.scene == scenePlace && gs.Cur != nil && gs.Cur.ID == id }
	carry := gs.Player.Count(sim.Mais)
	done := false
	switch g.tuto {
	case tutoMap:
		done = at("ferme")
	case tutoHarvest:
		done = gs.Stats.Harvested > g.tutoBase || carry > 0
	case tutoShelf, tutoShelf2:
		done = carry == 0
	case tutoServe, tutoServe2:
		done = farm.Served > g.tutoBase
	case tutoBuy:
		done = len(mais.Plants) > g.tutoBase || len(mais.Plants) >= len(mais.Slots)
	case tutoWait:
		done = g.ripeMais() >= min(2, len(mais.Plants))
	case tutoHarvest2:
		done = carry >= 2 || (carry > 0 && g.ripeMais() == 0)
	case tutoSchool:
		done = at("ecole")
	case tutoStudent:
		done = gs.Place("ecole").Welcomed > g.tutoBase
	case tutoMairie:
		done = g.panel == panelMairie || g.panel == panelPoints
	case tutoPoints:
		done = g.panel == panelNone
	case tutoBack:
		done = at("ferme")
	case tutoVest:
		done = g.panel == panelWardrobe
	case tutoWard:
		done = g.panel == panelNone
	case tutoBravo:
		if g.tutoT > 12 {
			g.tuto = 0
			g.save()
		}
		return
	}
	if done {
		g.setTuto(g.tuto + 1)
		g.save()
	}
}

// tutoText : la consigne de l'étape.
func (g *game) tutoText() string {
	switch g.tuto {
	case tutoMap:
		return "Ici, c'est la carte de la ville, où se trouvent tous les lieux achetables. Pour l'instant, la ferme et l'école sont disponibles. Allons à la ferme."
	case tutoHarvest:
		return "Va récolter ton premier maïs."
	case tutoShelf:
		return "Apporte ton premier maïs sur l'étal."
	case tutoServe:
		return "Sers ton premier client."
	case tutoBuy:
		return "Achète une autre pousse de maïs."
	case tutoWait:
		return "Attends que les deux pousses soient mûres."
	case tutoHarvest2:
		return "Récolte-les."
	case tutoShelf2:
		return "Amène-les à l'étal."
	case tutoServe2:
		return "Encaisse le client suivant."
	case tutoSchool:
		return "Allons maintenant visiter l'école. Rends-toi près de l'hélicoptère et reste immobile sur la zone « École »."
	case tutoStudent:
		return "Va accueillir ton premier élève."
	case tutoMairie:
		return "Dans quelques minutes, le professeur l'aura formé. Tu gagneras un point école. Ils servent à améliorer les lieux. Clique sur l'icône « maison »."
	case tutoPoints:
		return "Chaque lieu a ses propres améliorations. En haut, tu peux acheter un veilleur de nuit pour continuer à gagner de l'argent même hors ligne. Ferme cette page pour continuer."
	case tutoBack:
		if k := g.sim.FavKey("ferme"); k != "" {
			return tr.F("Retourne à la ferme. Pour cela, tu peux soit utiliser l'hélicoptère, soit passer par la carte de la ville, soit utiliser le raccourci clavier correspondant (touche {k}).", "k", k)
		}
		return "Retourne à la ferme. Pour cela, tu peux soit utiliser l'hélicoptère, soit passer par la carte de la ville."
	case tutoVest:
		return "Va au vestiaire."
	case tutoWard:
		return "Tu peux acheter des accessoires pour améliorer ton personnage. Les sandales le feront marcher plus vite, le sac porter davantage. Ferme cette page pour retourner à la ferme."
	case tutoBravo:
		return "Bravo ! Tu as acquis le minimum pour jouer. Le reste, tu le découvriras tout seul. Bon jeu !"
	}
	return ""
}

// tutoWhere : le lieu où se passe l'étape ("" : n'importe où).
func (g *game) tutoWhere() string {
	switch {
	case g.tuto >= tutoMap && g.tuto <= tutoServe2, g.tuto == tutoBack:
		return "ferme"
	case g.tuto == tutoSchool, g.tuto == tutoStudent:
		return "ecole"
	}
	return ""
}

// tutoHint : la consigne et l'endroit (dans le lieu) où pointe la flèche.
func (g *game) tutoHint() (string, V, bool) {
	txt := g.tutoText()
	p := g.sim.Cur
	if p == nil || g.tuto == tutoWelcome || g.tuto == tutoBravo {
		return txt, V{}, false
	}
	if id := g.tutoWhere(); id != "" && p.ID != id {
		// pas au bon endroit : la flèche montre l'hélicoptère qui y mène
		for _, d := range g.busDests(p) {
			if d.id == id {
				return txt, d.zone.C(), true
			}
		}
		return txt, V{}, false
	}
	switch g.tuto {
	case tutoHarvest, tutoWait, tutoHarvest2:
		if g.tuto == tutoWait {
			return txt, V{}, false
		}
		return txt, p.Station("mais").Body.C(), true
	case tutoShelf, tutoShelf2:
		return txt, p.Station("etal_1").Zone.C(), true
	case tutoServe, tutoServe2:
		if regs := p.Registers(); len(regs) > 0 {
			return txt, regs[0].Zone.C(), true
		}
	case tutoBuy:
		return txt, p.Station("mais").PlantZone().C(), true
	case tutoStudent:
		return txt, p.Station("accueil").Zone.C(), true
	case tutoVest:
		return txt, p.Vestiaire.C(), true
	}
	return txt, V{}, false
}

// drawStart : la fenêtre de bienvenue d'une nouvelle partie.
func (g *game) drawStart(dst *ebiten.Image) {
	w, h := 1000.0, 470.0
	x, y := g.panelFrame(dst, "Bienvenue !", w, h)
	for i, l := range wrap("Bienvenue à Toute une Ville, le jeu de gestion de Clairbourg !", 34, w-120, true) {
		say(dst, l, x+w/2, y+70+float64(i)*48, 34, true, 0, rgb(0x5a3d1a))
	}
	for i, l := range wrap("Tu peux jouer tout de suite, ou te laisser guider quelques minutes pas à pas.", 26, w-120, false) {
		say(dst, l, x+w/2, y+180+float64(i)*36, 26, false, 0, rgb(0x5a4a3a))
	}
	g.btn(dst, "Jouer", x+w/2-380, y+h-190, 340, 80, 0x2fa86b, true, func() { g.closePanel() })
	g.btn(dst, "Tutoriel", x+w/2+40, y+h-190, 340, 80, 0xd4a017, true, func() {
		g.panel = panelNone
		g.setTuto(tutoMap)
		g.save()
	})
}

// tutoArrow : une flèche jaune qui rebondit vers un rectangle de l'interface
// (au-dessus, pointe en bas ; en dessous, pointe en haut), et un halo autour.
func (g *game) tutoArrow(dst *ebiten.Image, r [4]float64, below bool) {
	if r[2] == 0 {
		return
	}
	k := 0.5 + 0.5*math.Sin(g.t*5)
	strokeRect(dst, r[0]-6, r[1]-6, r[2]+12, r[3]+12, 5, rgba(0xffd23f, 0.6+0.4*k))
	cx := r[0] + r[2]/2
	b := 14 * math.Abs(math.Sin(g.t*4))
	if below {
		ty := r[1] + r[3] + 14 + b
		polygon(dst, []V{{X: cx - 26, Y: ty + 32}, {X: cx + 26, Y: ty + 32}, {X: cx, Y: ty}}, rgb(0xffd23f))
		rect(dst, cx-12, ty+32, 24, 34, rgb(0xffd23f))
		return
	}
	ty := r[1] - 14 - b
	polygon(dst, []V{{X: cx - 26, Y: ty - 32}, {X: cx + 26, Y: ty - 32}, {X: cx, Y: ty}}, rgb(0xffd23f))
	rect(dst, cx-12, ty-66, 24, 34, rgb(0xffd23f))
}

// drawTuto : l'encadré de l'étape (en bas de l'écran, par-dessus tout) et les flèches
// vers les boutons de l'interface (celles du monde sont dessinées par drawGuide).
func (g *game) drawTuto(dst *ebiten.Image) {
	if g.tuto <= tutoWelcome || (g.panel != panelNone && g.panel != panelMairie && g.panel != panelPoints && g.panel != panelWardrobe) {
		return
	}
	// flèches vers l'interface
	switch {
	case g.scene == sceneMap && g.panel == panelNone:
		if id := g.tutoWhere(); id != "" {
			g.tutoArrow(dst, g.mapRects[id], false)
		}
	case g.tuto == tutoMairie && g.panel == panelNone:
		g.tutoArrow(dst, g.tutoMairie, true)
	case g.tuto == tutoPoints && (g.panel == panelMairie || g.panel == panelPoints):
		r := g.tutoVeil
		k := 0.5 + 0.5*math.Sin(g.t*5)
		strokeRect(dst, r[0]-6, r[1]-6, r[2]+12, r[3]+12, 5, rgba(0xffd23f, 0.6+0.4*k))
	}
	txt := g.tutoText()
	if txt == "" {
		return
	}
	if g.panel != panelNone { // une page est ouverte : la consigne remplace son titre
		r := g.panelRect
		bw := r[2] - 140
		size := 22.0
		lines := wrap(txt, size, bw-40, true)
		for len(lines) > 2 && size > 15 {
			size--
			lines = wrap(txt, size, bw-40, true)
		}
		roundRect(dst, r[0]+14, r[1]+6, bw, 78, 20, rgba(0x2b2140, 0.96))
		strokeRect(dst, r[0]+17, r[1]+9, bw-6, 72, 3, rgba(0xffd23f, 0.8))
		lh := size * 1.3
		for i, l := range lines {
			say(dst, l, r[0]+36, r[1]+45+(float64(i)-float64(len(lines)-1)/2)*lh, size, true, -1, rgb(0xffffff))
		}
		return
	}
	size := 28.0
	if scr.Small {
		size = 40
	}
	w := math.Min(1500, scr.VW-80)
	lines := wrap(txt, size, w-420, true)
	lh := size * 1.35
	h := math.Max(110, float64(len(lines))*lh+56)
	x, y := scr.VW/2-w/2, scr.VH-16-h
	roundRect(dst, x, y+6, w, h, 24, rgba(0x000000, 0.3))
	roundRect(dst, x, y, w, h, 24, rgba(0x2b2140, 0.94))
	strokeRect(dst, x+3, y+3, w-6, h-6, 3, rgba(0xffd23f, 0.8))
	say(dst, tr.F("Tutoriel · étape {n} / {m}", "n", itoa(g.tuto-tutoWelcome), "m", itoa(tutoBravo-tutoWelcome)), x+30, y+24, 17, true, -1, rgb(0xffd23f))
	for i, l := range lines {
		say(dst, l, x+30, y+36+lh/2+float64(i)*lh, size, true, -1, rgb(0xffffff))
	}
	if g.tuto == tutoBravo {
		g.btn(dst, "Merci !", x+w-300, y+h/2-38, 270, 76, 0x2fa86b, true, func() {
			g.tuto = 0
			g.save()
		})
		return
	}
	g.btn(dst, "Passer le tutoriel", x+w-370, y+h/2-34, 340, 68, 0x8a7a6a, true, func() {
		g.tuto = 0
		g.save()
	})
}
