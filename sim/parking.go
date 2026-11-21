package sim

import "math"

// Parking : sous la route de chaque boutique. Les clients arrivent en voiture,
// à vélo, à moto ou en trottinette, se garent, traversent au passage piéton
// et repartent avec leur véhicule. S'il n'y a plus de place, ils viennent à pied.

type VehKind int

const (
	VCar VehKind = iota
	VBike
	VMoto
	VTrot
)

// Vehicle : un véhicule de client (ou de passage, sans client).
type Vehicle struct {
	Kind   VehKind
	Pos    V
	Dir    V
	Look   int
	Rider  bool // quelqu'un est dessus (ou dedans)
	Moving bool
	State  int // 0 arrive, 1 garé, 2 repart
	T      float64
	path   []V
	spot   *PSpot
	speed  float64
	waitT  float64
	cust   *Customer
	Hidden bool // garée dans le parking à étages (on ne la voit plus)
}

// PSpot : une place de parking.
type PSpot struct {
	Pos   V
	Bike  bool    // place pour deux-roues
	Aisle float64 // supermarché : l'allée devant la place
	Etage bool    // supermarché : dans le parking à étages
	used  bool
}

// Used : la place est prise.
func (s *PSpot) Used() bool { return s.used }

// géométrie (la route va de roadY à roadY+150)
const (
	laneW     = roadY + 40  // voie vers la gauche
	laneE     = roadY + 110 // voie vers la droite
	ParkTop   = roadY + 190 // haut du parking
	ParkRow1  = roadY + 255 // centre des places du haut
	ParkAisle = roadY + 350 // allée
	ParkRow2  = roadY + 445 // centre des places du bas
	ParkBot   = roadY + 500
	ParkH     = ParkBot + 50 // hauteur du lieu avec son parking
)

// ParkX : le parking est centré sur le portail (sans dépasser du terrain).
func (p *Place) ParkX() (x0, x1, gate float64) {
	h := p.parkHalf
	if h == 0 {
		h = 580
	}
	c := math.Max(h+60, math.Min(p.W-h-60, p.Gate.X))
	if n := p.ParksBought(); n < len(p.Parks) {
		// crêperie : le parking grandit avec ses achats (de l'herbe avant), autour du portail
		far := 80.0
		for _, s := range p.Parks[:n] {
			far = math.Max(far, math.Abs(s.Pos.X-p.Gate.X))
		}
		return p.Gate.X - far - 130, p.Gate.X + far + 130, p.Gate.X
	}
	return c - h, c + h, p.Gate.X
}

// ParkEntry, ParkExit : les deux entrées (sens unique, de gauche à droite).
func (p *Place) ParkEntry() float64 { x0, _, _ := p.ParkX(); return x0 + 60 }
func (p *Place) ParkExit() float64  { _, x1, _ := p.ParkX(); return x1 - 60 }

func (p *Place) setupParking() {
	p.H = ParkH
	x0, x1, gx := p.ParkX()
	// le coin des deux-roues, à droite du passage (à gauche s'il n'y a pas la place)
	b0, b1 := gx+60, gx+250
	if b1 > x1-150 {
		b0, b1 = gx-250, gx-60
	}
	for x := x0 + 140; x <= x1-130; x += 80 {
		if math.Abs(x-gx) < 80 {
			continue // le passage des piétons, face au portail
		}
		if x > b0 && x < b1 {
			continue
		}
		p.Parks = append(p.Parks, &PSpot{Pos: V{x, ParkRow1}}, &PSpot{Pos: V{x, ParkRow2}})
	}
	for x := b0 + 20; x < b1-10; x += 32 {
		p.Parks = append(p.Parks, &PSpot{Pos: V{x, ParkRow1}, Bike: true})
	}
}

// arriveBy : le client vient avec un véhicule s'il reste une place.
func (p *Place) arriveBy(c *Customer) bool {
	if len(p.Parks) == 0 {
		return false
	}
	r := p.rand()
	kind := VCar
	switch {
	case r < 0.14:
		kind = VBike
	case r < 0.24:
		kind = VMoto
	case r < 0.36:
		kind = VTrot
	}
	var free []*PSpot
	for _, s := range p.Parks {
		if !s.used && s.Bike == (kind != VCar) {
			free = append(free, s)
		}
	}
	if len(free) == 0 {
		return false
	}
	s := free[p.rng.Intn(len(free))]
	s.used = true
	v := &Vehicle{Kind: kind, Look: c.Look, Rider: true, spot: s, cust: c}
	v.speed = []float64{300, 190, 320, 170}[kind]
	entry := p.ParkEntry()
	if p.portRoad() {
		p.portArrive(v, s)
	} else if p.rng.Intn(2) == 0 {
		v.Pos = V{-250, laneE}
		v.path = []V{{entry, laneE}}
	} else {
		v.Pos = V{p.W + 250, laneW}
		v.path = []V{{entry, laneW}}
	}
	if !p.portRoad() {
		v.path = append(v.path, V{entry, ParkAisle}, V{s.Pos.X, ParkAisle}, s.Pos)
	}
	v.Dir = v.path[0].Sub(v.Pos).Norm()
	p.Vehicles = append(p.Vehicles, v)
	c.veh = v
	c.Pos = s.Pos
	return true
}

// walkFromCar / walkToCar : du véhicule au portail, et retour.
func (p *Place) walkFromCar(s *PSpot) []V {
	if p.ID == "super" {
		return p.superWalkFromCar(s)
	}
	if p.portRoad() {
		return p.portWalkFromCar(s)
	}
	_, _, gx := p.ParkX()
	return []V{{s.Pos.X, ParkAisle - 30}, {gx, ParkAisle - 30}, {gx, ParkTop - 20}, p.Gate}
}

func (p *Place) walkToCar(s *PSpot) []V {
	if p.portRoad() {
		return p.portWalkToCar(s)
	}
	_, _, gx := p.ParkX()
	return []V{p.Gate, {gx, ParkTop - 20}, {gx, ParkAisle - 30}, {s.Pos.X, ParkAisle - 30}, s.Pos}
}

// InCar : le client est encore dans son véhicule (on ne le dessine pas).
func (c *Customer) InCar() bool { return c.veh != nil && c.veh.State == 0 }

// leaveByCar : le client a fini ; au lieu de partir au bout de la route,
// il retourne à son véhicule.
func (p *Place) leaveByCar(c *Customer) {
	if p.ID == "super" || c.veh == nil || c.toCar || len(c.path) == 0 {
		return
	}
	last := c.path[len(c.path)-1]
	if last != p.SpawnL && last != p.SpawnR {
		return
	}
	// on garde le début du chemin jusqu'au portail
	var keep []V
	for _, q := range c.path[:len(c.path)-1] {
		if q == p.Gate {
			break
		}
		keep = append(keep, q)
	}
	c.path = append(keep, p.walkToCar(c.veh.spot)...)
	c.toCar = true
}

// depart : le client est remonté dans son véhicule, qui s'en va.
func (p *Place) depart(v *Vehicle) {
	if p.ID == "super" {
		p.superDepart(v)
		return
	}
	if p.portRoad() {
		p.portDepart(v)
		return
	}
	v.State, v.Rider = 2, true
	exit := p.ParkExit()
	v.path = []V{{v.spot.Pos.X, ParkAisle}, {exit, ParkAisle}}
	if p.rng.Intn(2) == 0 {
		v.path = append(v.path, V{exit, laneE}, V{p.W + 300, laneE})
	} else {
		v.path = append(v.path, V{exit, laneW}, V{-300, laneW})
	}
	v.cust = nil
}

// passing : de temps en temps, une voiture passe sans s'arrêter.
func (p *Place) passing() {
	r := p.rand()
	kind := VCar
	if r < 0.15 {
		kind = VBike
	} else if r < 0.3 {
		kind = VMoto
	}
	v := &Vehicle{Kind: kind, Look: p.rng.Intn(1 << 20), Rider: true, State: 2}
	v.speed = []float64{320, 190, 340, 170}[kind]
	if p.portRoad() {
		p.portPassing(v)
	} else if p.rng.Intn(2) == 0 {
		v.Pos, v.path = V{-250, laneE}, []V{{p.W + 300, laneE}}
	} else {
		v.Pos, v.path = V{p.W + 250, laneW}, []V{{-300, laneW}}
	}
	p.Vehicles = append(p.Vehicles, v)
}

func (p *Place) stepVehicles(dt float64) {
	if len(p.Parks) == 0 {
		return
	}
	p.passT -= dt
	if p.passT <= 0 {
		p.passT = 6 + 10*p.rand()
		p.passing()
	}
	_, _, gx := p.ParkX()
	// quelqu'un sur le passage piéton ?
	crossing := false
	onCross := func(q V) bool { return math.Abs(q.X-gx) < 85 && q.Y > roadY-20 && q.Y < roadY+175 }
	if p.portRoad() { // (le passage piéton traverse la route verticale)
		onCross = func(q V) bool {
			return math.Abs(q.Y-PortGateY) < 85 && q.X > PortRoadX-20 && q.X < PortRoadX+PortRoadW+20
		}
	}
	for _, c := range p.Customers {
		if p.ID == "super" {
			break // (au supermarché, personne ne traverse la route)
		}
		if !c.InCar() && onCross(c.Pos) {
			crossing = true
			break
		}
	}
	if p.Crep != nil {
		for _, cg := range p.Crep.Groups {
			for _, m := range cg.Members {
				if cg.Visible() && onCross(m.Pos) {
					crossing = true
				}
			}
		}
	}
	if pl := p.g.Player; pl != nil && p.g.Cur == p && onCross(pl.Pos) {
		crossing = true
	}
	alive := p.Vehicles[:0]
	for _, v := range p.Vehicles {
		v.T += dt
		if v.State == 1 || len(v.path) == 0 {
			v.Moving = false
			if v.State == 2 {
				continue // sorti de l'écran
			}
			if v.State == 0 {
				// garé : le client descend
				v.State, v.Rider = 1, false
				v.Dir = V{0, 1}
				if v.spot != nil && v.spot.Etage {
					v.Hidden = true // dans le parking à étages
				}
				if c := v.cust; c != nil {
					c.Pos = v.spot.Pos
					c.path = append(p.walkFromCar(v.spot), c.path...)
				}
			}
			alive = append(alive, v)
			continue
		}
		to := v.path[0]
		d := to.Sub(v.Pos)
		dir := d.Norm()
		// on laisse passer : piétons sur le passage, véhicule devant
		stop := false
		onRoad := math.Abs(v.Pos.Y-laneW) < 4 || math.Abs(v.Pos.Y-laneE) < 4
		if crossing && onRoad && math.Abs(dir.Y) < 0.1 && !p.portRoad() {
			ahead := (gx - v.Pos.X) * dir.X
			if ahead > 60 && ahead < 220 {
				stop = true
			}
		}
		if crossing && p.portRoad() && math.Abs(dir.X) < 0.1 && (math.Abs(v.Pos.X-PortLaneS) < 4 || math.Abs(v.Pos.X-PortLaneN) < 4) {
			ahead := (PortGateY - v.Pos.Y) * dir.Y
			if ahead > 60 && ahead < 220 {
				stop = true
			}
		}
		if !stop {
			for _, o := range p.Vehicles {
				if o == v || o.State == 1 || o.Hidden {
					continue
				}
				r := o.Pos.Sub(v.Pos)
				along := r.X*dir.X + r.Y*dir.Y
				side := math.Abs(r.X*dir.Y - r.Y*dir.X)
				if along > 0 && along < 120 && side < 30 {
					stop = true
					break
				}
			}
		}
		if stop && v.waitT < 4 {
			v.waitT += dt
			v.Moving = false
			alive = append(alive, v)
			continue
		}
		if !stop {
			v.waitT = 0
		}
		step := v.speed * dt
		if d.Len() > 1 {
			// on ralentit dans le parking
			if (!p.portRoad() && v.Pos.Y > ParkTop-30) || (p.portRoad() && v.Pos.X > PortParkX0-10) {
				step *= 0.55
			}
		}
		v.Moving = true
		if d.Len() <= step {
			v.Pos = to
			v.path = v.path[1:]
		} else {
			v.Pos = v.Pos.Add(dir.Mul(step))
		}
		if d.Len() > 0.5 {
			v.Dir = dir
		}
		if v.State == 2 && len(v.path) <= 2 && v.spot != nil {
			v.spot.used = false // la place est libre dès qu'on a quitté l'allée
			v.spot = nil
		}
		alive = append(alive, v)
	}
	p.Vehicles = alive
}

// ParksBought : places de parking achetées (toutes, sauf à la crêperie : son parking s'agrandit).
func (p *Place) ParksBought() int {
	if p.Crep != nil && p.g != nil {
		return p.CrepParkN(p.g.Levels[p.key("parking")])
	}
	return len(p.Parks)
}

// ---------------------------------------------------------------- le port : route verticale
//
// Au port, la route passe à la verticale, à droite de la criée ; le parking est de l'autre
// côté, sous la zone du joueur : trois allées horizontales, des places au-dessus et en dessous
// de chacune. On traverse la route au passage piéton, face à l'entrée de la criée.

// PortAisles : les allées du parking du port (y).
var PortAisles = []float64{1880, 2440, 3000}

func (p *Place) portRoad() bool { return p.ID == "port" }

func (p *Place) setupPortParking() {
	for ai, ay := range PortAisles {
		for x := PortParkX0 + 80.0; x <= PortParkX1-60; x += 80 {
			if ai == 0 && x < PortParkX0+330 {
				continue // (le coin des deux-roues, près de la route)
			}
			p.Parks = append(p.Parks, &PSpot{Pos: V{x, ay - 95}, Aisle: ay}, &PSpot{Pos: V{x, ay + 95}, Aisle: ay})
		}
	}
	for x := PortParkX0 + 40.0; x < PortParkX0+300; x += 32 {
		p.Parks = append(p.Parks, &PSpot{Pos: V{x, PortAisles[0] - 95}, Bike: true, Aisle: PortAisles[0]})
	}
}

// portArrive : le véhicule descend (ou monte) la route verticale, puis entre dans son allée.
func (p *Place) portArrive(v *Vehicle, s *PSpot) {
	if p.rng.Intn(2) == 0 {
		v.Pos = V{PortLaneS, PortSalTop - 500}
		v.path = []V{{PortLaneS, s.Aisle}}
	} else {
		v.Pos = V{PortLaneN, PortH + 400}
		v.path = []V{{PortLaneN, s.Aisle}}
	}
	v.path = append(v.path, V{s.Pos.X, s.Aisle}, s.Pos)
}

func (p *Place) portWalkFromCar(s *PSpot) []V {
	side := PortRoadX + PortRoadW + 40 // le trottoir, côté parking
	return []V{{s.Pos.X, s.Aisle}, {PortParkX0 - 20, s.Aisle}, {side, s.Aisle}, {side, PortGateY}, p.Gate}
}

func (p *Place) portWalkToCar(s *PSpot) []V {
	side := PortRoadX + PortRoadW + 40
	return []V{p.Gate, {side, PortGateY}, {side, s.Aisle}, {PortParkX0 - 20, s.Aisle}, {s.Pos.X, s.Aisle}, s.Pos}
}

func (p *Place) portDepart(v *Vehicle) {
	v.State, v.Rider = 2, true
	v.path = []V{{v.spot.Pos.X, v.spot.Aisle}, {PortParkX0 + 20, v.spot.Aisle}}
	if p.rng.Intn(2) == 0 {
		v.path = append(v.path, V{PortLaneN, v.spot.Aisle}, V{PortLaneN, PortSalTop - 600})
	} else {
		v.path = append(v.path, V{PortLaneS, v.spot.Aisle}, V{PortLaneS, PortH + 500})
	}
	v.cust = nil
}

func (p *Place) portPassing(v *Vehicle) {
	if p.rng.Intn(2) == 0 {
		v.Pos, v.path = V{PortLaneS, PortSalTop - 500}, []V{{PortLaneS, PortH + 500}}
	} else {
		v.Pos, v.path = V{PortLaneN, PortH + 400}, []V{{PortLaneN, PortSalTop - 600}}
	}
}
