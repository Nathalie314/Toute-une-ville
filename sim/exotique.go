package sim

import "math/rand"

// La ferme exotique (après la crèmerie). Comme ma ferme : des champs, des
// bâtiments, des étals, des clients, des commandes de particuliers, des drones
// et des camions (supermarché, boulangerie, glacier).
//
//   fruits  : ananas, bananier, cocotier, manguier (une colonne chacun, comme le verger)
//   sucre   : canne à sucre → raffinerie (sucre en morceaux) → moulin (sucre en poudre)
//   cacao   : cacaoyers → chocolaterie (cacao + sucre en poudre → tablettes)
//   café    : caféiers → torréfacteur (café en grains) → moulin à café (café moulu)
//   riz     : une grande rizière (le riz se vend en sachets)
//   épices  : vanille, théiers → séchoir (boîtes de thé), cannelle
//
// Chaque espace a ses ouvriers (comme à la ferme d'élevage).

const exoticPrice = 35000

var exoticZones = []*ShopZone{
	{ID: "fruits", Name: "Fruits", X0: 420, X1: 1150, RestY: 970, Costs: []float64{3000, 12000, 30000, 60000}},
	{ID: "sucre", Name: "Sucre", X0: 1230, X1: 2050, RestY: 970, Costs: []float64{5000, 16000, 40000}},
	{ID: "cacao", Name: "Cacao", X0: 2130, X1: 2550, RestY: 970, Costs: []float64{9000, 24000}},
	{ID: "cafe", Name: "Café", X0: 2630, X1: 3450, RestY: 970, Costs: []float64{14000, 32000, 70000}},
	{ID: "riz", Name: "Riz", X0: 3540, X1: 4420, RestY: 970, Costs: []float64{20000}},
	{ID: "epices", Name: "Épices", X0: 4520, X1: 5600, RestY: 970, Costs: []float64{28000, 55000, 100000, 170000}},
}

// exoColumnIDs : les colonnes d'arbres exotiques.
var exoColumnIDs = []string{"ananas", "bananier", "cocotier", "manguier"}

// ExoDecoTop : le haut du décor de la ferme exotique (forêt, cascade, rivière).
const ExoDecoTop = -950.0

func newExotic(g *Game) *Place {
	W := 5800.0                                                                      // (aéré : on voit toutes les cases et on passe entre les champs)
	p := &Place{ID: "exotique", Name: "La ferme exotique", W: W, H: 1800, Top: -400, // (on monte jusqu'à la rivière)
		g: g, rng: rand.New(rand.NewSource(61)),
		Rep: 3, Price: exoticPrice, SellMult: 1, SpawnBase: 5, WantMax: 3, MaxCust: 14, LeaveOff: V{170, 60}, RepGain: 0.012,
		DecoTop: ExoDecoTop} // (au nord : la cascade, la rivière et la forêt)
	p.SpawnL = V{350, roadY + 40}
	p.SpawnR = V{W - 550, roadY + 40}
	p.Gate = V{1250, roadY + 40}
	p.Inside = V{1250, 1390}
	p.RestSpot = V{W - 400, 1440}

	field := func(id, name, team string, it Item, x, t float64, cols int, unl string) *Station {
		s := &Station{ID: id, Name: name, Kind: Field, Item: it, Time: t, Unlock: unl, Team: team}
		s.Body = Rect{x, fieldY, 40 + 80*float64(cols), fieldH}
		s.Zone = s.Body
		for k, r := range []int{2, 1, 0} {
			for i := 0; i < cols; i++ {
				c := i
				if k%2 == 1 {
					c = cols - 1 - i
				}
				s.Slots = append(s.Slots, V{x + 60 + float64(c)*80, fieldY + 60 + float64(r)*90})
			}
		}
		s.Plants = []Plant{{P: s.Slots[0], G: 1}}
		p.Stations = append(p.Stations, s)
		return s
	}
	machine := func(id, name, team string, out Item, in []Item, x, t float64, unl string) *Station {
		s := &Station{ID: id, Name: name, Kind: Machine, Item: out, Inputs: in, Time: t, PileMax: 12, BufMax: 10, Unlock: unl, Team: team}
		s.Body = Rect{x, bldY, 260, bldH}
		s.Zone = Rect{x - 10, bldY + bldH + 18, 130, 90}
		s.OutZone = Rect{x + 140, bldY + bldH + 18, 130, 90}
		p.Stations = append(p.Stations, s)
		return s
	}
	shelf := func(team string, it Item, x, w float64, unl string, market bool) {
		s := &Station{ID: "etal_" + itoa(int(it)), Name: "Étal : " + Items[it].Name, Kind: Shelf, Item: it, StockMax: 12, Unlock: unl, Team: team, Market: market}
		s.Body = Rect{x, shelfY, w, shelfH}
		s.Zone = Rect{x - 8, shelfY - 95, w + 16, 92}
		s.Spot = V{x + w/2, shelfY + shelfH + 38}
		p.Stations = append(p.Stations, s)
	}

	// --- fruits : quatre colonnes d'arbres (l'ananas, lui, pousse au sol…)
	names := []string{"Ananas", "Bananier", "Cocotier", "Manguier"}
	for i, it := range ExoFruits {
		id := exoColumnIDs[i]
		x := 440 + float64(i)*170
		s := &Station{ID: id, Name: names[i], Kind: Field, Item: it, Time: 9 + float64(i), Unlock: id, Column: true, Team: "fruits"}
		bottom := fieldY + 60 + 2*90
		s.Body = Rect{x + 10, bottom - float64(ColumnMax-1)*90 - 60, 130, float64(ColumnMax-1)*90 + 120}
		s.Zone = s.Body
		for k := 0; k < ColumnMax; k++ {
			s.Slots = append(s.Slots, V{x + 75, bottom - float64(k)*90})
		}
		s.Plants = []Plant{{P: s.Slots[0], Young: 1}} // l'arbre pousse d'abord
		p.Stations = append(p.Stations, s)
		shelf("fruits", it, x+12, 126, id, true)
	}
	if a := p.Station("ananas"); a != nil {
		a.Unlock = "" // la première colonne est offerte avec la ferme
		p.Station("etal_" + itoa(int(Ananas))).Unlock = ""
	}
	// --- sucre
	field("canne", "Canne à sucre", "sucre", Canne, 1250, 6, 6, "canne")
	machine("raffinerie", "Raffinerie", "sucre", SucreMorceaux, []Item{Canne}, 1250, 3, "raffinerie")
	ms := machine("moulin_sucre", "Moulin à sucre", "sucre", SucrePoudre, []Item{SucreMorceaux}, 1650, 2.5, "moulin_sucre")
	ms.Shared = true // la chocolaterie (espace cacao) s'y sert aussi
	shelf("sucre", SucreMorceaux, 1260, 220, "raffinerie", false)
	shelf("sucre", SucrePoudre, 1660, 220, "moulin_sucre", false)
	// --- cacao
	field("cacao", "Cacaoyers", "cacao", Cacao, 2150, 8, 4, "cacao")
	machine("chocolaterie", "Chocolaterie", "cacao", Chocolat, []Item{Cacao, SucrePoudre}, 2150, 4, "chocolaterie")
	shelf("cacao", Chocolat, 2160, 220, "chocolaterie", false)
	// --- café
	field("cafe", "Caféiers", "cafe", CafeVert, 2650, 8, 6, "cafe")
	machine("torrefacteur", "Torréfacteur", "cafe", CafeGrains, []Item{CafeVert}, 2650, 3.5, "torrefacteur")
	machine("moulin_cafe", "Moulin à café", "cafe", CafeMoulu, []Item{CafeGrains}, 3050, 3, "moulin_cafe")
	shelf("cafe", CafeGrains, 2660, 220, "torrefacteur", false)
	shelf("cafe", CafeMoulu, 3060, 220, "moulin_cafe", false)
	// --- la rizière
	field("riz", "Rizière", "riz", Riz, 3560, 6, 10, "riz")
	shelf("riz", Riz, 3800, 360, "riz", false)
	// --- épices : vanille, thé (et son séchoir), cannelle
	field("vanille", "Vanille", "epices", Vanille, 4540, 12, 3, "vanille")
	field("the", "Théiers", "epices", FeuillesThe, 4920, 7, 3, "the")
	field("cannelle", "Cannelliers", "epices", Cannelle, 5300, 11, 3, "cannelle")
	machine("sechoir", "Séchoir à thé", "epices", The, []Item{FeuillesThe}, 4930, 3.5, "sechoir")
	shelf("epices", Vanille, 4550, 220, "vanille", false)
	shelf("epices", The, 4930, 220, "sechoir", false)
	shelf("epices", Cannelle, 5310, 220, "cannelle", false)

	reg := &Station{ID: "caisse", Name: "Caisse", Kind: Register, HireKey: "caissiere", QueueDown: 5}
	reg.Body = Rect{180, shelfY, 210, shelfH}
	reg.Spot = V{285, shelfY - 40}
	reg.Zone = Rect{200, shelfY - 100, 170, 96}
	reg.QueueHead = V{285, shelfY + shelfH + 40}
	p.Stations = append(p.Stations, reg)
	trash := &Station{ID: "poubelle", Name: "Compost", Kind: Trash}
	trash.Body = Rect{W - 150, 925, 84, 76}
	trash.Zone = Rect{W - 170, 880, 124, 150}
	p.Stations = append(p.Stations, trash)
	p.dock("quai", "super", 1500)
	p.dock("quai_boul", "boulangerie", 1850)
	p.dock("quai_glace", "glacier", 2200)
	cmd := &Station{ID: "commande", Name: "Commandes", Kind: OrderBox, Built: true}
	p.Stations = append(p.Stations, cmd)

	// --- déblocages, dans l'ordre
	unlock := func(id, name string, cost float64, hint string) {
		var at Rect
		for _, s := range p.Stations {
			if s.Unlock == id && s.Kind != Shelf {
				c := s.Body.C()
				at = Rect{c.X - 90, c.Y - 65, 180, 130}
				if s.Column {
					at = Rect{s.Body.X, 290, 130, 130}
				}
				break
			}
		}
		p.Unlocks = append(p.Unlocks, &Unlock{ID: id, Name: name, Cost: cost, Zone: at, Hint: hint})
	}
	fruitHint := "Un arbre : il pousse d'abord, puis donne des fruits sans qu'on replante. Achète des cases au-dessus."
	unlock("canne", "Canne à sucre", 3000, "La canne ne se vend pas : elle part à la raffinerie.")
	unlock("raffinerie", "Raffinerie", 6000, "La canne devient du sucre en morceaux.")
	unlock("bananier", "Bananier", 8000, fruitHint)
	unlock("moulin_sucre", "Moulin à sucre", 11000, "Le sucre en morceaux devient du sucre en poudre (la boulangerie et le glacier en veulent).")
	unlock("cocotier", "Cocotier", 15000, fruitHint)
	unlock("cacao", "Cacaoyers", 20000, "Le cacao ne se vend pas : il part à la chocolaterie.")
	unlock("chocolaterie", "Chocolaterie", 26000, "Cacao + sucre en poudre = des tablettes de chocolat.")
	unlock("manguier", "Manguier", 32000, fruitHint)
	unlock("cafe", "Caféiers", 38000, "Le café vert ne se vend pas : il part au torréfacteur.")
	unlock("torrefacteur", "Torréfacteur", 45000, "Le café vert devient du café en grains.")
	unlock("moulin_cafe", "Moulin à café", 55000, "Le café en grains devient du café moulu.")
	unlock("riz", "Rizière", 65000, "Une grande rizière : le riz se vend en sachets.")
	unlock("vanille", "Vanille", 80000, "Des gousses de vanille, très chères.")
	unlock("the", "Théiers", 95000, "Les feuilles de thé partent au séchoir.")
	unlock("sechoir", "Séchoir à thé", 105000, "Les feuilles séchées deviennent des boîtes de thé.")
	unlock("cannelle", "Cannelliers", 120000, "Des bâtons de cannelle (la future pâtisserie en voudra).")
	for _, s := range p.Stations {
		if s.Unlock == "" {
			s.Built = true
		}
	}

	// --- améliorations
	k := p.key
	p.ups = []*Upgrade{
		{Key: k("caissiere"), Name: "Caissière", Max: 1, Costs: []float64{2500}, Desc: func(int) string { return "Elle encaisse les clients à ta place." }},
		{Key: k("etals"), Name: "Étals plus grands", Max: 6, Base: 3000, Mult: 2.3, Desc: func(l int) string { return "Chaque étal contient " + itoa(12+4*l) + " produits." }},
	}
	p.ups = append(p.ups, p.teamUpgrades("Ouvrier", "Il ne travaille que dans son espace : champs, bâtiments et étals.")...)
	p.ups = append(p.ups, p.droneUpgrades(2)...)
	p.Spots = append(p.Spots,
		&Spot{ID: "caissiere", Label: "Caissière", Zone: Rect{40, 960, 145, 130}, Done: func(int) string { return "Caissière embauchée !" }},
		&Spot{ID: "etals", Label: "Étals +", Zone: Rect{40, 1105, 130, 120}, Done: func(l int) string { return "Étals agrandis : niveau " + itoa(l) + " !" }})
	for _, st := range p.Stations {
		if st.Kind == Shelf {
			it := st.Item
			p.Spots = append(p.Spots, &Spot{ID: droneKey(it), Label: "Drone", Shelf: st, Zone: st.DronePad(),
				Done: func(int) string { return "Drone acheté pour l'étal « " + Items[it].Name + " » !" }})
		}
	}
	// les ouvriers s'embauchent sous le premier champ de leur espace (comme à la ferme)
	first := map[string]string{"fruits": "ananas", "sucre": "canne", "cacao": "cacao", "cafe": "cafe", "riz": "riz", "epices": "vanille"}
	for _, z := range exoticZones {
		z := z
		b := p.Station(first[z.ID]).Body
		zr := Rect{b.X + b.W/2 - 56, b.Y + b.H + 14, 112, 92}
		if z.ID == "fruits" {
			// une grande case au milieu, sous les quatre arbres, dans l'herbe entre les
			// deux chemins (comme la serre et le verger) : l'ouvrier est pour tous les fruits
			x0, x1 := p.Station(exoColumnIDs[0]).Body.X, p.Station(exoColumnIDs[len(exoColumnIDs)-1]).Body
			cx := (x0 + x1.X + x1.W) / 2
			zr = Rect{cx - 150, 684, 300, 92}
		}
		p.Spots = append(p.Spots, &Spot{ID: teamKey(z.ID), Label: "Ouvrier " + z.Name, Team: z.ID, Zone: zr,
			Done: func(l int) string { return "Ouvrier (" + z.Name + ") +1 !" }})
	}
	p.setupYard(-620)
	p.setupParking()
	p.spawnT = 0.5
	return p
}

// exoticOrder : un particulier commande deux ou trois produits de la ferme exotique.
func (g *Game) exoticOrder() *Order { return g.shopOrder("exotique") }
