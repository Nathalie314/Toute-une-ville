package sim

import (
	"math"
	"math/rand"
	"sort"
	"strings"

	"ville/tr"
)

// LES INCENDIES (événements, dès la boulangerie). Un lieu qui a des « zones chaudes » (fours,
// plaques de cuisson ; seulement celles qui sont achetées) peut prendre feu : 10 points par
// zone chaude dans le tirage du groupe. Le feu prend selon la prévention du lieu (100 % de
// risque ; −1 % tous les 10 PPI ; 10 % au moins). Sinon : « un début d'incendie… ».
//
// Un incendie dure 1 minute (comme un vol : tout le monde fuit, le joueur court à sa zone). Les
// camions de pompiers (avec leur pompier) sauvent chacun 1 élément (+1 tous les 1 000 points du
// primaire) : la caissière, puis les employés, puis les objets du plus cher au moins cher. Le
// reste est détruit : le lieu revient à son point de départ (il reste acheté ; ses points
// d'école restent ; ses étals, ses drones, ses stocks reviennent à zéro).
//
// LA CASERNE DE POMPIERS (rangée du bas, offerte avec la boulangerie) : le commandant (50 points
// d'université) ; à gauche la prévention (6 bureaux, leurs chargés de prévention : points du
// lycée ; chacun rapporte 1 PPI toutes les 2 heures, +1 % de vitesse par point d'université placé
// sur la caserne), à droite l'action (5 camions, leurs pompiers : points du collège).

// FireTime : la durée d'un incendie.
const FireTime = 60.0

// hotIDs : les zones chaudes de chaque lieu (nil : toutes ses machines).
var hotIDs = map[string][]string{
	"boulangerie": nil,
	"patisserie":  nil,
	"creperie":    nil,
	"boucherie":   {"rotisserie", "foie_gras", "jambon", "pate"},
	"domaine":     {"whisky", "vodka", "alcool_poire"},
}

// HotZones : les zones chaudes achetées de ce lieu.
func (g *Game) HotZones(id string) int {
	ids, ok := hotIDs[id]
	p := g.Place(id)
	if !ok || p == nil || !p.Owned {
		return 0
	}
	n := 0
	for _, s := range p.Stations {
		if s.Kind != Machine || !s.Built {
			continue
		}
		if ids == nil {
			n++
			continue
		}
		for _, h := range ids {
			if s.ID == h {
				n++
			}
		}
	}
	return n
}

// HotPlaces : les lieux qui peuvent brûler (dans l'ordre de la carte).
func HotPlaces() []string {
	var out []string
	for _, id := range TopRow() {
		if _, ok := hotIDs[id]; ok {
			out = append(out, id)
		}
	}
	return out
}

// firesActive : les incendies arrivent dès que la boulangerie est ouverte.
func (g *Game) firesActive() bool {
	q := g.Place("boulangerie")
	return q != nil && q.Owned
}

// PPIOf : la prévention placée sur ce lieu.
func (g *Game) PPIOf(id string) int { return g.Levels["caserne.ppi_"+id] }

// FireRisk : la chance que le feu prenne vraiment (1 ; −1 % tous les 10 PPI ; 0,1 au moins).
func (g *Game) FireRisk(id string) float64 {
	return math.Max(0.1, 1-float64(g.PPIOf(id))/1000)
}

// PPIMax : les PPI au plus sur un lieu (le risque ne descend pas sous 10 %).
const PPIMax = 900

// tryFire : le groupe « incendie » est tiré : le feu prend-il ?
func (g *Game) tryFire(id string) {
	p := g.Place(id)
	if p == nil {
		return
	}
	if g.rng.Float64() >= g.FireRisk(id) {
		g.Levels["caserne.evites"]++
		g.emit(Event{Kind: EvLevel, Place: id, Pos: p.Gate, Text: tr.F("Un début d'incendie s'est déclaré {alieu}, mais la prévention incendie a permis d'éviter le pire.", "alieu", placeAt(id), "lieu", theName(id))})
		g.orderT = g.eventGap()
		return
	}
	g.startFire(id)
}

// startFire : le feu prend.
func (g *Game) startFire(id string) {
	p := g.Place(id)
	g.Theft = &Theft{Place: id, Left: FireTime, Total: FireTime, Fire: true, robbed: map[int]bool{}}
	p.theftFlee()
	if g.Cur == p && p.Yard.W > 0 && !p.Yard.Has(g.Player.Pos) {
		g.Theft.Scared = true
	}
	txt := tr.F("Au feu ! {lieu} brûle !", "lieu", theName(id))
	if n := g.TrucksReady(); n > 0 {
		txt += " " + tr.F("Les pompiers arrivent ({n} camion(s)).", "n", itoa(n))
	} else {
		txt += " " + tr.T("Aucun camion de pompiers ne peut venir.")
	}
	g.emit(Event{Kind: EvLevel, Place: id, Pos: p.Gate, Text: txt})
	g.emit(Event{Kind: EvAngry, Place: id, Pos: p.Gate})
}

// FireAt : un incendie en cours dans ce lieu ?
func (g *Game) FireAt(id string) bool { return g.Theft != nil && g.Theft.Fire && g.Theft.Place == id }

// TrucksReady : les camions de pompiers qui partent (achetés, avec leur pompier, et le commandant).
func (g *Game) TrucksReady() int {
	if g.Levels["caserne.commandant"] == 0 {
		return 0
	}
	n := 0
	for k := 1; k <= CasTruckN; k++ {
		if g.TruckReady(k) {
			n++
		}
	}
	return n
}

// TruckReady : le camion n° k part-il ?
func (g *Game) TruckReady(k int) bool {
	return g.Levels["caserne.camion_"+itoa(k)] > 0 && g.Levels["caserne.pompier_"+itoa(k)] > 0
}

// TruckSaves : les éléments que sauve le camion n° k (1, +1 tous les 1 000 points du primaire).
func (g *Game) TruckSaves(k int) int { return 1 + g.Levels["caserne.camion_niv_"+itoa(k)] }

// SaveCapacity : tout ce que sauvent les camions.
func (g *Game) SaveCapacity() int {
	if g.Levels["caserne.commandant"] == 0 {
		return 0
	}
	n := 0
	for k := 1; k <= CasTruckN; k++ {
		if g.TruckReady(k) {
			n += g.TruckSaves(k)
		}
	}
	return n
}

// fireElem : un élément à sauver.
type fireElem struct {
	key  string  // le niveau (caissière, employés) ; "" pour un objet
	u    *Unlock // l'objet (four, table…)
	cost float64
}

// staffKey : une case d'embauche (caissière, employés, serveurs) ?
func staffKey(id string) bool {
	return strings.HasPrefix(id, "caissiere") || id == "ouvriers" || strings.HasPrefix(id, "ouvriers_") || id == "serveurs"
}

// fireElems : ce qu'il y a à sauver, dans l'ordre : la caissière, les employés, puis les objets
// du plus cher au moins cher.
func (p *Place) fireElems() []fireElem {
	g := p.g
	var cash, staff, objs []fireElem
	for _, sp := range p.Spots {
		u := p.SpotUpgrade(sp)
		if u == nil || !staffKey(sp.ID) {
			continue
		}
		for i := 0; i < g.Levels[u.Key]; i++ {
			e := fireElem{key: u.Key}
			if strings.HasPrefix(sp.ID, "caissiere") {
				cash = append(cash, e)
			} else {
				staff = append(staff, e)
			}
		}
	}
	for _, u := range p.Unlocks {
		if u.Done {
			objs = append(objs, fireElem{u: u, cost: u.Cost})
		}
	}
	sort.SliceStable(objs, func(i, j int) bool { return objs[i].cost > objs[j].cost })
	return append(append(cash, staff...), objs...)
}

// MaxFireElems : le plus d'éléments qu'un lieu à risque peut avoir (tout acheté) : de quoi fixer
// le plafond des améliorations des camions (5 camions doivent pouvoir sauver tout un lieu).
func (g *Game) MaxFireElems() int {
	best := 0
	for _, id := range HotPlaces() {
		p := g.Place(id)
		if p == nil {
			continue
		}
		n := len(p.Unlocks)
		for _, sp := range p.Spots {
			if u := p.SpotUpgrade(sp); u != nil && staffKey(sp.ID) {
				n += u.Max
			}
		}
		best = max(best, n)
	}
	return best
}

// TruckLvlMax : les améliorations au plus d'un camion.
func (g *Game) TruckLvlMax() int { return max(0, (g.MaxFireElems()+CasTruckN-1)/CasTruckN-1) }

// TruckUpCost : 1 000 points du primaire pour +1 élément.
const TruckUpCost = 1000

// BuyTruckUp : un élément de plus pour le camion n° k.
func (g *Game) BuyTruckUp(k int) bool {
	key := "caserne.camion_niv_" + itoa(k)
	if g.Levels["caserne.camion_"+itoa(k)] == 0 || g.Levels[key] >= g.TruckLvlMax() || g.Points("ouvrier") < TruckUpCost {
		return false
	}
	g.Levels["pts.ouvrier"] -= TruckUpCost
	g.Levels[key]++
	return true
}

// burn : la fin de l'incendie : ce qui n'est pas sauvé est détruit, le lieu revient à son point
// de départ.
func (g *Game) burn(p *Place) {
	elems := p.fireElems()
	cap := min(g.SaveCapacity(), len(elems))
	g.Levels["caserne.sauves"] += cap
	keepStaff := map[string]int{}
	keepObj := map[*Unlock]bool{}
	for _, e := range elems[:cap] {
		if e.u != nil {
			keepObj[e.u] = true
		} else {
			keepStaff[e.key]++
		}
	}
	// les cases : les employés sauvés restent ; tout le reste revient à zéro (drones, étals…)
	for _, sp := range p.Spots {
		u := p.SpotUpgrade(sp)
		if u == nil {
			continue
		}
		g.Levels[u.Key] = keepStaff[u.Key]
		if g.Levels[u.Key] == 0 {
			delete(g.Levels, u.Key)
		}
		sp.Paid = 0
	}
	// les objets : détruits s'ils ne sont pas sauvés
	gone := map[string]bool{}
	for _, u := range p.Unlocks {
		if u.Done && !keepObj[u] {
			u.Done, u.Paid = false, 0
			gone[u.ID] = true
		}
	}
	for _, s := range p.Stations {
		if s.Unlock != "" && gone[s.Unlock] {
			s.Built = false
			if s.Kind == Field && len(s.Slots) > 0 { // (le champ, ou la colonne d'arbres, repart de rien)
				pl := Plant{P: s.Slots[0]}
				if s.Column {
					pl.Young = 1
				}
				s.Plants = []Plant{pl}
			}
		}
		if s.Unlock == "" || gone[s.Unlock] { // (et son niveau, son stock agrandi)
			if s.Upgradable() {
				delete(g.Levels, p.StationUpgrade(s).Key)
			}
			if s.Stockable() {
				delete(g.Levels, p.StockUpgrade(s).Key)
			}
			s.LvlPaid, s.StkPaid = 0, 0
		}
		// tous les stocks brûlent
		s.Pile, s.Pile2, s.Stock = 0, 0, 0
		s.Buf = [NItems]int{}
		s.Drones = nil
	}
	// les employés partent, traumatisés (ceux qui sont sauvés reviennent)
	p.Workers = nil
	for _, r := range p.Registers() {
		r.Cashier = nil
	}
	if cp := p.Crep; cp != nil && len(cp.Waiters) > g.Levels[p.key("serveurs")] {
		cp.Waiters = cp.Waiters[:g.Levels[p.key("serveurs")]]
	}
	if g.Player != nil && g.Cur == p {
		g.Player.Carry = nil
	}
	p.applyCaps()
	p.syncStaff()
	g.emit(Event{Kind: EvLevel, Place: p.ID, Pos: p.Gate, Text: tr.F("L'incendie {delieu} est éteint : les pompiers ont sauvé {n} élément(s) sur {m}. Tout le reste est détruit.", "delieu", placeOf(p.ID), "lieu", theName(p.ID), "n", itoa(cap), "m", itoa(len(elems)))})
}

// ---------------------------------------------------------------- la caserne

const (
	CasOffN    = 6 // les bureaux de prévention
	CasTruckN  = 5 // les camions
	CasBldY    = 480.0
	CasRoomW   = 640.0
	CasOffW    = 300.0
	CasOffH    = 330.0
	CasCorH    = 200.0
	CasOffGap  = 50.0
	CasGap     = 60.0
	CasBayW    = 280.0
	CasBayGap  = 50.0
	CasYardX   = 300.0
	CasParkY   = roadY + 240
	PPITime    = 2 * 3600.0 // un PPI toutes les 2 heures, par chargé de prévention
	casBldH    = 60 + CasOffH + CasCorH + CasOffH + 60
	casOffCols = 3
)

// CasBldH : la hauteur du bâtiment.
const CasBldH = casBldH

func casLeftW() float64  { return casOffCols*(CasOffW+CasOffGap) - CasOffGap }
func casRightW() float64 { return CasTruckN*(CasBayW+CasBayGap) - CasBayGap }

// CasBldW : la largeur du bâtiment.
var CasBldW = 60 + casLeftW() + CasGap + CasRoomW + CasGap + casRightW() + 60

// CasW : la largeur de la caserne.
var CasW = math.Max(CasBldW+600, CasYardX+2150+400)

// CasBldX : le bord gauche du bâtiment.
var CasBldX = (CasW - CasBldW) / 2

func casTop() float64 { return CasBldY + 60 }

// CasChief : le bureau du commandant (au centre, en haut).
func CasChief() Rect { return Rect{CasBldX + 60 + casLeftW() + CasGap, casTop(), CasRoomW, CasOffH} }

// CasHall : le hall (sous le bureau du commandant) : les cases de la prévention et des camions.
func CasHall() Rect {
	c := CasChief()
	return Rect{c.X, c.Y + c.H + 20, c.W, CasCorH + CasOffH - 20}
}

// CasOffice : le bureau de prévention n° n (1 à 6), du centre vers la gauche, en haut puis en bas.
func CasOffice(n int) Rect {
	x1 := CasBldX + 60 + casLeftW()
	c, r := (n-1)/2, (n-1)%2
	return Rect{x1 - CasOffW - float64(c)*(CasOffW+CasOffGap), casTop() + float64(r)*(CasOffH+CasCorH), CasOffW, CasOffH}
}

// CasCorridorY : le haut du couloir de la prévention.
func CasCorridorY() float64 { return casTop() + CasOffH }

// CasGarage : le garage des camions (à droite ; ses portes donnent sur la route).
func CasGarage() Rect {
	c := CasChief()
	return Rect{c.X + c.W + CasGap, casTop(), casRightW(), CasBldH - 120}
}

// CasBay : la place du camion n° k.
func CasBay(k int) Rect {
	g := CasGarage()
	return Rect{g.X + float64(k-1)*(CasBayW+CasBayGap), g.Y + 60, CasBayW, g.H - 240}
}

// CasPrevZone, CasTruckZone : les cases du hall (le menu de la prévention, celui des camions).
func CasPrevZone() Rect  { h := CasHall(); return Rect{h.X + 60, h.Y + h.H - 190, 220, 150} }
func CasTruckZone() Rect { h := CasHall(); return Rect{h.X + h.W - 280, h.Y + h.H - 190, 220, 150} }

// Caserne : ce qui vit à la caserne (le temps des PPI en cours).
type Caserne struct {
	Prog float64 // le temps accumulé vers le prochain PPI (secondes × chargés)
}

func newCaserne(g *Game) *Place {
	W := CasW
	p := &Place{ID: "caserne", Name: "La caserne de pompiers", W: W, H: CasParkY + 520, Top: CasBldY - 150, g: g,
		rng: rand.New(rand.NewSource(307)), Rep: 3, Price: 0, SellMult: 1, MaxCust: 0}
	p.Cas = &Caserne{}
	door := V{CasHall().C().X, CasBldY + CasBldH}
	p.Gate = V{door.X, roadY + 40}
	p.SpawnL = V{350, roadY + 40}
	p.SpawnR = V{W - 350, roadY + 40}
	p.Inside = CasHall().C()
	p.RestSpot = p.Inside
	k := p.key
	up := func(id, name string, cost float64, pts, desc string) {
		p.ups = append(p.ups, &Upgrade{Key: k(id), Name: name, Max: 1, Costs: []float64{cost}, Pts: pts,
			Desc: func(int) string { return desc }})
	}
	spot := func(id, label, needs string, z Rect, done string) {
		p.Spots = append(p.Spots, &Spot{ID: id, Label: label, Needs: needs, Zone: z, Done: func(int) string { return done }})
	}
	c := CasChief()
	up("commandant", "Commandant", 50, "formation", "Sans lui, la caserne ne fonctionne pas. Il coûte 50 points d'université (l'école).")
	spot("commandant", "Commandant", "", Rect{c.X + c.W - 180, c.Y + c.H - 150, 140, 130}, "Le commandant est arrivé : la caserne peut fonctionner !")
	for n := 1; n <= CasOffN; n++ {
		r := CasOffice(n)
		prev := "commandant"
		if n > 1 {
			prev = "bureau_" + itoa(n-1)
		}
		up("bureau_"+itoa(n), "Bureau de prévention", niceRound(1e5*math.Pow(4, float64(n-1))), "", "Un chargé de prévention y travaillera : il rapporte des points de prévention incendie (PPI).")
		spot("bureau_"+itoa(n), "Bureau", prev, Rect{r.C().X - 70, r.C().Y - 65, 140, 130}, tr.F("Bureau de prévention n° {n} !", "n", itoa(n)))
		up("charge_"+itoa(n), "Chargé de prévention", float64(20*int(math.Pow(3, float64(n-1)))), "comptable", "Formé avec des points du lycée, il rapporte 1 PPI toutes les 2 heures (plus vite avec la formation de l'université).")
		spot("charge_"+itoa(n), "Chargé de prévention", "bureau_"+itoa(n), Rect{r.C().X - 70, r.Y + r.H - 140, 140, 130}, tr.F("Chargé de prévention n° {n} formé !", "n", itoa(n)))
	}
	for n := 1; n <= CasTruckN; n++ {
		b := CasBay(n)
		prev := "commandant"
		if n > 1 {
			prev = "camion_" + itoa(n-1)
		}
		up("camion_"+itoa(n), "Camion de pompiers", niceRound(2e5*math.Pow(5, float64(n-1))), "", "Avec son pompier, il file éteindre les incendies et sauve un élément du lieu (plus avec les améliorations).")
		spot("camion_"+itoa(n), "Camion", prev, Rect{b.C().X - 70, b.C().Y - 65, 140, 130}, "Nouveau camion de pompiers !")
		up("pompier_"+itoa(n), "Pompier", float64(10*int(math.Pow(5, float64(n-1)))), "marketing", "Formé avec des points du collège, il conduit son camion.")
		spot("pompier_"+itoa(n), "Pompier", "camion_"+itoa(n), Rect{b.C().X - 70, b.Y + b.H + 20, 140, 120}, tr.F("Pompier n° {n} formé !", "n", itoa(n)))
	}
	p.setupYardH(CasYardX+650, CasParkY)
	return p
}

// HasCommander : le commandant est-il là ?
func (g *Game) HasCommander() bool { return g.Levels["caserne.commandant"] > 0 }

// PrevOfficers : les chargés de prévention au travail.
func (g *Game) PrevOfficers() int {
	if !g.HasCommander() {
		return 0
	}
	n := 0
	for k := 1; k <= CasOffN; k++ {
		if g.Levels["caserne.bureau_"+itoa(k)] > 0 && g.Levels["caserne.charge_"+itoa(k)] > 0 {
			n++
		}
	}
	return n
}

// PPIRate : la vitesse des chargés (+1 % par point d'université placé sur la caserne).
func (g *Game) PPIRate() float64 { return 1 + float64(g.Levels["caserne.pts_prevention"])/100 }

// PPI : les points de prévention incendie à répartir.
func (g *Game) PPI() int { return g.Levels["pts.ppi"] }

// PPINext : le temps avant le prochain PPI (0 : aucun chargé).
func (g *Game) PPINext() float64 {
	n := g.PrevOfficers()
	p := g.Place("caserne")
	if n == 0 || p == nil {
		return 0
	}
	return (PPITime - p.Cas.Prog) / (float64(n) * g.PPIRate())
}

// SpendPPI : n PPI sur ce lieu (900 au plus).
func (g *Game) SpendPPI(id string, n int) bool {
	n = min(n, PPIMax-g.PPIOf(id), g.PPI())
	if n <= 0 {
		return false
	}
	g.Levels["pts.ppi"] -= n
	g.Levels["caserne.ppi_"+id] += n
	return true
}

// openCaserne : la caserne s'ouvre (offerte) avec la boulangerie.
func (g *Game) openCaserne() {
	p := g.Place("caserne")
	if p == nil || p.Owned || !g.firesActive() {
		return
	}
	p.Owned = true
	g.emit(Event{Kind: EvLevel, Text: "La caserne de pompiers ouvre (offerte) ! Il lui faut un commandant."})
}

func (p *Place) stepCaserne(dt float64, active bool) {
	g := p.g
	if active && g.Player != nil {
		p.interact(g.Player, dt)
	}
	if p.Cas.Prog == 0 && g.Levels["caserne.ppi_prog"] > 0 {
		p.Cas.Prog = float64(g.Levels["caserne.ppi_prog"]) // (une partie rechargée)
	}
	if n := g.PrevOfficers(); n > 0 {
		p.Cas.Prog += dt * float64(n) * g.PPIRate()
		for p.Cas.Prog >= PPITime {
			p.Cas.Prog -= PPITime
			g.Levels["pts.ppi"]++
			g.emit(Event{Kind: EvLevel, Place: p.ID, Pos: CasHall().C(), Text: "+1 point de prévention incendie (PPI) !"})
		}
	}
	g.Levels["caserne.ppi_prog"] = int(p.Cas.Prog)
}

// StartFire : un incendie tout de suite dans ce lieu (captures d'écran, essais).
func (g *Game) StartFire(id string) {
	if g.Theft == nil && g.HotZones(id) > 0 {
		g.Order = nil
		g.startFire(id)
	}
}
