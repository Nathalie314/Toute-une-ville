package sim

import "math/rand"

// Le port : l'eau à gauche, la criée (toute en bois, sans murs) au centre, la
// zone du joueur à droite en haut, le parking à droite en bas.
//
// Une jetée part de la criée vers l'ouest. Au nord de la jetée, les bateaux de
// pêche (sardine au plus près de la criée, puis cabillaud, thon, saumon) ; au
// sud, les parcs immergés (huîtres, moules, saint-jacques, gambas). Tout est
// automatique : les bateaux partent vers le nord, disparaissent, reviennent
// (20 s le voyage) et vident leur pêche dans le panier de leur poisson ; une
// drague par parc ramasse les coquillages mûrs et les met dans leur panier.
// Le joueur ne va pas sur l'eau : il ne récolte rien, il porte et encaisse.
// Les ouvriers du port portent les paniers aux étals de la criée ; un drone par
// étal porte la marchandise au quai du camion (supermarché).

const portPrice = 85000

// géométrie (voir le plan dans NOTES.md)
const (
	PortCrieeX0 = 2700.0 // bord ouest de la criée (le quai)
	PortCrieeX1 = 4150.0
	PortCrieeY0 = 150.0
	PortCrieeY1 = 3350.0 // (la criée longe toute la mer, jusqu'au second ponton)
	PortJettyX0 = 380.0  // bout de la jetée
	PortJettyY0 = 520.0  // (520 de large : on circule au milieu sans passer sur les paniers)
	PortJettyY1 = 1040.0
	PortSecW    = 520.0             // largeur d'une section (un poisson au nord, un parc au sud)
	PortSec0X   = 2050.0            // section la plus proche de la criée (sardine, huîtres)
	PortBerthY  = PortJettyY0 - 185 // proue d'un bateau à quai
	PortBoatL   = 170.0             // longueur d'un bateau
	PortSeaY    = -1500.0           // au large : le bateau a disparu
	PortDecoTop = -1150.0
	PortParcY   = PortJettyY1 + 40 // haut des parcs
	PortCell    = 115.0            // une case de parc
	PortBoatMax = 4                // bateaux par poisson
	// le marais salant, au nord de la criée : 2 rangées de 3 bassins, sur une terre de digues
	PortSalTop  = -830.0 // haut de la terre du marais
	PortSalX0   = 2755.0 // bord ouest du premier bassin (centré sur la criée)
	PortSalY0   = -700.0 // haut de la rangée du haut
	PortSalW    = 400.0  // un bassin
	PortSalH    = 220.0
	PortSalGapX = 70.0 // les digues entre les bassins
	PortSalGapY = 50.0
	PortSalMax  = 6
	// la route verticale, à droite de la criée ; la zone du joueur et le parking de l'autre côté
	PortRoadX = 4300.0 // bord gauche de la route
	PortRoadW = 150.0
	PortLaneS = PortRoadX + 40  // voie vers le sud (les voitures roulent à droite)
	PortLaneN = PortRoadX + 110 // voie vers le nord
	PortGateY = 1960.0          // le passage piéton, face à l'entrée de la criée
	PortH     = 3500.0
	// les quais et la voie de livraison (sous le passage piéton)
	PortDockY0   = 2110.0
	PortDockStep = 190.0
	PortBayX     = PortCrieeX1 + 85 // le milieu de la voie de livraison (les camions s'y garent)
	PortBayY0    = 2010.0
	PortBayY1    = 2470.0
	PortParkX0   = PortRoadX + PortRoadW + 110 // le parking, sous la zone du joueur
	PortParkX1   = PortParkX0 + 830
)

// PortSalPivot : le pied du râteau mécanique (sous les bassins, au milieu) ; le tas de sel est juste dessous.
var PortSalPivot = V{(PortCrieeX0 + PortCrieeX1) / 2, -150}

// PortSalBasin : le bassin n° k, dans l'ordre d'achat (le plus proche du râteau d'abord :
// en bas au milieu, en bas à gauche, en bas à droite, puis la rangée du haut).
func PortSalBasin(k int) Rect {
	col := []int{1, 0, 2, 1, 0, 2}[k]
	row := 1 - k/3
	return Rect{PortSalX0 + float64(col)*(PortSalW+PortSalGapX), PortSalY0 + float64(row)*(PortSalH+PortSalGapY), PortSalW, PortSalH}
}

// PortFish, PortShell : les poissons (bateaux) et les coquillages (parcs), du
// plus proche au plus loin de la criée, dans l'ordre de déblocage.
var PortFish = []Item{Sardine, Cabillaud, Thon, Saumon}
var PortShell = []Item{Huitre, Moule, SaintJacques, Gambas}
var PortSold = append(append(append([]Item{}, PortFish...), PortShell...), PortShell2...)

// PortShell2 : les produits du second ponton, dans l'ordre d'achat.
var PortShell2 = []Item{Bigorneau, Palourde, Crabe, Tourteau, Araignee, Langoustine, Homard, Langouste}

// le second ponton, sous les parcs : des bassins au nord et au sud
const (
	PortJ2Y0 = 2230.0
	PortJ2Y1 = 2750.0
)

// PortParc : un parc (ponton 1, au sud) ou un bassin (ponton 2, au nord ou au sud). Une drague
// par parc ramasse ce qui est mûr et le met dans son panier, au bord du ponton.
type PortParc struct {
	ID, Name string
	It       Item
	Time     float64 // pousse d'une case
	Case     float64 // prix de base des cases
	Cost     float64 // déblocage (ponton 2 ; ponton 1 : portUnlockCost)
	Sec      int     // section (0 : la plus proche de la criée)
	Side     int     // 0 : ponton 1, au sud ; 1 : ponton 2, au nord ; 2 : ponton 2, au sud
	Net      bool    // un filet (crustacés) plutôt qu'une benne
}

var PortParcs = []*PortParc{
	{ID: "huitres", Name: "Parc à huîtres", It: Huitre, Time: 12, Case: 1500, Sec: 0},
	{ID: "moules", Name: "Parc à moules", It: Moule, Time: 10, Case: 2200, Sec: 1},
	{ID: "stjacques", Name: "Parc à saint-jacques", It: SaintJacques, Time: 14, Case: 3200, Sec: 2},
	{ID: "gambas", Name: "Bassin à gambas", It: Gambas, Time: 16, Case: 4500, Sec: 3, Net: true},
	// ponton 2 (dans l'ordre d'achat, après le marais salant)
	{ID: "bigorneaux", Name: "Bassin à bigorneaux", It: Bigorneau, Time: 15, Case: 5500, Cost: 400000, Sec: 0, Side: 1},
	{ID: "palourdes", Name: "Bassin à palourdes", It: Palourde, Time: 16, Case: 6500, Cost: 480000, Sec: 1, Side: 1},
	{ID: "crabes", Name: "Bassin à crabes", It: Crabe, Time: 17, Case: 7500, Cost: 570000, Sec: 0, Side: 2, Net: true},
	{ID: "tourteaux", Name: "Bassin à tourteaux", It: Tourteau, Time: 18, Case: 8500, Cost: 680000, Sec: 1, Side: 2, Net: true},
	{ID: "araignees", Name: "Bassin à araignées de mer", It: Araignee, Time: 19, Case: 10000, Cost: 810000, Sec: 2, Side: 1, Net: true},
	{ID: "langoustines", Name: "Bassin à langoustines", It: Langoustine, Time: 20, Case: 11500, Cost: 960000, Sec: 2, Side: 2, Net: true},
	{ID: "homards", Name: "Bassin à homards", It: Homard, Time: 22, Case: 13000, Cost: 1140000, Sec: 3, Side: 1, Net: true},
	{ID: "langoustes", Name: "Bassin à langoustes", It: Langouste, Time: 24, Case: 15000, Cost: 1350000, Sec: 3, Side: 2, Net: true},
}

// ParcTop : le haut du parc ; les cases du ponton 2, côté nord, se remplissent depuis le ponton.
func (c *PortParc) ParcTop() float64 {
	switch c.Side {
	case 1:
		return PortJ2Y0 - 40 - 4*PortCell
	case 2:
		return PortJ2Y1 + 40
	}
	return PortParcY
}

// edge : le bord du ponton où sont le panier, la drague et la case « Case +1 » (top : le bord du haut).
func (c *PortParc) edge() (y float64, top bool) {
	switch c.Side {
	case 1:
		return PortJ2Y0, true
	case 2:
		return PortJ2Y1, false
	}
	return PortJettyY1, false
}

// LabelY : le nom du parc (sous le parc ; au-dessus pour le ponton 2, côté nord).
func (c *PortParc) LabelY() float64 {
	if c.Side == 1 {
		return c.ParcTop() - 34
	}
	return c.ParcTop() + 4*PortCell + 44
}

// portParcByID : le parc d'un ID (nil si aucun).
func portParcByID(id string) *PortParc {
	for _, c := range PortParcs {
		if c.ID == id {
			return c
		}
	}
	return nil
}

var portFishID = []string{"sardine", "cabillaud", "thon", "saumon"}
var portShellID = []string{"huitres", "moules", "stjacques", "gambas"}

// PortSecX : bord ouest de la section n° i (0 : la plus proche de la criée).
func PortSecX(i int) float64 { return PortSec0X - float64(i)*(PortSecW+20) }

// PortBerth : centre (x) du poste n° k de la section i.
func PortBerth(i, k int) float64 { return PortSecX(i) + 90 + float64(k)*115 }

// Boat : un bateau de pêche (pas sauvegardé : il repart de son poste).
type Boat struct {
	Sec, K int
	Pos    V
	Phase  int     // 0 : à quai ; 1 : il part ; 2 : au large ; 3 : il revient ; 4 : il décharge
	N      int     // poissons dans la soute
	T      float64 // attente (au large, à quai)
}

// Dredge : la drague d'un parc (un bras qui va chercher une case mûre).
type Dredge struct {
	Base   V   // pied du bras (sur la jetée)
	Tip    V   // bout du bras
	Target int // case visée (-1 : aucune)
	Phase  int // 0 : replié ; 1 : il s'étend ; 2 : il racle ; 3 : il revient (chargé)
	N      int // coquillages ramenés
	T      float64
	Net    bool   // un filet (crustacés) plutôt qu'une benne
	parc   string // la station du parc et celle de son panier
	basket string
}

// DredgeOpen : la drague travaille (son parc est acheté).
func (p *Place) DredgeOpen(d *Dredge) bool {
	s := p.Station(d.parc)
	return s != nil && s.Built
}

// BoatSpeed : vitesse d'un bateau (le voyage aller-retour dure 20 s au départ).
func (p *Place) BoatSpeed() float64 {
	return 2 * (PortBerthY - PortSeaY) / 17 * (1 + 0.03*float64(p.g.Levels[p.key("pts_bat_vit")])) * p.g.boostF("marin", "")
}

// BoatCap : poissons par voyage.
func BoatCap(pts int) int { return 1 + pts/3 }

// DredgeSpeed : vitesse du bras ; DredgeCap : coquillages par passage.
func DredgeSpeed(pts int) float64 { return 260 * (1 + 0.03*float64(pts)) }
func DredgeCap(pts int) int       { return 1 + pts/4 }

func newPort(g *Game) *Place {
	W := PortParkX1 + 110
	p := &Place{ID: "port", Name: "Le port", W: W, H: PortH, Top: 0, DecoTop: PortDecoTop, g: g, rng: rand.New(rand.NewSource(71)),
		Rep: 3, Price: portPrice, SellMult: 1, SpawnBase: 5, WantMax: 3, MaxCust: 16, LeaveOff: V{170, 60}, RepGain: 0.01, OnlyStocked: true}
	// les piétons arrivent par le trottoir de la route verticale et traversent au passage piéton
	p.SpawnL = V{PortRoadX + PortRoadW + 40, PortH + 300}
	p.SpawnR = V{PortRoadX + PortRoadW + 40, PortSalTop - 300}
	p.Gate = V{PortCrieeX1 + 70, PortGateY}
	p.Inside = V{PortCrieeX1 - 200, PortGateY}
	p.RestSpot = V{3900, 2450}
	p.Start = V{PortCrieeX1 - 250, PortGateY}
	// l'eau : tout ce qui est à l'ouest de la criée, sauf les jetées (on n'y marche pas), et au nord du marais
	p.Water = []Rect{
		{-3000, -4000, PortCrieeX0 + 3000, PortJettyY0 + 4000},              // au nord de la jetée
		{-3000, PortJettyY1, PortCrieeX0 + 3000, PortJ2Y0 - PortJettyY1},    // entre les deux pontons
		{-3000, PortJ2Y1, PortCrieeX0 + 3000, PortH + 3000 - PortJ2Y1},      // au sud du second ponton
		{-3000, PortJettyY0, PortJettyX0 + 3000, PortJettyY1 - PortJettyY0}, // au-delà du bout de la jetée
		{-3000, PortJ2Y0, PortJettyX0 + 3000, PortJ2Y1 - PortJ2Y0},          // au-delà du bout du second ponton
		{PortCrieeX0, -4000, PortCrieeX1 - PortCrieeX0, 4000 + PortSalTop},  // au nord du marais salant : la mer
	}
	// les bassins et leurs digues : personne n'y va (le joueur achète les bassins depuis le bord)
	p.Water = append(p.Water, Rect{PortSalX0 - 10, PortSalY0 - 10, 3*PortSalW + 2*PortSalGapX + 20, 2*PortSalH + PortSalGapY + 20})
	p.Top = PortSalTop

	// --- les sections : un panier par produit, les bateaux, les parcs et leur drague
	for i := 0; i < 4; i++ {
		sx := PortSecX(i)
		fid := portFishID[i]
		fish := PortFish[i]
		unlF := fid
		if i == 0 {
			unlF = "" // la sardine est là dès l'achat du port
		}
		bf := &Station{ID: "panier_" + fid, Name: "Panier : " + Items[fish].Name, Kind: Pallet, Item: fish, PileMax: 40, Unlock: unlF}
		bf.Body = Rect{sx + 20, PortJettyY0 + 14, 110, 56}
		bf.Zone = Rect{sx + 10, PortJettyY0 + 76, 130, 80}
		p.Stations = append(p.Stations, bf)
	}
	// --- les parcs (ponton 1) et les bassins (ponton 2) : un panier au bord du ponton, 4 × 4 cases
	// sous l'eau (une case au départ, les autres s'achètent depuis le ponton) et une drague
	for _, c := range PortParcs {
		sx := PortSecX(c.Sec)
		ey, top := c.edge()
		bs := &Station{ID: "panier_" + c.ID, Name: "Panier : " + Items[c.It].Name, Kind: Pallet, Item: c.It, PileMax: 40, Unlock: c.ID}
		if top {
			bs.Body = Rect{sx + 20, ey + 14, 110, 56}
			bs.Zone = Rect{sx + 10, ey + 76, 130, 80}
		} else {
			bs.Body = Rect{sx + 20, ey - 70, 110, 56}
			bs.Zone = Rect{sx + 10, ey - 156, 130, 80}
		}
		py := c.ParcTop()
		parc := &Station{ID: "parc_" + c.ID, Name: c.Name, Kind: Field, Item: c.It, Time: c.Time, Unlock: c.ID, Dredged: true}
		parc.Body = Rect{sx + 20, py - 15, 4*PortCell + 20, 4*PortCell + 20}
		parc.Zone = Rect{-1000, -1000, 0, 0}
		for r := 0; r < 4; r++ {
			rr := r
			if top {
				rr = 3 - r // (côté nord du ponton 2 : la première case contre le ponton)
			}
			for col := 0; col < 4; col++ {
				parc.Slots = append(parc.Slots, V{sx + 30 + PortCell/2 + float64(col)*PortCell, py + PortCell/2 + float64(rr)*PortCell})
			}
		}
		parc.Plants = []Plant{{P: parc.Slots[0]}}
		p.Stations = append(p.Stations, bs, parc)
		by := ey - 40
		if top {
			by = ey + 40
		}
		p.dredges = append(p.dredges, &Dredge{Base: V{sx + 440, by}, Tip: V{sx + 440, by}, Target: -1, Net: c.Net, parc: "parc_" + c.ID, basket: "panier_" + c.ID})
	}

	// --- le marais salant (après les gambas) : six bassins (un au départ, « Bassin +1 » ensuite) ;
	// l'eau s'évapore (bleu → blanc), le râteau mécanique ramasse le sel et le met en tas
	mar := &Station{ID: "marais", Name: "Marais salant", Kind: Field, Item: Sel, Time: 24, Unlock: "marais", Dredged: true}
	mar.Body = Rect{PortSalX0 - 10, PortSalY0 - 10, 3*PortSalW + 2*PortSalGapX + 20, 2*PortSalH + PortSalGapY + 20}
	mar.Zone = Rect{-1000, -1000, 0, 0}
	for k := 0; k < PortSalMax; k++ {
		mar.Slots = append(mar.Slots, PortSalBasin(k).C())
	}
	mar.Plants = []Plant{{P: mar.Slots[0]}}
	tas := &Station{ID: "tas_sel", Name: "Tas de sel", Kind: Pallet, Item: Sel, PileMax: 40, Unlock: "marais"}
	tas.Body = Rect{PortSalPivot.X - 65, PortSalPivot.Y + 50, 130, 64}
	tas.Zone = Rect{PortSalPivot.X - 75, PortSalPivot.Y + 124, 150, 90}
	p.Stations = append(p.Stations, mar, tas)
	p.salt = &Dredge{Base: PortSalPivot, Tip: PortSalPivot, Target: -1}

	// --- la criée : deux rangées d'étals (poissons en haut, coquillages en bas), la caisse au milieu
	shelf := func(it Item, x, y float64, unl string) {
		s := &Station{ID: "etal_" + itoa(int(it)), Name: "Étal : " + Items[it].Name, Kind: Shelf, Item: it, StockMax: 12, Unlock: unl}
		s.Body = Rect{x, y, 170, 66}
		s.Zone = Rect{x - 8, y - 95, 186, 92}
		s.Spot = V{x + 85, y + 66 + 38}
		p.Stations = append(p.Stations, s)
	}
	for i := 0; i < 4; i++ {
		x := 2790 + float64(i)*330
		unlF := portFishID[i]
		if i == 0 {
			unlF = ""
		}
		shelf(PortFish[i], x, 760, unlF)
	}
	// les étals des parcs : les coquillages sous les poissons ; ceux du second ponton plus bas,
	// sous la caisse (nord du ponton 2, puis sud), chacun dans la colonne de sa section
	for _, c := range PortParcs {
		y := []float64{1200, 2150, 2700}[c.Side]
		shelf(c.It, 2790+float64(c.Sec)*330, y, c.ID)
	}
	// le sel : en haut de la criée, au milieu (face au marais salant)
	shelf(Sel, (PortCrieeX0+PortCrieeX1)/2-85, 330, "marais")
	// la caisse : sous les deux rangées
	reg := &Station{ID: "caisse", Name: "Caisse", Kind: Register, HireKey: "caissiere", QueueDown: 3}
	reg.Body = Rect{3290, 1650, 210, 66}
	reg.Spot = V{3395, 1610}
	reg.Zone = Rect{3310, 1522, 170, 96}
	reg.QueueHead = V{3395, 1756}
	p.Stations = append(p.Stations, reg)
	trash := &Station{ID: "poubelle", Name: "Poubelle", Kind: Trash}
	trash.Body = Rect{4040, 3220, 70, 70}
	trash.Zone = Rect{4020, 3180, 110, 150}
	p.Stations = append(p.Stations, trash)
	// les quais : contre la route verticale, sous le passage piéton ; les camions se garent dans la
	// voie de livraison (une baie entre la criée et la route, hors de la circulation)
	for i, to := range []string{"super", "creperie"} { // (le thon part aussi à la crêperie)
		id := "quai"
		if i > 0 {
			id = "quai_" + to
		}
		y := PortDockY0 + float64(i)*PortDockStep
		q := &Station{ID: id, Name: "Quai du camion", Kind: Dock, BufMax: 20, Built: true, To: to}
		q.Body = Rect{PortCrieeX1 - 180, y, 170, 56}
		q.Zone = Rect{PortCrieeX1 - 190, y - 72, 200, 154} // (le texte au-dessus, le quai dessous)
		p.Stations = append(p.Stations, q)
	}
	cmd := &Station{ID: "commande", Name: "Commandes", Kind: OrderBox, Built: true}
	p.Stations = append(p.Stations, cmd)

	// --- déblocages, dans l'ordre (sur la jetée, devant la section)
	for i, id := range append(append([]string{}, portFishID...), portShellID...) {
		if id == "sardine" {
			continue
		}
		k := i % 4
		sx := PortSecX(k)
		y := PortJettyY0 + 10.0
		name, hint := "Bateaux : "+Items[PortFish[k]].Name, "Un bateau de pêche automatique : il part au large et revient la soute pleine."
		if i >= 4 {
			y = PortJettyY1 - 140
			name, hint = portParcName[k], "Un parc sous l'eau : une drague ramasse ce qui est mûr. Achète des cases depuis la jetée."
		}
		p.Unlocks = append(p.Unlocks, &Unlock{ID: id, Name: name, Cost: portUnlockCost[i], Zone: Rect{sx + 170, y, 200, 130}, Hint: hint})
	}
	p.Unlocks = append(p.Unlocks, &Unlock{ID: "marais", Name: "Marais salant", Cost: 330000, Zone: Rect{PortSalPivot.X + 140, PortSalPivot.Y - 10, 200, 130},
		Hint: "Six bassins d'eau de mer : l'eau s'évapore, le râteau mécanique ramasse le sel. Achète les bassins depuis le marais."})
	for _, c := range PortParcs[4:] { // le second ponton, après le marais salant
		sx := PortSecX(c.Sec)
		ey, top := c.edge()
		y := ey - 140
		if top {
			y = ey + 10
		}
		p.Unlocks = append(p.Unlocks, &Unlock{ID: c.ID, Name: c.Name, Cost: c.Cost, Zone: Rect{sx + 170, y, 200, 130},
			Hint: "Un bassin sous l'eau : une drague ramasse ce qui est mûr. Achète des cases depuis le ponton."})
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
		{Key: k("etals"), Name: "Étals plus grands", Max: 6, Base: 5000, Mult: 2.3, Desc: func(l int) string { return "Chaque étal contient " + itoa(12+4*l) + " produits." }},
		{Key: k("entrepot"), Name: "Paniers plus grands", Max: 12, Base: 6000, Mult: 1.6, Desc: func(l int) string {
			return "Chaque panier de la jetée contient " + itoa(PalletMax(p.ID, l)) + " produits."
		}},
		{Key: k("ouvriers"), Name: "Ouvrier du port", Max: 17, Now: true, Costs: []float64{4000, 8000, 15000, 25000, 40000, 60000, 90000, 130000, 180000, // (le 9e : le marais salant ;
			240000, 320000, 420000, 550000, 720000, 940000, 1220000, 1600000}, // les 8 derniers : le second ponton)
			Desc: func(l int) string {
				return "Il porte les paniers de la jetée jusqu'aux étals de la criée. Embauchés : " + itoa(l) + "."
			}},
	}
	for i := 0; i < 4; i++ {
		i := i
		p.ups = append(p.ups,
			&Upgrade{Key: k("bateaux_" + portFishID[i]), Name: "Bateau : " + Items[PortFish[i]].Name, Max: PortBoatMax - 1, Now: true,
				Costs: portBoatCosts[i], Desc: func(l int) string {
					return "Un bateau de pêche de plus (automatique). Bateaux : " + itoa(l+1) + " / " + itoa(PortBoatMax) + "."
				}},
		)
	}
	for _, c := range PortParcs {
		p.ups = append(p.ups, &Upgrade{Key: k("cases_" + c.ID), Name: "Case : " + c.Name, Max: 15, Now: true,
			Base: c.Case, Mult: 1.16, Desc: func(l int) string {
				return "Une case de plus dans le parc (la drague y ramasse les coquillages). Cases : " + itoa(l+1) + " / 16."
			}})
	}
	p.ups = append(p.ups, &Upgrade{Key: k("cases_marais"), Name: "Bassin : Marais salant", Max: PortSalMax - 1, Now: true, Base: 9000, Mult: 1.35,
		Desc: func(l int) string {
			return "Un bassin de plus (le râteau y ramasse le sel). Bassins : " + itoa(l+1) + " / " + itoa(PortSalMax) + "."
		}})
	p.ups = append(p.ups, p.droneUpgrades(2)...)
	p.Spots = append(p.Spots,
		&Spot{ID: "cases_marais", Label: "Bassin +1", Zone: Rect{PortSalPivot.X + 140, PortSalPivot.Y - 10, 130, 120}, NeedsBuilt: "marais",
			Done: func(l int) string { return "Marais salant : un bassin de plus !" }},
		&Spot{ID: "caissiere", Label: "Caissière", Zone: Rect{3560, 1540, 145, 116}, Done: func(int) string { return "Caissière embauchée !" }},
		&Spot{ID: "etals", Label: "Étals +", Zone: Rect{2830, 1540, 130, 116}, Done: func(l int) string { return "Étals agrandis : niveau " + itoa(l) + " !" }},
		&Spot{ID: "entrepot", Label: "Paniers +", Zone: Rect{3720, 1540, 130, 116}, Done: func(l int) string {
			return "Paniers agrandis : " + itoa(PalletMax(p.ID, l)) + " par panier !"
		}})
	for _, st := range p.Stations {
		if st.Kind == Shelf {
			it := st.Item
			p.Spots = append(p.Spots, &Spot{ID: droneKey(it), Label: "Drone", Shelf: st, Zone: st.DronePad(),
				Done: func(int) string { return "Drone acheté pour l'étal « " + Items[it].Name + " » !" }})
		}
	}
	for i := 0; i < 4; i++ {
		i := i
		sx := PortSecX(i)
		p.Spots = append(p.Spots,
			&Spot{ID: "bateaux_" + portFishID[i], Label: "Bateau +1", Zone: Rect{sx + 205, PortJettyY0 + 14, 130, 120}, NeedsBuilt: "panier_" + portFishID[i],
				Done: func(l int) string { return "Nouveau bateau : " + Items[PortFish[i]].Name + " !" }})
	}
	for _, c := range PortParcs {
		c := c
		sx := PortSecX(c.Sec)
		ey, top := c.edge()
		y := ey - 134
		if top {
			y = ey + 14
		}
		p.Spots = append(p.Spots, &Spot{ID: "cases_" + c.ID, Label: "Case +1", Zone: Rect{sx + 205, y, 130, 120}, NeedsBuilt: "parc_" + c.ID,
			Done: func(l int) string { return c.Name + " : une case de plus !" }})
	}
	p.Spots = append(p.Spots, &Spot{ID: "ouvriers", Label: "Ouvrier du port", Done: func(l int) string { return "Ouvrier du port +1 !" }})
	p.setupYard(PortParkX0 + YardW - yardOldW) // (la zone, de l'autre côté de la route verticale)
	// (les ouvriers s'embauchent à gauche de la caisse, pas dans la zone du joueur)
	p.spot("ouvriers").Zone = Rect{2985, 1540, 145, 116} // (à gauche de la caisse ; « Étals + » à sa gauche)
	p.setupPortParking()
	p.spawnT = 0.5
	return p
}

var portParcName = []string{"Parc à huîtres", "Parc à moules", "Parc à saint-jacques", "Bassin à gambas"}

// déblocages : cabillaud, thon, saumon, puis les quatre parcs
var portUnlockCost = []float64{0, 15000, 32000, 60000, 90000, 130000, 180000, 250000}
var portBoatCosts = [][]float64{{5000, 9000, 15000}, {12000, 20000, 32000}, {26000, 40000, 60000}, {45000, 70000, 100000}}

// portOrder : un particulier commande deux ou trois produits du port.
func (g *Game) portOrder() *Order { return g.shopOrder("port") }

// syncPort : autant de bateaux et de cases de parc que de niveaux achetés.
func (p *Place) syncPort() {
	g := p.g
	for i := 0; i < 4; i++ {
		if b := p.Station("panier_" + portFishID[i]); b.Built {
			n := 1 + min(PortBoatMax-1, g.Levels[p.key("bateaux_"+portFishID[i])])
			have := 0
			for _, bt := range p.Boats {
				if bt.Sec == i {
					have++
				}
			}
			for ; have < n; have++ {
				b := &Boat{Sec: i, K: have, Pos: V{PortBerth(i, have), PortSeaY}, Phase: 3} // un nouveau bateau arrive du large
				if !p.boatsInit {
					// (à l'achat du port ou en rouvrant le jeu : il est à quai)
					b.Pos.Y, b.Phase, b.T = PortBerthY+PortBoatL/2, 0, 0.5+0.8*float64(have)
				}
				p.Boats = append(p.Boats, b)
			}
		}
	}
	for _, c := range PortParcs {
		parc := p.Station("parc_" + c.ID)
		if parc.Built {
			n := 1 + min(15, g.Levels[p.key("cases_"+c.ID)])
			for len(parc.Plants) < n {
				parc.Plants = append(parc.Plants, Plant{P: parc.Slots[len(parc.Plants)]})
			}
		}
	}
	if mar := p.Station("marais"); mar != nil && mar.Built {
		n := 1 + min(PortSalMax-1, g.Levels[p.key("cases_marais")])
		for len(mar.Plants) < n {
			mar.Plants = append(mar.Plants, Plant{P: mar.Slots[len(mar.Plants)]})
		}
	}
	p.boatsInit = true
}

// SaltSpeed : vitesse du râteau du marais ; SaltCap : sel ramassé par bassin.
func SaltSpeed(pts int) float64 { return 300 * (1 + 0.03*float64(pts)) }
func SaltCap(pts int) int       { return 3 + pts/3 }

// Salt : le râteau mécanique du marais salant (dessin).
func (p *Place) Salt() *Dredge { return p.salt }

// stepSalt : le râteau va au bassin blanc le plus proche, racle le sel (le bassin se remplit
// d'eau de mer et redevient bleu) et le rapporte sur le tas.
func (p *Place) stepSalt(dt float64) {
	g := p.g
	mar, tas, d := p.Station("marais"), p.Station("tas_sel"), p.salt
	if mar == nil || !mar.Built || d == nil {
		return
	}
	v := SaltSpeed(g.Levels[p.key("pts_sel_vit")])
	switch d.Phase {
	case 0:
		d.Tip = d.Base
		d.T -= dt
		if d.T > 0 || tas.Pile >= tas.PileMax {
			return
		}
		best, bd := -1, 1e9
		for k, pl := range mar.Plants {
			if pl.G >= 1 {
				if dd := pl.P.Dist(d.Base); dd < bd {
					best, bd = k, dd
				}
			}
		}
		if best >= 0 {
			d.Target, d.Phase = best, 1
		}
	case 1:
		if d.moveTip(mar.Plants[d.Target].P, v*dt) {
			d.Phase, d.T = 2, 1.2
		}
	case 2:
		d.T -= dt
		if d.T <= 0 {
			pl := &mar.Plants[d.Target]
			if pl.G >= 1 {
				d.N = SaltCap(g.Levels[p.key("pts_sel_cap")])
				pl.G = 0
				g.Stats.Harvested += d.N
				g.emit(Event{Kind: EvHarvest, Pos: pl.P, It: Sel})
			}
			d.Phase = 3
		}
	case 3:
		if d.moveTip(d.Base, v*dt) {
			for d.N > 0 && tas.Pile < tas.PileMax {
				d.N--
				tas.Pile++
				g.made(Sel)
			}
			d.N = 0
			d.Phase, d.T = 0, 0.3
		}
	}
}

// stepPort : les bateaux et les dragues.
func (p *Place) stepPort(dt float64) {
	g := p.g
	p.syncPort()
	speed := p.BoatSpeed()
	capa := BoatCap(g.Levels[p.key("pts_bat_cap")])
	for _, b := range p.Boats {
		basket := p.Station("panier_" + portFishID[b.Sec])
		berth := V{PortBerth(b.Sec, b.K), PortBerthY + PortBoatL/2}
		switch b.Phase {
		case 0: // à quai, vide : il repart (les bateaux d'une section partent l'un après l'autre)
			b.T -= dt
			if b.T <= 0 {
				b.Phase = 1
			}
		case 1: // vers le nord
			b.Pos.Y -= speed * dt
			if b.Pos.Y <= PortSeaY {
				b.Pos.Y = PortSeaY
				b.Phase, b.T = 2, 1.5
			}
		case 2: // au large : il pêche
			b.T -= dt
			if b.T <= 0 {
				b.N = capa
				b.Phase = 3
			}
		case 3: // il revient
			b.Pos.Y += speed * dt
			if b.Pos.Y >= berth.Y {
				b.Pos.Y = berth.Y
				b.Phase, b.T = 4, 0
			}
		case 4: // il vide sa pêche dans le panier (un poisson toutes les 0,1 s)
			b.T -= dt
			for b.T <= 0 && b.N > 0 && basket.Pile < basket.PileMax {
				b.N--
				basket.Pile++
				g.made(basket.Item)
				b.T += 0.1
			}
			if b.N == 0 {
				b.Phase, b.T = 0, 0.5+0.4*float64(b.K)
			}
		}
		b.Pos.X = berth.X
	}
	p.stepSalt(dt)
	// dragues
	for _, d := range p.dredges {
		parc := p.Station(d.parc)
		basket := p.Station(d.basket)
		if !parc.Built {
			continue
		}
		v := DredgeSpeed(g.Levels[p.key("pts_drag_vit")])
		switch d.Phase {
		case 0:
			d.Tip = d.Base
			d.T -= dt
			if d.T > 0 || basket.Pile >= basket.PileMax {
				continue
			}
			// la case mûre la plus proche
			best, bd := -1, 1e9
			for k, pl := range parc.Plants {
				if pl.G >= 1 {
					if dd := pl.P.Dist(d.Base); dd < bd {
						best, bd = k, dd
					}
				}
			}
			if best >= 0 {
				d.Target, d.Phase = best, 1
			}
		case 1:
			to := parc.Plants[d.Target].P
			if d.moveTip(to, v*dt) {
				d.Phase, d.T = 2, 0.6
			}
		case 2:
			d.T -= dt
			if d.T <= 0 {
				pl := &parc.Plants[d.Target]
				if pl.G >= 1 {
					d.N = DredgeCap(g.Levels[p.key("pts_drag_cap")])
					pl.G = 0
					g.Stats.Harvested += d.N
					g.emit(Event{Kind: EvHarvest, Pos: pl.P, It: parc.Item})
				}
				d.Phase = 3
			}
		case 3:
			if d.moveTip(d.Base, v*dt) {
				for d.N > 0 && basket.Pile < basket.PileMax {
					d.N--
					basket.Pile++
					g.made(basket.Item)
				}
				d.N = 0 // (panier plein : le reste retombe à l'eau)
				d.Phase, d.T = 0, 0.3
			}
		}
	}
}

func (d *Dredge) moveTip(to V, step float64) bool {
	dv := to.Sub(d.Tip)
	if dv.Len() <= step {
		d.Tip = to
		return true
	}
	d.Tip = d.Tip.Add(dv.Norm().Mul(step))
	return false
}

// Dredges : les dragues (dessin).
func (p *Place) Dredges() []*Dredge { return p.dredges }

// PortSecOpen : la section i des bateaux (fish) ou des parcs est-elle ouverte ?
func (p *Place) PortSecOpen(i int, shell bool) bool {
	if shell {
		return p.Station("parc_" + portShellID[i]).Built
	}
	return p.Station("panier_" + portFishID[i]).Built
}

// PortSecName : « Sardine », « Parc à huîtres »…
func PortSecName(i int, shell bool) string {
	if shell {
		return portParcName[i]
	}
	return "Bateaux : " + Items[PortFish[i]].Name
}
