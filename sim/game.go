package sim

import (
	"math"
	"math/rand"
	"sort"
	"strings"
)

// ---------------------------------------------------------------- événements

type EvKind int

const (
	EvHarvest EvKind = iota
	EvTake
	EvDrop
	EvMade
	EvTrash
	EvPay
	EvUnlock
	EvCoin
	EvAngry
	EvHire
	EvPlant
	EvLevel
	EvBadge
	EvTruck
	EvPoint // un élève quitte l'école (Text : type de point)
	EvBoost // une classe reçoit son matériel (Text : nom de la classe)
	EvOrder // une nouvelle commande arrive (Text : le message)
)

type Event struct {
	Kind   EvKind
	Pos    V
	It     Item
	Amount float64
	Text   string
	Place  string // lieu où ça se passe ("" : toute la ville)
}

// ---------------------------------------------------------------- tenues

type Outfit struct {
	ID, Name, Desc string
	Cost           float64
	Speed          float64   // multiplicateur de vitesse
	Cap            int       // objets portés en plus (ou en moins)
	ChefBoost      float64   // machines proches plus rapides
	TipBonus       float64   // pourboire quand tu encaisses toi-même
	Colors         [3]uint32 // haut, bas, chapeau
}

// Outfits : plus de tenues à acheter (4.1) ; la tenue de départ donne les couleurs du personnage.
var Outfits = []Outfit{
	{ID: "fermier", Name: "Fermier", Desc: "La tenue de départ.", Speed: 1, Colors: [3]uint32{0xd9534f, 0x3b6ea8, 0xe8c26a}},
}

func OutfitByID(id string) *Outfit {
	for i := range Outfits {
		if Outfits[i].ID == id {
			return &Outfits[i]
		}
	}
	return &Outfits[0]
}

// ---------------------------------------------------------------- améliorations

type Upgrade struct {
	Key   string // clé dans Levels
	Name  string
	Desc  func(lvl int) string
	Now   bool // la description parle du niveau actuel (et non du suivant)
	Max   int
	Costs []float64 // prix de chaque niveau (si absent : Base × Mult^niveau)
	Base  float64
	Mult  float64
}

func (u *Upgrade) Cost(lvl int) float64 {
	c := 0.0
	if lvl < len(u.Costs) {
		c = u.Costs[lvl]
	} else {
		c = math.Round(u.Base * math.Pow(u.Mult, float64(lvl)))
	}
	if CostScale != nil {
		if s := CostScale(u.Key); s != 1 {
			c = niceRound(c * s) // (lieu loin sur la carte : prices.go)
		}
	}
	if CostMult != nil {
		c = math.Round(c * CostMult(u.Key))
	}
	return c
}

// CostScale : coefficient des prix selon le rang du lieu de l'amélioration sur la carte.
var CostScale func(key string) float64

// CostMult : réduction des prix (négociation, points de lycée), selon le lieu de l'amélioration.
var CostMult func(key string) float64

var NightHours = []float64{1, 2, 4, 8, 12, 24}

// ---------------------------------------------------------------- partie

type Stats struct {
	Served    int
	Missed    int
	Earned    float64
	Played    float64
	Harvested int
	Replanted int            // sachets de graines plantés (serre)
	Made      map[string]int // produits fabriqués (machines et animaux), par nom
	Happy     int
	Angry     int
	Delivered int
	Orders    int // commandes spéciales livrées
}

type incomeRec struct {
	t, a float64
}

type Game struct {
	stepping string // lieu en train d'avancer (pour les événements)
	Money    float64
	Places   []*Place
	Cur      *Place // lieu où se trouve le joueur (nil : carte de la ville)
	Player   *Agent
	Levels   map[string]int
	Outfit   string
	Owned    map[string]bool
	T        float64
	Stats    Stats
	Badges   map[string]bool
	badgeT   float64
	diagT    float64            // (rayons vides du supermarché)
	TruckT   map[string]float64 // chrono de chaque camion (par quai)

	Events    []Event
	Order     *Order            // commande spéciale en cours (ferme)
	Boost     *Boost            // cadeau d'une commande en cours (un seul à la fois)
	Fav       map[string]string // favoris : touche ("1" à "9", "0") → lieu
	orderT    float64           // avant la prochaine commande
	lastOrder string            // la dernière commande (qui + quoi)
	income    []incomeRec
	rng       *rand.Rand
}

// DefaultFav : les favoris au départ (les touches d'avant).
var DefaultFav = map[string]string{"1": "ferme", "2": "ecole", "3": "super", "4": "cremerie", "5": "exotique",
	"6": "boulangerie", "7": "glacier", "8": "elevage", "9": "boucherie", "0": "port"}

// FavKey : la touche d'un lieu ("" : pas un favori).
func (g *Game) FavKey(id string) string {
	for k, v := range g.Fav {
		if v == id {
			return k
		}
	}
	return ""
}

// SetFav : ce lieu va sur cette touche (il quitte son ancienne touche ; le lieu
// qui y était n'a plus de touche).
func (g *Game) SetFav(key, id string) {
	for k, v := range g.Fav {
		if v == id {
			delete(g.Fav, k)
		}
	}
	g.Fav[key] = id
}

func NewGame() *Game {
	g := &Game{Levels: map[string]int{}, Badges: map[string]bool{}, Owned: map[string]bool{"fermier": true}, Outfit: "fermier", rng: rand.New(rand.NewSource(7))}
	g.Fav = map[string]string{}
	for k, v := range DefaultFav {
		g.Fav[k] = v
	}
	g.Places = []*Place{newFarm(g), newSuper(g), newWorkshop(g, bakerySpec), newWorkshop(g, cremerieSpec), newWorkshop(g, iceSpec), newSchool(g), newElevage(g), newBoucherie(g), newExotic(g), newPort(g), newDomaine(g)}
	g.sortPlaces()
	for _, p := range g.Places {
		p.applyCaps()
		for _, s := range p.Stations {
			s.place = p.ID
		}
	}
	// négociation (lycée) : les améliorations du lieu coûtent moins cher
	CostMult = func(key string) float64 {
		id, _, ok := strings.Cut(key, ".")
		if !ok {
			return 1
		}
		return 1 / (1 + float64(g.Levels[id+".pts_nego"])/100)
	}
	// les rayons du supermarché (d'après les prix de base des lieux d'origine : le
	// supermarché garde ses prix), puis les prix selon le rang du lieu sur la carte
	g.fixRayonCosts()
	CostScale = g.keyScale
	g.applyMapPrices()
	// l'école est ouverte dès le début, gratuitement, avec une table et un prof
	if sc := g.Place("ecole"); sc != nil {
		sc.Owned, sc.Price = true, 0
		g.Levels["ecole.prof_"+sc.Rooms[0].ID] = 1
	}
	g.Player = &Agent{Kind: PlayerAgent, Face: V{0, 1}}
	g.Enter(g.Places[0])
	g.Money = 0
	return g
}

// MapOrder : les lieux dans l'ordre de la carte de la ville (ui.go, drawMap). Il
// fixe l'ordre des listes (points d'école…) et le rang de chaque lieu, qui fixe
// ses prix (PlaceScale). Un nouveau lieu se met ici, à sa place sur la carte.
var MapOrder = []string{"ferme", "ecole", "super", "cremerie", "exotique", "boulangerie", "glacier", "port", "elevage", "boucherie", "domaine"}

// MapRank : position du lieu sur la carte (1 : ma ferme) ; 0 si inconnu.
func MapRank(id string) int {
	for i, m := range MapOrder {
		if m == id {
			return i + 1
		}
	}
	return 0
}

// sortPlaces : g.Places dans l'ordre de la carte (les lieux inconnus à la fin).
func (g *Game) sortPlaces() {
	sort.SliceStable(g.Places, func(i, j int) bool {
		ri, rj := MapRank(g.Places[i].ID), MapRank(g.Places[j].ID)
		if ri == 0 {
			ri = 1 << 20
		}
		if rj == 0 {
			rj = 1 << 20
		}
		return ri < rj
	})
}

func (g *Game) Place(id string) *Place {
	for _, p := range g.Places {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (g *Game) Enter(p *Place) {
	g.Cur = p
	if p != nil {
		g.Player.Pos = p.Start
		g.Player.Carry = nil
	}
}

func (g *Game) outfit() *Outfit    { return OutfitByID(g.Outfit) }
func (g *Game) OutfitNow() *Outfit { return g.outfit() }

func (g *Game) emit(e Event) {
	if e.Place == "" {
		e.Place = g.stepping
	}
	g.Events = append(g.Events, e)
}

func (g *Game) earn(a float64, at V, place string) {
	a = math.Round(a)
	g.income = append(g.income, incomeRec{g.T, a}) // (sans le cadeau : il ne compte pas pour les absences)
	a *= g.boostF("ventes", place)                 // cadeau : ventes ×2 dans ce lieu
	g.Money += a
	g.Stats.Earned += a
	g.emit(Event{Kind: EvCoin, Pos: at, Amount: a, Text: place})
}

func (g *Game) priceMult(p *Place) float64 {
	return 1 + p.ptsF("pts_comptable")
}

// Rate : argent gagné par seconde sur les 5 dernières minutes.
func (g *Game) Rate() float64 { return g.rateOver(300) }

// RateLong : revenu moyen sur 15 minutes (sert aux gains pendant l'absence,
// pour qu'un pic juste avant de quitter ne compte pas trop).
func (g *Game) RateLong() float64 { return g.rateOver(900) }

func (g *Game) rateOver(win float64) float64 {
	k := 0
	for k < len(g.income) && g.income[k].t < g.T-900 {
		k++
	}
	g.income = g.income[k:]
	sum := 0.0
	for _, r := range g.income {
		if r.t >= g.T-win {
			sum += r.a
		}
	}
	span := math.Min(win, math.Max(60, g.T))
	return sum / span
}

// earnBonus : de l'argent qui ne compte pas dans le revenu régulier (commandes).
func (g *Game) earnBonus(a float64, at V, place string) {
	a = math.Round(a)
	g.Money += a
	g.Stats.Earned += a
	g.emit(Event{Kind: EvCoin, Pos: at, Amount: a, Text: place})
}

// Step fait avancer toute la ville d'un pas de temps.
func (g *Game) Step(dt float64) {
	g.T += dt
	g.Stats.Played += dt
	for _, p := range g.Places {
		if p.Owned {
			g.stepping = p.ID // (les sons ne s'entendent que dans le lieu où l'on est)
			p.step(dt, p == g.Cur)
		}
	}
	g.stepping = ""
	g.syncPallets()
	g.stepTrucks(dt)
	g.stepRayonDiag(dt)
	g.stepOrders(dt)
	g.stepBoost(dt)
	g.badgeT += dt
	if g.badgeT > 1 {
		g.badgeT = 0
		g.checkBadges(true)
	}
}

// ---------------------------------------------------------------- personnel

func (p *Place) syncStaff() {
	g := p.g
	if p.ID == "ferme" {
		// une équipe par zone
		for _, t := range FarmTeams {
			n := min(g.Levels[p.key(teamKey(t.ID))], t.Max) // (une sauvegarde d'avant peut en avoir plus)
			have := 0
			for _, w := range p.Workers {
				if w.Team == t.ID {
					have++
				}
			}
			for ; have < n; have++ {
				pos := p.teamRest(t.ID)
				w := &Agent{Kind: WorkerAgent, Pos: pos, Look: len(p.Workers) + 1, Face: V{0, 1}, Team: t.ID}
				w.lastPos = w.Pos
				p.Workers = append(p.Workers, w)
				g.emit(Event{Kind: EvHire, Pos: w.Pos})
			}
		}
	}
	if p.ID == "super" {
		// des magasiniers par espace
		for _, z := range SuperZones {
			n := g.Levels[p.key(teamKey(z.ID))]
			have := 0
			for _, w := range p.Workers {
				if w.Team == z.ID {
					have++
				}
			}
			for ; have < n; have++ {
				w := &Agent{Kind: WorkerAgent, Pos: superRest(z.ID), Look: len(p.Workers) + 1, Face: V{0, 1}, Team: z.ID}
				w.lastPos = w.Pos
				p.Workers = append(p.Workers, w)
				g.emit(Event{Kind: EvHire, Pos: w.Pos})
			}
		}
	}
	// ferme d'élevage, boucherie : des employés par espace
	for _, z := range teamZonesOf(p.ID) {
		n := min(g.Levels[p.key(teamKey(z.ID))], z.Max()) // (au plus le maximum de l'espace)
		have := 0
		for _, w := range p.Workers {
			if w.Team == z.ID {
				have++
			}
		}
		for ; have < n; have++ {
			w := &Agent{Kind: WorkerAgent, Pos: p.zoneRest(z.ID), Look: len(p.Workers) + 1, Face: V{0, 1}, Team: z.ID}
			w.lastPos = w.Pos
			p.Workers = append(p.Workers, w)
			g.emit(Event{Kind: EvHire, Pos: w.Pos})
		}
	}
	n := g.Levels[p.key("ouvriers")]
	for p.ID != "ferme" && p.ID != "super" && teamZonesOf(p.ID) == nil && len(p.Workers) < n {
		w := &Agent{Kind: WorkerAgent, Pos: p.Inside, Look: len(p.Workers) + 1, Face: V{0, 1}}
		w.lastPos = w.Pos
		p.Workers = append(p.Workers, w)
		g.emit(Event{Kind: EvHire, Pos: w.Pos})
	}
	for i, r := range p.Registers() {
		if r.Cashier == nil && r.HireKey != "" && g.Levels[p.key(r.HireKey)] > 0 {
			r.Cashier = &Agent{Kind: CashierAgent, Pos: r.Spot, Face: V{0, 1}, Look: 99 + i*7}
			g.emit(Event{Kind: EvHire, Pos: r.Spot})
		}
	}
}

// ---------------------------------------------------------------- achats

// Buy achète un niveau d'amélioration. Renvoie faux si impossible.
func (g *Game) Buy(u *Upgrade) bool {
	l := g.Levels[u.Key]
	if l >= u.Max {
		return false
	}
	if !g.hireOK(u.Key) || !g.droneOK(u.Key) || !g.schoolOK(u.Key) {
		return false // les ouvriers de la ferme s'embauchent dans l'ordre
	}
	c := u.Cost(l)
	if g.Money < c {
		return false
	}
	g.Money -= c
	g.Levels[u.Key] = l + 1
	g.autoWear(u.Key)
	g.syncFarmTotal()
	for _, p := range g.Places {
		p.applyCaps()
	}
	return true
}

// ---------------------------------------------------------------- catalogue

func pct(v float64) string { return itoa(int(math.Round(v))) + " %" }

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}

// MairieUpgrades : améliorations valables pour toute la ville.
func MairieUpgrades() []*Upgrade {
	return []*Upgrade{
		{Key: "mairie.veilleur", Name: "Veilleur de nuit", Max: 5, Costs: []float64{1500, 6000, 25000, 90000, 300000},
			Desc: func(l int) string {
				return "Tes employés travaillent pendant ton absence : " + itoa(int(NightHours[l])) + " h maximum."
			}},
	}
}

// PlaceUpgrades : améliorations propres à un lieu.
func (p *Place) PlaceUpgrades() []*Upgrade {
	if p.ups != nil {
		return p.ups
	}
	k := p.key
	ups := []*Upgrade{
		{Key: k("caissiere"), Name: "Caissière", Max: 1, Costs: []float64{150},
			Desc: func(l int) string {
				if l > 0 {
					return "Elle encaisse les clients à ta place."
				}
				return "Encaisse les clients à ta place. Indispensable quand tu n'es pas là !"
			}},
		{Key: k("etals"), Name: "Étals plus grands", Max: 6, Base: 200, Mult: 2.3,
			Desc: func(l int) string { return "Chaque étal contient " + itoa(12+4*l) + " produits." }},
	}
	if p.ID == "ferme" {
		ups = append(ups, p.droneUpgrades(1)...)
		for _, t := range FarmTeams {
			t := t
			ups = append(ups, &Upgrade{Key: k(teamKey(t.ID)), Name: "Ouvrier " + t.Name, Max: t.Max, Now: true, Costs: teamCosts(t.ID),
				Desc: func(l int) string {
					return "Il ne travaille que dans la zone " + t.Name + " : champ, bâtiments et étals. Embauchés : " + itoa(l) + "."
				}})
		}
	}
	return ups
}

// StationUpgrade : amélioration du niveau d'une station (vitesse, stock).
func (p *Place) StationUpgrade(s *Station) *Upgrade {
	base := math.Max(30, s.BaseCost*0.6)
	if s.UpMult > 0 {
		base = math.Max(30, base*s.UpMult)
	}
	return &Upgrade{Key: p.key(s.ID), Name: s.Name, Max: 9, Base: base, Mult: 1.9,
		Desc: func(l int) string {
			txt := "Niveau " + itoa(l+1) + " : " + pct(100*lvlFactor(l+1)-100) + " plus rapide."
			if s.Kind == Field {
				if y := FieldYield(l + 1); y > 1 {
					txt += " Chaque pousse donne " + itoa(y) + " récoltes."
				} else {
					txt += " (Au niveau 5 : 2 récoltes par pousse.)"
				}
			}
			return txt
		}}
}

// PriceMult : multiplicateur de prix de la ferme (label bio).
func (g *Game) PriceMult() float64 { return g.priceMult(g.Places[0]) }

// Shopping : le client fait ses courses dans les rayons.
func (c *Customer) Shopping() bool { return c.State == cShop && len(c.path) == 0 }

// Wanted : produit que le client cherche en ce moment (-1 si aucun).
func (c *Customer) Wanted() int {
	if c.State != cShop || c.idx >= len(c.Wants) {
		return -1
	}
	return int(c.Wants[c.idx].It)
}

// QueueLen : nombre de clients qui attendent à la caisse.
func (p *Place) QueueLen() int {
	n := 0
	for _, c := range p.Customers {
		if c.State == cQueue {
			n++
		}
	}
	return n
}

// hireOK : un ouvrier de la ferme ne s'embauche que si son champ existe.
func (g *Game) hireOK(key string) bool {
	for _, p := range g.Places {
		for _, z := range teamZonesOf(p.ID) {
			if key == p.key(teamKey(z.ID)) {
				return p.ZoneOpen(z.ID) // (ferme d'élevage, boucherie : l'espace doit exister)
			}
		}
	}
	for _, t := range FarmTeams {
		if key == "ferme."+teamKey(t.ID) {
			f := g.farm().Station(teamAnchor(t.ID))
			if isColumnTeam(t.ID) && g.Levels[key] >= g.farm().columnHires(t.ID) {
				return false // (un ouvrier par colonne ouverte)
			}
			return f != nil && f.Built
		}
	}
	return true
}
