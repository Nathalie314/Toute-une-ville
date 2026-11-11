package sim

import (
	"math"
	"math/rand"
	"time"
)

// Save est ce qui est écrit dans le fichier de sauvegarde.
type Save struct {
	Version  int
	Money    float64
	Levels   map[string]int
	Outfit   string
	Owned    map[string]bool
	Stats    Stats
	Rate     float64 // revenu par seconde au moment de la sauvegarde
	Unix     int64   // heure de la sauvegarde
	Places   map[string]*PlaceSave
	Badges   map[string]bool
	Commande *Order            `json:",omitempty"` // commande spéciale en cours (le temps s'arrête jeu fermé)
	Cadeau   *Boost            `json:",omitempty"` // cadeau d'une commande en cours (en pause jeu fermé)
	Favoris  map[string]string `json:",omitempty"` // touches 1 à 9 et 0 → lieu
}

type PlaceSave struct {
	Paid       map[string]float64 // déblocages (payé en partie ou en totalité)
	Stock      map[string]int
	Pile       map[string]int
	Buf        map[string][]int
	Pousses    map[string]int // nombre de pousses par champ
	PoussePaye map[string]float64
	NiveauPaye map[string]float64
	StockPaye  map[string]float64
	CasePaye   map[string]float64
	Reputation float64
	Achete     bool
	Servis     int
	Eleves     []EleveSave `json:",omitempty"`
}

func (g *Game) Snapshot(unix int64) *Save {
	s := &Save{Version: SaveVersion, Money: g.Money, Levels: g.Levels, Outfit: g.Outfit, Owned: g.Owned, Stats: g.Stats,
		Rate: g.RateLong(), Unix: unix, Places: map[string]*PlaceSave{}, Badges: g.Badges, Commande: g.Order, Cadeau: g.Boost, Favoris: g.Fav}
	for _, p := range g.Places {
		ps := &PlaceSave{Paid: map[string]float64{}, Stock: map[string]int{}, Pile: map[string]int{}, Buf: map[string][]int{}, Pousses: map[string]int{}, PoussePaye: map[string]float64{}, NiveauPaye: map[string]float64{}, StockPaye: map[string]float64{}, CasePaye: map[string]float64{}}
		for _, u := range p.Unlocks {
			if u.Paid > 0 {
				ps.Paid[u.ID] = u.Paid
			}
		}
		ps.Reputation = p.Rep
		ps.Achete = p.Owned
		ps.Servis = p.Served
		if p.ID == "ecole" {
			ps.Eleves = p.saveStudents()
		}
		for _, sp := range p.Spots {
			if sp.Paid > 0 {
				ps.CasePaye[sp.ID] = sp.Paid
			}
		}
		for _, st := range p.Stations {
			if st.Stock > 0 {
				ps.Stock[st.ID] = st.Stock
			}
			if st.Pile > 0 {
				ps.Pile[st.ID] = st.Pile
			}
			if st.LvlPaid > 0 {
				ps.NiveauPaye[st.ID] = st.LvlPaid
			}
			if st.StkPaid > 0 {
				ps.StockPaye[st.ID] = st.StkPaid
			}
			if st.Kind == Field {
				ps.Pousses[st.ID] = len(st.Plants)
				if st.PlantPaid > 0 {
					ps.PoussePaye[st.ID] = st.PlantPaid
				}
			}
			if st.Kind == Machine || st.Kind == Dock {
				b := make([]int, NItems)
				copy(b, st.Buf[:])
				ps.Buf[st.ID] = b
			}
		}
		s.Places[p.ID] = ps
	}
	return s
}

// Restore recharge une sauvegarde. Renvoie l'argent gagné pendant l'absence
// et la durée prise en compte (en secondes).
// SaveVersion : version des sauvegardes. La 10 est celle de la version 4.0 du
// jeu (nouvelle ferme, crèmerie…) : les parties d'avant ne se convertissent pas.
const SaveVersion = 14 // 14 : la ferme exotique

func (g *Game) Restore(s *Save, nowUnix int64) (offline, secs float64) {
	// le hasard repart d'une graine différente à chaque lancement (sinon la
	// première commande après le chargement était toujours la même)
	g.rng = rand.New(rand.NewSource(nowUnix ^ time.Now().UnixNano()))
	for i, p := range g.Places {
		p.rng = rand.New(rand.NewSource(nowUnix + int64(i)*7919 + time.Now().UnixNano()))
	}
	g.Money = s.Money
	if s.Levels != nil {
		g.Levels = s.Levels
	}
	// la serre et le verger n'ont plus qu'une équipe chacun : les ouvriers d'une
	// colonne (anciennes sauvegardes) rejoignent l'équipe de la serre ou du verger
	for i, id := range ColumnIDs {
		k := "ferme." + teamKey(id)
		if n := g.Levels[k]; n > 0 {
			to := "ferme." + teamKey("serre")
			if i >= 6 {
				to = "ferme." + teamKey("verger")
			}
			g.Levels[to] += n
		}
		delete(g.Levels, k)
	}
	if s.Version < 4 {
		// version 4 : les ouvriers de la ferme sont répartis par zone, dans l'ordre d'embauche
		old := []float64{250, 1000, 3500, 12000, 35000, 100000, 220000, 450000, 900000, 1800000}
		n := g.Levels["ferme.ouvriers"]
		for i := 0; i < n; i++ {
			if i < len(HireOrder) {
				g.Levels["ferme."+teamKey(HireOrder[i])]++
			} else if i < len(old) {
				g.Money += old[i] // ouvrier en trop : remboursé
			}
		}
	}
	if s.Version < 5 {
		// version 5 : la pancarte et le label bio (et leurs équivalents) deviennent
		// des points de marketing et de comptabilité dans le même lieu
		for _, id := range []string{"ferme", "super", "boulangerie", "glacier"} {
			g.Levels[id+".pts_marketing"] += 20 * g.Levels[id+".pancarte"]
			g.Levels[id+".pts_comptable"] += 12 * g.Levels[id+".bio"]
			delete(g.Levels, id+".pancarte")
			delete(g.Levels, id+".bio")
		}
	}
	if l := g.Levels["mairie.pub"]; l > 0 {
		// version 6 : les panneaux publicitaires (+15 % de clients partout par niveau)
		// deviennent des points de marketing dans chaque lieu
		for _, id := range []string{"ferme", "super", "boulangerie", "glacier"} {
			g.Levels[id+".pts_marketing"] += 15 * l
		}
		delete(g.Levels, "mairie.pub")
	}
	if s.Version < 9 {
		// version 9 : plus d'amélioration du camion à la mairie (les mécaniciens
		// s'en occupent) : les niveaux achetés sont remboursés
		for l := 0; l < g.Levels["mairie.camion"]; l++ {
			g.Money += 3000 * math.Pow(2.6, float64(l))
		}
		delete(g.Levels, "mairie.camion")
	}
	if s.Version < 8 {
		// version 8 : plus de niveaux pour les étagères de l'école : remboursés
		if sc := g.Place("ecole"); sc != nil {
			for _, st := range sc.Stations {
				if st.Kind != Animal {
					continue
				}
				u := sc.StationUpgrade(st)
				for l := 0; l < g.Levels[u.Key]; l++ {
					g.Money += u.Cost(l)
				}
				delete(g.Levels, u.Key)
			}
		}
	}
	g.syncFarmTotal()
	if s.Owned != nil {
		g.Owned = s.Owned
	}
	if s.Version < 7 {
		// version 7 : le sac à dos passe de la mairie au personnage ; les tenues
		// deviennent décoratives et moins chères (la différence est remboursée)
		if l := g.Levels["mairie.sac"]; l > 0 {
			g.Levels["perso.sac"] = l
			delete(g.Levels, "mairie.sac")
		}
		old := map[string]float64{"survet": 1500, "salopette": 4000, "tablier": 9000, "rollers": 18000, "chic": 40000, "heros": 250000}
		for id, c := range old {
			if g.Owned[id] {
				g.Money += c - OutfitByID(id).Cost
			}
		}
	}
	g.Owned["fermier"] = true
	g.Outfit = s.Outfit
	if !g.Owned[g.Outfit] {
		g.Outfit = "fermier"
	}
	g.Stats = s.Stats
	g.Order = s.Commande
	if s.Favoris != nil {
		g.Fav = s.Favoris
	}
	if s.Badges != nil {
		g.Badges = s.Badges
	}
	for _, p := range g.Places {
		ps := s.Places[p.ID]
		if ps == nil {
			continue
		}
		for _, sp := range p.Spots {
			sp.Paid = ps.CasePaye[sp.ID]
		}
		if ps.Reputation >= 1 {
			p.Rep = ps.Reputation
		}
		p.Owned = ps.Achete || p.ID == "ferme"
		p.Served = ps.Servis
		if p.ID == "ferme" && p.Served == 0 {
			p.Served = s.Stats.Served // ancienne sauvegarde
		}
		if p.ID == "ferme" && s.Version < 3 {
			// version 3 : la ferme démarre au maïs ; le champ de blé était déjà là,
			// le four à pain devient le moulin à maïs, la pâtisserie part à la boulangerie
			ps.Paid["ble"] = 400
			if ps.Paid["four"] >= 300 {
				ps.Paid["moulin_mais"] = 60
			}
			if ps.Paid["patisserie"] >= 50000 {
				g.Money += 50000 // remboursée
			}
		}
		for _, u := range p.Unlocks {
			u.Paid = ps.Paid[u.ID]
			if u.Paid >= u.Cost-1e-6 && u.Cost > 0 {
				u.Done = false
				p.complete(u)
			}
		}
		p.applyCaps()
		for _, st := range p.Stations {
			n, ok := ps.Pousses[st.ID]
			if !ok && s.Version < 2 {
				n, ok = len(st.Slots), true // ancienne sauvegarde : champs complets
			}
			if ok && st.Kind == Field {
				n = max(1, min(n, len(st.Slots)))
				st.Plants = st.Plants[:0]
				for i := 0; i < n; i++ {
					st.Plants = append(st.Plants, Plant{P: st.Slots[i], G: 1})
				}
				st.PlantPaid = ps.PoussePaye[st.ID]
			}
			st.LvlPaid = ps.NiveauPaye[st.ID]
			st.StkPaid = ps.StockPaye[st.ID]
			st.Stock = min(ps.Stock[st.ID], st.StockMax)
			st.Pile = min(ps.Pile[st.ID], st.PileMax)
			if b, ok := ps.Buf[st.ID]; ok {
				for i := 0; i < int(NItems) && i < len(b); i++ {
					st.Buf[i] = b[i]
				}
			}
		}
		p.syncStaff()
		if p.ID == "ecole" {
			p.restoreStudents(ps.Eleves)
		}
	}
	g.checkBadges(false) // badges mérités avant leur invention : sans fanfare
	g.Events = nil
	// gains pendant l'absence
	if s.Unix > 0 && nowUnix > s.Unix {
		secs = float64(nowUnix - s.Unix)
		maxS := NightHours[min(g.Levels["mairie.veilleur"], len(NightHours)-1)] * 3600
		secs = math.Min(secs, maxS)
		// sans personne pour encaisser, rien ne rentre
		// 4 % du revenu, +1 % par ouvrier (jusqu'à 8 %) : l'absence aide, sans remplacer le jeu
		f := 0.0
		for _, p := range g.Places {
			if p.Owned && p.HasCashier() && p.ID != "ecole" {
				f = math.Max(f, 0.04+0.01*math.Min(4, float64(g.Levels[p.key("ouvriers")])))
			}
		}
		if sc := g.Place("ecole"); sc != nil {
			_ = sc // (l'école est fermée quand le jeu l'est : hors ligne, seulement de l'argent)
		}
		// clients fidèles (points du collège) : moyenne de tes lieux qui vendent
		fid, nf := 0.0, 0
		for _, p := range g.Places {
			if p.Owned && p.sells() {
				fid += p.ptsF("pts_fidele")
				nf++
			}
		}
		if nf > 0 {
			f *= 1 + fid/float64(nf)
		}
		offline = math.Round(s.Rate * secs * f)
		if offline > 0 {
			g.Money += offline
			g.Stats.Earned += offline
		}
		// à ton retour, les champs ont poussé
		for _, p := range g.Places {
			for _, st := range p.Stations {
				for i := range st.Plants {
					st.Plants[i].G = 1
				}
			}
		}
	}
	// on garde la mémoire du revenu pour la prochaine absence
	if s.Rate > 0 {
		g.income = append(g.income, incomeRec{0, s.Rate * 60})
	}
	g.Boost = s.Cadeau // (après les gains d'absence : un cadeau ne travaille pas jeu fermé)
	return
}
