package sim

// Drones de la ferme : un par étal. Chacun prend les produits de SON étal et
// les apporte au quai d'un camion. Les ouvriers ne s'occupent plus des quais ;
// le joueur peut toujours y déposer des produits pour aider.

// Drone : l'appareil d'un étal (pas sauvegardé : il repart de sa base).
type Drone struct {
	Pos   V
	Phase int // 0 : posé ; 1 : vers l'étal ; 2 : vers le quai (chargé) ; 3 : retour
	It    Item
	N     int // objets prévus puis transportés
	dock  *Station
	wait  float64 // petite pause sur la base
}

// FreightMult : un drone de fret (grange à foin, frigos) porte 5 fois plus.
const FreightMult = 5

func droneKey(it Item) string { return "drone_" + itoa(int(it)) }

// DroneSpeed : vitesse de vol. Au départ, un trajet base → étal → quai → base
// dure à peu près l'intervalle du camion : environ un objet par livraison.
func DroneSpeed(pts int) float64 { return 80 * (1 + 0.03*float64(pts)) }

// DroneCap : nombre d'objets portés (1 au départ, +1 tous les 2 points de mécaniciens).
func DroneCap(pts int) int { return 1 + pts/2 }

// DronePad : base du drone, sous son étal.
func (s *Station) DronePad() Rect {
	off := 140.0
	if s.PadOff > 0 {
		off = s.PadOff
	}
	return Rect{s.Body.X + 15, s.Body.Y + s.Body.H + off, 120, 104}
}

// hasDrones : lieux où les vitrines ont des drones.
func (p *Place) hasDrones() bool {
	return p.ID == "ferme" || p.ID == "exotique" || p.ID == "boulangerie" || p.ID == "cremerie" || p.ID == "glacier" || p.ID == "elevage" || p.ID == "boucherie" || p.ID == "port"
}

// droneDest : un camion part-il de ce lieu ? (ferme : vers un lieu acheté ;
// boulangerie et glacier : vers le supermarché)
func (p *Place) droneDest() bool {
	for _, q := range p.Stations {
		if q.Kind == Dock && p.g.DockActive(q) {
			return true
		}
	}
	return false
}

// droneOK : un drone ne s'achète que si son étal existe et qu'un camion part ;
// la soute et les hélices, qu'après le premier drone du lieu.
func (g *Game) droneOK(key string) bool {
	for _, p := range g.Places {
		if !p.hasDrones() {
			continue
		}
		for _, s := range p.Stations {
			if s.Kind == Shelf && key == p.key(droneKey(s.Item)) {
				return p.droneUseful(s)
			}
		}
	}
	return true
}

func (p *Place) droneUpgrades(scale float64) []*Upgrade {
	k := p.key
	var ups []*Upgrade
	for _, s := range p.Stations {
		if s.Kind != Shelf {
			continue
		}
		it := s.Item
		ups = append(ups, &Upgrade{Key: k(droneKey(it)), Name: "Drone : " + Items[it].Name, Max: 1,
			Costs: []float64{dronePrice[p.ID]}, // même prix pour tous les drones d'un lieu
			Desc:  func(int) string { return "Il prend les produits de cette vitrine et les porte au quai du camion." }})
	}
	return ups
}

func (p *Place) stepDrones(dt float64) {
	g := p.g
	// (plus de soute ni d'hélices à acheter : seulement les points « mécaniciens » de l'école)
	capa := DroneCap(g.Levels[p.key("pts_dr_cap")])
	speed := DroneSpeed(g.Levels[p.key("pts_dr_vit")]) * g.boostF("garage", "") // (cadeau du garagiste)
	for _, s := range p.Stations {
		if s.Kind != Shelf || !s.Built || g.Levels[p.key(droneKey(s.Item))] == 0 {
			continue
		}
		capa := capa
		if s.NoSale {
			capa *= FreightMult // grange à foin, frigos : un gros drone de fret
		}
		pad := s.DronePad().C()
		d := s.Drone
		if d == nil {
			d = &Drone{Pos: pad, wait: p.rand() * 2}
			s.Drone = d
		}
		// vol en ligne droite, à vitesse constante
		fly := func(to V) bool {
			v := to.Sub(d.Pos)
			step := speed * dt
			if v.Len() <= step {
				d.Pos = to
				return true
			}
			d.Pos = d.Pos.Add(v.Norm().Mul(step))
			return false
		}
		above := V{s.Body.C().X, s.Body.Y - 10}
		switch d.Phase {
		case 0:
			d.Pos = pad
			if d.wait > 0 {
				d.wait -= dt
				continue
			}
			if s.Stock-s.booked <= 0 {
				continue
			}
			// le quai qui en manque le plus
			var best *Station
			room := 0
			for _, q := range p.Stations {
				if q.Kind != Dock || !g.DockActive(q) || !g.dockAccepts(q, s.Item) {
					continue
				}
				if r := q.BufMax - q.Buf[s.Item] - q.claim[s.Item]; r > room {
					best, room = q, r
				}
			}
			if best == nil {
				continue
			}
			d.It, d.N, d.dock = s.Item, min(capa, room), best
			best.claim[s.Item] += d.N
			d.Phase = 1
		case 1:
			// d'abord à l'étal, prendre la marchandise
			if !fly(above) {
				continue
			}
			k := min(d.N, s.Stock-s.booked)
			d.dock.claim[d.It] = max(0, d.dock.claim[d.It]-(d.N-max(0, k)))
			if k <= 0 {
				d.N, d.Phase = 0, 3
				continue
			}
			s.Stock -= k
			d.N = k
			g.emit(Event{Kind: EvTake, Pos: above, It: d.It})
			d.Phase = 2
		case 2:
			if !fly(d.dock.Zone.C()) {
				continue
			}
			q := d.dock
			q.claim[d.It] = max(0, q.claim[d.It]-d.N)
			put := min(d.N, q.BufMax-q.Buf[d.It])
			q.Buf[d.It] += put
			s.Stock = min(s.StockMax, s.Stock+d.N-put) // le surplus revient à l'étal
			if put > 0 {
				g.emit(Event{Kind: EvDrop, Pos: q.Body.C(), It: d.It})
			}
			d.N, d.Phase = 0, 3
		case 3:
			if fly(pad) {
				d.Phase = 0
				d.wait = 0.5 + p.rand()
			}
		}
	}
}

// DroneCount : nombre de drones achetés dans ce lieu.
func (p *Place) DroneCount() int {
	n := 0
	for _, s := range p.Stations {
		if s.Kind == Shelf {
			n += p.g.Levels[p.key(droneKey(s.Item))]
		}
	}
	return n
}

// DroneCount : nombre de drones de la ferme.
func (g *Game) DroneCount() int { return g.farm().DroneCount() }

// droneUseful : le drone de cet étal aurait-il un quai où livrer ?
// (au supermarché : son rayon est ouvert ; à la boulangerie ou au glacier :
// une machine utilise ce produit). Sinon, on ne le propose pas.
func (p *Place) droneUseful(s *Station) bool {
	if !s.Built {
		return false
	}
	for _, q := range p.Stations {
		if q.Kind != Dock {
			continue
		}
		to := p.g.Place(q.To)
		if to == nil || !to.Owned {
			continue
		}
		if p.g.DockActive(q) && p.g.dockAccepts(q, s.Item) {
			return true
		}
		if to.ID == "super" {
			// son rayon est ouvert, ou on peut déjà l'acheter
			for _, u := range to.Unlocks {
				if u.ID == "r_"+itoa(int(s.Item)) && (u.Done || u.Requires == nil || u.Requires()) {
					return true
				}
			}
			continue
		}
		// le lieu s'en sert (même sans la machine : ex. le beurre, dès que la boulangerie est achetée)
		if to.pallet(s.Item) != nil {
			return true
		}
	}
	return false
}

// DroneLock : (deux lignes séparées par « | ») pourquoi le drone de cet étal n'est pas encore proposé ("" s'il l'est,
// s'il est déjà acheté, ou s'il est trop tôt pour en parler).
func (p *Place) DroneLock(sp *Spot) string {
	s := sp.Shelf
	if s == nil || !s.Built || p.g.Levels[p.key(sp.ID)] > 0 || p.droneUseful(s) || !p.droneDest() {
		return ""
	}
	if sup := p.g.Place("super"); sup == nil || !sup.Owned {
		return "Il faut le|supermarché"
	}
	return "Ouvre son rayon|au supermarché"
}

// dronePrice : prix d'un drone, le même pour tous les étals d'un lieu.
var dronePrice = map[string]float64{"ferme": 800, "boulangerie": 2500, "cremerie": 3200, "glacier": 4000, "elevage": 5000, "boucherie": 6000, "exotique": 2500, "port": 5000}

// StorePad : base (prévue) du drone d'une réserve, sous sa zone de dépôt.
func (s *Station) StorePad() Rect {
	return Rect{s.Zone.X + 10, s.Zone.Y + s.Zone.H + 8, s.Zone.W - 20, fieldY + fieldH - (s.Zone.Y + s.Zone.H + 8) - 4}
}
