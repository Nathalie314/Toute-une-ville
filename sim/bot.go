package sim

import (
	"fmt"
	"math"
	"strings"
)

// Bot joue tout seul (sert à équilibrer le rythme du jeu).
type Bot struct {
	Log      []string
	goal     *Unlock
	plant    *Station
	plantN   int
	atReg    bool
	nextBuyT float64
	// Human : joue comme une vraie joueuse (achète tout un champ « pour le plaisir »,
	// prend un veilleur de nuit pour gagner pendant ses absences)
	Human bool
}

func (b *Bot) Step(g *Game, dt float64) {
	p := g.Cur
	pl := g.Player
	// le joueur automatique ne livre pas les commandes : il les laisse tomber
	// sans pénalité (un vrai joueur les livre), pour que les mesures restent justes
	if g.Order != nil && g.Order.Left < 1 {
		g.Order = nil
		g.orderT = 180
	}
	// achats d'améliorations
	if g.T >= b.nextBuyT {
		b.nextBuyT = g.T + 2
		b.shop(g)
		b.spendPoints(g)
	}
	// caisse sans caissière ?
	q := 0
	for _, c := range p.Customers {
		if c.State == cQueue {
			q++
		}
	}
	reg := p.Register()
	if !p.HasCashier() && (q >= 2 || (b.atReg && q > 0)) {
		b.atReg = true
		p.endJob(pl)
		if reg.Zone.Has(pl.Pos) && pl.Pos.Dist(reg.Zone.C()) < 30 {
			pl.Moving = false
			return
		}
		p.walk(pl, reg.Zone.C(), p.agentSpeed(pl), dt)
		return
	}
	b.atReg = false
	// achat abordable (déblocage ou pousse) : le moins cher d'abord
	if b.goal == nil && b.plant == nil {
		best := 1e18
		for _, u := range p.Unlocks {
			if np := b.nextPlace(g); np != nil && u.Cost >= 5500 && np.Price < 600*g.Rate()+g.Money && (!b.Human || b.ready(g)) {
				continue // (il garde de quoi acheter le prochain lieu, s'il est à portée : ~10 min de recettes)
			}
			if !u.Done && p.Visible(u) && g.Money >= u.Cost-u.Paid && u.Cost < best {
				best, b.goal = u.Cost, u
			}
		}
		unlockLeft := false
		for _, u := range p.Unlocks {
			if !u.Done && p.Visible(u) {
				unlockLeft = true
			}
		}
		for _, s := range p.Stations {
			if s.Kind == Field && len(s.Plants) >= 30 && unlockLeft && !b.Human {
				continue // grand champ : les bâtiments d'abord (comme un joueur raisonnable)
			}
			if s.Kind == Field && s.Built && len(s.Plants) < len(s.Slots) && g.Money >= s.PlantCost() && s.PlantCost() < best {
				best, b.goal, b.plant = s.PlantCost(), nil, s
			}
		}
	}
	if b.plant != nil {
		s := b.plant
		if len(s.Plants) != b.plantN && b.plantN > 0 || len(s.Plants) >= len(s.Slots) || g.Money < 1 {
			if len(s.Plants) != b.plantN && b.plantN > 0 {
				b.Log = append(b.Log, fmt.Sprintf("%7s  pousse %-20s n°%d", hms(g.T), s.Name, len(s.Plants)))
			}
			b.plant, b.plantN = nil, 0
			return
		}
		b.plantN = len(s.Plants)
		p.endJob(pl)
		p.walk(pl, s.PlantZone().C(), p.agentSpeed(pl), dt)
		return
	}
	if b.goal != nil {
		if b.goal.Done {
			b.Log = append(b.Log, fmt.Sprintf("%7s  débloque %-18s (%.0f €)", hms(g.T), b.goal.Name, b.goal.Cost))
			b.goal = nil
			return
		}
		if g.Money < 1 {
			b.goal = nil
		} else {
			p.endJob(pl)
			p.walk(pl, b.goal.Zone.C(), p.agentSpeed(pl), dt)
			return
		}
	}
	p.workerAI(pl, dt)
}

func (b *Bot) shop(g *Game) {
	for _, id := range []string{"ecole", "super", "cremerie", "exotique", "boulangerie", "glacier", "port", "elevage", "boucherie", "domaine", "patisserie"} {
		if q := g.Place(id); b.Human && !q.Owned && !b.ready(g) {
			break // (elle finit d'abord ce qu'elle a avant d'acheter un nouveau lieu)
		}
		if !b.shopRemote(g, g.Place(id)) {
			break // on n'achète pas le lieu suivant avant celui-ci
		}
	}
	p := g.Cur
	// réserve pour le prochain déblocage visible
	var next *Unlock
	for _, u := range p.Unlocks {
		if !u.Done && p.Visible(u) {
			next = u
			break
		}
	}
	var ups []*Upgrade
	ups = append(ups, p.PlaceUpgrades()...)
	ups = append(ups, MairieUpgrades()...)
	for _, gr := range Gears {
		if gr.Slot != "main" && gr.Desc(1) != diplomeDesc(1) { // (l'ensemble du diplômé ne sert qu'en classe)
			ups = append(ups, gr.Upgrade())
		}
	}
	for _, s := range p.Stations {
		if s.Built && s.Upgradable() {
			ups = append(ups, p.StationUpgrade(s))
		}
		if s.Built && s.Stockable() {
			ups = append(ups, p.StockUpgrade(s))
		}
	}
	for {
		var best *Upgrade
		bc := 1e18
		for _, u := range ups {
			l := g.Levels[u.Key]
			if l >= u.Max || (u.Key == "mairie.veilleur" && !b.Human) || !g.hireOK(u.Key) || !g.droneOK(u.Key) || !g.schoolOK(u.Key) {
				continue
			}
			c := u.Cost(l)
			w := c
			if u.Key == p.key("caissiere") || (b.Human && (u.Key == "mairie.veilleur" || strings.Contains(u.Key, ".ouvriers"))) || (u.Key == p.key("ouvriers") && l < 2) || (p.ID == "ferme" && strings.HasPrefix(u.Key, "ferme.ouvriers_") && g.Levels["ferme.ouvriers"] < 4) || botUrgent(u.Key) {
				w *= 0.1
			}
			if w < bc {
				best, bc = u, w
			}
		}
		if best == nil || (!g.sup().Owned && bc >= 3000) {
			return // (avant le supermarché, il économise au-delà de 3 000 €)
		}
		c := best.Cost(g.Levels[best.Key])
		reserve := 0.0
		if next != nil {
			reserve = (next.Cost - next.Paid) * 0.5
		}
		if best.Key == p.key("caissiere") || (b.Human && best.Key == "mairie.veilleur") || (best.Key == p.key("ouvriers") && g.Levels[best.Key] < 2) || (strings.HasPrefix(best.Key, "ferme.ouvriers_") && g.Levels["ferme.ouvriers"] < 4) || botUrgent(best.Key) {
			reserve = 0
		}
		if np := b.nextPlace(g); np != nil && np.ID != "super" && np.Price < 600*g.Rate()+g.Money && (!b.Human || b.ready(g)) {
			reserve = math.Max(reserve, np.Price*1.1) // il économise pour le prochain lieu
		}
		if g.Money-c < reserve {
			return
		}
		if !g.Buy(best) {
			return
		}
		b.Log = append(b.Log, fmt.Sprintf("%7s    achète %-28s niv %d (%.0f €)", hms(g.T), best.Name, g.Levels[best.Key], c))
	}
}

// nextPlace : le prochain lieu à acheter (dans l'ordre).
func (b *Bot) nextPlace(g *Game) *Place {
	for _, id := range []string{"super", "cremerie", "exotique", "boulangerie", "glacier", "elevage", "boucherie", "domaine", "patisserie"} {
		if p := g.Place(id); p != nil && !p.Owned {
			return p
		}
	}
	return nil
}

func hms(t float64) string {
	s := int(t)
	return fmt.Sprintf("%d:%02d:%02d", s/3600, s/60%60, s%60)
}

// shopRemote : le bot gère un lieu « à distance » (achats seulement).
// Renvoie vrai si le lieu est acheté.
func (b *Bot) shopRemote(g *Game, s *Place) bool {
	if !s.Owned {
		if g.Money >= g.PriceOf(s)*1.1 {
			g.BuyPlace(s)
			b.Log = append(b.Log, fmt.Sprintf("%7s  ACHÈTE %s", hms(g.T), s.Name))
		}
		return s.Owned
	}
	for {
		var bestU *Unlock
		var bestUp *Upgrade
		var bestF *Station
		bc := 1e18
		for _, u := range s.Unlocks {
			if !u.Done && s.Visible(u) && u.Cost < bc {
				bc, bestU, bestUp = u.Cost, u, nil
			}
		}
		ups := append([]*Upgrade{}, s.PlaceUpgrades()...)
		if b.Human {
			// comme au lieu courant : pousses des champs, niveaux et stocks des bâtiments
			for _, st := range s.Stations {
				if st.Built && st.Upgradable() {
					ups = append(ups, s.StationUpgrade(st))
				}
				if st.Built && st.Stockable() {
					ups = append(ups, s.StockUpgrade(st))
				}
				if st.Kind == Field && st.Built && len(st.Plants) < len(st.Slots) && st.PlantCost() < bc {
					bc, bestU, bestUp, bestF = st.PlantCost(), nil, nil, st
				}
			}
		}
		for _, u := range ups {
			l := g.Levels[u.Key]
			if l >= u.Max || !g.schoolOK(u.Key) || !g.droneOK(u.Key) || !g.hireOK(u.Key) {
				continue
			}
			if strings.HasPrefix(u.Key, "super.caissiere") && len(u.Key) > len("super.caissiere") {
				if r := s.Station("caisse" + u.Key[len("super.caissiere"):]); r != nil && !r.Built {
					continue
				}
			}
			c := u.Cost(l)
			if u.Key == s.key("caissiere") || (b.Human && strings.Contains(u.Key, ".ouvriers")) {
				c *= 0.1 // (elle embauche volontiers : moins de travail pour elle)
			}
			if c < bc {
				bc, bestU, bestUp, bestF = c, nil, u, nil
			}
		}
		if !g.sup().Owned && bc >= 3000 {
			return true // il économise pour le supermarché
		}
		if np := b.nextPlace(g); np != nil && np.ID != "super" && np.Price < 600*g.Rate()+g.Money && g.Money-bc < np.Price*1.1 && (!b.Human || b.ready(g)) {
			return true // il économise pour le prochain lieu
		}
		switch {
		case bestF != nil && g.Money >= bestF.PlantCost()*1.5:
			g.Money -= bestF.PlantCost()
			pl := Plant{P: bestF.Slots[len(bestF.Plants)]}
			if bestF.Column && !bestF.Serre {
				pl.Young = 1
			}
			bestF.Plants = append(bestF.Plants, pl)
		case bestU != nil && g.Money >= bestU.Cost*1.5:
			g.Money -= bestU.Cost
			s.complete(bestU)
			b.Log = append(b.Log, fmt.Sprintf("%7s  %s : %s", hms(g.T), s.ID, bestU.Name))
		case bestUp != nil && g.Money >= bestUp.Cost(g.Levels[bestUp.Key])*1.5:
			if !g.Buy(bestUp) {
				return true
			}
			b.Log = append(b.Log, fmt.Sprintf("%7s  %s : %s niv %d", hms(g.T), s.ID, bestUp.Name, g.Levels[bestUp.Key]))
		default:
			return true
		}
	}
}

// spendPoints : le joueur automatique place ses points de l'école, un par un,
// toujours là où il y en a le moins parmi ses usages préférés.
func (b *Bot) spendPoints(g *Game) {
	if b.Human {
		b.spendPointsHuman(g)
		return
	}
	prefs := map[string][]string{
		"ouvrier":   {"ferme.pts_agro", "ferme.pts_soin", "ferme.pts_mac_vit", "ferme.pts_ent_vit", "ferme.pts_dr_cap", "ferme.pts_dr_vit", "ferme.pts_cam_vit", "ferme.pts_cam_cap", "boulangerie.pts_chef_vit", "cremerie.pts_chef_vit", "glacier.pts_chef_vit", "super.pts_caisse", "super.pts_caddie", "ferme.pts_caisse", "boulangerie.pts_caisse"},
		"marketing": {"ferme.pts_panier", "super.pts_panier", "ferme.pts_fidele", "boulangerie.pts_panier", "cremerie.pts_panier", "ferme.pts_reclame"},
		"comptable": {"ferme.pts_comptable", "super.pts_comptable", "boulangerie.pts_comptable", "cremerie.pts_comptable"},
		"formation": {"ecole.pts_formation", "ecole.pts_pedagogie"},
	}
	for t, keys := range prefs {
		for g.Points(t) > 0 {
			var bp *Place
			var bk PointKind
			bl := 1 << 30
			for _, full := range keys {
				id, key, _ := strings.Cut(full, ".")
				p := g.Place(id)
				if p == nil || !p.Owned {
					continue
				}
				if key == "pts_cam_vit" || key == "pts_cam_cap" {
					if sup := g.Place("super"); sup == nil || !sup.Owned {
						continue
					}
				}
				for _, k := range p.PointKinds() {
					if k.Key == key && g.Levels[full] < bl {
						bp, bk, bl = p, k, g.Levels[full]
					}
				}
			}
			if bp == nil || !g.SpendPoints(bp, bk, 1) {
				break
			}
		}
	}
}

// botUrgent : les drones de la ferme passent avant le reste (sinon rien n'arrive au supermarché).
func botUrgent(key string) bool {
	return strings.HasPrefix(key, "ferme.drone_") && key != "ferme.drone_cap" && key != "ferme.drone_vit"
}

// spendPointsHuman : elle répartit ses points sur tout ce qui existe dans ses
// lieux (pas une liste figée), toujours là où il y en a le moins.
func (b *Bot) spendPointsHuman(g *Game) {
	for _, t := range PointTypes {
		for g.Points(t) > 0 {
			var bp *Place
			var bk PointKind
			bl := 1 << 30
			for _, p := range g.Places {
				if !p.Owned {
					continue
				}
				for _, k := range p.PointKinds() {
					if k.Type == t && g.Levels[p.key(k.Key)] < bl {
						bp, bk, bl = p, k, g.Levels[p.key(k.Key)]
					}
				}
			}
			if bp == nil || !g.SpendPoints(bp, bk, 1) {
				break
			}
		}
	}
}

// ready : la joueuse a « fini » ses lieux et peut économiser pour le suivant :
// plus rien de bon marché à côté du prix du prochain lieu (bâtiments, embauches,
// pousses : elle aime « avoir tout acheté »).
func (b *Bot) ready(g *Game) bool {
	np := b.nextPlace(g)
	if np == nil {
		return true
	}
	lim := 0.3 * np.Price
	for _, p := range g.Places {
		if !p.Owned || p.ID == "ecole" {
			continue
		}
		for _, u := range p.Unlocks {
			if !u.Done && p.Visible(u) && u.Cost < lim {
				return false
			}
		}
		for _, u := range p.PlaceUpgrades() {
			if l := g.Levels[u.Key]; strings.Contains(u.Key, ".ouvriers") && l < u.Max && g.hireOK(u.Key) && u.Cost(l) < lim {
				return false
			}
		}
		for _, st := range p.Stations {
			if st.Kind == Field && st.Built && len(st.Plants) < len(st.Slots) && st.PlantCost() < lim/3 {
				return false
			}
		}
	}
	return true
}
