package sim

import "math"

// Les stands du parc : ballons, glaces, pop-corn… Ils sont au bord du
// boulevard, près d'une attraction (on les découvre quand elle est ouverte).
// Personne ne les tient et rien ne leur est livré (« c'est magique ») : un
// visiteur de bonne humeur qui passe devant achète parfois quelque chose.
// Points de l'école « Panier moyen » : les visiteurs y dépensent plus.

type StandKind struct {
	ID, Name string
	Cost     float64
	Share    float64 // dépense d'un visiteur (fraction du prix d'entrée de base)
	Slot     int     // terrain voisin (le stand est au bord du boulevard, devant lui)
	DX       float64 // position le long du terrain
	Ride     string  // attraction qui le fait apparaître
}

var StandKinds = []StandKind{
	{ID: "ballons", Name: "Ballons", Cost: 40000, Share: 0.06, Slot: 5, DX: 1500, Ride: "train"},
	{ID: "glaces", Name: "Marchand de glaces", Cost: 100000, Share: 0.08, Slot: 4, DX: 1500, Ride: "fantome"},
	{ID: "popcorn", Name: "Pop-corn", Cost: 250000, Share: 0.08, Slot: 6, DX: 1580, Ride: "volante"},
	{ID: "barbe", Name: "Barbe à papa", Cost: 600000, Share: 0.09, Slot: 7, DX: 550, Ride: "bateaux"},
	{ID: "photo", Name: "Photo souvenir", Cost: 1300000, Share: 0.16, Slot: 1, DX: 2400, Ride: "grandhuit"},
	{ID: "foodtruck", Name: "Food truck", Cost: 3000000, Share: 0.14, Slot: 2, DX: 2300, Ride: "splash"},
	{ID: "canards", Name: "Stand de crêpes", Cost: 6500000, Share: 0.12, Slot: 0, DX: 2300, Ride: "vide"},
	{ID: "souvenirs", Name: "Boutique de souvenirs", Cost: 12000000, Share: 0.2, Slot: 3, DX: 700, Ride: "vide"},
}

// Stand : un stand dans la partie.
type Stand struct {
	Kind *StandKind
	N    int // rang (bit dans Visitor.shopped)
	Pos  V   // pied du stand
	Top  bool
	Sold float64 // temps depuis la dernière vente (animation)
}

func (s *Stand) Zone() Rect { return Rect{s.Pos.X - 90, s.Pos.Y - 70, 180, 140} }

func standPos(k *StandKind) (V, bool) {
	pl := ParcSlot(k.Slot)
	top := k.Slot < 4
	if top {
		y := ParcBoulY - ParcBoulH/2 + 85
		if k.ID == "foodtruck" || k.ID == "souvenirs" {
			y += 50 // (plus grands : un peu plus bas sur le boulevard)
		}
		return V{pl.X + k.DX, y}, true
	}
	return V{pl.X + k.DX, ParcBoulY + ParcBoulH/2 - 45}, false
}

func (p *Place) addStands() {
	pc := p.Parc
	for i := range StandKinds {
		k := &StandKinds[i]
		pos, top := standPos(k)
		s := &Stand{Kind: k, N: i, Pos: pos, Top: top, Sold: 99}
		pc.Stands = append(pc.Stands, s)
		st := &Station{ID: "stand_" + k.ID, Name: k.Name, Kind: Trash, Unlock: "stand_" + k.ID}
		st.Body, st.Zone = Rect{-9000, -9000, 0, 0}, Rect{-9000, -9000, 0, 0}
		p.Stations = append(p.Stations, st)
		ride := p.Station("attr_" + k.Ride)
		p.Unlocks = append(p.Unlocks, &Unlock{ID: st.Unlock, Name: k.Name, Cost: k.Cost, Zone: s.Zone(),
			Requires: func() bool { return ride != nil && ride.Built }, Secret: true,
			Hint: "Un stand : les visiteurs heureux qui passent devant y dépensent un peu d'argent."})
	}
}

// StandBuilt : le stand est-il ouvert ?
func (p *Place) StandBuilt(s *Stand) bool {
	st := p.Station("stand_" + s.Kind.ID)
	return st != nil && st.Built
}

// StandSpend : ce que dépense un visiteur à ce stand.
func (p *Place) StandSpend(s *Stand) float64 {
	return math.Round(parcFeeBase * s.Kind.Share * (1 + p.ptsF("pts_panier")) * p.SellMult * p.g.priceMult(p))
}

const (
	parcStandJoy  = 55.0 // il faut être de bonne humeur pour acheter
	parcStandOdds = 0.5
	parcHoldTime  = 25.0 // il garde son achat à la main (ballon, glace…)
)

// parcShop : un visiteur passe devant les stands.
func (p *Place) parcShop(v *Visitor) {
	for _, s := range p.Parc.Stands {
		bit := uint32(1) << uint(s.N)
		if v.shopped&bit != 0 || math.Abs(v.Pos.X-s.Pos.X) > 50 || math.Abs(v.Pos.Y-s.Pos.Y) > 280 {
			continue
		}
		v.shopped |= bit
		if !p.StandBuilt(s) || v.Joy < parcStandJoy || p.rand() > parcStandOdds {
			continue
		}
		p.g.earn(p.StandSpend(s), s.Pos.Add(V{0, -120}), p.ID)
		s.Sold = 0
		v.Joy = math.Min(100, v.Joy+3)
		v.Holding, v.HoldT = s.N+1, parcHoldTime
		p.Parc.StandSales++
	}
}
