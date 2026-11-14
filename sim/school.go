package sim

import (
	"math"
	"math/rand"
)

// L'école de Clairbourg. Les élèves arrivent au portail, on les inscrit à
// l'accueil, ils s'assoient dans une classe de primaire. Une classe avance si
// elle a un prof, ou si le joueur s'y tient. Formation finie : l'élève monte au
// niveau supérieur (table libre + prof), sinon il quitte l'école et rapporte
// un point du type de sa classe.

const schoolPrice = 5000

// Niveaux de l'école.
const (
	Primaire = iota
	College
	Lycee
	Universite
)

// PointTypes : un type de point par niveau.
var PointTypes = []string{"ouvrier", "marketing", "comptable", "formation"}

// Supplies : le matériel de chaque niveau.
var Supplies = [][]Item{{Crayons, Cahiers}, {Geometrie, Atlas}, {Calculatrice, Microscope}, {Ordinateur, Blouse}}

var levelNames = []string{"Primaire", "Collège", "Lycée", "Université"}

// LevelName : nom d'un niveau.
func LevelName(l int) string { return levelNames[l] }

// formation : durée d'une formation (secondes), sans points de formation.
var formation = []float64{600, 1200, 2400, 4800}

// Room : une salle de classe.
type Room struct {
	ID     string
	Name   string
	Level  int
	Body   Rect
	built  bool
	Prof   *Agent
	seats  [6]*Student // élève assis (ou attendu) à chaque table
	Row    int         // 0 : rangée du haut (porte en bas) ; 1 : rangée du bas (porte en haut) ; 2 : amphithéâtre
	Want   Item        // matériel demandé (-1 : rien)
	wantT  float64     // avant la prochaine demande
	BoostT float64     // vitesse double pendant encore… (secondes)
	// formation express : un intervenant vient former toute la classe en 10 s
	Expert  *Agent  // l'intervenant (nil : personne)
	ExpertT float64 // secondes de cours qui restent (une fois arrivé)
	expPaid float64 // déjà payé
	expOut  bool    // il a fini : il repart
	expWait bool    // payée : il faut quitter la case avant d'en racheter une
}

// expressPrice : prix de la formation express, par élève (formation complète).
var expressPrice = []float64{200, 800, 3000, 12000}

// ExpressTime : durée du cours de l'intervenant.
const ExpressTime = 10.0

// ExpressZone : la case « Formation express » de la salle.
func (r *Room) ExpressZone() Rect {
	if r.Row == 2 {
		return Rect{r.Body.X + r.Body.W/2 - 100, r.Body.Y + r.Body.H - 110, 200, 92}
	}
	return Rect{r.Body.X + 20, r.Body.Y + 12, 130, 80} // (en haut à gauche, à côté du tableau)
}

// ExpressCost : prix pour former tout de suite les élèves assis (selon ce qu'il leur reste).
func (p *Place) ExpressCost(r *Room) float64 {
	c := 0.0
	for _, s := range p.Students {
		if s.Room == r && s.State == 2 {
			c += (1 - math.Min(1, s.Prog)) * expressPrice[r.Level]
		}
	}
	if c == 0 {
		return 0
	}
	return math.Max(50, math.Round(c/10)*10)
}

// ExpressOK : la case est-elle proposée ? (des élèves assis, pas d'intervenant en route)
func (p *Place) ExpressOK(r *Room) bool { return r.built && r.Expert == nil && p.ExpressCost(r) > 0 }

// stepExpress : l'intervenant arrive par le portail, fait cours 10 s, et repart.
func (p *Place) stepExpress(r *Room, dt float64) {
	a := r.Expert
	if a == nil {
		return
	}
	switch {
	case r.expOut:
		if p.walk(a, V{p.Gate.X, roadY + 40}, 260, dt) {
			r.Expert, r.expOut = nil, false
		}
	case r.ExpertT > 0:
		a.Moving = false
		r.ExpertT -= dt
		if r.ExpertT <= 0 {
			// toute la classe est formée
			for _, s := range p.Students {
				if s.Room == r && s.State == 2 {
					s.Prog = 1
				}
			}
			r.expOut = true
			p.g.emit(Event{Kind: EvLevel, Pos: r.Body.C(), Text: "Formation express terminée : « " + r.Name + " » est formée !"})
		}
	default:
		to := r.ProfPos().Add(V{40, 60})
		if r.Row == 2 {
			to = r.ProfPos().Add(V{90, 0})
		}
		if p.walk(a, to, 260, dt) {
			r.ExpertT = ExpressTime
			a.Face = V{1, 0}
			if r.Row == 2 {
				a.Face = V{0, 1}
			}
		}
	}
}

// payExpress : le joueur, arrêté sur la case, paie la formation express.
func (p *Place) payExpress(a *Agent) bool {
	g := p.g
	for _, r := range p.Rooms {
		if r.expWait && !r.ExpressZone().Has(a.Pos) {
			r.expWait = false
		}
		if r.expWait || !p.ExpressOK(r) || !r.ExpressZone().Has(a.Pos) || g.Money < 1 {
			continue
		}
		cost := p.ExpressCost(r)
		step := math.Max(1, math.Ceil(cost/30))
		pay := math.Min(step, math.Min(g.Money, cost-r.expPaid))
		if pay > 0 {
			g.Money -= pay
			r.expPaid += pay
		}
		g.emit(Event{Kind: EvPay, Pos: r.ExpressZone().C()})
		if r.expPaid >= cost-1e-6 {
			r.expPaid, r.expWait = 0, true
			r.Expert = &Agent{Kind: WorkerAgent, Pos: V{p.Gate.X, roadY + 40}, Face: V{0, -1}, Look: 4242 + int(r.Body.X)}
			r.Expert.lastPos = r.Expert.Pos
			g.emit(Event{Kind: EvHire, Pos: r.ExpressZone().C()})
			g.emit(Event{Kind: EvLevel, Pos: r.ExpressZone().C(), Text: "Formation express : un intervenant arrive pour « " + r.Name + " » !"})
		}
		return true
	}
	return false
}

// Student : un élève.
type Student struct {
	A     *Agent
	State int // 0 : file d'attente ; 1 : va à sa table ; 2 : assis ; 3 : s'en va
	Room  *Room
	Seat  int
	Prog  float64 // avancement de la formation (0 à 1)
}

// RoomAt : salle qui contient ce point.
func (p *Place) RoomAt(v V) *Room {
	for _, r := range p.Rooms {
		if r.built && r.Body.Has(v) {
			return r
		}
	}
	return nil
}

// Tables : nombre de tables d'une salle (1 à 6).
func (p *Place) Tables(r *Room) int { return 1 + p.g.Levels[p.key("tables_"+r.ID)] }

// HasVideo : la salle a-t-elle un vidéoprojecteur ?
func (p *Place) HasVideo(r *Room) bool { return p.g.Levels[p.key("video_"+r.ID)] > 0 }

// HasProf : la salle a-t-elle un prof ?
func (p *Place) HasProf(r *Room) bool { return p.g.Levels[p.key("prof_"+r.ID)] > 0 }

// Built : la salle est-elle ouverte ?
func (r *Room) Built() bool { return r.built }

// Seat : position de la table n.
func (r *Room) SeatPos(n int) V {
	if r.Row == 2 {
		// amphithéâtre : deux arcs face au prof
		a := (float64(n%3) - 1) * 0.55
		rad := 260.0 + float64(n/3)*170
		c := V{r.Body.X + r.Body.W/2, r.Body.Y + 110}
		return V{c.X + math.Sin(a)*rad, c.Y + math.Cos(a)*rad*0.9}
	}
	return V{r.Body.X + 200 + float64(n%3)*110, r.Body.Y + 150 + float64(n/3)*120}
}

// ProfPos : place du prof dans la salle.
func (r *Room) ProfPos() V {
	if r.Row == 2 {
		return V{r.Body.X + r.Body.W/2, r.Body.Y + 120}
	}
	return V{r.Body.X + 75, r.Body.Y + 160}
}

// Géométrie de l'école (voir le plan) : la cour des petits à gauche, la cour
// de récréation en haut, deux rangées de classes de part et d'autre d'un
// couloir, l'amphithéâtre à droite ; l'accueil et les fournitures en bas.
const (
	SchoolX0    = 760.0  // première colonne de salles
	SchoolColW  = 540.0  // largeur d'une colonne (salle + mur)
	SchoolRow1Y = 540.0  // rangée du haut
	SchoolRow2Y = 1040.0 // rangée du bas
	SchoolRoomH = 380.0
	SchoolW     = 5940.0
)

// SchoolCol : bord gauche de la colonne i.
func SchoolCol(i int) float64 { return SchoolX0 + float64(i)*SchoolColW }

// SchoolAmphi : l'amphithéâtre (université), à droite, sur les deux rangées.
func SchoolAmphi() Rect {
	return Rect{SchoolCol(8), SchoolRow1Y, 760, SchoolRow2Y + SchoolRoomH - SchoolRow1Y}
}

// SchoolFourn : le local des fournitures (sous C2 à L2).
// (plus haut que les classes : le nom des étagères ne touche plus le mur)
func SchoolFourn() Rect { return Rect{SchoolCol(4), SchoolRow2Y, 4*SchoolColW - 40, SchoolRoomH + 110} }

// fournDoors : les trois portes du local, une entre deux colonnes d'étagères (x relatif, largeur 120).
var fournDoors = []float64{335, 855, 1375}

// FournWalls : les murs du local des fournitures (le haut est coupé par les portes).
func FournWalls() []Rect {
	fr := SchoolFourn()
	ws := []Rect{{fr.X - 6, fr.Y - 6, 10, fr.H + 12}, {fr.X + fr.W - 4, fr.Y - 6, 10, fr.H + 12}, {fr.X - 6, fr.Y + fr.H - 4, fr.W + 12, 10}}
	x := fr.X - 6
	for _, d := range fournDoors {
		ws = append(ws, Rect{x, fr.Y - 6, fr.X + d - x, 10})
		x = fr.X + d + 120
	}
	return append(ws, Rect{x, fr.Y - 6, fr.X + fr.W + 6 - x, 10})
}

// SchoolAccueil : l'accueil (entre P4 et les fournitures).
func SchoolAccueil() Rect { return Rect{SchoolCol(3), SchoolRow2Y, 500, SchoolRoomH} }

// SeatStudent : élève assis à la table n (ou nil).
func (r *Room) SeatStudent(n int) *Student { return r.seats[n] }

// Teaching : la classe avance-t-elle en ce moment ?
func (p *Place) Teaching(r *Room) bool {
	if p.HasProf(r) {
		return true
	}
	pl := p.g.Player
	return p.g.Cur == p && pl != nil && r.Body.Has(pl.Pos)
}

// FormationMult : vitesse de formation (points de formation).
func (g *Game) FormationMult() float64 { return 1 + float64(g.Levels["ecole.pts_formation"])/100 }

// FormationTime : durée d'une formation à ce niveau, en secondes.
func (g *Game) FormationTime(level int) float64 {
	return formation[level] / (g.FormationMult() * g.SupplyBonus(level) * g.boostF("ecole", "")) // (cadeau de la cantine)
}

// SupplyBonus : chaque étagère de fournitures d'un niveau fait travailler ses classes 15 % plus vite.
func (g *Game) SupplyBonus(level int) float64 {
	p := g.Place("ecole")
	if p == nil || level >= len(Supplies) {
		return 1
	}
	n := 0
	for _, it := range Supplies[level] {
		if sh := p.supplyShelf(it); sh != nil && sh.Built {
			n++
		}
	}
	return 1 + 0.15*float64(n)
}

func newSchool(g *Game) *Place {
	p := &Place{ID: "ecole", Name: "L'école", W: SchoolW, H: 1800, g: g, rng: rand.New(rand.NewSource(31)), Rep: 3, Price: schoolPrice, SellMult: 1}
	acc := SchoolAccueil()
	gx := acc.C().X
	p.SpawnL = V{350, roadY + 40}
	p.SpawnR = V{SchoolW - 300, roadY + 40}
	p.Gate = V{gx, roadY + 40}
	p.Inside = V{gx, 1480}
	p.RestSpot = V{gx, 980}
	p.Start = V{gx, 1480}

	// salles : P1 P2 P3 C1 C2 C3 L1 L2 en haut ; P6 P5 P4 en bas ; l'amphithéâtre à droite
	add := func(level, n, row int, b Rect) {
		r := &Room{ID: itoa(level) + "_" + itoa(n), Level: level, Body: b, Want: -1, wantT: 20, Row: row}
		r.Name = levelNames[level]
		if level != Universite {
			r.Name += " " + itoa(n+1)
		}
		p.Rooms = append(p.Rooms, r)
	}
	top := func(i int) Rect { return Rect{SchoolCol(i), SchoolRow1Y, 500, SchoolRoomH} }
	bot := func(i int) Rect { return Rect{SchoolCol(i), SchoolRow2Y, 500, SchoolRoomH} }
	add(Primaire, 0, 0, top(0))
	add(Primaire, 1, 0, top(1))
	add(Primaire, 2, 0, top(2))
	add(Primaire, 3, 1, bot(2))
	add(Primaire, 4, 1, bot(1))
	add(Primaire, 5, 1, bot(0))
	for i := 0; i < 3; i++ {
		add(College, i, 0, top(3+i))
	}
	for i := 0; i < 2; i++ {
		add(Lycee, i, 0, top(6+i))
	}
	add(Universite, 0, 2, SchoolAmphi())
	p.Rooms[0].built = true
	// murs (on ne les traverse pas) : porte en bas (rangée du haut), en haut à
	// droite (rangée du bas), à gauche sur le couloir (amphithéâtre)
	wall := func(id, unl string, w Rect) {
		st := &Station{ID: id, Name: "Mur", Kind: Trash, Unlock: unl, Built: unl == ""}
		st.Body = w
		st.Zone = Rect{-1000, -1000, 0, 0}
		p.Stations = append(p.Stations, st)
	}
	for _, r := range p.Rooms {
		b := r.Body
		unl := "salle_" + r.ID
		if r == p.Rooms[0] {
			unl = ""
		}
		var ws []Rect
		switch r.Row {
		case 0:
			ws = []Rect{{b.X - 6, b.Y - 6, b.W + 12, 10}, {b.X - 6, b.Y - 6, 10, b.H + 12}, {b.X + b.W - 4, b.Y - 6, 10, b.H + 12},
				{b.X - 6, b.Y + b.H - 4, b.W/2 - 50, 10}, {b.X + b.W/2 + 56, b.Y + b.H - 4, b.W/2 - 50, 10}}
		case 1:
			ws = []Rect{{b.X - 6, b.Y + b.H - 4, b.W + 12, 10}, {b.X - 6, b.Y - 6, 10, b.H + 12}, {b.X + b.W - 4, b.Y - 6, 10, b.H + 12},
				{b.X - 6, b.Y - 6, b.W - 100, 10}, {b.X + b.W - 10, b.Y - 6, 16, 10}}
		case 2:
			cy := (SchoolRow1Y + SchoolRoomH + SchoolRow2Y) / 2
			ws = []Rect{{b.X - 6, b.Y - 6, b.W + 12, 10}, {b.X - 6, b.Y + b.H - 4, b.W + 12, 10}, {b.X + b.W - 4, b.Y - 6, 10, b.H + 12},
				{b.X - 6, b.Y - 6, 10, cy - 50 - b.Y + 6}, {b.X - 6, cy + 56, 10, b.Y + b.H + 6 - cy - 56}}
		}
		for k, w := range ws {
			wall("mur_"+r.ID+"_"+itoa(k), unl, w)
		}
	}
	// (l'accueil n'a pas de murs : on y circule librement)
	// local des fournitures : trois portes sur le couloir (une entre deux colonnes d'étagères)
	fr := SchoolFourn()
	for k, w := range FournWalls() {
		wall("mur_fourn_"+itoa(k), "", w)
	}

	// accueil (le bureau fait office de caisse : on s'y tient pour inscrire)
	reg := &Station{ID: "accueil", Name: "Accueil", Kind: Register, HireKey: "secretaire", Built: true}
	reg.Body = Rect{gx - 110, acc.Y + 120, 220, 66}
	reg.Spot = V{gx, acc.Y + 80}
	reg.Zone = Rect{gx - 85, acc.Y + 20, 170, 96}
	reg.QueueHead = V{gx, acc.Y + 240}
	p.Stations = append(p.Stations, reg)

	// local de fournitures : une étagère par matériel, qui se remplit toute seule
	supplyTime := []float64{12, 18, 25, 35}
	supplyCost := [][]float64{{0, 800}, {5000, 8000}, {20000, 30000}, {60000, 80000}}
	for l, its := range Supplies {
		for k, it := range its {
			i := l*2 + k
			x := fr.X + 70 + float64(i%4)*520 + float64(i%4/2)*0
			y := fr.Y + 75 + float64(i/4)*235
			st := &Station{ID: "fourn_" + itoa(int(it)), Name: Items[it].Name, Kind: Animal, Item: it, Time: supplyTime[l], PileMax: 6,
				BaseCost: math.Max(600, supplyCost[l][k])}
			st.Body = Rect{x, y, 130, 60}
			st.Zone = Rect{x, y + 72, 130, 84}
			if supplyCost[l][k] > 0 {
				st.Unlock = "fourn_" + itoa(int(it))
			} else {
				st.Built = true
			}
			p.Stations = append(p.Stations, st)
		}
	}
	// achat des salles
	roomCost := map[int][]float64{
		Primaire:   {0, 1500, 4000, 9000, 20000, 45000},
		College:    {6000, 15000, 35000},
		Lycee:      {30000, 90000},
		Universite: {200000},
	}
	idx := map[int]int{}
	var prev = map[int]*Room{}
	for _, r := range p.Rooms {
		r := r
		n := idx[r.Level]
		idx[r.Level]++
		cost := roomCost[r.Level][n]
		if cost == 0 {
			prev[r.Level] = r
			continue
		}
		var req func() bool
		if before := prev[r.Level]; before != nil {
			req = func() bool { return before.built }
		} else {
			// la première salle d'un niveau : il faut 2 salles du niveau d'avant
			lv := r.Level - 1
			req = func() bool {
				k := 0
				for _, o := range p.Rooms {
					if o.Level == lv && o.built {
						k++
					}
				}
				return k >= 2 || (lv == Lycee && k >= 1)
			}
		}
		prev[r.Level] = r
		c := r.Body.C()
		p.Unlocks = append(p.Unlocks, &Unlock{ID: "salle_" + r.ID, Name: r.Name, Cost: cost, Zone: Rect{c.X - 90, c.Y - 65, 180, 130},
			Requires: req, Locked: "Ouvre d'abord d'autres classes", Hint: "Une salle de plus : 1 table, jusqu'à 6."})
	}

	// déblocage des étagères (avec la première salle de leur niveau)
	for l, its := range Supplies {
		for k, it := range its {
			l := l
			c := supplyCost[l][k]
			if c == 0 {
				continue
			}
			st := p.Station("fourn_" + itoa(int(it)))
			cc := st.Body.C()
			p.Unlocks = append(p.Unlocks, &Unlock{ID: st.Unlock, Name: Items[it].Name, Cost: c, Zone: Rect{cc.X - 90, cc.Y - 40, 180, 130},
				Requires: func() bool {
					for _, r := range p.Rooms {
						if r.Level == l && r.built {
							return true
						}
					}
					return false
				}, Locked: "Avec la 1re classe : " + levelNames[l], Hint: "+15 % de vitesse pour les classes de ce niveau, et de nouvelles demandes de matériel."})
		}
	}

	// améliorations au sol : tables et prof de chaque salle, secrétaire
	k := p.key
	tableBase := []float64{60, 1000, 5000, 20000} // (les premières tables : pas chères, pour bien démarrer)
	profBase := []float64{800, 6000, 25000, 80000}
	nth := map[int]int{}
	p.ups = []*Upgrade{
		{Key: k("secretaire"), Name: "Secrétaire", Max: 1, Costs: []float64{1200}, Desc: func(int) string { return "Elle inscrit les élèves à ta place." }},
		{Key: k("ouvriers"), Name: "Surveillant", Max: 2, Now: true, Costs: []float64{3000, 15000},
			Desc: func(l int) string {
				return "Il apporte aux classes le matériel qu'elles demandent (moins vite que toi). Embauchés : " + itoa(l) + "."
			}},
	}
	spot := func(id, label string, z Rect, done func(int) string) *Spot {
		sp := &Spot{ID: id, Label: label, Zone: z, Done: done}
		p.Spots = append(p.Spots, sp)
		return sp
	}
	spot("secretaire", "Secrétaire", Rect{acc.X + acc.W - 140, acc.Y + 30, 120, 110}, func(int) string { return "Secrétaire embauchée !" })
	spot("ouvriers", "Surveillant", Rect{}, func(l int) string { return "Surveillant +1 !" })
	for _, r := range p.Rooms {
		r := r
		n := nth[r.Level]
		nth[r.Level]++
		tb := tableBase[r.Level] * (1 + 0.5*float64(n))
		var tc []float64
		for j := 0; j < 5; j++ {
			tc = append(tc, math.Round(tb*math.Pow(1.5, float64(j))/10)*10)
		}
		p.ups = append(p.ups,
			&Upgrade{Key: k("tables_" + r.ID), Name: r.Name + " : table", Max: 5, Now: true, Costs: tc,
				Desc: func(l int) string { return "Une table de plus dans la classe. Tables : " + itoa(l+1) + "/6." }},
			&Upgrade{Key: k("prof_" + r.ID), Name: r.Name + " : prof", Max: 1, Costs: []float64{math.Round(profBase[r.Level] * (1 + 0.5*float64(n)))},
				Desc: func(int) string { return "Il fait cours à ta place : la classe avance même quand tu n'es pas là." }},
		)
		sp := spot("tables_"+r.ID, "Table", Rect{}, func(l int) string { return r.Name + " : " + itoa(l+1) + " tables !" })
		sp.Room = r
		pz, vz := Rect{r.Body.X + 20, r.Body.Y + 100, 110, 100}, Rect{r.Body.X + 20, r.Body.Y + 240, 110, 100}
		if r.Row == 2 {
			pz, vz = Rect{r.Body.X + 30, r.Body.Y + 40, 110, 100}, Rect{r.Body.X + r.Body.W - 140, r.Body.Y + 40, 110, 100}
		}
		sp2 := spot("prof_"+r.ID, "Prof", pz, func(int) string { return "Un prof pour « " + r.Name + " » !" })
		sp2.Room = r
		vc := math.Round(profBase[r.Level]*1.5*(1+0.5*float64(n))/10) * 10
		p.ups = append(p.ups, &Upgrade{Key: k("video_" + r.ID), Name: r.Name + " : vidéoprojecteur", Max: 1, Costs: []float64{vc},
			Desc: func(int) string { return "Des cours en images : la formation va 50 % plus vite dans cette salle." }})
		sp3 := spot("video_"+r.ID, "Vidéoprojecteur", vz, func(int) string { return "Vidéoprojecteur installé dans « " + r.Name + " » !" })
		sp3.Room = r
	}
	p.setupYard(-620)
	return p
}

// schoolSpots : la case « Table » se place là où ira la prochaine table.
func (p *Place) schoolSpots() {
	for _, sp := range p.Spots {
		if sp.Room != nil && sp.ID == "tables_"+sp.Room.ID {
			n := p.Tables(sp.Room)
			if n < 6 {
				c := sp.Room.SeatPos(n)
				sp.Zone = Rect{c.X - 52, c.Y - 46, 104, 92}
			}
		}
	}
}

func (p *Place) stepSchool(dt float64, active bool) {
	g := p.g
	for _, u := range p.Unlocks {
		if u.Done {
			for _, r := range p.Rooms {
				if "salle_"+r.ID == u.ID {
					r.built = true
				}
			}
		}
	}
	p.schoolSpots()
	p.syncStaff()
	// profs
	for _, r := range p.Rooms {
		if r.built && p.HasProf(r) && r.Prof == nil {
			face := V{1, 0}
			if r.Row == 2 {
				face = V{0, 1}
			}
			r.Prof = &Agent{Kind: WorkerAgent, Pos: r.ProfPos(), Face: face, Look: 700 + len(r.ID)*13 + r.Level*31 + int(r.Body.X)}
			g.emit(Event{Kind: EvHire, Pos: r.Prof.Pos})
		}
	}
	for _, r := range p.Rooms {
		p.stepExpress(r, dt)
	}
	if active && g.Player != nil {
		if pl := g.Player; !pl.Moving {
			p.expTick -= dt
			if p.expTick <= 0 && p.payExpress(pl) {
				p.expTick = 0.05
			}
		}
		p.interact(g.Player, dt)
	}
	p.stepSupplies(dt, active)

	// arrivée des élèves : ils font la queue à l'accueil
	reg := p.Station("accueil")
	waiting := 0
	for _, s := range p.Students {
		if s.State == 0 {
			waiting++
		}
	}
	p.spawnT -= dt
	if p.spawnT <= 0 {
		// des places libres en primaire : les élèves arrivent bien plus vite
		// (sinon l'école, plus rapide, attendait ses élèves)
		p.spawnT = 8 + 4*p.rand()
		if p.freeSeats(Primaire) > waiting {
			p.spawnT = 2 + 2*p.rand()
		}
		if waiting < 6 {
			from := p.SpawnL
			if p.rand() < 0.5 {
				from = p.SpawnR
			}
			a := &Agent{Kind: StudentAgent, Pos: from, Face: V{0, -1}, Look: 1000 + p.rng.Intn(100000)}
			a.lastPos = a.Pos
			p.Students = append(p.Students, &Student{A: a})
		}
	}
	// inscription de l'élève en tête de file
	var head *Student
	q := 0
	for _, s := range p.Students {
		if s.State != 0 {
			continue
		}
		spot := reg.QueueHead.Add(V{0, float64(q) * 60})
		if q == 0 {
			head = s
		}
		q++
		p.walk(s.A, spot, 200, dt)
	}
	if head != nil && head.A.Pos.Dist(reg.QueueHead) < 12 {
		pl := g.Player
		by := 0.0
		if active && pl != nil && reg.Zone.Has(pl.Pos) {
			by = 1.2
		} else if reg.Cashier != nil {
			by = 2.5
		}
		if by > 0 {
			if r, n := p.freeSeat(Primaire, false); r != nil {
				p.regT += dt
				if p.regT >= by {
					p.regT = 0
					r.seats[n] = head
					head.Room, head.Seat, head.State = r, n, 1
					g.emit(Event{Kind: EvPay, Pos: reg.Body.C()})
				}
			}
		}
	}

	// les classes
	alive := p.Students[:0]
	for _, s := range p.Students {
		switch s.State {
		case 1:
			if p.walk(s.A, s.Room.SeatPos(s.Seat).Add(V{0, 40}), 200, dt) {
				s.State = 2
				s.A.Face = V{0, -1}
			}
		case 2:
			r := s.Room
			if !p.Teaching(r) && s.Prog < 1 { // (formation express : Prog = 1 même sans prof)
				break
			}
			sp := 1.0
			if r.BoostT > 0 {
				sp = 2 // le matériel demandé est arrivé
			}
			sp *= p.ClassBoost(r) // l'ensemble du diplômé, si le joueur est dans la classe
			if p.HasVideo(r) {
				sp *= 1.5 // vidéoprojecteur
			}
			s.Prog += sp * dt / g.FormationTime(r.Level)
			if s.Prog < 1 {
				break
			}
			// formation terminée
			r.seats[s.Seat] = nil
			if p.rand() < g.FailChance() {
				// échec : il s'en va ; il rapporte quand même le point du niveau
				// qu'il a validé avant (rien s'il était en primaire)
				if r.Level > 0 {
					p.graduate(s.A.Pos, r.Level-1)
				}
				g.Levels["ecole.echecs"]++
				g.emit(Event{Kind: EvAngry, Pos: s.A.Pos})
				s.State = 3
				break
			}
			if nr, n := p.freeSeat(r.Level+1, true); nr != nil {
				nr.seats[n] = s
				s.Room, s.Seat, s.Prog, s.State = nr, n, 0, 1
			} else {
				p.graduate(s.A.Pos, r.Level)
				s.State = 3
			}
		case 3:
			if s.A.Pos.Y > roadY-20 {
				if p.walk(s.A, V{p.SpawnR.X + 200, roadY + 40}, 220, dt) {
					continue
				}
			} else {
				p.walk(s.A, p.Gate, 220, dt)
			}
		}
		alive = append(alive, s)
	}
	p.Students = alive
}

// freeSeat : une table libre dans une salle de ce niveau (avec un prof si demandé).
func (p *Place) freeSeat(level int, needProf bool) (*Room, int) {
	if level > Universite {
		return nil, 0
	}
	for _, r := range p.Rooms {
		if !r.built || r.Level != level || (needProf && !p.HasProf(r)) {
			continue
		}
		for n := 0; n < p.Tables(r); n++ {
			if r.seats[n] == nil {
				return r, n
			}
		}
	}
	return nil, 0
}

// freeSeats : tables libres d'un niveau (classes construites).
func (p *Place) freeSeats(level int) int {
	n := 0
	for _, r := range p.Rooms {
		if !r.built || r.Level != level {
			continue
		}
		for k := 0; k < p.Tables(r); k++ {
			if r.seats[k] == nil {
				n++
			}
		}
	}
	return n
}

// graduate : un élève quitte l'école et rapporte un point.
func (p *Place) graduate(at V, level int) {
	g := p.g
	t := PointTypes[level]
	g.Levels["pts."+t]++
	g.Levels["ecole.diplomes"]++
	g.emit(Event{Kind: EvPoint, Pos: at, Text: t})
}

// ---------------------------------------------------------------- points

// PointKind : où l'on peut dépenser un type de point.
type PointKind struct {
	Key   string // suffixe de la clé dans Levels (lieu.pts_xxx)
	Type  string // type de point dépensé (ouvrier = primaire, marketing = collège, comptable = lycée, formation = université)
	Group string // famille (Mécaniciens, Chefs…)
	Name  string
	Desc  string
	Val   func(n int) string // effet actuel, en clair
}

// TypeOfLevel : type de point d'un niveau (onglet).
func TypeOfLevel(l int) string { return PointTypes[l] }

func pctVal(n int) string   { return "+" + itoa(n) + " %" }
func per10Val(n int) string { return "+" + itoa(n/10) + " objet(s)" }

// animalIDs : bâtiments à animaux (Soigneurs) ; les autres machines de la ferme sont pour les mécaniciens.
var animalIDs = map[string]bool{"poulailler": true, "mare": true, "etable": true, "chevrerie": true, "bergerie": true, "porcherie": true, "ruche": true}

// cooks : lieux dont les machines cuisinent (Chefs).
func (p *Place) cooks() bool {
	return p.ID == "boulangerie" || p.ID == "cremerie" || p.ID == "glacier" || p.ID == "boucherie" || p.ID == "patisserie"
}

// sells : lieux qui reçoivent des clients.
func (p *Place) sells() bool { return p.ID != "ecole" && p.ID != "elevage" }

// hasStaff : lieux qui ont des employés (marche, charge).
func (p *Place) hasStaff() bool { return p.ID != "ecole" }

// hasOrders : lieux où l'on livre des commandes.
func (p *Place) hasOrders() bool { return p.Station("commande") != nil }

// PointKinds : les usages possibles des points dans un lieu, tous onglets confondus.
func (p *Place) PointKinds() []PointKind {
	var ks []PointKind
	add := func(key, t, grp, name, desc string, val func(int) string) {
		ks = append(ks, PointKind{key, t, grp, name, desc, val})
	}
	// primaire
	if p.ID == "ferme" || p.ID == "exotique" || p.ID == "domaine" {
		add("pts_agro", "ouvrier", "Agronomes", "Champs", "Les champs poussent plus vite (+1 % par point).", pctVal)
	}
	if p.ID == "elevage" {
		add("pts_soin", "ouvrier", "Soigneurs", "Animaux", "Poulailler, mare, prés, porcherie vont plus vite (+3 % par point).", func(n int) string { return pctVal(3 * n) })
		add("pts_soin_rent", "ouvrier", "Soigneurs", "Sélection génétique", "Les animaux mangent moins pour produire autant.", rentVal)
	}
	if p.ID == "ferme" {
		add("pts_soin", "ouvrier", "Soigneurs", "Animaux", "Poulailler, étable, ruches… vont plus vite (+3 % par point).", func(n int) string { return pctVal(3 * n) })
		add("pts_soin_rent", "ouvrier", "Soigneurs", "Sélection génétique", "Les animaux (et les abeilles) mangent moins pour produire autant.", rentVal)
	}
	if p.hasStaff() {
		add("pts_ent_vit", "ouvrier", "Entraînement", "Vitesse de marche", "Les employés marchent plus vite (+1 % par point).", pctVal)
		add("pts_mains", "ouvrier", "Entraînement", "Mains expertes", "Les employés prennent et posent la marchandise plus vite (+1 % par point).", pctVal)
		add("pts_ent_cap", "ouvrier", "Entraînement", "Charge", "Les employés portent plus (+1 objet tous les 5 points).",
			func(n int) string { return "+" + itoa(n/5) + " objet(s)" })
	}
	if p.hasDrones() {
		add("pts_dr_vit", "ouvrier", "Mécaniciens", "Drones : vitesse", "+3 % de vitesse par point.", func(n int) string { return pctVal(3 * n) })
		add("pts_dr_cap", "ouvrier", "Mécaniciens", "Drones : charge", "+1 objet tous les 2 points (1 au départ).",
			func(n int) string { return itoa(DroneCap(n)) + " objet(s)" })
		add("pts_cam_vit", "ouvrier", "Mécaniciens", "Camions : départs", "Les camions partent plus souvent (+1 % par point).", pctVal)
		add("pts_cam_cap", "ouvrier", "Mécaniciens", "Camions : charge", "+2 produits par voyage par point ; quais : +1 place tous les 2 points.",
			func(n int) string { return "+" + itoa(2*n) + " produits" })
	}
	if p.sells() {
		add("pts_caisse", "ouvrier", "Mécaniciens", "Caisses", "Les caisses passent les articles plus vite (+2 % par point).", func(n int) string { return pctVal(2 * n) })
	}
	if p.ID == "super" {
		add("pts_caddie", "ouvrier", "Mécaniciens", "Caddies", "Avec de bons caddies, les clients marchent plus vite (+1 % par point).", pctVal)
	}
	if p.ID == "port" {
		add("pts_bat_vit", "ouvrier", "Mécaniciens", "Bateaux : vitesse", "Les bateaux vont et viennent plus vite (+3 % par point).", func(n int) string { return pctVal(3 * n) })
		add("pts_bat_cap", "ouvrier", "Mécaniciens", "Bateaux : soute", "+1 poisson par voyage tous les 3 points (1 au départ).",
			func(n int) string { return itoa(BoatCap(n)) + " poissons" })
		add("pts_drag_vit", "ouvrier", "Mécaniciens", "Dragues : vitesse", "Les bras des dragues vont plus vite (+3 % par point).", func(n int) string { return pctVal(3 * n) })
		add("pts_drag_cap", "ouvrier", "Mécaniciens", "Dragues : quantité", "+1 coquillage par passage tous les 4 points (1 au départ).",
			func(n int) string { return itoa(DredgeCap(n)) + " par passage" })
	}
	if p.ID == "exotique" {
		add("pts_mac_vit", "ouvrier", "Mécaniciens", "Machines : vitesse", "Raffinerie, moulins, chocolaterie, torréfacteur, séchoir vont plus vite (+3 % par point).", func(n int) string { return pctVal(3 * n) })
		add("pts_mac_rent", "ouvrier", "Mécaniciens", "Machines : rentabilité", "Elles produisent autant en consommant moins.", rentVal)
	}
	if p.ID == "domaine" {
		add("pts_mac_vit", "ouvrier", "Mécaniciens", "Machines : vitesse", "Presses, brasserie, distilleries, pressoir vont plus vite (+3 % par point).", func(n int) string { return pctVal(3 * n) })
		add("pts_mac_rent", "ouvrier", "Mécaniciens", "Machines : rentabilité", "Elles produisent autant en consommant moins.", rentVal)
	}
	if p.ID == "elevage" {
		add("pts_mac_vit", "ouvrier", "Mécaniciens", "Abattoirs : vitesse", "Les abattoirs vont plus vite (+3 % par point ; ils consomment autant).", func(n int) string { return pctVal(3 * n) })
		add("pts_mac_rent", "ouvrier", "Mécaniciens", "Abattoirs : rentabilité", "Ils produisent autant en consommant moins.", rentVal)
	}
	if p.ID == "ferme" {
		add("pts_mac_vit", "ouvrier", "Mécaniciens", "Machines : vitesse", "Moulins… vont plus vite (+3 % par point ; ils consomment autant).", func(n int) string { return pctVal(3 * n) })
		add("pts_mac_rent", "ouvrier", "Mécaniciens", "Machines : rentabilité", "Elles produisent autant en consommant moins.", rentVal)
	}
	if p.cooks() {
		add("pts_chef_vit", "ouvrier", "Chefs", "Machines : vitesse", "Les machines cuisinent plus vite (+3 % par point ; elles consomment autant).", func(n int) string { return pctVal(3 * n) })
		add("pts_chef_rent", "ouvrier", "Chefs", "Machines : rentabilité", "Elles produisent autant en consommant moins.", rentVal)
	}
	// collège
	// (« Clients » et « Patience » ne s'achètent plus : les points déjà placés comptent toujours)
	if p.sells() {
		add("pts_panier", "marketing", "Panier moyen", "Panier", "Chaque client achète plus de produits (+1 % par point : à 100, un produit de plus par article).", pctVal)
		add("pts_fidele", "marketing", "Clients fidèles", "Absences", "Tes gains pendant ton absence augmentent (+1 % par point, moyenne de tes lieux).", pctVal)
	}
	if p.hasOrders() {
		add("pts_reclame", "marketing", "Réclame", "Cadeaux", "Les cadeaux des commandes de ce lieu durent plus longtemps (+1 % par point) et les cadeaux spéciaux viennent plus souvent.", pctVal)
	}
	// lycée
	if p.sells() {
		add("pts_comptable", "comptable", "Comptables", "Gains", "Les ventes rapportent plus (+1 % par point).", pctVal)
	}
	// (« Négociation » ne s'achète plus : inutile une fois tout acheté ; les points déjà placés comptent toujours)
	if p.hasOrders() {
		add("pts_logist", "comptable", "Logistique", "Commandes", "Plus de temps pour livrer les commandes (+1 % par point).", pctVal)
	}
	// université
	if p.ID == "ecole" {
		add("pts_formation", "formation", "Formation", "Vitesse", "Les formations vont plus vite (+1 % par point).", pctVal)
		add("pts_pedagogie", "formation", "Pédagogie", "Réussite", "Moins d'élèves échouent.",
			func(n int) string { return "1 échec sur " + itoa(6+n) })
	}
	return ks
}

// rentVal : part des ingrédients économisés.
func rentVal(n int) string { return "−" + itoa(int(math.Round(100*RentSave(n)))) + " % consommé" }

// RentSave : chance de ne pas consommer un ingrédient (rentabilité), au plus 75 %.
func RentSave(n int) float64 { return math.Min(0.75, float64(n)/(float64(n)+100)) }

// FailChance : chance qu'un élève échoue (1 sur 6, puis 1 sur 7… avec la pédagogie).
func (g *Game) FailChance() float64 { return 1 / (6 + float64(g.Levels["ecole.pts_pedagogie"])) }

// SpendPoints : place n points d'un usage dans un lieu (définitif).
func (g *Game) SpendPoints(p *Place, k PointKind, n int) bool {
	if n <= 0 || g.Levels["pts."+k.Type] < n {
		return false
	}
	g.Levels["pts."+k.Type] -= n
	g.Levels[p.key(k.Key)] += n
	return true
}

// Points : points disponibles d'un type.
func (g *Game) Points(t string) int { return g.Levels["pts."+t] }

// ptsF : bonus en fraction (1 point = 1 %).
func (p *Place) ptsF(key string) float64 { return float64(p.g.Levels[p.key(key)]) / 100 }

// ---------------------------------------------------------------- sauvegarde

// EleveSave : un élève assis (les autres repartent de zéro).
type EleveSave struct {
	Room string
	Seat int
	Prog float64
}

func (p *Place) saveStudents() []EleveSave {
	var out []EleveSave
	for _, s := range p.Students {
		if s.State == 1 || s.State == 2 {
			out = append(out, EleveSave{Room: s.Room.ID, Seat: s.Seat, Prog: s.Prog})
		}
	}
	return out
}

func (p *Place) restoreStudents(es []EleveSave) {
	for _, u := range p.Unlocks {
		for _, r := range p.Rooms {
			if u.Done && "salle_"+r.ID == u.ID {
				r.built = true
			}
		}
	}
	for _, e := range es {
		for _, r := range p.Rooms {
			if r.ID == e.Room && r.built && e.Seat < p.Tables(r) && r.seats[e.Seat] == nil {
				a := &Agent{Kind: StudentAgent, Pos: r.SeatPos(e.Seat).Add(V{0, 40}), Face: V{0, -1}, Look: 1000 + p.rng.Intn(100000)}
				a.lastPos = a.Pos
				s := &Student{A: a, State: 2, Room: r, Seat: e.Seat, Prog: e.Prog}
				r.seats[e.Seat] = s
				p.Students = append(p.Students, s)
			}
		}
	}
}

// offlineSchool : pendant l'absence, les classes avec un prof continuent
// (à moitié moins vite) ; chaque formation finie rapporte un point.
func (p *Place) offlineSchool(secs float64) {
	if !p.Owned {
		return
	}
	for _, r := range p.Rooms {
		if !r.built || !p.HasProf(r) {
			continue
		}
		n := int(float64(p.Tables(r)) * secs * 0.5 / p.g.FormationTime(r.Level))
		if n > 0 {
			p.g.Levels["pts."+PointTypes[r.Level]] += n
			p.g.Levels["ecole.diplomes"] += n
			p.g.Levels["ecole.absence"] += n
		}
	}
}

// schoolOK : tables et prof ne s'achètent que pour une salle ouverte.
func (g *Game) schoolOK(key string) bool {
	p := g.Place("ecole")
	if p == nil {
		return true
	}
	for _, r := range p.Rooms {
		if key == p.key("tables_"+r.ID) || key == p.key("prof_"+r.ID) {
			return r.built
		}
	}
	return true
}

// ---------------------------------------------------------------- fournitures

type survJob struct {
	room  *Room
	shelf *Station
	phase int
}

func (p *Place) supplyShelf(it Item) *Station {
	return p.Station("fourn_" + itoa(int(it)))
}

func (p *Place) stepSupplies(dt float64, active bool) {
	g := p.g
	// les étagères se remplissent
	for _, s := range p.Stations {
		if s.Kind != Animal || !s.Built {
			continue
		}
		s.Working = s.Pile < s.PileMax
		if s.Working {
			s.Timer += dt * lvlFactor(p.Level(s))
			if s.Timer >= s.Time {
				s.Timer = 0
				s.Pile++
			}
		}
	}
	// demandes des classes et vitesse double
	for _, r := range p.Rooms {
		if !r.built {
			continue
		}
		if r.BoostT > 0 {
			r.BoostT = math.Max(0, r.BoostT-dt)
		}
		busy := false
		for n := 0; n < p.Tables(r); n++ {
			if st := r.seats[n]; st != nil && st.State == 2 {
				busy = true
			}
		}
		if r.Want >= 0 || r.BoostT > 0 || !busy {
			continue
		}
		r.wantT -= dt
		if r.wantT > 0 {
			continue
		}
		var cand []Item
		for _, it := range Supplies[r.Level] {
			if sh := p.supplyShelf(it); sh != nil && sh.Built {
				cand = append(cand, it)
			}
		}
		if len(cand) == 0 {
			r.wantT = 15
			continue
		}
		r.Want = cand[p.rng.Intn(len(cand))]
	}
	// le joueur apporte le matériel
	if pl := g.Player; active && pl != nil {
		if r := p.RoomAt(pl.Pos); r != nil && r.Want >= 0 && pl.count(r.Want) > 0 {
			pl.remove(r.Want)
			p.deliver(r)
		}
	}
	// les surveillants
	if p.survJobs == nil {
		p.survJobs = map[*Agent]*survJob{}
	}
	for _, w := range p.Workers {
		j := p.survJobs[w]
		speed := p.agentSpeed(w) * 0.85
		if j == nil {
			for _, r := range p.Rooms {
				if !r.built || r.Want < 0 || p.roomClaimed(r) {
					continue
				}
				if sh := p.supplyShelf(r.Want); sh != nil && sh.Built && sh.Pile > 0 {
					j = &survJob{room: r, shelf: sh}
					p.survJobs[w] = j
					break
				}
			}
		}
		if j == nil {
			w.wanderT -= dt
			if w.wanderT <= 0 || !w.hasT {
				// dans le couloir, devant l'accueil et le local des fournitures (avant : dehors, hors de vue)
				w.target = V{2300 + p.rand()*900, (SchoolRow1Y + SchoolRoomH + SchoolRow2Y) / 2}
				w.hasT, w.wanderT = true, 3+3*p.rand()
			}
			p.walk(w, w.target, speed*0.6, dt)
			continue
		}
		if j.phase == 0 {
			if p.walk(w, j.shelf.Zone.C(), speed, dt) {
				if j.shelf.Pile > 0 && j.room.Want == j.shelf.Item {
					j.shelf.Pile--
					w.Carry = []Item{j.shelf.Item}
					j.phase = 1
				} else {
					delete(p.survJobs, w)
				}
			}
			continue
		}
		if j.room.Want != w.Carry[0] {
			// plus besoin : on le repose
			j.shelf.Pile = min(j.shelf.PileMax, j.shelf.Pile+1)
			w.Carry = nil
			delete(p.survJobs, w)
			continue
		}
		if p.walk(w, j.room.Body.C(), speed, dt) {
			w.Carry = nil
			p.deliver(j.room)
			delete(p.survJobs, w)
		}
	}
}

func (p *Place) roomClaimed(r *Room) bool {
	for _, j := range p.survJobs {
		if j.room == r {
			return true
		}
	}
	return false
}

// deliver : le matériel demandé arrive : vitesse double pendant une minute.
func (p *Place) deliver(r *Room) {
	r.Want = -1
	r.BoostT = 60
	r.wantT = 30 + 30*p.rand()
	p.g.emit(Event{Kind: EvBoost, Pos: V{r.Body.X + 300, r.Body.Y + 50}, Text: r.Name})
}

// wanted : nombre de classes qui demandent ce matériel.
func (p *Place) wanted(it Item) int {
	n := 0
	for _, r := range p.Rooms {
		if r.built && r.Want == it {
			n++
		}
	}
	return n
}

// NeedsWelcome : des élèves attendent à l'accueil, une table est libre et
// personne (pas de secrétaire) ne les inscrit.
func (p *Place) NeedsWelcome() bool {
	if p.ID != "ecole" || p.g.Levels[p.key("secretaire")] > 0 {
		return false
	}
	if r, _ := p.freeSeat(Primaire, false); r == nil {
		return false
	}
	for _, s := range p.Students {
		if s.State == 0 {
			return true
		}
	}
	return false
}
