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
	PolBldY    = 480.0                                 // le bâtiment, seul au-dessus de la route
	PolBldH    = 60 + PolOffH + PolCorH + PolOffH + 60 // deux rangées de pièces, un couloir entre les deux
	PolCorH    = 200.0                                 // le couloir (bureaux des capitaines, garde à vue)
	PolOffGap  = 50.0                                  // entre deux bureaux de capitaines
	PolCellGap = 34.0                                  // entre deux cellules
	PolParkY   = 1880.0                                // sous la route : la zone du joueur à gauche, le parking des voitures à droite
	PolYardX   = 300.0                                 // le bord gauche de la zone du joueur (à l'horizontale, 2150 de large)
	PolParkX0  = PolYardX + 2150 + 200                 // le bord gauche du parking
	PolParkH   = 540.0                                 // (sous chaque place : la case de son lieutenant)
	PolParkM   = 270.0                                 // de la place à gauche et à droite des voitures (la case des patrouilles à droite)
	PolBayW    = 300.0                                 // une place de parking
	PolBayGap  = 140.0                                 // (de la place entre deux voitures : le jeu reste aéré)
	PolRoomW   = 640.0                                 // le bureau du commissaire, le hall
	PolOffW    = 300.0                                 // un bureau de capitaine (deux rangées)
	PolOffH    = 330.0                                 //
	PolCellW   = 170.0                                 // une cellule de garde à vue (deux rangées de cinq)
	PolCellH   = 330.0                                 //
	PolCellN   = 10                                    // cellules au plus (quel que soit le nombre de bureaux)
	PolCellCap = 5                                     // voleurs par cellule
	PolGap     = 60.0                                  // l'épaisseur des murs entre les pièces
)

// survRank : le rang (dans la rangée du haut) du lieu que protège en théorie le bureau n° n :
// 3, puis 6 et 7, 10 et 11, 14 et 15… (les voitures : 1 et 2, 4 et 5, 8 et 9…). Le bureau
// coûte le prix de ce lieu ; la cellule n° n, la moitié.
func survRank(n int) int {
	c := 0
	for r := 3; ; r++ {
		if r == 3 || (r >= 6 && (r%4 == 2 || r%4 == 3)) {
			if c++; c == n {
				return r
			}
		}
	}
}

// PolOffices : le nombre de bureaux de capitaines (un par lieu « à surveiller » qui existe).
func PolOffices() int {
	n := 0
	for survRank(n+1) <= len(TopRow()) {
		n++
	}
	return n
}

func polOffCols() int { return max(1, (PolOffices()+1)/2) }

// polSide : la largeur d'une aile (les bureaux à gauche, les cellules à droite : la plus large des deux).
func polSide() float64 {
	return math.Max(float64(polOffCols())*(PolOffW+PolOffGap)-PolOffGap, 5*(PolCellW+PolCellGap)-PolCellGap)
}

// PolBldW : la largeur du bâtiment : au centre, le commissaire (en haut) et le hall (en bas) ;
// à gauche, les bureaux des capitaines ; à droite, la garde à vue.
var PolBldW = 60 + polSide() + PolGap + PolRoomW + PolGap + polSide() + 60

// PolW : la largeur du commissariat (le bâtiment, la zone du joueur et le parking).
var PolW = math.Max(PolBldW+600, PolParkX0+2*PolParkM+float64(PolCars())*(PolBayW+PolBayGap)-PolBayGap+300)

// PolBldX : le bord gauche du bâtiment.
var PolBldX = (PolW - PolBldW) / 2

func polCX() float64 { return PolBldX + PolBldW/2 }

// PolOffice : le bureau du commissaire : au sommet, au centre (le symbole de sa mission).
func PolOffice() Rect { return Rect{polCX() - PolRoomW/2, PolBldY + 60, PolRoomW, PolOffH} }

// PolHall : le hall, sous son bureau (la porte, l'accueil, la case de la vidéosurveillance).
func PolHall() Rect {
	return Rect{polCX() - PolRoomW/2, PolBldY + 60 + PolOffH + 20, PolRoomW, PolCorH + PolOffH - 20} // (jusqu'en bas : ses portes de côté donnent sur les couloirs)
}

// PolWing : l'aile des bureaux des capitaines (à gauche).
func PolWing() Rect {
	w := float64(polOffCols())*(PolOffW+PolOffGap) - PolOffGap
	return Rect{polCX() - PolRoomW/2 - PolGap - w, PolBldY + 60, w, PolBldH - 120}
}

// PolCapOffice : le bureau n° n (1, 2…) : du centre vers la gauche, en haut puis en bas.
func PolCapOffice(n int) Rect {
	w := PolWing()
	c, r := (n-1)/2, (n-1)%2
	return Rect{w.X + w.W - PolOffW - float64(c)*(PolOffW+PolOffGap), w.Y + float64(r)*(PolOffH+PolCorH), PolOffW, PolOffH}
}

// PolCells : la garde à vue (à droite).
func PolCells() Rect {
	return Rect{polCX() + PolRoomW/2 + PolGap, PolBldY + 60, 5*(PolCellW+PolCellGap) - PolCellGap, PolBldH - 120}
}

// PolCell : la cellule n° n (1 à 10), du centre vers la droite, en haut puis en bas.
func PolCell(n int) Rect {
	a := PolCells()
	c, r := (n-1)%5, (n-1)/5
	return Rect{a.X + float64(c)*(PolCellW+PolCellGap), a.Y + float64(r)*(PolCellH+PolCorH), PolCellW, PolCellH}
}

// PolVideoZone : la case du hall qui ouvre le menu de la vidéosurveillance.
func PolVideoZone() Rect { h := PolHall(); return Rect{h.C().X - 110, h.Y + h.H - 190, 220, 150} }

// PolCorridorY : le haut des couloirs (entre les deux rangées de bureaux et de cellules).
func PolCorridorY() float64 { return PolBldY + 60 + PolOffH }

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
	x0 := PolParkX0 + PolParkM
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
	p := &Place{ID: "commissariat", Name: "Le commissariat", W: W, H: PolParkY + PolParkH + 130, Top: PolBldY - 150, g: g, rng: rand.New(rand.NewSource(113)),
		Rep: 3, Price: 0, SellMult: 1, MaxCust: 0}
	p.Pol = &Police{}
	p.Gate = V{PolDoor().X, roadY + 40}
	p.SpawnL = V{350, roadY + 40}
	p.SpawnR = V{W - 350, roadY + 40}
	p.Inside = PolDoor().Add(V{0, -200})
	p.RestSpot = PolBay(1).C().Add(V{0, 200})
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
	p.Spots = append(p.Spots, &Spot{ID: "commissaire", Label: "Commissaire", Zone: Rect{PolOffice().X + PolOffice().W - 180, PolOffice().Y + PolOffice().H - 150, 140, 130},
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
	// les bureaux des capitaines (de l'argent : le prix du lieu qu'ils surveillent en théorie),
	// leurs capitaines (des points du lycée), les cellules de garde à vue (moitié prix)
	for n := 1; n <= PolOffices(); n++ {
		n, r := n, PolCapOffice(n)
		prev := "commissaire"
		if n > 1 {
			prev = "bureau_" + itoa(n-1)
		}
		p.ups = append(p.ups, &Upgrade{Key: k("bureau_" + itoa(n)), Name: "Bureau de capitaine", Max: 1, Costs: []float64{PlacePrice(survRank(n))},
			Desc: func(int) string {
				return "Un capitaine y surveillera un lieu par vidéosurveillance : il peut faire arrêter les voleurs."
			}})
		p.ups = append(p.ups, &Upgrade{Key: k("capitaine_" + itoa(n)), Name: "Capitaine", Max: 1, Pts: "comptable", Costs: []float64{float64(CaptCost(n))},
			Desc: func(int) string {
				return "Formé avec des points du lycée, il surveille un lieu par vidéosurveillance (case « Vidéosurveillance » du hall)."
			}})
		p.Spots = append(p.Spots, &Spot{ID: "bureau_" + itoa(n), Label: "Bureau", Needs: prev, Zone: Rect{r.C().X - 70, r.C().Y - 65, 140, 130},
			Done: func(int) string { return tr.F("Bureau n° {n} construit !", "n", itoa(n)) }})
		p.Spots = append(p.Spots, &Spot{ID: "capitaine_" + itoa(n), Label: "Capitaine", Needs: "bureau_" + itoa(n), Zone: Rect{r.C().X - 70, r.Y + r.H - 140, 140, 130},
			Done: func(int) string { return tr.F("Capitaine n° {n} formé !", "n", itoa(n)) }})
	}
	for n := 1; n <= PolCellN; n++ {
		n, r := n, PolCell(n)
		prev := "commissaire"
		if n > 1 {
			prev = "cellule_" + itoa(n-1)
		}
		p.ups = append(p.ups, &Upgrade{Key: k("cellule_" + itoa(n)), Name: "Cellule de garde à vue", Max: 1, Costs: []float64{math.Round(PlacePrice(survRank(n)) / 2)},
			Desc: func(int) string {
				return "Elle garde 5 voleurs arrêtés pendant 30 minutes (même jeu fermé) : pendant ce temps, ils n'attaquent plus."
			}})
		p.Spots = append(p.Spots, &Spot{ID: "cellule_" + itoa(n), Label: "Cellule", Needs: prev, Zone: Rect{r.C().X - 70, r.C().Y - 65, 140, 130},
			Done: func(int) string { return "Nouvelle cellule de garde à vue !" }})
	}
	p.setupYardH(PolYardX+650, PolParkY) // (de l'autre côté de la route, comme au port)
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
	if id != "" { // (une voiture arrive : le capitaine qui surveillait ce lieu est libéré)
		for c := 1; c <= PolOffices(); c++ {
			if g.PolCamera(c) == id {
				delete(g.Levels, "commissariat.camera_"+itoa(c))
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
	if n := g.Levels[p.key("lieutenants")]; n > 0 { // (une partie d'avant : un lieutenant par voiture)
		for k := 1; k <= min(n, PolCars()); k++ {
			g.Levels[p.key("lieutenant_"+itoa(k))] = 1
		}
		delete(g.Levels, p.key("lieutenants"))
	}
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

// ---------------------------------------------------------------- vidéosurveillance, garde à vue

// CaptCost : points du lycée pour former le capitaine n° n (50, 500, 5 000… : ×10).
func CaptCost(n int) int { return 50 * int(math.Pow(10, float64(n-1))) }

// PolOfficeOwned : le bureau n° n est-il construit ?
func (g *Game) PolOfficeOwned(n int) bool { return g.Levels["commissariat.bureau_"+itoa(n)] > 0 }

// PolCaptain : le capitaine du bureau n° n est-il formé ?
func (g *Game) PolCaptain(n int) bool { return g.Levels["commissariat.capitaine_"+itoa(n)] > 0 }

// PolWatching : le capitaine n° n travaille-t-il (commissaire, bureau, capitaine) ?
func (g *Game) PolWatching(n int) bool { return g.HasChief() && g.PolOfficeOwned(n) && g.PolCaptain(n) }

// PolCamera : le lieu que surveille le capitaine n° n ("" : aucun).
func (g *Game) PolCamera(n int) string {
	r := g.Levels["commissariat.camera_"+itoa(n)]
	top := TopRow()
	if r < 1 || r > len(top) {
		return ""
	}
	return top[r-1]
}

// SetCamera : le capitaine n° n surveille ce lieu ("" : aucun). Un capitaine par lieu.
func (g *Game) SetCamera(n int, id string) {
	if id != "" && g.PatrolAssigned(id) > 0 {
		return // (un lieu confié à une voiture de police n'a pas besoin de capitaine)
	}
	if id != "" {
		for c := 1; c <= PolOffices(); c++ {
			if g.PolCamera(c) == id {
				delete(g.Levels, "commissariat.camera_"+itoa(c))
			}
		}
	}
	if r := PriceRank(id); id != "" && r > 0 {
		g.Levels["commissariat.camera_"+itoa(n)] = r
	} else {
		delete(g.Levels, "commissariat.camera_"+itoa(n))
	}
}

// CameraOf : le capitaine qui surveille ce lieu (0 : aucun).
func (g *Game) CameraOf(id string) int {
	for n := 1; n <= PolOffices(); n++ {
		if g.PolWatching(n) && g.PolCamera(n) == id {
			return n
		}
	}
	return 0
}

// VideoMax : les points du primaire au plus dans un bureau (99 % de plus : 100 %).
const VideoMax = 990

// CatchChance : la chance que le capitaine n° n fasse arrêter les voleurs (1 %, +1 % tous les
// 10 points du primaire placés dans SON bureau, 100 % au plus).
func (g *Game) CatchChance(n int) float64 {
	return math.Min(1, 0.01+float64(g.Levels["commissariat.video_"+itoa(n)])/1000)
}

// BuyVideo : des techniciens pour le bureau n° n (pts points du primaire, par 10).
func (g *Game) BuyVideo(n, pts int) bool {
	key := "commissariat.video_" + itoa(n)
	pts = min(pts, VideoMax-g.Levels[key])
	if pts <= 0 || !g.PolOfficeOwned(n) || g.Points("ouvrier") < pts {
		return false
	}
	g.Levels["pts.ouvrier"] -= pts
	g.Levels[key] += pts
	return true
}

// PolCellsOwned : le nombre de cellules de garde à vue.
func (g *Game) PolCellsOwned() int {
	n := 0
	for c := 1; c <= PolCellN; c++ {
		if g.Levels["commissariat.cellule_"+itoa(c)] > 0 {
			n++
		}
	}
	return n
}

// Custody : des voleurs en garde à vue (relâchés au bout de 30 minutes, même jeu fermé).
type Custody struct {
	N    int
	Left float64
}

// CustodyTime : 30 minutes (demande ; 2 heures avant).
const CustodyTime = 30 * 60.0

// InCustody : le nombre de voleurs en garde à vue.
func (g *Game) InCustody() int {
	n := 0
	for _, c := range g.Custody {
		n += c.N
	}
	return n
}

// CellRoom : les places libres en garde à vue.
func (g *Game) CellRoom() int { return max(0, g.PolCellsOwned()*PolCellCap-g.InCustody()) }

// arrest : les voleurs repérés sont arrêtés à leur départ (s'il y a de la place en cellule).
func (g *Game) arrest(id string, n int, left float64) {
	at := V{}
	if p := g.Place(id); p != nil {
		at = p.Gate
	}
	k := min(n, g.CellRoom())
	if k <= 0 {
		g.emit(Event{Kind: EvLevel, Place: id, Pos: at, Text: "La vidéosurveillance a repéré les voleurs, mais il n'y a pas de place en garde à vue : ils restent libres."})
		return
	}
	g.Custody = append(g.Custody, &Custody{N: k, Left: left})
	g.Levels["commissariat.arrestations"] += k
	g.emit(Event{Kind: EvLevel, Place: id, Pos: at, Text: tr.F("Grâce à la vidéosurveillance, {n} voleur(s) arrêté(s) en partant : garde à vue pendant 30 minutes.", "n", itoa(k))})
}

// stepCustody : le temps de la garde à vue s'écoule ; au bout, les voleurs sont relâchés.
func (g *Game) stepCustody(dt float64) {
	keep := g.Custody[:0]
	for _, c := range g.Custody {
		if c.Left -= dt; c.Left > 0 {
			keep = append(keep, c)
		} else {
			g.emit(Event{Kind: EvLevel, Text: tr.F("Fin de la garde à vue : {n} voleur(s) relâché(s).", "n", itoa(c.N))})
		}
	}
	g.Custody = keep
}

// GardeSave : la garde à vue, dans la sauvegarde.
type GardeSave struct {
	N     int
	Reste float64
}

func (g *Game) custodySave() []GardeSave {
	var out []GardeSave
	for _, c := range g.Custody {
		out = append(out, GardeSave{N: c.N, Reste: c.Left})
	}
	return out
}

// PatrolAssigned : la voiture à qui ce lieu est confié, même sans son lieutenant (0 : aucune).
func (g *Game) PatrolAssigned(id string) int {
	for k := 1; k <= PolCars(); k++ {
		for s := 0; s < g.PolSlots(k); s++ {
			if g.PolPatrol(k, s) == id {
				return k
			}
		}
	}
	return 0
}

// RankPrice : le prix d'achat du lieu n° r de la rangée du haut (son vrai prix s'il existe ;
// sinon celui que donnera la courbe des prix : prévu pour les lieux à venir).
func (g *Game) RankPrice(r int) float64 {
	if top := TopRow(); r >= 1 && r <= len(top) {
		if q := g.Place(top[r-1]); q != nil {
			return q.Price
		}
	}
	return PlacePrice(r)
}

// policePrices : les voitures, les bureaux et les cellules au prix réel des lieux de la carte
// (PlacePrice donne 0 avant le 4e lieu : le 1er bureau, au prix de la crèmerie, était gratuit).
func (g *Game) policePrices() {
	p := g.Place("commissariat")
	if p == nil {
		return
	}
	for _, u := range p.ups {
		var n int
		switch {
		case scanN(u.Key, "commissariat.voiture_", &n):
			u.Costs = []float64{g.RankPrice(4 * (n - 1))}
		case scanN(u.Key, "commissariat.bureau_", &n):
			u.Costs = []float64{g.RankPrice(survRank(n))}
		case scanN(u.Key, "commissariat.cellule_", &n):
			u.Costs = []float64{math.Round(g.RankPrice(survRank(n)) / 2)}
		}
	}
}

// scanN : « préfixe » suivi d'un nombre ?
func scanN(key, prefix string, n *int) bool {
	if len(key) <= len(prefix) || key[:len(prefix)] != prefix {
		return false
	}
	v := 0
	for _, c := range key[len(prefix):] {
		if c < '0' || c > '9' {
			return false
		}
		v = v*10 + int(c-'0')
	}
	*n = v
	return true
}
