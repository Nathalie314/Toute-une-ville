package sim

import (
	"math"

	"ville/tr"
)

// LES VOLEURS (événements, étape 2). Dès que le supermarché est ouvert, les voleurs font partie
// des groupes qui peuvent se manifester dans un lieu de la rangée du haut (eventGroups) : 50,
// plus 5 par lieu débloqué de la rangée du haut (ils attaquent par 5 ; plus la ville grossit,
// plus elle les attire). Un vol dure autant de minutes qu'il y a de lieux débloqués en haut.
//
// Pendant un vol, dans le lieu attaqué : les clients fuient en jetant leurs achats, sans payer ;
// les employés qui se déplacent se réfugient au vestiaire ; ceux qui ne bougent pas (caissières,
// serveurs de la crêperie, employés des attractions) lèvent les bras ; les drones jettent leur
// chargement et restent sur leur base ; plus aucun camion ne part ni n'arrive. Les champs
// poussent et les animaux produisent toujours. Cinq voleurs en noir (visage découvert : la
// vidéosurveillance servira) vident chaque caisse une fois : une minute des gains du lieu.
// À la crêperie, sans caisse, ils raflent les pourboires laissés sur les tables.
// Le joueur, lui, ne peut pas sortir de sa zone tant que les voleurs sont là ; s'il est ailleurs
// dans le lieu quand ils arrivent, terrorisé, il y court tout seul (on ne le dirige plus avant).
// Les événements n'arrivent que pendant qu'on joue ; un vol en cours est sauvegardé (TheftSave).

const (
	thievesBase    = 50
	thievesPerLieu = 5
	robbersN       = 5
	robberSpeed    = 240.0
)

// Theft : le vol en cours.
type Theft struct {
	Place   string
	Left    float64 // secondes restantes
	Total   float64
	Robbers []*Robber
	Stolen  float64 // ce qui a été pris dans les caisses
	robbed  map[int]bool
	Fire    bool         // un incendie (pas de voleurs : à la fin, le lieu brûle ; fire.go)
	Spotted bool         // repérés par la vidéosurveillance : ils seront arrêtés en partant
	targets []V          // les caisses (ou, sans caisse, quelques points du lieu)
	tables  []*CrepTable // à la crêperie : les tables, dont ils raflent les pourboires
	hasRegs bool
	Scared  bool // le joueur, terrorisé, court se réfugier dans sa zone (on ne le dirige plus)
}

// Robber : un voleur.
type Robber struct {
	Pos, Face V
	WalkT     float64
	Moving    bool
	Look      int
	path      []V
	target    int // caisse visée (-1 : aucune)
	waitT     float64
}

// TopOwned : le nombre de lieux débloqués de la rangée du haut de la carte.
func (g *Game) TopOwned() int {
	n := 0
	for _, id := range TopRow() {
		if q := g.Place(id); q != nil && q.Owned {
			n++
		}
	}
	return n
}

// Thieves : le nombre de voleurs dans la ville.
// (ceux qui sont en garde à vue n'en font plus partie : ils ne peuvent plus attaquer)
func (g *Game) Thieves() int {
	return max(0, thievesBase+thievesPerLieu*g.TopOwned()-g.InCustody()-g.PrisonHeld()-g.Levels["prison.reinseres"])
}

// thievesActive : les voleurs sont-ils arrivés en ville (le supermarché est ouvert) ?
func (g *Game) thievesActive() bool {
	q := g.Place("super")
	return q != nil && q.Owned
}

// TheftAt : le vol en cours dans ce lieu (nil sinon).
func (g *Game) TheftAt(id string) *Theft {
	if g.Theft != nil && g.Theft.Place == id {
		return g.Theft
	}
	return nil
}

func (p *Place) robbed() bool { return p.g != nil && p.g.TheftAt(p.ID) != nil }

// startTheft : les voleurs attaquent ce lieu.
func (g *Game) startTheft(id string) {
	p := g.Place(id)
	if p == nil {
		return
	}
	d := 60 * float64(max(1, g.TopOwned()))
	g.startTheftQuiet(id, d, min(robbersN, g.Thieves()))
	p.theftFlee()
	if g.Cur == p && p.Yard.W > 0 && !p.Yard.Has(g.Player.Pos) {
		g.Theft.Scared = true // (terrorisé, le joueur court tout seul jusqu'à sa zone)
	}
	g.emit(Event{Kind: EvLevel, Place: id, Pos: p.Gate, Text: tr.F("Des voleurs attaquent {lieu} ! Tout s'arrête pendant {duree}.", "lieu", theName(id), "duree", durText(d))})
	g.emit(Event{Kind: EvAngry, Place: id, Pos: p.Gate})
	if n := g.CameraOf(id); n > 0 && g.rng.Float64() < g.CatchChance(n) { // (un seul tirage pour tout le groupe)
		g.Theft.Spotted = true
		g.emit(Event{Kind: EvLevel, Place: id, Pos: p.Gate, Text: "La vidéosurveillance a repéré les voleurs : la police les attend à la sortie !"})
	}
}

// startTheftQuiet : met les voleurs en place (sans message).
func (g *Game) startTheftQuiet(id string, d float64, nr int) {
	p := g.Place(id)
	th := &Theft{Place: id, Left: d, Total: d, robbed: map[int]bool{}}
	for _, s := range p.Stations {
		if s.Kind == Register && s.Built {
			th.targets = append(th.targets, s.Body.C().Add(V{0, -60}))
		}
	}
	if p.Crep != nil { // (pas de caisse à la crêperie : les pourboires laissés sur les tables)
		for _, t := range p.Crep.Tables {
			if t.St == nil || t.St.Built {
				th.tables = append(th.tables, t)
				th.targets = append(th.targets, t.Pos.Add(V{0, -70}))
			}
		}
	}
	if p.Parc != nil {
		for _, r := range p.Parc.Regs {
			th.targets = append(th.targets, r.Zone().C())
		}
	}
	th.hasRegs = len(th.targets) > 0
	if !th.hasRegs { // pas de caisse : ils rôdent au milieu du lieu
		c := p.Inside
		for k := 0; k < robbersN; k++ {
			a := float64(k) / robbersN * 2 * math.Pi
			th.targets = append(th.targets, c.Add(V{math.Cos(a) * 220, math.Sin(a) * 120}))
		}
	}
	for k := 0; k < max(1, nr); k++ { // (moins de 5 voleurs en liberté : ils viennent à ce qu'ils sont)
		r := &Robber{Pos: p.Gate.Add(V{float64(k-2) * 40, 30}), Face: V{0, -1}, Look: 500 + 97*k, target: -1, waitT: float64(k) * 0.6}
		th.Robbers = append(th.Robbers, r)
	}
	g.Theft = th
}

// theftFlee : au début du vol, tous les clients s'enfuient (sans payer, en jetant leurs achats).
func (p *Place) theftFlee() {
	for _, c := range p.Customers {
		p.fleeCustomer(c)
	}
	if cp := p.Crep; cp != nil {
		for _, cg := range cp.Groups {
			if cg.State != cgLeave && cg.State != cgGone {
				p.crepLeave(cg)
			}
		}
	}
	if pc := p.Parc; pc != nil {
		for _, r := range pc.Rides {
			r.Queue, r.Riders, r.State, r.Prog = nil, nil, 0, 0
		}
		for _, r := range pc.Regs {
			r.Queue = nil
		}
		for _, v := range pc.Visitors {
			p.parcFlee(v)
		}
	}
}

func (p *Place) fleeCustomer(c *Customer) {
	if c.State == cLeave || c.State == cGone || c.InCar() {
		return
	}
	c.Mood = 2
	c.State = cLeave
	c.Carry = nil
	c.CheckN, c.CheckT = 0, 0
	c.path = p.leavePath(c)
}

// parcFlee : un visiteur du parc s'en va tout de suite.
func (p *Place) parcFlee(v *Visitor) {
	if v.State == vCar || v.State == vLeave || v.State == vGone || len(p.Parc.Rides) == 0 {
		return
	}
	v.Joy = 0
	p.parcNext(v, v.Pos, p.Parc.Rides[0]) // (jauge à zéro : il prend le chemin de la sortie)
}

// stepTheft : les voleurs vident les caisses ; au bout du temps, ils s'en vont.
func (g *Game) stepTheft(dt float64) {
	th := g.Theft
	p := g.Place(th.Place)
	th.Left -= dt
	if th.Left <= 0 || p == nil {
		g.Theft = nil
		if th.Fire { // (l'incendie s'éteint : ce qui n'est pas sauvé est détruit)
			if p != nil {
				g.burn(p)
			}
			g.orderT = g.eventGap()
			return
		}
		if th.Spotted {
			g.arrest(th.Place, len(th.Robbers), CustodyTime)
		}
		if p != nil {
			for _, w := range p.Workers {
				w.Hidden = false // (ils sortent du vestiaire)
			}
			g.emit(Event{Kind: EvLevel, Place: p.ID, Pos: p.Gate, Text: tr.F("Les voleurs sont partis de {lieu}. Tout reprend.", "lieu", theName(p.ID))})
		}
		g.orderT = g.eventGap()
		return
	}
	if th.Scared {
		g.playerFlee(p, dt)
	}
	if th.Fire {
		return
	}
	for i, r := range th.Robbers {
		if r.waitT > 0 {
			r.waitT -= dt
			continue
		}
		if r.target < 0 { // la caisse suivante pas encore vidée (ou un point où rôder)
			r.target = (i + int(th.Total-th.Left)/20) % len(th.targets)
			for k := range th.targets { // (la suivante qu'aucun autre voleur ne vise déjà)
				if th.hasRegs && !th.robbed[k] && !th.aimed(k, r) {
					r.target = k
					break
				}
			}
			r.path = p.route(r.Pos, th.targets[r.target].Add(V{float64(i-2) * 30, 0}))
		}
		if len(r.path) == 0 {
			r.Moving = false
			if th.hasRegs && !th.robbed[r.target] {
				th.robbed[r.target] = true
				if th.tables != nil {
					g.robTips(p, th, th.tables[r.target])
				} else {
					g.robRegister(p, th, th.targets[r.target])
				}
			}
			r.waitT = 4 + 3*g.rng.Float64()
			r.target = -1
			continue
		}
		to := r.path[0]
		d := to.Sub(r.Pos)
		step := robberSpeed * dt
		r.Moving = true
		r.WalkT += dt
		if d.Len() <= step {
			r.Pos = to
			r.path = r.path[1:]
		} else {
			r.Face = d.Norm()
			r.Pos = r.Pos.Add(r.Face.Mul(step))
		}
	}
}

// aimed : un autre voleur vise-t-il déjà cette caisse ?
func (th *Theft) aimed(k int, me *Robber) bool {
	for _, o := range th.Robbers {
		if o != me && o.target == k {
			return true
		}
	}
	return false
}

// robRegister : une caisse vidée : une minute des gains du lieu (si le joueur a de quoi).
func (g *Game) robRegister(p *Place, th *Theft, at V) {
	earn, span := g.EarningsByPlace(300)
	rate := 0.0
	if span > 0 {
		rate = earn[p.ID] / span
	}
	a := math.Min(math.Round(60*rate), math.Max(0, g.Money))
	if a <= 0 {
		return
	}
	g.Money -= a
	th.Stolen += a
	g.emit(Event{Kind: EvLevel, Place: p.ID, Pos: at, Text: tr.F("Les voleurs ont vidé une caisse : −{prix}", "prix", moneyText(a))})
}

// robTips : à la crêperie, les voleurs raflent les pourboires laissés sur une table.
func (g *Game) robTips(p *Place, th *Theft, t *CrepTable) {
	if t.Tips <= 0 {
		return
	}
	a := t.Tips
	t.Tips = 0
	th.Stolen += a
	g.emit(Event{Kind: EvLevel, Place: p.ID, Pos: t.Pos, Text: tr.F("Les voleurs ont raflé les pourboires d'une table : −{prix}", "prix", moneyText(a))})
}

// playerFlee : le joueur terrorisé court jusqu'à sa zone ; là seulement, on le dirige de nouveau.
func (g *Game) playerFlee(p *Place, dt float64) {
	pl := g.Player
	if g.Cur != p || pl == nil {
		g.Theft.Scared = false
		return
	}
	in := p.Yard
	in.X, in.Y, in.W, in.H = in.X+40, in.Y+40, in.W-80, in.H-80
	if in.Has(pl.Pos) {
		g.Theft.Scared = false
		pl.Moving, pl.route = false, nil
		return
	}
	p.walk(pl, p.Start, p.agentSpeed(pl)*1.4, dt)
}

// hideWorker : pendant un vol, l'employé court se réfugier au vestiaire.
func (p *Place) hideWorker(w *Agent, dt float64) {
	if w.Hidden {
		return
	}
	to := p.Vestiaire.C()
	if p.Vestiaire.W == 0 {
		to = p.Gate
	}
	w.route = nil
	if p.walk(w, to, p.agentSpeed(w)*1.4, dt) || w.Pos.Dist(to) < 40 {
		w.Hidden = true
		w.Moving = false
	}
}

// StartTheft : lance un vol tout de suite dans ce lieu (captures d'écran, essais).
func (g *Game) StartTheft(id string) {
	if g.Theft == nil {
		g.Order = nil
		g.startTheft(id)
	}
}

// TheftSave : un vol en cours, dans la sauvegarde. Son temps s'écoule aussi jeu fermé : les
// voleurs partent à l'heure prévue. Revenir avant, c'est les retrouver (quitter ne les chasse pas) ;
// pendant l'absence, le lieu ne perd que le reste du vol, puis rapporte comme d'habitude (Taux).
type TheftSave struct {
	Lieu    string
	Reste   float64
	Duree   float64
	Taux    float64 `json:",omitempty"` // ce que le lieu rapporte par seconde (retiré de Save.Rate)
	Vole    float64 `json:",omitempty"`
	Caisses []int   `json:",omitempty"` // les caisses déjà vidées
	Repere  bool    `json:",omitempty"` // repérés par la vidéosurveillance
	Voleurs int     `json:",omitempty"` // combien ils sont (0 : 5)
	Feu     bool    `json:",omitempty"` // un incendie
}

func (g *Game) theftSave() *TheftSave {
	th := g.Theft
	if th == nil {
		return nil
	}
	ts := &TheftSave{Lieu: th.Place, Reste: th.Left, Duree: th.Total, Vole: th.Stolen, Taux: g.placeRate(th.Place), Repere: th.Spotted, Voleurs: len(th.Robbers), Feu: th.Fire}
	for k, done := range th.robbed {
		if done {
			ts.Caisses = append(ts.Caisses, k)
		}
	}
	return ts
}

// placeRate : ce que ce lieu a rapporté par seconde ces 15 dernières minutes (comme RateOffline).
func (g *Game) placeRate(id string) float64 {
	a := 0.0
	for _, x := range g.income {
		if x.place == id && x.t >= g.T-900 {
			a += x.a
		}
	}
	return a / math.Min(900, math.Max(60, g.T))
}

func (g *Game) theftRestore(ts *TheftSave) {
	if ts == nil {
		return
	}
	if ts.Feu { // un incendie : fini pendant l'absence, le lieu a brûlé ; sinon, il brûle encore
		p := g.Place(ts.Lieu)
		if p == nil || !p.Owned {
			return
		}
		if ts.Reste <= 0 {
			g.burn(p)
			return
		}
		g.Order = nil
		g.Theft = &Theft{Place: ts.Lieu, Left: math.Min(ts.Reste, ts.Duree), Total: ts.Duree, Fire: true, robbed: map[int]bool{}}
		for _, w := range p.Workers {
			w.Hidden = true
		}
		g.emit(Event{Kind: EvLevel, Place: p.ID, Pos: p.Gate, Text: tr.F("Le feu brûle toujours {alieu} : encore {duree}.", "alieu", placeAt(p.ID), "lieu", theName(p.ID), "duree", durText(ts.Reste))})
		return
	}
	if ts.Reste <= 0 { // (ils sont partis pendant l'absence : arrêtés en partant, s'ils étaient repérés)
		if ts.Repere {
			n := ts.Voleurs
			if n == 0 {
				n = robbersN
			}
			g.arrest(ts.Lieu, n, CustodyTime) // (la garde à vue commence au retour : rien ne tourne jeu fermé)
		}
		return
	}
	p := g.Place(ts.Lieu)
	if p == nil || !p.Owned {
		return
	}
	g.Order = nil
	nr := ts.Voleurs
	if nr == 0 {
		nr = robbersN
	}
	g.startTheftQuiet(ts.Lieu, ts.Duree, nr)
	th := g.Theft
	th.Spotted = ts.Repere
	th.Left, th.Stolen = math.Min(ts.Reste, ts.Duree), ts.Vole
	for _, k := range ts.Caisses {
		th.robbed[k] = true
	}
	for _, w := range p.Workers {
		w.Hidden = true // (ils étaient déjà au vestiaire)
	}
	g.emit(Event{Kind: EvLevel, Place: p.ID, Pos: p.Gate, Text: tr.F("Les voleurs sont toujours là : {lieu} reste arrêté encore {duree}.", "lieu", theName(p.ID), "duree", durText(th.Left))})
}
