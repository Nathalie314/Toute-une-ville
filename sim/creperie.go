package sim

import (
	"math"
	"sort"
	"strings"
)

// La crêperie (12e lieu, après le domaine) : le premier restaurant.
//
// Le bâtiment a deux pièces. En haut, la cuisine : la réserve (palettes livrées
// par camion), le poste des crêpes nature (son chef en fait sans arrêt) et les
// postes des crêpes garnies (un chef chacun, seulement sur commande). Les commis
// portent les ingrédients aux chefs et les plats sur le comptoir (le passe-plat,
// dans le mur entre les deux pièces). Un chef dont le plat n'est pas encore
// emporté n'en commence pas un autre.
//
// En bas, la salle : des tables de 2, puis de 4, puis une terrasse dehors, plus
// loin (tables de 2 et de 4). Les clients arrivent en voiture (2 par voiture,
// puis 2 ou 4 quand les tables de 4 existent), entrent et s'assoient seuls à une
// table propre ; sans table, ils attendent devant la porte. Un serveur prend la
// commande à la tablette (elle part directement en cuisine), apporte les crêpes
// depuis le comptoir et le cidre depuis le frigo de la salle, dans n'importe quel
// ordre, puis nettoie la table quand ils sont partis. Pas de caisse : une table
// paie ce qu'elle a reçu en partant. Pas de réputation : la jauge de bonheur de
// la table décide du pourboire (bonheur = chance d'en laisser un), que seul le
// joueur peut ramasser. Rien n'est offert : la première table, le premier poste…
// tout s'achète. Seul le joueur passe d'une pièce à l'autre.

// géométrie
const (
	CrepW      = 5300.0
	CrepTop    = -1180.0 // la route des livraisons, au-dessus de la cuisine
	CrepBX0    = 120.0   // mur de gauche
	CrepBX1    = 3480.0  // mur de droite (la terrasse est au-delà)
	CrepKY0    = -1010.0 // haut de la cuisine
	CrepPalY   = -960.0  // palettes
	CrepRowA   = -660.0  // 1re rangée de postes (crêpe nature, crêpes sucrées)
	CrepRowB   = -300.0  // 2e rangée (crêpes salées)
	CrepWallY  = 110.0   // le mur entre la cuisine et la salle (le comptoir y est pris)
	CrepWallH  = 70.0
	CrepBotY   = 1514.0 // mur du bas
	CrepDoorX  = 1780.0 // la porte des clients (mur du bas)
	CrepTerX0  = 3700.0 // la terrasse
	CrepTerX1  = 5050.0
	CrepTerDY  = 780.0 // la porte de la terrasse (mur de droite)
	crepParkN0 = 5     // places au départ
	crepParkUp = 5     // places par achat
)

// CrepTable : une table de la salle ou de la terrasse.
type CrepTable struct {
	ID      string
	Seats   int
	Pos     V // centre
	Terrace bool
	St      *Station // son achat (Built) et son obstacle
	Group   *CrepGroup
	Dirty   bool    // à nettoyer avant les clients suivants
	Tips    float64 // pourboires laissés (seul le joueur les ramasse)
	cleanT  float64
	claim   *Agent // serveur qui s'en occupe (commande ou nettoyage)
}

// Body : le plateau de la table.
func (t *CrepTable) Body() Rect {
	if t.Seats == 4 {
		return Rect{t.Pos.X - 90, t.Pos.Y - 50, 180, 100}
	}
	return Rect{t.Pos.X - 60, t.Pos.Y - 42, 120, 84}
}

// ServeAt : où se tient le serveur (et le joueur) pour s'occuper de la table.
func (t *CrepTable) ServeAt() V { return V{t.Pos.X, t.Body().Y + t.Body().H + 50} }

// Zone : la zone de service (le joueur s'y arrête).
func (t *CrepTable) Zone() Rect {
	b := t.Body()
	return Rect{b.X - 70, b.Y - 60, b.W + 140, b.H + 150}
}

// Seat : la chaise n° i (à gauche puis à droite de la table).
func (t *CrepTable) Seat(i int) V {
	b := t.Body()
	side := -1.0
	if i%2 == 1 {
		side = 1
	}
	x := t.Pos.X + side*(b.W/2+34)
	if t.Seats == 4 {
		return V{x, t.Pos.Y - 24 + float64(i/2)*48}
	}
	return V{x, t.Pos.Y + 4}
}

// Diner : un client de la crêperie (il fait partie d'une table).
type Diner struct {
	Pos, Face V
	WalkT     float64
	Moving    bool
	Look      int
	path      []V
}

// états d'une table de clients
const (
	cgCar      = iota // dans la voiture
	cgWalk            // vers la porte
	cgQueue           // devant la porte, sans table
	cgToTable         // vers sa table
	cgSeated          // attend qu'on prenne la commande
	cgWaitFood        // commande prise, attend crêpes et cidre
	cgEat             // tout est servi : on mange
	cgLeave           // repart vers la voiture
	cgGone
)

// CrepGroup : les clients d'une voiture (une table).
type CrepGroup struct {
	N       int
	Members []*Diner
	State   int
	T       float64 // temps dans l'état
	wait    float64 // attente totale depuis l'arrivée à table
	Table   *CrepTable
	Order   []Item // la commande complète (choisie avant d'entrer)
	Wants   []Item // ce qui reste à servir (une fois la commande prise)
	Got     []Item
	Happy   float64 // jauge de bonheur (0 à 1)
	Angry   bool
	orderT  float64 // avancement de la prise de commande
	veh     *Vehicle
	slot    int
}

// Remaining : combien de ce produit la table attend encore.
func (cg *CrepGroup) Remaining(it Item) int {
	n := 0
	for _, w := range cg.Wants {
		if w == it {
			n++
		}
	}
	return n
}

func (cg *CrepGroup) receive(it Item) bool {
	for i, w := range cg.Wants {
		if w == it {
			cg.Wants = append(cg.Wants[:i], cg.Wants[i+1:]...)
			cg.Got = append(cg.Got, it)
			return true
		}
	}
	return false
}

// Center : le milieu du groupe (bulles, effets).
func (cg *CrepGroup) Center() V {
	if cg.Table != nil && cg.State >= cgSeated && cg.State <= cgEat {
		return cg.Table.Pos
	}
	if len(cg.Members) > 0 {
		return cg.Members[0].Pos
	}
	return V{}
}

// waiter : ce que fait un serveur.
type waiter struct {
	A     *Agent
	task  int // 0 rien, 1 commande, 2 service, 3 nettoyage
	table *CrepTable
	stops []crepStop
	path  []V
	t     float64
}

type crepStop struct {
	from *Station // étagère du comptoir ou frigo (palette du cidre)
	it   Item
}

// Creperie : l'état propre au restaurant.
type Creperie struct {
	Tables  []*CrepTable
	Groups  []*CrepGroup
	Waiters []*waiter
	claims  map[Item]int // plats déjà promis à un serveur
}

// CrepRecipe : les postes de la cuisine, dans l'ordre d'achat (salées et sucrées mêlées).
var crepRecipes = []recipe{
	{"crepe_nature", "Crêpes nature", CrepeNature, []Item{Lait, Oeuf, Beurre}, 3, 20000, "Le chef fait des crêpes nature sans arrêt (lait, œufs, beurre) : toutes les crêpes garnies en ont besoin.", "cremerie:beurre"},
	{"crepe_sucre", "Crêpes au sucre", CrepeSucre, []Item{CrepeNature, SucrePoudre}, 4, 30000, "Une crêpe et du sucre en poudre (ferme exotique).", "exotique:moulin_sucre"},
	{"crepe_saucisse", "Crêpes saucisse-œufs", CrepeSaucisseOeuf, []Item{CrepeNature, Saucisse, Oeuf}, 6, 45000, "Une crêpe, une saucisse et deux œufs (un sourire !).", "boucherie:saucisse"},
	{"crepe_pomme", "Crêpes pomme-cannelle", CrepePommeCannelle, []Item{CrepeNature, Pomme, Cannelle}, 5, 60000, "Une crêpe, des pommes et de la cannelle.", "ferme:pommier,exotique:cannelle"},
	{"crepe_thon", "Crêpes thon-poireaux", CrepeThonPoireau, []Item{CrepeNature, Thon, Poireau}, 6, 80000, "Une crêpe, du thon (port) et des poireaux.", "port:thon,ferme:poireau"},
	{"crepe_banane", "Crêpes banane-chocolat", CrepeBananeChoco, []Item{CrepeNature, Banane, Chocolat}, 5, 100000, "Une crêpe, des bananes et du chocolat.", "exotique:bananier,exotique:chocolaterie"},
	{"crepe_steak", "Crêpes steak haché-brebis", CrepeSteakBrebis, []Item{CrepeNature, SteakHache, FromageBrebis}, 7, 130000, "Une crêpe, un steak haché et du fromage de brebis.", "boucherie:hache,cremerie:fromage_brebis"},
	{"crepe_poire", "Crêpes poire-glace vanille", CrepePoireVanille, []Item{CrepeNature, Poire, GlaceVanille}, 6, 170000, "Une crêpe, des poires et de la glace à la vanille (glacier).", "ferme:poirier,glacier:glace_vanille"},
	{"crepe_canard", "Crêpes au canard flambé", CrepeCanard, []Item{CrepeNature, Magret, Whisky}, 8, 220000, "Une crêpe, un magret flambé au whisky.", "boucherie:magret,domaine:whisky"},
}

// CrepInputs : la réserve (une palette par ingrédient) ; le cidre et la salade ont leur frigo
// en salle (les serveurs s'y servent : une salade par client).
var CrepInputs = []Item{Lait, Oeuf, Beurre, SucrePoudre, Chou, Saucisse, Pomme, Cannelle, Thon, Poireau, Banane, Chocolat, SteakHache, FromageBrebis, Poire, GlaceVanille, Magret, Whisky}

var crepSpec = workshopSpec{
	id: "creperie", name: "La crêperie", price: 30000000, seed: 89,
	inputs: CrepInputs, recipes: crepRecipes,
	staff: "Commis", staffCosts: []float64{15000, 35000, 80000, 180000, 400000, 850000, 1800000, 3800000, 8000000}, upBase: 20000, // (un par poste)
	sign: "Comptoir", bio: "Fait maison", w: CrepW, perRow: 5, docks: []string{},
}

// tables : 6 de 2, puis 6 de 4 (salle), puis la terrasse : 4 de 2 et 4 de 4.
var crepTableCosts = []float64{8000, 10000, 12000, 15000, 18000, 22000, 30000, 36000, 43000, 52000, 62000, 75000,
	40000, 48000, 57000, 68000, 85000, 100000, 120000, 145000}

const crepTerraceCost = 120000

// CrepCiderMult : le prix d'une bouteille de cidre au restaurant (× son prix au domaine).
const CrepCiderMult = 5.0

// CrepSaladMult : le prix d'une salade au restaurant (× son prix à la ferme).
const CrepSaladMult = 3.0

// crepFridge : ce que les serveurs prennent au frigo de la salle (pas au comptoir).
func crepFridge(it Item) bool { return it == Cidre || it == Chou }

func newCreperie(g *Game) *Place {
	p := newWorkshop(g, crepSpec)
	p.Top = CrepTop
	p.Rep = 0
	p.Crep = &Creperie{claims: map[Item]int{}}
	// pas de caisse, pas de vitrine pour la crêpe nature, pas de drones, pas de quai
	var keep []*Station
	for _, s := range p.Stations {
		if s.Kind == Register || s.Kind == Dock || s.ID == "etal_"+itoa(int(CrepeNature)) {
			continue
		}
		keep = append(keep, s)
	}
	p.Stations = keep
	var spots []*Spot
	for _, sp := range p.Spots {
		if sp.Shelf != nil || sp.ID == "caissiere" || sp.ID == "etals" {
			continue
		}
		spots = append(spots, sp)
	}
	p.Spots = spots
	var ups []*Upgrade
	for _, u := range p.ups {
		if strings.HasSuffix(u.Key, ".caissiere") || strings.HasSuffix(u.Key, ".etals") || strings.Contains(u.Key, ".drone") || strings.Contains(u.Key, ".pts_dr") {
			continue
		}
		ups = append(ups, u)
	}
	p.ups = ups
	// la réserve : en haut de la cuisine
	i := 0
	for _, s := range p.Stations {
		if s.Kind == Pallet && s.Item != Chou {
			x := 320 + float64(i)*172
			s.Body = Rect{x, CrepPalY, 150, 90}
			s.Zone = Rect{x - 5, CrepPalY + 110, 160, 90}
			i++
		}
	}
	// les postes : la crêpe nature et les sucrées en haut, les salées en dessous
	xa, xb := 320.0, 520.0
	for _, r := range crepRecipes {
		s := p.Station(r.id)
		mw := machineW(r)
		y, x := CrepRowA, xa
		if isSalee(r.out) {
			y, x = CrepRowB, xb
			xb += mw + stockBoxW + 260
		} else {
			xa += mw + stockBoxW + 230
		}
		s.Body = Rect{x, y, mw, 150}
		s.Zone = Rect{x - 10, y + 168, mw - 130, 90}
		s.OutZone = Rect{x + mw - 120, y + 168, 130, 90}
		if r.out != CrepeNature {
			s.PileMax, s.BufMax = 1, 3 // (sur commande : le chef attend qu'on emporte son plat)
		} else {
			s.PileMax, s.BufMax = 12, 6 // (les crêpes nature s'empilent : « Stock + »)
		}
		for _, u := range p.Unlocks {
			if u.ID == r.id {
				c := s.Body.C()
				u.Zone = Rect{c.X - 90, c.Y - 65, 180, 130}
			}
		}
	}
	// le comptoir : une étagère par crêpe garnie, dans le mur
	k := 0
	for _, s := range p.Stations {
		if s.Kind == Shelf {
			x := 640 + float64(k)*250
			s.Body = Rect{x, CrepWallY - 4, 160, 78}
			s.Zone = Rect{x - 10, CrepWallY - 100, 180, 92} // (côté cuisine : les commis y posent les plats)
			s.Spot = V{x + 80, CrepWallY + 120}
			s.Name = "Comptoir : " + Items[s.Item].Name
			s.NoSale = true
			s.StockMax = 6
			k++
		}
	}
	// les murs (la porte de la cuisine, à gauche du comptoir ; celle de la salle, à gauche
	// aussi ; la porte des clients en bas ; celle de la terrasse à droite)
	wall := func(id string, r Rect) {
		w := &Station{ID: "mur_" + id, Name: "Mur", Kind: Trash, Built: true}
		w.Body = r
		w.Zone = Rect{-1000, -1000, 0, 0}
		p.Stations = append(p.Stations, w)
	}
	// mur du milieu : de la porte de la cuisine (250 à 430) au mur de droite, sauf le comptoir
	x := 430.0
	var shelves []*Station
	for _, s := range p.Stations {
		if s.Kind == Shelf {
			shelves = append(shelves, s)
		}
	}
	sort.Slice(shelves, func(a, b int) bool { return shelves[a].Body.X < shelves[b].Body.X })
	for n, s := range shelves {
		wall("milieu"+itoa(n), Rect{x, CrepWallY, s.Body.X - x, CrepWallH})
		x = s.Body.X + s.Body.W
	}
	wall("milieu_fin", Rect{x, CrepWallY, CrepBX1 - x, CrepWallH})
	wall("milieu_debut", Rect{CrepBX0, CrepWallY, 250 - CrepBX0, CrepWallH})
	wall("gauche_haut", Rect{CrepBX0 - 30, CrepKY0, 30, 1250 - CrepKY0})
	wall("gauche_bas", Rect{CrepBX0 - 30, 1420, 30, CrepBotY - 1420}) // (la porte de la salle : de 1250 à 1420)
	wall("droite_haut", Rect{CrepBX1, CrepKY0, 30, CrepTerDY - 90 - CrepKY0})
	wall("droite_bas", Rect{CrepBX1, CrepTerDY + 90, 30, CrepBotY - CrepTerDY - 90})
	wall("haut", Rect{CrepBX0 - 30, CrepKY0 - 30, CrepBX1 - CrepBX0 + 60, 30})
	wall("bas_g", Rect{CrepBX0 - 30, CrepBotY, CrepDoorX - 90 - CrepBX0 + 30, 26})
	wall("bas_d", Rect{CrepDoorX + 90, CrepBotY, CrepBX1 + 30 - CrepDoorX - 90, 26})

	// (la caisse des commandes, des crêpes nature pour la cantine et le parc, est dans la zone du
	// joueur, comme partout)
	// le compost, en cuisine (en bas à droite, près du coin des commis)
	if tr := p.Station("poubelle"); tr != nil {
		tr.Body = Rect{3360, -230, 70, 70}
		tr.Zone = Rect{3340, -270, 110, 150}
	}
	// le frigo du cidre, en salle à droite du comptoir (les serveurs s'y servent)
	fr := &Station{ID: "palette_" + itoa(int(Cidre)), Name: "Frigo : cidre", Kind: Pallet, Item: Cidre, PileMax: 40, Built: true}
	fr.Body = Rect{3050, CrepWallY + CrepWallH, 200, 80}
	fr.Zone = Rect{3040, CrepWallY + CrepWallH + 90, 220, 90}
	p.Stations = append(p.Stations, fr)
	// le frigo des salades, à sa gauche (une salade par client, apportée par les serveurs)
	if sa := p.pallet(Chou); sa != nil {
		sa.Name = "Frigo : salades"
		sa.Body = Rect{2790, CrepWallY + CrepWallH, 200, 80}
		sa.Zone = Rect{2780, CrepWallY + CrepWallH + 90, 220, 90}
	}

	// les tables
	cp := p.Crep
	add := func(id string, seats int, pos V, terrace bool, cost float64, req func() bool) {
		st := &Station{ID: "table_" + id, Name: "Table", Kind: Trash, Unlock: "table_" + id}
		t := &CrepTable{ID: id, Seats: seats, Pos: pos, Terrace: terrace, St: st}
		st.Body = t.Body()
		st.Zone = Rect{-1000, -1000, 0, 0}
		p.Stations = append(p.Stations, st)
		cp.Tables = append(cp.Tables, t)
		name := "Table de 2"
		if seats == 4 {
			name = "Table de 4"
		}
		if terrace {
			name += " (terrasse)"
		}
		b := t.Body()
		p.Unlocks = append(p.Unlocks, &Unlock{ID: "table_" + id, Name: name, Cost: cost, Zone: Rect{b.C().X - 90, b.C().Y - 65, 180, 130},
			Hint: "Les clients s'y assoient seuls ; un serveur prend leur commande.", Requires: req, Secret: true})
	}
	done := func(id string) func() bool {
		return func() bool { u := p.unlock(id); return u != nil && u.Done }
	}
	prev := func(id string, extra func() bool) func() bool {
		return func() bool { return done(id)() && (extra == nil || extra()) }
	}
	cols := []float64{560, 1300, 2040, 2780}
	rows := []float64{470, 820, 1170}
	n := 0
	for r, y := range rows {
		for _, cx := range cols {
			seats := 2
			if n >= 6 {
				seats = 4
			}
			id := itoa(n + 1)
			var req func() bool
			if n > 0 {
				req = prev("table_"+itoa(n), nil)
			}
			add(id, seats, V{cx, y}, false, crepTableCosts[n], req)
			n++
		}
		_ = r
	}
	// la terrasse (dehors, plus loin : les serveurs passent par la porte de droite)
	terU := &Unlock{ID: "terrasse", Name: "Terrasse", Cost: crepTerraceCost, Zone: Rect{(CrepTerX0+CrepTerX1)/2 - 90, 820, 180, 130},
		Hint: "Une terrasse, dehors : des tables de 2 et de 4 à acheter (aucune n'est offerte).", Requires: done("table_12"), Secret: true}
	p.Unlocks = append(p.Unlocks, terU)
	for k := 0; k < 8; k++ {
		seats := 2
		if k >= 4 {
			seats = 4
		}
		x := 4020.0
		if k%2 == 1 {
			x = 4740
		}
		y := 440 + float64(k/2)*290
		req := done("terrasse")
		if k > 0 {
			req = prev("table_"+itoa(12+k), nil)
		}
		add(itoa(13+k), seats, V{x, y}, true, crepTableCosts[12+k], req)
	}
	// postes : dans l'ordre, chacun après le précédent (et ses ingrédients ailleurs)
	for i, r := range crepRecipes {
		u := p.unlock(r.id)
		if u == nil {
			continue
		}
		need, prevID := r.needs, ""
		if i > 0 {
			prevID = crepRecipes[i-1].id
		}
		u.Requires = func() bool {
			if prevID != "" && !done(prevID)() {
				return false
			}
			return need == "" || g.needsMet(need)
		}
		if need != "" {
			u.Locked = needsText(need)
		}
	}
	// améliorations : serveurs (autant que de tables), parking (5 places à la fois)
	kk := p.key
	var wc []float64
	for i := 0; i < len(cp.Tables); i++ {
		wc = append(wc, math.Round(12000*math.Pow(1.45, float64(i))/100)*100)
	}
	p.ups = append(p.ups,
		&Upgrade{Key: kk("serveurs"), Name: "Serveurs", Max: len(cp.Tables), Now: true, Costs: wc,
			Desc: func(l int) string {
				return "Ils prennent les commandes, apportent crêpes et cidre, et nettoient les tables. Embauchés : " + itoa(l) + "."
			}},
		&Upgrade{Key: kk("parking"), Name: "Parking +5", Max: (len(cp.Tables) + 5 - crepParkN0) / crepParkUp, Base: 10000, Mult: 1.9,
			Desc: func(l int) string { return itoa(p.CrepParkN(l)) + " places de parking." }},
	)
	p.Spots = append(p.Spots,
		&Spot{ID: "serveurs", Label: "Serveur +1", Zone: Rect{180, 1250, 140, 124}, Done: func(l int) string { return "Nouveau serveur : +1 !" }},
		&Spot{ID: "parking", Label: "Parking +5", Zone: Rect{180, 1080, 140, 124}, Done: func(l int) string { return itoa(p.CrepParkN(l)) + " places de parking !" }},
	)
	if sp := p.spot("ouvriers"); sp != nil {
		sp.Label = "Commis +1"
		sp.Zone = Rect{180, -150, 140, 124} // (en cuisine, en bas à gauche)
	}
	if sp := p.spot("entrepot"); sp != nil {
		sp.Zone = Rect{150, CrepPalY, 130, 124}
	}
	// le parking : seulement des voitures, devant la porte
	p.Gate = V{CrepDoorX, roadY + 40}
	p.Inside = V{3100, -60} // (les commis attendent en cuisine)
	p.RestSpot = V{3100, -60}
	// (le joueur arrive dans sa zone, à gauche, comme partout : pas sur l'embauche des serveurs)
	p.Parks = nil
	p.parkHalf = 760
	x0, x1, gx := p.ParkX()
	for xx := x0 + 140; xx <= x1-130; xx += 80 {
		if math.Abs(xx-gx) < 80 {
			continue
		}
		p.Parks = append(p.Parks, &PSpot{Pos: V{xx, ParkRow1}}, &PSpot{Pos: V{xx, ParkRow2}})
	}
	sort.SliceStable(p.Parks, func(a, b int) bool {
		return math.Abs(p.Parks[a].Pos.X-gx) < math.Abs(p.Parks[b].Pos.X-gx)
	})
	for _, s := range p.Stations {
		if s.Unlock == "" {
			s.Built = true
		}
	}
	return p
}

func isSalee(it Item) bool {
	for _, s := range CrepesSalees {
		if s == it {
			return true
		}
	}
	return false
}

// unlock : le déblocage de cet identifiant.
func (p *Place) unlock(id string) *Unlock {
	for _, u := range p.Unlocks {
		if u.ID == id {
			return u
		}
	}
	return nil
}

// CrepParkN : places de parking à ce niveau d'agrandissement.
func (p *Place) CrepParkN(l int) int {
	return min(len(p.Parks), crepParkN0+crepParkUp*l)
}

// CrepTablesBuilt : les tables achetées.
func (p *Place) CrepTablesBuilt() []*CrepTable {
	var out []*CrepTable
	for _, t := range p.Crep.Tables {
		if t.St.Built {
			out = append(out, t)
		}
	}
	return out
}

// crepPostes : les postes de cuisine achetés.
func (p *Place) crepPostes() int {
	n := 0
	for _, r := range crepRecipes {
		if s := p.Station(r.id); s != nil && s.Built {
			n++
		}
	}
	return n
}

// crepMade : la crêpe garnie se fait-elle ici (poste acheté) ?
func (p *Place) crepMade(it Item) bool {
	for _, r := range crepRecipes {
		if r.out == it {
			s := p.Station(r.id)
			return s != nil && s.Built
		}
	}
	return false
}

// crepDemand : combien de ce plat les tables attendent encore, moins ce qui est
// déjà prêt (comptoir, poste, bras des commis, du joueur et des serveurs).
func (p *Place) crepDemand(it Item) int {
	n := 0
	for _, cg := range p.Crep.Groups {
		if cg.State == cgWaitFood {
			n += cg.Remaining(it)
		}
	}
	if sh := p.shelfFor(it); sh != nil {
		n -= sh.Stock
	}
	for _, r := range crepRecipes {
		if r.out == it {
			if s := p.Station(r.id); s != nil {
				n -= s.Pile
			}
		}
	}
	for _, w := range p.Workers {
		n -= w.count(it)
	}
	for _, w := range p.Crep.Waiters {
		n -= w.A.count(it)
	}
	if pl := p.g.Player; pl != nil && p.g.Cur == p {
		n -= pl.count(it)
	}
	return n
}

// crepCapacity : combien de tables de clients à la fois (une par table ; plus avec la réclame).
func (p *Place) crepCapacity() int {
	n := len(p.CrepTablesBuilt())
	return int(math.Round(float64(n) * (1 + p.ptsF("pts_reclame"))))
}

// CrepTipRate : part de l'addition laissée en pourboire (1 %, +1 % tous les 5 points « Pourboires »).
func (p *Place) CrepTipRate() float64 {
	return 0.01 + float64(p.g.Levels[p.key("pts_pourboire")])/500
}

// ---------------------------------------------------------------- mise à jour

func (p *Place) stepCreperie(dt float64, active bool) {
	cp := p.Crep
	g := p.g
	p.crepWaiters()
	// une nouvelle voiture
	p.spawnT -= dt
	if p.spawnT <= 0 {
		p.spawnT = (4 + 4*p.rand()) / (1 + p.ptsF("pts_reclame"))
		alive := 0
		for _, cg := range cp.Groups {
			if cg.State != cgGone {
				alive++
			}
		}
		if alive < p.crepCapacity() && p.crepMenuOK() {
			p.crepSpawn()
		}
	}
	for _, cg := range cp.Groups {
		p.stepCrepGroup(cg, dt)
	}
	alive := cp.Groups[:0]
	for _, cg := range cp.Groups {
		if cg.State != cgGone {
			alive = append(alive, cg)
		}
	}
	cp.Groups = alive
	for _, w := range cp.Waiters {
		p.stepWaiter(w, dt)
	}
	if active && g.Player != nil {
		p.crepPlayer(g.Player, dt)
	}
}

// crepMenuOK : on vient manger s'il y a une table et au moins une crêpe garnie.
func (p *Place) crepMenuOK() bool {
	if len(p.CrepTablesBuilt()) == 0 {
		return false
	}
	for _, it := range append(append([]Item{}, CrepesSucrees...), CrepesSalees...) {
		if p.crepMade(it) {
			return true
		}
	}
	return false
}

// crepSpawn : une voiture arrive (2 clients, ou 4 s'il y a des tables de 4).
func (p *Place) crepSpawn() {
	cp := p.Crep
	g := p.g
	free := []*PSpot{}
	for i, s := range p.Parks {
		if i >= p.CrepParkN(g.Levels[p.key("parking")]) {
			break
		}
		if !s.used {
			free = append(free, s)
		}
	}
	if len(free) == 0 {
		return
	}
	t2, t4 := 0, 0
	for _, t := range p.CrepTablesBuilt() {
		if t.Seats == 4 {
			t4++
		} else {
			t2++
		}
	}
	n := 2
	if t4 > 0 && p.rand() < float64(t4)/float64(t2+t4) {
		n = 4
	}
	s := free[p.rng.Intn(len(free))]
	s.used = true
	cg := &CrepGroup{N: n, Happy: 1, State: cgCar}
	look := p.rng.Intn(1 << 20)
	v := &Vehicle{Kind: VCar, Look: look, Rider: true, spot: s}
	v.speed = 300
	entry := p.ParkEntry()
	if p.rng.Intn(2) == 0 {
		v.Pos, v.path = V{-250, laneE}, []V{{entry, laneE}}
	} else {
		v.Pos, v.path = V{p.W + 250, laneW}, []V{{entry, laneW}}
	}
	v.path = append(v.path, V{entry, ParkAisle}, V{s.Pos.X, ParkAisle}, s.Pos)
	v.Dir = v.path[0].Sub(v.Pos).Norm()
	p.Vehicles = append(p.Vehicles, v)
	cg.veh = v
	for i := 0; i < n; i++ {
		cg.Members = append(cg.Members, &Diner{Pos: s.Pos, Face: V{0, -1}, Look: look + i*7919})
	}
	// le menu, choisi avant d'entrer : une salée et une sucrée chacun (le plat préféré,
	// ou un autre s'il n'est pas à la carte : un peu moins de bonheur), du cidre
	pick := func(list []Item) (Item, bool) {
		fav := list[p.rng.Intn(len(list))]
		if p.crepMade(fav) {
			return fav, true
		}
		var ok []Item
		for _, it := range list {
			if p.crepMade(it) {
				ok = append(ok, it)
			}
		}
		cg.Happy -= 0.15
		if len(ok) == 0 {
			return 0, false
		}
		return ok[p.rng.Intn(len(ok))], true
	}
	for i := 0; i < n; i++ {
		if it, ok := pick(CrepesSalees); ok {
			cg.Order = append(cg.Order, it)
		}
		if it, ok := pick(CrepesSucrees); ok {
			cg.Order = append(cg.Order, it)
		}
	}
	if g.needsMet("domaine:cidre") {
		for i := 0; i < n/2; i++ {
			cg.Order = append(cg.Order, Cidre)
		}
	} else {
		cg.Happy -= 0.15 // (pas de cidre à la carte)
	}
	if g.needsMet("ferme:chou") { // une salade pour chacun (du frigo, les serveurs l'apportent)
		for i := 0; i < n; i++ {
			cg.Order = append(cg.Order, Chou)
		}
	}
	cg.Happy = math.Max(0, cg.Happy)
	cp.Groups = append(cp.Groups, cg)
}

// les chemins (sans obstacle : on passe par les allées et les portes)
var (
	crepDoorOut = V{CrepDoorX, CrepBotY + 50}
	crepDoorIn  = V{CrepDoorX, CrepBotY - 60}
	crepTerOut  = V{CrepBX1 + 110, CrepTerDY}
	crepTerIn   = V{CrepBX1 - 90, CrepTerDY}
	crepTerGate = V{CrepTerX0 - 60, CrepBotY - 40}
)

// tableRoute : de la porte des clients jusqu'à la chaise (salle ou terrasse).
func crepRouteIn(t *CrepTable, seat V) []V {
	aisle := t.ServeAt().Y + 40
	if t.Terrace {
		return []V{{crepTerGate.X, crepDoorOut.Y}, crepTerGate, {crepTerGate.X, aisle}, {seat.X, aisle}, seat}
	}
	return []V{crepDoorIn, {crepDoorIn.X, aisle}, {seat.X, aisle}, seat}
}

// crepWalk : d'un point de la salle à un autre (par la porte de la terrasse si besoin).
func crepWalk(from, to V) []V {
	inTer := func(q V) bool { return q.X > CrepBX1 }
	switch {
	case inTer(from) && !inTer(to):
		return []V{{from.X, CrepTerDY}, crepTerOut, crepTerIn, to}
	case !inTer(from) && inTer(to):
		return []V{crepTerIn, crepTerOut, {to.X, CrepTerDY}, to}
	}
	return []V{to}
}

func (d *Diner) walk(speed, dt float64) bool {
	d.Moving = false
	for len(d.path) > 0 {
		to := d.path[0]
		v := to.Sub(d.Pos)
		if v.Len() < 2 {
			d.Pos = to
			d.path = d.path[1:]
			continue
		}
		step := speed * dt
		if step >= v.Len() {
			d.Pos = to
			d.path = d.path[1:]
		} else {
			d.Pos = d.Pos.Add(v.Norm().Mul(step))
		}
		d.Face = v.Norm()
		d.Moving = true
		d.WalkT += dt
		return false
	}
	return true
}

// crepFreeTable : une table propre et libre pour ce groupe (de 2 de préférence pour 2).
func (p *Place) crepFreeTable(n int) *CrepTable {
	var best *CrepTable
	for _, t := range p.CrepTablesBuilt() {
		if t.Group != nil || t.Dirty || t.Seats < n {
			continue
		}
		if best == nil || (t.Seats < best.Seats) || (t.Seats == best.Seats && !t.Terrace && best.Terrace) {
			best = t
		}
	}
	return best
}

func (p *Place) stepCrepGroup(cg *CrepGroup, dt float64) {
	cg.T += dt
	const speed = 160.0
	switch cg.State {
	case cgCar:
		if cg.veh != nil && cg.veh.State == 1 {
			cg.State, cg.T = cgWalk, 0
			for i, m := range cg.Members {
				m.Pos = cg.veh.spot.Pos.Add(V{float64(i%2)*30 - 15, 0})
				m.path = append(p.walkFromCar(cg.veh.spot), crepDoorOut.Add(V{float64(i)*26 - 40, 0}))
			}
		}
	case cgWalk:
		all := true
		for _, m := range cg.Members {
			if !m.walk(speed, dt) {
				all = false
			}
		}
		if all {
			cg.State, cg.T = cgQueue, 0
		}
	case cgQueue:
		for _, m := range cg.Members {
			m.walk(speed, dt)
		}
		if t := p.crepFreeTable(cg.N); t != nil && p.crepFirstInLine(cg) {
			t.Group = cg
			cg.Table = t
			cg.State, cg.T = cgToTable, 0
			for i, m := range cg.Members {
				m.path = crepRouteIn(t, t.Seat(i))
			}
			return
		}
		// on attend devant la porte (en file, le long du trottoir)
		k := p.crepQueueIndex(cg)
		for i, m := range cg.Members {
			to := V{CrepDoorX + 150 + float64(k)*110 + float64(i%2)*36, CrepBotY + 50 + float64(i/2)*30}
			if m.Pos.Dist(to) > 4 && len(m.path) == 0 {
				m.path = []V{to}
			}
		}
		if cg.T > 90 { // trop long : on repart, fâché
			cg.Happy, cg.Angry = 0, true
			p.crepLeave(cg)
		}
	case cgToTable:
		all := true
		for _, m := range cg.Members {
			if !m.walk(speed, dt) {
				all = false
			}
		}
		if all {
			cg.State, cg.T = cgSeated, 0
			for i, m := range cg.Members {
				if i%2 == 0 {
					m.Face = V{1, 0}
				} else {
					m.Face = V{-1, 0}
				}
			}
		}
	case cgSeated, cgWaitFood:
		cg.wait += dt
		// l'attente use la patience (au-delà d'un moment, la jauge baisse)
		calm := 25.0
		if cg.State == cgWaitFood {
			calm = 60
		}
		if cg.T > calm {
			cg.Happy = math.Max(0, cg.Happy-dt*0.004*math.Max(0, 1-p.ptsF("pts_bonheur")/2))
		}
		if cg.State == cgWaitFood && len(cg.Wants) == 0 {
			cg.State, cg.T = cgEat, 0
			return
		}
		if cg.wait > 300 { // bien trop long : on paie ce qu'on a reçu et on part, fâché
			cg.Angry = true
			cg.Happy = math.Max(0, cg.Happy-0.3)
			p.crepPay(cg)
			p.crepLeave(cg)
		}
	case cgEat:
		if cg.T > 20 {
			p.crepPay(cg)
			p.crepLeave(cg)
		}
	case cgLeave:
		all := true
		for _, m := range cg.Members {
			if !m.walk(speed, dt) {
				all = false
			}
		}
		if all {
			cg.State = cgGone
			if cg.veh != nil {
				p.depart(cg.veh)
			}
		}
	}
}

// crepFirstInLine : les premiers arrivés devant la porte entrent d'abord (une table de 4
// laisse passer une table de 2 s'il n'y a de place que pour elle).
func (p *Place) crepFirstInLine(cg *CrepGroup) bool {
	for _, o := range p.Crep.Groups {
		if o == cg {
			return true
		}
		if o.State == cgQueue && o.N <= cg.N && p.crepFreeTable(o.N) != nil {
			return false
		}
	}
	return true
}

func (p *Place) crepQueueIndex(cg *CrepGroup) int {
	k := 0
	for _, o := range p.Crep.Groups {
		if o == cg {
			return k
		}
		if o.State == cgQueue {
			k++
		}
	}
	return k
}

// crepPay : la table paie ce qu'elle a reçu ; un pourboire selon sa jauge de bonheur.
func (p *Place) crepPay(cg *CrepGroup) {
	g := p.g
	if len(cg.Got) == 0 {
		return
	}
	total := 0.0
	for _, it := range cg.Got {
		pr := Items[it].Price
		if it == Cidre {
			pr *= CrepCiderMult // (au restaurant, la bouteille est bien plus chère qu'au domaine)
		}
		if it == Chou {
			pr *= CrepSaladMult // (la salade, assaisonnée, au restaurant)
		}
		total += pr
	}
	total *= g.priceMult(p) * p.SellMult
	at := cg.Center()
	g.earn(total, at, p.ID)
	cg.Got = nil
	if len(cg.Wants) > 0 || cg.Angry {
		g.emit(Event{Kind: EvAngry, Pos: at, Place: p.ID})
		g.Stats.Angry++
	} else {
		g.Stats.Happy++
	}
	g.Stats.Served += cg.N
	p.Served += cg.N
	if cg.Table != nil && p.rand() < cg.Happy {
		cg.Table.Tips += math.Round(total * p.CrepTipRate())
	}
}

// crepLeave : on rend la table (sale) et on retourne à la voiture.
func (p *Place) crepLeave(cg *CrepGroup) {
	if t := cg.Table; t != nil {
		t.Group = nil
		t.Dirty = true
		if t.claim != nil {
			t.claim = nil
		}
	}
	for _, m := range cg.Members {
		var path []V
		if cg.Table != nil && cg.State >= cgSeated {
			t := cg.Table
			aisle := t.ServeAt().Y + 40
			if t.Terrace {
				path = []V{{m.Pos.X, aisle}, {crepTerGate.X, aisle}, crepTerGate, {crepTerGate.X, crepDoorOut.Y}, crepDoorOut}
			} else {
				path = []V{{m.Pos.X, aisle}, {crepDoorIn.X, aisle}, crepDoorIn, crepDoorOut}
			}
		} else {
			path = []V{crepDoorOut}
		}
		if cg.veh != nil && cg.veh.spot != nil {
			path = append(path, p.walkToCar(cg.veh.spot)...)
		}
		m.path = path
	}
	cg.Table = nil
	cg.State, cg.T = cgLeave, 0
	if cg.veh != nil && cg.veh.State == 0 {
		// encore dans la voiture (parking plein…) : elle repart tout de suite
		cg.State = cgGone
	}
}

// ---------------------------------------------------------------- serveurs

func (p *Place) crepWaiters() {
	g := p.g
	cp := p.Crep
	n := g.Levels[p.key("serveurs")]
	for len(cp.Waiters) < n {
		a := &Agent{Kind: WorkerAgent, Pos: p.crepWaiterRest(len(cp.Waiters)), Face: V{0, 1}, Look: 4000 + len(cp.Waiters)*53, Team: "salle"}
		cp.Waiters = append(cp.Waiters, &waiter{A: a})
		g.emit(Event{Kind: EvHire, Pos: a.Pos})
	}
}

func (p *Place) crepWaiterRest(i int) V {
	return V{2400 + float64(i%6)*90, CrepWallY + 200 + float64(i/6)*60}
}

func (p *Place) waiterSpeed() float64 { return 230 * (1 + p.ptsF("pts_serv_vit")) }

// stepWaiter : le serveur choisit une tâche (commande, service, nettoyage) et la fait.
func (p *Place) stepWaiter(w *waiter, dt float64) {
	a := w.A
	cp := p.Crep
	move := func() bool {
		a.Moving = false
		for len(w.path) > 0 {
			to := w.path[0]
			v := to.Sub(a.Pos)
			if v.Len() < 2 {
				a.Pos = to
				w.path = w.path[1:]
				continue
			}
			step := p.waiterSpeed() * dt
			if step >= v.Len() {
				a.Pos = to
				w.path = w.path[1:]
			} else {
				a.Pos = a.Pos.Add(v.Norm().Mul(step))
			}
			a.Face = v.Norm()
			a.Moving = true
			a.WalkT += dt
			return false
		}
		return true
	}
	if w.task == 0 {
		p.waiterPick(w)
		if w.task == 0 {
			rest := p.crepWaiterRest(indexOf(cp.Waiters, w))
			if a.Pos.Dist(rest) > 4 && len(w.path) == 0 {
				w.path = crepWalk(a.Pos, rest)
			}
			move()
			return
		}
	}
	if !move() {
		return
	}
	t := w.table
	switch w.task {
	case 1: // la commande, à la tablette
		cg := t.Group
		if cg == nil || cg.State != cgSeated {
			p.waiterDone(w)
			return
		}
		a.Face = V{0, -1}
		cg.orderT += dt * (1 + p.ptsF("pts_tablette"))
		if cg.orderT >= 6 {
			p.crepOrdered(cg)
			p.waiterDone(w)
		}
	case 2: // le service : aller chercher, puis apporter
		if len(w.stops) > 0 {
			st := w.stops[0]
			w.stops = w.stops[1:]
			cp.claims[st.it]--
			if p.takeFrom(st.from, st.it) {
				a.Carry = append(a.Carry, st.it)
			}
			if len(w.stops) > 0 {
				w.path = crepWalk(a.Pos, crepStopAt(w.stops[0].from))
			} else {
				w.path = crepWalk(a.Pos, t.ServeAt())
			}
			return
		}
		if cg := t.Group; cg != nil && cg.State == cgWaitFood {
			var keep []Item
			for _, it := range a.Carry {
				if !cg.receive(it) {
					keep = append(keep, it)
				}
			}
			a.Carry = keep
		}
		if len(a.Carry) > 0 {
			p.waiterDrop(w) // (la table est partie : on rapporte au comptoir)
		}
		p.waiterDone(w)
	case 3: // le nettoyage
		if !t.Dirty || t.Group != nil {
			p.waiterDone(w)
			return
		}
		a.Face = V{0, -1}
		t.cleanT += dt * (1 + p.ptsF("pts_nettoyage"))
		if t.cleanT >= 5 {
			t.cleanT = 0
			t.Dirty = false
			p.waiterDone(w)
		}
	}
}

func indexOf(ws []*waiter, w *waiter) int {
	for i, x := range ws {
		if x == w {
			return i
		}
	}
	return 0
}

func (p *Place) waiterDone(w *waiter) {
	if w.table != nil && w.table.claim == w.A {
		w.table.claim = nil
	}
	for _, st := range w.stops {
		p.Crep.claims[st.it]--
	}
	w.stops = nil
	w.task, w.table = 0, nil
}

// waiterDrop : des plats sans table (elle est partie) : remis au comptoir (ou au frigo).
func (p *Place) waiterDrop(w *waiter) {
	for _, it := range w.A.Carry {
		if sh := p.shelfFor(it); sh != nil && sh.Stock < sh.StockMax {
			sh.Stock++
		} else if pal := p.pallet(it); pal != nil && pal.Pile < pal.PileMax {
			pal.Pile++
		}
	}
	w.A.Carry = nil
}

func crepStopAt(s *Station) V {
	if s.Kind == Pallet {
		return s.Zone.C()
	}
	return s.Spot
}

// takeFrom : prend un produit au comptoir ou au frigo.
func (p *Place) takeFrom(s *Station, it Item) bool {
	if s.Kind == Pallet {
		if s.Pile > 0 {
			s.Pile--
			return true
		}
		return false
	}
	if s.Stock > 0 {
		s.Stock--
		return true
	}
	return false
}

// crepAvail : combien de ce produit un serveur peut encore prendre (comptoir ou frigo).
func (p *Place) crepAvail(it Item) (int, *Station) {
	if crepFridge(it) {
		pal := p.pallet(it)
		if pal == nil {
			return 0, nil
		}
		return pal.Pile - p.Crep.claims[it], pal
	}
	sh := p.shelfFor(it)
	if sh == nil || !sh.Built {
		return 0, nil
	}
	return sh.Stock - p.Crep.claims[it], sh
}

// waiterPick : la prochaine tâche du serveur. D'abord servir ce qui est prêt (les
// plats refroidissent), puis prendre les commandes, puis nettoyer.
func (p *Place) waiterPick(w *waiter) {
	cp := p.Crep
	a := w.A
	cap := 4 // (un plateau)
	// servir : une table qui attend quelque chose de disponible (la plus ancienne)
	var best *CrepGroup
	for _, cg := range cp.Groups {
		if cg.State != cgWaitFood || cg.Table == nil || cg.Table.claim != nil {
			continue
		}
		for _, it := range cg.Wants {
			if n, _ := p.crepAvail(it); n > 0 {
				if best == nil || cg.wait > best.wait {
					best = cg
				}
				break
			}
		}
	}
	if best != nil {
		t := best.Table
		var stops []crepStop
		taken := map[Item]int{}
		for _, it := range best.Wants {
			if len(stops) >= cap {
				break
			}
			n, from := p.crepAvail(it)
			if n-taken[it] <= 0 {
				continue
			}
			taken[it]++
			cp.claims[it]++
			stops = append(stops, crepStop{from: from, it: it})
		}
		sort.SliceStable(stops, func(i, j int) bool { return stops[i].from.Body.X < stops[j].from.Body.X })
		w.task, w.table, w.stops = 2, t, stops
		t.claim = a
		w.path = crepWalk(a.Pos, crepStopAt(stops[0].from))
		return
	}
	// la commande
	for _, cg := range cp.Groups {
		if cg.State == cgSeated && cg.Table != nil && cg.Table.claim == nil {
			w.task, w.table = 1, cg.Table
			cg.Table.claim = a
			w.path = crepWalk(a.Pos, cg.Table.ServeAt())
			return
		}
	}
	// le nettoyage
	for _, t := range cp.Tables {
		if t.St.Built && t.Dirty && t.Group == nil && t.claim == nil {
			w.task, w.table = 3, t
			t.claim = a
			w.path = crepWalk(a.Pos, t.ServeAt())
			return
		}
	}
}

// crepOrdered : la commande est prise : elle part en cuisine (la tablette).
func (p *Place) crepOrdered(cg *CrepGroup) {
	cg.State, cg.T = cgWaitFood, 0
	cg.Wants = append([]Item{}, cg.Order...)
	p.g.emit(Event{Kind: EvLevel, Pos: cg.Center(), Place: p.ID})
}

// ---------------------------------------------------------------- le joueur en salle

// crepPlayer : le joueur peut tout faire en salle : prendre une commande (arrêté
// près de la table), servir ce qu'il porte, nettoyer, ramasser les pourboires,
// prendre au comptoir (côté salle) et au frigo ce que les tables attendent.
func (p *Place) crepPlayer(pl *Agent, dt float64) {
	g := p.g
	cp := p.Crep
	for _, t := range cp.Tables {
		if !t.St.Built || !t.Zone().Has(pl.Pos) {
			continue
		}
		if t.Tips > 0 {
			g.earn(t.Tips, t.Pos, p.ID)
			g.Stats.Tips += t.Tips
			t.Tips = 0
		}
		cg := t.Group
		if cg != nil && cg.State == cgWaitFood && len(pl.Carry) > 0 {
			var keep []Item
			for _, it := range pl.Carry {
				if !cg.receive(it) {
					keep = append(keep, it)
				}
			}
			pl.Carry = keep
		}
		if pl.Moving {
			continue
		}
		if cg != nil && cg.State == cgSeated && t.claim == nil {
			cg.orderT += dt * (1 + p.ptsF("pts_tablette"))
			if cg.orderT >= 6 {
				p.crepOrdered(cg)
			}
		}
		if cg == nil && t.Dirty && t.claim == nil {
			t.cleanT += dt * (1 + p.ptsF("pts_nettoyage"))
			if t.cleanT >= 5 {
				t.cleanT = 0
				t.Dirty = false
			}
		}
	}
	// le comptoir (côté salle) et le frigo : ce que les tables attendent, pas plus
	if pl.Moving {
		return
	}
	cap := p.carryCap(pl)
	for _, s := range p.Stations {
		if !s.Built || len(pl.Carry) >= cap {
			continue
		}
		var zone Rect
		switch {
		case s.Kind == Shelf:
			zone = Rect{s.Spot.X - 90, s.Spot.Y - 60, 180, 110}
		case s.Kind == Pallet && crepFridge(s.Item):
			zone = s.Zone
		default:
			continue
		}
		if !zone.Has(pl.Pos) {
			continue
		}
		need := 0
		for _, cg := range cp.Groups {
			if cg.State == cgWaitFood {
				need += cg.Remaining(s.Item)
			}
		}
		need -= pl.count(s.Item)
		for _, w := range cp.Waiters {
			need -= w.A.count(s.Item)
		}
		if need > 0 {
			if n, _ := p.crepAvail(s.Item); n > 0 && p.takeFrom(s, s.Item) {
				pl.Carry = append(pl.Carry, s.Item)
			}
		}
	}
}

// CrepNeeds : ce qu'attend la table (pour les bulles).
func (cg *CrepGroup) Needs() []Item { return cg.Wants }

// Waiting : la table attend qu'on prenne sa commande.
func (cg *CrepGroup) Waiting() bool { return cg.State == cgSeated }

// Seated : la table est assise (commande prise ou non, ou en train de manger).
func (cg *CrepGroup) Seated() bool { return cg.State >= cgSeated && cg.State <= cgEat }

// Eating : tout est servi.
func (cg *CrepGroup) Eating() bool { return cg.State == cgEat }

// Outside : visible dehors (en marche, en file).
func (cg *CrepGroup) Visible() bool { return cg.State != cgCar && cg.State != cgGone }

// CrepWaiters : les serveurs (pour le dessin).
func (p *Place) CrepWaiters() []*Agent {
	var out []*Agent
	for _, w := range p.Crep.Waiters {
		out = append(out, w.A)
	}
	return out
}

// CrepOrderProgress : avancement de la prise de commande (0 à 1).
func (cg *CrepGroup) OrderProgress() float64 { return math.Min(1, cg.orderT/6) }

// CrepCleanProgress : avancement du nettoyage (0 à 1).
func (t *CrepTable) CleanProgress() float64 { return math.Min(1, t.cleanT/5) }

// crepPrevDone : le poste précédent de ce déblocage est-il acheté ?
func (p *Place) crepPrevDone(u *Unlock) bool {
	for i, r := range crepRecipes {
		if r.id == u.ID {
			if i == 0 {
				return true
			}
			v := p.unlock(crepRecipes[i-1].id)
			return v != nil && v.Done
		}
	}
	return false
}

// InQueue : la table attend devant la porte (pas de table libre).
func (cg *CrepGroup) InQueue() bool { return cg.State == cgQueue }

// Leaving : la table repart.
func (cg *CrepGroup) Leaving() bool { return cg.State == cgLeave }

// QueueT : depuis combien de temps elle attend dehors.
func (cg *CrepGroup) QueueT() float64 {
	if cg.State == cgQueue {
		return cg.T
	}
	return 0
}

// TerraceOpen : la terrasse est achetée.
func (p *Place) TerraceOpen() bool { u := p.unlock("terrasse"); return u != nil && u.Done }

// crepPointKinds : les points d'école de la crêperie (ses postes, ses serveurs, ses pourboires).
func (p *Place) crepPointKinds(add func(key, t, grp, name, desc string, val func(int) string)) {
	// primaire : les commis, les serveurs, les chefs (crêpes nature, salées, sucrées à part)
	add("pts_ent_vit", "ouvrier", "Commis", "Vitesse de marche", "Les commis marchent plus vite (+1 % par point).", pctVal)
	add("pts_mains", "ouvrier", "Commis", "Mains expertes", "Les commis prennent et posent plus vite (+1 % par point).", pctVal)
	add("pts_ent_cap", "ouvrier", "Commis", "Charge", "Les commis portent plus (+1 objet tous les 5 points).",
		func(n int) string { return "+" + itoa(n/5) + " objet(s)" })
	add("pts_serv_vit", "ouvrier", "Serveurs", "Vitesse", "Les serveurs marchent plus vite (+1 % par point).", pctVal)
	add("pts_tablette", "ouvrier", "Serveurs", "Tablette", "Les commandes se prennent plus vite (+1 % par point).", pctVal)
	add("pts_nettoyage", "ouvrier", "Serveurs", "Nettoyage", "Les tables se nettoient plus vite (+1 % par point).", pctVal)
	add("pts_chef_nat_vit", "ouvrier", "Chef des crêpes nature", "Vitesse", "Les crêpes nature cuisent plus vite (+3 % par point ; elles consomment autant).", func(n int) string { return pctVal(3 * n) })
	add("pts_chef_nat_rent", "ouvrier", "Chef des crêpes nature", "Rentabilité", "Chaque crêpe nature peut être gratuite : elle ne consomme alors ni lait, ni œufs, ni beurre.", rentVal)
	add("pts_chef_sal_vit", "ouvrier", "Chefs des crêpes salées", "Vitesse", "Les crêpes salées se préparent plus vite (+3 % par point ; elles consomment autant).", func(n int) string { return pctVal(3 * n) })
	add("pts_chef_sal_rent", "ouvrier", "Chefs des crêpes salées", "Rentabilité", "Chaque crêpe salée peut être gratuite : elle ne consomme alors aucun ingrédient.", rentVal)
	add("pts_chef_suc_vit", "ouvrier", "Chefs des crêpes sucrées", "Vitesse", "Les crêpes sucrées se préparent plus vite (+3 % par point ; elles consomment autant).", func(n int) string { return pctVal(3 * n) })
	add("pts_chef_suc_rent", "ouvrier", "Chefs des crêpes sucrées", "Rentabilité", "Chaque crêpe sucrée peut être gratuite : elle ne consomme alors aucun ingrédient.", rentVal)
	// collège : plus de clients à la fois ; les absences (comme partout)
	add("pts_reclame", "marketing", "Réclame", "Clients", "Plus de tables de clients à la fois (+1 % par point) : attention à ne pas les faire trop attendre.", pctVal)
	// (la crêperie n'est ouverte que quand on joue : pas de « Clients fidèles », mais des clients plus patients)
	add("pts_bonheur", "marketing", "Clients heureux", "Bonheur", "La jauge de bonheur des clients baisse moins vite (-0,5 % par point).",
		func(n int) string { return "-" + itoa(n/2) + " %" })
	// lycée : les pourboires (sans maximum), les gains, le temps des commandes
	add("pts_pourboire", "comptable", "Pourboires", "Pourboires", "Une table contente laisse plus : +1 % de l'addition tous les 5 points (1 % au départ, sans maximum).",
		func(n int) string { return itoa(1+n/5) + " %" })
	add("pts_comptable", "comptable", "Comptables", "Gains", "Les ventes rapportent plus (+1 % par point).", pctVal)
	add("pts_logist", "comptable", "Logistique", "Commandes", "Plus de temps pour livrer les commandes (+1 % par point).", pctVal)
}

// ---------------------------------------------------------------- sauvegarde

// CrepGroupSave : une table de clients assis (le restaurant fermé, ils sont « en pause »).
type CrepGroupSave struct {
	Table              string
	N, State, Look     int
	T, Wait, Happy, OT float64
	Angry              bool
	Order, Wants, Got  []Item
}

// saveSeated : les clients attablés (ceux qui ne sont pas encore assis repartent).
func (p *Place) saveSeated() []CrepGroupSave {
	var out []CrepGroupSave
	for _, cg := range p.Crep.Groups {
		if cg.Table == nil || cg.State < cgSeated || cg.State > cgEat {
			continue
		}
		look := 0
		if len(cg.Members) > 0 {
			look = cg.Members[0].Look
		}
		out = append(out, CrepGroupSave{Table: cg.Table.ID, N: cg.N, State: cg.State, Look: look, T: cg.T, Wait: cg.wait,
			Happy: cg.Happy, OT: cg.orderT, Angry: cg.Angry, Order: cg.Order, Wants: cg.Wants, Got: cg.Got})
	}
	return out
}

// restoreSeated : à ton retour, les clients reprennent là où ils en étaient (à leur
// table, leur voiture garée au parking).
func (p *Place) restoreSeated(gs []CrepGroupSave) {
	for _, s := range gs {
		var t *CrepTable
		for _, x := range p.Crep.Tables {
			if x.ID == s.Table && x.St.Built && x.Group == nil && !x.Dirty && x.Seats >= s.N {
				t = x
			}
		}
		if t == nil || s.State < cgSeated || s.State > cgEat {
			continue
		}
		cg := &CrepGroup{N: s.N, State: s.State, T: s.T, wait: s.Wait, Happy: s.Happy, orderT: s.OT, Angry: s.Angry,
			Order: s.Order, Wants: s.Wants, Got: s.Got, Table: t}
		for i := 0; i < s.N; i++ {
			face := V{1, 0}
			if i%2 == 1 {
				face = V{-1, 0}
			}
			cg.Members = append(cg.Members, &Diner{Pos: t.Seat(i), Face: face, Look: s.Look + i*7919})
		}
		// leur voiture, garée
		for i, sp := range p.Parks {
			if i >= p.CrepParkN(p.g.Levels[p.key("parking")]) {
				break
			}
			if !sp.used {
				sp.used = true
				v := &Vehicle{Kind: VCar, Look: s.Look, State: 1, Pos: sp.Pos, Dir: V{0, 1}, spot: sp, speed: 300}
				p.Vehicles = append(p.Vehicles, v)
				cg.veh = v
				break
			}
		}
		t.Group = cg
		p.Crep.Groups = append(p.Crep.Groups, cg)
	}
}
