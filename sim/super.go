package sim

import (
	"fmt"
	"math"
	"math/rand"
)

// Le supermarché de Clairbourg : il ne vend que ce que produit la ferme.
// Les produits déposés sur le quai de la ferme partent en camion vers la
// réserve du supermarché (une palette par produit), puis on remplit les rayons.

const superPrice = 7000

// SuperZone : un espace du supermarché (ses rayons, ses palettes dans
// l'entrepôt, ses magasiniers). Fermé au public tant que son lieu n'existe pas.
type SuperZone struct {
	ID, Name string
	Items    []Item
	Cols     int
	Rows     int     // rangées de rayons : 3 au-dessus des caisses, 4 plus loin (plus de place en bas)
	X, W     float64 // position dans le magasin
	Bot      float64 // bas de l'espace (et de sa barrière)
	open     func(g *Game) bool
	Soon     bool // pas encore dans le jeu
	rank     int  // rang pour le prix des magasiniers (0 : sa place dans la liste)
}

var SuperZones = []*SuperZone{
	{ID: "ferme", Name: "Produits de la ferme", Items: []Item{Mais, FarineMais, Oeuf, Farine, Lait, LaitChevre, LaitBrebis, Miel}, Cols: 3,
		open: func(g *Game) bool { return true }},
	{ID: "legumes", Name: "Légumes", Items: Legumes, Cols: 2,
		open: func(g *Game) bool { return g.farm().Station("carotte").Built }},
	{ID: "fruits", Name: "Fruits", Items: append(append(append([]Item{}, Fruits...), ExoFruits...), Raisin, Olive), Cols: 4, // (le raisin du domaine, puis les olives d'un futur lieu)
		open: func(g *Game) bool { return g.farm().Station("fraisier").Built }},
	{ID: "cremerie", Name: "Crèmerie", Items: CremerieSold, Cols: 3, open: ownedFn("cremerie")},
	{ID: "epicerie", Name: "Épicerie", Items: EpicerieSold, Cols: 3, open: ownedFn("exotique")},
	{ID: "alcool", Name: "Boissons", rank: 10, // (juste après l'épicerie ; ses magasiniers gardent leur prix)
		Items: append([]Item{Limonade, JusOrange}, DrinksSold...), Cols: 3, // (limonade et jus d'orange de la ferme exotique en premier ; puis jus, vins, bière, cidre, alcools du domaine)
		open: func(g *Game) bool {
			if q := g.Place("domaine"); q != nil && q.Owned {
				return true
			}
			q := g.Place("exotique") // (les jus de la ferme exotique arrivent souvent avant le domaine)
			return q != nil && q.Owned && (q.Station("pressoir_citron").Built || q.Station("pressoir_orange").Built)
		}},
	{ID: "boulangerie", Name: "Boulangerie", Items: BakerySold, Cols: 3, open: ownedFn("boulangerie")},
	{ID: "glace", Name: "Glaces", Items: IceSold, Cols: 2, open: ownedFn("glacier")},
	{ID: "poisson", Name: "Poissonnerie", Items: PortSold, Cols: 3, open: ownedFn("port")}, // (le port s'ouvre avant la boucherie)
	{ID: "boucherie", Name: "Boucherie-charcuterie", Items: BoucherieSold, Cols: 6, open: ownedFn("boucherie")},
	{ID: "patisserie", Name: "Pâtisserie", Items: PatisserieSold, Cols: 2, open: ownedFn("patisserie")},
	{ID: "huiles", Name: "Huiles", rank: 13, Items: HuilesSold, Cols: 2, open: ownedFn("huilerie")}, // (les six huiles de l'huilerie, après la pâtisserie)
}

func ownedFn(id string) func(g *Game) bool {
	return func(g *Game) bool { q := g.Place(id); return q != nil && q.Owned }
}

// Open : l'espace est-il ouvert au public ?
func (z *SuperZone) Open(g *Game) bool { return !z.Soon && z.open(g) }

// SuperZoneOf : l'espace où se vend ce produit ("" : aucun).
func SuperZoneOf(it Item) string {
	for _, z := range SuperZones {
		for _, x := range z.Items {
			if x == it {
				return z.ID
			}
		}
	}
	return ""
}

func superZone(id string) *SuperZone {
	for _, z := range SuperZones {
		if z.ID == id {
			return z
		}
	}
	return nil
}

// superOrder : tout ce que le supermarché peut vendre (espace par espace).
var superOrder = func() []Item {
	var out []Item
	for _, z := range SuperZones {
		out = append(out, z.Items...)
	}
	return out
}()

// géométrie du supermarché (voir le plan) : à gauche la zone du joueur, puis
// l'entrepôt (en haut) et le magasin, la colonne des caddies, le parking.
const (
	SuperBX0     = 100.0 // mur gauche (la zone du joueur est sous la route, devant les caddies)
	superColW    = 320.0 // (rayon de 180 : un large passage entre deux rayons voisins)
	superGap     = 60.0
	SuperRow0    = 640.0 // première rangée de rayons
	superRowGap  = 220.0
	SuperRegY    = 1330.0 // caisses (on passe derrière les caissières sans entrer dans un espace fermé ; la file serpente dessous)
	SuperFloorY1 = 1604.0 // bas du magasin
	SuperBarY1   = 1170.0 // bas des espaces à trois rangées (au-dessus des caisses)
	SuperBarY4   = 1410.0 // bas des espaces à quatre rangées (plus loin que les caisses)
	SuperRegs    = 12
	superRegStep = 380.0 // les caisses, serrées près de l'entrée (caisse et case de la caissière)
	SuperMaxPark = 300   // places et caddies au maximum
	superStart   = 10    // au départ : 10 places et 10 caddies
)

// SuperMaxClients : jamais plus de clients à la fois, même avec beaucoup de
// rayons et de marketing (une marge sous les 300 places : des voitures
// arrivent et repartent). (Une variable : les tests la font varier.)
var SuperMaxClients = 290

// SuperYardY : haut de la zone du joueur (sous la route du bas).
const SuperYardY = 1850.0

// superStaff : magasiniers pour le badge « Équipe de choc » (un espace en a autant que de rayons).
const superStaff = 4

var (
	SuperBX1   float64 // mur droit du magasin
	SuperCartX float64 // colonne des caddies (bord gauche)
	SuperParkX float64 // parking (bord gauche)
	SuperW     float64
)

func init() {
	x := SuperBX0 + 60
	// les caisses sont serrées près de l'entrée : les espaces au-dessus d'elles gardent trois
	// rangées ; les suivants, plus loin, descendent jusqu'à une quatrième (le magasin est moins long)
	band := 460 + float64(SuperRegs-1)*superRegStep + 120
	for _, z := range SuperZones {
		if x-SuperBX0-60 >= band {
			z.Rows, z.Bot = 4, SuperBarY4
			z.Cols = max(2, (len(z.Items)+3)/4)
		} else {
			z.Rows, z.Bot = 3, SuperBarY1
			z.Cols = max(z.Cols, (len(z.Items)+2)/3) // (trois rayons par colonne : tout tient dans l'espace)
		}
		z.X = x
		z.W = float64(z.Cols)*superColW + 40
		x += z.W + superGap
	}
	SuperBX1 = x
	// en miroir : les premiers espaces (ferme, légumes…) près de l'entrée, à droite
	for _, z := range SuperZones {
		z.X = SuperBX0 + SuperBX1 - z.X - z.W
	}
	SuperCartX = SuperBX1 + 40
	SuperParkX = SuperCartX + 510 // (après la route d'accès, qui va jusqu'à SuperCartX + 480 : elle ne passe plus sur le parking)
	SuperW = SuperParkX + 2600    // (la route de sortie passe entre les places au sol et le parking à étages)
}

func newSuper(g *Game) *Place {
	p := &Place{ID: "super", Name: "Le supermarché", W: SuperW, H: 2300, g: g, rng: rand.New(rand.NewSource(23)),
		Rep: 3, Price: superPrice, SellMult: 1, SpawnBase: 5.5, WantMax: 3, MaxCust: 50, LeaveOff: V{140, 0}, OnlyStocked: true, RepGain: 0.004}
	p.Gate = V{SuperCartX + 150, 1520}
	p.Inside = V{SuperBX1 - 60, 1530}
	p.SpawnL = V{-300, 1700}
	p.SpawnR = V{SuperW + 300, 1700}
	p.RestSpot = V{SuperBX1 - 300, 1250}
	// les clients ne vont jamais dans l'entrepôt
	p.NoCust = []Rect{{SuperBX0, 130, SuperBX1 - SuperBX0, 456 - 130}}
	p.cartSpot = V{SuperCartX + 150, 1340}
	p.setupSuperParking()

	for _, z := range SuperZones {
		// entrepôt : une palette par produit, au-dessus de son espace
		if n := len(z.Items); n > 0 {
			pw := (z.W - 20) / float64(n)
			for i, it := range z.Items {
				x := z.X + 10 + float64(n-1-i)*pw // (le premier produit à droite, côté entrée)
				s := &Station{ID: "palette_" + itoa(int(it)), Name: "Palette : " + Items[it].Name, Kind: Pallet, Item: it, PileMax: 300, Built: true}
				s.Body = Rect{x + 2, 180, pw - 8, 90}
				s.Zone = Rect{x, 290, pw - 4, 90}
				p.Stations = append(p.Stations, s)
			}
		}
		// rayons : trois ou quatre rangées
		for i, it := range z.Items {
			x := z.X + z.W - 270 - float64(i/z.Rows)*superColW // (les premiers rayons à droite, côté entrée ; marges égales)
			y := SuperRow0 + float64(i%z.Rows)*superRowGap
			s := &Station{ID: "rayon_" + itoa(int(it)), Name: "Rayon : " + Items[it].Name, Kind: Shelf, Item: it, StockMax: 30, Unlock: "r_" + itoa(int(it))}
			s.Body = Rect{x, y, 180, 66}
			s.Zone = Rect{x - 10, y - 95, 200, 92}
			s.Spot = V{x + 90, y + 66 + 38}
			p.Stations = append(p.Stations, s)
		}
		// un espace fermé : une barrière (on ne peut pas y entrer)
		bar := &Station{ID: "barriere_" + z.ID, Name: z.Name, Kind: Trash}
		bar.Body = Rect{z.X - 10, 470, z.W + 20, z.Bot - 470}
		bar.Zone = Rect{-1000, -1000, 0, 0}
		p.Stations = append(p.Stations, bar)
	}
	// caisses : une rangée en bas, serrées près de l'entrée
	for i := 0; i < SuperRegs; i++ {
		x := SuperBX1 - 460 - float64(i)*superRegStep // la première près de la sortie
		id, key, unl := "caisse", "caissiere", ""
		if i > 0 {
			id, key, unl = "caisse"+itoa(i+1), "caissiere"+itoa(i+1), "caisse"+itoa(i+1)
		}
		r := &Station{ID: id, Name: "Caisse " + itoa(i+1), Kind: Register, HireKey: key, Unlock: unl}
		r.Body = Rect{x, SuperRegY, 220, 66}
		r.Spot = V{x + 110, SuperRegY - 40}
		r.Zone = Rect{x + 25, SuperRegY - 100, 170, 96}
		r.QueueHead = V{x + 110, SuperRegY + 106}
		r.QueueDown = 2
		r.QueueLane = true
		p.Stations = append(p.Stations, r)
	}

	// déblocages : les rayons (si le produit est fabriqué quelque part) et les caisses
	for _, it := range superOrder {
		it := it
		s := p.Station("rayon_" + itoa(int(it)))
		c := s.Body.C()
		z := superZone(SuperZoneOf(it))
		u := &Unlock{ID: "r_" + itoa(int(it)), Name: "Rayon " + Items[it].Name, Cost: math.Round(Items[it].Price * 150),
			Zone: Rect{c.X - 90, c.Y - 60, 180, 120}, Requires: func() bool { return z.Open(g) && g.produced(it) }, Locked: whereLocked(it)}
		if it == Mais {
			u.Cost = 0
		}
		p.Unlocks = append(p.Unlocks, u)
	}
	for i := 2; i <= SuperRegs; i++ {
		r := p.Station("caisse" + itoa(i))
		c := r.Body.C()
		cost := math.Round(1500*math.Pow(1.7, float64(i-2))/100) * 100
		p.Unlocks = append(p.Unlocks, &Unlock{ID: "caisse" + itoa(i), Name: "Caisse " + itoa(i), Cost: cost, Zone: Rect{c.X - 90, c.Y - 60, 180, 120}})
	}
	free := "r_" + itoa(int(Mais))
	for _, s := range p.Stations {
		if s.Unlock == "" || s.Unlock == free {
			s.Built = true
		}
	}
	for _, u := range p.Unlocks {
		if u.ID == free {
			u.Done = true
		}
	}

	// améliorations
	k := p.key
	p.ups = []*Upgrade{}
	for i := 1; i <= SuperRegs; i++ {
		key, n := "caissiere", i
		if i > 1 {
			key += itoa(i)
		}
		cost := math.Round(600*math.Pow(1.6, float64(i-1))/100) * 100
		p.ups = append(p.ups, &Upgrade{Key: k(key), Name: "Caissière (caisse " + itoa(i) + ")", Max: 1, Costs: []float64{cost},
			Desc: func(int) string { return "Elle encaisse à la caisse " + itoa(n) + "." }})
	}
	for zi, z := range SuperZones {
		if z.Soon {
			continue
		}
		z := z
		r := zi
		if z.rank > 0 {
			r = z.rank
		} else if zi > 5 {
			r = zi - 1 // (les espaces après les boissons gardent le prix de leur ancienne place)
		}
		f := math.Pow(2.2, float64(r))
		var costs []float64
		for l := 0; l < len(z.Items); l++ { // un magasinier par rayon
			c := 1500 * f * math.Pow(3, math.Min(3, float64(l))) * math.Pow(1.15, math.Max(0, float64(l-3)))
			costs = append(costs, math.Round(c/100)*100)
		}
		p.ups = append(p.ups, &Upgrade{Key: k(teamKey(z.ID)), Name: "Magasinier (" + z.Name + ")", Max: len(z.Items), Now: true, Costs: costs,
			Desc: func(l int) string {
				return "Il remplit les rayons de son espace. Embauchés : " + itoa(l) + "."
			}})
	}
	p.ups = append(p.ups,
		&Upgrade{Key: k("etals"), Name: "Rayons plus grands", Max: 6, Base: 2000, Mult: 2.3, Desc: func(l int) string { return "Chaque rayon contient " + itoa(30+4*l) + " produits." }},
		&Upgrade{Key: k("entrepot"), Name: "Entrepôt", Max: 12, Base: 2500, Mult: 1.6, Desc: func(l int) string {
			return "Chaque palette de l'entrepôt contient " + itoa(PalletMax(p.ID, l)) + " produits."
		}},
		&Upgrade{Key: k("parking"), Name: "Places de parking", Max: (SuperMaxPark - superStart) / 10, Base: 600, Mult: 1.17,
			Desc: func(l int) string {
				return itoa(p.ParkPlaces(l)) + " places de parking (un client sans place repart mécontent)."
			}},
		&Upgrade{Key: k("caddies"), Name: "Caddies", Max: (SuperMaxPark - superStart) / 10, Base: 400, Mult: 1.17,
			Desc: func(l int) string {
				return itoa(p.CartsTotalAt(l)) + " caddies (un client sans caddie repart mécontent)."
			}},
	)
	spot := func(id, label string, z Rect, needs string, done func(int) string) {
		p.Spots = append(p.Spots, &Spot{ID: id, Label: label, Zone: z, Needs: needs, Done: done})
	}
	for i := 0; i < SuperRegs; i++ {
		r := p.Registers()
		_ = r
		st := p.Station("caisse")
		if i > 0 {
			st = p.Station("caisse" + itoa(i+1))
		}
		id := "caissiere"
		if i > 0 {
			id = "caissiere" + itoa(i+1)
		}
		n := i + 1
		spot(id, "Caissière "+itoa(n), Rect{st.Body.X + 232, SuperRegY - 104, 130, 120}, "", func(int) string { return "Caissière embauchée à la caisse " + itoa(n) + " !" })
	}
	for _, z := range SuperZones {
		if z.Soon {
			continue
		}
		z := z
		p.Spots = append(p.Spots, &Spot{ID: teamKey(z.ID), Label: "Magasinier", Team: z.ID, Zone: Rect{z.X + z.W - 130, 462, 120, 80},
			Done: func(l int) string {
				return "Nouveau magasinier (" + z.Name + ") : +1 !"
			}})
	}
	spot("parking", "Parking +10", Rect{SuperParkX - 390, 150, 140, 124}, "", func(l int) string { return itoa(p.ParkPlaces(l)) + " places de parking !" })
	spot("caddies", "Caddies +10", Rect{SuperCartX + 80, 470, 140, 124}, "", func(l int) string { return itoa(p.CartsTotalAt(l)) + " caddies !" })
	spot("etals", "Rayons +", Rect{}, "", func(l int) string { return "Rayons agrandis : niveau " + itoa(l) + " !" })
	spot("entrepot", "Entrepôt", Rect{}, "", func(l int) string { return "Entrepôt agrandi : " + itoa(PalletMax(p.ID, l)) + " par palette !" })
	// la zone du joueur : sous la route, à l'horizontale ; on traverse au passage
	// piéton pour entrer au niveau des caddies
	p.setupYardH(SuperCartX+150-760, SuperYardY)
	p.spawnT = 1
	return p
}

// syncBarriers : un espace fermé (lieu pas encore acheté) a sa barrière.
func (p *Place) syncBarriers() {
	if len(p.g.Places) == 0 {
		return // (pendant la création de la partie)
	}
	for _, z := range SuperZones {
		if b := p.Station("barriere_" + z.ID); b != nil {
			b.Built = !z.Open(p.g)
		}
	}
}

// superRest : où les magasiniers d'un espace attendent (en haut de l'espace).
func superRest(team string) V {
	if z := superZone(team); z != nil {
		return V{z.X + z.W/2, 520}
	}
	return V{SuperBX1 - 300, 1250}
}

// supermarket : vrai si une caisse n'est visible qu'une fois la précédente achetée.
func (p *Place) spotHidden(sp *Spot) bool {
	if p.ID != "super" {
		return false
	}
	if len(sp.ID) > len("caissiere") && sp.ID[:len("caissiere")] == "caissiere" {
		return !p.Station("caisse" + sp.ID[len("caissiere"):]).Built
	}
	return false
}

// BuyPlace achète un lieu de la carte.
func (g *Game) BuyPlace(p *Place) bool {
	price := g.PriceOf(p) // (moitié prix pendant le cadeau du maire)
	if p.Owned || g.Money < price {
		return false
	}
	g.Money -= price
	p.Owned = true
	g.emit(Event{Kind: EvUnlock, Text: p.Name})
	return true
}

// PriceMultOf : multiplicateur des prix affichés dans un lieu.
func (g *Game) PriceMultOf(p *Place) float64 { return g.priceMult(p) * p.SellMult }

// produced : ce produit est-il fabriqué quelque part (ferme, boulangerie, glacier) ?
func (g *Game) produced(it Item) bool {
	if g.farm().shelfFor(it) != nil {
		return true
	}
	for _, id := range []string{"exotique", "port", "domaine", "huilerie"} { // (ce qu'on y vend tel quel ; domaine : le raisin, les jus ; huilerie : olives, noisettes, noix)
		if q := g.Place(id); q != nil && q.Owned && q.shelfFor(it) != nil {
			return true
		}
	}
	for _, id := range []string{"boulangerie", "cremerie", "glacier", "boucherie", "domaine", "patisserie", "huilerie"} {
		p := g.Place(id)
		if p == nil || !p.Owned {
			continue
		}
		for _, s := range p.Stations {
			if s.Kind == Machine && s.Built && s.Item == it {
				return true
			}
		}
	}
	return false
}

func whereLocked(it Item) string {
	for _, b := range BakerySold {
		if b == it {
			return "Depuis la boulangerie"
		}
	}
	for _, b := range CremerieSold {
		if b == it {
			return "Depuis la crèmerie"
		}
	}
	for _, b := range IceSold {
		if b == it {
			return "Depuis le glacier"
		}
	}
	for _, b := range BoucherieSold {
		if b == it {
			return "Depuis la boucherie"
		}
	}
	for _, b := range ExoticSold {
		if b == it {
			return "Depuis la ferme exotique"
		}
	}
	for _, b := range PortSold {
		if b == it {
			return "Depuis le port"
		}
	}
	if it == Raisin {
		return "Depuis le domaine"
	}
	if it == Sel {
		return "Depuis le port"
	}
	for _, b := range append(append([]Item{}, HuilesSold...), HuilerieRaw...) {
		if b == it {
			return "Depuis l'huilerie"
		}
	}
	for _, b := range FutureSold {
		if b == it {
			return "Bientôt : un nouveau lieu"
		}
	}
	for _, b := range PatisserieSold {
		if b == it {
			return "Depuis la pâtisserie"
		}
	}
	for _, b := range DrinksSold {
		if b == it {
			return "Depuis le domaine"
		}
	}
	return "Depuis la ferme"
}

// fixRayonCosts : le prix d'un rayon suit la progression du jeu (plus le produit
// arrive tard, plus son rayon coûte cher), au lieu de dépendre de son prix de vente.
func (g *Game) fixRayonCosts() {
	sup := g.Place("super")
	srcCost := func(it Item) (float64, string) {
		for _, p := range g.Places {
			if p.ID == "super" {
				continue
			}
			for _, s := range p.Stations {
				if (s.Kind == Field || s.Kind == Machine || s.Kind == Animal) && s.Item == it {
					for _, u := range p.Unlocks {
						if u.ID == s.Unlock {
							return u.Cost, p.ID
						}
					}
					return 0, p.ID
				}
			}
		}
		return 0, ""
	}
	for _, u := range sup.Unlocks {
		for _, it := range superOrder {
			if u.ID != "r_"+itoa(int(it)) {
				continue
			}
			c, from := srcCost(it)
			switch from {
			case "ferme":
				if it == Mais {
					u.Cost = 0
				} else {
					u.Cost = 800 + 0.5*c
				}
			case "boulangerie":
				u.Cost = 3000 + 0.5*c
			case "cremerie":
				u.Cost = 4000 + 0.5*c
			case "glacier":
				u.Cost = 5000 + 0.5*c
			case "boucherie":
				u.Cost = 7000 + 0.5*c
			case "exotique":
				u.Cost = 2500 + 0.5*c
			case "domaine":
				u.Cost = 9000 + 0.5*c
			case "patisserie":
				u.Cost = 10000 + 0.5*c
			case "huilerie":
				u.Cost = 11000 + 0.5*c
			}
			u.Cost = math.Round(u.Cost/50) * 50
		}
	}
}

// ---------------------------------------------------------------- prix (règle des paliers)
//
// Au supermarché, chaque produit appartient au lieu qui le fabrique : son rayon et le
// magasinier qui va avec suivent le palier de ce lieu. Pour un lieu, ses rayons coûtent
// ensemble superRayons × son prix d'achat, ses magasiniers superMagas × son prix. Le k-ième
// magasinier d'un espace coûte selon le k-ième produit de l'espace (dans l'ordre d'arrivée).
// Les produits de la ferme (début du jeu, tutoriel) gardent leurs prix d'avant.
const (
	superRayons = 0.25
	superMagas  = 0.5
)

// SourceOf : le lieu qui fabrique un produit vendu au supermarché ("" : inconnu).
func (g *Game) SourceOf(it Item) string {
	for _, p := range g.Places {
		if p.ID == "super" {
			continue
		}
		for _, s := range p.Stations {
			if (s.Kind == Field || s.Kind == Machine || s.Kind == Animal) && s.Item == it {
				return p.ID
			}
		}
	}
	for _, it2 := range PortFish { // (les poissons du port : pêchés par les bateaux)
		if it2 == it {
			return "port"
		}
	}
	return ""
}

// superSource : le lieu d'origine et son prix d'achat (0 si la règle ne s'applique pas).
func (g *Game) superSource(it Item) (string, float64) {
	id := g.SourceOf(it)
	q := g.Place(id)
	if q == nil || PriceRank(id) < palierFrom {
		return id, 0
	}
	return id, q.Price
}

func (g *Game) superPrices() {
	sup := g.Place("super")
	if sup == nil {
		return
	}
	n := map[string]int{} // produits de chaque lieu au supermarché
	for _, z := range SuperZones {
		for _, it := range z.Items {
			if id, pr := g.superSource(it); pr > 0 {
				n[id]++
			}
		}
	}
	round := func(v float64) float64 { return math.Max(50, math.Round(v/50)*50) }
	for _, u := range sup.Unlocks {
		var it Item
		if _, err := fmt.Sscanf(u.ID, "r_%d", &it); err != nil {
			continue
		}
		if id, pr := g.superSource(it); pr > 0 {
			u.Cost = round(superRayons * pr / float64(n[id]))
		}
	}
	for _, z := range SuperZones {
		for _, u := range sup.ups {
			if u.Key != sup.key(teamKey(z.ID)) {
				continue
			}
			for l, it := range z.Items {
				if l >= len(u.Costs) {
					break
				}
				if id, pr := g.superSource(it); pr > 0 {
					u.Costs[l] = round(superMagas * pr / float64(n[id]))
				}
			}
		}
	}
}
