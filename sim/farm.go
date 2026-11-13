package sim

import "math/rand"

// La ferme de Clairbourg.
//
//   rangée 1 : les champs (on marche dedans)
//   rangée 2 : les bâtiments (moulin, four, animaux…)
//   rangée 3 : la boutique (étals + caisse)
//   en bas   : la route et le portail

const (
	fieldY = 150.0
	fieldH = 300.0
	bldY   = 600.0
	bldH   = 170.0
	shelfY = 1130.0
	shelfH = 66.0
	roadY  = 1640.0
)

// la serre, le verger et les fleurs (avant le décalage de la cour)
const (
	ColumnW   = 150.0
	ColumnMax = 8
	SerreX    = 2980.0
	VergerX   = SerreX + 6*ColumnW + 100
	FleursX   = VergerX + 6*ColumnW + 100
	FarmW     = FleursX + 910 // (champ de fleurs doublé : 6 colonnes)
	FarmTop   = -420.0
)

// ColumnIDs : les colonnes de la serre puis du verger (= ID du champ et de son équipe).
var ColumnIDs = []string{"carotte", "tomate", "courgette", "poireau", "chou", "pdt", "fraisier", "pommier", "poirier", "cerisier", "abricotier", "figuier"}

type columnBuy struct {
	id, name string
	cost     float64
	fruit    bool
}

// ColumnOrder : ordre et prix d'achat des colonnes (serre et verger mêlés).
var ColumnOrder = []columnBuy{
	{"carotte", "Serre : carottes", 6000, false},
	{"tomate", "Tomates", 7500, false},
	{"fraisier", "Verger : fraisiers", 9000, true},
	{"courgette", "Courgettes", 11000, false},
	{"pommier", "Pommier", 13000, true},
	{"poireau", "Poireaux", 16000, false},
	{"poirier", "Poirier", 19000, true},
	{"chou", "Choux", 22000, false},
	{"cerisier", "Cerisier", 26000, true},
	{"pdt", "Pommes de terre", 30000, false},
	{"abricotier", "Abricotier", 35000, true},
	{"figuier", "Figuier", 42000, true},
}

func newFarm(g *Game) *Place {
	p := &Place{ID: "ferme", Name: "Ma ferme", Rep: 3, Owned: true, SellMult: 1, SpawnBase: 4.5, WantMax: 3, MaxCust: 16, LeaveOff: V{170, 60}, RepGain: 0.012, W: FarmW, H: 1800, Top: FarmTop, g: g, rng: rand.New(rand.NewSource(11))}
	p.SpawnL = V{350, roadY + 40}
	p.SpawnR = V{2150, roadY + 40}
	p.Gate = V{1250, roadY + 40}
	p.Inside = V{1250, 1390}
	p.Exit = V{FarmW - 120, roadY + 40}
	p.RestSpot = V{3050, 1440}
	p.Start = V{1250, 1440} // juste après le portail, comme dans les autres lieux

	field := func(id, name string, it Item, x float64, t float64, unlock string, cost float64, cols int) {
		s := &Station{ID: id, Name: name, Kind: Field, Item: it, Time: t, Unlock: unlock, BaseCost: cost}
		s.Body = Rect{x, fieldY, 40 + 80*float64(cols), fieldH}
		s.Zone = s.Body
		// on achète les pousses une par une, en serpent : la rangée du bas de
		// gauche à droite, celle du milieu de droite à gauche, puis celle du haut
		//   7 8 9
		//   6 5 4
		//   1 2 3
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
	}
	bx := func(i int) float64 { return 140 + float64(i)*470 }
	machine := func(id, name string, out Item, in []Item, i int, t float64, unlock string, cost float64) {
		x := bx(i)
		s := &Station{ID: id, Name: name, Kind: Machine, Item: out, Inputs: in, Time: t, PileMax: 12, BufMax: 10, Unlock: unlock, BaseCost: cost}
		s.Body = Rect{x, bldY, 260, bldH}
		s.Zone = Rect{x - 10, bldY + bldH + 18, 130, 90}
		s.OutZone = Rect{x + 140, bldY + bldH + 18, 130, 90}
		p.Stations = append(p.Stations, s)
	}

	// --- champs, chacun au-dessus des bâtiments qu'il nourrit
	//   maïs  → moulin à maïs, poulailler
	//   blé   → moulin à blé, étable, chèvrerie, bergerie
	//   serre (6 légumes) et verger (6 arbres) : une colonne chacun, avec son petit marché dessous
	//   fleurs → ruches
	// (4.1 : plus de viande ; la canne, le cacao et le café iront à la ferme exotique)
	fx := func(i int) float64 { return 140 + float64(i)*470 }
	field("mais", "Champ de maïs", Mais, fx(0), 5.5, "", 20, 10) // au-dessus de ses 2 bâtiments (5,5 s : un début de partie plus rapide ; avant 7)
	field("ble", "Champ de blé", Ble, fx(2), 5, "ble", 400, 22)  // au-dessus de ses 4 bâtiments
	// colonnes de la serre et du verger : une case au départ, les autres
	// s'achètent une à une vers le haut (jusqu'à ColumnMax)
	// (les légumes de la serre repoussent tout seuls, comme le reste : plus de graines)
	column := func(id, name string, it Item, x float64, t float64, seed bool) {
		s := &Station{ID: id, Name: name, Kind: Field, Item: it, Time: t, Unlock: id, Column: true, Serre: seed}
		bottom := fieldY + 60 + 2*90 // comme la rangée du bas des autres champs
		s.Body = Rect{x + 10, bottom - float64(ColumnMax-1)*90 - 60, 130, float64(ColumnMax-1)*90 + 120}
		s.Zone = s.Body
		for k := 0; k < ColumnMax; k++ {
			s.Slots = append(s.Slots, V{x + 75, bottom - float64(k)*90})
		}
		s.Plants = []Plant{{P: s.Slots[0], G: 1}}
		if !seed {
			s.Plants[0] = Plant{P: s.Slots[0], Young: 1} // verger : l'arbre pousse d'abord
		}
		p.Stations = append(p.Stations, s)
		// son étal, au petit marché
		m := &Station{ID: "etal_" + itoa(int(it)), Name: "Étal : " + Items[it].Name, Kind: Shelf, Item: it, StockMax: 12, Unlock: id, Market: true}
		m.Body = Rect{x + 12, shelfY, 126, shelfH}
		m.Zone = Rect{x + 5, shelfY - 95, 140, 92}
		m.Spot = V{x + 75, shelfY + shelfH + 38}
		p.Stations = append(p.Stations, m)
	}
	vegNames := []string{"Carottes", "Tomates", "Courgettes", "Poireaux", "Choux", "Pommes de terre"}
	fruitNames := []string{"Fraisiers", "Pommier", "Poirier", "Cerisier", "Abricotier", "Figuier"}
	for i, it := range Legumes {
		column(ColumnIDs[i], vegNames[i], it, SerreX+float64(i)*ColumnW, 9+float64(i), true)
	}
	for i, it := range Fruits {
		column(ColumnIDs[6+i], fruitNames[i], it, VergerX+float64(i)*ColumnW, 8+float64(i), false)
	}
	field("fleurs", "Champ de fleurs", Fleur, FleursX, 14, "fleurs", 16000, 6) // (doublé : 6 colonnes)
	// réserve de fleurs, à droite du champ : elle garde le surplus pour le
	// futur fleuriste (un drone viendra s'y servir : sa base est prévue dessous)
	{
		f := p.Station("fleurs").Body
		st := &Station{ID: "reserve_fleurs", Name: "Réserve de fleurs", Kind: Store, Item: Fleur, StockMax: 40, Unlock: "fleuriste"} // invisible tant que le fleuriste n'existe pas
		st.Body = Rect{f.X + f.W + 15, fieldY + 10, 160, 132}
		st.Zone = Rect{st.Body.X - 5, st.Body.Y + st.Body.H + 10, 170, 80}
		p.Stations = append(p.Stations, st)
	}
	// l'étang (décor) au-dessus du maïs et du blé, là où la serre et le verger montent
	etang := &Station{ID: "etang", Name: "Étang", Kind: Trash, Built: true}
	etang.Body = Rect{330, -330, 860, 330} // (étape 6 : plus petit, le grenier à foin est à côté)
	etang.Zone = Rect{-1000, -1000, 0, 0}
	p.Stations = append(p.Stations, etang)
	// le grenier à foin, au-dessus du champ de blé (il n'apparaît qu'avec la ferme
	// d'élevage) : le blé devient du foin, rangé dans la grange (un drone l'emporte)
	{
		s := &Station{ID: "grenier", Name: "Grenier à foin", Kind: Machine, Item: Foin, Inputs: []Item{Ble}, Time: 2.5, PileMax: 12, BufMax: 10, Unlock: "grenier", BaseCost: 3000}
		s.Body = Rect{1850, -330, 260, 170} // (pile au milieu du champ de blé, au-dessus)
		s.Zone = Rect{1840, -142, 130, 90}
		s.OutZone = Rect{1990, -142, 130, 90}
		p.Stations = append(p.Stations, s)
		gr := &Station{ID: "grange", Name: "Grange à foin", Kind: Shelf, Item: Foin, StockMax: 30, Unlock: "grenier", NoSale: true}
		gr.Body = Rect{2320, -300, 200, 66} // (assez loin : la case « Stock + » du grenier ne la touche pas)
		gr.Zone = Rect{2310, -225, 220, 92}
		gr.Spot = V{2420, -120}
		p.Stations = append(p.Stations, gr)
		// le silo à blé, à gauche du grenier (il n'apparaît qu'avec le domaine) :
		// les ouvriers du blé y rangent du blé, un drone de fret l'emporte au domaine (whisky)
		si := &Station{ID: "silo", Name: "Silo à blé", Kind: Shelf, Item: Ble, StockMax: 30, Unlock: "silo", NoSale: true}
		si.Body = Rect{1540, -300, 200, 66}
		si.Zone = Rect{1530, -225, 220, 92}
		si.Spot = V{1640, -120}
		p.Stations = append(p.Stations, si)
	}

	// --- bâtiments
	machine("moulin_mais", "Moulin à maïs", FarineMais, []Item{Mais}, 0, 2.5, "moulin_mais", 60)
	machine("poulailler", "Poulailler", Oeuf, []Item{Mais}, 1, 3.5, "poulailler", 200) // les poules mangent du maïs
	machine("moulin", "Moulin à blé", Farine, []Item{Ble}, 2, 2.2, "moulin", 700)
	machine("etable", "Étable", Lait, []Item{Ble}, 3, 3, "etable", 1800) // les vaches mangent du blé
	machine("chevrerie", "Chèvrerie", LaitChevre, []Item{Ble}, 4, 3.5, "chevrerie", 3500)
	machine("bergerie", "Bergerie", LaitBrebis, []Item{Ble}, 5, 4, "bergerie", 5000)
	{
		// six ruches (avant : trois) : deux fois plus de miel (6 s → 3 s)
		s := &Station{ID: "ruche", Name: "Ruches", Kind: Machine, Item: Miel, Inputs: []Item{Fleur}, Time: 3, PileMax: 12, BufMax: 10, Unlock: "ruche"} // les abeilles butinent les fleurs
		s.Body = Rect{FleursX, bldY, 500, bldH}
		s.Zone = Rect{FleursX - 10, bldY + bldH + 18, 130, 90}
		s.OutZone = Rect{FleursX + 140, bldY + bldH + 18, 130, 90}
		p.Stations = append(p.Stations, s)
	}
	// l'étable s'améliore à petit prix, pour suivre la demande en lait
	p.Station("etable").UpMult = 0.3
	// une ration de blé donne 2 bidons de lait (il en faut pour l'étal et pour
	// les camions de la crèmerie, du glacier et de la boulangerie)
	for _, id := range []string{"etable", "chevrerie", "bergerie"} {
		p.Station(id).Yield = 2
	}

	// --- boutique : les étals (larges) jusqu'à la serre, puis les petits marchés, puis le miel
	unl := map[Item]string{Mais: "", Farine: "moulin", FarineMais: "moulin_mais", Oeuf: "poulailler",
		Lait: "etable", LaitChevre: "chevrerie", LaitBrebis: "bergerie"}
	big := []Item{Mais, FarineMais, Oeuf, Farine, Lait, LaitChevre, LaitBrebis}
	step := (SerreX - 70 - 450) / float64(len(big))
	for i, it := range big {
		x := 450 + float64(i)*step
		s := &Station{ID: "etal_" + itoa(int(it)), Name: "Étal : " + Items[it].Name, Kind: Shelf, Item: it, StockMax: 12, Unlock: unl[it]}
		s.Body = Rect{x, shelfY, step - 50, shelfH}
		s.Zone = Rect{x - 10, shelfY - 95, step - 30, 92}
		s.Spot = V{x + (step-50)/2, shelfY + shelfH + 38}
		p.Stations = append(p.Stations, s)
	}
	{
		x := FleursX + 30.0
		s := &Station{ID: "etal_" + itoa(int(Miel)), Name: "Étal : " + Items[Miel].Name, Kind: Shelf, Item: Miel, StockMax: 12, Unlock: "ruche"}
		s.Body = Rect{x, shelfY, 200, shelfH}
		s.Zone = Rect{x - 10, shelfY - 95, 220, 92}
		s.Spot = V{x + 100, shelfY + shelfH + 38}
		p.Stations = append(p.Stations, s)
	}
	reg := &Station{ID: "caisse", Name: "Caisse", Kind: Register, HireKey: "caissiere", QueueDown: 5}
	reg.Body = Rect{180, shelfY, 210, shelfH}
	reg.Spot = V{285, shelfY - 40}
	reg.Zone = Rect{200, shelfY - 100, 170, 96}
	p.Stations = append(p.Stations, reg)
	reg.QueueHead = V{285, shelfY + shelfH + 40}
	trash := &Station{ID: "poubelle", Name: "Compost", Kind: Trash}
	// à côté de l'étal du miel (un peu à droite, pas collé)
	trash.Body = Rect{5290, shelfY - 10, 84, 76} // (avant le décalage de la ferme)
	trash.Zone = Rect{5270, shelfY - 100, 124, 180}
	// quais d'expédition : un par destination (servent une fois le lieu acheté)
	p.dock("quai", "super", 1700)
	p.dock("quai_boul", "boulangerie", 2050)
	p.dock("quai_creme", "cremerie", 2400)
	p.dock("quai_glace", "glacier", 2750)
	p.dock("quai_elevage", "elevage", 3100)
	p.dock("quai_domaine", "domaine", 3450)
	// la cour, à gauche de la boutique (près du portail) : caisse des commandes et hangar
	cmd := &Station{ID: "commande", Name: "Commandes", Kind: OrderBox, Built: true}
	cmd.Body = Rect{230, 1110, 180, 70}
	cmd.Zone = Rect{210, 1190, 220, 100}
	p.Stations = append(p.Stations, cmd)
	cmd.yard = true
	// améliorations au sol
	spot := func(id, label string, z Rect, needs string, done func(int) string) {
		p.Spots = append(p.Spots, &Spot{ID: id, Label: label, Zone: z, Needs: needs, Done: done})
	}
	spot("caissiere", "Caissière", Rect{40, 960, 145, 130}, "", func(int) string { return "Caissière embauchée !" })
	// drones : un par étal (sa base est sous l'étal), et leurs améliorations au hangar
	for _, st := range p.Stations {
		if st.Kind == Shelf {
			it := st.Item
			p.Spots = append(p.Spots, &Spot{ID: droneKey(it), Label: "Drone", Shelf: st, Zone: st.DronePad(),
				Done: func(int) string { return "Drone acheté pour l'étal « " + Items[it].Name + " » !" }})
		}
	}
	// les ouvriers s'embauchent dans leur zone, sur le chemin sous le champ (à côté de son « Niveau »)
	for _, t := range FarmTeams {
		t := t
		f := p.Station(teamAnchor(t.ID))
		b := f.Body
		z := Rect{b.X + b.W/2 - 56, b.Y + b.H + 14, 112, 92} // sur le chemin, sous le milieu du champ
		if f.Column {
			// serre et verger : une grande case au milieu, sous les six colonnes, dans
			// l'herbe entre les deux chemins (580 à 880) : les ouvriers sont pour
			// toute la serre, tout le verger
			cx := SerreX + 3*ColumnW
			if t.ID == "verger" {
				cx = VergerX + 3*ColumnW
			}
			z = Rect{cx - 150, 684, 300, 92}
		}
		p.Spots = append(p.Spots, &Spot{ID: teamKey(t.ID), Label: "Ouvrier", Team: t.ID, Zone: z,
			Done: func(l int) string { return "Ouvrier " + t.Name + " +1 !" }})
	}
	spot("etals", "Étals +", Rect{40, 1105, 130, 120}, "", func(l int) string { return "Étals agrandis : niveau " + itoa(l) + " !" })
	p.Stations = append(p.Stations, trash)

	// --- déblocages, dans l'ordre
	unlock := func(id, name string, cost float64, hint string) {
		var at Rect
		for _, s := range p.Stations {
			if s.Unlock == id && s.Kind != Shelf && s.Kind != Store && s.Kind != Pallet {
				c := s.Body.C()
				at = Rect{c.X - 90, c.Y - 65, 180, 130}
				if s.Column {
					at = Rect{s.Body.X, 290, 130, 130} // sur la case du bas, près du chemin
				}
			}
		}
		p.Unlocks = append(p.Unlocks, &Unlock{ID: id, Name: name, Cost: cost, Zone: at, Hint: hint})
	}
	unlock("moulin_mais", "Moulin à maïs", 60, "Le maïs devient de la farine de maïs.")
	unlock("poulailler", "Poulailler", 200, "Donne du maïs aux poules : elles pondent.")
	unlock("ble", "Champ de blé", 400, "Le blé ne se vend pas : il nourrit le moulin et les vaches.")
	unlock("moulin", "Moulin à blé", 700, "Le blé devient de la farine.")
	unlock("etable", "Étable", 1800, "Donne du blé aux vaches : elles donnent du lait.")
	unlock("chevrerie", "Chèvrerie", 3500, "Les chèvres mangent du blé : lait de chèvre.")
	unlock("bergerie", "Bergerie", 5000, "Les brebis mangent du blé : lait de brebis.")
	for _, c := range ColumnOrder {
		hint := "Va chercher les sachets de graines au pied de la colonne : il faut replanter après chaque récolte."
		if c.fruit {
			hint = "Un arbre : on récolte sans replanter. Achète des cases au-dessus pour en avoir plus."
		}
		if c.id == "carotte" {
			hint = "La serre ! " + hint
		}
		unlock(c.id, c.name, c.cost, hint)
	}
	unlock("fleurs", "Champ de fleurs", 50000, "Les fleurs nourrissent les abeilles.")
	unlock("ruche", "Ruches", 65000, "Les abeilles transforment les fleurs en miel.")
	unlock("grenier", "Grenier à foin", 20000, "Le blé devient du foin, pour les bovins et les ovins de la ferme d'élevage.")
	{
		// il passe avant les autres (dès que la ferme d'élevage est achetée), sans se montrer avant
		u := p.Unlocks[len(p.Unlocks)-1]
		u.Requires = func() bool { q := g.Place("elevage"); return q != nil && q.Owned }
		u.Locked, u.Secret = "Il faut la ferme d'élevage", true
		p.Unlocks = append([]*Unlock{u}, p.Unlocks[:len(p.Unlocks)-1]...)
	}
	unlock("silo", "Silo à blé", 30000, "Les ouvriers du blé y rangent du blé ; un drone l'emporte au domaine (whisky).")
	{
		// comme le grenier : dès que le domaine est acheté, sans se montrer avant
		u := p.Unlocks[len(p.Unlocks)-1]
		c := p.Station("silo").Body.C()
		u.Zone = Rect{c.X - 90, c.Y - 65, 180, 130}
		u.Requires = func() bool { q := g.Place("domaine"); return q != nil && q.Owned }
		u.Locked, u.Secret = "Il faut le domaine", true
		p.Unlocks = append([]*Unlock{u}, p.Unlocks[:len(p.Unlocks)-1]...)
	}

	for _, s := range p.Stations {
		if s.Unlock == "" {
			s.Built = true
		}
	}
	p.shiftFarm(FarmX0)
	p.setupYard(70) // la cour, à gauche
	p.setupParking()
	p.spawnT = 0.5
	return p
}

// FarmX0 : la ferme commence après la cour (hangar et commandes), tout à gauche.
const FarmX0 = 650.0

// shiftFarm décale la ferme vers la droite pour laisser la cour à gauche
// (le portail, l'entrée et le point d'arrivée ne bougent pas : la cour est près de l'arrivée).
func (p *Place) shiftFarm(dx float64) {
	p.W += dx
	p.Exit.X += dx
	p.RestSpot.X += dx
	mv := func(r *Rect) {
		if r.W > 0 || r.H > 0 {
			r.X += dx
		}
	}
	for _, s := range p.Stations {
		if s.yard || s.Kind == Dock {
			continue
		}
		mv(&s.Body)
		mv(&s.Zone)
		mv(&s.OutZone)
		if s.Spot != (V{}) {
			s.Spot.X += dx
		}
		if s.QueueHead != (V{}) {
			s.QueueHead.X += dx
		}
		for i := range s.Slots {
			s.Slots[i].X += dx
		}
		for i := range s.Plants {
			s.Plants[i].P.X += dx
		}
	}
	for _, u := range p.Unlocks {
		mv(&u.Zone)
	}
	for _, sp := range p.Spots {
		switch sp.ID {
		case "vitesse", "capacite", "drone_cap", "drone_vit":
			continue // dans la cour, près du hangar
		}
		mv(&sp.Zone)
	}
}

// dock ajoute un quai d'expédition au bord de la route, vers un autre lieu.
func (p *Place) dock(id, to string, x float64) {
	q := &Station{ID: id, Name: "Quai du camion", Kind: Dock, BufMax: 20, Built: true, To: to}
	q.Body = Rect{x, 1585, 230, 50}
	q.Zone = Rect{x - 10, 1470, 250, 110}
	p.Stations = append(p.Stations, q)
}

// ---------------------------------------------------------------- équipes

// Team : les ouvriers d'une zone de la ferme. Ils ne travaillent que pour
// les bâtiments et les étals de leur zone, et ne remplissent pas les quais.
type Team struct {
	ID    string // = ID du champ de la zone
	Name  string // ex. « du maïs »
	Max   int
	Items []Item // ce que produit la zone
	Field string // « le champ de maïs »…
}

var FarmTeams = func() []*Team {
	ts := []*Team{
		{ID: "mais", Name: "du maïs", Max: 4, Items: []Item{Mais, FarineMais, Oeuf}, Field: "le champ de maïs"}, // champ, moulin, poulailler (+1 demandé par la conceptrice)
		{ID: "ble", Name: "du blé", Max: 5, // champ, moulin, étable, chèvrerie, bergerie
			Items: []Item{Ble, Farine, Lait, LaitChevre, LaitBrebis, Foin}, Field: "le champ de blé"},
	}
	// la serre et le verger : une équipe chacun (un ouvrier n'est pas attaché à un
	// légume ou à un fruit : il va où il y a du travail), autant d'ouvriers que de colonnes
	ts = append(ts, &Team{ID: "serre", Name: "de la serre", Max: 7, Items: append([]Item{}, Legumes...), Field: "la serre"},
		&Team{ID: "verger", Name: "du verger", Max: 6, Items: append([]Item{}, Fruits...), Field: "le verger"})
	return append(ts, &Team{ID: "fleurs", Name: "des fleurs", Max: 2, Items: []Item{Fleur, Miel}, Field: "le champ de fleurs"})
}()

// HireOrder : ordre conseillé des embauches ; il fixe les prix (le 1er le moins cher).
var HireOrder = []string{"mais", "ble", "mais", "ble", "serre", "ble", "serre", "verger", "ble", "serre", "verger", "ble",
	"serre", "verger", "serre", "verger", "mais", "serre", "verger", "verger", "fleurs", "fleurs", "mais", "serre"}

// teamAnchor : le champ d'une équipe (serre : la colonne des carottes ; verger : les fraisiers).
func teamAnchor(id string) string {
	switch id {
	case "serre":
		return "carotte"
	case "verger":
		return "fraisier"
	}
	return id
}

// columnHires : ouvriers possibles dans la serre (ou le verger) : un par colonne
// ouverte ; toutes les colonnes ouvertes, le maximum de l'équipe (la serre a un
// 7e ouvrier).
func (p *Place) columnHires(team string) int {
	n := p.columnsBuiltOf(team)
	if n >= 6 {
		for _, t := range FarmTeams {
			if t.ID == team {
				return t.Max
			}
		}
	}
	return n
}

// columnsBuiltOf : colonnes ouvertes de la serre (ou du verger).
func (p *Place) columnsBuiltOf(team string) int {
	ids := ColumnIDs[:6]
	if team == "verger" {
		ids = ColumnIDs[6:]
	}
	n := 0
	for _, id := range ids {
		if s := p.Station(id); s != nil && s.Built {
			n++
		}
	}
	return n
}

var HireCosts = []float64{250, 1000, 3500, 6000, 7000, 8000, 9000, 10000, 12000, 13000, 14000, 16000,
	18000, 20000, 23000, 26000, 30000, 34000, 40000, 48000, 60000, 75000, 95000, 120000}

func teamOf(it Item) string {
	for _, t := range FarmTeams {
		for _, x := range t.Items {
			if x == it {
				return t.ID
			}
		}
	}
	return ""
}

func teamKey(id string) string { return "ouvriers_" + id }

// teamCosts : prix des embauches d'une équipe, selon leur rang dans l'ordre d'embauche.
func teamCosts(id string) []float64 {
	var c []float64
	for i, t := range HireOrder {
		if t == id {
			c = append(c, HireCosts[i])
		}
	}
	return c
}

// NextHire : l'équipe du prochain ouvrier à embaucher ("" si tout le monde est là).
func (g *Game) NextHire() string {
	n := 0
	for _, t := range FarmTeams {
		n += g.Levels["ferme."+teamKey(t.ID)]
	}
	if n >= len(HireOrder) {
		return ""
	}
	return HireOrder[n]
}

// syncFarmTotal : « ferme.ouvriers » = nombre total d'ouvriers (badges, bottes, brouettes…).
func (g *Game) syncFarmTotal() {
	n := 0
	for _, t := range FarmTeams {
		n += g.Levels["ferme."+teamKey(t.ID)]
	}
	g.Levels["ferme.ouvriers"] = n
	for _, p := range g.Places {
		if p.hasDrones() {
			g.Levels[p.key("drones")] = p.DroneCount()
		}
	}
}

// teamAllows : un ouvrier d'équipe peut-il apporter ce produit à cette station ?
func teamAllows(w *Agent, dst *Station, it Item) bool {
	if dst.Kind == Dock && w.Kind == WorkerAgent {
		return false // les quais : les drones (ou toi)
	}
	if w.Team == "" {
		return true
	}
	if dst.place == "super" {
		return SuperZoneOf(dst.Item) == w.Team // un magasinier ne remplit que son espace
	}
	if dst.Team != "" || len(teamZonesOf(dst.place)) > 0 {
		return dst.Team == w.Team || dst.Kind == Trash // boucherie, ferme d'élevage : chaque espace a ses employés
	}
	return teamOf(dst.Item) == w.Team || (IsSeed(it) && dst.Kind == Field && teamOf(dst.Item) == w.Team)
}

// teamRest : point de repos d'une équipe, sur le chemin sous son champ.
func (p *Place) teamRest(id string) V {
	if p.ID == "super" {
		return superRest(id)
	}
	if teamZonesOf(p.ID) != nil {
		return p.zoneRest(id)
	}
	if s := p.Station(teamAnchor(id)); s != nil {
		return V{s.Body.C().X, s.Body.Y + s.Body.H + 70}
	}
	return p.RestSpot
}

// HireLock : plus utilisé (l'embauche est libre ; seuls les prix guident).
func (p *Place) HireLock(sp *Spot) string { return "" }
