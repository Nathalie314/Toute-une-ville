package sim

import "math/rand"

// Le domaine (11e lieu de la carte, après la boucherie) : un lieu « hybride »,
// à la fois vignoble, brasserie et distillerie (plan de la conceptrice).
//
//   au nord   : le houblon (à gauche) et la vigne (à droite)
//   dessous   : le stockage du raisin (il ne part qu'au supermarché, en camion)
//   un chemin, puis les fabriques : bière, vin blanc, vin rouge (une presse et un
//   foudre : le vin rouge fermente plus longtemps)
//   à droite  : la distillerie en colonne (whisky, vodka, alcool de poire, cidre,
//   jus de pomme), chaque fabrique en face de sa palette ; la réserve contre la
//   route des camions, verticale, au bord droit (particularité du lieu)
//   au centre : la caisse et les étals (vins, bière, cidre, alcools) ; les jus ne
//   se vendent pas sur place : leurs réserves (sans clients) sont à côté du quai du
//   supermarché, avec leur drone ; le houblon est plus bas que la vigne
//
// Rien n'est offert : même le premier pied de vigne s'achète. Ordre : vigne,
// jus de raisin, vin blanc, vin rouge, houblon, bière, pommes, jus de pomme,
// cidre, poires, alcool de poire, pommes de terre, vodka, blé, whisky.
// Pommes, poires, pommes de terre et blé arrivent de ma ferme : on achète leur
// palette dans la réserve (il faut que la ferme les produise).

const domainePrice = 32000000 // (remplacé par la courbe des prix : PlacePrice(11))

// Géométrie du domaine.
const (
	DomW        = 4900.0
	DomTop      = -430.0 // on monte jusqu'aux champs
	DomFieldY   = -170.0 // haut des champs (300 de haut)
	DomStoreY   = 250.0  // stockage du raisin, jus de raisin
	DomHopY     = 190.0  // haut du champ de houblon (plus bas que la vigne)
	DomJusX     = 3660.0 // les réserves de jus, à gauche du quai du supermarché
	DomJusY     = -300.0
	DomPathY    = 520.0  // le chemin (60 de haut)
	DomRowY     = 640.0  // les fabriques (vins, bière)
	DomColX     = 3650.0 // la distillerie (colonne)
	DomColY0    = 160.0  // haut de la 1re fabrique de la colonne
	DomColStep  = 225.0  // d'une fabrique à l'autre
	DomPalX     = 4210.0 // les palettes de la réserve
	DomRoadX    = 4520.0 // la route des camions (verticale)
	DomRoadW    = 160.0
	DomDockY0   = -330.0 // les quais (supermarché, crêperie, zoo), au nord de la réserve
	DomDockStep = 150.0
)

var domaineZones = []*ShopZone{
	{ID: "vigne", Name: "Vigne", X0: 1250, X1: 2900, RestY: 980, Costs: []float64{6000, 18000, 45000, 110000}},
	{ID: "houblon", Name: "Houblon", X0: 300, X1: 1200, RestY: 980, Costs: []float64{30000, 80000}},
	{ID: "distillerie", Name: "Distillerie", X0: 3450, X1: 4150, RestY: 980, Costs: []float64{50000, 120000, 260000, 520000, 1000000}},
}

func newDomaine(g *Game) *Place {
	W := DomW
	p := &Place{ID: "domaine", Name: "Le domaine", W: W, H: 1800, Top: DomTop, g: g, rng: rand.New(rand.NewSource(71)),
		Rep: 3, Price: domainePrice, SellMult: 1, SpawnBase: 6, WantMax: 3, MaxCust: 14, LeaveOff: V{170, 60}, RepGain: 0.01, OnlyStocked: true}
	p.SpawnL = V{350, roadY + 40}
	p.SpawnR = V{DomRoadX - 300, roadY + 40}
	p.Gate = V{1250, roadY + 40}
	p.Inside = V{1250, 1390}
	p.RestSpot = V{3000, 1440}

	field := func(id, name, team string, it Item, x, y, t float64, cols int) {
		s := &Station{ID: id, Name: name, Kind: Field, Item: it, Time: t, Unlock: id, Team: team}
		s.Body = Rect{x, y, 40 + 80*float64(cols), fieldH}
		s.Zone = s.Body
		for k, r := range []int{2, 1, 0} {
			for i := 0; i < cols; i++ {
				c := i
				if k%2 == 1 {
					c = cols - 1 - i
				}
				s.Slots = append(s.Slots, V{x + 60 + float64(c)*80, y + 60 + float64(r)*90})
			}
		}
		s.Plants = []Plant{{P: s.Slots[0], G: 1}}
		p.Stations = append(p.Stations, s)
	}
	// une fabrique de la rangée (zones dessous)
	machine := func(id, name, team string, out Item, in []Item, x, y, w, t float64) *Station {
		s := &Station{ID: id, Name: name, Kind: Machine, Item: out, Inputs: in, Time: t, PileMax: 12, BufMax: 10, Unlock: id, Team: team}
		s.Body = Rect{x, y, w, 150}
		s.Zone = Rect{x - 10, y + 168, 130, 90}
		s.OutZone = Rect{x + w - 120, y + 168, 130, 90}
		p.Stations = append(p.Stations, s)
		return s
	}
	// une fabrique de la colonne (distillerie) : on dépose à gauche (en haut),
	// on reprend à gauche (en bas) ; sa case « Stock + » est à droite
	colMachine := func(id, name string, out Item, in []Item, k int, t float64) *Station {
		y := DomColY0 + float64(k)*DomColStep
		s := &Station{ID: id, Name: name, Kind: Machine, Item: out, Inputs: in, Time: t, PileMax: 12, BufMax: 10, Unlock: id, Team: "distillerie"}
		s.Body = Rect{DomColX, y, 260, 150}
		s.Zone = Rect{DomColX - 160, y - 10, 140, 80}
		s.OutZone = Rect{DomColX - 160, y + 80, 140, 80}
		p.Stations = append(p.Stations, s)
		return s
	}
	// réserve sans clients (raisin, jus) : un drone emporte vers les quais
	store := func(id, name, team string, it Item, x, y float64, zoneAbove bool, unl string) *Station {
		s := &Station{ID: id, Name: name, Kind: Shelf, Item: it, StockMax: 30, Unlock: unl, NoSale: true, Team: team}
		s.Body = Rect{x, y, 220, shelfH}
		if zoneAbove {
			s.Zone = Rect{x - 10, y - 95, 240, 92}
		} else {
			s.Zone = Rect{x - 10, y + shelfH + 6, 240, 92}
		}
		s.Spot = V{x + 110, y + shelfH + 38}
		p.Stations = append(p.Stations, s)
		return s
	}

	// --- au nord : le houblon et la vigne
	field("houblon", "Houblon", "houblon", Houblon, 380, DomHopY, 8, 8) // (plus bas : juste au-dessus du chemin et de la brasserie)
	field("vigne", "Vigne", "vigne", Raisin, 1330, DomFieldY, 9, 12)
	// le stockage du raisin (sous la vigne : on y dépose depuis le haut) et la machine à jus
	rs := store("stock_raisin", "Stockage du raisin", "vigne", Raisin, 1360, DomStoreY+40, true, "vigne")
	rs.PadOff = 40
	machine("jus_raisin", "Machine à jus de raisin", "vigne", JusRaisin, []Item{Raisin}, 1900, DomStoreY-60, 260, 3)
	// --- les fabriques, après le chemin
	machine("vin_blanc", "Presse à vin blanc", "vigne", VinBlanc, []Item{Raisin}, 1250, DomRowY, 380, 6)
	machine("vin_rouge", "Presse à vin rouge", "vigne", VinRouge, []Item{Raisin}, 1880, DomRowY, 380, 10) // (fermentation plus longue)
	machine("biere", "Brasserie", "houblon", Biere, []Item{Houblon}, 520, DomRowY, 300, 5)
	// --- la distillerie, en colonne, chaque fabrique en face de sa palette
	colMachine("whisky", "Distillerie de whisky", Whisky, []Item{Ble}, 0, 9)
	colMachine("vodka", "Distillerie de vodka", Vodka, []Item{PommeDeTerre}, 1, 7)
	colMachine("alcool_poire", "Alambic : alcool de poire", AlcoolPoire, []Item{Poire}, 2, 7)
	colMachine("cidre", "Cidrerie", Cidre, []Item{JusPomme}, 3, 6)
	colMachine("jus_pomme", "Pressoir à pommes", JusPomme, []Item{Pomme}, 4, 3)
	// --- la réserve (palettes, contre la route des camions), chacune s'achète
	for _, pl := range []struct {
		it  Item
		k   float64
		unl string
	}{{Ble, 0, "pal_ble"}, {PommeDeTerre, 1, "pal_pdt"}, {Poire, 2, "pal_poire"}, {Pomme, 3.5, "pal_pomme"}} {
		y := DomColY0 + 30 + pl.k*DomColStep
		s := &Station{ID: "palette_" + itoa(int(pl.it)), Name: "Palette : " + Items[pl.it].Name, Kind: Pallet, Item: pl.it, PileMax: 40, Unlock: pl.unl}
		s.Body = Rect{DomPalX + 120, y, 160, 90}
		s.Zone = Rect{DomPalX, y, 110, 90}
		p.Stations = append(p.Stations, s)
	}
	// --- la boutique : la caisse, les étals ; puis les réserves de jus (sans clients)
	shelfTeam := map[Item]string{VinBlanc: "vigne", VinRouge: "vigne", Biere: "houblon", Cidre: "distillerie", AlcoolPoire: "distillerie", Vodka: "distillerie", Whisky: "distillerie"}
	shelfUnl := map[Item]string{VinBlanc: "vin_blanc", VinRouge: "vin_rouge", Biere: "biere", Cidre: "cidre", AlcoolPoire: "alcool_poire", Vodka: "vodka", Whisky: "whisky"}
	for i, it := range DomaineSold {
		x := 450 + float64(i)*370 // (étals écartés sur toute la boutique : un lieu aéré)
		s := &Station{ID: "etal_" + itoa(int(it)), Name: "Étal : " + Items[it].Name, Kind: Shelf, Item: it, StockMax: 12, Unlock: shelfUnl[it], Team: shelfTeam[it]}
		s.Body = Rect{x, shelfY, 170, shelfH}
		s.Zone = Rect{x - 10, shelfY - 95, 190, 92}
		s.Spot = V{x + 85, shelfY + shelfH + 38}
		p.Stations = append(p.Stations, s)
	}
	// les jus ne se vendent pas ici : leurs réserves sont à côté du quai du supermarché
	// (on y dépose par le haut ; le drone, dessous, n'a que quelques pas jusqu'au quai)
	for i, j := range []struct {
		id, name, team, unl string
		it                  Item
	}{{"stock_jus_raisin", "Réserve : jus de raisin", "vigne", "jus_raisin", JusRaisin}, {"stock_jus_pomme", "Réserve : jus de pomme", "distillerie", "jus_pomme", JusPomme}} {
		s := store(j.id, j.name, j.team, j.it, DomJusX+float64(i)*260, DomJusY, true, j.unl)
		s.PadOff = 40
	}
	reg := &Station{ID: "caisse", Name: "Caisse", Kind: Register, HireKey: "caissiere", QueueDown: 5}
	reg.Body = Rect{180, shelfY, 210, shelfH}
	reg.Spot = V{285, shelfY - 40}
	reg.Zone = Rect{200, shelfY - 100, 170, 96}
	reg.QueueHead = V{285, shelfY + shelfH + 40}
	p.Stations = append(p.Stations, reg)
	trash := &Station{ID: "poubelle", Name: "Compost", Kind: Trash}
	trash.Body = Rect{3080, shelfY - 10, 84, 76}
	trash.Zone = Rect{3060, shelfY - 100, 124, 180}
	p.Stations = append(p.Stations, trash)
	// --- les quais, au nord de la réserve, contre la route verticale (la crêperie
	// et le zoo n'existent pas encore : leurs quais restent invisibles)
	for i, to := range []string{"super", "creperie", "zoo"} {
		id := "quai"
		if i > 0 {
			id = "quai_" + to
		}
		y := DomDockY0 + float64(i)*DomDockStep
		q := &Station{ID: id, Name: "Quai du camion", Kind: Dock, BufMax: 20, Built: true, To: to}
		q.Body = Rect{DomRoadX - 230, y, 210, 56}
		q.Zone = Rect{DomPalX - 10, y - 30, 270, 112}
		p.Stations = append(p.Stations, q)
	}
	cmd := &Station{ID: "commande", Name: "Commandes", Kind: OrderBox, Built: true}
	p.Stations = append(p.Stations, cmd)

	// --- déblocages, dans l'ordre (rien n'est offert)
	unlock := func(id, name string, cost float64, hint, needs, locked string) {
		var at Rect
		for _, s := range p.Stations {
			if s.Unlock == id && (s.Kind == Field || s.Kind == Machine || s.Kind == Pallet) {
				c := s.Body.C()
				if s.Kind == Field {
					c = V{s.Body.X + 120, s.Body.Y + s.Body.H - 70} // (près du chemin, sur la première pousse)
				}
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
	unlock("vigne", "Vigne", 5000, "Le raisin ne se vend pas ici : un drone le porte au camion du supermarché (rayon Fruits).", "", "")
	unlock("jus_raisin", "Machine à jus de raisin", 9000, "Le jus part en camion (supermarché, et plus tard le zoo).", "", "")
	unlock("vin_blanc", "Presse à vin blanc", 15000, "Le raisin pressé fermente dans le foudre : du vin blanc.", "", "")
	unlock("vin_rouge", "Presse à vin rouge", 22000, "Plus long : le vin rouge fermente avec la peau du raisin.", "", "")
	unlock("houblon", "Houblon", 30000, "Le houblon ne se vend pas : il part à la brasserie.", "", "")
	unlock("biere", "Brasserie", 38000, "Le houblon devient de la bière.", "", "")
	unlock("pal_pomme", "Palette de pommes", 45000, "Le camion de ma ferme apporte des pommes.", "ferme:pommier", "Il faut le pommier (ma ferme)")
	unlock("jus_pomme", "Pressoir à pommes", 52000, "Les pommes deviennent du jus (supermarché, et plus tard le zoo).", "", "")
	unlock("cidre", "Cidrerie", 60000, "Le jus de pomme fermente : du cidre.", "", "")
	unlock("pal_poire", "Palette de poires", 70000, "Le camion de ma ferme apporte des poires.", "ferme:poirier", "Il faut le poirier (ma ferme)")
	unlock("alcool_poire", "Alambic : alcool de poire", 80000, "Les poires deviennent de l'alcool de poire.", "", "")
	unlock("pal_pdt", "Palette de pommes de terre", 90000, "Le camion de ma ferme apporte des pommes de terre.", "ferme:pdt", "Il faut les pommes de terre (ma ferme)")
	unlock("vodka", "Distillerie de vodka", 100000, "Les pommes de terre deviennent de la vodka.", "", "")
	unlock("pal_ble", "Palette de blé", 110000, "Le camion de ma ferme apporte du blé (son silo à blé).", "ferme:silo", "Il faut le silo à blé (ma ferme)")
	p.Unlocks[len(p.Unlocks)-1].Secret = true // (rien d'écrit tant qu'on ne peut pas l'acheter)
	unlock("whisky", "Distillerie de whisky", 125000, "Le blé devient du whisky : le plus cher du domaine.", "", "")
	// chaque machine de la distillerie attend sa palette (et le cidre, le pressoir)
	for _, u := range p.Unlocks {
		need := map[string]string{"jus_pomme": "pal_pomme", "cidre": "jus_pomme", "alcool_poire": "pal_poire", "vodka": "pal_pdt", "whisky": "pal_ble"}[u.ID]
		if need != "" {
			u.Requires = func() bool { return p.builtOrUnlocked(need) }
		}
	}
	for _, s := range p.Stations {
		if s.Unlock == "" {
			s.Built = true
		}
	}

	// --- améliorations
	k := p.key
	p.ups = []*Upgrade{
		{Key: k("caissiere"), Name: "Caissière", Max: 1, Costs: []float64{4000}, Desc: func(int) string { return "Elle encaisse les clients à ta place." }},
		{Key: k("etals"), Name: "Étals plus grands", Max: 6, Base: 4000, Mult: 2.3, Desc: func(l int) string { return "Chaque étal contient " + itoa(12+4*l) + " bouteilles." }},
		{Key: k("entrepot"), Name: "Entrepôt", Max: 12, Base: 5000, Mult: 1.6, Desc: func(l int) string {
			return "Chaque palette de la réserve contient " + itoa(PalletMax(p.ID, l)) + " produits."
		}},
	}
	p.ups = append(p.ups, p.teamUpgrades("Ouvrier", "Il ne travaille que dans son espace.")...)
	p.ups = append(p.ups, p.droneUpgrades(3)...)
	p.Spots = append(p.Spots,
		&Spot{ID: "caissiere", Label: "Caissière", Zone: Rect{40, 960, 145, 130}, Done: func(int) string { return "Caissière embauchée !" }},
		&Spot{ID: "etals", Label: "Étals +", Zone: Rect{40, 1105, 130, 120}, Done: func(l int) string { return "Étals agrandis : niveau " + itoa(l) + " !" }},
		&Spot{ID: "entrepot", Label: "Entrepôt", Zone: Rect{DomPalX + 10, DomColY0 + 30 + 4.6*DomColStep, 130, 124},
			Done: func(l int) string { return "Entrepôt agrandi : " + itoa(PalletMax(p.ID, l)) + " par palette !" }})
	for _, st := range p.Stations {
		if st.Kind == Shelf {
			it := st.Item
			p.Spots = append(p.Spots, &Spot{ID: droneKey(it), Label: "Drone", Shelf: st, Zone: st.DronePad(),
				Done: func(int) string { return "Drone acheté pour « " + Items[it].Name + " » !" }})
		}
	}
	p.setupYard(-620)
	p.teamSpots("Ouvrier")
	p.setupParking()
	p.spawnT = 1
	return p
}

// domaineOrder : un particulier commande deux ou trois bouteilles du domaine.
func (g *Game) domaineOrder() *Order { return g.shopOrder("domaine") }
