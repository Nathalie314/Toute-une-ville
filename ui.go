package main

import (
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// ---------------------------------------------------------------- particules

type particle struct {
	pos, vel  V
	life, max float64
	text      string
	col       uint32
	size      float64
	kind      int // 0 texte, 1 confetti, 2 feuille
	it        sim.Item
}

type toastMsg struct {
	text, sub string
	t         float64
}

func (g *game) toast(text, sub string) {
	g.toasts = append(g.toasts, toastMsg{text: text, sub: sub})
}

// toastSplit : un message, le cadeau d'une commande sur une seconde ligne.
func (g *game) toastSplit(text string) {
	if i := strings.Index(text, " Cadeau"); i >= 0 {
		g.toast(text[:i], text[i+1:])
		return
	}
	g.toast(text, "")
}

func (g *game) handleEvents() {
	cur := g.sim.Cur
	for _, e := range g.sim.Events {
		// un événement d'un autre lieu : pas de son ni d'étincelles (les messages restent)
		here := e.Place == "" || (cur != nil && g.scene == scenePlace && e.Place == cur.ID)
		if !here {
			switch e.Kind {
			case sim.EvLevel:
				g.toastSplit(e.Text)
			case sim.EvUnlock:
				g.toast("Nouveau : "+e.Text+" !", "")
			case sim.EvBadge:
				g.toast("Badge : "+e.Text+" !", "")
			}
			continue
		}
		switch e.Kind {
		case sim.EvCoin:
			if cur != nil && g.scene == scenePlace && (e.Text == "" || e.Text == cur.ID) {
				g.parts = append(g.parts, particle{pos: e.Pos, vel: V{Y: -70}, life: 1.4, max: 1.4, text: "+" + money(e.Amount), col: 0xffe066, size: 30})
				if g.t-g.coinSnd > 0.12 {
					g.snd.play(sCoin)
					g.coinSnd = g.t
				}
			}
		case sim.EvHarvest, sim.EvTake:
			if g.scene == scenePlace {
				g.snd.play(sPop)
				for k := 0; k < 3; k++ {
					a := g.t*7 + float64(k)*2.1
					g.parts = append(g.parts, particle{pos: e.Pos.Add(V{Y: -20}), vel: V{X: math.Cos(a) * 80, Y: -60 - 40*math.Sin(a*1.3)}, life: 0.5, max: 0.5, kind: 2, col: 0x8fd16a, size: 5})
				}
			}
		case sim.EvLevel:
			g.snd.play(sUnlock)
			g.toastSplit(e.Text)
			for k := 0; k < 30; k++ {
				a := float64(k) * 2.4
				sp := 140 + 200*math.Mod(float64(k)*0.37, 1)
				g.parts = append(g.parts, particle{pos: e.Pos, vel: V{X: math.Cos(a) * sp, Y: math.Sin(a)*sp - 220}, life: 1.2, max: 1.2, kind: 1, col: 0xffd23f, size: 7})
			}
		case sim.EvTruck:
			if cur != nil && cur.ID == "super" {
				g.parts = append(g.parts, particle{pos: V{X: 560, Y: 100}, vel: V{Y: -40}, life: 2, max: 2, text: "+" + itoa(int(e.Amount)) + " produits livrés", col: 0x9fd3ff, size: 26})
			}
		case sim.EvOrder:
			g.snd.play(sPhone) // « drriing » : on la reconnaît partout
			g.toastSplit(e.Text)
		case sim.EvBadge:
			g.snd.play(sBadge)
			g.toast("Badge : "+e.Text+" !", "")
		case sim.EvPlant:
			switch e.Text {
			case "achat": // une nouvelle case achetée
				g.snd.play(sHire)
			case "": // le joueur replante : un tout petit « pof » (les ouvriers : rien)
				g.snd.play(sSeed)
			}
			for k := 0; k < 14; k++ {
				a := float64(k) * 0.45
				g.parts = append(g.parts, particle{pos: e.Pos.Add(V{Y: -10}), vel: V{X: math.Cos(a) * 120, Y: -150 - 60*math.Sin(a*2)}, life: 0.8, max: 0.8, kind: 2, col: 0x7ed957, size: 7})
			}
		case sim.EvDrop:
			if g.scene == scenePlace {
				g.snd.play(sDrop)
			}
		case sim.EvPay:
			if g.t-g.lastPay > 0.07 {
				g.snd.play(sTick)
				g.lastPay = g.t
			}
		case sim.EvUnlock:
			g.snd.play(sUnlock)
			g.toast("Nouveau : "+e.Text+" !", "")
			for k := 0; k < 60; k++ {
				a := float64(k) * 2.4
				sp := 180 + 260*math.Mod(float64(k)*0.37, 1)
				cols := []uint32{0xff6fa8, 0xffd23f, 0x4fc3ff, 0x7ed957, 0xff8a3d}
				g.parts = append(g.parts, particle{pos: e.Pos, vel: V{X: math.Cos(a) * sp, Y: math.Sin(a)*sp - 250}, life: 1.6, max: 1.6, kind: 1, col: cols[k%5], size: 7})
			}
		case sim.EvAngry:
			if g.scene == scenePlace {
				g.snd.play(sAngry)
			}
		case sim.EvHire:
			g.snd.play(sHire)
		case sim.EvBoost:
			if cur != nil && cur.ID == "ecole" && g.scene == scenePlace {
				g.snd.play(sUnlock)
				g.parts = append(g.parts, particle{pos: e.Pos.Add(V{Y: -40}), vel: V{Y: -50}, life: 2.2, max: 2.2, text: "Merci ! Vitesse ×2 pendant 1 min", col: 0xffd23f, size: 26})
			}
		case sim.EvPoint:
			if cur != nil && cur.ID == "ecole" && g.scene == scenePlace {
				g.snd.play(sUnlock)
				g.parts = append(g.parts, particle{pos: e.Pos.Add(V{Y: -60}), vel: V{Y: -60}, life: 2, max: 2, text: "+1 point " + e.Text, col: 0xffd23f, size: 26})
			}
		}
	}
	g.sim.Events = g.sim.Events[:0]
}

func (g *game) updateParticles(dt float64) {
	out := g.parts[:0]
	for _, p := range g.parts {
		p.life -= dt
		if p.life <= 0 {
			continue
		}
		p.pos = p.pos.Add(p.vel.Mul(dt))
		if p.kind > 0 {
			p.vel.Y += 600 * dt
		}
		out = append(out, p)
	}
	g.parts = out
	for i := range g.toasts {
		g.toasts[i].t += dt
	}
	for len(g.toasts) > 0 && g.toasts[0].t > 3.2 {
		g.toasts = g.toasts[1:]
	}
}

func (g *game) drawParticles(dst *ebiten.Image) {
	for _, p := range g.parts {
		a := math.Min(1, p.life/p.max*2)
		switch p.kind {
		case 0:
			sayOutlined(dst, p.text, p.pos.X, p.pos.Y, p.size, 0, rgba(p.col, a), rgba(0x3b2a10, a))
		default:
			rect(dst, p.pos.X-p.size/2, p.pos.Y-p.size/2, p.size, p.size*0.7, rgba(p.col, a))
		}
	}
}

func (g *game) drawToasts(dst *ebiten.Image) {
	y := 150.0
	for _, t := range g.toasts {
		a := math.Min(1, math.Min(t.t*4, (3.2-t.t)*3))
		dy := -20 * (1 - math.Min(1, t.t*4))
		sz := 38.0
		for sz > 22 && textWidth(t.text, sz, true)+80 > scr.VW-40 {
			sz -= 2
		}
		w := textWidth(t.text, sz, true) + 80
		h := 72.0
		if t.sub != "" {
			w = math.Max(w, textWidth(t.sub, 28, true)+110)
			h = 118
		}
		roundRect(dst, scr.VW/2-w/2, y+dy-36, w, h, 36, rgba(0x2b2140, 0.85*a))
		say(dst, t.text, scr.VW/2, y+dy, sz, true, 0, rgba(0xffe066, a))
		if t.sub != "" { // le cadeau d'une commande : il faut que ça donne envie
			tw := textWidth(t.sub, 28, true)
			drawGift(dst, scr.VW/2-tw/2-22, y+dy+44, 13)
			say(dst, t.sub, scr.VW/2+14, y+dy+46, 28, true, 0, rgba(0xffffff, a))
		}
		y += h + 18
	}
}

// ---------------------------------------------------------------- guide

// hint renvoie le conseil du moment et l'endroit vers lequel il pointe.
func (g *game) hint() (string, V, bool) {
	gs := g.sim
	p := gs.Cur
	if p == nil {
		return "", V{}, false
	}
	pl := gs.Player
	for _, reg := range p.Registers() {
		if reg.Cashier == nil && p.QueueAt(reg) > 0 {
			if reg.Zone.Has(pl.Pos) {
				return "Reste à la caisse tant que des clients attendent.", V{}, false
			}
			return "Un client attend à la caisse : va encaisser !", reg.Zone.C(), true
		}
	}
	if gs.Boosted("maire", "") { // le cadeau du maire : on le rappelle tant qu'il dure
		for _, q := range gs.Places {
			if !q.Owned && q.Price > 0 && gs.Money >= gs.PriceOf(q) {
				r := []rune(q.Name)
				return "Cadeau du maire : " + strings.ToLower(string(r[:1])) + string(r[1:]) + " à " + map[bool]string{true: "moitié prix", false: "un quart du prix"}[gs.Boost.Mul() == 2] + " ! Va sur la carte (Tab).", V{}, false
			}
		}
	}
	if p.ID == "super" {
		if txt, at, arrow := g.superRefused(p); txt != "" {
			return txt, at, arrow
		}
	}
	// l'école : des élèves attendent à l'accueil et personne ne les inscrit
	// (après les tout premiers clients de la ferme, pour ne pas cacher les premiers conseils)
	if sc := gs.Place("ecole"); sc != nil && p.ID != "ecole" && gs.Place("ferme").Served >= 5 && sc.NeedsWelcome() {
		return "Va accueillir les élèves à l'école (touche 2) !", V{}, false
	}
	// près d'une amélioration au sol : on explique à quoi elle sert
	for _, sp := range p.Spots {
		if !p.SpotVisible(sp) {
			continue
		}
		z := sp.Zone
		if !(sim.Rect{X: z.X - 70, Y: z.Y - 70, W: z.W + 140, H: z.H + 140}).Has(pl.Pos) {
			continue
		}
		u := p.SpotUpgrade(sp)
		if u == nil || u.Desc == nil {
			continue
		}
		l := gs.Levels[u.Key]
		switch {
		case l >= u.Max && u.Max == 1:
			return u.Name + " : " + u.Desc(l), V{}, false
		case l >= u.Max:
			return u.Name + " (maximum) : " + u.Desc(l), V{}, false
		case u.Now:
			return u.Name + " : " + u.Desc(l), V{}, false
		case u.Max == 1:
			return u.Name + " : " + u.Desc(1), V{}, false
		default:
			return u.Name + ", niveau " + itoa(l+1) + " : " + u.Desc(l+1), V{}, false
		}
	}
	if p.ID == "ecole" {
		if txt, at, arrow := g.schoolHint(p); txt != "" {
			return txt, at, arrow
		}
	}

	if o := gs.Order; o != nil && p.ID == o.Place() {
		box := p.Station("commande")
		for _, it := range pl.Carry {
			if o.Need(it) > 0 {
				if box.Zone.Has(pl.Pos) {
					return "Reste là : tu remplis la commande (" + o.Who + ").", V{}, false
				}
				return "Tu as de quoi remplir la commande (" + o.Who + ") : va à la caisse des commandes.", box.Zone.C(), true
			}
		}
	}
	// des sachets de graines dans les bras : vers la case vide la plus proche
	for _, f := range p.Stations {
		if !f.Built || !f.Replant || pl.Count(f.Seed) == 0 {
			continue
		}
		bd, at := 1e18, V{}
		for _, q := range f.Plants {
			if q.Empty() && q.P.Dist(pl.Pos) < bd {
				bd, at = q.P.Dist(pl.Pos), q.P
			}
		}
		if bd < 1e17 {
			if bd < 90 {
				return "Les graines se plantent toutes seules sur les cases vides.", V{}, false
			}
			return "Marche sur les cases vides de la colonne pour planter tes graines.", at, true
		}
	}
	// près d'un achat (bâtiment, champ, salle…) : à quoi il sert
	for _, u := range p.Unlocks {
		if u.Done || u.Hint == "" || !p.Visible(u) {
			continue
		}
		z := u.Zone
		if (sim.Rect{X: z.X - 60, Y: z.Y - 60, W: z.W + 120, H: z.H + 120}).Has(pl.Pos) {
			return u.Name + " : " + u.Hint, V{}, false
		}
	}
	if p.ID == "super" { // les rayons presque vides passent avant les achats possibles
		if txt, at, arrow := g.superHint(p); txt != "" {
			return txt, at, arrow
		}
		if txt := gs.RayonSummary(); txt != "" {
			return txt, V{}, false
		}
	}
	// achat abordable le moins cher : déblocage ou nouvelle pousse
	best, bestTxt := 1e18, ""
	var bestZone sim.Rect
	for _, u := range p.Unlocks {
		if !u.Done && p.Visible(u) && gs.Money >= u.Cost-u.Paid && u.Cost < 1e6 && u.Cost < best {
			best, bestZone = u.Cost, u.Zone
			bestTxt = "Tu peux acheter « " + u.Name + " » : arrête-toi sur sa zone."
		}
	}
	for _, s := range p.Stations {
		if s.Kind == sim.Field && s.Built && len(s.Plants) < len(s.Slots) && gs.Money >= s.PlantCost()-s.PlantPaid && s.PlantCost() < best {
			best, bestZone = s.PlantCost(), s.PlantZone()
			bestTxt = "Achète une pousse de plus : arrête-toi sur la case « +1 »."
		}
	}
	if bestTxt != "" {
		if bestZone.Has(pl.Pos) {
			return "Reste immobile sur la zone : l'argent s'y verse…", V{}, false
		}
		return bestTxt, bestZone.C(), true
	}
	if p.ID == "ecole" {
		return "", V{}, false
	}
	for _, q := range p.Stations {
		if q.Kind == sim.Dock && gs.DockActive(q) && gs.Levels[p.ID+".drones"] == 0 {
			for _, sp := range p.Spots {
				if sp.Shelf != nil && p.SpotVisible(sp) {
					u := p.SpotUpgrade(sp)
					if gs.Money >= u.Cost(0) {
						if sp.Zone.Has(pl.Pos) {
							return "Reste immobile sur la case : l'argent s'y verse…", V{}, false
						}
						return "Achète un drone : il portera les produits de son étal jusqu'au quai du camion.", sp.Zone.C(), true
					}
				}
			}
		}
	}
	if p.ID != "ferme" {
		return g.superHint(p)
	}
	// une colonne de la serre vide, sans ouvrier pour la replanter
	for _, f := range p.Stations {
		if !f.Built || !f.Replant || gs.Levels["ferme.ouvriers_"+f.ID] > 0 || p.Empties(f.Seed) < len(f.Plants) {
			continue
		}
		c := p.Station("graines_" + f.ID)
		if c.Zone.Has(pl.Pos) {
			return "Prends les sachets, puis marche sur les cases vides.", V{}, false
		}
		return "La colonne « " + f.Name + " » est vide : prends des sachets de graines à son pied.", c.Zone.C(), true
	}
	bleShelf := p.Station("etal_1") // l'étal du maïs
	if gs.Levels["ferme.caissiere"] == 0 && gs.Money >= 150 {
		z := p.Spots[0].Zone
		if z.Has(pl.Pos) {
			return "Reste immobile sur la case : l'argent s'y verse…", V{}, false
		}
		return "Embauche une caissière : arrête-toi sur sa case, à gauche de la caisse.", z.C(), true
	}
	if gs.Levels["ferme.ouvriers"] == 0 && gs.Money >= 250 && gs.Levels["ferme.caissiere"] > 0 {
		for _, sp := range p.Spots {
			if sp.Team != "" && p.SpotVisible(sp) {
				z := sp.Zone
				if z.Has(pl.Pos) {
					return "Reste immobile sur la case : l'argent s'y verse…", V{}, false
				}
				return "Embauche un ouvrier du maïs : sa case est sous le champ de maïs.", z.C(), true
			}
		}
	}
	if p.Served < 3 {
		if len(pl.Carry) == 0 {
			if bleShelf.Stock < 6 {
				return "Marche dans le champ de maïs pour récolter.", p.Station("mais").Body.C(), true
			}
		} else if bleShelf.Stock < bleShelf.StockMax {
			return "Dépose le maïs sur l'étal (la zone au-dessus du comptoir).", bleShelf.Zone.C(), true
		}
	}
	for _, q := range p.Stations {
		if q.Kind == sim.Dock && gs.DockActive(q) && gs.Stats.Delivered == 0 {
			return "Dépose des produits sur le quai du camion (au bord de la route) : il les emmène.", q.Zone.C(), true
		}
	}
	return "", V{}, false
}

func (g *game) drawGuide(dst *ebiten.Image, p *sim.Place) {
	_, tgt, ok := g.hint()
	if !ok {
		return
	}
	pl := g.sim.Player.Pos
	d := tgt.Sub(pl)
	if d.Len() < 90 {
		return
	}
	n := d.Norm()
	// flèche autour du joueur
	c := pl.Add(n.Mul(70 + 8*math.Sin(g.t*6)))
	side := V{X: -n.Y, Y: n.X}
	polygon(dst, []V{c.Add(n.Mul(22)), c.Add(side.Mul(16)).Sub(n.Mul(8)), c.Sub(side.Mul(16)).Sub(n.Mul(8))}, rgba(0xffe066, 0.95))
	// flèche qui rebondit au-dessus de la cible
	by := tgt.Y - 90 - 14*math.Abs(math.Sin(g.t*4))
	polygon(dst, []V{{X: tgt.X - 24, Y: by - 30}, {X: tgt.X + 24, Y: by - 30}, {X: tgt.X, Y: by}}, rgb(0xffe066))
	polygon(dst, []V{{X: tgt.X - 12, Y: by - 60}, {X: tgt.X + 12, Y: by - 60}, {X: tgt.X + 12, Y: by - 30}, {X: tgt.X - 12, Y: by - 30}}, rgb(0xffe066))
}

// ---------------------------------------------------------------- HUD

func (g *game) moneyBox(dst *ebiten.Image, showRate bool) {
	roundRect(dst, 20, 18, 380, 92, 22, rgba(0x2b2140, 0.82))
	coin(dst, 66, 56, 24)
	sayFit(dst, money(g.sim.Money), 104, 52, 42, true, -1, 280, rgb(0xffe066))
	if showRate {
		say(dst, "+"+ftoa(g.sim.Rate(), 1)+" € / s", 106, 90, 20, false, -1, rgb(0xd8d0f0))
	}
}

// btn dessine un bouton et l'enregistre pour les clics.
func (g *game) btn(dst *ebiten.Image, label string, x, y, w, h float64, col uint32, enabled bool, fn func()) {
	c := col
	if !enabled {
		c = 0x8a8a8a
	}
	hv := g.hover(x, y, w, h) && enabled
	roundRect(dst, x, y+5, w, h, h*0.3, shade(c, 0.6))
	yy := y
	if hv {
		yy -= 2
	}
	roundRect(dst, x, yy, w, h, h*0.3, rgb(c))
	if hv {
		roundRect(dst, x, yy, w, h, h*0.3, rgba(0xffffff, 0.15))
	}
	sayFit(dst, label, x+w/2, yy+h/2, math.Min(30, h*0.45), true, 0, w-20, rgb(0xffffff))
	if enabled {
		g.addButton(x, y, w, h, fn)
		if label == "Quitter" || label == "Oui, tout effacer" {
			g.btns[len(g.btns)-1].danger = true
		}
	}
}

func (g *game) drawHUD(dst *ebiten.Image) {
	g.moneyBox(dst, true)
	x := scr.VW - 20
	x -= 90
	g.soundButton(dst, x, 24)
	x -= 250
	g.btn(dst, "Ville  [Tab]", x, 24, 236, 70, 0x5b7fd6, true, func() { g.goMap() })
	x -= 230
	g.btn(dst, "Badges  [B]", x, 24, 216, 70, 0xd4a017, true, func() { g.openPanel(panelBadges) })
	x -= 90
	g.btn(dst, "?", x, 24, 76, 70, 0x9b6bd6, true, func() { g.openPanel(panelHelp) })
	// raccourci : la mairie (il s'illumine quand on peut acheter)
	x -= 90
	g.roundShortcut(dst, x, 24, g.canBuyMairie() || g.hasPoints(), func(cx, cy float64) {
		// petite mairie
		rect(dst, cx-18, cy-4, 36, 22, rgb(0x5b7fd6))
		polygon(dst, []V{{X: cx - 24, Y: cy - 2}, {X: cx, Y: cy - 20}, {X: cx + 24, Y: cy - 2}}, rgb(0x2b4a8a))
		rect(dst, cx-5, cy+6, 10, 12, rgb(0x6e4322))
		rect(dst, cx-1, cy-34, 2, 16, rgb(0x555555))
		polygon(dst, []V{{X: cx + 1, Y: cy - 34}, {X: cx + 14, Y: cy - 30}, {X: cx + 1, Y: cy - 26}}, rgb(0x2045a8))
	}, func() { g.openPanel(panelMairie) })
	// (le personnage s'ouvre au vestiaire, dans la zone du joueur, ou avec P)
	g.repBox(dst, g.sim.Cur, 416, 18) // à droite de l'argent, pour ne pas cacher les champs
	g.drawGoal(dst, 672, 18)
	g.drawOrderHUD(dst, 20, 124)

	// alerte : clients en attente sans caissière
	p := g.sim.Cur
	// conseil du moment
	if txt, _, _ := g.hint(); txt != "" {
		w := textWidth(txt, 30, true) + 70
		y := scr.VH - 110
		roundRect(dst, scr.VW/2-w/2, y, w, 70, 35, rgba(0x2b2140, 0.85))
		say(dst, txt, scr.VW/2, y+35, 30, true, 0, rgb(0xffffff))
	}
	// petit rappel des contrôles
	if g.sim.Stats.Played < 240 {
		say(dst, "Se déplacer : flèches, ZQSD ou clic maintenu · Zoom : + et - (ou molette) · F11 : plein écran", scr.VW/2, scr.VH-24, 20, false, 0, rgba(0x1e2a12, 0.8))
	}
	_ = p
}

func (g *game) soundButton(dst *ebiten.Image, x, y float64) {
	c := uint32(0x6a6a8a)
	g.btn(dst, "", x, y, 76, 70, c, true, func() { g.toggleMute() })
	// haut-parleur
	cx, cy := x+30, y+35
	rect(dst, cx-14, cy-8, 10, 16, rgb(0xffffff))
	polygon(dst, []V{{X: cx - 6, Y: cy - 8}, {X: cx + 6, Y: cy - 18}, {X: cx + 6, Y: cy + 18}, {X: cx - 6, Y: cy + 8}}, rgb(0xffffff))
	if g.snd != nil && g.snd.muted {
		line(dst, cx+12, cy-10, cx+28, cy+10, 4, rgb(0xff6b6b))
		line(dst, cx+28, cy-10, cx+12, cy+10, 4, rgb(0xff6b6b))
	} else {
		ring(dst, cx+8, cy, 12, 3, rgb(0xffffff))
		rect(dst, cx-4, cy-20, 12, 40, rgb(c))
	}
}

// ---------------------------------------------------------------- panneaux

// panelFrame dessine le cadre d'une fenêtre et renvoie la zone intérieure.
func (g *game) panelFrame(dst *ebiten.Image, title string, w, h float64) (x, y float64) {
	g.panelBtn0 = len(g.btns) // la main ne vise que les boutons de la fenêtre
	rect(dst, 0, 0, scr.VW, scr.VH, rgba(0x000000, 0.55))
	x = scr.VW/2 - w/2
	y = scr.VH/2 - h/2
	roundRect(dst, x, y+8, w, h, 30, rgb(0x5a4a3a))
	roundRect(dst, x, y, w, h, 30, rgb(0xfff8ea))
	roundRect(dst, x, y, w, 90, 30, rgb(0x8a5a33))
	rect(dst, x, y+60, w, 30, rgb(0x8a5a33))
	say(dst, title, x+w/2, y+46, 44, true, 0, rgb(0xffeab0))
	g.btn(dst, "X", x+w-80, y+15, 60, 60, 0xd9534f, true, func() { g.closePanel() })
	// clic hors de la fenêtre : rien (on bloque)
	return x, y + 90
}

type row struct {
	title, desc string
	lvl, max    int
	cost        float64
	label       string // texte du bouton (sinon le prix)
	ok          bool   // achetable
	icon        func(d *ebiten.Image, x, y float64)
	fn          func()
	header      bool
}

// drawRows dessine une liste défilante de lignes.
func (g *game) drawRows(dst *ebiten.Image, x, y, w, h float64, rows []row) {
	const rh = 118.0
	total := 0.0
	for _, r := range rows {
		if r.header {
			total += 60
		} else {
			total += rh
		}
	}
	maxS := math.Max(0, total-h+20)
	g.scroll = math.Max(0, math.Min(maxS, g.scroll))
	sub := dst.SubImage(rectI(x, y, w, h)).(*ebiten.Image)
	yy := y + 10 - g.scroll
	for _, r := range rows {
		if r.header {
			say(sub, r.title, x+40, yy+32, 28, true, -1, rgb(0x8a5a33))
			yy += 60
			continue
		}
		if yy > y+h || yy+rh < y {
			yy += rh
			continue
		}
		roundRect(sub, x+24, yy, w-48, rh-12, 18, rgb(0xf3e6cc))
		if r.icon != nil {
			roundRect(sub, x+38, yy+12, 82, 82, 16, rgb(0xe6d3ae))
			r.icon(sub, x+79, yy+53)
		}
		tx := x + 140
		say(sub, r.title, tx, yy+30, 28, true, -1, rgb(0x3b2a10))
		lines := wrap(r.desc, 21, w-560, false)
		for i, l := range lines {
			if i < 2 {
				say(sub, l, tx, yy+62+float64(i)*24, 21, false, -1, rgb(0x5a4a3a))
			}
		}
		// niveaux
		if r.max > 1 {
			for k := 0; k < r.max; k++ {
				c := rgb(0xd6c6a6)
				if k < r.lvl {
					c = rgb(0xffb400)
				}
				roundRect(sub, x+w-400+float64(k)*(min(26, 200/float64(r.max))), yy+20, min(20, 200/float64(r.max)-4), 14, 5, c)
			}
		}
		label := r.label
		if label == "" {
			label = money(r.cost)
		}
		bx, by, bw, bh := x+w-250, yy+26, 200.0, 64.0
		visible := by >= y && by+bh <= y+h
		col := uint32(0x2fa86b)
		if !r.ok {
			col = 0xb0a890
		}
		if r.lvl >= r.max && r.max > 0 {
			label = "MAX"
			col = 0xb0a890
		}
		c := col
		roundRect(sub, bx, by+5, bw, bh, 18, shade(c, 0.6))
		hv := g.hover(bx, by, bw, bh) && r.ok && visible
		roundRect(sub, bx, by, bw, bh, 18, rgb(c))
		if hv {
			roundRect(sub, bx, by, bw, bh, 18, rgba(0xffffff, 0.18))
		}
		if r.label == "" && label != "MAX" {
			coin(sub, bx+30, by+bh/2, 13)
			sayFit(sub, label, bx+bw/2+14, by+bh/2, 26, true, 0, bw-60, rgb(0xffffff))
		} else {
			sayFit(sub, label, bx+bw/2, by+bh/2, 26, true, 0, bw-20, rgb(0xffffff))
		}
		if r.ok && visible && r.fn != nil {
			fn := r.fn
			g.addButton(bx, by, bw, bh, fn)
		}
		yy += rh
	}
	// barre de défilement
	if maxS > 0 {
		bh := h * h / total
		by := y + (h-bh)*g.scroll/maxS
		roundRect(dst, x+w-16, by, 8, bh, 4, rgba(0x8a5a33, 0.5))
	}
}

func rectI(x, y, w, h float64) imageRect {
	return imageRect{Min: imagePt{int(cam.X(x)), int(cam.Y(y))}, Max: imagePt{int(cam.X(x + w)), int(cam.Y(y + h))}}
}

func (g *game) upgradeRow(u *sim.Upgrade, icon func(d *ebiten.Image, x, y float64)) row {
	l := g.sim.Levels[u.Key]
	r := row{title: u.Name, lvl: l, max: u.Max, icon: icon}
	if l < u.Max {
		r.cost = u.Cost(l)
		r.ok = g.sim.Money >= r.cost
		r.desc = u.Desc(l + 1)
		if u.Now {
			r.desc = u.Desc(l)
		}
		uu := u
		r.fn = func() {
			if g.sim.Buy(uu) {
				g.snd.play(sBuy)
			}
		}
	} else {
		r.desc = u.Desc(l)
	}
	return r
}

func (g *game) drawUpgradePanel(dst *ebiten.Image) {
	p := g.sim.Cur
	if p == nil {
		p = g.sim.Places[0]
	}
	w, h := math.Min(1300, scr.VW-80), 900.0
	x, y := g.panelFrame(dst, "Améliorations de la ferme", w, h)
	ups := p.PlaceUpgrades()
	person := func(lk look) func(d *ebiten.Image, x, y float64) {
		return func(d *ebiten.Image, x, y float64) {
			scaled(x, y, 0.72, func() { drawPerson(d, V{X: x, Y: y + 46}, V{Y: 1}, 0, false, lk, nil) })
		}
	}
	worker := staffLook(p.ID, sim.WorkerAgent, 0)
	cashier := staffLook(p.ID, sim.CashierAgent, 0)
	var rows []row
	rows = append(rows, row{header: true, title: "Personnel"})
	rows = append(rows, g.upgradeRow(ups[0], person(cashier)))
	rows = append(rows, g.upgradeRow(ups[1], person(worker)))
	rows = append(rows, g.upgradeRow(ups[2], func(dst *ebiten.Image, x, y float64) {
		roundRect(dst, x-22, y-10, 20, 34, 6, rgb(0x6b4423))
		roundRect(dst, x+2, y-10, 20, 34, 6, rgb(0x6b4423))
		roundRect(dst, x-26, y+16, 26, 12, 5, rgb(0x4a2f18))
		roundRect(dst, x+2, y+16, 26, 12, 5, rgb(0x4a2f18))
	}))
	rows = append(rows, g.upgradeRow(ups[3], func(dst *ebiten.Image, x, y float64) {
		polygon(dst, []V{{X: x - 30, Y: y - 12}, {X: x + 24, Y: y - 12}, {X: x + 16, Y: y + 12}, {X: x - 22, Y: y + 12}}, rgb(0x6f8fa8))
		circle(dst, x+20, y+20, 10, rgb(0x333333))
		line(dst, x-30, y-12, x-40, y+18, 4, rgb(0x8a5a33))
		drawItem(dst, sim.Ble, x-4, y-20, 14)
	}))
	rows = append(rows, row{header: true, title: "Ventes"})
	rows = append(rows, g.upgradeRow(ups[4], func(dst *ebiten.Image, x, y float64) {
		rect(dst, x-4, y-6, 8, 40, rgb(0x6e4322))
		roundRect(dst, x-34, y-30, 68, 34, 6, rgb(0xf3ead8))
		sayFit(dst, "MA FERME", x, y-13, 15, true, 0, 62, rgb(0x6e4322))
	}))
	rows = append(rows, g.upgradeRow(ups[5], func(dst *ebiten.Image, x, y float64) {
		circle(dst, x, y, 30, rgb(0x2fa86b))
		say(dst, "BIO", x, y, 22, true, 0, rgb(0xffffff))
	}))
	g.drawRows(dst, x, y, w, h-100, rows)
}

// drawMairiePanel : la mairie, c'est la page des points de l'école (avec le
// veilleur de nuit en bandeau, tout en haut).
func (g *game) drawMairiePanel(dst *ebiten.Image) { g.drawPointsPanel(dst) }

// veilleurBand : le veilleur de nuit, sur une ligne.
func (g *game) veilleurBand(dst *ebiten.Image, x, y, w float64) {
	u := sim.MairieUpgrades()[0]
	l := g.sim.Levels[u.Key]
	roundRect(dst, x, y, w, 84, 16, rgb(0xe8eef8))
	cx, cy := x+50, y+42
	circle(dst, cx, cy, 26, rgb(0x2b2b5a))
	circle(dst, cx+9, cy-7, 19, rgb(0xe8eef8))
	star(dst, cx-12, cy+10, 6, rgb(0xffe066))
	say(dst, "Veilleur de nuit", x+95, y+26, 24, true, -1, rgb(0x2b3a5a))
	cur := "Tes employés travaillent pendant ton absence : " + itoa(int(sim.NightHours[min(l, len(sim.NightHours)-1)])) + " h maximum (il faut une caissière)."
	if l < u.Max {
		cur += "  Niveau suivant : " + itoa(int(sim.NightHours[l+1])) + " h."
	}
	sayFit(dst, cur, x+95, y+58, 17, false, -1, w-400, rgb(0x4a5a7a))
	for k := 0; k < u.Max; k++ {
		c := rgba(0x2b3a5a, 0.2)
		if k < l {
			c = rgb(0x5b7fd6)
		}
		roundRect(dst, x+w-330+float64(k)*20, y+34, 14, 14, 4, c)
	}
	if l >= u.Max {
		say(dst, "Maximum", x+w-110, y+42, 22, true, 0, rgb(0x2fa86b))
		return
	}
	cost := u.Cost(l)
	g.btn(dst, money(cost), x+w-210, y+14, 190, 56, 0x2fa86b, g.sim.Money >= cost, func() {
		if g.sim.Buy(u) {
			g.snd.play(sBuy)
		}
	})
}

func (g *game) drawWardrobePanel(dst *ebiten.Image) { g.drawPersoPanel(dst) }

func (g *game) drawWelcome(dst *ebiten.Image) {
	w, h := 900.0, 470.0
	x, y := g.panelFrame(dst, "Bon retour à Clairbourg !", w, h)
	secs := g.welcome.secs
	hh := int(secs) / 3600
	mm := int(secs) / 60 % 60
	dur := itoa(mm) + " min"
	if hh > 0 {
		dur = itoa(hh) + " h " + itoa(mm)
	}
	say(dst, "Pendant ton absence ("+dur+") :", x+w/2, y+60, 30, false, 0, rgb(0x5a4a3a))
	if g.welcome.amount > 0 {
		coin(dst, x+w/2-textWidth("+"+money(g.welcome.amount), 56, true)/2-36, y+140, 26)
		say(dst, "+"+money(g.welcome.amount), x+w/2+14, y+140, 56, true, 0, rgb(0x2fa86b))
		say(dst, "Tes employés ont bien travaillé.", x+w/2, y+210, 26, false, 0, rgb(0x5a4a3a))
	} else {
		for i, l := range wrap("Personne n'a encaissé les clients. Embauche une caissière pour gagner de l'argent même quand le jeu est fermé.", 28, w-120, false) {
			say(dst, l, x+w/2, y+130+float64(i)*38, 28, false, 0, rgb(0x5a4a3a))
		}
	}
	if g.welcome.points > 0 {
		say(dst, "L'école a formé "+itoa(g.welcome.points)+" élève(s) : des points à répartir à la mairie.", x+w/2, y+262, 22, true, 0, rgb(0x8a6a10))
	}
	g.btn(dst, "Super !", x+w/2-130, y+h-180, 260, 76, 0x2fa86b, true, func() { g.closePanel() })
}

func (g *game) drawHelp(dst *ebiten.Image) {
	if g.sim.Cur != nil && g.sim.Cur.ID == "ecole" && g.scene == scenePlace {
		g.drawSchoolHelp(dst)
		return
	}
	w, h := math.Min(1300, scr.VW-80), 920.0
	x, y := g.panelFrame(dst, "Comment jouer", w, h)
	lines := []string{
		"• Déplace-toi avec les flèches (ou ZQSD, ou la souris). Carte et menus : une main suit les flèches ; reste sur un bouton (ou Entrée).",
		"• Marche dans un champ mûr pour récolter. Ce que tu portes s'empile dans tes bras.",
		"• Chaque champ commence avec une seule pousse : achète les suivantes sur la case « +1 ».",
		"• Dépose la marchandise sur la zone claire au-dessus d'un étal : les clients se servent.",
		"• Les clients paient à la caisse… s'il y a quelqu'un pour encaisser !",
		"• Bâtiments (moulins, poulailler, étable…) : dépose les ingrédients à gauche (bleu), ramasse à droite.",
		"• Arrête-toi sur une zone à vendre (pointillés jaunes) : ton argent s'y verse petit à petit.",
		"• À côté des bâtiments : « Niveau » (plus rapide) et « Stock + » (stocks plus grands).",
		"• Chaque zone a ses ouvriers (case sous le champ). Les quais des camions : un drone par étal, ou toi.",
		"• L'école (offerte) forme des élèves : chacun qui part te donne un point, à répartir dans tes lieux depuis la mairie.",
		"• Zone grise réservée au joueur : hélicoptère vers les autres lieux (ou touches 1 à 9 et 0 : tes favoris, à choisir sur la carte ; Tab : carte de la ville), vestiaire (P), V : mairie.",
		"• Tu n'arrives pas à te débarrasser de quelque chose ? Arrête-toi au compost, tout à droite.",
		"• Supermarché, boulangerie, crèmerie, glacier (sur la carte) : chacun a son quai à la ferme, un camion y emmène les produits.",
		"• La réputation (étoiles) monte avec les clients contents et attire du monde. B : tes badges.",
		"• Le jeu se sauvegarde tout seul. F11 : plein écran ou fenêtre. M : couper le son.",
	}
	for i, l := range lines {
		say(dst, l, x+50, y+34+float64(i)*45, 22, false, -1, rgb(0x3b2a10))
	}
	g.btn(dst, "C'est parti !", x+w/2-150, y+h-190, 300, 76, 0x2fa86b, true, func() { g.closePanel() })
}

// drawSchoolHelp : l'aide de l'école (temps de formation, échecs).
func (g *game) drawSchoolHelp(dst *ebiten.Image) {
	w, h := math.Min(1300, scr.VW-80), 920.0
	x, y := g.panelFrame(dst, "L'école", w, h)
	roundRect(dst, x+40, y+20, w-80, 110, 18, rgb(0xd9534f))
	say(dst, "L'école ne fonctionne que quand le jeu est ouvert.", x+w/2, y+58, 34, true, 0, rgb(0xffffff))
	say(dst, "Hors ligne, tes lieux ne te rapportent que de l'argent : pas d'élèves, pas de points.", x+w/2, y+102, 22, false, 0, rgb(0xffffff))
	lines := []string{
		"• Les élèves arrivent à l'accueil : accompagne-les jusqu'à une table libre d'une classe qui a un prof.",
		"• Chaque élève qui finit sa formation te donne un point du niveau de sa classe.",
		"• Primaire → points Ouvrier, Collège → Marketing, Lycée → Comptable, Université → Formation.",
		"• Les points se répartissent à la mairie (V), lieu par lieu, onglet par onglet.",
		"• Les fournitures (étagères) font travailler les classes plus vite.",
	}
	for i, l := range lines {
		say(dst, l, x+50, y+170+float64(i)*42, 22, false, -1, rgb(0x3b2a10))
	}
	ty := y + 400
	say(dst, "Durée d'une formation aujourd'hui", x+w/2, ty, 26, true, 0, rgb(0x5a3a1a))
	cw := (w - 100) / 4
	for i := range sim.PointTypes {
		cx := x + 50 + float64(i)*cw
		roundRect(dst, cx+6, ty+26, cw-12, 90, 14, rgba(levelColUI(i), 0.3))
		say(dst, pointTabs[i].name, cx+cw/2, ty+52, 22, true, 0, rgb(0x3b2a10))
		say(dst, clock(g.sim.FormationTime(i)), cx+cw/2, ty+88, 26, true, 0, rgb(0x3b2a10))
	}
	f := g.sim.FailChance()
	say(dst, "Échec : 1 élève sur "+itoa(int(math.Round(1/f)))+" rate sa formation (il repart avec un point du niveau d'en dessous, rien au primaire).", x+w/2, ty+160, 21, false, 0, rgb(0x7a3a2a))
	say(dst, "La Pédagogie (points Université) réduit les échecs ; la Formation accélère toutes les classes.", x+w/2, ty+200, 21, false, 0, rgb(0x5a4a3a))
	g.btn(dst, "Compris !", x+w/2-150, y+h-190, 300, 76, 0x2fa86b, true, func() { g.closePanel() })
}

// ---------------------------------------------------------------- carte

type tile struct {
	id, name, status string
	place            string
	open             bool
	col              uint32
	fn               func()
}

func (g *game) drawMap(dst *ebiten.Image) {
	t := g.t
	// fond : prairie et rivière
	rect(dst, 0, 0, scr.VW, scr.VH, rgb(0x9fd27a))
	for i := 0; i < 60; i++ {
		x := math.Mod(float64(i)*397.3, scr.VW)
		y := math.Mod(float64(i)*211.7, scr.VH)
		circle(dst, x, y, 30+math.Mod(float64(i)*13.1, 40), rgba(0x8cc56a, 0.6))
	}
	var pts []V
	for k := 0; k <= 40; k++ {
		y := float64(k) / 40 * scr.VH
		pts = append(pts, V{X: scr.VW - 140 + math.Sin(y/120+t*0.2)*30, Y: y})
	}
	for k := 40; k >= 0; k-- {
		y := float64(k) / 40 * scr.VH
		pts = append(pts, V{X: scr.VW - 40 + math.Sin(y/120+t*0.2)*30, Y: y})
	}
	polygon(dst, pts, rgb(0x6fb7e6))

	sayOutlined(dst, "Clairbourg", scr.VW/2, 70, 72, 0, rgb(0xffffff), rgb(0x3b6a2a))
	g.moneyBox(dst, true)
	farm := g.sim.Places[0]

	// à gauche : la mairie et le personnage ; à droite : les lieux
	side := []tile{
		{id: "mairie", name: "Mairie", open: true, col: 0x5b7fd6, fn: func() { g.openPanel(panelMairie) }},
		{id: "vetements", name: "Personnage", open: true, col: 0x9b6bd6, fn: func() { g.openPanel(panelWardrobe) }},
	}
	tiles := []tile{
		{id: "ferme", name: "Ma ferme", open: true, col: 0xc0623a, fn: func() { g.goPlace(farm) }},
		g.placeTile("ecole", "École", 0xf1c40f),
		g.superTile(),
		g.placeTile("cremerie", "Crèmerie", 0x7fb3d5),
		g.placeTile("exotique", "Ferme exotique", 0x2e9e6a),
		g.placeTile("boulangerie", "Boulangerie", 0xc0623a),
		g.placeTile("glacier", "Glacier", 0x5aa0e0),
		g.placeTile("port", "Port", 0x2b6cd6),
		g.placeTile("elevage", "Ferme d'élevage", 0x8a5a2b),
		g.placeTile("boucherie", "Boucherie", 0xb03a3a),
		{id: "caviste", name: "Caviste", status: "Bientôt", col: 0x7a2a4a},
		{id: "patisserie", name: "Pâtisserie", status: "Bientôt", col: 0xe0567f},
		{id: "creperie", name: "Crêperie", status: "Bientôt", col: 0xd9a55a},
		{id: "fleuriste", name: "Fleuriste", status: "Bientôt", col: 0xff6fa8},
		{id: "restaurant", name: "Restaurant", status: "Bientôt", col: 0xe67e22},
		{id: "hotel", name: "Hôtel", status: "Bientôt", col: 0x8e44ad},
		{id: "zoo", name: "Zoo", status: "Bientôt", col: 0x7a9a2a},
		{id: "hopital", name: "Hôpital", status: "Bientôt", col: 0xe74c3c},
		{id: "prison", name: "Prison", status: "Bientôt", col: 0x5d6d7e},
		{id: "parc", name: "Parc d'attractions", status: "Bientôt", col: 0xff8a3d},
	}
	cols := (len(tiles) + 2) / 3 // trois rangées
	gap := 26.0
	sep := 70.0
	tw := math.Min(300, (scr.VW-200-sep-float64(cols-1)*gap)/float64(cols+1))
	th := tw * 236 / 300
	gw := tw + sep + float64(cols)*tw + float64(cols-1)*gap
	x0 := (scr.VW-160)/2 - gw/2
	y0 := 150.0
	g.mapHover = ""
	draw := func(tl tile, x, y float64) {
		id := tl.place
		if tl.id == "ferme" {
			id = "ferme"
		}
		if id != "" && g.hover(x, y, tw, th) {
			g.mapHover = id // (un chiffre en fait un favori)
		}
		hot := (tl.id == "mairie" && (g.canBuyMairie() || g.hasPoints())) || (tl.id == "vetements" && g.canBuyGear()) || (tl.id == "ferme" && g.canBuyPlace(farm))
		if pl := g.sim.Place(tl.place); pl != nil && tl.place != "ferme" {
			hot = (!pl.Owned && g.sim.Money >= g.sim.PriceOf(pl)) || (pl.Owned && g.canBuyPlace(pl))
		}
		if hot {
			glow(dst, x, y, tw, th, g.t)
		}
		g.drawTile(dst, tl, x, y, tw, th, farm)
		if hot {
			badge(dst, x+tw-8, y+8, g.t)
		}
		if k := g.sim.FavKey(id); id != "" && k != "" { // sa touche de favori
			by := y + th*160/236 - 46 // (juste au-dessus du nom, sous les étoiles)
			roundRect(dst, x+8, by, 40, 40, 10, rgba(0x2b2140, 0.85))
			say(dst, k, x+28, by+21, 24, true, 0, rgb(0xffe066))
		}
		if o := g.sim.Order; o != nil && (o.Place() == tl.place || o.Place() == tl.id) {
			// la commande en cours est à livrer ici
			k := 0.5 + 0.5*math.Sin(g.t*4)
			roundRect(dst, x+tw/2-tw*0.3, y+th-30-3*k, tw*0.6, 36, 12, rgb(0xffd23f))
			sayFit(dst, "Commande à livrer !", x+tw/2, y+th-12-3*k, 16, true, 0, tw*0.58, rgb(0x5a3d00))
		}
	}
	for i, tl := range side {
		draw(tl, x0, y0+float64(i)*(th+gap))
	}
	// séparation
	lx := x0 + tw + sep/2
	rect(dst, lx-3, y0, 6, 3*th+2*gap, rgba(0x3b6a2a, 0.35))
	for i, tl := range tiles {
		c := i % cols
		r := i / cols
		draw(tl, x0+tw+sep+float64(c)*(tw+gap), y0+float64(r)*(th+gap))
	}
	g.btn(dst, "Quitter", scr.VW-260, scr.VH-100, 210, 70, 0xd9534f, true, func() {
		g.save()
		g.quit = true
	})
	g.soundButton(dst, scr.VW-360, scr.VH-100)
	g.btn(dst, "Recommencer à zéro", 50, scr.VH-100, 330, 70, 0x8a7a6a, true, func() { g.openPanel(panelReset) })
	g.btn(dst, "Badges", scr.VW-600, scr.VH-100, 220, 70, 0xd4a017, true, func() { g.openPanel(panelBadges) })
	say(dst, "Clique sur un lieu, ou vas-y avec les flèches et reste dessus.", 420, scr.VH-62, 24, true, -1, rgba(0x1e3a12, 0.85))
	if b := g.sim.Boost; b != nil && b.Kind == "maire" { // le cadeau du maire : bien visible
		txt := "Cadeau du maire : " + strings.ToLower(b.Text()[:1]) + b.Text()[1:] + " ! Encore " + clock(b.Left)
		w := textWidth(txt, 22, true) + 70
		k := 0.5 + 0.5*math.Sin(g.t*4)
		roundRect(dst, scr.VW/2-w/2, 108-2*k, w, 38, 19, rgb(0xd9534f))
		drawGift(dst, scr.VW/2-w/2+26, 126-2*k, 10)
		say(dst, txt, scr.VW/2+14, 127-2*k, 22, true, 0, rgb(0xffffff))
	} else {
		say(dst, "Favoris : mets la main sur un lieu et appuie sur un chiffre (1 à 9, 0).", scr.VW/2, 126, 20, true, 0, rgba(0xffffff, 0.9))
	}
}

func (g *game) drawTile(dst *ebiten.Image, tl tile, x, y, w, h float64, farm *sim.Place) {
	// la tuile est dessinée en 300 × 236, puis réduite à la taille voulue
	hv := g.hover(x, y, w, h) && tl.open
	scaled(x, y, w/300, func() { g.drawTileBody(dst, tl, x, y, 300, 236, farm, hv) })
	if tl.open && tl.fn != nil {
		g.addButton(x, y, w, h, tl.fn)
	}
}

func (g *game) drawTileBody(dst *ebiten.Image, tl tile, x, y, w, h float64, farm *sim.Place, hv bool) {
	yy := y
	if hv {
		yy -= 4
	}
	roundRect(dst, x, y+8, w, h, 26, rgba(0x000000, 0.25))
	bg := uint32(0xfff8ea)
	if !tl.open {
		bg = 0xd8d4c8
	}
	roundRect(dst, x, yy, w, h, 26, rgb(bg))
	// illustration : un petit bâtiment
	bx, by := x+w/2, yy+125
	wall := tl.col
	if !tl.open {
		wall = 0xa8a8a0
	}
	switch tl.id {
	case "ferme":
		rect(dst, x+20, yy+130, w-40, 30, rgb(0xa47447))
		for k := 0; k < 6; k++ {
			drawPlant(dst, sim.Ble, x+50+float64(k)*26, yy+150, 1, g.t+float64(k), k)
		}
		building(dst, sim.Rect{X: bx + 10, Y: by - 70, W: 120, H: 90}, 0xb8352c, 0x5a4a44)
	case "vetements":
		// le personnage lui-même
		scaled(bx, by, 1.4, func() { g.drawHero(dst, V{X: bx, Y: by + 10}, V{Y: 1}, 0, false, nil) })
	default:
		building(dst, sim.Rect{X: bx - 80, Y: by - 80, W: 160, H: 100}, wall, shadeHex(wall, 0.6))
		roundRect(dst, bx-18, by-20, 36, 40, 6, rgb(0x6e4322))
	}
	if tl.id == "mairie" {
		rect(dst, bx-3, by-150, 6, 40, rgb(0x555555))
		polygon(dst, []V{{X: bx + 3, Y: by - 150}, {X: bx + 40, Y: by - 140}, {X: bx + 3, Y: by - 128}}, rgb(0x2045a8))
	}
	if tl.id == "supermarche" {
		roundRect(dst, bx-60, by-56, 120, 24, 6, rgb(0xffffff))
		say(dst, "SUPER", bx, by-44, 17, true, 0, rgb(0x2fa86b))
	}
	// nom
	roundRect(dst, x, yy+h-76, w, 76, 26, rgb(wall))
	rect(dst, x, yy+h-76, w, 30, rgb(wall))
	sayFit(dst, tl.name, x+w/2, yy+h-48, 32, true, 0, w-30, rgb(0xffffff))
	sub := tl.status
	if tl.id == "ferme" {
		sub = "+" + ftoa(g.sim.Rate(), 1) + " € / s"
		if !farm.HasCashier() && farm.QueueLen() > 0 {
			sub = itoa(farm.QueueLen()) + " client(s) attendent !"
		}
	}
	if tl.id == "mairie" {
		sub = "Points de l'école, veilleur"
		if n := g.totalPoints(); n > 0 {
			sub = itoa(n) + " point(s) à répartir !"
		}
	}
	if tl.id == "vetements" {
		sub = "Équipe ton personnage"
	}
	say(dst, sub, x+w/2, yy+h-18, 20, false, 0, rgb(0xffffff))
	if tl.id == "ferme" {
		stars(dst, x+24, yy+28, 13, farm.Rep)
	}
	if pl := g.sim.Place(tl.place); pl != nil && tl.place != "ferme" && tl.place != "ecole" && pl.Owned {
		stars(dst, x+24, yy+28, 13, pl.Rep)
	}
	if !tl.open {
		// cadenas
		lx, ly := x+w-50, yy+44
		ring(dst, lx, ly-10, 14, 6, rgb(0x6a6a6a))
		roundRect(dst, lx-20, ly-6, 40, 32, 6, rgb(0x6a6a6a))
		circle(dst, lx, ly+8, 5, rgb(0xd8d4c8))
	}
}

func shadeHex(c uint32, k float64) uint32 {
	s := shade(c, k)
	return uint32(s.R)<<16 | uint32(s.G)<<8 | uint32(s.B)
}

// scaled dessine fn réduit (ou agrandi) d'un facteur k autour de (cx, cy).
func scaled(cx, cy, k float64, fn func()) {
	saved := cam
	cam = view{s: saved.s * k, ox: saved.ox + cx*saved.s*(1-k), oy: saved.oy + cy*saved.s*(1-k)}
	fn()
	cam = saved
}

// ---------------------------------------------------------------- « achat possible »

func (g *game) affordable(u *sim.Upgrade) bool {
	l := g.sim.Levels[u.Key]
	return l < u.Max && g.sim.Money >= u.Cost(l)
}

// canBuyPlace : une amélioration du lieu est-elle à portée de bourse ?
func (g *game) canBuyPlace(p *sim.Place) bool {
	if p == nil {
		return false
	}
	// seulement ce qu'on peut vraiment acheter dans ce lieu (cases visibles)
	for _, sp := range p.Spots {
		if p.SpotVisible(sp) {
			if u := p.SpotUpgrade(sp); u != nil && g.affordable(u) {
				return true
			}
		}
	}
	for _, u := range p.Unlocks {
		if !u.Done && p.Visible(u) && u.Cost < 1e9 && g.sim.Money >= u.Cost-u.Paid {
			return true
		}
	}
	for _, s := range p.Stations {
		if !s.Built {
			continue
		}
		if s.Upgradable() && g.affordable(p.StationUpgrade(s)) {
			return true
		}
		if s.Stockable() && g.affordable(p.StockUpgrade(s)) {
			return true
		}
	}
	return false
}

func (g *game) canBuyMairie() bool {
	for _, u := range sim.MairieUpgrades() {
		if g.affordable(u) {
			return true
		}
	}
	return false
}

// glow : halo doré qui pulse autour d'un bouton.
func glow(dst *ebiten.Image, x, y, w, h, t float64) {
	a := 0.7 + 0.3*math.Sin(t*5)
	for i, d := range []float64{20, 13, 7} {
		roundRect(dst, x-d, y-d, w+2*d, h+2*d+5, h*0.3+d, rgba(0xffd23f, a*float64(i+1)/3))
	}
}

// badge : pastille rouge avec un « ! ».
func badge(dst *ebiten.Image, x, y, t float64) {
	r := 17 + 2*math.Sin(t*5)
	circle(dst, x, y, r+3, rgb(0xffffff))
	circle(dst, x, y, r, rgb(0xe74c3c))
	say(dst, "!", x, y, 24, true, 0, rgb(0xffffff))
}

// ---------------------------------------------------------------- remise à zéro

func (g *game) drawReset(dst *ebiten.Image) {
	w, h := 940.0, 480.0
	x, y := g.panelFrame(dst, "Recommencer à zéro ?", w, h)
	for i, l := range wrap("Tout sera effacé : l'argent, les champs, les bâtiments, les employés, les améliorations de la mairie et les tenues. On repart avec une seule pousse de blé.", 28, w-120, false) {
		say(dst, l, x+w/2, y+60+float64(i)*40, 28, false, 0, rgb(0x5a4a3a))
	}
	say(dst, "Impossible de revenir en arrière !", x+w/2, y+220, 28, true, 0, rgb(0xc0392b))
	g.btn(dst, "Non, je garde tout", x+70, y+h-180, 360, 76, 0x2fa86b, true, func() { g.closePanel() })
	g.btn(dst, "Oui, tout effacer", x+w-430, y+h-180, 360, 76, 0xd9534f, true, func() { g.resetGame() })
}

func (g *game) resetGame() {
	g.sim = sim.NewGame()
	g.parts = nil
	g.toasts = nil
	g.panel = panelNone
	g.scene = scenePlace
	g.camPos = g.sim.Player.Pos
	g.save()
	g.toast("Nouvelle partie !", "")
}

// ---------------------------------------------------------------- réputation

// stars dessine 5 étoiles remplies selon la note (de gauche à droite).
func stars(dst *ebiten.Image, x, y, r, note float64) {
	for i := 0; i < 5; i++ {
		cx := x + float64(i)*r*2.2 + r
		star(dst, cx, y, r+2, rgb(0x5a4a2a))
		f := math.Max(0, math.Min(1, note-float64(i)))
		star(dst, cx, y, r, rgb(0xd8d0c0))
		if f > 0 {
			// étoile remplie en partie : on la découpe
			sub := dst.SubImage(rectI(cx-r, y-r-2, 2*r*f, 2*r+4)).(*ebiten.Image)
			star(sub, cx, y, r, rgb(0xffc21a))
			flush()
		}
	}
}

func (g *game) repBox(dst *ebiten.Image, p *sim.Place, x, y float64) {
	if p == nil {
		return
	}
	roundRect(dst, x, y, 240, 92, 22, rgba(0x2b2140, 0.82))
	if p.ID == "ecole" {
		say(dst, "Élèves formés", x+120, y+28, 20, false, 0, rgb(0xd8d0f0))
		say(dst, itoa(g.sim.Levels["ecole.diplomes"]), x+120, y+62, 30, true, 0, rgb(0xffe066))
		return
	}
	say(dst, "Réputation", x+120, y+28, 20, false, 0, rgb(0xd8d0f0))
	stars(dst, x+120-5*2.2*17/2, y+62, 17, p.Rep)
}

// ---------------------------------------------------------------- badges

func (g *game) drawBadges(dst *ebiten.Image) {
	w, h := math.Min(1500, scr.VW-80), 940.0
	x, y := g.panelFrame(dst, "Badges", w, h)
	sub := dst.SubImage(rectI(x, y, w, h-90)).(*ebiten.Image)
	cols := 4
	cw := (w - 80) / float64(cols)
	ch := 150.0
	total := 0.0
	type sect struct{ id, title string }
	sects := []sect{{"ferme", "Ma ferme"}, {"super", "Le supermarché"}, {"boulangerie", "La boulangerie"}, {"cremerie", "La crèmerie"}, {"glacier", "Le glacier"}, {"exotique", "La ferme exotique"}, {"elevage", "La ferme d'élevage"}, {"boucherie", "La boucherie"}, {"ecole", "L'école"}, {"ville", "Clairbourg"}}
	for _, sc := range sects {
		n := 0
		for _, b := range sim.Badges {
			if b.Place == sc.id {
				n++
			}
		}
		total += 70 + math.Ceil(float64(n)/float64(cols))*ch
	}
	maxS := math.Max(0, total-(h-110))
	g.scroll = math.Max(0, math.Min(maxS, g.scroll))
	yy := y + 10 - g.scroll
	for _, sc := range sects {
		var list []sim.Badge
		got := 0
		for _, b := range sim.Badges {
			if b.Place == sc.id {
				list = append(list, b)
				if g.sim.Badges[b.ID] {
					got++
				}
			}
		}
		say(sub, sc.title+"  ("+itoa(got)+"/"+itoa(len(list))+")", x+44, yy+36, 30, true, -1, rgb(0x8a5a33))
		yy += 70
		for i, b := range list {
			cx := x + 40 + float64(i%cols)*cw
			cy := yy + float64(i/cols)*ch
			if cy > y+h || cy+ch < y {
				continue
			}
			ok := g.sim.Badges[b.ID]
			bg := uint32(0xeee4cf)
			if ok {
				bg = 0xfff0c2
			}
			roundRect(sub, cx+6, cy+6, cw-12, ch-12, 18, rgb(bg))
			medal(sub, b.Icon, cx+66, cy+ch/2, ok)
			tc := uint32(0x3b2a10)
			if !ok {
				tc = 0x8a7d66
			}
			sayFit(sub, b.Name, cx+118, cy+44, 22, true, -1, cw-140, rgb(tc))
			for k, l := range wrap(b.Desc, 17, cw-140, false) {
				if k < 3 {
					say(sub, l, cx+118, cy+76+float64(k)*22, 17, false, -1, rgb(0x6a5a44))
				}
			}
		}
		yy += math.Ceil(float64(len(list))/float64(cols)) * ch
	}
	flush()
}

// medal dessine une médaille (dorée si gagnée, grise sinon).
func medal(dst *ebiten.Image, icon int, x, y float64, ok bool) {
	rib1, rib2, disc, rim := uint32(0x3b6ea8), uint32(0xc0392b), uint32(0xffd23f), uint32(0xc99a1a)
	if !ok {
		rib1, rib2, disc, rim = 0xb0a890, 0xa09880, 0xd8d0c0, 0xb8b0a0
	}
	polygon(dst, []V{{X: x - 22, Y: y - 50}, {X: x - 4, Y: y - 50}, {X: x + 6, Y: y - 10}, {X: x - 12, Y: y - 10}}, rgb(rib1))
	polygon(dst, []V{{X: x + 4, Y: y - 50}, {X: x + 22, Y: y - 50}, {X: x + 12, Y: y - 10}, {X: x - 6, Y: y - 10}}, rgb(rib2))
	circle(dst, x, y+8, 38, rgb(rim))
	circle(dst, x, y+8, 32, rgb(disc))
	if !ok {
		say(dst, "?", x, y+8, 34, true, 0, rgb(0xa09880))
		return
	}
	cy := y + 8
	switch icon {
	case sim.IconWheat:
		drawItem(dst, sim.Ble, x, cy, 20)
	case sim.IconCorn:
		drawItem(dst, sim.Mais, x, cy, 20)
	case sim.IconVeg:
		drawItem(dst, sim.Carotte, x, cy, 20)
	case sim.IconFlower:
		drawItem(dst, sim.Fleur, x, cy, 20)
	case sim.IconCane:
		drawItem(dst, sim.Canne, x, cy, 20)
	case sim.IconMill:
		drawItem(dst, sim.Farine, x, cy, 20)
	case sim.IconBread:
		drawItem(dst, sim.Pain, x, cy, 20)
	case sim.IconEgg:
		drawItem(dst, sim.Oeuf, x, cy, 20)
	case sim.IconMilk:
		drawItem(dst, sim.Lait, x, cy, 20)
	case sim.IconHoney:
		drawItem(dst, sim.Miel, x, cy, 20)
	case sim.IconSausage:
		drawItem(dst, sim.Saucisse, x, cy, 20)
	case sim.IconFruit:
		drawItem(dst, sim.Pomme, x, cy, 20)
	case sim.IconChoco:
		drawItem(dst, sim.Chocolat, x, cy, 20)
	case sim.IconIce:
		drawItem(dst, sim.GlaceFraise, x, cy, 20)
	case sim.IconCheese:
		drawItem(dst, sim.Fromage, x, cy, 20)
	case sim.IconCake:
		drawItem(dst, sim.Croissant, x, cy, 20)
	case sim.IconCoin:
		coin(dst, x, cy, 18)
	case sim.IconCart:
		roundRect(dst, x-16, cy-12, 30, 18, 4, rgb(0x7f93a8))
		line(dst, x-16, cy-12, x-22, cy-18, 3, rgb(0x555555))
		circle(dst, x-10, cy+12, 4, rgb(0x333333))
		circle(dst, x+8, cy+12, 4, rgb(0x333333))
	case sim.IconTruck:
		scaled(x, cy, 0.16, func() { drawTruck(dst, x, cy, false) })
	case sim.IconHeart:
		circle(dst, x-8, cy-5, 10, rgb(0xe74c3c))
		circle(dst, x+8, cy-5, 10, rgb(0xe74c3c))
		polygon(dst, []V{{X: x - 17, Y: cy}, {X: x + 17, Y: cy}, {X: x, Y: cy + 18}}, rgb(0xe74c3c))
	case sim.IconClock:
		circle(dst, x, cy, 18, rgb(0xffffff))
		line(dst, x, cy, x, cy-13, 3, rgb(0x333333))
		line(dst, x, cy, x+9, cy+4, 3, rgb(0x333333))
	case sim.IconShirt:
		polygon(dst, []V{{X: x - 18, Y: cy - 12}, {X: x - 6, Y: cy - 16}, {X: x + 6, Y: cy - 16}, {X: x + 18, Y: cy - 12}, {X: x + 14, Y: cy - 2}, {X: x + 10, Y: cy - 4}, {X: x + 10, Y: cy + 16}, {X: x - 10, Y: cy + 16}, {X: x - 10, Y: cy - 4}, {X: x - 14, Y: cy - 2}}, rgb(0x9b6bd6))
	case sim.IconMoon:
		circle(dst, x, cy, 17, rgb(0x2b2b5a))
		circle(dst, x+8, cy-6, 13, rgb(disc))
	case sim.IconWorker, sim.IconCashier:
		lk := staffLook(staffPlace, sim.WorkerAgent, 0)
		if icon == sim.IconCashier {
			lk = staffLook(staffPlace, sim.CashierAgent, 0)
		}
		scaled(x, cy, 0.38, func() { drawPerson(dst, V{X: x, Y: cy + 50}, V{Y: 1}, 0, false, lk, nil) })
	default:
		spotIconOrStar(dst, icon, x, cy)
	}
}

func spotIconOrStar(dst *ebiten.Image, icon int, x, y float64) {
	switch icon {
	case sim.IconBag:
		roundRect(dst, x-14, y-14, 28, 32, 8, rgb(0xc0623a))
	case sim.IconSign:
		spotIcon(dst, "pancarte", x, y)
	case sim.IconBio:
		spotIcon(dst, "bio", x, y)
	case sim.IconFarm:
		building(dst, sim.Rect{X: x - 18, Y: y - 6, W: 36, H: 24}, 0xb8352c, 0x5a4a44)
	default:
		star(dst, x, y, 20, rgb(0xff9f1a))
	}
}

// superRefused : des clients repartent faute de place ou de caddie. Ce conseil
// passe avant les autres (sinon « Tu peux acheter… » le cachait toujours) ; en
// arrivant au supermarché, on voit aussi les refus survenus pendant qu'on était ailleurs.
func (g *game) superRefused(p *sim.Place) (string, V, bool) {
	pl := g.sim.Player
	if p.RefusedPark > g.seenPark {
		g.seenPark, g.refusedAt = p.RefusedPark, g.t
		g.refusedWhat = "parking"
	}
	if p.RefusedCart > g.seenCart {
		g.seenCart, g.refusedAt = p.RefusedCart, g.t
		g.refusedWhat = "caddies"
	}
	if g.t-g.refusedAt < 20 && g.refusedAt > 0 {
		for _, sp := range p.Spots {
			if sp.ID == g.refusedWhat && p.SpotVisible(sp) {
				if g.refusedWhat == "parking" {
					return "Parking plein : des clients repartent et ta réputation baisse ! Achète des places.", sp.Zone.C(), !sp.Zone.Has(pl.Pos)
				}
				return "Plus de caddies : des clients repartent et ta réputation baisse ! Achètes-en.", sp.Zone.C(), !sp.Zone.Has(pl.Pos)
			}
		}
		if g.refusedWhat == "parking" {
			return "Parking plein : des clients repartent et ta réputation baisse !", V{}, false
		}
		return "Plus de caddies : des clients repartent et ta réputation baisse !", V{}, false
	}
	return "", V{}, false
}

// superHint : conseils propres au supermarché.
func (g *game) superHint(p *sim.Place) (string, V, bool) {
	gs := g.sim
	pl := gs.Player
	if p.Station("caisse") != nil && gs.Levels[p.ID+".caissiere"] == 0 && gs.Money >= p.PlaceUpgrades()[0].Cost(0) { // (pas de caisse : pas de caissière)
		z := p.Spots[0].Zone
		return "Embauche une caissière : arrête-toi sur sa case.", z.C(), !z.Has(pl.Pos)
	}
	if p.ID == "super" {
		for _, z := range sim.SuperZones {
			if gs.Levels["super.ouvriers_"+z.ID] > 0 {
				return "", V{}, false
			}
		}
	} else if gs.Levels[p.ID+".ouvriers"] > 0 {
		return "", V{}, false
	}
	for _, z := range sim.TeamZones(p.ID) {
		if gs.Levels[p.ID+".ouvriers_"+z.ID] > 0 {
			return "", V{}, false
		}
	}
	if p.ID != "super" {
		return g.workshopHint(p)
	}
	// on remplit les rayons soi-même tant qu'il n'y a pas d'employé
	if len(pl.Carry) > 0 {
		for _, it := range pl.Carry {
			for _, s := range p.Stations {
				if s.Kind == sim.Shelf && s.Built && s.Item == it && s.Stock < s.StockMax {
					return "Mets les produits en rayon (zone au-dessus du rayon).", s.Zone.C(), true
				}
			}
		}
		return "", V{}, false
	}
	empty := true
	for _, s := range p.Stations {
		if s.Kind == sim.Shelf && s.Built && s.Stock < s.StockMax/2 {
			pal := p.Station("palette_" + itoa(int(s.Item)))
			if pal.Pile > 0 {
				return "Va chercher des produits à l'entrepôt, en haut.", pal.Zone.C(), true
			}
		}
	}
	for _, s := range p.Stations {
		if s.Kind == sim.Pallet && s.Pile > 0 {
			empty = false
		}
	}
	if empty && gs.Stats.Delivered < 30 {
		return "Les produits arrivent de la ferme en camion : remplis le quai de la ferme.", V{}, false
	}
	return "", V{}, false
}

// ---------------------------------------------------------------- achat d'un lieu

func (g *game) drawBuyPlace(dst *ebiten.Image) {
	p := g.buying
	if p == nil {
		g.panel = panelNone
		return
	}
	w, h := 940.0, 520.0
	x, y := g.panelFrame(dst, "Acheter "+strings.ToLower(p.Name[:1])+p.Name[1:]+" ?", w, h)
	txt := "L'immense supermarché vend les produits de tes lieux, espace par espace (ferme, légumes, fruits, puis crèmerie, boulangerie, glaces…). Ce qui part en camion arrive dans son entrepôt (1 objet livré = 3 en réserve). Les clients viennent en voiture et prennent un caddie : achète des places et des caddies !"
	switch p.ID {
	case "boulangerie":
		txt = "La boulangerie fait du pain et des viennoiseries : baguettes, pain blé-maïs, brioches, puis croissants et pains au chocolat (avec le beurre de la crèmerie). Les ingrédients arrivent en camion depuis le quai « → boulangerie » de la ferme."
	case "cremerie":
		txt = "La crèmerie transforme les trois laits de la ferme (vache, chèvre, brebis) en crème, beurre, fromages et yaourts. Son beurre part aussi en camion à la boulangerie, pour les croissants."
	case "ecole":
		txt = "À l'école, tu accueilles des élèves et tu leur fais cours (puis des profs le font pour toi). Chaque élève qui s'en va te donne un point (ouvrier, marketing, comptable ou formation) à placer depuis la mairie pour améliorer tes lieux."
	case "exotique":
		txt = "La ferme exotique fait pousser ananas, bananes, noix de coco, mangues, canne à sucre, cacao, café, riz, vanille, thé et cannelle. Elle a ses clients et ses commandes, et livre le sucre et le chocolat à la boulangerie et au glacier. Chaque espace a ses ouvriers."
	case "elevage":
		txt = "La ferme d'élevage ne reçoit pas de clients : elle élève volailles, bovins, ovins et porcs avec le maïs, le foin (grenier à foin de ta ferme) et les pommes de terre de ta ferme, puis les abat. Ses frigos partent en camion frigorifique à la boucherie. Chaque espace a ses éleveurs."
	case "boucherie":
		txt = "La boucherie-charcuterie découpe la viande de la ferme d'élevage : poulets, canards, bœuf, veau, mouton, agneau, porc. Ses clients sont rares mais achètent cher, et elle remplit l'espace boucherie du supermarché. Chaque espace a ses bouchers."
	case "glacier":
		txt = "Le glacier fait des sorbets (fraise, mangue, ananas) avec les fruits, puis des glaces : crème et œufs avec de la vanille, du café, de la noix de coco, du chocolat ou des poires. Les ingrédients arrivent en camion (ferme, crèmerie, ferme exotique)."
	}
	for i, l := range wrap(txt, 26, w-120, false) {
		say(dst, l, x+w/2, y+56+float64(i)*36, 26, false, 0, rgb(0x5a4a3a))
	}
	price := g.sim.PriceOf(p)
	coin(dst, x+w/2-textWidth(money(price), 44, true)/2-34, y+270, 22)
	say(dst, money(price), x+w/2+10, y+270, 44, true, 0, rgb(0x3b2a10))
	if price < p.Price { // cadeau du maire : l'ancien prix barré
		ow, nw := textWidth(money(p.Price), 28, true), textWidth(money(price), 44, true)
		ox := x + w/2 + 10 + nw/2 + 30 + ow/2
		say(dst, money(p.Price), ox, y+272, 28, true, 0, rgb(0x9a8a7a))
		line(dst, ox-ow/2-6, y+272, ox+ow/2+6, y+272, 3, rgb(0xd9534f))
		say(dst, map[bool]string{true: "Moitié prix", false: "Quart du prix"}[g.sim.Boost.Mul() == 2]+" : cadeau du maire !", x+w/2, y+320, 24, true, 0, rgb(0xd9534f))
	}
	ok := g.sim.Money >= price
	label := "Acheter"
	if !ok {
		label = "Pas assez d'argent"
	}
	g.btn(dst, label, x+w/2-180, y+h-180, 360, 76, 0x2fa86b, ok, func() {
		if g.sim.BuyPlace(p) {
			g.snd.play(sUnlock)
			g.closePanel()
			g.goPlace(p)
		}
	})
}

func (g *game) superTile() tile {
	t := g.placeTile("super", "Supermarché", 0x2fa86b)
	t.id = "supermarche"
	return t
}

// placeTile : tuile d'un lieu qui s'achète sur la carte.
func (g *game) placeTile(id, name string, col uint32) tile {
	sup := g.sim.Place(id)
	t := tile{id: id, name: name, col: col, open: true, place: id}
	if sup.Owned && id == "ecole" {
		t.status = "Ouvert"
		if n := g.totalPoints(); n > 0 {
			t.status = itoa(n) + " point(s) à répartir (mairie)"
		}
		t.fn = func() { g.goPlace(sup) }
		return t
	}
	if sup.Owned {
		t.status = "Ouvert"
		if !sup.HasCashier() && sup.QueueLen() > 0 {
			t.status = itoa(sup.QueueLen()) + " client(s) attendent !"
		}
		t.fn = func() { g.goPlace(sup) }
	} else {
		t.status = "À vendre : " + money(g.sim.PriceOf(sup))
		if g.sim.PriceOf(sup) < sup.Price {
			t.status += " (−" + itoa(int(math.Round(100-100*g.sim.PriceOf(sup)/sup.Price))) + " %)"
		}
		t.fn = func() { g.buying = sup; g.openPanel(panelBuy) }
	}
	return t
}

// workshopHint : boulangerie et glacier, tant qu'il n'y a pas d'employé.
func (g *game) workshopHint(p *sim.Place) (string, V, bool) {
	pl := g.sim.Player
	if len(pl.Carry) > 0 {
		for _, it := range pl.Carry {
			for _, s := range p.Stations {
				if !s.Built {
					continue
				}
				if s.Kind == sim.Shelf && s.Item == it && s.Stock < s.StockMax {
					return "Mets-le en vitrine (zone au-dessus de la vitrine).", s.Zone.C(), true
				}
				if s.Kind == sim.Machine {
					for _, in := range s.Inputs {
						if in == it && s.Buf[in] < s.BufMax {
							return "Dépose les ingrédients dans le dépôt (bleu) d'une machine.", s.Zone.C(), true
						}
					}
				}
			}
		}
		return "", V{}, false
	}
	for _, s := range p.Stations {
		if s.Kind == sim.Machine && s.Built && s.Pile > 0 {
			return "Des produits sont prêts : ramasse-les à la sortie de la machine.", s.OutZone.C(), true
		}
	}
	for _, s := range p.Stations {
		if s.Kind == sim.Machine && s.Built {
			for _, in := range s.Inputs {
				if s.Buf[in] < 3 {
					if pal := p.Station("palette_" + itoa(int(in))); pal != nil && pal.Pile > 0 {
						return "Va chercher des ingrédients à la réserve, en haut.", pal.Zone.C(), true
					}
				}
			}
		}
	}
	return "Les ingrédients arrivent de la ferme : remplis son quai « → " + strings.ToLower(p.Name[3:]) + " ».", V{}, false
}

// ---------------------------------------------------------------- points de l'école

func (g *game) totalPoints() int {
	n := 0
	for _, t := range sim.PointTypes {
		n += g.sim.Points(t)
	}
	return n
}

func (g *game) hasPoints() bool { return g.totalPoints() > 0 }

var pointNames = map[string]string{"ouvrier": "Ouvrier", "marketing": "Marketing", "comptable": "Comptable", "formation": "Formation"}

// tabs des points : un onglet par niveau de l'école.
var pointTabs = []struct{ name, sub string }{
	{"Primaire", "Ouvriers : agronomes, soigneurs, entraînement, mécaniciens, chefs"},
	{"Collège", "Clients : panier moyen, clients fidèles, réclame"},
	{"Lycée", "Argent : comptables, logistique"},
	{"Université", "École : formation et pédagogie"},
}

// drawPointsPanel : répartir les points de l'école dans les lieux (à la mairie).
func (g *game) drawPointsPanel(dst *ebiten.Image) {
	if g.pending == nil {
		g.pending = map[string]int{}
	}
	w, h := math.Min(1560, scr.VW-60), 1000.0
	x, y := g.panelFrame(dst, "Mairie de Clairbourg", w, h)
	g.veilleurBand(dst, x+30, y+16, w-60)
	y0 := y
	y += 104 // sous le veilleur
	// ce qui reste à répartir (moins ce qui est choisi)
	left := map[string]int{}
	for _, t := range sim.PointTypes {
		left[t] = g.sim.Points(t)
	}
	for key, n := range g.pending {
		left[pendingType(g.sim, key)] -= n
	}
	// onglets
	tw := (w - 60) / 4
	for i, t := range sim.PointTypes {
		i := i
		tx := x + 30 + float64(i)*tw
		sel := g.ptsTab == i
		a := 0.3
		if sel {
			a = 1
		}
		roundRect(dst, tx+4, y+8, tw-8, 74, 16, rgba(levelColUI(i), a))
		if g.hover(tx+4, y+8, tw-8, 74) && !sel {
			roundRect(dst, tx+4, y+8, tw-8, 74, 16, rgba(0xffffff, 0.2))
		}
		c := rgb(0x3b2a10)
		if sel {
			c = rgb(0xffffff)
		}
		say(dst, pointTabs[i].name, tx+tw/2, y+32, 24, true, 0, c)
		say(dst, pointNames[t]+" : "+itoa(left[t])+" point(s)", tx+tw/2, y+62, 19, false, 0, c)
		if left[t] > 0 && !sel {
			circle(dst, tx+tw-22, y+18, 11, rgb(0xd9534f))
		}
		g.addButton(tx+4, y+8, tw-8, 74, func() { g.ptsTab = i })
	}
	tab := g.ptsTab
	t := sim.PointTypes[tab]
	y += 92
	say(dst, pointTabs[tab].sub+". Une fois validés, les points ne peuvent plus être déplacés.", x+w/2, y+16, 19, false, 0, rgb(0x5a4a3a))
	y += 40
	// lieux qui acceptent ce type de point
	var places []*sim.Place
	for _, p := range g.sim.Places {
		if !p.Owned {
			continue
		}
		for _, k := range p.PointKinds() {
			if k.Type == t {
				places = append(places, p)
				break
			}
		}
	}
	var cur *sim.Place
	for _, p := range places {
		if p.ID == g.ptsPlace {
			cur = p
		}
	}
	if cur == nil && len(places) > 0 {
		cur = places[0]
	}
	by := y0 + h - 182 // boutons en bas de la fenêtre
	listH := by - 20 - y
	lw := 300.0
	roundRect(dst, x+30, y, lw, listH, 16, rgb(0xeadcc2))
	// ascenseur : trop de lieux pour la hauteur → on fait défiler (molette ou flèches)
	vis := max(1, int((listH-10)/70))
	if len(places) > vis {
		vis = max(1, int((listH-70)/70)) // (la place des flèches en bas)
		maxS := len(places) - vis
		if g.hover(x+30, y, lw, listH) {
			if _, wy := ebiten.Wheel(); wy > 0 {
				g.ptsScroll--
			} else if wy < 0 {
				g.ptsScroll++
			}
		}
		g.ptsScroll = max(0, min(maxS, g.ptsScroll))
		ay := y + listH - 56
		g.btn(dst, "Haut", x+40, ay, (lw-30)/2, 46, 0x8a6a3a, g.ptsScroll > 0, func() { g.ptsScroll-- })
		g.btn(dst, "Bas", x+50+(lw-30)/2, ay, (lw-30)/2, 46, 0x8a6a3a, g.ptsScroll < maxS, func() { g.ptsScroll++ })
		places = places[g.ptsScroll : g.ptsScroll+vis]
	} else {
		g.ptsScroll = 0
	}
	for i, p := range places {
		p := p
		py := y + 10 + float64(i)*70
		sel := p == cur
		col := rgb(0xf3e9d6)
		if sel {
			col = rgba(levelColUI(tab), 0.85)
		} else if g.hover(x+40, py, lw-20, 60) {
			col = rgb(0xfbf4e6)
		}
		roundRect(dst, x+40, py, lw-20, 60, 12, col)
		sayFit(dst, p.Name, x+40+(lw-20)/2, py+30, 22, true, 0, lw-40, rgb(0x3b2a10))
		// des points en attente ici ?
		n := 0
		for key, v := range g.pending {
			if strings.HasPrefix(key, p.ID+".") && pendingType(g.sim, key) == t {
				n += v
			}
		}
		if n > 0 {
			roundRect(dst, x+lw-12-44, py+8, 44, 26, 10, rgb(0x2fa86b))
			say(dst, "+"+itoa(n), x+lw-12-22, py+21, 16, true, 0, rgb(0xffffff))
		}
		g.addButton(x+40, py, lw-20, 60, func() { g.ptsPlace = p.ID })
	}
	// cartes des usages
	if cur != nil {
		var kinds []sim.PointKind
		for _, k := range cur.PointKinds() {
			if k.Type == t {
				kinds = append(kinds, k)
			}
		}
		rx := x + 30 + lw + 20
		rw := x + w - 30 - rx
		cols := 3
		if len(kinds) <= 4 {
			cols = 2
		}
		cw, ch := rw/float64(cols), 128.0
		if rows := (len(kinds) + cols - 1) / cols; rows > 0 {
			ch = math.Min(128, (by-20-y-float64(rows-1)*8)/float64(rows)) // (tout doit tenir au-dessus des boutons)
		}
		sc := ch / 128 // cartes plus petites quand il y en a beaucoup
		fs := math.Max(0.8, sc)
		for i, k := range kinds {
			k := k
			kx := rx + float64(i%cols)*cw
			ky := y + float64(i/cols)*(ch+8)
			key := cur.ID + "." + k.Key
			have := g.sim.Levels[key]
			pend := g.pending[key]
			roundRect(dst, kx+5, ky, cw-10, ch, 14, rgba(levelColUI(tab), 0.22))
			say(dst, strings.ToUpper(k.Group), kx+20, ky+20*sc, 15*fs, true, -1, rgb(0x8a6a3a))
			say(dst, itoa(have)+" pt", kx+cw-22, ky+20*sc, 15*fs, true, 1, rgb(0x8a6a3a))
			sayFit(dst, k.Name, kx+20, ky+46*sc, 22*fs, true, -1, cw-50, rgb(0x3b2a10))
			if pend > 0 {
				sayFit(dst, k.Val(have)+"  →  "+k.Val(have+pend), kx+20, ky+76*sc, 19*fs, true, -1, cw-200, rgb(0x2fa86b))
			} else {
				sayFit(dst, k.Val(have), kx+20, ky+76*sc, 19*fs, true, -1, cw-200, rgb(0x5a3a1a))
			}
			sayFit(dst, k.Desc, kx+20, ky+ch-20*sc, 14*fs, false, -1, cw-50, rgb(0x7a6a5a))
			for j, n := range []int{1, 10} {
				// « +10 » place jusqu'à 10 points (ce qui reste s'il y en a moins)
				add := min(n, left[t])
				g.btn(dst, "+"+itoa(n), kx+cw-15-float64(2-j)*84, ky+62*sc, 76, 34*sc, 0x2fa86b, add > 0, func() {
					g.pending[key] += add
				})
			}
		}
	}
	any := len(g.pending) > 0
	g.btn(dst, "Valider", x+w/2-260, by, 240, 70, 0x2fa86b, any, func() {
		for key, n := range g.pending {
			id, k, _ := strings.Cut(key, ".")
			if p := g.sim.Place(id); p != nil {
				for _, kk := range p.PointKinds() {
					if kk.Key == k {
						g.sim.SpendPoints(p, kk, n)
					}
				}
			}
		}
		g.pending = nil
		g.snd.play(sBuy)
		g.toast("Points placés !", "")
	})
	g.btn(dst, "Annuler", x+w/2+20, by, 240, 70, 0x8a7a6a, any, func() { g.pending = nil })
}

// pendingType : type de point d'une clé « lieu.pts_xxx ».
func pendingType(gs *sim.Game, key string) string {
	id, k, _ := strings.Cut(key, ".")
	if p := gs.Place(id); p != nil {
		for _, kk := range p.PointKinds() {
			if kk.Key == k {
				return kk.Type
			}
		}
	}
	return "formation"
}

func levelColUI(i int) uint32 { return levelCol[i] }

// schoolHint : conseils à l'école.
func (g *game) schoolHint(p *sim.Place) (string, V, bool) {
	gs := g.sim
	pl := gs.Player
	reg := p.Station("accueil")
	waiting, seated := 0, 0
	for _, s := range p.Students {
		switch s.State {
		case 0:
			waiting++
		case 1, 2:
			seated++
		}
	}
	// du matériel demandé
	for _, r := range p.Rooms {
		if !r.Built() || r.Want < 0 {
			continue
		}
		if pl.Count(r.Want) > 0 {
			return "« " + r.Name + " » attend ce matériel : apporte-le dans la classe.", r.Body.C(), true
		}
		if sh := p.Station("fourn_" + itoa(int(r.Want))); sh != nil && sh.Pile > 0 && len(p.Workers) == 0 {
			return "« " + r.Name + " » demande : " + sim.Items[r.Want].Name + ". Va le chercher au local des fournitures.", sh.Zone.C(), true
		}
	}
	// une salle avec des élèves mais sans prof ni toi
	for _, r := range p.Rooms {
		if !r.Built() || p.HasProf(r) {
			continue
		}
		busy := false
		for n := 0; n < p.Tables(r); n++ {
			if s := r.SeatStudent(n); s != nil && s.State == 2 {
				busy = true
			}
		}
		if busy && !p.Teaching(r) && (waiting == 0 || reg.Cashier != nil || seated >= 1) {
			return "Des élèves t'attendent en « " + r.Name + " » : reste dans la classe pour faire cours.", r.Body.C(), true
		}
		if busy && p.Teaching(r) {
			return "Tu fais cours ! Reste dans la classe jusqu'à la fin de la formation (ou embauche un prof).", V{}, false
		}
	}
	if waiting > 0 && reg.Cashier == nil {
		if reg.Zone.Has(pl.Pos) {
			return "Reste à l'accueil : tu inscris les élèves qui attendent.", V{}, false
		}
		return "Des élèves attendent : va à l'accueil pour les inscrire.", reg.Zone.C(), true
	}
	if g.hasPoints() {
		return "Tu as des points à répartir : va à la mairie (Tab, puis Mairie).", V{}, false
	}
	return "", V{}, false
}

// roundShortcut : bouton rond (raccourci), qui brille quand un achat est possible.
func (g *game) roundShortcut(dst *ebiten.Image, x, y float64, hot bool, icon func(cx, cy float64), fn func()) {
	d := 70.0
	cx, cy := x+d/2+3, y+d/2
	if hot {
		k := 0.5 + 0.5*math.Sin(g.t*4)
		circle(dst, cx, cy, d/2+8+3*k, rgba(0xffd23f, 0.35+0.3*k))
	}
	circle(dst, cx, cy+4, d/2, rgba(0x000000, 0.25))
	circle(dst, cx, cy, d/2, rgb(0xfff8ea))
	if g.hover(cx-d/2, cy-d/2, d, d) {
		circle(dst, cx, cy, d/2, rgba(0xffffff, 0.4))
	}
	icon(cx, cy)
	if hot {
		badge(dst, cx+d/2-6, cy-d/2+6, g.t)
	}
	g.addButton(cx-d/2, cy-d/2, d, d, fn)
}
