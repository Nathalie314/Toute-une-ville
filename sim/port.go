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
	PortCrieeY1 = 1500.0
	PortJettyX0 = 380.0 // bout de la jetée
	PortJettyY0 = 520.0 // (520 de large : on circule au milieu sans passer sur les paniers)
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
)

// PortFish, PortShell : les poissons (bateaux) et les coquillages (parcs), du
// plus proche au plus loin de la criée, dans l'ordre de déblocage.
var PortFish = []Item{Sardine, Cabillaud, Thon, Saumon}
var PortShell = []Item{Huitre, Moule, SaintJacques, Gambas}
var PortSold = append(append([]Item{}, PortFish...), PortShell...)

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
	W := 4900.0
	p := &Place{ID: "port", Name: "Le port", W: W, H: 1800, Top: 0, DecoTop: PortDecoTop, g: g, rng: rand.New(rand.NewSource(71)),
		Rep: 3, Price: portPrice, SellMult: 1, SpawnBase: 5, WantMax: 3, MaxCust: 16, LeaveOff: V{170, 60}, RepGain: 0.01, OnlyStocked: true}
	gx := 3900.0
	p.SpawnL = V{2400, roadY + 40}
	p.SpawnR = V{W + 600, roadY + 40}
	p.Gate = V{gx, roadY + 40}
	p.Inside = V{gx, 1380}
	p.RestSpot = V{4000, 1250}
	// l'eau : tout ce qui est à l'ouest de la criée, sauf la jetée (on n'y marche pas)
	p.Water = []Rect{
		{-3000, -4000, PortCrieeX0 + 3000, PortJettyY0 + 4000},                         // au nord de la jetée
		{-3000, PortJettyY1, PortCrieeX0 + 3000, roadY - 12 - PortJettyY1},             // au sud (jusqu'au pont de la route)
		{-3000, PortJettyY0, PortJettyX0 + 3000, PortJettyY1 - PortJettyY0},            // au-delà du bout de la jetée
		{-3000, roadY + 162, PortCrieeX0 - 120 + 3000, 2000},                           // sous le pont
		{PortCrieeX0, -4000, PortCrieeX1 + 600 - PortCrieeX0, PortCrieeY0 + 4000 - 20}, // au nord de la criée : la mer aussi
	}

	// --- les sections : un panier par produit, les bateaux, les parcs et leur drague
	for i := 0; i < 4; i++ {
		sx := PortSecX(i)
		fid, sid := portFishID[i], portShellID[i]
		fish, shell := PortFish[i], PortShell[i]
		unlF, unlS := fid, sid
		if i == 0 {
			unlF = "" // la sardine est là dès l'achat du port
		}
		bf := &Station{ID: "panier_" + fid, Name: "Panier : " + Items[fish].Name, Kind: Pallet, Item: fish, PileMax: 40, Unlock: unlF}
		bf.Body = Rect{sx + 20, PortJettyY0 + 14, 110, 56}
		bf.Zone = Rect{sx + 10, PortJettyY0 + 76, 130, 80}
		bs := &Station{ID: "panier_" + sid, Name: "Panier : " + Items[shell].Name, Kind: Pallet, Item: shell, PileMax: 40, Unlock: unlS}
		bs.Body = Rect{sx + 20, PortJettyY1 - 70, 110, 56}
		bs.Zone = Rect{sx + 10, PortJettyY1 - 156, 130, 80}
		// le parc : 4 × 4 cases sous l'eau (une case au départ, les autres s'achètent depuis la jetée)
		parc := &Station{ID: "parc_" + sid, Name: portParcName[i], Kind: Field, Item: shell, Time: portParcTime[i], Unlock: unlS, Dredged: true}
		parc.Body = Rect{sx + 20, PortParcY - 15, 4*PortCell + 20, 4*PortCell + 20}
		parc.Zone = Rect{-1000, -1000, 0, 0}
		for r := 0; r < 4; r++ {
			for c := 0; c < 4; c++ {
				parc.Slots = append(parc.Slots, V{sx + 30 + PortCell/2 + float64(c)*PortCell, PortParcY + PortCell/2 + float64(r)*PortCell})
			}
		}
		parc.Plants = []Plant{{P: parc.Slots[0]}}
		p.Stations = append(p.Stations, bf, bs, parc)
		p.dredges = append(p.dredges, &Dredge{Base: V{sx + 440, PortJettyY1 - 40}, Tip: V{sx + 440, PortJettyY1 - 40}, Target: -1})
	}

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
		shelf(PortFish[i], x, 330, unlF)
		shelf(PortShell[i], x, 1020, portShellID[i])
	}
	reg := &Station{ID: "caisse", Name: "Caisse", Kind: Register, HireKey: "caissiere", QueueDown: 1}
	reg.Body = Rect{3290, 770, 210, 66}
	reg.Spot = V{3395, 730}
	reg.Zone = Rect{3310, 642, 170, 96}
	reg.QueueHead = V{3395, 876}
	p.Stations = append(p.Stations, reg)
	trash := &Station{ID: "poubelle", Name: "Poubelle", Kind: Trash}
	trash.Body = Rect{4040, 1330, 70, 70}
	trash.Zone = Rect{4020, 1290, 110, 150}
	p.Stations = append(p.Stations, trash)
	p.dock("quai", "super", 3560)
	p.dock("quai_creperie", "creperie", 3910) // (le thon)
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
		{Key: k("ouvriers"), Name: "Ouvrier du port", Max: 8, Now: true, Costs: []float64{4000, 8000, 15000, 25000, 40000, 60000, 90000, 130000},
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
			&Upgrade{Key: k("cases_" + portShellID[i]), Name: "Case : " + portParcName[i], Max: 15, Now: true,
				Base: portCaseBase[i], Mult: 1.16, Desc: func(l int) string {
					return "Une case de plus dans le parc (la drague y ramasse les coquillages). Cases : " + itoa(l+1) + " / 16."
				}})
	}
	p.ups = append(p.ups, p.droneUpgrades(2)...)
	p.Spots = append(p.Spots,
		&Spot{ID: "caissiere", Label: "Caissière", Zone: Rect{3560, 650, 145, 116}, Done: func(int) string { return "Caissière embauchée !" }},
		&Spot{ID: "etals", Label: "Étals +", Zone: Rect{3720, 650, 130, 116}, Done: func(l int) string { return "Étals agrandis : niveau " + itoa(l) + " !" }},
		&Spot{ID: "entrepot", Label: "Paniers +", Zone: Rect{3870, 650, 130, 116}, Done: func(l int) string {
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
				Done: func(l int) string { return "Nouveau bateau : " + Items[PortFish[i]].Name + " !" }},
			&Spot{ID: "cases_" + portShellID[i], Label: "Case +1", Zone: Rect{sx + 205, PortJettyY1 - 134, 130, 120}, NeedsBuilt: "parc_" + portShellID[i],
				Done: func(l int) string { return portParcName[i] + " : une case de plus !" }})
	}
	p.Spots = append(p.Spots, &Spot{ID: "ouvriers", Label: "Ouvrier du port", Done: func(l int) string { return "Ouvrier du port +1 !" }})
	p.setupYard(4570) // (la zone, 830 de large, commence à 4300 : à droite de la criée)
	// (les ouvriers s'embauchent à l'entrée de la jetée, pas dans la zone du joueur)
	p.spot("ouvriers").Zone = Rect{PortSec0X + PortSecW + 8, PortJettyY0 + 200, 118, 120}
	p.setupParking()
	p.spawnT = 0.5
	return p
}

var portParcName = []string{"Parc à huîtres", "Parc à moules", "Parc à saint-jacques", "Bassin à gambas"}
var portParcTime = []float64{12, 10, 14, 16}

// déblocages : cabillaud, thon, saumon, puis les quatre parcs
var portUnlockCost = []float64{0, 15000, 32000, 60000, 90000, 130000, 180000, 250000}
var portBoatCosts = [][]float64{{5000, 9000, 15000}, {12000, 20000, 32000}, {26000, 40000, 60000}, {45000, 70000, 100000}}
var portCaseBase = []float64{1500, 2200, 3200, 4500}

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
		parc := p.Station("parc_" + portShellID[i])
		if parc.Built {
			n := 1 + min(15, g.Levels[p.key("cases_"+portShellID[i])])
			for len(parc.Plants) < n {
				parc.Plants = append(parc.Plants, Plant{P: parc.Slots[len(parc.Plants)]})
			}
		}
	}
	p.boatsInit = true
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
	// dragues
	for i, d := range p.dredges {
		parc := p.Station("parc_" + portShellID[i])
		basket := p.Station("panier_" + portShellID[i])
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
