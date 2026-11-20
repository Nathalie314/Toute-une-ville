package sim

import "math/rand"

// La ferme exotique (après la crèmerie). Comme ma ferme : des champs, des
// bâtiments, des étals, des clients, des commandes de particuliers, des drones
// et des camions (supermarché, boulangerie, glacier).
//
//   fruits  : ananas, bananier, cocotier, manguier, avocatier, citronnier, oranger (une colonne chacun, comme le verger)
//   sucre   : canne à sucre → raffinerie (sucre en morceaux) → moulin (sucre en poudre)
//   cacao   : cacaoyers → chocolaterie (cacao + sucre en poudre → tablettes)
//   café    : caféiers → torréfacteur (café en grains) → moulin à café (café moulu)
//   riz     : une grande rizière (le riz se vend en sachets)
//   épices  : vanille, cannelle, théiers → séchoir (boîtes de thé), poivriers → séchoir à poivre
//
// Chaque espace a ses ouvriers (comme à la ferme d'élevage).

const exoticPrice = 35000

var exoticZones = []*ShopZone{
	{ID: "fruits", Name: "Fruits", X0: 420, X1: 2060, RestY: 970, Costs: []float64{3000, 12000, 30000, 60000, 100000, 150000, 200000, 260000, 330000}},
	{ID: "sucre", Name: "Sucre", X0: 2140, X1: 2960, RestY: 970, Costs: []float64{5000, 16000, 40000}},
	{ID: "cacao", Name: "Cacao", X0: 3040, X1: 3460, RestY: 970, Costs: []float64{9000, 24000}},
	{ID: "cafe", Name: "Café", X0: 3540, X1: 4360, RestY: 970, Costs: []float64{14000, 32000, 70000}},
	{ID: "riz", Name: "Riz", X0: 4450, X1: 5330, RestY: 970, Costs: []float64{20000}},
	{ID: "epices", Name: "Épices", X0: 5430, X1: 6890, RestY: 970, Costs: []float64{28000, 55000, 100000, 170000, 280000, 420000}}, // (6 ressources avec les poivriers et leur séchoir)
}

// exoColumnIDs : les colonnes d'arbres exotiques.
var exoColumnIDs = []string{"ananas", "bananier", "cocotier", "manguier", "avocatier", "citronnier", "oranger"}

// ExoDecoTop : le haut du décor de la ferme exotique (forêt, cascade, rivière).
const ExoDecoTop = -950.0

func newExotic(g *Game) *Place {
	W := 7240.0                                                                      // (aéré : on voit toutes les cases et on passe entre les champs)
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

	// --- fruits : sept colonnes d'arbres (l'ananas, lui, pousse au sol…)
	names := []string{"Ananas", "Bananier", "Cocotier", "Manguier", "Avocatier", "Citronnier", "Oranger"}
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
	// les pressoirs, sous les arbres : citrons → limonade, oranges → jus d'orange (vendus au
	// rayon « Boissons » du supermarché) ; leurs étals à droite des fruits
	machine("pressoir_citron", "Pressoir à citrons", "fruits", Limonade, []Item{Citron}, 1160, 3, "pressoir_citron")
	machine("pressoir_orange", "Pressoir à oranges", "fruits", JusOrange, []Item{Orange}, 1570, 3, "pressoir_orange")
	shelf("fruits", Limonade, 1630, 220, "pressoir_citron", false)
	shelf("fruits", JusOrange, 1880, 220, "pressoir_orange", false)
	if a := p.Station("ananas"); a != nil {
		a.Unlock = "" // la première colonne est offerte avec la ferme
		p.Station("etal_" + itoa(int(Ananas))).Unlock = ""
	}
	// --- sucre
	field("canne", "Canne à sucre", "sucre", Canne, 2160, 6, 6, "canne")
	machine("raffinerie", "Raffinerie", "sucre", SucreMorceaux, []Item{Canne}, 2160, 3, "raffinerie")
	ms := machine("moulin_sucre", "Moulin à sucre", "sucre", SucrePoudre, []Item{SucreMorceaux}, 2560, 2.5, "moulin_sucre")
	ms.Shared = true // la chocolaterie (espace cacao) s'y sert aussi
	shelf("sucre", SucreMorceaux, 2170, 220, "raffinerie", false)
	shelf("sucre", SucrePoudre, 2570, 220, "moulin_sucre", false)
	// --- cacao
	field("cacao", "Cacaoyers", "cacao", Cacao, 3060, 8, 4, "cacao")
	machine("chocolaterie", "Chocolaterie", "cacao", Chocolat, []Item{Cacao, SucrePoudre}, 3060, 4, "chocolaterie")
	shelf("cacao", Chocolat, 3070, 220, "chocolaterie", false)
	// --- café
	field("cafe", "Caféiers", "cafe", CafeVert, 3560, 8, 6, "cafe")
	machine("torrefacteur", "Torréfacteur", "cafe", CafeGrains, []Item{CafeVert}, 3560, 3.5, "torrefacteur")
	machine("moulin_cafe", "Moulin à café", "cafe", CafeMoulu, []Item{CafeGrains}, 3960, 3, "moulin_cafe")
	shelf("cafe", CafeGrains, 3570, 220, "torrefacteur", false)
	shelf("cafe", CafeMoulu, 3970, 220, "moulin_cafe", false)
	// --- la rizière
	field("riz", "Rizière", "riz", Riz, 4470, 6, 10, "riz")
	shelf("riz", Riz, 4710, 360, "riz", false)
	// --- épices : vanille, cannelle, thé (le thé en dernier : son séchoir dessous, à
	// droite ; l'embauche des épices au milieu, sous les cannelliers)
	field("vanille", "Vanille", "epices", Vanille, 5450, 12, 3, "vanille")
	field("cannelle", "Cannelliers", "epices", Cannelle, 5830, 11, 3, "cannelle")
	field("the", "Théiers", "epices", FeuillesThe, 6210, 7, 3, "the")
	field("poivre", "Poivriers", "epices", BaiesPoivre, 6590, 10, 3, "poivre") // (lianes sur tuteurs, grappes vertes puis rouges)
	machine("sechoir", "Séchoir à thé", "epices", The, []Item{FeuillesThe}, 6220, 3.5, "sechoir")
	machine("sechoir_poivre", "Séchoir à poivre", "epices", Poivre, []Item{BaiesPoivre}, 6600, 3.5, "sechoir_poivre") // (les baies séchées deviennent du poivre)
	shelf("epices", Vanille, 5460, 220, "vanille", false)
	shelf("epices", Cannelle, 5840, 220, "cannelle", false)
	shelf("epices", The, 6220, 220, "sechoir", false)
	shelf("epices", Poivre, 6600, 220, "sechoir_poivre", false)

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
	// les quais : au bord de la rivière, au nord (la ferme exotique est loin : tout part en bateau)
	p.riverDock("quai", "super", 2010)
	p.riverDock("quai_boul", "boulangerie", 2360)
	p.riverDock("quai_glace", "glacier", 2710)
	p.riverDock("quai_patisserie", "patisserie", 3060)
	p.riverDock("quai_creperie", "creperie", 3410)
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
	unlock("avocatier", "Avocatier", 34000, fruitHint)
	unlock("citronnier", "Citronnier", 36000, fruitHint) // (prix relevés avec l'avocatier : migrateAvocatier)
	unlock("oranger", "Oranger", 38000, fruitHint)
	unlock("pressoir_citron", "Pressoir à citrons", 40000, "Les citrons deviennent de la limonade (le supermarché la vend au rayon « Boissons »).")
	unlock("pressoir_orange", "Pressoir à oranges", 43000, "Les oranges deviennent du jus d'orange (le supermarché le vend au rayon « Boissons »).")
	unlock("cafe", "Caféiers", 38000, "Le café vert ne se vend pas : il part au torréfacteur.")
	unlock("torrefacteur", "Torréfacteur", 45000, "Le café vert devient du café en grains.")
	unlock("moulin_cafe", "Moulin à café", 55000, "Le café en grains devient du café moulu.")
	unlock("riz", "Rizière", 65000, "Une grande rizière : le riz se vend en sachets.")
	// (épices dans l'ordre où elles sont plantées : vanille, cannelle, thé et son séchoir ;
	// les anciens prix d'une sauvegarde sont convertis : migrateEpices)
	unlock("vanille", "Vanille", 80000, "Des gousses de vanille, très chères.")
	unlock("cannelle", "Cannelliers", 95000, "Des bâtons de cannelle (la future pâtisserie en voudra).")
	unlock("the", "Théiers", 105000, "Les feuilles de thé partent au séchoir.")
	unlock("sechoir", "Séchoir à thé", 120000, "Les feuilles séchées deviennent des boîtes de thé.")
	unlock("poivre", "Poivriers", 140000, "Les baies de poivre partent au séchoir.")
	unlock("sechoir_poivre", "Séchoir à poivre", 160000, "Les baies séchées deviennent du poivre (le supermarché le vend à l'Épicerie).")
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
		if z.ID == "epices" {
			// une grande case sous la vanille et la cannelle (centrée sur elles), dans l'herbe : l'ouvrier
			// est pour toutes les épices (vanille, cannelle, thé et son séchoir, poivre et son séchoir)
			v, c := p.Station("vanille").Body, p.Station("cannelle").Body
			zr = Rect{(v.X+c.X+c.W)/2 - 150, 684, 300, 92}
		}
		if z.ID == "fruits" {
			// une grande case sous les trois premiers arbres, dans l'herbe entre les deux
			// chemins (les pressoirs sont sous les suivants) : l'ouvrier est pour tous les fruits
			x0, x1 := p.Station(exoColumnIDs[0]).Body.X, p.Station(exoColumnIDs[2]).Body
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
