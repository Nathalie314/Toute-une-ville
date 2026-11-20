package sim

import "math/rand"

// L'huilerie (15e lieu de la carte, après le parc d'attractions) : aucune vente sur
// place, aucune commande. Six cultures, chacune dans son espace avec ses ouvriers :
//
//   champs : tournesol, colza, lin (des graines, qui ne se vendent pas)
//   arbres : oliviers, noisetiers, noyers (trois colonnes chacun, comme le verger)
//
// Sous chaque culture, son extracteur (presse ou moulin) fait l'huile. Les huiles
// partent au supermarché (espace « Huiles »), les olives au rayon Fruits, les
// noisettes et les noix à l'Épicerie ; l'huile d'olive ira aussi à la future
// pizzeria (son quai attend). Rien n'est offert : même le premier champ s'achète.

// Géométrie de l'huilerie.
const (
	HuilW      = 5420.0
	HuilTop    = -420.0 // (les arbres montent haut)
	HuilZoneX0 = 420.0  // bord gauche du premier espace
	HuilZoneW  = 760.0  // largeur d'un espace
	HuilZoneDX = 800.0  // d'un espace au suivant
)

// HuilCulture : une culture de l'huilerie (son espace, sa matière, son huile).
type HuilCulture struct {
	ID, Name   string // ID : celui de l'espace et du champ (ou de la 1re colonne)
	Raw, Oil   Item
	Tree       bool // des arbres (trois colonnes) plutôt qu'un champ
	Press      string
	PressName  string
	Cost, PCst float64 // prix de base : la culture, son extracteur
	Col2, Col3 float64 // arbres : prix des 2e et 3e colonnes
}

// HuilCultures : dans l'ordre d'achat (et de gauche à droite).
var HuilCultures = []HuilCulture{
	{ID: "tournesol", Name: "Tournesol", Raw: Tournesol, Oil: HuileTournesol, Press: "presse_tournesol", PressName: "Presse à tournesol", Cost: 5000, PCst: 9000},
	{ID: "colza", Name: "Colza", Raw: Colza, Oil: HuileColza, Press: "presse_colza", PressName: "Presse à colza", Cost: 15000, PCst: 22000},
	{ID: "olivier", Name: "Oliviers", Raw: Olive, Oil: HuileOlive, Tree: true, Press: "moulin_olive", PressName: "Moulin à huile d'olive", Cost: 30000, PCst: 38000, Col2: 45000, Col3: 52000},
	{ID: "lin", Name: "Lin", Raw: Lin, Oil: HuileLin, Press: "presse_lin", PressName: "Presse à lin", Cost: 60000, PCst: 70000},
	{ID: "noisetier", Name: "Noisetiers", Raw: Noisette, Oil: HuileNoisette, Tree: true, Press: "presse_noisette", PressName: "Presse à noisettes", Cost: 80000, PCst: 90000, Col2: 100000, Col3: 110000},
	{ID: "noyer", Name: "Noyers", Raw: Noix, Oil: HuileNoix, Tree: true, Press: "presse_noix", PressName: "Presse à noix", Cost: 125000, PCst: 140000, Col2: 155000, Col3: 170000},
}

// huilZones : un espace par culture ; un employé par ressource (le champ et sa presse ;
// les trois colonnes d'arbres et leur presse).
var huilZones = func() []*ShopZone {
	var zs []*ShopZone
	for i, c := range HuilCultures {
		x0 := HuilZoneX0 + float64(i)*HuilZoneDX
		costs := []float64{20000, 50000}
		if c.Tree {
			costs = []float64{20000, 50000, 110000, 240000}
		}
		for k := range costs {
			costs[k] *= 1 + float64(i)*0.5 // (les espaces suivants coûtent un peu plus)
		}
		zs = append(zs, &ShopZone{ID: c.ID, Name: c.Name, X0: x0, X1: x0 + HuilZoneW, RestY: 980, Costs: costs})
	}
	return zs
}()

// huilTreeTeam : un espace d'arbres (ses ouvriers ont une cagette pour toute une colonne).
func huilTreeTeam(team string) bool {
	for _, c := range HuilCultures {
		if c.ID == team {
			return c.Tree
		}
	}
	return false
}

func newHuilerie(g *Game) *Place {
	W := HuilW
	p := &Place{ID: "huilerie", Name: "L'huilerie", W: W, H: 1800, Top: HuilTop, g: g, rng: rand.New(rand.NewSource(97)),
		Rep: 3, Price: 1, SellMult: 1, SpawnBase: 5, WantMax: 3, MaxCust: 0, LeaveOff: V{170, 60}, OnlyStocked: true}
	p.SpawnL = V{350, roadY + 40}
	p.SpawnR = V{W - 550, roadY + 40}
	p.Gate = V{1250, roadY + 40}
	p.Inside = V{1250, 1390}
	p.H = ParkH // (pas de parking : pas de clients)
	p.RestSpot = V{W - 400, 1440}

	// une réserve sans clients (huile, olives, noisettes, noix) : un drone l'emporte au quai
	rack := func(id, name, team string, it Item, x float64, unl string) {
		s := &Station{ID: id, Name: name, Kind: Shelf, Item: it, StockMax: 30, Unlock: unl, NoSale: true, Team: team}
		s.Body = Rect{x, shelfY, 220, shelfH}
		s.Zone = Rect{x - 10, shelfY - 95, 240, 92}
		s.Spot = V{x + 110, shelfY + shelfH + 38}
		p.Stations = append(p.Stations, s)
	}
	for i, c := range HuilCultures {
		x0 := HuilZoneX0 + float64(i)*HuilZoneDX
		if c.Tree {
			// trois colonnes d'arbres (une case au départ, les autres s'achètent vers le haut)
			for k := 0; k < 3; k++ {
				id := c.ID
				if k > 0 {
					id = c.ID + "_" + itoa(k+1)
				}
				x := x0 + 30 + float64(k)*(ColumnW+40)
				s := &Station{ID: id, Name: c.Name, Kind: Field, Item: c.Raw, Time: 10 + float64(i), Unlock: id, Column: true, Team: c.ID}
				bottom := fieldY + 60 + 2*90
				s.Body = Rect{x + 10, bottom - float64(ColumnMax-1)*90 - 60, 130, float64(ColumnMax-1)*90 + 120}
				s.Zone = s.Body
				for j := 0; j < ColumnMax; j++ {
					s.Slots = append(s.Slots, V{x + 75, bottom - float64(j)*90})
				}
				s.Plants = []Plant{{P: s.Slots[0], Young: 1}} // l'arbre pousse d'abord
				p.Stations = append(p.Stations, s)
			}
		} else {
			// un champ de graines (huit colonnes, trois rangées)
			cols := 8
			s := &Station{ID: c.ID, Name: c.Name, Kind: Field, Item: c.Raw, Time: 6 + float64(i), Unlock: c.ID, Team: c.ID}
			s.Body = Rect{x0 + 40, fieldY, 40 + 80*float64(cols), fieldH}
			s.Zone = s.Body
			for k, r := range []int{2, 1, 0} {
				for j := 0; j < cols; j++ {
					cc := j
					if k%2 == 1 {
						cc = cols - 1 - j
					}
					s.Slots = append(s.Slots, V{x0 + 100 + float64(cc)*80, fieldY + 60 + float64(r)*90})
				}
			}
			s.Plants = []Plant{{P: s.Slots[0], G: 1}}
			p.Stations = append(p.Stations, s)
		}
		// l'extracteur, sous la culture
		m := &Station{ID: c.Press, Name: c.PressName, Kind: Machine, Item: c.Oil, Inputs: []Item{c.Raw}, Time: 4 + float64(i)*0.5,
			PileMax: 12, BufMax: 10, Unlock: c.Press, Team: c.ID}
		mx := x0 + 60
		m.Body = Rect{mx, bldY, 260, bldH}
		m.Zone = Rect{mx - 10, bldY + bldH + 18, 130, 90}
		m.OutZone = Rect{mx + 140, bldY + bldH + 18, 130, 90}
		p.Stations = append(p.Stations, m)
		// les réserves : l'huile ; pour les arbres, aussi le fruit tel quel (supermarché)
		rack("stock_"+itoa(int(c.Oil)), "Réserve : "+Items[c.Oil].Name, c.ID, c.Oil, x0+40, c.Press)
		if c.Tree {
			rack("stock_"+itoa(int(c.Raw)), "Réserve : "+Items[c.Raw].Name, c.ID, c.Raw, x0+300, c.ID)
		}
	}
	trash := &Station{ID: "poubelle", Name: "Compost", Kind: Trash}
	trash.Body = Rect{W - 250, shelfY - 10, 84, 76}
	trash.Zone = Rect{W - 270, shelfY - 100, 124, 180}
	p.Stations = append(p.Stations, trash)
	// les quais (au bord de la route) : le supermarché, et la pizzeria (pas encore là : son quai reste invisible)
	p.dock("quai", "super", 2620)
	p.dock("quai_pizzeria", "pizzeria", 2970)

	// --- déblocages, dans l'ordre (rien n'est offert)
	unlock := func(id, name string, cost float64, hint string) {
		var at Rect
		for _, s := range p.Stations {
			if s.Unlock == id && (s.Kind == Field || s.Kind == Machine) {
				c := s.Body.C()
				switch {
				case s.Column:
					at = Rect{s.Body.X, 290, 130, 130} // (sur la case du bas, près du chemin)
				case s.Kind == Field:
					at = Rect{s.Body.X + 30, s.Body.Y + s.Body.H - 140, 180, 130}
				default:
					at = Rect{c.X - 90, c.Y - 65, 180, 130}
				}
				break
			}
		}
		p.Unlocks = append(p.Unlocks, &Unlock{ID: id, Name: name, Cost: cost, Zone: at, Hint: hint})
	}
	treeHint := "Un arbre : il pousse d'abord, puis donne sans qu'on replante. Achète des cases au-dessus."
	for _, c := range HuilCultures {
		if c.Tree {
			unlock(c.ID, c.Name, c.Cost, treeHint)
		} else {
			unlock(c.ID, "Champ : "+lowerFirst(c.Name), c.Cost, "Les graines ne se vendent pas : elles partent à la presse.")
		}
		unlock(c.Press, c.PressName, c.PCst, "Elle fait l'huile, qui part en camion au supermarché (espace « Huiles »).")
		if c.Tree {
			unlock(c.ID+"_2", c.Name+" (2e colonne)", c.Col2, treeHint)
			unlock(c.ID+"_3", c.Name+" (3e colonne)", c.Col3, treeHint)
		}
	}

	// --- améliorations
	k := p.key
	p.ups = []*Upgrade{
		{Key: k("etals"), Name: "Réserves plus grandes", Max: 6, Base: 6000, Mult: 2.3, Desc: func(l int) string { return "Chaque réserve contient " + itoa(30+4*l) + " produits." }},
	}
	p.ups = append(p.ups, p.teamUpgrades("Ouvrier", "Il ne travaille que dans son espace.")...)
	p.ups = append(p.ups, p.droneUpgrades(3)...)
	for _, st := range p.Stations {
		if st.Kind == Shelf {
			it := st.Item
			p.Spots = append(p.Spots, &Spot{ID: droneKey(it), Label: "Drone", Shelf: st, Zone: st.DronePad(),
				Done: func(int) string { return "Drone acheté pour « " + Items[it].Name + " » !" }})
		}
	}
	p.setupYard(-620)
	p.teamSpots("Ouvrier")
	p.Spots = append(p.Spots, &Spot{ID: "etals", Label: "Réserves +", Zone: Rect{p.Yard.X + 10, yardRow1 + 140, 130, 124},
		Done: func(l int) string { return "Réserves agrandies : niveau " + itoa(l) + " !" }})
	return p
}
