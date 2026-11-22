package sim

import (
	"math"
	"math/rand"

	"ville/tr"
)

// LE COMMISSARIAT (événements, étape 3 ; rangée du bas de la carte, après l'école). Il
// s'ouvre, offert, en même temps que le supermarché (quand les voleurs arrivent en ville).
//
//   - Le bureau du commissaire : rien ne marche sans lui. Il ne s'achète pas avec de
//     l'argent : 10 points d'université.
//   - Les voitures de police, sur le parking : la n-ième coûte le prix du lieu n° 4(n−1) de la
//     rangée du haut, et n'existe que si ce lieu existe. La première est offerte avec le
//     commissaire ; la 2e coûte le prix du 4e lieu, la 3e du 8e, la 4e du 12e, la 5e du 16e…
//   - Les lieutenants : un par voiture, formé sur la case sous sa voiture (points du collège :
//     10, 50, 250… ×5 d'une voiture à la suivante). Le lieutenant n° k conduit la voiture n° k ; sans voiture (ou sans
//     patrouille), il erre sur le parking en sifflotant.
//   - Le parking (menu) : chaque voiture conduite par son lieutenant patrouille dans un lieu de
//     la rangée du haut. Ce lieu ne peut plus être attaqué par les voleurs ; on y voit passer,
//     de temps en temps, la voiture de police. Des mécaniciens (points du primaire : 100 fois
//     ce qu'a coûté son lieutenant en points du collège) rendent une voiture plus rapide : elle
//     protège alors un deuxième lieu. Au mieux, un lieu sur deux est protégé.
//
// Rien n'est sauvegardé à part : tout est dans Levels (commissariat.commissaire,
// .voiture_k, .lieutenant_k, .double_k, .patrouille_k_s = rang du lieu dans TopRow).

// Géométrie du commissariat.
const (
	PolBldY   = 150.0  //
	PolBldW   = 2480.0 //
	PolBldH   = 700.0  //
	PolParkY  = 1010.0 // le parking des voitures de police (sous le bâtiment)
	PolParkH  = 540.0  // (sous chaque place : la case de son lieutenant)
	PolParkM  = 270.0  // de la place à gauche et à droite des voitures (la case des patrouilles à droite)
	PolBayW   = 300.0  // une place de parking
	PolBayGap = 140.0  // (de la place entre deux voitures : le jeu reste aéré)
)

// PolW : la largeur du commissariat (le parking s'allonge avec le nombre de voitures) ; le
// bâtiment est centré.
var PolW = math.Max(3000, float64(PolCars())*(PolBayW+PolBayGap)+1000)

// PolBldX : le bord gauche du bâtiment.
var PolBldX = (PolW - PolBldW) / 2

// PolOffice : le bureau du commissaire (à gauche dans le bâtiment).
func PolOffice() Rect { return Rect{PolBldX + 60, PolBldY + 60, 700, PolBldH - 120} }

// PolHall : le hall, au milieu (l'accueil).
func PolHall() Rect { return Rect{PolBldX + 880, PolBldY + 60, 720, PolBldH - 120} }

// PolWing : l'aile de droite (les bureaux des capitaines, plus tard).
func PolWing() Rect { return Rect{PolBldX + 1720, PolBldY + 60, PolBldW - 1780, PolBldH - 120} }

// PolDoor : la porte du bâtiment (en bas, au milieu du hall).
func PolDoor() V { return V{PolHall().C().X, PolBldY + PolBldH} }

// Police : ce qui vit au commissariat (le commissaire, les lieutenants à pied).
type Police struct {
	Cops []*Cop // les lieutenants (dans l'ordre : le n° k conduit la voiture n° k)
	tuto bool
}

// Cop : un lieutenant à pied (sans patrouille, il erre sur le parking en sifflotant).
type Cop struct {
	Pos, Face V
	WalkT     float64
	Moving    bool
	Look      int
	InCar     bool // il patrouille (on ne le voit plus au commissariat)
	to        V
	waitT     float64
}

// PolCars : le nombre de voitures au plus. La voiture n° k coûte le prix du lieu n° 4(k−1) (la
// 1re : celui du « lieu 0 », gratuite ; la 2e : le 4e lieu ; la 3e : le 8e…) : elle n'existe que
// si ce lieu existe. 14 lieux : 4 voitures ; la 5e viendra avec le 16e lieu, la 6e avec le 20e…
func PolCars() int { return len(TopRow())/4 + 1 }

// PolBay : la place de parking de la voiture n° k (1, 2…), centrées sur le lieu.
func PolBay(k int) Rect {
	n := float64(PolCars())
	w := n*PolBayW + (n-1)*PolBayGap
	x0 := PolW/2 - w/2
	return Rect{x0 + float64(k-1)*(PolBayW+PolBayGap), PolParkY + 60, PolBayW, 250}
}

// PolPatrolZone : la case du parking qui ouvre le menu des patrouilles.
func PolPatrolZone() Rect {
	b := PolBay(PolCars())
	return Rect{b.X + b.W + 50, PolParkY + 150, 190, 150}
}

// LieutCost : points du collège pour former le lieutenant n° k (10, 50, 250… : ×5).
func LieutCost(k int) int { return 10 * int(math.Pow(5, float64(k-1))) }

// DoubleCost : points du primaire (mécaniciens) pour qu'une voiture protège deux lieux.
func DoubleCost(k int) int { return 100 * LieutCost(k) }

func newPolice(g *Game) *Place {
	W := PolW
	p := &Place{ID: "commissariat", Name: "Le commissariat", W: W, H: 1800, g: g, rng: rand.New(rand.NewSource(113)),
		Rep: 3, Price: 0, SellMult: 1, MaxCust: 0}
	p.Pol = &Police{}
	p.Gate = V{W / 2, roadY + 40}
	p.SpawnL = V{350, roadY + 40}
	p.SpawnR = V{W - 350, roadY + 40}
	p.Inside = V{W / 2, PolParkY + 300}
	p.RestSpot = V{W / 2, PolParkY + 300}
	k := p.key
	p.ups = []*Upgrade{
		{Key: k("commissaire"), Name: "Commissaire", Max: 1, Costs: []float64{10}, Pts: "formation",
			Desc: func(l int) string {
				if l > 0 {
					return "Il dirige le commissariat : les voitures et les lieutenants sont débloqués."
				}
				return "Sans lui, le commissariat ne fonctionne pas. Il coûte 10 points d'université (l'école). La première voiture est offerte avec lui."
			}},
	}
	for c := 1; c <= PolCars(); c++ {
		c := c
		p.ups = append(p.ups, &Upgrade{Key: k("lieutenant_" + itoa(c)), Name: "Lieutenant", Max: 1, Pts: "marketing", Costs: []float64{float64(LieutCost(c))},
			Desc: func(int) string {
				return tr.F("Formé avec des points du collège, il conduit la voiture n° {n}.", "n", itoa(c))
			}})
		cost := PlacePrice(4 * (c - 1))
		p.ups = append(p.ups, &Upgrade{Key: k("voiture_" + itoa(c)), Name: "Voiture de police", Max: 1, Costs: []float64{cost},
			Desc: func(int) string {
				return "Avec son lieutenant, elle protège un lieu des voleurs (case « Patrouilles »)."
			}})
	}
	p.Spots = append(p.Spots, &Spot{ID: "commissaire", Label: "Commissaire", Zone: Rect{PolOffice().C().X - 70, PolOffice().Y + PolOffice().H - 170, 140, 130},
		Done: func(int) string { return "Le commissaire est arrivé ! La première voiture de police est offerte." }})
	for c := 1; c <= PolCars(); c++ { // le lieutenant de chaque voiture : sa case sous la voiture
		c, b := c, PolBay(c)
		p.Spots = append(p.Spots, &Spot{ID: "lieutenant_" + itoa(c), Label: "Lieutenant", Needs: "voiture_" + itoa(c), Zone: Rect{b.C().X - 70, b.Y + b.H + 14, 140, 120},
			Done: func(int) string { return tr.F("Lieutenant n° {n} formé !", "n", itoa(c)) }})
	}
	for c := 2; c <= PolCars(); c++ {
		b := PolBay(c)
		prev := "voiture_" + itoa(c-1)
		p.Spots = append(p.Spots, &Spot{ID: "voiture_" + itoa(c), Label: "Voiture de police", Needs: prev, Zone: Rect{b.C().X - 70, b.C().Y - 65, 140, 130},
			Done: func(int) string { return "Nouvelle voiture de police !" }})
	}
	p.setupYard(-620)
	return p
}

// HasChief : le commissaire est-il là ?
func (g *Game) HasChief() bool { return g.Levels["commissariat.commissaire"] > 0 }

// PolCarOwned : la voiture n° k est-elle achetée ?
func (g *Game) PolCarOwned(k int) bool {
	return g.Levels["commissariat.voiture_"+itoa(k)] > 0 || (k == 1 && g.HasChief()) // (la 1re : offerte avec lui)
}

// PolLieutenant : le lieutenant de la voiture n° k est-il formé ?
func (g *Game) PolLieutenant(k int) bool { return g.Levels["commissariat.lieutenant_"+itoa(k)] > 0 }

// PolLieutenants : le nombre de lieutenants formés.
func (g *Game) PolLieutenants() int {
	n := 0
	for k := 1; k <= PolCars(); k++ {
		if g.PolLieutenant(k) {
			n++
		}
	}
	return n
}

// PolDriven : la voiture n° k a-t-elle son lieutenant (et le commissaire) ?
func (g *Game) PolDriven(k int) bool {
	return g.HasChief() && g.PolCarOwned(k) && g.PolLieutenant(k)
}

// PolDouble : la voiture n° k protège-t-elle deux lieux (mécaniciens) ?
func (g *Game) PolDouble(k int) bool { return g.Levels["commissariat.double_"+itoa(k)] > 0 }

// PolSlots : le nombre de lieux que la voiture n° k peut protéger.
func (g *Game) PolSlots(k int) int {
	if g.PolDouble(k) {
		return 2
	}
	return 1
}

// PolPatrol : le lieu où patrouille la voiture n° k dans sa case s (0 ou 1) ("" : aucun).
func (g *Game) PolPatrol(k, s int) string {
	r := g.Levels["commissariat.patrouille_"+itoa(k)+"_"+itoa(s)]
	top := TopRow()
	if r < 1 || r > len(top) || s >= g.PolSlots(k) {
		return ""
	}
	return top[r-1]
}

// SetPatrol : la voiture n° k patrouille dans ce lieu (case s) ; "" : elle rentre au parking.
// Un lieu n'a qu'une patrouille : s'il en avait une autre, elle rentre.
func (g *Game) SetPatrol(k, s int, id string) {
	if id != "" {
		for c := 1; c <= PolCars(); c++ {
			for t := 0; t < 2; t++ {
				if g.PolPatrol(c, t) == id {
					delete(g.Levels, "commissariat.patrouille_"+itoa(c)+"_"+itoa(t))
				}
			}
		}
	}
	key := "commissariat.patrouille_" + itoa(k) + "_" + itoa(s)
	if r := PriceRank(id); id != "" && r > 0 {
		g.Levels[key] = r
	} else {
		delete(g.Levels, key)
	}
}

// PatrolOf : la voiture qui protège ce lieu (0 : aucune). Il faut qu'elle soit conduite.
func (g *Game) PatrolOf(id string) int {
	for c := 1; c <= PolCars(); c++ {
		if !g.PolDriven(c) {
			continue
		}
		for s := 0; s < g.PolSlots(c); s++ {
			if g.PolPatrol(c, s) == id {
				return c
			}
		}
	}
	return 0
}

// Patrolled : ce lieu est-il protégé par une patrouille (les voleurs n'y viennent pas) ?
func (g *Game) Patrolled(id string) bool { return g.PatrolOf(id) > 0 }

// PolOnPatrol : la voiture n° k est-elle partie en patrouille (conduite, et un lieu choisi) ?
func (g *Game) PolOnPatrol(k int) bool {
	if !g.PolDriven(k) {
		return false
	}
	for s := 0; s < g.PolSlots(k); s++ {
		if g.PolPatrol(k, s) != "" {
			return true
		}
	}
	return false
}

// BuyDouble : les mécaniciens de la voiture n° k (points du primaire).
func (g *Game) BuyDouble(k int) bool {
	c := DoubleCost(k)
	if g.PolDouble(k) || !g.PolCarOwned(k) || g.Points("ouvrier") < c {
		return false
	}
	g.Levels["pts.ouvrier"] -= c
	g.Levels["commissariat.double_"+itoa(k)] = 1
	return true
}

// stepPolice : le commissariat s'ouvre avec le supermarché ; ses lieutenants errent sur le
// parking quand ils ne patrouillent pas.
func (p *Place) stepPolice(dt float64, active bool) {
	g := p.g
	if g.HasChief() && g.Levels[p.key("voiture_1")] == 0 {
		g.Levels[p.key("voiture_1")] = 1 // (la première voiture est offerte avec le commissaire)
	}
	pol := p.Pol
	for len(pol.Cops) < PolCars() {
		pol.Cops = append(pol.Cops, nil)
	}
	for k := 1; k <= PolCars(); k++ { // (le lieutenant n° k : Cops[k-1])
		if pol.Cops[k-1] == nil && g.PolLieutenant(k) {
			c := &Cop{Pos: PolBay(k).C().Add(V{0, 200}), Face: V{0, 1}, Look: 700 + 53*k}
			c.to = c.Pos
			pol.Cops[k-1] = c
			g.emit(Event{Kind: EvHire, Pos: c.Pos, Place: p.ID})
		}
	}
	for i, c := range pol.Cops {
		if c == nil {
			continue
		}
		c.InCar = g.PolOnPatrol(i + 1)
		if c.InCar {
			c.Moving = false
			continue
		}
		if c.waitT > 0 {
			c.waitT -= dt
			c.Moving = false
			continue
		}
		d := c.to.Sub(c.Pos)
		if d.Len() < 6 { // une autre place du parking, puis une pause en sifflotant
			c.to = V{PolBay(1).X - 120 + p.rng.Float64()*(PolBay(PolCars()).X+PolBayW-PolBay(1).X+240), PolParkY + 60 + p.rng.Float64()*(PolParkH-120)}
			c.waitT = 2 + 4*p.rng.Float64()
			continue
		}
		step := 140 * dt
		c.Moving = true
		c.WalkT += dt
		c.Face = d.Norm()
		if d.Len() <= step {
			c.Pos = c.to
		} else {
			c.Pos = c.Pos.Add(c.Face.Mul(step))
		}
	}
	if active && g.Player != nil {
		p.interact(g.Player, dt)
	}
}

// PolTutoSeen : le petit tutoriel du commissariat a-t-il déjà été proposé ?
func (g *Game) PolTutoSeen() bool { return g.Levels["commissariat.tuto"] > 0 }

// SetPolTutoSeen : il ne sera plus proposé.
func (g *Game) SetPolTutoSeen() { g.Levels["commissariat.tuto"] = 1 }

// openPolice : le commissariat s'ouvre (offert) en même temps que le supermarché.
func (g *Game) openPolice() {
	pol := g.Place("commissariat")
	if pol == nil || pol.Owned || !g.thievesActive() {
		return
	}
	pol.Owned = true
	g.emit(Event{Kind: EvLevel, Text: "Le commissariat ouvre ses portes (offert) ! Il lui faut un commissaire."})
}
