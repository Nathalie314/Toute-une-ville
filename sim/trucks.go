package sim

import "math"

// Camions : chaque quai d'expédition a son camion, qui emmène ce qu'il y a sur
// le quai vers les palettes de la réserve du lieu de destination.

// Les camions d'un lieu s'améliorent avec les points « mécaniciens » de ce lieu
// (il n'y a plus d'amélioration du camion à la mairie).

// TruckInterval : secondes entre deux voyages (+1 % de départs par point).
func TruckInterval(pts int) float64 { return math.Round(250/(1+float64(pts)/100)) / 10 }

// TruckCap : nombre de produits par voyage (+2 par point).
func TruckCap(pts int) int { return 60 + 2*pts }

// DockSize : places du quai, par produit (+1 tous les 2 points de mécaniciens,
// +2 par niveau de soute des drones : le drone a toujours de quoi déposer).
func DockSize(pts, soute int) int { return 20 + pts/2 + 2*soute }

// meca : points « mécaniciens » du lieu d'un quai.
func (g *Game) meca(dock *Station) int { return g.Levels[dock.place+".pts_cam_vit"] }

// DockEvery : secondes entre deux camions de ce quai.
func (g *Game) DockEvery(dock *Station) float64 { return TruckInterval(g.meca(dock)) }

// pallet : palette d'un produit dans la réserve d'un lieu.
func (p *Place) pallet(it Item) *Station {
	for _, s := range p.Stations {
		if s.Kind == Pallet && s.Item == it {
			return s
		}
	}
	return nil
}

// usesInput : une machine construite du lieu a besoin de ce produit.
func (p *Place) usesInput(it Item) bool {
	for _, s := range p.Stations {
		if s.Kind == Machine && s.Built {
			for _, in := range s.Inputs {
				if in == it {
					return true
				}
			}
		}
	}
	return false
}

// dockAccepts : ce quai prend-il ce produit ?
// Vers le supermarché : s'il y a un rayon ouvert pour lui ;
// vers un atelier (boulangerie, glacier) : si une de ses machines s'en sert.
func (g *Game) dockAccepts(dock *Station, it Item) bool {
	to := g.Place(dock.To)
	if to == nil || !to.Owned || to.pallet(it) == nil {
		return false
	}
	if !g.palletRoom(to, it) {
		return false // la réserve est pleine : on dit « stop » à l'expéditeur
	}
	if to.ID == "super" {
		return to.shelfFor(it) != nil
	}
	return to.usesInput(it)
}

// palletRoom : la palette de destination a-t-elle encore de la place, en comptant
// ce qui attend déjà sur les quais (et ce que les drones apportent) ?
func (g *Game) palletRoom(to *Place, it Item) bool {
	pal := to.pallet(it)
	mult := DockMult(to)
	coming := 0
	for _, from := range g.Places {
		for _, q := range from.Stations {
			if q.Kind == Dock && q.To == to.ID {
				coming += q.Buf[it] + q.claim[it]
			}
		}
	}
	return pal.Pile+(coming+1)*mult <= pal.PileMax
}

// TruckLeft : secondes avant le prochain départ du camion d'un quai.
func (g *Game) TruckLeft(dock *Station) float64 {
	return math.Max(0, g.DockEvery(dock)-g.TruckT[dock.ID+"@"+dock.place])
}

// DockActive : le quai sert-il (son lieu de destination est acheté) ?
func (g *Game) DockActive(dock *Station) bool {
	to := g.Place(dock.To)
	if to == nil || !to.Owned {
		return false
	}
	// au supermarché : il faut au moins un rayon ouvert pour un produit de ce lieu
	// (sinon le camion de la boulangerie tournait à vide)
	if to.ID == "super" {
		from := g.Place(dock.place)
		if from == nil {
			return true
		}
		for _, s := range from.Stations {
			if s.Kind == Shelf && to.shelfFor(s.Item) != nil {
				return true
			}
		}
		return false
	}
	return true
}

// DockMult : au supermarché, chaque produit livré en devient 3
// (il achète en gros : sinon ses rayons se vident bien plus vite qu'on ne les remplit).
func DockMult(to *Place) int {
	if to != nil && to.ID == "super" {
		return 3 // (4.1 : 1 objet livré = 3 en réserve ; avant : 2, et au tout début 10)
	}
	return 1
}

func (g *Game) stepTrucks(dt float64) {
	if g.TruckT == nil {
		g.TruckT = map[string]float64{}
	}
	for _, from := range g.Places {
		if !from.Owned {
			continue
		}
		pv, pc := g.Levels[from.key("pts_cam_vit")], g.Levels[from.key("pts_cam_cap")]
		every, capa := TruckInterval(pv), TruckCap(pc)
		for _, dock := range from.Stations {
			if dock.Kind != Dock {
				continue
			}
			dock.BufMax = DockSize(pc, 0)
			if !g.DockActive(dock) {
				continue
			}
			key := dock.ID + "@" + dock.place
			g.TruckT[key] += dt * g.boostF("garage", "") // (cadeau du garagiste)
			if g.TruckT[key] < every {
				continue
			}
			g.TruckT[key] = 0
			to := g.Place(dock.To)
			mult := DockMult(to) // au supermarché : 1 objet livré = 10 en réserve
			moved := 0
			for moved < capa {
				any := false
				for it := Item(0); it < NItems && moved < capa; it++ {
					pal := to.pallet(it)
					if dock.Buf[it] > 0 && pal != nil && pal.Pile+mult <= pal.PileMax { // rien ne se perd
						dock.Buf[it]--
						pal.Pile += mult
						moved++
						any = true
					}
				}
				if !any {
					break
				}
			}
			dock.lastFull = moved >= capa
			if moved > 0 {
				g.Stats.Delivered += moved * mult
				g.emit(Event{Kind: EvTruck, Amount: float64(moved), Text: to.ID})
			}
		}
	}
}

// PalletMax : contenu d'une palette de réserve, selon le niveau de l'entrepôt
// (ateliers : 40 → 100 par 5 ; supermarché : 300 → 600 par 25).
func PalletMax(place string, l int) int {
	if place == "super" {
		return 300 + 25*l
	}
	return 40 + 5*l
}

// syncPallets : applique le niveau de l'entrepôt aux palettes.
func (g *Game) syncPallets() {
	for _, p := range g.Places {
		if !p.Owned {
			continue
		}
		m := PalletMax(p.ID, g.Levels[p.key("entrepot")])
		for _, s := range p.Stations {
			if s.Kind == Pallet {
				s.PileMax = m
			}
		}
	}
}
