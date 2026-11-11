package sim

import "math"

// Le parking du supermarché (à droite, entre les deux routes) et ses caddies.
// Tous les clients viennent en voiture et prennent un caddie. S'il n'y a plus
// de place ou plus de caddie, le client repart mécontent.
// Au départ : 10 places et 10 caddies ; on en achète 10 par 10 (300 au plus).
// Au-delà des places au sol, les voitures montent dans le parking à étages.

// géométrie du parking
const (
	superTopW    = 40.0  // route du haut : voie vers la gauche
	superTopE    = 80.0  // voie vers la droite
	superSpotW   = 85.0  // largeur d'une place
	superGround  = 168   // places au sol (8 rangées de 21)
	SuperEtageN  = 70    // places par étage du parking à étages
	superCartTop = 600.0 // haut de la colonne des caddies (rangement)
)

// SuperAisles : les allées du parking au sol (une rangée de places de chaque côté).
var SuperAisles = []float64{330, 650, 970, 1290}

// SuperAccessX : la route d'accès au parking, entre les caddies et le parking.
func SuperAccessX() float64 { return SuperCartX + 410 }

// Sens unique (plus aucun croisement) : les voitures arrivent par la route du bas,
// montent la route d'accès (à gauche du parking), roulent vers la droite dans les
// allées et ressortent par la route de sortie (à droite du parking), qui monte
// jusqu'à la route du haut ; elles repartent vers la droite.

// SuperExitX : la route de sortie, à droite des places au sol.
func SuperExitX() float64 { return SuperParkX + 21*superSpotW + 140 }

// SuperEtageX : le parking à étages, de l'autre côté de la route de sortie.
func SuperEtageX() float64 { return SuperExitX() + 110 }

// SuperEtageDoor : l'entrée (voitures) et la porte (piétons) du parking à étages.
func SuperEtageDoor() V { return V{SuperEtageX(), SuperAisles[2]} }

func (p *Place) setupSuperParking() {
	// places au sol, les plus proches des caddies d'abord
	for c := 0; c < 21; c++ {
		x := SuperParkX + 60 + float64(c)*superSpotW
		for _, a := range SuperAisles {
			p.Parks = append(p.Parks, &PSpot{Pos: V{x, a - 105}, Aisle: a}, &PSpot{Pos: V{x, a + 105}, Aisle: a})
		}
	}
	// puis les étages (leurs places sont « dans » le bâtiment)
	for i := superGround; i < SuperMaxPark; i++ {
		p.Parks = append(p.Parks, &PSpot{Pos: SuperEtageDoor(), Aisle: SuperAisles[2], Etage: true})
	}
}

// ParkPlaces : places de parking achetées (au niveau l de l'amélioration).
func (p *Place) ParkPlaces(l int) int { return min(SuperMaxPark, superStart+10*l) }

// CartsTotalAt : caddies achetés (au niveau l).
func (p *Place) CartsTotalAt(l int) int { return min(SuperMaxPark, superStart+10*l) }

// CartsTotal, ParkTotal : ce qu'il y a aujourd'hui.
func (p *Place) CartsTotal() int { return p.CartsTotalAt(p.g.Levels[p.key("caddies")]) }
func (p *Place) ParkTotal() int  { return p.ParkPlaces(p.g.Levels[p.key("parking")]) }

// ParkUsed : places occupées.
func (p *Place) ParkUsed() int {
	n := 0
	for _, s := range p.Parks[:min(len(p.Parks), p.ParkTotal())] {
		if s.used {
			n++
		}
	}
	return n
}

// Etages : nombre d'étages du parking à étages (0 : pas encore construit).
func (p *Place) Etages() int {
	n := p.ParkTotal() - superGround
	if n <= 0 {
		return 0
	}
	return (n + SuperEtageN - 1) / SuperEtageN
}

// SuperMaxCust : clients en même temps = rayons ouverts × 5 (+ marketing).
func (p *Place) SuperMaxCust() int {
	n := 0
	for _, s := range p.Stations {
		if s.Kind == Shelf && s.Built {
			n++
		}
	}
	return min(SuperMaxClients, int(float64(n*5)*(1+p.ptsF("pts_marketing"))))
}

// superSpawn : un client essaie de venir.
func (p *Place) superSpawn() {
	// ce qu'il veut acheter : entre 3 et (rayons ouverts × 2) articles
	var shelves []*Station
	open, free := 0, 0
	for _, s := range p.Stations {
		if s.Kind == Shelf && s.Built {
			open++
			if s.Stock-s.booked > 0 {
				shelves = append(shelves, s)
				free += s.Stock - s.booked
			}
		}
	}
	// il ne vient que s'il y a au moins 3 articles libres en rayon (sinon il
	// repartirait avec 1 ou 2 objets : mieux vaut moins de clients, mais des caddies pleins)
	if free < 3 {
		return
	}
	// une place de parking et un caddie ?
	var spot *PSpot
	for _, s := range p.Parks[:p.ParkTotal()] {
		if !s.used {
			spot = s
			break
		}
	}
	// parking et caddies déjà au maximum : personne de plus ne vient (il n'y a
	// plus rien à acheter, ce n'est pas la faute du joueur : pas de mécontent)
	if (spot == nil && p.ParkTotal() >= SuperMaxPark) || (p.CartsFree() <= 0 && p.CartsTotal() >= SuperMaxPark) {
		return
	}
	if spot == nil || p.CartsFree() <= 0 {
		// il repart aussitôt, mécontent
		p.Refused++
		p.g.Stats.Missed++
		p.reputation(-0.02)
		at := V{SuperAccessX(), 200}
		if spot != nil {
			at = V{SuperCartX + 150, superCartTop}
			p.RefusedCart++
		} else {
			p.RefusedPark++
		}
		p.g.emit(Event{Kind: EvAngry, Pos: at})
		return
	}
	c := &Customer{Look: p.rng.Intn(1 << 20)}
	total := 3
	if open*2 > 3 {
		total = 3 + p.rng.Intn(open*2-2)
	}
	p.rng.Shuffle(len(shelves), func(i, j int) { shelves[i], shelves[j] = shelves[j], shelves[i] })
	for _, s := range shelves {
		if total <= 0 {
			break
		}
		n := min(total, 1+p.rng.Intn(4), s.Stock-s.booked)
		if n <= 0 {
			continue
		}
		s.booked += n
		total -= n
		c.Wants = append(c.Wants, Want{It: s.Item, N: n, Booked: n})
	}
	// le reste de sa liste : sur les rayons ouverts, sans réserver (il prendra ce
	// qu'il trouve à son passage, ou attendra un peu que le magasinier remplisse)
	if total > 0 {
		var all []*Station
		for _, s := range p.Stations {
			if s.Kind == Shelf && s.Built {
				all = append(all, s)
			}
		}
		p.rng.Shuffle(len(all), func(i, j int) { all[i], all[j] = all[j], all[i] })
		for total > 0 {
			for _, s := range all {
				if total <= 0 {
					break
				}
				n := min(total, 1+p.rng.Intn(3))
				total -= n
				found := false
				for i := range c.Wants {
					if c.Wants[i].It == s.Item {
						c.Wants[i].N += n
						found = true
					}
				}
				if !found {
					c.Wants = append(c.Wants, Want{It: s.Item, N: n})
				}
			}
		}
	}
	// on entre par la droite : on visite les rayons de droite à gauche
	for i := range c.Wants {
		for j := i + 1; j < len(c.Wants); j++ {
			if p.shelfFor(c.Wants[j].It).Spot.X > p.shelfFor(c.Wants[i].It).Spot.X {
				c.Wants[i], c.Wants[j] = c.Wants[j], c.Wants[i]
			}
		}
	}
	p.cartsOut++
	c.cartRes = true
	spot.used = true
	v := &Vehicle{Kind: VCar, Look: c.Look, Rider: true, spot: spot, cust: c, speed: 300}
	ax, ex := SuperAccessX(), SuperExitX()
	// on arrive par la route du bas (voie la plus proche), on monte la route d'accès
	v.Pos = V{p.W + 250, laneW}
	if spot.Etage {
		// parking à étages : par l'allée du bas, puis la route de sortie jusqu'à sa porte
		low := SuperAisles[len(SuperAisles)-1]
		v.path = []V{{ax, laneW}, {ax, low}, {ex, low}, {ex, SuperEtageDoor().Y}, SuperEtageDoor()}
	} else {
		v.path = []V{{ax, laneW}, {ax, spot.Aisle}, {spot.Pos.X, spot.Aisle}, spot.Pos}
	}
	v.Dir = v.path[0].Sub(v.Pos).Norm()
	p.Vehicles = append(p.Vehicles, v)
	c.veh = v
	c.Pos = spot.Pos
	// à pied : vers les caddies, le sas, puis dans le magasin
	c.path = []V{p.cartSpot, p.Gate, V{SuperBX1 + 10, p.Gate.Y}, p.Inside}
	p.Customers = append(p.Customers, c)
}

// superWalkFromCar : de la voiture à la route d'accès (le reste est dans c.path).
func (p *Place) superWalkFromCar(s *PSpot) []V {
	ax := SuperAccessX()
	if s.Etage {
		return []V{{ax + 70, s.Aisle}, {ax - 60, s.Aisle}}
	}
	return []V{{s.Pos.X, s.Aisle - 40}, {ax - 60, s.Aisle - 40}}
}

// superLeavePath : après la caisse, il ressort par le sas, rend son caddie et
// retourne à sa voiture.
func (p *Place) superLeavePath(c *Customer) []V {
	path := []V{c.Pos.Add(V{0, 60}), p.Inside, V{SuperBX1 + 10, p.Gate.Y}, p.Gate, p.cartSpot}
	if c.veh == nil || c.veh.spot == nil {
		return append(path, V{SuperAccessX(), 1500})
	}
	s := c.veh.spot
	ax := SuperAccessX()
	if s.Etage {
		return append(path, V{ax - 60, s.Aisle}, V{ax + 70, s.Aisle}, s.Pos)
	}
	return append(path, V{ax - 60, s.Aisle - 40}, V{s.Pos.X, s.Aisle - 40}, s.Pos)
}

// superDepart : la voiture recule dans l'allée et repart par une des deux routes.
func (p *Place) superDepart(v *Vehicle) {
	v.State, v.Rider = 2, true
	s := v.spot
	ex := SuperExitX()
	v.Hidden = false
	if s.Etage {
		v.Pos = SuperEtageDoor()
		v.path = []V{{ex, s.Aisle}}
	} else {
		// on recule dans l'allée, on continue vers la droite jusqu'à la sortie
		v.path = []V{{s.Pos.X, s.Aisle}, {ex, s.Aisle}}
	}
	// la route de sortie monte jusqu'à la route du haut ; on repart vers la droite
	v.path = append(v.path, V{ex, superTopE}, V{p.W + 300, superTopE})
	v.cust = nil
}

// superCustSpeed : les caddies (points « mécaniciens ») font marcher plus vite.
func (p *Place) superCustSpeed() float64 { return 1 + p.ptsF("pts_caddie") }

// superRegister : la caisse où aller (file courte, pas trop loin).
func (p *Place) superRegister(c *Customer) *Station {
	var best *Station
	bs := math.Inf(1)
	for _, r := range p.Registers() {
		sc := float64(p.QueueAt(r)) + c.Pos.Dist(r.QueueHead)/700
		if !p.manned(r) {
			sc += 4
		}
		if sc < bs {
			best, bs = r, sc
		}
	}
	return best
}
