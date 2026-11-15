package sim

import (
	"math"
	"math/rand"
	"sort"
)

// Le parc d'attractions (13e lieu, après la pâtisserie).
//
// Il ne consomme rien et ne livre rien : il est autonome. L'argent vient de
// l'entrée (payée aux caisses, son prix monte avec le nombre d'attractions) et,
// plus tard, des stands. Les visiteurs arrivent en voiture (1 à 4 par voiture)
// sur le parking au sud de la route, paient l'entrée, prennent une trottinette
// s'il en reste, puis font une attraction, et une autre tant que leur jauge de
// bonheur est haute.
//
// Chaque attraction a une « taille » (niveau 1 à 10) : plus elle est grande,
// plus le tour dure et plus la jauge monte. Ses wagons (bateaux, rondins,
// nacelles…) s'achètent un par un, jamais plus que sa taille ne le permet. Un
// employé (ou le joueur, debout dans sa cabine) fait monter et descendre les
// visiteurs ; sans lui, l'attraction ne tourne pas.
//
// La file d'attente fait baisser la jauge. À la sortie d'une attraction, une
// jauge au plus bas fait perdre de la réputation, une jauge haute en fait
// gagner. Un visiteur ne quitte jamais une file, sauf dans une attraction
// sans employé quand sa jauge tombe à zéro (le parc perd alors de la réputation).

// géométrie
const (
	ParcW         = 13800.0
	ParcTop       = -2950.0
	ParcBoulY     = -900.0 // centre du grand boulevard, entre les deux rangées d'attractions
	ParcBoulH     = 460.0
	ParcAvX       = 6900.0 // l'avenue qui monte de l'entrée jusqu'au boulevard
	ParcAvW       = 500.0
	ParcPlazaY    = 1080.0 // haut de la place de l'entrée
	ParcRegY      = 1200.0 // les cabines des caisses
	ParcRegGap    = 190.0
	ParcRackY     = 1000.0        // en montant l'avenue après les caisses (on rend sa trottinette en repassant ici)
	ParcRackX     = ParcAvX + 960 // le râtelier des trottinettes, sur la place, à droite des caisses
	ParcRackW     = 400.0
	ParcRackTop   = 1130.0
	parcSpotW     = 85.0
	ParcRegMax    = 10
	ParcParkMax   = 300
	ParcTrottMax  = 600
	parcParkStart = 10
)

// ParcAisles : les allées du parking (une rangée de places de chaque côté).
var ParcAisles = []float64{roadY + 330, roadY + 660, roadY + 990}

// ParcH : le bas du lieu (sous le parking).
const ParcH = roadY + 1150

// ---------------------------------------------------------------- attractions

// RideKind : ce qui distingue une attraction des autres.
type RideKind struct {
	ID, Name string
	Seats    int     // places par wagon (bateau, rondin, nacelle, rangée)
	Units    int     // wagons au plus (paires de nacelles pour la tour volante)
	PerLevel float64 // wagons permis par niveau (0,5 : un tous les deux niveaux)
	Pair     bool    // les nacelles s'achètent par deux (face à face)
	Cost     float64 // prix d'achat (avant le coefficient du lieu)
	Time     float64 // durée d'un tour au niveau 1 (secondes)
	TimeLvl  float64 // durée en plus par niveau
	Water    bool    // sur l'eau
	Slot     int     // emplacement dans le parc
}

// RideKinds : les attractions, de la moins chère à la plus chère.
var RideKinds = []RideKind{
	{ID: "train", Name: "Petit train", Seats: 2, Units: 10, PerLevel: 1, Cost: 60000, Time: 18, TimeLvl: 7, Slot: 5},
	{ID: "fantome", Name: "Train fantôme", Seats: 2, Units: 10, PerLevel: 1, Cost: 150000, Time: 22, TimeLvl: 7, Slot: 4},
	{ID: "volante", Name: "Tour volante", Seats: 1, Units: 10, PerLevel: 1, Pair: true, Cost: 350000, Time: 20, TimeLvl: 6, Slot: 6},
	{ID: "bateaux", Name: "Petits bateaux", Seats: 4, Units: 5, PerLevel: 0.5, Cost: 800000, Time: 30, TimeLvl: 9, Water: true, Slot: 7},
	{ID: "grandhuit", Name: "Grand huit", Seats: 4, Units: 5, PerLevel: 0.5, Cost: 1800000, Time: 20, TimeLvl: 6, Slot: 1},
	{ID: "splash", Name: "Grand splash", Seats: 4, Units: 5, PerLevel: 0.5, Cost: 4000000, Time: 24, TimeLvl: 7, Water: true, Slot: 2},
	{ID: "vide", Name: "Les pieds dans le vide", Seats: 2, Units: 10, PerLevel: 1, Cost: 9000000, Time: 22, TimeLvl: 6, Slot: 0},
}

// ParcSlot : l'emplacement n° i (0 à 3 : rangée du haut ; 4 à 7 : rangée du bas).
func ParcSlot(i int) Rect {
	xs := []float64{200, 3500, 7300, 10600}
	x := xs[i%4]
	if i < 4 {
		return Rect{x, ParcTop + 100, 3000, ParcBoulY - ParcBoulH/2 - (ParcTop + 100) - 100}
	}
	y := ParcBoulY + ParcBoulH/2 + 100
	return Rect{x, y, 3000, ParcPlazaY - 120 - y}
}

// Ride : une attraction dans la partie.
type Ride struct {
	Kind     *RideKind
	Plot     Rect
	Top      bool // rangée du haut (son entrée est en bas)
	Queue    []*Visitor
	Riders   []*Visitor
	State    int     // 0 embarquement, 1 tour, 2 débarquement
	T        float64 // temps écoulé dans l'état
	Prog     float64 // avancement du tour (0 à 1)
	boardT   float64
	idleT    float64
	Trotts   int         // trottinettes garées à l'entrée
	Boats    []*RideBoat // attractions sur l'eau : bateaux (rondins) autonomes
	gapT     float64
	Operator *Agent
}

// Edge : la hauteur du bord du terrain qui donne sur le boulevard.
func (r *Ride) Edge() float64 {
	if r.Top {
		return r.Plot.Y + r.Plot.H
	}
	return r.Plot.Y
}

// in : vers l'intérieur du terrain depuis le boulevard (+1 vers le bas, −1 vers le haut).
func (r *Ride) in() float64 {
	if r.Top {
		return -1
	}
	return 1
}

// Entrance, Exit : sur le boulevard, à gauche (entrée) et à droite (sortie).
func (r *Ride) Entrance() V { return V{r.Plot.X + 180, r.Edge() - r.in()*60} }
func (r *Ride) Exit() V     { return V{r.Plot.X + r.Plot.W - 180, r.Edge() - r.in()*60} }

// QueuePos : la place n° k dans la file (en serpentin, dans le terrain).
func (r *Ride) QueuePos(k int) V {
	row, col := k/8, k%8
	if row%2 == 1 {
		col = 7 - col
	}
	return V{r.Plot.X + 70 + float64(col)*38, r.Edge() + r.in()*(50+float64(row)*40)}
}

// QueueBox : la zone de la file d'attente (barrières).
func (r *Ride) QueueBox() Rect {
	if r.Top {
		return Rect{r.Plot.X + 40, r.Edge() - 420, 330, 420}
	}
	return Rect{r.Plot.X + 40, r.Edge(), 330, 420}
}

// Booth : la cabine de l'employé (le joueur s'y tient pour faire tourner l'attraction).
func (r *Ride) Booth() Rect {
	return Rect{r.Plot.X + 400, r.Edge() + r.in()*60 - 50, 120, 100}
}

// Track : la zone du manège (tout le terrain sauf la file et les cases d'achat).
func (r *Ride) Track() Rect {
	p := r.Plot
	if r.Top {
		return Rect{p.X + 560, p.Y + 40, p.W - 600, p.H - 80}
	}
	return Rect{p.X + 560, p.Y + 40, p.W - 600, p.H - 80}
}

// ID, Key : identifiants (clés des améliorations).
func (r *Ride) ID() string { return r.Kind.ID }

// Level : la taille de l'attraction (1 à 10).
func (p *Place) RideLevel(r *Ride) int { return 1 + p.g.Levels[p.key("taille_"+r.ID())] }

// RideUnits : wagons (ou paires de nacelles) achetés.
func (p *Place) RideUnits(r *Ride) int { return 1 + p.g.Levels[p.key("wagons_"+r.ID())] }

// RideUnitsCap : wagons permis à cette taille.
func (p *Place) RideUnitsCap(r *Ride, level int) int {
	return min(r.Kind.Units, max(1, int(math.Ceil(float64(level)*r.Kind.PerLevel-1e-9))))
}

// RideSeats : places à bord.
func (p *Place) RideSeats(r *Ride) int {
	n := p.RideUnits(r) * r.Kind.Seats
	if r.Kind.Pair {
		n *= 2
	}
	return n
}

// RideBuilt : l'attraction est achetée.
func (p *Place) RideBuilt(r *Ride) bool {
	s := p.Station("attr_" + r.ID())
	return s != nil && s.Built
}

// RideTime : durée d'un tour.
func (p *Place) RideTime(r *Ride) float64 {
	return r.Kind.Time + r.Kind.TimeLvl*float64(p.RideLevel(r)-1)
}

// RideJoy : ce qu'un tour ajoute à la jauge de bonheur.
func RideJoy(level int) float64 { return 6 + 3.5*float64(level) }

// RideManned : quelqu'un fait monter et descendre les visiteurs.
func (p *Place) RideManned(r *Ride) bool {
	if p.g.Levels[p.key("operateur_"+r.ID())] > 0 {
		return true
	}
	pl := p.g.Player
	return p.g.Cur == p && pl != nil && r.Booth().Grow(30).Has(pl.Pos)
}

// RideHired : un employé est embauché (sans lui, la file peut se vider de mécontents).
func (p *Place) RideHired(r *Ride) bool { return p.g.Levels[p.key("operateur_"+r.ID())] > 0 }

// ---------------------------------------------------------------- caisses

// ParcReg : une caisse de l'entrée.
type ParcReg struct {
	N       int // 1 à 10
	Pos     V   // la cabine
	Queue   []*Visitor
	T       float64
	Cashier *Agent
}

// ParcRegX : la cabine n° n (de 1 à 10), du centre vers les côtés.
func ParcRegX(n int) float64 {
	k := n - 1
	off := float64((k+1)/2) * ParcRegGap
	if k%2 == 1 {
		off = -off
	}
	return ParcAvX + off
}

// QueuePos : la place n° k dans la file (vers la route, sur deux colonnes).
func (r *ParcReg) QueuePos(k int) V {
	return V{r.Pos.X - 22 + float64(k%2)*44, r.Pos.Y + 80 + float64(k/2)*36}
}

// Zone : où se tient la caissière (ou le joueur), derrière la cabine.
func (r *ParcReg) Zone() Rect { return Rect{r.Pos.X - 60, r.Pos.Y - 130, 120, 90} }

// ---------------------------------------------------------------- visiteurs

const (
	vCar    = iota // dans la voiture
	vToReg         // vers les caisses
	vRegQ          // dans la file d'une caisse
	vToRide        // vers une attraction
	vQueue         // dans la file d'une attraction
	vRiding        // à bord
	vLeave         // vers sa voiture
	vGone
)

// Visitor : un visiteur du parc.
type Visitor struct {
	Pos, Face V
	Look      int
	WalkT     float64
	Moving    bool
	Joy       float64 // jauge de bonheur (0 à 100)
	State     int
	Trott     bool // il roule en trottinette
	path      []V
	ride      *Ride
	reg       *ParcReg
	done      uint32 // attractions déjà faites
	veh       *Vehicle
	exitT     float64
	shopped   uint32  // stands déjà passés
	Holding   int     // ce qu'il a acheté (1 + rang du stand ; 0 : rien)
	HoldT     float64 // combien de temps il le garde encore
}

// OnTrott : il roule en trottinette en ce moment (pas dans une file).
func (v *Visitor) OnTrott() bool { return v.Trott && (v.State == vToRide || v.State == vLeave) }

// ---------------------------------------------------------------- le lieu

// Parc : ce qui est propre au parc (les autres lieux n'en ont pas).
type Parc struct {
	Rides    []*Ride
	Regs     []*ParcReg
	Visitors []*Visitor
	trottOut int
	Refused  int
	Stands   []*Stand
	// StandSales : ventes des stands (badges)
	StandSales int
	cars       map[*Vehicle]int // visiteurs pas encore revenus à chaque voiture
}

func newParc(g *Game) *Place {
	p := &Place{ID: "parc", Name: "Le parc d'attractions", W: ParcW, H: ParcH, Top: ParcTop, g: g, rng: rand.New(rand.NewSource(77)), Rep: 3, Price: 1, SellMult: 1}
	p.Parc = &Parc{cars: map[*Vehicle]int{}}
	pc := p.Parc
	p.Gate = V{ParcAvX, roadY + 40}
	p.SpawnL, p.SpawnR = V{-250, laneE}, V{ParcW + 250, laneW}
	k := p.key

	// les attractions (une station chacune : « attr_train »… construite à l'achat)
	for i := range RideKinds {
		rk := &RideKinds[i]
		r := &Ride{Kind: rk, Plot: ParcSlot(rk.Slot), Top: rk.Slot < 4}
		pc.Rides = append(pc.Rides, r)
		st := &Station{ID: "attr_" + rk.ID, Name: rk.Name, Kind: Trash, Unlock: "attr_" + rk.ID}
		st.Body, st.Zone = Rect{-9000, -9000, 0, 0}, Rect{-9000, -9000, 0, 0}
		p.Stations = append(p.Stations, st)
		c := r.Track().C()
		p.Unlocks = append(p.Unlocks, &Unlock{ID: "attr_" + rk.ID, Name: rk.Name, Cost: rk.Cost, Zone: Rect{c.X - 100, c.Y - 70, 200, 140},
			Hint: "Une attraction : les visiteurs viennent pour elle (1 à 2 par place à bord)."})
	}
	// les caisses (la première est ouverte ; les suivantes s'achètent)
	for n := 1; n <= ParcRegMax; n++ {
		reg := &ParcReg{N: n, Pos: V{ParcRegX(n), ParcRegY}}
		pc.Regs = append(pc.Regs, reg)
		st := &Station{ID: "caisse" + itoa(n), Name: "Caisse " + itoa(n), Kind: Trash, Built: n == 1}
		st.Body, st.Zone = Rect{-9000, -9000, 0, 0}, Rect{-9000, -9000, 0, 0}
		if n > 1 {
			st.Unlock = "caisse" + itoa(n)
			prev := p.Station("caisse" + itoa(n-1))
			p.Unlocks = append(p.Unlocks, &Unlock{ID: st.Unlock, Name: st.Name, Cost: math.Round(40000 * math.Pow(1.45, float64(n-2))),
				Zone: Rect{reg.Pos.X - 80, reg.Pos.Y - 60, 160, 120}, Requires: func() bool { return prev.Built },
				Locked: "Ouvre d'abord la caisse d'avant", Secret: true, Hint: "Une caisse de plus : l'attente à l'entrée diminue."})
		}
		p.Stations = append(p.Stations, st)
	}

	// améliorations au sol
	spot := func(id, label string, z Rect, built string, done func(int) string) {
		p.Spots = append(p.Spots, &Spot{ID: id, Label: label, Zone: z, NeedsBuilt: built, Done: done})
	}
	for _, reg := range pc.Regs {
		n := reg.N
		p.ups = append(p.ups, &Upgrade{Key: k("caissiere" + itoa(n)), Name: "Caissière (caisse " + itoa(n) + ")", Max: 1,
			Costs: []float64{math.Round(25000 * (1 + 0.5*float64(n-1)))},
			Desc:  func(int) string { return "Elle encaisse les entrées à ta place." }})
		z := reg.Zone()
		spot("caissiere"+itoa(n), "Caissière", Rect{z.X - 5, z.Y - 150, 130, 124}, "caisse"+itoa(n), func(int) string { return "Caissière embauchée !" })
	}
	p.ups = append(p.ups,
		&Upgrade{Key: k("parking"), Name: "Parking +10", Max: (ParcParkMax - parcParkStart) / 10, Now: true, Base: 9000, Mult: 1.13,
			Desc: func(l int) string {
				return itoa(p.ParcParkTotalAt(l)) + " places de parking (un visiteur sans place repart mécontent)."
			}},
		&Upgrade{Key: k("trott"), Name: "Trottinettes +10", Max: ParcTrottMax / 10, Now: true, Base: 6000, Mult: 1.075,
			Desc: func(l int) string {
				return itoa(10*l) + " trottinettes électriques : les visiteurs qui en ont une vont plus vite d'une attraction à l'autre."
			}},
	)
	spot("parking", "Parking +10", Rect{ParcAvX - 1150, ParcPlazaY + 150, 130, 124}, "", func(l int) string {
		return itoa(p.ParcParkTotalAt(l)) + " places de parking !"
	})
	spot("trott", "Trottinettes +10", Rect{ParcRackX + ParcRackW + 20, ParcRackTop + 60, 130, 124}, "", func(l int) string {
		return itoa(10*l) + " trottinettes !"
	})
	for _, r := range pc.Rides {
		r := r
		id := r.ID()
		rk := r.Kind
		p.ups = append(p.ups,
			&Upgrade{Key: k("operateur_" + id), Name: "Employé (" + rk.Name + ")", Max: 1, Costs: []float64{math.Round(rk.Cost * 0.35)},
				Desc: func(int) string { return "Il fait monter et descendre les visiteurs à ta place." }},
			&Upgrade{Key: k("taille_" + id), Name: rk.Name + " : taille", Max: 9, Now: false, Base: math.Round(rk.Cost * 0.45), Mult: 1.55,
				Desc: func(l int) string {
					return "Taille " + itoa(l) + " : un tour plus long et plus beau (+" + itoa(int(RideJoy(l))) + " de bonheur par tour)."
				}},
			&Upgrade{Key: k("wagons_" + id), Name: rk.Name + " : " + unitName(rk), Max: rk.Units - 1, Now: true, Base: math.Round(rk.Cost * 0.18), Mult: 1.3,
				Desc: func(l int) string {
					return itoa(p.RideSeats(r)) + " places à bord. Pas plus de wagons que la taille ne le permet."
				}},
		)
		b := r.Booth()
		y := r.Edge() + r.in()*60 - 62
		spot("operateur_"+id, "Employé", Rect{b.X + 140, y, 130, 124}, "attr_"+id, func(int) string { return "Employé embauché !" })
		spot("taille_"+id, "Taille +1", Rect{b.X + 290, y, 130, 124}, "attr_"+id, func(l int) string {
			return rk.Name + " : taille " + itoa(l+1) + " !"
		})
		spot("wagons_"+id, unitLabel(rk), Rect{b.X + 440, y, 130, 124}, "attr_"+id, func(int) string {
			return rk.Name + " : " + itoa(p.RideSeats(r)) + " places à bord !"
		})
	}

	// parking : trois allées de deux rangées de 50 places. Il grandit allée par
	// allée (la plus proche de la route d'abord), du centre vers les côtés.
	for _, a := range ParcAisles {
		for c := 0; c < 50; c++ {
			x := ParcAvX + (float64(c)-24.5)*parcSpotW
			p.Parks = append(p.Parks, &PSpot{Pos: V{x, a - 105}, Aisle: a}, &PSpot{Pos: V{x, a + 105}, Aisle: a})
		}
	}
	sort.SliceStable(p.Parks, func(i, j int) bool {
		a, b := p.Parks[i], p.Parks[j]
		if a.Aisle != b.Aisle {
			return a.Aisle < b.Aisle
		}
		da, db := math.Round(math.Abs(a.Pos.X-ParcAvX)), math.Round(math.Abs(b.Pos.X-ParcAvX))
		if da != db {
			return da < db
		}
		if a.Pos.X != b.Pos.X {
			return a.Pos.X < b.Pos.X
		}
		return a.Pos.Y < b.Pos.Y
	})
	p.addStands()
	p.setupYard(-620)
	// l'agrandissement du parking : dans la zone du joueur
	if sp := p.spot("parking"); sp != nil {
		sp.Zone = Rect{p.Yard.X + 10, yardRow1, 130, 124}
	}
	return p
}

func unitName(rk *RideKind) string {
	switch rk.ID {
	case "bateaux":
		return "bateaux"
	case "splash":
		return "rondins"
	case "volante":
		return "nacelles"
	case "vide":
		return "rangées"
	}
	return "wagons"
}

func unitLabel(rk *RideKind) string {
	switch rk.ID {
	case "bateaux":
		return "Bateau +1"
	case "splash":
		return "Rondin +1"
	case "volante":
		return "Nacelles +2"
	case "vide":
		return "Rangée +1"
	}
	return "Wagon +1"
}

// ParcParkTotalAt : places de parking au niveau l de l'amélioration.
func (p *Place) ParcParkTotalAt(l int) int { return min(ParcParkMax, parcParkStart+10*l) }

// ParcParkTotal, ParcTrottTotal : ce qu'il y a aujourd'hui.
func (p *Place) ParcParkTotal() int  { return p.ParcParkTotalAt(p.g.Levels[p.key("parking")]) }
func (p *Place) ParcTrottTotal() int { return min(ParcTrottMax, 10*p.g.Levels[p.key("trott")]) }

// ParcAccessX, ParcExitX : les routes d'accès (à gauche) et de sortie (à droite)
// du parking ; elles s'écartent quand le parking grandit.
func (p *Place) ParcAccessX() float64 {
	x := ParcAvX
	for _, s := range p.Parks[:p.ParcParkTotal()] {
		x = math.Min(x, s.Pos.X)
	}
	return x - 140
}

func (p *Place) ParcExitX() float64 {
	x := ParcAvX
	for _, s := range p.Parks[:p.ParcParkTotal()] {
		x = math.Max(x, s.Pos.X)
	}
	return x + 140
}

// ParkUsedParc : places de parking occupées.
func (p *Place) ParkUsedParc() int {
	n := 0
	for _, s := range p.Parks[:p.ParcParkTotal()] {
		if s.used {
			n++
		}
	}
	return n
}

// ParcTrottFree : trottinettes au râtelier de l'entrée.
func (p *Place) ParcTrottFree() int { return max(0, p.ParcTrottTotal()-p.Parc.trottOut) }

// ParcRegOpen : la caisse n° n est ouverte.
func (p *Place) ParcRegOpen(r *ParcReg) bool {
	s := p.Station("caisse" + itoa(r.N))
	return s != nil && s.Built
}

// ParcRegManned : quelqu'un encaisse à cette caisse.
func (p *Place) ParcRegManned(r *ParcReg) bool {
	if p.g.Levels[p.key("caissiere"+itoa(r.N))] > 0 {
		return true
	}
	pl := p.g.Player
	return p.g.Cur == p && pl != nil && r.Zone().Has(pl.Pos)
}

// ParcPlayerHelps : le joueur se tient près d'une caissière embauchée (il l'aide).
func (p *Place) ParcPlayerHelps(r *ParcReg) bool {
	pl := p.g.Player
	return r.Cashier != nil && p.g.Cur == p && pl != nil && pl.Pos.Dist(r.Cashier.Pos) < HelpDist
}

// ParcFee : le prix d'entrée (il monte avec le nombre d'attractions ouvertes).
func (p *Place) ParcFee() float64 {
	n := 0
	for _, r := range p.Parc.Rides {
		if p.RideBuilt(r) {
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return math.Round(parcFeeBase * (1 + 0.6*float64(n-1)) * p.SellMult * p.g.priceMult(p))
}

// parcFeeBase : le prix d'entrée avec une seule attraction (réglé pour que le
// parc rapporte, comme le veut sa place sur la carte, environ deux fois plus
// que la pâtisserie).
var parcFeeBase = 3000.0

// ParcSeats : places à bord de toutes les attractions ouvertes.
func (p *Place) ParcSeats() int {
	n := 0
	for _, r := range p.Parc.Rides {
		if p.RideBuilt(r) {
			n += p.RideSeats(r)
		}
	}
	return n
}

// ParcTarget : visiteurs voulus en même temps (1 par place à bord à 1 étoile,
// 2 à 5 étoiles ; la réclame en fait venir plus).
func (p *Place) ParcTarget() int {
	return int(float64(p.ParcSeats()) * (1 + (p.Rep-1)/4) * (1 + p.ptsF("pts_reclame")))
}

// ParcVisitors : visiteurs dans le parc (ou sur son parking).
func (p *Place) ParcVisitors() int { return len(p.Parc.Visitors) }

// ---------------------------------------------------------------- simulation

const (
	parcWalk     = 115.0 // vitesse d'un visiteur à pied
	parcJoyStart = 60.0
	parcJoyWait  = 6.0 // jauge perdue par minute d'attente
	parcJoyNext  = 75.0
)

func (p *Place) stepParc(dt float64, active bool) {
	g := p.g
	pc := p.Parc
	if active && g.Player != nil {
		p.interact(g.Player, dt) // (achats au sol)
	}
	p.stepVehicles(dt)
	p.parcStaff()
	// nouveaux visiteurs
	p.spawnT -= dt
	if p.spawnT <= 0 {
		p.spawnT = 1.2 + 1.6*p.rand()
		if p.ParcSeats() > 0 && len(pc.Visitors) < p.ParcTarget() {
			p.parcSpawn()
		}
	}
	for _, r := range pc.Regs {
		p.stepParcReg(r, dt)
	}
	for _, s := range pc.Stands {
		s.Sold += dt
	}
	for _, r := range pc.Rides {
		if p.RideBuilt(r) {
			p.stepRide(r, dt)
		}
	}
	alive := pc.Visitors[:0]
	for _, v := range pc.Visitors {
		p.stepVisitor(v, dt)
		if v.State != vGone {
			alive = append(alive, v)
		}
	}
	pc.Visitors = alive
}

// parcStaff : les caissières et les employés des attractions (ils ne bougent pas).
func (p *Place) parcStaff() {
	g := p.g
	pc := p.Parc
	for _, r := range pc.Regs {
		if g.Levels[p.key("caissiere"+itoa(r.N))] > 0 && r.Cashier == nil {
			z := r.Zone()
			r.Cashier = &Agent{Kind: CashierAgent, Pos: V{z.C().X, z.Y + 40}, Face: V{0, 1}, Look: 900 + r.N*37}
			g.emit(Event{Kind: EvHire, Pos: r.Cashier.Pos})
		}
	}
	for _, r := range pc.Rides {
		if p.RideHired(r) && r.Operator == nil {
			b := r.Booth()
			r.Operator = &Agent{Kind: WorkerAgent, Pos: b.C(), Face: V{0, -r.in()}, Look: 1200 + r.Kind.Slot*41}
			g.emit(Event{Kind: EvHire, Pos: b.C()})
		}
	}
}

// parcSpawn : une voiture arrive (1 à 4 visiteurs) ; sans place, elle repart.
func (p *Place) parcSpawn() {
	pc := p.Parc
	var spot *PSpot
	for _, s := range p.Parks[:p.ParcParkTotal()] {
		if !s.used {
			spot = s
			break
		}
	}
	if spot == nil {
		if p.ParcParkTotal() < ParcParkMax {
			pc.Refused++
			p.g.Stats.Missed++
			p.reputation(-0.01)
			p.g.emit(Event{Kind: EvAngry, Pos: V{p.ParcAccessX(), roadY + 200}})
		}
		return
	}
	spot.used = true
	n := 1 + p.rng.Intn(4)
	veh := &Vehicle{Kind: VCar, Look: p.rng.Intn(1 << 20), Rider: true, spot: spot, speed: 320}
	if p.rng.Intn(2) == 0 {
		veh.Pos = V{ParcW + 250, laneW}
	} else {
		veh.Pos = V{-250, laneE}
	}
	ax := p.ParcAccessX()
	veh.path = []V{{ax, veh.Pos.Y}, {ax, spot.Aisle}, {spot.Pos.X, spot.Aisle}, spot.Pos}
	veh.Dir = veh.path[0].Sub(veh.Pos).Norm()
	p.Vehicles = append(p.Vehicles, veh)
	pc.cars[veh] = n
	for i := 0; i < n; i++ {
		v := &Visitor{Pos: spot.Pos, Look: p.rng.Intn(1 << 20), Joy: parcJoyStart, State: vCar, veh: veh}
		v.exitT = float64(i) * 0.5
		pc.Visitors = append(pc.Visitors, v)
	}
}

// visitorSpeed : à pied, ou en trottinette (points « Mécaniciens : trottinettes »).
func (p *Place) visitorSpeed(v *Visitor) float64 {
	if v.OnTrott() {
		return parcWalk * 1.9 * (1 + p.ptsF("pts_trott"))
	}
	return parcWalk
}

// follow : avance le long du chemin ; vrai une fois arrivé au bout.
func (v *Visitor) follow(speed, dt float64) bool {
	for len(v.path) > 0 {
		d := v.path[0].Sub(v.Pos)
		step := speed * dt
		if d.Len() <= step {
			v.Pos = v.path[0]
			v.path = v.path[1:]
			dt -= d.Len() / speed
			if dt <= 0 {
				break
			}
			continue
		}
		v.Face = d.Norm()
		v.Pos = v.Pos.Add(v.Face.Mul(step))
		v.Moving = true
		v.WalkT += dt
		return false
	}
	v.Moving = false
	return len(v.path) == 0
}

// toBoulevard : le chemin d'un point du parc jusqu'à un autre, par le boulevard.
func toBoulevard(from, to V) []V {
	return []V{{from.X, ParcBoulY + 40*math.Sin(from.X/300)}, {to.X, ParcBoulY + 40*math.Sin(to.X/300)}, to}
}

func (p *Place) stepVisitor(v *Visitor, dt float64) {
	pc := p.Parc
	if v.Holding > 0 {
		if v.HoldT -= dt; v.HoldT <= 0 {
			v.Holding = 0
		}
	}
	if v.State == vToRide || v.State == vLeave {
		p.parcShop(v)
	}
	switch v.State {
	case vCar:
		if v.veh == nil || v.veh.State != 1 {
			return
		}
		v.exitT -= dt
		if v.exitT > 0 {
			return
		}
		// il descend et va aux caisses (passage piéton)
		s := v.veh.spot
		v.Pos = s.Pos
		v.path = []V{{s.Pos.X, s.Aisle - 40}, {ParcAvX + 40*float64(v.Look%3-1), s.Aisle - 40}, {ParcAvX, roadY + 180}, {ParcAvX, roadY - 20}}
		v.State = vToReg
	case vToReg:
		if v.follow(parcWalk, dt) {
			v.reg = p.parcPickReg(v)
			v.reg.Queue = append(v.reg.Queue, v)
			v.State = vRegQ
		}
	case vRegQ:
		// il avance dans la file (sa place est donnée par son rang)
		for i, o := range v.reg.Queue {
			if o == v {
				v.path = []V{v.reg.QueuePos(i)}
				break
			}
		}
		v.follow(parcWalk*0.8, dt)
	case vToRide:
		if v.follow(p.visitorSpeed(v), dt) {
			r := v.ride
			if v.Trott {
				r.Trotts++ // il pose sa trottinette à l'entrée
			}
			r.Queue = append(r.Queue, v)
			v.State = vQueue
		}
	case vQueue:
		r := v.ride
		v.Joy = math.Max(0, v.Joy-parcJoyWait*dt/60)
		for i, o := range r.Queue {
			if o == v {
				v.path = []V{r.QueuePos(i)}
				break
			}
		}
		v.follow(parcWalk*0.7, dt)
		// une attraction sans employé : à bout, il s'en va (et le parc perd de la réputation)
		if v.Joy <= 0 && !p.RideHired(r) {
			for i, o := range r.Queue {
				if o == v {
					r.Queue = append(r.Queue[:i], r.Queue[i+1:]...)
					break
				}
			}
			p.reputation(-0.03)
			p.g.Stats.Angry++
			p.g.emit(Event{Kind: EvAngry, Pos: v.Pos})
			if v.Trott {
				r.Trotts--
			}
			p.parcNext(v, v.Pos, r)
		}
	case vRiding:
		if v.ride != nil && v.ride.State == 1 {
			v.Joy = math.Min(100, v.Joy) // (le tour lui-même ne fait pas attendre)
		}
	case vLeave:
		if v.Trott && v.Pos.Y >= ParcRackY-5 {
			pc.trottOut-- // il rend sa trottinette au râtelier, en sortant
			v.Trott = false
		}
		if v.follow(p.visitorSpeed(v), dt) {
			v.State = vGone
			p.Served++
			p.g.Stats.Served++
			if veh := v.veh; veh != nil {
				pc.cars[veh]--
				if pc.cars[veh] <= 0 {
					delete(pc.cars, veh)
					p.parcDepart(veh)
				}
			}
		}
	}
}

// parcPickReg : la caisse ouverte à la file la plus courte (une caisse sans
// caissière compte pour une longue file : on l'évite, mais celui qui y va y reste).
func (p *Place) parcPickReg(v *Visitor) *ParcReg {
	var best *ParcReg
	bs := math.Inf(1)
	for _, r := range p.Parc.Regs {
		if !p.ParcRegOpen(r) {
			continue
		}
		sc := float64(len(r.Queue)) + math.Abs(r.Pos.X-v.Pos.X)/400
		if !p.ParcRegManned(r) {
			sc += 8
		}
		if sc < bs {
			best, bs = r, sc
		}
	}
	return best
}

// stepParcReg : la caissière (ou le joueur) encaisse l'entrée du premier de la file.
func (p *Place) stepParcReg(r *ParcReg, dt float64) {
	if len(r.Queue) == 0 || !p.ParcRegManned(r) {
		r.T = 0
		return
	}
	v := r.Queue[0]
	if v.Pos.Dist(r.QueuePos(0)) > 12 {
		return
	}
	r.T += dt * (1 + p.ptsF("pts_caisse"))
	if p.ParcPlayerHelps(r) {
		r.T += dt // le joueur aide la caissière : deux fois plus vite
	}
	if r.T < 2.2 {
		return
	}
	r.T = 0
	r.Queue = r.Queue[1:]
	p.g.earn(p.ParcFee(), r.Pos, p.ID)
	// une trottinette, s'il en reste
	if p.ParcTrottFree() > 0 {
		p.Parc.trottOut++
		v.Trott = true
	}
	v.Pos = V{r.Pos.X, r.Pos.Y - 70}
	p.parcNext(v, v.Pos, nil)
}

// parcNext : la prochaine attraction (la première est obligatoire ; ensuite,
// seulement si sa jauge est haute) ou la sortie.
func (p *Place) parcNext(v *Visitor, from V, after *Ride) {
	var best *Ride
	if after == nil || v.Joy >= parcJoyNext {
		bs := math.Inf(1)
		for i, r := range p.Parc.Rides {
			if !p.RideBuilt(r) || v.done&(1<<uint(i)) != 0 || r == after {
				continue
			}
			// attente estimée : la file, au rythme de l'attraction, plus le trajet
			per := p.RideTime(r) + 8
			wait := float64(len(r.Queue))/float64(max(1, p.RideSeats(r)))*per + from.Dist(r.Entrance())/(parcWalk*1.5)
			if !p.RideManned(r) {
				wait += 120
			}
			if wait < bs {
				best, bs = r, wait
			}
		}
	}
	if best != nil {
		v.ride = best
		v.State = vToRide
		v.path = append(p.parcFromPlace(from), toBoulevard(V{from.X, ParcBoulY}, best.Entrance())...)
		return
	}
	// la sortie : par le boulevard, l'avenue, entre les caisses, la route, sa voiture
	v.ride = nil
	v.State = vLeave
	gapX := ParcAvX + float64(v.Look%5-2)*ParcRegGap + ParcRegGap/2
	path := p.parcFromPlace(from)
	path = append(path, V{from.X, ParcBoulY}, V{ParcAvX, ParcBoulY}, V{ParcAvX, ParcRackY}, V{gapX, ParcRegY - 40}, V{gapX, roadY - 20}, V{ParcAvX, roadY + 180})
	if s := v.veh; s != nil && s.spot != nil {
		path = append(path, V{ParcAvX, s.spot.Aisle - 40}, V{s.spot.Pos.X, s.spot.Aisle - 40}, s.spot.Pos)
	}
	v.path = path
}

// parcFromPlace : des caisses, on monte d'abord l'avenue.
func (p *Place) parcFromPlace(from V) []V {
	if from.Y > ParcPlazaY-100 {
		return []V{{ParcAvX + float64(int(from.X)%7-3)*20, ParcRackY}, {ParcAvX, ParcBoulY}}
	}
	return nil
}

// parcDepart : tout le monde est remonté, la voiture repart.
func (p *Place) parcDepart(veh *Vehicle) {
	veh.State, veh.Rider = 2, true
	s := veh.spot
	ex := p.ParcExitX()
	veh.path = []V{{s.Pos.X, s.Aisle}, {ex, s.Aisle}, {ex, laneE}, {ParcW + 300, laneE}}
}

// stepRide : embarquement, tour, débarquement.
func (p *Place) stepRide(r *Ride, dt float64) {
	if r.Indep() {
		p.stepBoats(r, dt)
		return
	}
	manned := p.RideManned(r)
	board := 1.0 / (1 + p.ptsF("pts_embarq")) // (points « Opérateurs : embarquement »)
	switch r.State {
	case 0: // embarquement
		r.Prog = 0
		if !manned {
			return
		}
		seats := p.RideSeats(r)
		r.boardT += dt
		if len(r.Queue) > 0 && len(r.Riders) < seats {
			v := r.Queue[0]
			if v.Pos.Dist(r.QueuePos(0)) < 30 && r.boardT >= board {
				r.boardT = 0
				r.Queue = r.Queue[1:]
				r.Riders = append(r.Riders, v)
				v.State = vRiding
				v.Pos = r.Booth().C()
				r.idleT = 0
			}
			return
		}
		if len(r.Riders) == 0 {
			return
		}
		r.idleT += dt
		if len(r.Riders) >= seats || r.idleT > 4 {
			r.State, r.T = 1, 0
		}
	case 1: // le tour
		r.T += dt
		r.Prog = math.Min(1, r.T/p.RideTime(r))
		if r.T >= p.RideTime(r) {
			r.State, r.T, r.Prog = 2, 0, 1
		}
	case 2: // débarquement
		if !manned {
			for _, v := range r.Riders {
				v.Joy = math.Max(0, v.Joy-parcJoyWait*dt/60)
			}
			return
		}
		r.boardT += dt
		if len(r.Riders) == 0 {
			r.State, r.Prog = 0, 0
			return
		}
		if r.boardT < board*0.6 {
			return
		}
		r.boardT = 0
		v := r.Riders[0]
		r.Riders = r.Riders[1:]
		p.parcRideDone(r, v)
	}
}

// parcRideDone : un visiteur descend ; sa jauge monte selon la taille de
// l'attraction ; il fait gagner ou perdre de la réputation au parc.
func (p *Place) parcRideDone(r *Ride, v *Visitor) {
	for i, o := range p.Parc.Rides {
		if o == r {
			v.done |= 1 << uint(i)
		}
	}
	v.Joy = math.Min(100, v.Joy+RideJoy(p.RideLevel(r)))
	switch {
	case v.Joy <= 0:
		p.reputation(-0.03)
		p.g.Stats.Angry++
		p.g.emit(Event{Kind: EvAngry, Pos: r.Exit()})
	case v.Joy >= 50:
		p.reputation(0.004)
		p.g.Stats.Happy++
	}
	if v.Trott {
		r.Trotts-- // il reprend sa trottinette
	}
	v.Pos = r.Exit()
	p.parcNext(v, v.Pos, r)
}

// ParcJoyAvg : la jauge moyenne des visiteurs (pour l'affichage).
func (p *Place) ParcJoyAvg() float64 {
	if len(p.Parc.Visitors) == 0 {
		return 0
	}
	s := 0.0
	for _, v := range p.Parc.Visitors {
		s += v.Joy
	}
	return s / float64(len(p.Parc.Visitors))
}

// ParcSpotUpgrade : les wagons ne dépassent pas ce que la taille permet.
func (p *Place) parcSpotMax(sp *Spot, u *Upgrade) *Upgrade {
	for _, r := range p.Parc.Rides {
		if sp.ID == "wagons_"+r.ID() {
			c := *u
			c.Max = p.RideUnitsCap(r, p.RideLevel(r)) - 1
			return &c
		}
	}
	return u
}

// États d'un visiteur, pour le dessin.
func VCarState() int    { return vCar }
func VRidingState() int { return vRiding }
func VGoneState() int   { return vGone }
func VQueueState() int  { return vQueue }
func VRegQState() int   { return vRegQ }
