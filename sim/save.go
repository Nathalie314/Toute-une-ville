package sim

import (
	"math"
	"math/rand"
	"strings"
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
	Vol      *TheftSave        `json:",omitempty"` // vol en cours (en pause jeu fermé : quitter ne chasse pas les voleurs)
	Garde    []GardeSave       `json:",omitempty"` // les voleurs en garde à vue (commissariat)
	Prison   *PrisonSave       `json:",omitempty"` // la prison : ses détenus, ses fourgons en route
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
	Eleves     []EleveSave        `json:",omitempty"`
	Stands     int                `json:",omitempty"` // parc : ventes des stands
	Pourboires map[string]float64 `json:",omitempty"` // crêperie : pourboires laissés sur les tables
	Attables   []CrepGroupSave    `json:",omitempty"` // crêperie : les clients attablés (en pause)
	Visiteurs  []ParcVisitorSave  `json:",omitempty"` // parc : les visiteurs entrés (en pause)
}

func (g *Game) Snapshot(unix int64) *Save {
	s := &Save{Version: SaveVersion, Money: g.Money, Levels: g.Levels, Outfit: g.Outfit, Owned: g.Owned, Stats: g.Stats,
		Rate: g.RateOffline(), Unix: unix, Places: map[string]*PlaceSave{}, Badges: g.Badges, Commande: g.Order, Cadeau: g.Boost, Favoris: g.Fav, Vol: g.theftSave(), Garde: g.custodySave(), Prison: g.prisonSave()}
	for _, p := range g.Places {
		ps := &PlaceSave{Paid: map[string]float64{}, Stock: map[string]int{}, Pile: map[string]int{}, Buf: map[string][]int{}, Pousses: map[string]int{}, PoussePaye: map[string]float64{}, NiveauPaye: map[string]float64{}, StockPaye: map[string]float64{}, CasePaye: map[string]float64{}}
		// les montants payés sont écrits sans le coefficient du lieu (prices.go) :
		// si la courbe des prix change, ce qui était payé reste payé
		k := p.scale1()
		for _, u := range p.Unlocks {
			if u.Paid > 0 {
				ps.Paid[u.ID] = u.Paid / k
			}
		}
		ps.Reputation = p.Rep
		ps.Achete = p.Owned
		ps.Servis = p.Served
		if p.Parc != nil {
			ps.Stands = p.Parc.StandSales
			ps.Visiteurs = p.saveVisitors()
		}
		if p.Crep != nil {
			for _, t := range p.Crep.Tables {
				if t.Tips > 0 {
					if ps.Pourboires == nil {
						ps.Pourboires = map[string]float64{}
					}
					ps.Pourboires[t.ID] = t.Tips
				}
			}
			ps.Attables = p.saveSeated()
		}
		if p.ID == "ecole" {
			ps.Eleves = p.saveStudents()
		}
		for _, sp := range p.Spots {
			if sp.Paid > 0 {
				ps.CasePaye[sp.ID] = sp.Paid / k
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
				ps.NiveauPaye[st.ID] = st.LvlPaid / k
			}
			if st.StkPaid > 0 {
				ps.StockPaye[st.ID] = st.StkPaid / k
			}
			if st.Kind == Field {
				ps.Pousses[st.ID] = len(st.Plants)
				if st.PlantPaid > 0 {
					ps.PoussePaye[st.ID] = st.PlantPaid / k
				}
			}
			if st.Kind == Machine || st.Kind == Dock {
				b := make([]int, NItems)
				copy(b, st.Buf[:])
				ps.Buf[st.ID] = b
			}
		}
		// les drones en vol vers un quai : leur chargement est posé sur ce quai (il ne se perd plus)
		for _, st := range p.Stations {
			for _, d := range st.Drones {
				if d.Phase == 2 && d.N > 0 && d.dock != nil {
					if b, ok := ps.Buf[d.dock.ID]; ok {
						b[d.It] += d.N
					}
				}
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
	migrateGlaces(s)
	migrateEpices(s)
	migrateAbattoirs(s)
	migrateAvocatier(s)
	migrateRayonsAgrumes(s)
	// crêperie : « Clients fidèles » (absences) est devenu « Clients heureux » (elle n'ouvre que quand on joue)
	// (de même au parc : « Visiteurs heureux »)
	for _, id := range []string{"creperie", "parc"} {
		if n := s.Levels[id+".pts_fidele"]; n > 0 {
			s.Levels[id+".pts_bonheur"] += n
			delete(s.Levels, id+".pts_fidele")
		}
	}
	g.Money = s.Money
	if s.Levels != nil {
		g.Levels = s.Levels
		// rentabilité : 450 points au plus ; ce qui dépasse retourne aux points ouvrier
		for k, n := range g.Levels {
			if strings.HasSuffix(k, ".pts_mac_rent") || strings.HasSuffix(k, ".pts_soin_rent") || strings.HasSuffix(k, ".pts_chef_rent") {
				if n > RentMax {
					g.Levels["pts.ouvrier"] += n - RentMax
					g.Levels[k] = RentMax
				}
			}
		}
		// parc : moins de places de parking et de trottinettes qu'avant : on borne
		if g.Levels["parc.parking"] > (ParcParkMax-parcParkStart)/10 {
			g.Levels["parc.parking"] = (ParcParkMax - parcParkStart) / 10
		}
		if g.Levels["parc.trott"] > ParcTrottMax/10 {
			g.Levels["parc.trott"] = ParcTrottMax / 10
		}
	}
	// la serre et le verger n'ont plus qu'une équipe chacun : les ouvriers d'une
	// colonne (anciennes sauvegardes) rejoignent l'équipe de la serre ou du verger
	for i, id := range ColumnIDs {
		k := "ferme." + teamKey(id)
		if n := g.Levels[k]; n > 0 {
			to := "ferme." + teamKey("serre")
			if i >= NSerre {
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
		// (montants écrits sans le coefficient du lieu : voir Snapshot ; une sauvegarde
		// d'avant la courbe des prix, avec les anciens prix, redonne donc tout ce qui était payé)
		k := p.scale1()
		for _, sp := range p.Spots {
			sp.Paid = ps.CasePaye[sp.ID] * k
		}
		if ps.Reputation >= 1 {
			p.Rep = ps.Reputation
		}
		p.Owned = ps.Achete || p.ID == "ferme"
		p.Served = ps.Servis
		if p.Parc != nil {
			p.Parc.StandSales = ps.Stands
		}
		if p.Crep != nil {
			for _, t := range p.Crep.Tables {
				t.Tips = ps.Pourboires[t.ID]
			}
		}
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
			u.Paid = ps.Paid[u.ID] * k
			if u.Paid >= u.Cost*0.99 && u.Cost > 0 { // (prix arrondis : niceRound)
				u.Paid = u.Cost
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
				st.PlantPaid = ps.PoussePaye[st.ID] * k
			}
			st.LvlPaid = ps.NiveauPaye[st.ID] * k
			st.StkPaid = ps.StockPaye[st.ID] * k
			st.Stock = min(ps.Stock[st.ID], st.StockMax)
			if st.Kind == Pallet && p.Owned {
				// (la taille de l'entrepôt d'abord : sinon les palettes étaient coupées à leur taille
				// de départ à chaque lancement, et les drones se jetaient sur les quais pour les remplir)
				st.PileMax = PalletMax(p.ID, g.Levels[p.key("entrepot")])
			}
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
	// la crêperie : ses clients attablés reprennent où ils en étaient
	for _, p := range g.Places {
		if p.Crep != nil && s.Places[p.ID] != nil {
			p.restoreSeated(s.Places[p.ID].Attables)
		}
		if p.Parc != nil && s.Places[p.ID] != nil && p.Owned {
			p.restoreVisitors(s.Places[p.ID].Visiteurs) // (le parc aussi : ses visiteurs reprennent leur visite)
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
		// (la crêperie n'en fait pas partie : son revenu est retiré de s.Rate, RateOffline)
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
			if p.Owned && p.sells() && p.Crep == nil && p.Parc == nil { // (la crêperie et le parc ne gagnent rien en ton absence)
				fid += p.ptsF("pts_fidele")
				nf++
			}
		}
		if nf > 0 {
			f *= 1 + fid/float64(nf)
		}
		rate := s.Rate * secs
		if v := s.Vol; v != nil && secs > v.Reste {
			rate += v.Taux * (secs - v.Reste) // (les voleurs sont partis pendant l'absence : le lieu a repris ensuite)
		}
		offline = math.Round(rate * f)
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
		g.income = append(g.income, incomeRec{t: g.T, a: s.Rate * 60})
	}
	g.Boost = s.Cadeau // (après les gains d'absence : un cadeau ne travaille pas jeu fermé)
	g.Custody = nil
	g.prisonRestore(s.Prison)
	for _, c := range s.Garde { // (le commissariat ne fonctionne pas jeu fermé : la garde à vue reprend où elle en était)
		left := c.Reste
		if left > 0 && c.N > 0 {
			g.Custody = append(g.Custody, &Custody{N: c.N, Left: left})
		}
	}
	if s.Vol != nil && s.Unix > 0 && nowUnix > s.Unix {
		s.Vol.Reste -= float64(nowUnix - s.Unix) // (le temps du vol s'écoule aussi jeu fermé)
	}
	g.theftRestore(s.Vol) // (s'il en reste, les voleurs sont toujours là)
	g.sessionStart = g.T
	return
}

// migrateGlaces : la glace à la cerise est devenue la glace à la vanille (même
// place au glacier, même prix) : ce qui avait été payé pour elle (fabrique,
// stock, drone de sa vitrine, rayon du supermarché) passe à la vanille.
func migrateGlaces(s *Save) {
	move := func(m map[string]float64, from, to string) {
		if v, ok := m[from]; ok && m != nil {
			if m[to] < v {
				m[to] = v
			}
			delete(m, from)
		}
	}
	if s.Levels != nil {
		for _, k := range [][2]string{
			{"glacier.glace_cerise_stock", "glacier.glace_vanille_stock"},
			{"glacier." + droneKey(GlaceCerise), "glacier." + droneKey(GlaceVanille)},
		} {
			if v, ok := s.Levels[k[0]]; ok {
				s.Levels[k[1]] = max(s.Levels[k[1]], v)
				delete(s.Levels, k[0])
			}
		}
	}
	if ps := s.Places["glacier"]; ps != nil {
		move(ps.Paid, "glace_cerise", "glace_vanille")
		move(ps.StockPaye, "glace_cerise", "glace_vanille")
		move(ps.CasePaye, droneKey(GlaceCerise), droneKey(GlaceVanille))
	}
	if ps := s.Places["super"]; ps != nil {
		move(ps.Paid, "r_"+itoa(int(GlaceCerise)), "r_"+itoa(int(GlaceVanille)))
	}
}

// migrateEpices : les épices de la ferme exotique s'achètent maintenant dans l'ordre
// de leurs champs (vanille, cannelle, thé, séchoir) ; le thé (95 000 → 105 000 de
// base) et le séchoir (105 000 → 120 000) coûtent plus : ce qui était entièrement
// payé à l'ancien prix reste payé (les montants partiels restent tels quels).
func migrateEpices(s *Save) {
	ps := s.Places["exotique"]
	if ps == nil || ps.Paid == nil {
		return
	}
	for id, c := range map[string][2]float64{"the": {95000, 105000}, "sechoir": {105000, 120000}} {
		if v := ps.Paid[id]; v >= c[0]*0.99 && v < c[1] {
			ps.Paid[id] = c[1]
		}
	}
}

// AbattoirsKey : marque des parties où les abattoirs de la ferme d'élevage s'achètent à
// part (toutes les nouvelles parties).
const AbattoirsKey = "elevage.abattoirs_a_part"

// abattoirOf : chaque abattoir et le bâtiment de ses animaux ("" : le poulailler, offert),
// avec son prix de base.
var abattoirOf = []struct {
	id, animal string
	cost       float64
}{
	{"abattoir_poulet", "", 4000}, {"abattoir_canard", "mare", 10000},
	{"abattoir_boeuf", "bovin", 25000}, {"abattoir_veau", "bovin", 30000},
	{"abattoir_mouton", "ovin", 35000}, {"abattoir_agneau", "ovin", 40000},
	{"abattoir_porc", "porc", 50000},
}

// migrateAbattoirs : avant, l'abattoir venait avec le bâtiment de ses animaux. Dans une
// ancienne partie, un abattoir dont les animaux sont achetés reste acheté (rien ne se
// perd) ; les suivants s'achètent à part.
func migrateAbattoirs(s *Save) {
	if s.Levels == nil {
		s.Levels = map[string]int{}
	}
	if s.Levels[AbattoirsKey] > 0 {
		return
	}
	s.Levels[AbattoirsKey] = 1
	ps := s.Places["elevage"]
	if ps == nil || !ps.Achete {
		return
	}
	if ps.Paid == nil {
		ps.Paid = map[string]float64{}
	}
	full := map[string]float64{"mare": 15000, "bovin": 40000, "ovin": 60000, "porc": 80000}
	for _, a := range abattoirOf {
		if a.animal == "" || ps.Paid[a.animal] >= full[a.animal]*0.99 {
			ps.Paid[a.id] = math.Max(ps.Paid[a.id], a.cost)
		}
	}
}

// AvocatierKey : marque des parties où l'avocatier existe (toutes les nouvelles parties).
const AvocatierKey = "exotique.avocatier"

// migrateAvocatier : avec l'avocatier (après le manguier), le citronnier, l'oranger et les
// deux pressoirs coûtent plus. Dans une ancienne partie, ce qui était entièrement payé à
// l'ancien prix reste payé (les montants partiels restent tels quels).
func migrateAvocatier(s *Save) {
	if s.Levels == nil {
		s.Levels = map[string]int{}
	}
	if s.Levels[AvocatierKey] > 0 {
		return
	}
	s.Levels[AvocatierKey] = 1
	ps := s.Places["exotique"]
	if ps == nil || ps.Paid == nil {
		return
	}
	for id, c := range map[string][2]float64{"citronnier": {34000, 36000}, "oranger": {36000, 38000},
		"pressoir_citron": {37000, 40000}, "pressoir_orange": {40000, 43000}} {
		if v := ps.Paid[id]; v >= c[0]*0.99 && v < c[1] {
			ps.Paid[id] = c[1]
		}
	}
}

// RayonsAgrumesKey : marque des parties où les rayons des agrumes ont leurs prix d'après l'avocatier.
const RayonsAgrumesKey = "super.rayons_agrumes"

// migrateRayonsAgrumes : le prix d'un rayon suit celui de ce qui le fabrique (fixRayonCosts) ;
// avec l'avocatier, les rayons citrons, oranges, limonade et jus d'orange coûtent plus.
// Un rayon payé à l'ancien prix reste payé (sinon il resterait 1 000 à 1 500 à payer).
func migrateRayonsAgrumes(s *Save) {
	if s.Levels == nil {
		s.Levels = map[string]int{}
	}
	if s.Levels[RayonsAgrumesKey] > 0 {
		return
	}
	s.Levels[RayonsAgrumesKey] = 1
	ps := s.Places["super"]
	if ps == nil || ps.Paid == nil {
		return
	}
	for it, c := range map[Item][2]float64{Citron: {19500, 20500}, Orange: {20500, 21500},
		Limonade: {21000, 22500}, JusOrange: {22500, 24000}} {
		id := "r_" + itoa(int(it))
		if v := ps.Paid[id]; v >= c[0]*0.99 && v < c[1] {
			ps.Paid[id] = c[1]
		}
	}
}
