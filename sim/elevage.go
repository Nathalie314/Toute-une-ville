package sim

import (
	"math"
	"math/rand"
	"strings"
)

// Étape 6 : la ferme d'élevage et la boucherie-charcuterie.
//
// La ferme d'élevage n'a pas de clients : elle reçoit de la ferme du maïs, du
// foin et des pommes de terre, élève les animaux, les abat et range les
// carcasses dans ses frigos. Des drones les portent au quai du camion
// frigorifique, qui part vers la boucherie.
//
// Les deux lieux sont découpés en espaces (volaille, bovin, ovin, porc) : chaque
// espace a ses propres employés, qui ne travaillent que chez eux.

// ShopZone : un espace d'un lieu à équipes.
type ShopZone struct {
	ID, Name string
	X0, X1   float64 // étendue (pour le sol peint et le repos)
	RestY    float64
	Costs    []float64 // prix des embauches (un par employé)
}

// Max : nombre d'employés au plus.
func (z *ShopZone) Max() int { return len(z.Costs) }

var elevageZones = []*ShopZone{
	{ID: "volaille", Name: "Volaille", X0: 150, X1: 1110, RestY: 1262, Costs: []float64{3000, 9000, 24000, 55000}},
	{ID: "bovin", Name: "Bovins", X0: 1160, X1: 2010, RestY: 1262, Costs: []float64{12000, 30000, 70000}},
	{ID: "ovin", Name: "Ovins", X0: 2060, X1: 2910, RestY: 1262, Costs: []float64{18000, 42000, 90000}},
	{ID: "porc", Name: "Porcs", X0: 2960, X1: 3410, RestY: 1262, Costs: []float64{24000, 55000}},
}

var boucherieZones = []*ShopZone{
	{ID: "volaille", Name: "Volaille", X0: 440, X1: 2300, RestY: 775, Costs: []float64{5000, 14000, 35000, 70000, 130000, 220000}},
	{ID: "bovin", Name: "Bœuf et veau", X0: 2300, X1: 3540, RestY: 775, Costs: []float64{16000, 40000, 90000, 180000}},
	{ID: "ovin", Name: "Mouton et agneau", X0: 3540, X1: 4160, RestY: 775, Costs: []float64{22000, 60000}},
	{ID: "porc", Name: "Charcuterie", X0: 4160, X1: 5400, RestY: 775, Costs: []float64{28000, 65000, 140000, 280000}},
}

// TeamZones : les espaces d'un lieu (nil : pas d'espaces).
func TeamZones(place string) []*ShopZone { return teamZonesOf(place) }

func teamZonesOf(place string) []*ShopZone {
	switch place {
	case "elevage":
		return elevageZones
	case "boucherie":
		return boucherieZones
	case "exotique":
		return exoticZones
	case "domaine":
		return domaineZones
	}
	return nil
}

func shopZone(place, id string) *ShopZone {
	for _, z := range teamZonesOf(place) {
		if z.ID == id {
			return z
		}
	}
	return nil
}

// ZoneOpen : l'espace a-t-il au moins une station construite ?
func (p *Place) ZoneOpen(id string) bool {
	for _, s := range p.Stations {
		if s.Team == id && s.Built {
			return true
		}
	}
	return false
}

// prix des deux lieux
const (
	elevagePrice   = 90000
	boucheriePrice = 140000
)

// animaux : les bâtiments d'élevage (points « soigneurs ») ; les abattoirs sont des machines.
var elevageAnimals = map[string]bool{"poulailler": true, "mare": true, "pre_bovin": true, "pre_ovin": true, "porcherie": true}

func newElevage(g *Game) *Place {
	W := 3850.0 // (bâtiments d'animaux et abattoirs espacés : leur case « Stock + » à droite)
	p := &Place{ID: "elevage", Name: "La ferme d'élevage", W: W, H: 1800, g: g, rng: rand.New(rand.NewSource(53)),
		Rep: 3, Price: elevagePrice, SellMult: 1, SpawnBase: 5, WantMax: 3, MaxCust: 0, LeaveOff: V{170, 60}, OnlyStocked: true}
	p.SpawnL = V{350, roadY + 40}
	p.SpawnR = V{W - 550, roadY + 40}
	p.Gate = V{1250, roadY + 40}
	p.Inside = V{1250, 1390}
	p.H = ParkH // (pas de parking : pas de clients)
	p.RestSpot = V{W - 550, 1440}

	// réserve : maïs (volaille), foin (bovins et ovins), pommes de terre (porcs)
	for _, pl := range []struct {
		it Item
		x  float64
	}{{Mais, 330}, {Foin, 1960}, {PommeDeTerre, 3100}} {
		s := &Station{ID: "palette_" + itoa(int(pl.it)), Name: "Palette : " + Items[pl.it].Name, Kind: Pallet, Item: pl.it, PileMax: 40, Built: true}
		s.Body = Rect{pl.x, 150, 160, 90}
		s.Zone = Rect{pl.x - 5, 258, 170, 90}
		p.Stations = append(p.Stations, s)
	}
	const rowA, rowB = 400.0, 735.0
	mach := func(id, name, team string, in, out Item, x, y, w, h, t float64, unl string) *Station {
		s := &Station{ID: id, Name: name, Kind: Machine, Item: out, Inputs: []Item{in}, Time: t, PileMax: 12, BufMax: 10, Unlock: unl, BaseCost: 3000, Team: team}
		s.Body = Rect{x, y, w, h}
		s.Zone = Rect{x - 10, y + h + 18, 130, 90}
		s.OutZone = Rect{x + 140, y + h + 18, 130, 90}
		p.Stations = append(p.Stations, s)
		return s
	}
	pre := func(id, name, team string, in, a, b Item, x, t float64, unl string) {
		s := mach(id, name, team, in, a, x, rowA, 420, 180, t, unl)
		s.Two, s.Item2 = true, b
		s.OutZone2 = Rect{x + 280, rowA + 180 + 18, 130, 90}
	}
	// (chaque abattoir s'achète à part, après le bâtiment de ses animaux ; son frigo vient avec lui)
	frigo := func(team string, it Item, x float64, unl string) {
		s := &Station{ID: "frigo_" + itoa(int(it)), Name: "Frigo : " + Items[it].Name, Kind: Shelf, Item: it, StockMax: 30, Unlock: unl, NoSale: true, Team: team}
		s.Body = Rect{x, shelfY, 180, shelfH}
		s.Zone = Rect{x - 10, shelfY - 95, 200, 92}
		s.Spot = V{x + 90, shelfY + shelfH + 38}
		p.Stations = append(p.Stations, s)
	}
	// volaille : le poulailler et la mare (maïs), l'abattoir à deux entrées
	// (chaque bâtiment d'animaux a sa case « Stock + » à sa droite : on laisse la place)
	mach("poulailler", "Poulailler", "volaille", Mais, Poulet, 200, rowA, 260, 180, 4, "")
	mach("mare", "Mare aux canards", "volaille", Mais, Canard, 680, rowA, 260, 180, 5, "mare")
	mach("abattoir_poulet", "Abattoir : poulets", "volaille", Poulet, PouletAbattu, 200, rowB, 260, 140, 3, "abattoir_poulet")
	mach("abattoir_canard", "Abattoir : canards", "volaille", Canard, CanardAbattu, 680, rowB, 260, 140, 3, "abattoir_canard") // (sous la mare)
	frigo("volaille", PouletAbattu, 200, "abattoir_poulet")
	frigo("volaille", CanardAbattu, 680, "abattoir_canard")
	// bovins : un grand pré (foin → vaches et veaux), ses deux abattoirs
	pre("pre_bovin", "Pré des bovins", "bovin", Foin, Vache, Veau, 1210, 6, "bovin")
	mach("abattoir_boeuf", "Abattoir : bœufs", "bovin", Vache, Boeuf, 1210, rowB, 260, 140, 4, "abattoir_boeuf")
	mach("abattoir_veau", "Abattoir : veaux", "bovin", Veau, VeauCarcasse, 1600, rowB, 260, 140, 4, "abattoir_veau")
	frigo("bovin", Boeuf, 1210, "abattoir_boeuf")
	frigo("bovin", VeauCarcasse, 1450, "abattoir_veau")
	// ovins : un grand pré (foin → moutons et agneaux), ses deux abattoirs
	pre("pre_ovin", "Pré des ovins", "ovin", Foin, Mouton, Agneau, 2110, 5, "ovin")
	mach("abattoir_mouton", "Abattoir : moutons", "ovin", Mouton, MoutonCarcasse, 2110, rowB, 260, 140, 4, "abattoir_mouton")
	mach("abattoir_agneau", "Abattoir : agneaux", "ovin", Agneau, AgneauCarcasse, 2500, rowB, 260, 140, 4, "abattoir_agneau")
	frigo("ovin", MoutonCarcasse, 2110, "abattoir_mouton")
	frigo("ovin", AgneauCarcasse, 2350, "abattoir_agneau")
	// porcs : la porcherie (pommes de terre), son abattoir
	mach("porcherie", "Porcherie", "porc", PommeDeTerre, Cochon, 3010, rowA, 260, 180, 5, "porc")
	mach("abattoir_porc", "Abattoir : porcs", "porc", Cochon, Porc, 3010, rowB, 260, 140, 4, "abattoir_porc")
	frigo("porc", Porc, 3060, "abattoir_porc")

	trash := &Station{ID: "poubelle", Name: "Poubelle", Kind: Trash}
	trash.Body = Rect{3530, rowA + 50, 70, 70} // (le compost, à droite de la porcherie et de sa case « Stock + »)
	trash.Zone = Rect{3510, rowA + 10, 110, 150}
	p.Stations = append(p.Stations, trash)
	p.dock("quai", "boucherie", 1740) // le camion frigorifique (entre les frigos des bovins et ceux des ovins)

	// déblocages, espace par espace
	unlock := func(id, name string, cost float64, hint, needs, locked string) {
		var at Rect
		for _, s := range p.Stations {
			if s.Unlock == id && s.Kind == Machine {
				c := s.Body.C()
				at = Rect{c.X - 90, c.Y - 65, 180, 130}
				break
			}
		}
		u := &Unlock{ID: id, Name: name, Cost: cost, Zone: at, Hint: hint}
		if needs != "" {
			u.Requires = func() bool { return g.needsMet(needs) }
			u.Locked = locked
		}
		p.Unlocks = append(p.Unlocks, u)
	}
	unlock("abattoir_poulet", "Abattoir : poulets", 4000, "Les poulets du poulailler y deviennent des poulets abattus, rangés au frigo.", "", "")
	unlock("mare", "Mare aux canards", 15000, "Les canards mangent du maïs.", "", "")
	unlock("abattoir_canard", "Abattoir : canards", 10000, "Les canards de la mare y deviennent des canards abattus.", "elevage:mare", "Il faut d'abord la mare aux canards")
	unlock("bovin", "Pré des bovins", 40000, "Vaches, taureau et veaux mangent du foin.", "ferme:grenier", "Il faut le grenier à foin (ma ferme)")
	unlock("abattoir_boeuf", "Abattoir : bœufs", 25000, "Les vaches et le taureau y deviennent du bœuf.", "elevage:bovin", "Il faut d'abord le pré des bovins")
	unlock("abattoir_veau", "Abattoir : veaux", 30000, "Les veaux y deviennent des carcasses de veau.", "elevage:bovin", "Il faut d'abord le pré des bovins")
	unlock("ovin", "Pré des ovins", 60000, "Brebis, moutons et agneaux mangent du foin.", "ferme:grenier", "Il faut le grenier à foin (ma ferme)")
	unlock("abattoir_mouton", "Abattoir : moutons", 35000, "Les moutons y deviennent des carcasses de mouton.", "elevage:ovin", "Il faut d'abord le pré des ovins")
	unlock("abattoir_agneau", "Abattoir : agneaux", 40000, "Les agneaux y deviennent des carcasses d'agneau.", "elevage:ovin", "Il faut d'abord le pré des ovins")
	unlock("porc", "Porcherie", 80000, "Les cochons mangent des pommes de terre.", "ferme:pdt", "Il faut les pommes de terre (serre)")
	unlock("abattoir_porc", "Abattoir : porcs", 50000, "Les cochons y deviennent du porc.", "elevage:porc", "Il faut d'abord la porcherie")
	for _, s := range p.Stations {
		if s.Unlock == "" {
			s.Built = true
		}
	}

	// améliorations
	k := p.key
	p.ups = []*Upgrade{
		{Key: k("etals"), Name: "Frigos plus grands", Max: 6, Base: 6000, Mult: 2.3, Desc: func(l int) string { return "Chaque frigo contient " + itoa(30+4*l) + " carcasses." }},
		{Key: k("entrepot"), Name: "Entrepôt", Max: 12, Base: 4000, Mult: 1.6, Desc: func(l int) string {
			return "Chaque palette de la réserve contient " + itoa(PalletMax(p.ID, l)) + " produits."
		}},
	}
	p.ups = append(p.ups, p.teamUpgrades("Éleveur", "Il ne travaille que dans son espace : bâtiments, abattoir et frigos.")...)
	p.ups = append(p.ups, p.droneUpgrades(3.5)...)
	for _, st := range p.Stations {
		if st.Kind == Shelf {
			it := st.Item
			p.Spots = append(p.Spots, &Spot{ID: droneKey(it), Label: "Drone", Shelf: st, Zone: st.DronePad(),
				Done: func(int) string { return "Drone acheté pour le frigo « " + Items[it].Name + " » !" }})
		}
	}
	p.Spots = append(p.Spots,
		&Spot{ID: "etals", Label: "Frigos +", Done: func(l int) string { return "Frigos agrandis : niveau " + itoa(l) + " !" }},
		&Spot{ID: "entrepot", Label: "Entrepôt", Zone: Rect{150, 170, 130, 124}, Done: func(l int) string { return "Entrepôt agrandi : " + itoa(PalletMax(p.ID, l)) + " par palette !" }})
	p.setupYard(-620)
	p.teamSpots("Éleveur")
	// les frigos s'agrandissent dans la zone du joueur, après les éleveurs
	p.spot("etals").Zone = Rect{p.Yard.X + 10 + float64(len(teamZonesOf(p.ID)))*135, yardRow1, 130, 124}
	return p
}

// teamUpgrades : l'embauche des employés de chaque espace.
func (p *Place) teamUpgrades(who, desc string) []*Upgrade {
	var ups []*Upgrade
	for _, z := range teamZonesOf(p.ID) {
		z := z
		d := desc
		if p.ID == "exotique" {
			// on nomme tout ce dont il s'occupe (« Épices » : pas seulement la vanille)
			var names []string
			for _, s := range p.Stations {
				if s.Team == z.ID && (s.Kind == Field || s.Kind == Machine) {
					r := []rune(s.Name)
					names = append(names, strings.ToLower(string(r[:1]))+string(r[1:]))
				}
			}
			if len(names) > 1 {
				d = "Il s'occupe de tout l'espace « " + z.Name + " » : " + strings.Join(names[:len(names)-1], ", ") + " et " + names[len(names)-1] + "."
			}
		}
		ups = append(ups, &Upgrade{Key: p.key(teamKey(z.ID)), Name: who + " (" + z.Name + ")", Max: z.Max(), Now: true, Costs: z.Costs,
			Desc: func(l int) string { return d + " Embauchés : " + itoa(l) + "." }})
	}
	return ups
}

// teamSpots : les embauches, dans la zone du joueur (une case par espace).
func (p *Place) teamSpots(who string) {
	for i, z := range teamZonesOf(p.ID) {
		z := z
		p.Spots = append(p.Spots, &Spot{ID: teamKey(z.ID), Label: z.Name, Team: z.ID,
			Zone: Rect{p.Yard.X + 10 + float64(i)*135, yardRow1, 130, 124},
			Done: func(l int) string {
				return "Nouvel employé (" + z.Name + ") : +1 !"
			}})
	}
	_ = who
}

// ---------------------------------------------------------------- boucherie

// emplacement de chaque recette de la boucherie : colonne et rangée
// (volaille : colonnes 0 à 2, bovins : 3 et 4, ovins : 5, porc : 6 et 7).
var boucherieCell = map[string][3]int{ // colonne, rangée, espace
	"rotisserie": {0, 0, 0}, "cuisses": {1, 0, 0}, "pilons": {2, 0, 0},
	"aiguillettes": {0, 1, 0}, "magret": {1, 1, 0}, "foie_gras": {2, 1, 0},
	"steak": {3, 0, 1}, "hache": {4, 0, 1}, "roti_boeuf": {3, 1, 1}, "veau": {4, 1, 1},
	"gigot": {5, 0, 2}, "cote_agneau": {5, 1, 2},
	"saucisse": {6, 0, 3}, "jambon": {7, 0, 3}, "cotelette": {6, 1, 3}, "pate": {7, 1, 3},
}

var boucherieSpec = workshopSpec{
	id: "boucherie", name: "La boucherie-charcuterie", price: boucheriePrice, seed: 59,
	inputs: []Item{PouletAbattu, CanardAbattu, Boeuf, VeauCarcasse, MoutonCarcasse, AgneauCarcasse, Porc},
	recipes: []recipe{
		{"rotisserie", "Rôtisserie", PouletRoti, []Item{PouletAbattu}, 5, 0, "", ""}, // (la préparation des poulets n'existe plus : la rôtisserie est offerte)
		{"cuisses", "Cuisses de poulet", CuissePoulet, []Item{PouletAbattu}, 4, 20000, "Avec les poulets de la ferme d'élevage.", ""},
		{"pilons", "Pilons de poulet", PilonPoulet, []Item{PouletAbattu}, 4, 25000, "Avec les poulets de la ferme d'élevage.", ""},
		{"aiguillettes", "Aiguillettes de canard", Aiguillettes, []Item{CanardAbattu}, 4, 30000, "Avec les canards de la ferme d'élevage.", "elevage:abattoir_canard"},
		{"magret", "Magrets de canard", Magret, []Item{CanardAbattu}, 5, 45000, "Avec les canards de la ferme d'élevage.", "elevage:abattoir_canard"},
		{"foie_gras", "Foie gras", FoieGras, []Item{CanardAbattu}, 7, 70000, "Avec les canards de la ferme d'élevage.", "elevage:abattoir_canard"},
		{"steak", "Steaks", Steak, []Item{Boeuf}, 4, 40000, "Le bœuf de la ferme d'élevage devient des steaks.", "elevage:abattoir_boeuf"},
		{"hache", "Hachoir", SteakHache, []Item{Steak}, 3, 55000, "Les steaks deviennent des steaks hachés.", "elevage:abattoir_boeuf"},
		{"roti_boeuf", "Rôtis de bœuf", RotiBoeuf, []Item{Boeuf}, 5, 75000, "Avec le bœuf de la ferme d'élevage.", "elevage:abattoir_boeuf"},
		{"veau", "Morceaux de veau", MorceauxVeau, []Item{VeauCarcasse}, 5, 90000, "Avec les veaux de la ferme d'élevage.", "elevage:abattoir_veau"},
		{"gigot", "Gigots", Gigot, []Item{MoutonCarcasse}, 5, 60000, "Avec les moutons de la ferme d'élevage.", "elevage:abattoir_mouton"},
		{"cote_agneau", "Côtes d'agneau", CoteAgneau, []Item{AgneauCarcasse}, 5, 85000, "Avec les agneaux de la ferme d'élevage.", "elevage:abattoir_agneau"},
		{"saucisse", "Saucisses", Saucisse, []Item{Porc}, 4, 50000, "Le porc de la ferme d'élevage devient des saucisses.", "elevage:abattoir_porc"},
		{"jambon", "Jambons", Jambon, []Item{Porc}, 5, 65000, "Avec le porc de la ferme d'élevage.", "elevage:abattoir_porc"},
		{"cotelette", "Côtelettes", Cotelette, []Item{Porc}, 4, 80000, "Avec le porc de la ferme d'élevage.", "elevage:abattoir_porc"},
		{"pate", "Pâtés", Pate, []Item{Porc}, 6, 100000, "Avec le porc de la ferme d'élevage.", "elevage:abattoir_porc"},
	},
	staff: "Bouchers", cashCost: 8000, staffCosts: []float64{1}, upBase: 11000,
	sign: "Vitrine réfrigérée", bio: "Viande d'ici", w: 5700, perRow: 8, tall: true, docks: []string{"super", "creperie"},
}

func newBoucherie(g *Game) *Place {
	p := newWorkshop(g, boucherieSpec)
	p.SpawnBase = 11 // les clients sont rares (mais achètent cher)
	// machines et vitrines : chaque espace dans ses colonnes
	sy := float64(workshopShY)
	for _, r := range boucherieSpec.recipes {
		c := boucherieCell[r.id]
		team := boucherieZones[c[2]].ID
		x := 500 + float64(c[0])*620
		y := workshopRow0 + float64(c[1])*workshopRowG
		s := p.Station(r.id)
		s.Team = team
		s.Body = Rect{x, y, 260, 150}
		s.Zone = Rect{x - 10, y + 168, 130, 90}
		s.OutZone = Rect{x + 140, y + 168, 130, 90}
		sh := p.Station("etal_" + itoa(int(r.out)))
		sh.Team = team
		// les vitrines : dans chaque espace, dans l'ordre où l'on achète les machines
		k := 0
		for _, o := range boucherieSpec.recipes {
			if boucherieCell[o.id][2] == c[2] && (o.cost < r.cost || (o.cost == r.cost && o.id < r.id)) {
				k++
			}
		}
		first := 99
		for _, o := range boucherieSpec.recipes {
			if cc := boucherieCell[o.id]; cc[2] == c[2] {
				first = min(first, cc[0])
			}
		}
		sx := 500 + float64(first)*620 + float64(k)*200
		sh.Body = Rect{sx, sy, 150, shelfH}
		sh.Zone = Rect{sx - 10, sy - 95, 170, 92}
		sh.Spot = V{sx + 75, sy + shelfH + 38}
		for _, u := range p.Unlocks {
			if u.ID == r.id {
				c := s.Body.C()
				u.Zone = Rect{c.X - 90, c.Y - 65, 180, 130}
			}
		}
	}
	// palettes au-dessus de leur espace
	palX := map[Item]float64{PouletAbattu: 600, CanardAbattu: 1000, Boeuf: 2500, VeauCarcasse: 2900, MoutonCarcasse: 3600, AgneauCarcasse: 3820, Porc: 4500}
	for _, s := range p.Stations {
		if s.Kind == Pallet {
			x := palX[s.Item]
			s.Body = Rect{x, 170, 160, 90}
			s.Zone = Rect{x - 5, 280, 170, 90}
		}
	}
	// les drones suivent leur vitrine
	for _, sp := range p.Spots {
		if sp.Shelf != nil {
			sp.Zone = sp.Shelf.DronePad()
		}
	}
	// pas d'« ouvriers » communs : des bouchers par espace
	ups := p.ups[:0]
	for _, u := range p.ups {
		if u.Key != p.key("ouvriers") {
			ups = append(ups, u)
		}
	}
	p.ups = append(ups, p.teamUpgrades("Boucher", "Il ne travaille que dans son espace : machines et vitrines.")...)
	spots := p.Spots[:0]
	for _, sp := range p.Spots {
		if sp.ID != "ouvriers" {
			spots = append(spots, sp)
		}
	}
	p.Spots = spots
	p.teamSpots("Boucher")
	return p
}

// firstInZone : à la boucherie, chaque espace montre son prochain déblocage.
func (p *Place) firstInZone(u *Unlock) bool {
	team := ""
	for _, s := range p.Stations {
		if s.Unlock == u.ID && s.Kind == Machine {
			team = s.Team
		}
	}
	for _, v := range p.Unlocks {
		if v.Done {
			continue
		}
		for _, s := range p.Stations {
			if s.Unlock == v.ID && s.Kind == Machine && s.Team == team {
				return v == u
			}
		}
	}
	return false
}

// zoneRest : où attendent les employés d'un espace.
func (p *Place) zoneRest(id string) V {
	if z := shopZone(p.ID, id); z != nil {
		return V{(z.X0 + z.X1) / 2, z.RestY}
	}
	return p.RestSpot
}

var _ = math.Max
