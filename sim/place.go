package sim

import (
	"math"
	"math/rand"
)

// ---------------------------------------------------------------- stations

type Kind int

const (
	Field    Kind = iota // champ : on marche dedans pour récolter
	Animal               // animaux : produisent tout seuls, on ramasse
	Machine              // transformation : on dépose à gauche, on ramasse à droite
	Shelf                // étal : on dépose la marchandise, les clients se servent
	Register             // caisse
	Trash                // poubelle
	Pallet               // palette de la réserve (remplie par le camion)
	Dock                 // quai d'expédition vers le supermarché
	OrderBox             // caisse des commandes spéciales
	Store                // réserve (ex. : les fleurs, en attendant le fleuriste)
)

type Plant struct {
	P     V
	G     float64 // croissance : 1 = mûr ; -1 = case vide, à replanter (serre)
	Young float64 // verger : l'arbre pousse encore (1 : case toute neuve ; 0 : arbre adulte, il fait des fruits)
}

// TreeTime : un arbre (ou un arbuste) du verger met ce multiple d'une pousse de fruits à grandir.
const TreeTime = 1.5

// Empty : la case attend des graines.
func (pl Plant) Empty() bool { return pl.G < 0 }

type Station struct {
	ID, Name                     string
	Kind                         Kind
	Body                         Rect // emprise du bâtiment (solide, sauf les champs)
	Zone                         Rect // zone principale : récolte, dépôt, ramassage
	OutZone                      Rect // machines : zone de sortie
	Item                         Item // ce qui est produit / vendu
	Inputs                       []Item
	Buf                          [NItems]int
	BufMax                       int
	Pile                         int
	PileMax                      int
	Timer                        float64
	Time                         float64 // durée d'un cycle au niveau 1
	Yield                        int     // produits par cycle (machines ; 0 = 1)
	Plants                       []Plant // pousses plantées
	Market                       bool    // étal du petit marché (sous la serre et le verger)
	Shared                       bool    // machine d'un espace où les ouvriers des autres espaces peuvent se servir
	NoSale                       bool    // étal où aucun client ne se sert (grange à foin, frigos)
	Team                         string  // espace de l'atelier (boucherie, ferme d'élevage) : ses employés seulement
	Two                          bool    // deux sorties (un pré : vaches et veaux…) : Item2, Pile2, OutZone2
	Item2                        Item
	Pile2                        int
	OutZone2                     Rect
	flip                         bool
	Column                       bool // colonne de la serre ou du verger : cases achetées vers le haut, plus chères
	Replant                      bool // après la récolte, la case est vide : il faut replanter (graines : Seed) — plus utilisé
	Serre                        bool // une colonne de la serre (légumes)
	Dredged                      bool // parc du port : seule la drague récolte (personne ne va dans l'eau)
	Seed                         Item
	Slots                        []V     // emplacements possibles (dans l'ordre d'achat)
	PlantPaid                    float64 // argent déjà versé pour la prochaine pousse
	LvlPaid                      float64 // argent déjà versé pour le prochain niveau
	StkPaid                      float64 // argent déjà versé pour le prochain agrandissement du stock
	baseBuf, basePile, baseStock int
	Stock                        int
	StockMax                     int
	Spot                         V // étal : place du client ; caisse : place du caissier
	Built                        bool
	BaseCost                     float64          // sert à calculer le prix des améliorations
	Working                      bool             // pour l'animation
	diag                         [NBlocks]float64 // rayon du supermarché : temps passé presque vide, par cause (30 dernières secondes environ)
	lastFull                     bool             // quai : le dernier camion est parti plein
	Unlock                       string           // déblocage qui construit cette station

	QueueHead V       // caisse : début de la file
	QueueDown int     // caisse : places vers le bas avant de tourner
	Cashier   *Agent  // caisse : caissière embauchée
	HireKey   string  // caisse : amélioration qui embauche sa caissière
	To        string  // quai : lieu de destination du camion
	place     string  // lieu où se trouve la station
	scale     float64 // coefficient des prix du lieu (pousses) ; 0 : 1
	PadOff    float64 // étal : distance entre l'étal et la base de son drone (0 : 140)
	yard      bool    // ferme : dans la cour à gauche (ne se décale pas)
	UpMult    float64 // multiplicateur du prix des niveaux (0 : normal)

	claim  [NItems]int // quantités promises par des employés en route
	booked int         // étal : articles réservés par des clients en route
	Drone  *Drone      // étal de la ferme : son drone
}

func (s *Station) Solid() bool { return s.Kind != Field }

// PlaceID : le lieu de la station.
func (s *Station) PlaceID() string { return s.place }

// ---------------------------------------------------------------- personnages

type AgentKind int

const (
	PlayerAgent AgentKind = iota
	WorkerAgent
	CashierAgent
	StudentAgent // élève de l'école
)

type Agent struct {
	Kind   AgentKind
	Pos    V
	Face   V
	Moving bool
	WalkT  float64
	Carry  []Item
	Look   int
	Team   string // ouvrier de la ferme : sa zone
	// ennui : temps passé sans rien à faire ; une fois installé (Bored > BoredAfter
	// et arrivé à sa place), l'ouvrier s'assoit, bâille ou siffle (Mood 0, 1, 2)
	Bored   float64
	Mood    int
	Settled bool

	tick   float64
	target V
	hasT   bool
	// employé
	job     *job
	idleT   float64
	stuckT  float64
	lastPos V
	sideT   float64
	sideDir V
	noDestT float64
	wanderT float64
	route   []V // points de passage pour contourner les bâtiments
	routeTo V
	routeT  float64
}

type job struct {
	src, dst *Station
	it       Item
	want     int
	got      int
	phase    int  // 0 vers la source, 1 vers la destination
	seeded   bool // serre : il a déjà pris ses sachets de graines
	waitT    float64
}

// load : ce qu'il a dans les bras (les sachets de graines ne prennent pas de place).
func (a *Agent) load() int {
	n := 0
	for _, it := range a.Carry {
		if !IsSeed(it) {
			n++
		}
	}
	return n
}

// Count : combien d'objets de ce type il porte.
func (a *Agent) Count(it Item) int { return a.count(it) }

func (a *Agent) count(it Item) int {
	n := 0
	for _, c := range a.Carry {
		if c == it {
			n++
		}
	}
	return n
}

func (a *Agent) remove(it Item) bool {
	for i := len(a.Carry) - 1; i >= 0; i-- {
		if a.Carry[i] == it {
			a.Carry = append(a.Carry[:i], a.Carry[i+1:]...)
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------- clients

type cState int

const (
	cEnter cState = iota
	cShop
	cQueue
	cLeave
	cGone
)

type Want struct {
	It     Item
	N      int
	Got    int
	Booked int // articles réservés en rayon (le reste : « en plus », s'il en reste à son passage)
}

type Customer struct {
	Pos      V
	Face     V
	Moving   bool
	WalkT    float64
	Look     int
	Wants    []Want
	State    cState
	idx      int
	path     []V
	Wait     float64 // temps d'attente au rayon ou dans la file
	Mood     int     // 0 content, 1 impatient, 2 fâché
	Paid     float64
	tick     float64
	Carry    []Item
	CheckN   float64 // objets à passer en caisse (0 : pas encore en caisse)
	CheckT   float64 // objets restant à passer
	slot     int
	reg      *Station
	ticket   int      // numéro d'arrivée dans la file
	veh      *Vehicle // venu en voiture, à vélo… (nil : à pied)
	toCar    bool     // retourne à son véhicule
	cartRes  bool     // un caddie lui est réservé (supermarché)
	Cart     bool     // pousse un caddie
	cartBack bool     // le caddie est (ou va être) rendu
	route    []V
	routeTo  V
	routeT   float64
}

// ---------------------------------------------------------------- déblocages

type Unlock struct {
	ID, Name string
	Cost     float64
	Paid     float64
	Zone     Rect
	Done     bool
	Hint     string
	Requires func() bool // condition (ex. produit débloqué à la ferme)
	Locked   string      // texte quand la condition manque
	Secret   bool        // pas montré (même verrouillé) tant que la condition manque
}

// ShowLocked : montrer ce déblocage verrouillé (avec la raison) ?
func (p *Place) ShowLocked(u *Unlock) bool {
	if u.Done || u.Secret || u.Requires == nil || u.Requires() {
		return false
	}
	return p.ID != "boucherie" || p.firstInZone(u)
}

// ---------------------------------------------------------------- lieu

type Place struct {
	expTick     float64 // école : rythme des paiements de la formation express
	yardH       bool    // zone du joueur à l'horizontale (supermarché)
	ID, Name    string
	W, H        float64
	Stations    []*Station
	Unlocks     []*Unlock
	Workers     []*Agent
	Customers   []*Customer
	Spots       []*Spot // améliorations achetées au sol (personnel, ventes)
	Rep         float64 // réputation, de 1 à 5 étoiles
	Owned       bool    // lieu acheté
	Price       float64 // prix d'achat sur la carte
	Scale       float64 // coefficient de tous les prix du lieu (rang sur la carte, prices.go) ; 0 : 1
	Served      int     // clients servis ici
	SellMult    float64 // multiplicateur des prix de vente
	SpawnBase   float64 // temps moyen entre deux clients (1 produit)
	WantMax     int     // nombre maximum de produits différents par client
	MaxCust     int
	RepGain     float64 // réputation gagnée par client content
	OnlyStocked bool    // les clients ne viennent chercher que ce qui est en rayon
	LeaveOff    V
	ups         []*Upgrade
	Rooms       []*Room    // école : salles de classe
	Students    []*Student // école : élèves
	regT        float64
	survJobs    map[*Agent]*survJob

	SpawnL, SpawnR V          // bouts de la route
	Parks          []*PSpot   // parking
	parkHalf       float64    // demi-largeur du parking (580 par défaut)
	Vehicles       []*Vehicle // voitures, vélos… des clients et de passage
	passT          float64
	cartSpot       V   // caddies : où on les prend et les rend (supermarché)
	cartsOut       int // caddies réservés par des clients
	Refused        int // clients repartis faute de place ou de caddie
	RefusedPark    int
	RefusedCart    int
	Gate           V // portail
	Inside         V // juste après le portail
	Exit           V
	RestSpot       V
	Start          V
	Yard           Rect    // zone du joueur (dalle de béton)
	Top            float64 // haut du terrain (négatif : la ferme monte au-dessus des champs)
	DecoTop        float64 // haut du décor que la caméra peut montrer (0 : Top)
	Water          []Rect  // l'eau (port) : personne n'y marche
	NoCust         []Rect  // interdit aux clients (l'entrepôt du supermarché)
	navC           *navGrid
	Boats          []*Boat // port : les bateaux de pêche
	dredges        []*Dredge
	boatsInit      bool
	Vestiaire      Rect // porte du vestiaire : on s'y arrête pour ouvrir la page « Personnage »

	spawnT   float64
	queueSeq int
	nav      *navGrid
	rng      *rand.Rand
	g        *Game
}

func (p *Place) Station(id string) *Station {
	for _, s := range p.Stations {
		if s.ID == id {
			return s
		}
	}
	return nil
}

func (p *Place) Register() *Station {
	for _, s := range p.Stations {
		if s.Kind == Register {
			return s
		}
	}
	return nil
}

// Registers : caisses construites.
func (p *Place) Registers() []*Station {
	var r []*Station
	for _, s := range p.Stations {
		if s.Kind == Register && s.Built {
			r = append(r, s)
		}
	}
	return r
}

// HasCashier : au moins une caissière embauchée.
func (p *Place) HasCashier() bool {
	for _, r := range p.Registers() {
		if r.Cashier != nil {
			return true
		}
	}
	return false
}

// manned : quelqu'un encaisse à cette caisse.
func (p *Place) manned(r *Station) bool {
	if r.Cashier != nil {
		return true
	}
	g := p.g
	return g.Cur == p && g.Player != nil && r.Zone.Has(g.Player.Pos)
}

// QueueAt : nombre de clients dans la file d'une caisse.
func (p *Place) QueueAt(r *Station) int {
	n := 0
	for _, c := range p.Customers {
		if c.State == cQueue && c.reg == r {
			n++
		}
	}
	return n
}

func (p *Place) key(s string) string { return p.ID + "." + s }

// Level renvoie le niveau d'une station (1 au départ).
func (p *Place) Level(s *Station) int { return 1 + p.g.Levels[p.key(s.ID)] }

func lvlFactor(l int) float64 { return 1 + 0.35*float64(l-1) }

// shelfFor renvoie l'étal construit qui vend un produit.
func (p *Place) shelfFor(it Item) *Station {
	for _, s := range p.Stations {
		if s.Kind == Shelf && s.Built && s.Item == it {
			return s
		}
	}
	return nil
}

// ---------------------------------------------------------------- mise à jour

func (p *Place) step(dt float64, active bool) {
	g := p.g
	if p.ID == "ecole" {
		p.stepSchool(dt, active)
		return
	}
	for _, s := range p.Stations {
		if !s.Built {
			continue
		}
		// vitesse : niveau du bâtiment × points de l'école (agronomes pour les champs,
		// soigneurs pour les animaux, mécaniciens pour les machines de la ferme, chefs ailleurs)
		f, rent := p.speedOf(s)
		if active && s.Kind == Machine {
			f *= p.toqueBoost(s) // toque de chef
		}
		switch s.Kind {
		case Field:
			for i := range s.Plants {
				if s.Plants[i].Young > 0 {
					// verger : l'arbre grandit d'abord (petit, moyen, grand), les fruits viendront ensuite
					s.Plants[i].Young = math.Max(0, s.Plants[i].Young-dt*f*p.waterBoost(s.Plants[i].P)/(TreeTime*s.GrowTime()))
					continue
				}
				if s.Plants[i].G >= 0 && s.Plants[i].G < 1 {
					s.Plants[i].G = math.Min(1, s.Plants[i].G+dt*f*p.waterBoost(s.Plants[i].P)/s.GrowTime())
				}
			}
		case Animal:
			s.Working = s.Pile < s.PileMax
			if s.Working {
				s.Timer += dt * f
				if s.Timer >= s.Time {
					s.Timer = 0
					s.Pile++
					g.made(s.Item)
				}
			}
		case Machine:
			ok := s.Pile < s.PileMax
			if s.Two {
				ok = s.Pile < s.PileMax || s.Pile2 < s.PileMax
			}
			for _, in := range s.Inputs {
				if s.Buf[in] == 0 {
					ok = false
				}
			}
			s.Working = ok
			if ok {
				s.Timer += dt * f
				if s.Timer >= s.Time {
					s.Timer = 0
					for _, in := range s.Inputs {
						if rent > 0 && p.rand() < rent {
							continue // rentabilité : cet ingrédient n'a pas été consommé
						}
						s.Buf[in]--
					}
					n := max(1, s.Yield)
					out, oz := s.Item, s.OutZone
					// deux sorties : chacune son tour (sauf si l'une est pleine)
					if s.Two && ((s.flip && s.Pile2 < s.PileMax) || s.Pile >= s.PileMax) {
						s.Pile2 = min(s.PileMax, s.Pile2+n)
						out, oz = s.Item2, s.OutZone2
					} else {
						s.Pile = min(s.PileMax, s.Pile+n)
					}
					s.flip = !s.flip
					for k := 0; k < n; k++ {
						g.made(out)
					}
					g.emit(Event{Kind: EvMade, Pos: oz.C(), It: out})
				}
			} else {
				s.Timer = math.Min(s.Timer, s.Time*0.999)
			}
		}
	}

	if p.hasDrones() {
		p.stepDrones(dt)
	}
	if p.ID == "port" {
		p.stepPort(dt)
	}
	// joueur
	if active && g.Player != nil {
		p.interact(g.Player, dt)
	}
	// employés
	p.syncStaff()
	for _, w := range p.Workers {
		p.workerAI(w, dt)
		p.interact(w, dt)
	}
	// clients
	p.spawnT -= dt
	if p.ID == "super" {
		p.syncBarriers()
		if p.spawnT <= 0 {
			// la demande : de quoi remplir le magasin (les visites durent quelques minutes)
			m := p.SuperMaxCust()
			p.spawnT = 100 / float64(max(1, m)) / p.RepFactor() / p.megaBoost() * (0.6 + 0.8*p.rng.Float64())
			if len(p.Customers) < m {
				p.superSpawn()
			}
		}
	} else if p.spawnT <= 0 {
		p.spawnT = p.spawnInterval() * (0.6 + 0.8*p.rng.Float64())
		n := 0
		for _, s := range p.Stations {
			if s.Kind == Shelf && s.Built {
				n++
			}
		}
		here := 0 // ceux encore dans leur voiture ne comptent pas
		for _, c := range p.Customers {
			if !c.InCar() && !c.toCar {
				here++
			}
		}
		if here < min(p.ShopMaxCust(), 3+2*n) {
			p.spawnCustomer()
		}
	}
	for _, c := range p.Customers {
		p.customerStep(c, dt)
	}
	p.stepVehicles(dt)
	alive := p.Customers[:0]
	for _, c := range p.Customers {
		if c.State != cGone {
			alive = append(alive, c)
		} else {
			c.Cart = false
			p.releaseCart(c) // (filet de sécurité : le caddie retourne dans sa colonne)
		}
	}
	p.Customers = alive
}

// ShopMaxCust : dans un petit magasin, 10 clients au plus en même temps
// (+1 tous les 10 % de marketing).
func (p *Place) ShopMaxCust() int {
	return 10 + p.g.Levels[p.key("pts_marketing")]/10
}

// spawnInterval : temps moyen entre deux clients.
func (p *Place) spawnInterval() float64 {
	n := 0
	for _, s := range p.Stations {
		if s.Kind == Shelf && s.Built {
			n++
		}
	}
	// plus il y a de produits, plus la ferme attire du monde ;
	// la réputation joue aussi (3 étoiles : normal)
	base := p.SpawnBase / (1 + 0.1*float64(n-1)) / p.RepFactor()
	f := (1 + p.ptsF("pts_marketing")) * p.megaBoost()
	return base / f
}

// ---------------------------------------------------------------- interactions

func (p *Place) carryCap(a *Agent) int {
	g := p.g
	if a.Kind == PlayerAgent {
		c := 6 + g.playerCapBonus()
		if c < 2 {
			c = 2
		}
		return c
	}
	// (4.1 : plus de brouettes ni de plateaux : seulement l'entraînement de l'école)
	c := 4 + g.Levels[p.key("capacite")] + g.Levels[p.key("pts_ent_cap")]/5
	if (p.ID == "ferme" && isColumnTeam(a.Team)) || (p.ID == "exotique" && a.Team == "fruits") {
		c = max(c, ColumnMax) // serre et verger : une cagette, pour emporter toute la colonne
	}
	return c
}

// isColumnTeam : une équipe de la serre ou du verger.
func isColumnTeam(id string) bool { return id == "serre" || id == "verger" }

func (p *Place) agentSpeed(a *Agent) float64 {
	g := p.g
	if a.Kind == PlayerAgent {
		return 300 * g.playerSpeedF()
	}
	return 170 * (1 + 0.15*float64(g.Levels[p.key("vitesse")])) * (1 + p.ptsF("pts_ent_vit")) * p.whistleBoost(a)
}

// interact fait un échange (au plus) toutes les quelques fractions de seconde.
func (p *Place) interact(a *Agent, dt float64) {
	g := p.g
	a.tick -= dt
	if a.tick > 0 {
		return
	}
	period := 0.09
	if a.Kind == WorkerAgent {
		period = 0.16 / (1 + p.ptsF("pts_mains")) // (points « mains expertes » de l'école)
	}
	capa := p.carryCap(a)
	// nouvelles pousses (joueur seulement)
	// (on ne paie que si l'on s'arrête sur la zone : traverser ne coûte rien)
	if a.Kind == PlayerAgent && !a.Moving && g.Money >= 1 {
		for _, s := range p.Stations {
			if !s.Built || s.Kind != Field || len(s.Plants) >= len(s.Slots) || !s.PlantZone().Has(a.Pos) {
				continue
			}
			cost := s.PlantCost()
			step := math.Max(1, math.Ceil(cost/25))
			pay := math.Min(step, math.Min(g.Money, cost-s.PlantPaid))
			g.Money -= pay
			s.PlantPaid += pay
			a.tick = 0.05
			g.emit(Event{Kind: EvPay, Pos: s.PlantZone().C()})
			if s.PlantPaid >= cost-1e-6 {
				s.PlantPaid = 0
				pos := s.Slots[len(s.Plants)]
				pl := Plant{P: pos}
				if s.Column && !s.Serre {
					pl.Young = 1 // verger : une case vide, où un arbre va pousser
				}
				s.Plants = append(s.Plants, pl)
				g.emit(Event{Kind: EvPlant, Pos: pos, It: s.Item, Text: "achat"}) // (une nouvelle case achetée)
			}
			return
		}
	}
	// améliorations posées à côté des champs et bâtiments (joueur arrêté)
	if a.Kind == PlayerAgent && !a.Moving && g.Money >= 1 {
		for _, s := range p.Stations {
			if !s.Built || !s.Upgradable() || !s.LevelZone().Has(a.Pos) {
				continue
			}
			u := p.StationUpgrade(s)
			l := g.Levels[u.Key]
			if l >= u.Max {
				continue
			}
			cost := u.Cost(l)
			step := math.Max(1, math.Ceil(cost/30))
			pay := math.Min(step, math.Min(g.Money, cost-s.LvlPaid))
			g.Money -= pay
			s.LvlPaid += pay
			a.tick = 0.05
			g.emit(Event{Kind: EvPay, Pos: s.LevelZone().C()})
			if s.LvlPaid >= cost-1e-6 {
				s.LvlPaid = 0
				g.Levels[u.Key] = l + 1
				g.emit(Event{Kind: EvLevel, Pos: s.LevelZone().C(), Text: s.Name + " : niveau " + itoa(l+2) + " !"})
			}
			return
		}
	}
	// agrandissement des stocks des bâtiments (joueur arrêté)
	if a.Kind == PlayerAgent && !a.Moving && g.Money >= 1 {
		for _, s := range p.Stations {
			if !s.Built || !s.Stockable() || !s.StockZone().Has(a.Pos) {
				continue
			}
			u := p.StockUpgrade(s)
			l := g.Levels[u.Key]
			if l >= u.Max {
				continue
			}
			cost := u.Cost(l)
			step := math.Max(1, math.Ceil(cost/30))
			pay := math.Min(step, math.Min(g.Money, cost-s.StkPaid))
			g.Money -= pay
			s.StkPaid += pay
			a.tick = 0.05
			g.emit(Event{Kind: EvPay, Pos: s.StockZone().C()})
			if s.StkPaid >= cost-1e-6 {
				s.StkPaid = 0
				g.Levels[u.Key] = l + 1
				p.applyCaps()
				g.emit(Event{Kind: EvLevel, Pos: s.StockZone().C(), Text: s.Name + " : stock agrandi !"})
			}
			return
		}
	}
	// améliorations du lieu posées au sol (caissière, hangar, bord de route)
	if a.Kind == PlayerAgent && !a.Moving && g.Money >= 1 {
		for _, sp := range p.Spots {
			u := p.SpotUpgrade(sp)
			l := g.Levels[u.Key]
			if l >= u.Max || !p.SpotVisible(sp) || !sp.Zone.Has(a.Pos) {
				continue
			}
			cost := u.Cost(l)
			step := math.Max(1, math.Ceil(cost/30))
			pay := math.Min(step, math.Min(g.Money, cost-sp.Paid))
			g.Money -= pay
			sp.Paid += pay
			a.tick = 0.05
			g.emit(Event{Kind: EvPay, Pos: sp.Zone.C()})
			if sp.Paid >= cost-1e-6 {
				sp.Paid = 0
				g.Levels[u.Key] = l + 1
				g.syncFarmTotal()
				p.applyCaps()
				g.emit(Event{Kind: EvLevel, Pos: sp.Zone.C(), Text: sp.Done(l + 1)})
			}
			return
		}
	}
	// déblocages (joueur seulement, arrêté)
	if a.Kind == PlayerAgent && !a.Moving {
		for _, u := range p.Unlocks {
			if u.Done || !p.Visible(u) || !u.Zone.Has(a.Pos) || g.Money < 1 {
				continue
			}
			step := math.Max(1, math.Ceil(u.Cost/45))
			pay := math.Min(step, math.Min(g.Money, u.Cost-u.Paid))
			g.Money -= pay
			u.Paid += pay
			a.tick = 0.05
			g.emit(Event{Kind: EvPay, Pos: u.Zone.C()})
			if u.Paid >= u.Cost-1e-6 {
				p.complete(u)
			}
			return
		}
	}
	for _, s := range p.Stations {
		if !s.Built {
			continue
		}
		switch s.Kind {
		case Field:
			// on replante les cases vides avec les graines que l'on porte
			if s.Replant && s.Zone.Has(a.Pos) && a.count(s.Seed) > 0 {
				r := 80.0
				if a.Kind == WorkerAgent {
					r = 45
				}
				best, bd := -1, r
				for i, pl := range s.Plants {
					if pl.Empty() {
						if d := pl.P.Dist(a.Pos); d < bd {
							best, bd = i, d
						}
					}
				}
				if best >= 0 {
					s.Plants[best].G = 0
					g.Stats.Replanted++
					a.remove(s.Seed)
					a.tick = period
					who := ""
					if a.Kind == WorkerAgent {
						who = "ouvrier" // (les ouvriers replantent sans bruit)
					}
					g.emit(Event{Kind: EvPlant, Pos: s.Plants[best].P, It: s.Item, Text: who})
					return
				}
			}
			if !s.Zone.Has(a.Pos) || a.load() >= capa {
				continue
			}
			if a.Kind == WorkerAgent && (a.job == nil || a.job.phase != 0 || a.job.src != s) {
				continue
			}
			// on ne ramasse pas ce qui ne sert à rien (ex. du blé sans moulin ni étable)
			if a.Kind == PlayerAgent && !p.useful(s.Item) {
				continue
			}
			r := 80.0
			if a.Kind == WorkerAgent {
				r = 45
			}
			best, bd := -1, r
			for i, pl := range s.Plants {
				if pl.G >= 1 {
					if d := pl.P.Dist(a.Pos); d < bd {
						best, bd = i, d
					}
				}
			}
			if best >= 0 {
				s.Plants[best].G = 0
				if s.Replant {
					s.Plants[best].G = -1 // il faudra replanter
				}
				n := min(FieldYield(p.Level(s)), capa-a.load())
				if a.job != nil {
					n = max(1, min(n, a.job.want-a.job.got)) // un ouvrier ne prend que ce qu'il lui faut
				}
				for k := 0; k < n; k++ {
					a.Carry = append(a.Carry, s.Item)
				}
				a.tick = period
				g.Stats.Harvested += n
				g.emit(Event{Kind: EvHarvest, Pos: s.Plants[best].P, It: s.Item})
				if a.job != nil {
					a.job.got += n
				}
				return
			}
		case Animal, Machine, Pallet:
			out := s.Zone
			if s.Kind == Machine {
				out = s.OutZone
			}
			if p.ID == "ecole" && a.Kind == PlayerAgent && a.count(s.Item) >= p.wanted(s.Item) {
				continue // à l'école, on ne prend que le matériel demandé
			}
			if IsSeed(s.Item) && a.Kind == PlayerAgent && a.count(s.Item) >= p.Empties(s.Item) {
				continue // on ne prend que les graines qu'on va planter
			}
			if s.Two && s.OutZone2.Has(a.Pos) && s.Pile2 > 0 && a.load() < capa && (a.Kind == PlayerAgent && p.useful(s.Item2) || (a.job != nil && a.job.phase == 0 && a.job.src == s && a.job.it == s.Item2)) {
				s.Pile2--
				a.Carry = append(a.Carry, s.Item2)
				a.tick = period
				g.emit(Event{Kind: EvTake, Pos: s.OutZone2.C(), It: s.Item2})
				if a.job != nil {
					a.job.got++
				}
				return
			}
			if out.Has(a.Pos) && s.Pile > 0 && (a.load() < capa || IsSeed(s.Item)) && (a.Kind == PlayerAgent || (a.job != nil && a.job.phase == 0 && a.job.src == s && a.job.it == s.Item)) {
				if !IsSeed(s.Item) {
					s.Pile-- // (les sachets de graines ne s'épuisent pas)
				}
				a.Carry = append(a.Carry, s.Item)
				a.tick = period
				g.emit(Event{Kind: EvTake, Pos: out.C(), It: s.Item})
				if a.job != nil {
					a.job.got++
				}
				return
			}
			if s.Kind == Machine && s.Zone.Has(a.Pos) {
				for _, in := range s.Inputs {
					if s.Buf[in] < s.BufMax && a.count(in) > 0 {
						a.remove(in)
						s.Buf[in]++
						a.tick = period
						g.emit(Event{Kind: EvDrop, Pos: s.Zone.C(), It: in})
						return
					}
				}
			}
		case Shelf:
			if s.Zone.Has(a.Pos) && s.Stock < s.StockMax && a.count(s.Item) > 0 {
				a.remove(s.Item)
				s.Stock++
				a.tick = period
				g.emit(Event{Kind: EvDrop, Pos: s.Body.C(), It: s.Item})
				return
			}
		case Store:
			// on y dépose ; l'ouvrier seulement si c'est son travail (le surplus)
			if s.Zone.Has(a.Pos) && s.Stock < s.StockMax && a.count(s.Item) > 0 &&
				(a.Kind == PlayerAgent || (a.job != nil && a.job.dst == s)) {
				a.remove(s.Item)
				s.Stock++
				a.tick = period
				g.emit(Event{Kind: EvDrop, Pos: s.Body.C(), It: s.Item})
				return
			}
		case OrderBox:
			if a.Kind == PlayerAgent && s.Zone.Has(a.Pos) && g.Order != nil && g.Order.Place() == p.ID {
				for _, it := range a.Carry {
					if g.Order.Need(it) > 0 {
						a.remove(it)
						g.Order.add(it)
						a.tick = period
						g.emit(Event{Kind: EvDrop, Pos: s.Body.C(), It: it})
						if g.Order.Done() {
							g.finishOrder(s.Body.C())
						}
						return
					}
				}
			}
		case Dock:
			if s.Zone.Has(a.Pos) {
				for _, it := range a.Carry {
					if p.need(s, it)+s.claim[it] > 0 || (a.Kind == PlayerAgent && p.g.dockAccepts(s, it) && s.Buf[it] < s.BufMax) {
						a.remove(it)
						s.Buf[it]++
						a.tick = period
						g.emit(Event{Kind: EvDrop, Pos: s.Body.C(), It: it})
						return
					}
				}
			}
		case Trash:
			// on ne jette que volontairement : le joueur s'arrête dessus,
			// l'ouvrier y vient exprès (il n'a trouvé aucune livraison)
			wants := (a.Kind == PlayerAgent && !a.Moving) || (a.Kind == WorkerAgent && a.job == nil && a.noDestT >= 4)
			if wants && s.Zone.Has(a.Pos) && len(a.Carry) > 0 {
				a.Carry = a.Carry[:len(a.Carry)-1]
				a.tick = period
				g.emit(Event{Kind: EvTrash, Pos: s.Body.C()})
				return
			}
		}
	}
	a.tick = 0.03
}

// Visible : on ne montre que les deux prochains déblocages.
func (p *Place) Visible(u *Unlock) bool {
	if p.ID == "super" || p.ID == "ecole" {
		return !u.Done && (u.Requires == nil || u.Requires())
	}
	if p.ID == "boucherie" {
		return !u.Done && (u.Requires == nil || u.Requires()) && p.firstInZone(u)
	}
	if u.Requires != nil && !u.Requires() {
		return false // (montré verrouillé, avec la raison)
	}
	shown := 0
	for _, v := range p.Unlocks {
		if v.Done || (v.Requires != nil && !v.Requires()) {
			continue
		}
		if v == u {
			return shown < 2
		}
		shown++
	}
	return false
}

func (p *Place) complete(u *Unlock) {
	u.Done = true
	u.Paid = u.Cost
	for _, s := range p.Stations {
		if s.Unlock == u.ID {
			s.Built = true
		}
	}
	p.g.emit(Event{Kind: EvUnlock, Pos: u.Zone.C(), Text: u.Name})
	// un client vient tout de suite
	p.spawnT = math.Min(p.spawnT, 1)
}

// ---------------------------------------------------------------- déplacement

func (p *Place) solidAt(q V, r float64) bool {
	for _, s := range p.Stations {
		if !s.Built || !s.Solid() {
			continue
		}
		if s.Body.Grow(r).Has(q) {
			return true
		}
	}
	if q.X < p.MinX()+20 || q.Y < p.Top+20 || q.X > p.MaxX()-20 || q.Y > p.H-20 {
		return true
	}
	for _, w := range p.Water {
		if w.Grow(r).Has(q) {
			return true
		}
	}
	return false
}

// moveAgent déplace dans la direction voulue en glissant le long des murs.
func (p *Place) moveAgent(pos *V, dir V, dist float64) bool {
	if dir.Len() < 1e-6 || dist <= 0 {
		return false
	}
	d := dir.Norm().Mul(dist)
	const r = 16.0
	if p.solidAt(*pos, r) {
		// coincé dans un bâtiment qui vient d'apparaître : on en sort librement
		*pos = pos.Add(d)
		return true
	}
	moved := false
	nx := V{pos.X + d.X, pos.Y}
	if !p.solidAt(nx, r) {
		*pos = nx
		moved = true
	}
	ny := V{pos.X, pos.Y + d.Y}
	if !p.solidAt(ny, r) {
		*pos = ny
		moved = true
	}
	return moved
}

// MovePlayer applique la direction choisie au clavier (longueur ≤ 1).
func (g *Game) MovePlayer(dir V, dt float64) {
	pl := g.Player
	p := g.Cur
	if p == nil || pl == nil {
		return
	}
	pl.Moving = dir.Len() > 0.05
	if pl.Moving {
		pl.Face = dir.Norm()
		p.moveAgent(&pl.Pos, dir, p.agentSpeed(pl)*math.Min(1, dir.Len())*dt)
		pl.WalkT += dt
	}
}

// walk rapproche un PNJ de sa cible, en contournant les obstacles s'il bloque.
func (p *Place) walk(a *Agent, to V, speed, dt float64) bool {
	d := to.Sub(a.Pos)
	if d.Len() < 8 {
		a.Moving = false
		return true
	}
	a.Moving = true
	a.WalkT += dt
	// un bâtiment entre nous et la cible ? on calcule un chemin qui le contourne
	a.routeT -= dt
	if a.routeTo.Dist(to) > 20 || a.routeT <= 0 {
		a.route = nil
		a.routeTo = to
		a.routeT = 1.5
		if !p.clear(a.Pos, to) {
			a.route = p.route(a.Pos, to)
		}
	}
	for len(a.route) > 1 && a.Pos.Dist(a.route[0]) < 14 {
		a.route = a.route[1:]
	}
	next := to
	if len(a.route) > 0 {
		next = a.route[0]
	}
	nd := next.Sub(a.Pos)
	step := math.Min(nd.Len(), speed*dt)
	dir := nd.Norm()
	if a.sideT > 0 {
		a.sideT -= dt
		dir = a.sideDir.Add(dir.Mul(0.3)).Norm()
	}
	a.Face = dir
	p.moveAgent(&a.Pos, dir, step)
	// coincé ?
	a.stuckT += dt
	if a.stuckT > 0.4 {
		if a.Pos.Dist(a.lastPos) < speed*0.4*0.25 {
			// on fait un petit pas de côté (par rapport au prochain point du chemin),
			// court même quand on va vite, puis on recalcule le chemin
			side := V{-nd.Y, nd.X}.Norm()
			if p.rng.Intn(2) == 0 {
				side = side.Mul(-1)
			}
			a.sideDir = side
			a.sideT = (0.3 + 0.3*p.rng.Float64()) * 170 / math.Max(170, speed)
			a.routeT = 0
		}
		a.stuckT = 0
		a.lastPos = a.Pos
	}
	return false
}

func (p *Place) rand() float64 { return p.rng.Float64() }

// PlantCost : prix de la prochaine pousse d'un champ (de plus en plus cher).
// (une case de serre ou de verger : bien plus chère qu'une pousse de maïs ;
// les lieux loin sur la carte : tout est plus cher, prices.go)
func (s *Station) PlantCost() float64 {
	c := s.rawPlantCost(len(s.Plants))
	if s.scale > 0 && s.scale != 1 {
		c = niceRound(c * s.scale)
	}
	return c
}

// Empties : cases vides (à replanter) de la colonne qui utilise ces graines.
func (p *Place) Empties(seed Item) int {
	n := 0
	for _, s := range p.Stations {
		if s.Built && s.Replant && s.Seed == seed {
			for _, pl := range s.Plants {
				if pl.Empty() {
					n++
				}
			}
		}
	}
	return n
}

// PlantZone : case où l'on se place pour acheter la prochaine pousse.
func (s *Station) PlantZone() Rect {
	if len(s.Plants) >= len(s.Slots) {
		return Rect{}
	}
	c := s.Slots[len(s.Plants)]
	return Rect{c.X - 38, c.Y - 40, 76, 76}
}

// Upgradable : champs, animaux et machines ont des niveaux.
func (s *Station) Upgradable() bool {
	if s.place == "ecole" {
		return false // les étagères de fournitures n'ont pas de niveaux
	}
	// (4.1 : plus de niveaux du tout : seuls les points de l'école font aller plus vite
	// les champs, les animaux et les machines ; il reste l'augmentation des stocks)
	return false
}

// LevelZone : case d'amélioration, à côté du champ ou du bâtiment.
func (s *Station) LevelZone() Rect {
	b := s.Body
	if s.Kind == Field {
		return Rect{b.X + b.W - 104, b.Y + b.H + 14, 100, 84}
	}
	return Rect{b.X + b.W + 14, b.Y + 40, 104, 104}
}

// Spot : case au sol pour une amélioration du lieu.
type Spot struct {
	ID         string // suffixe de la clé d'amélioration (caissiere, ouvriers…)
	Label      string
	Zone       Rect
	Paid       float64
	Done       func(lvl int) string // message une fois acheté
	Needs      string               // n'apparaît qu'une fois cette amélioration achetée
	Team       string               // embauche d'un ouvrier de cette zone (ferme)
	Shelf      *Station             // achat du drone de cet étal (ferme)
	NeedsBuilt string               // n'apparaît qu'une fois cette station construite (port), tant qu'il reste à acheter
	Room       *Room                // école : table ou prof de cette salle
}

func (p *Place) SpotUpgrade(sp *Spot) *Upgrade {
	for _, u := range p.PlaceUpgrades() {
		if u.Key == p.key(sp.ID) {
			return u
		}
	}
	return nil
}

func (p *Place) SpotVisible(sp *Spot) bool {
	if p.spotHidden(sp) {
		return false
	}
	if sp.Room != nil {
		u := p.SpotUpgrade(sp)
		return sp.Room.built && p.g.Levels[u.Key] < u.Max
	}
	if sp.Shelf != nil {
		return p.g.Levels[p.key(sp.ID)] > 0 || p.droneUseful(sp.Shelf)
	}
	if sp.Team != "" && p.ID == "super" {
		z := superZone(sp.Team)
		u := p.SpotUpgrade(sp)
		return z != nil && z.Open(p.g) && u != nil && p.g.Levels[u.Key] < u.Max
	}
	if sp.Team != "" && teamZonesOf(p.ID) != nil {
		u := p.SpotUpgrade(sp)
		return p.ZoneOpen(sp.Team) && u != nil && p.g.Levels[u.Key] < u.Max
	}
	if sp.Team != "" {
		// dès que la zone existe, tant qu'il reste des ouvriers à embaucher
		// (les prix guident vers l'ordre maïs, blé, maïs, blé…)
		f := p.Station(teamAnchor(sp.Team))
		u := p.SpotUpgrade(sp)
		if isColumnTeam(sp.Team) && p.g.Levels[u.Key] >= p.columnHires(sp.Team) {
			return false // (un ouvrier par colonne ouverte)
		}
		return f != nil && f.Built && p.g.Levels[u.Key] < u.Max
	}
	if sp.NeedsBuilt != "" {
		st := p.Station(sp.NeedsBuilt)
		u := p.SpotUpgrade(sp)
		return st != nil && st.Built && u != nil && p.g.Levels[u.Key] < u.Max
	}
	return sp.Needs == "" || p.g.Levels[p.key(sp.Needs)] > 0
}

// RepFactor : effet de la réputation sur la fréquentation (×0,67 à ×1,33).
func (p *Place) RepFactor() float64 { return 0.5 + p.Rep/6 }

// reputation fait évoluer la réputation après le passage d'un client.
func (p *Place) reputation(d float64) {
	p.Rep = math.Max(1, math.Min(5, p.Rep+d))
}

// ---------------------------------------------------------------- stocks

// Stockable : bâtiments dont on peut agrandir le stock (entrée et sortie).
func (s *Station) Stockable() bool {
	if s.place == "elevage" {
		return elevageAnimals[s.ID] // (les bâtiments d'animaux, pas les abattoirs)
	}
	return (s.Kind == Machine || s.Kind == Animal) && s.place != "ecole"
}

// StockZone : case « stock », sous la case de niveau, à droite du bâtiment.
func (s *Station) StockZone() Rect {
	b := s.Body
	return Rect{b.X + b.W + 14, b.Y + 40, 104, 104} // (à la place de l'ancienne case « Niveau »)
}

// StockUpgrade : agrandir le stock d'un bâtiment.
func (p *Place) StockUpgrade(s *Station) *Upgrade {
	return &Upgrade{Key: p.key(s.ID + "_stock"), Name: s.Name + " : stock", Max: 5, Base: math.Max(60, s.BaseCost*0.5), Mult: 2.2,
		Desc: func(l int) string { return "Stock du bâtiment agrandi (niveau " + itoa(l) + ")." }}
}

// applyCaps recalcule les capacités selon les améliorations achetées.
func (p *Place) applyCaps() {
	g := p.g
	for _, s := range p.Stations {
		if s.baseBuf == 0 && s.basePile == 0 && s.baseStock == 0 {
			s.baseBuf, s.basePile, s.baseStock = s.BufMax, s.PileMax, s.StockMax
		}
		if s.Stockable() {
			l := g.Levels[p.key(s.ID+"_stock")]
			s.BufMax = s.baseBuf + 5*l
			s.PileMax = s.basePile + 6*l
		}
		if s.Kind == Shelf {
			s.StockMax = s.baseStock + 4*g.Levels[p.key("etals")]
		}
	}
}

// useful : ce produit peut-il aller quelque part dans ce lieu (étal, machine, quai) ?
func (p *Place) useful(it Item) bool {
	if IsSeed(it) {
		return p.Empties(it) > 0
	}
	for _, s := range p.Stations {
		if !s.Built {
			continue
		}
		switch s.Kind {
		case Shelf:
			if s.Item == it {
				return true
			}
		case Machine:
			for _, in := range s.Inputs {
				if in == it {
					return true
				}
			}
		case Dock:
			if p.g.dockAccepts(s, it) {
				return true
			}
		}
	}
	return false
}

// FieldTimeMult : durée de pousse d'un champ complet, par rapport à une seule pousse.
var FieldTimeMult = 6.0

// GrowTime : durée de pousse au niveau 1. Les premières pousses d'un champ
// poussent vite (bon départ) ; un grand champ pousse plus lentement, pour que
// les améliorations de vitesse servent vraiment.
func (s *Station) GrowTime() float64 {
	if s.Column {
		return s.Time * (1 + 0.25*float64(len(s.Plants)-1)) // 8 cases : ×2,75
	}
	k := math.Min(1, float64(len(s.Plants)-1)/10)
	return s.Time * (1 + (FieldTimeMult-1)*k)
}

// FieldYield : récoltes par pousse mûre selon le niveau du champ.
func FieldYield(level int) int {
	switch {
	case level >= 9:
		return 3
	case level >= 5:
		return 2
	}
	return 1
}

// CashTime : secondes par objet passé en caisse (caissière ; le joueur va plus vite).
func (p *Place) CashTime(r *Station) float64 {
	if r.Cashier == nil {
		return CashPlayer // le joueur
	}
	if p.ID == "super" {
		return CashPlayer / (1 + 2*p.ptsF("pts_caisse")) // caisses à scanner
	}
	// (4.1 : plus de tiroir-caisse ni de caisse rapide : seulement les points « Caisses » de l'école)
	return CashBase / (1 + 2*p.ptsF("pts_caisse"))
}

// cashUpgrade : vitesse d'encaissement des caissières d'un lieu.
func (p *Place) cashUpgrade(name string, base float64) *Upgrade {
	return &Upgrade{Key: p.key("encaisse"), Name: name, Max: 8, Base: base, Mult: 2.1,
		Desc: func(l int) string {
			t := CashBase / (1 + 0.5*float64(l))
			return "Une caissière passe un objet en " + ftoa1(t) + " s (toi : " + ftoa1(CashPlayer) + " s)."
		}}
}

func ftoa1(v float64) string {
	n := int(math.Round(v * 10))
	if n%10 == 0 {
		return itoa(n / 10)
	}
	return itoa(n/10) + "," + itoa(n%10)
}

// temps d'encaissement d'un objet (secondes)
var CashBase, CashPlayer = 1.2, 0.5

// speedOf : vitesse d'une station (niveau × points de l'école) et chance de ne
// pas consommer un ingrédient (rentabilité).
func (p *Place) speedOf(s *Station) (f, rent float64) {
	f, rent = p.speedOf0(s)
	if (p.ID == "elevage" && elevageAnimals[s.ID]) || (p.ID == "ferme" && animalIDs[s.ID]) {
		f *= p.g.boostF("veto", "") // cadeau du vétérinaire : animaux (et abeilles) ×2
	}
	return
}

func (p *Place) speedOf0(s *Station) (f, rent float64) {
	g := p.g
	switch {
	case s.Kind == Field:
		return 1 + p.ptsF("pts_agro"), 0
	// (plus de niveaux : +3 % par point de soigneur, de mécanicien ou de chef)
	case p.ID == "elevage" && elevageAnimals[s.ID]:
		return 1 + 3*p.ptsF("pts_soin"), RentSave(g.Levels[p.key("pts_soin_rent")])
	case p.ID == "elevage" && s.Kind == Machine:
		return 1 + 3*p.ptsF("pts_mac_vit"), RentSave(g.Levels[p.key("pts_mac_rent")])
	case p.ID == "exotique" && s.Kind == Machine:
		return 1 + 3*p.ptsF("pts_mac_vit"), RentSave(g.Levels[p.key("pts_mac_rent")])
	case p.ID == "ferme" && animalIDs[s.ID]:
		return 1 + 3*p.ptsF("pts_soin"), RentSave(g.Levels[p.key("pts_soin_rent")]) // (sélection génétique)
	case p.ID == "ferme" && s.Kind == Machine:
		return 1 + 3*p.ptsF("pts_mac_vit"), RentSave(g.Levels[p.key("pts_mac_rent")])
	case p.cooks() && s.Kind == Machine:
		return 1 + 3*p.ptsF("pts_chef_vit"), RentSave(g.Levels[p.key("pts_chef_rent")])
	}
	return lvlFactor(p.Level(s)), 0
}

// UnlockNow : débloque tout de suite (captures d'écran).
func (p *Place) UnlockNow(id string) {
	for _, u := range p.Unlocks {
		if u.ID == id && !u.Done {
			p.complete(u)
		}
	}
}
