package sim

import (
	"math"
	"math/rand"
	"strings"
)

// Ateliers : la boulangerie, la crèmerie et le glacier.
// Les ingrédients arrivent de la ferme en camion dans la réserve (palettes),
// on les porte aux machines, puis on met les produits en vitrine.
// Un quai envoie aussi les produits au supermarché.

type recipe struct {
	id, name string
	out      Item
	in       []Item
	time     float64
	cost     float64 // 0 : déjà là
	hint     string
	needs    string // lieu qu'il faut avoir acheté (ex. la crèmerie, pour le beurre)
}

type workshopSpec struct {
	id, name   string
	price      float64
	inputs     []Item
	recipes    []recipe
	staff      string // nom des employés
	cashCost   float64
	staffCosts []float64
	upBase     float64
	sign, bio  string
	seed       int64
	w          float64  // largeur minimale (2700 par défaut ; agrandie s'il faut pour les machines)
	perRow     int      // machines par rangée (4 par défaut)
	docks      []string // lieux où partent ses camions (le supermarché par défaut)
	shift2     bool     // 2e rangée décalée d'un cran à droite (rien derrière la caissière)
	gap        float64  // allée entre deux machines (machineGap par défaut)
	dockNear   bool     // quais juste après la dernière vitrine
	tall       bool     // file de la caisse plus courte (elle ne doit pas sortir par la vitrine)
}

var bakerySpec = workshopSpec{
	id: "boulangerie", name: "La boulangerie", price: 50000, seed: 31,
	inputs: []Item{Farine, FarineMais, Oeuf, Lait, Beurre, SucrePoudre, Chocolat},
	recipes: []recipe{
		{"four_baguette", "Tradition", Pain, []Item{Farine}, 3, 0, "", ""},
		{"pain_mie", "Pain de mie", PainMie, []Item{Farine}, 3.5, 1500, "Farine de blé → pain de mie.", ""},
		{"four_mais", "Baguette maïs", BaguetteMais, []Item{FarineMais}, 3, 2500, "Farine de maïs → baguette de maïs.", ""},
		{"four_mele", "Pain blé-maïs", PainMele, []Item{Farine, FarineMais}, 4, 5000, "Les deux farines ensemble.", ""},
		{"brioches", "Brioches", Brioche, []Item{Farine, Oeuf, Beurre, SucrePoudre}, 5, 9000, "Farine, œufs, beurre et sucre.", "cremerie:beurre,exotique:moulin_sucre"},
		{"croissants", "Croissants", Croissant, []Item{Farine, Oeuf, Lait, Beurre, SucrePoudre}, 5, 30000, "Farine, œufs, lait, beurre et sucre.", "cremerie:beurre,exotique:moulin_sucre"},
		{"pains_choco", "Pains au chocolat / chocolatines", PainChoco, []Item{Farine, Oeuf, Beurre, Chocolat}, 6, 60000, "Farine, œufs, beurre et chocolat.", "cremerie:beurre,exotique:chocolaterie"},
	},
	staff: "Boulangers", cashCost: 3000, staffCosts: []float64{4000, 10000, 25000, 60000, 150000, 350000, 800000}, upBase: 5000, // (un boulanger par machine)
	sign: "Vitrine", bio: "Fait maison", shift2: true, tall: true, dockNear: true, // (quai juste à droite des pains au chocolat)
}

// cremerieSpec : la crèmerie, avec les trois laits de la ferme. Le beurre se
// fait avec la crème de la crèmerie (les employés la portent au baratteur).
var cremerieSpec = workshopSpec{
	id: "cremerie", name: "La crèmerie", price: 22000, seed: 41,
	inputs: []Item{Lait, LaitChevre, LaitBrebis},
	recipes: []recipe{
		{"creme", "Écrémeuse", Creme, []Item{Lait}, 3, 0, "", ""},
		{"beurre", "Baratte à beurre", Beurre, []Item{Creme}, 4, 6000, "La crème devient du beurre.", ""},
		{"fromage_blanc", "Fromage blanc", FromageBlanc, []Item{Lait}, 3, 9000, "Lait de vache.", ""},
		{"yaourt", "Yaourts de vache", Yaourt, []Item{Lait}, 3, 12000, "Lait de vache.", ""},
		{"fromage", "Fromage de vache", Fromage, []Item{Lait}, 5, 18000, "Lait de vache.", ""},
		{"yaourt_chevre", "Yaourts de chèvre", YaourtChevre, []Item{LaitChevre}, 4, 25000, "Lait de chèvre.", ""},
		{"fromage_chevre", "Fromage de chèvre", FromageChevre, []Item{LaitChevre}, 5, 35000, "Lait de chèvre.", ""},
		{"yaourt_brebis", "Yaourts de brebis", YaourtBrebis, []Item{LaitBrebis}, 4, 45000, "Lait de brebis.", ""},
		{"fromage_brebis", "Fromage de brebis", FromageBrebis, []Item{LaitBrebis}, 6, 60000, "Lait de brebis.", ""},
	},
	staff: "Crémiers", cashCost: 4500, staffCosts: []float64{6000, 15000, 38000, 90000, 220000, 500000, 1100000, 2400000, 5000000}, upBase: 7000,
	sign: "Vitrine", bio: "Fermier",
	w: 4000, perRow: 5, docks: []string{"super", "boulangerie", "glacier", "patisserie", "creperie"}, shift2: true, tall: true, dockNear: true,
}

// patisserieSpec : la pâtisserie (12e lieu). Sa particularité : chaque recette
// demande au moins 3 ingrédients, venus de ma ferme (pommes, farine, cerises,
// œufs, lait, fraises), de la crèmerie (beurre) et de la ferme exotique (riz,
// vanille, chocolat, café moulu). Rien n'est offert : même la tarte s'achète.
var patisserieSpec = workshopSpec{
	id: "patisserie", name: "La pâtisserie", price: 64000000, seed: 83,
	inputs: []Item{Pomme, Beurre, Farine, Riz, Lait, Vanille, Cerise, Oeuf, Chocolat, CafeMoulu, Fraise},
	recipes: []recipe{
		{"tarte_pommes", "Tartes aux pommes", TartePommes, []Item{Pomme, Beurre, Farine}, 5, 5000, "Pommes, beurre et farine.", "cremerie:beurre,ferme:pommier"},
		{"riz_lait", "Riz au lait", RizLait, []Item{Riz, Lait, Vanille}, 5, 9000, "Riz, lait et vanille.", "exotique:riz,exotique:vanille"},
		{"clafoutis", "Clafoutis aux cerises", Clafoutis, []Item{Cerise, Farine, Oeuf, Lait}, 6, 15000, "Cerises, farine, œufs et lait.", "ferme:cerisier"},
		{"gateau", "Gâteaux au chocolat", Gateau, []Item{Chocolat, Beurre, Farine, Oeuf}, 6, 25000, "Chocolat, beurre, farine et œufs.", "exotique:chocolaterie,cremerie:beurre"},
		{"eclair", "Éclairs au café", Eclair, []Item{Farine, Oeuf, Lait, Beurre, CafeMoulu}, 7, 40000, "Farine, œufs, lait, beurre et café moulu.", "exotique:moulin_cafe,cremerie:beurre"},
		{"fraisier", "Fraisiers", Fraisier, []Item{Farine, Oeuf, Lait, Beurre, Fraise}, 7, 60000, "Farine, œufs, lait, beurre et fraises.", "ferme:fraisier,cremerie:beurre"},
	},
	staff: "Pâtissiers", cashCost: 9000, staffCosts: []float64{10000, 25000, 60000, 150000, 350000, 800000}, upBase: 12000, // (un par machine)
	sign: "Vitrine", bio: "Fait maison", shift2: true, tall: true, dockNear: true, gap: 260,
}

// iceSpec : le glacier. Les sorbets : seulement le fruit (et de l'eau) ; les
// glaces : crème, œufs et le parfum (vanille, café, noix de coco, chocolat, poire et chocolat).
var iceSpec = workshopSpec{
	id: "glacier", name: "Le glacier", price: 80000, seed: 37, w: 3300,
	inputs: []Item{Fraise, Mangue, Ananas, Vanille, CafeMoulu, NoixCoco, Creme, Oeuf, Chocolat, Poire},
	recipes: []recipe{
		{"sorbet_fraise", "Sorbet à la fraise", SorbetFraise, []Item{Fraise}, 3, 0, "", ""},
		{"sorbet_mangue", "Sorbet à la mangue", SorbetMangue, []Item{Mangue}, 4, 15000, "Des mangues (ferme exotique).", "exotique:manguier"},
		{"sorbet_ananas", "Sorbet à l'ananas", SorbetAnanas, []Item{Ananas}, 4, 25000, "Des ananas (ferme exotique).", "exotique:ananas"},
		{"glace_vanille", "Glace à la vanille", GlaceVanille, []Item{Creme, Oeuf, Vanille}, 5, 35000, "Crème, œufs et vanille.", "cremerie:creme,exotique:vanille"},
		{"glace_cafe", "Glace au café", GlaceCafe, []Item{Creme, Oeuf, CafeMoulu}, 5, 42000, "Crème, œufs et café moulu.", "cremerie:creme,exotique:moulin_cafe"},
		{"glace_coco", "Glace à la noix de coco", GlaceCoco, []Item{Creme, Oeuf, NoixCoco}, 5, 50000, "Crème, œufs et noix de coco.", "cremerie:creme,exotique:cocotier"},
		{"glace_choco", "Glace au chocolat", GlaceChoco, []Item{Creme, Oeuf, Chocolat}, 6, 65000, "Crème, œufs et chocolat.", "cremerie:creme,exotique:chocolaterie"},
		{"glace_poire", "Glace poire-chocolat", GlacePoireChoco, []Item{Creme, Oeuf, Poire, Chocolat}, 7, 90000, "Crème, œufs, poires et chocolat.", "cremerie:creme,exotique:chocolaterie"},
	},
	staff: "Serveurs", cashCost: 6000, staffCosts: []float64{8000, 20000, 50000, 120000, 300000, 700000, 1500000, 3200000}, upBase: 9000, // (un par fabrique)
	sign: "Parasols", bio: "Recettes maison", shift2: true, tall: true,
	gap: 320, dockNear: true, docks: []string{"super", "creperie"}, // (fabriques larges : on circule bien entre elles ; la glace vanille part aussi à la crêperie)
}

// Disposition des ateliers (jeu aéré) : chaque machine est suivie de sa case
// « Stock + » (stockBoxW) puis d'une allée (machineGap) ; deux rangées de
// machines ; sous la 2e rangée, un couloir libre (on y passe sans marcher dans
// aucune zone) avant les vitrines.
const (
	stockBoxW    = 118.0 // case « Stock + » à droite d'une machine (14 + 104)
	machineGap   = 240.0 // allée entre une machine (avec sa case) et la suivante
	workshopRow0 = 450.0 // haut de la 1re rangée de machines
	workshopRowG = 350.0 // d'une rangée à l'autre
	workshopShY  = shelfY + 120.0
	workshopPad  = 85.0 // bases des drones plus près des vitrines (le mur du bas est à 1514)
)

// machineW : largeur d'une machine (beaucoup d'ingrédients : plus large, sa zone de dépôt aussi).
func machineW(r recipe) float64 { return 260 + 70*math.Max(0, float64(len(r.in)-2)) }

func newWorkshop(g *Game, sp workshopSpec) *Place {
	per := sp.perRow
	if per == 0 {
		per = 4
	}
	gap := machineGap
	if sp.gap > 0 {
		gap = sp.gap
	}
	// positions des machines, rangée par rangée
	mx := make([]float64, len(sp.recipes))
	end := 0.0
	x := 200.0
	for i, r := range sp.recipes {
		if i%per == 0 && i > 0 {
			x = 200
			if sp.shift2 && len(mx) > 1 {
				x = mx[1] // (jamais de machine derrière la caissière : la 2e rangée commence à la 2e colonne)
			}
		}
		mx[i] = x
		x += machineW(r) + stockBoxW
		end = math.Max(end, x)
		x += gap
	}
	// assez large pour les machines, une allée, puis le compost contre le mur de droite
	W := math.Max(math.Max(sp.w, 2700), end+gap+330)
	if sp.docks == nil {
		sp.docks = []string{"super"}
	}
	p := &Place{ID: sp.id, Name: sp.name, W: W, H: 1800, g: g, rng: rand.New(rand.NewSource(sp.seed)),
		Rep: 3, Price: sp.price, SellMult: 1, SpawnBase: 5, WantMax: 3, MaxCust: 18, LeaveOff: V{170, 60}, OnlyStocked: true, RepGain: 0.008}
	p.SpawnL = V{350, roadY + 40}
	p.SpawnR = V{W - 550, roadY + 40}
	p.Gate = V{1250, roadY + 40}
	p.Inside = V{1250, 1390}
	p.setupParking()
	p.RestSpot = V{W - 550, 1440}
	p.Start = V{1250, 1420} // (sous les vitrines)

	// réserve
	pal := math.Min(290, (W-300)/float64(len(sp.inputs))) // (resserrées s'il y en a beaucoup)
	for i, it := range sp.inputs {
		x := 200 + float64(i)*pal
		s := &Station{ID: "palette_" + itoa(int(it)), Name: "Palette : " + Items[it].Name, Kind: Pallet, Item: it, PileMax: 40, Built: true}
		s.Body = Rect{x, 170, 160, 90}
		s.Zone = Rect{x - 5, 280, 170, 90}
		p.Stations = append(p.Stations, s)
	}
	// machines (deux rangées) et vitrines
	row0, rowGap, sy := workshopRow0, workshopRowG, workshopShY
	for i, r := range sp.recipes {
		x := mx[i]
		y := row0 + float64(i/per)*rowGap
		unl := r.id
		if r.cost == 0 {
			unl = ""
		}
		s := &Station{ID: r.id, Name: r.name, Kind: Machine, Item: r.out, Inputs: r.in, Time: r.time, PileMax: 12, BufMax: 10, Unlock: unl, BaseCost: math.Max(r.cost, 3000)}
		// beaucoup d'ingrédients : une machine et une zone de dépôt plus larges (jeu aéré)
		mw := machineW(r)
		s.Body = Rect{x, y, mw, 150}
		s.Zone = Rect{x - 10, y + 150 + 18, mw - 130, 90}
		s.OutZone = Rect{x + mw - 120, y + 150 + 18, 130, 90}
		p.Stations = append(p.Stations, s)
		sx := 450 + float64(i)*190
		sh := &Station{ID: "etal_" + itoa(int(r.out)), Name: "Vitrine : " + Items[r.out].Name, Kind: Shelf, Item: r.out, StockMax: 12, Unlock: unl}
		sh.Body = Rect{sx, sy, 150, shelfH}
		sh.Zone = Rect{sx - 10, sy - 95, 170, 92}
		sh.Spot = V{sx + 75, sy + shelfH + 38}
		sh.PadOff = workshopPad
		p.Stations = append(p.Stations, sh)
	}
	qd := 5
	if sp.tall {
		qd = 4 // la file ne doit pas sortir par la vitrine
	}
	reg := &Station{ID: "caisse", Name: "Caisse", Kind: Register, HireKey: "caissiere", QueueDown: qd}
	reg.Body = Rect{180, sy, 210, shelfH}
	reg.Spot = V{285, sy - 40}
	reg.Zone = Rect{200, sy - 100, 170, 96}
	reg.QueueHead = V{285, sy + shelfH + 40}
	p.Stations = append(p.Stations, reg)
	// le compost contre le mur de droite, à hauteur de la 1re rangée : hors des
	// allées et loin des cases (le mur est à W - 96)
	trash := &Station{ID: "poubelle", Name: "Poubelle", Kind: Trash}
	trash.Body = Rect{W - 250, row0 + 40, 70, 70}
	trash.Zone = Rect{W - 270, row0, 110, 150}
	// les quais juste après la dernière vitrine (le premier vers le supermarché)
	dockX := func(i int) float64 { return W - 500 - float64(i)*350 }
	if sp.dockNear {
		last := 450 + float64(len(sp.recipes)-1)*190 + 150
		dockX = func(i int) float64 { return last + 60 + float64(i)*350 }
	}
	p.Stations = append(p.Stations, trash)
	// quais : un par destination (le premier vers le supermarché)
	for i, to := range sp.docks {
		id := "quai"
		if i > 0 {
			id = "quai_" + to
		}
		p.dock(id, to, dockX(i))
	}
	if sp.id == "boulangerie" || sp.id == "cremerie" || sp.id == "patisserie" || sp.id == "glacier" {
		// caisse des commandes (boulangerie, crèmerie, glacier : la cantine de l'école ; crèmerie, pâtisserie : des particuliers)
		cmd := &Station{ID: "commande", Name: "Commandes", Kind: OrderBox, Built: true}
		cmd.Body = Rect{W - 280, 1300, 180, 60}
		cmd.Zone = Rect{W - 290, 1372, 200, 96}
		p.Stations = append(p.Stations, cmd)
	}

	// déblocages
	for _, r := range sp.recipes {
		if r.cost == 0 {
			continue
		}
		s := p.Station(r.id)
		c := s.Body.C()
		u := &Unlock{ID: r.id, Name: r.name, Cost: r.cost, Zone: Rect{c.X - 90, c.Y - 65, 180, 130}, Hint: r.hint}
		if need := r.needs; need != "" {
			u.Requires = func() bool { return g.needsMet(need) }
			u.Locked = needsText(need)
		}
		p.Unlocks = append(p.Unlocks, u)
	}
	for _, s := range p.Stations {
		if s.Unlock == "" {
			s.Built = true
		}
	}

	// améliorations au sol
	k := p.key
	b := sp.upBase
	p.ups = []*Upgrade{
		{Key: k("caissiere"), Name: "Caissière", Max: 1, Costs: []float64{sp.cashCost}, Desc: func(int) string { return "Elle encaisse à ta place." }},
		{Key: k("ouvriers"), Name: sp.staff, Max: len(sp.staffCosts), Now: true, Costs: sp.staffCosts, // (autant que de machines)
			Desc: func(l int) string {
				return "Ils portent les ingrédients aux machines et garnissent les vitrines. Embauchés : " + itoa(l) + "."
			}},
		{Key: k("etals"), Name: "Vitrines plus grandes", Max: 6, Base: b * 0.6, Mult: 2.3, Desc: func(l int) string { return "Chaque vitrine contient " + itoa(12+4*l) + " produits." }},
		{Key: k("entrepot"), Name: "Entrepôt", Max: 12, Base: b * 0.4, Mult: 1.6, Desc: func(l int) string {
			return "Chaque palette de la réserve contient " + itoa(PalletMax(p.ID, l)) + " produits."
		}},
	}
	scale := map[string]float64{"boulangerie": 2, "cremerie": 2.5, "glacier": 3, "boucherie": 4, "patisserie": 4}[sp.id]
	p.ups = append(p.ups, p.droneUpgrades(scale)...)
	spot := func(id, label string, z Rect, needs string, done func(int) string) {
		p.Spots = append(p.Spots, &Spot{ID: id, Label: label, Zone: z, Needs: needs, Done: done})
	}
	// drones : un par vitrine (base sous la vitrine) ; soute et hélices près des employés
	for _, st := range p.Stations {
		if st.Kind == Shelf {
			it := st.Item
			p.Spots = append(p.Spots, &Spot{ID: droneKey(it), Label: "Drone", Shelf: st, Zone: st.DronePad(),
				Done: func(int) string { return "Drone acheté pour la vitrine « " + Items[it].Name + " » !" }})
		}
	}
	spot("caissiere", "Caissière", Rect{40, 960, 145, 130}, "", func(int) string { return "Caissière embauchée !" })
	spot("etals", "Vitrines +", Rect{40, 1105, 130, 120}, "", func(l int) string { return "Vitrines agrandies : niveau " + itoa(l) + " !" })
	// (l'entrepôt s'agrandit à gauche de la réserve, côté zone du joueur)
	spot("entrepot", "Entrepôt", Rect{45, 170, 130, 124}, "", func(l int) string { return "Entrepôt agrandi : " + itoa(PalletMax(p.ID, l)) + " par palette !" })
	spot("ouvriers", sp.staff, Rect{W - 710, 1290, 130, 124}, "", func(l int) string { return "Nouvel employé : +1 !" })
	p.setupYard(-620)
	p.spawnT = 1
	return p
}

// placeName : « la crèmerie », « le supermarché »…
func placeName(id string) string {
	return map[string]string{"ferme": "la ferme", "super": "le supermarché", "boulangerie": "la boulangerie",
		"cremerie": "la crèmerie", "glacier": "le glacier", "ecole": "l'école",
		"elevage": "la ferme d'élevage", "boucherie": "la boucherie-charcuterie", "exotique": "la ferme exotique", "domaine": "le domaine", "creperie": "la crêperie", "patisserie": "la pâtisserie", "parc": "le parc d'attractions"}[id]
}

// needsMet : « lieu » (acheté) ou « lieu:station » (station construite), séparés par des virgules.
func (g *Game) needsMet(needs string) bool {
	for _, n := range strings.Split(needs, ",") {
		id, st, _ := strings.Cut(n, ":")
		q := g.Place(id)
		if q == nil || !q.Owned {
			return false
		}
		if st != "" && !q.builtOrUnlocked(st) {
			return false
		}
	}
	return true
}

// needsText : ce qu'il manque, en clair.
func needsText(needs string) string {
	txt := map[string]string{"cremerie:beurre": "le beurre de la crèmerie", "cremerie:creme": "la crème (crèmerie)",
		"exotique:chocolaterie": "la chocolaterie (ferme exotique)", "exotique:raffinerie": "la raffinerie (ferme exotique)",
		"exotique:moulin_sucre": "le sucre en poudre (ferme exotique)",
		"exotique:manguier":     "les mangues (ferme exotique)", "exotique:ananas": "les ananas (ferme exotique)",
		"exotique:cocotier": "les noix de coco (ferme exotique)",
		"ferme:cerisier":    "le cerisier (ma ferme)", "ferme:fraisier": "les fraisiers (ma ferme)", "exotique:riz": "le riz (ferme exotique)",
		"ferme:pommier": "le pommier (ma ferme)", "ferme:poirier": "le poirier (ma ferme)", "ferme:pdt": "les pommes de terre (ma ferme)", "ferme:silo": "le silo à blé (ma ferme)",
		"exotique:vanille": "la vanille (ferme exotique)", "exotique:moulin_cafe": "le café moulu (ferme exotique)",
		"elevage:mare": "les canards (ferme d'élevage)", "elevage:bovin": "les bovins (ferme d'élevage)",
		"elevage:ovin": "les ovins (ferme d'élevage)", "elevage:porc": "les porcs (ferme d'élevage)",
		"elevage:abattoir_canard": "l'abattoir des canards (ferme d'élevage)", "elevage:abattoir_boeuf": "l'abattoir des bœufs (ferme d'élevage)",
		"elevage:abattoir_veau": "l'abattoir des veaux (ferme d'élevage)", "elevage:abattoir_mouton": "l'abattoir des moutons (ferme d'élevage)",
		"elevage:abattoir_agneau": "l'abattoir des agneaux (ferme d'élevage)", "elevage:abattoir_porc": "l'abattoir des porcs (ferme d'élevage)",
		"boucherie:saucisse": "les saucisses (boucherie)", "boucherie:hache": "le steak haché (boucherie)", "boucherie:magret": "les magrets (boucherie)",
		"ferme:chou": "la salade (ma ferme)", "ferme:poireau": "les poireaux (ma ferme)", "port:thon": "le thon (port)",
		"exotique:bananier": "les bananes (ferme exotique)", "exotique:cannelle": "la cannelle (ferme exotique)",
		"cremerie:fromage_brebis": "le fromage de brebis (crèmerie)", "glacier:glace_vanille": "la glace à la vanille (glacier)", "domaine:whisky": "le whisky (domaine)"}
	var parts []string
	for _, n := range strings.Split(needs, ",") {
		t := txt[n]
		if t == "" {
			id, _, _ := strings.Cut(n, ":")
			t = placeName(id)
		}
		parts = append(parts, t)
	}
	return "Il faut " + strings.Join(parts, " et ")
}

// builtOrUnlocked : la station « id » est construite, ou le déblocage « id » est
// payé (« elevage:bovin » : le déblocage des bovins, dont le pré s'appelle pre_bovin).
func (p *Place) builtOrUnlocked(id string) bool {
	if s := p.Station(id); s != nil {
		return s.Built
	}
	for _, u := range p.Unlocks {
		if u.ID == id {
			return u.Done
		}
	}
	return false
}
