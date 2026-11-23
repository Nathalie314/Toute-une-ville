package sim

import (
	"math"
	"math/rand"

	"ville/tr"
)

// LA PRISON (rangée du bas de la carte, après le commissariat). Elle s'ouvre, offerte, dès qu'un
// premier groupe de voleurs est mis en garde à vue. Rien n'y fonctionne sans son directeur
// (100 points d'université). Elle ne tourne que pendant qu'on joue (comme le commissariat).
//
//   - Les fourgons (5 au plus) : chacun va chercher un voleur en garde à vue s'il y a une cellule
//     libre, 1 minute pour aller au commissariat, 1 minute pour revenir.
//   - Les cellules (20 : deux quartiers de 10), un prisonnier chacune ; un surveillant pour deux
//     cellules (points du lycée : 10, 20, 40… ; pas de cellule sans son surveillant).
//   - La cantine (tables : 4, 4 et 2 places), la salle de bain (10 cabines), la salle de sport
//     (10 places), les salles des familles (5), la salle de travail (10 postes), la psy et
//     l'assistant social (1 000 points d'université chacun).
//   - Les niveaux (10 au plus, 10 × n points le niveau n) : chaque cellule et chaque cabine
//     (primaire), la cantine et la salle de sport (collège), chaque salle des familles et la
//     salle de travail (lycée). La psy et l'assistant social comptent au niveau 10.
//
// Un prisonnier arrive avec 100 points de récidive et reste 90 minutes. Aux minutes 10, 20… 80,
// il réclame un de ses 8 besoins pas encore remplis : d'abord les plus rares (psy, assistant
// social, famille, travail, sport, bain, repas ; dormir en dernier, sa cellule est toujours là).
// S'il peut le remplir (5 minutes sur place), il perd autant de points que le niveau de la pièce ;
// sinon, il en gagne 10. À sa sortie, un tirage de 1 à 100 : au-dessus de ses points, il est
// réinséré (un voleur de moins dans la ville, pour toujours) ; sinon, il retourne voler.

// Les besoins, du plus rare au plus facile (l'ordre de préférence du prisonnier).
const (
	NeedPsy = iota
	NeedAS
	NeedFamille
	NeedTravail
	NeedSport
	NeedBain
	NeedRepas
	NeedDormir
	NeedN
)

// NeedNames : le nom de chaque besoin.
var NeedNames = []string{"Voir la psy", "Voir l'assistant social", "Voir sa famille", "Travailler", "Faire du sport", "Se laver", "Manger", "Dormir"}

const (
	PrisCellN     = 20
	PrisVanN      = 5
	PrisStay      = 90 * 60.0 // 1 h 30 en prison
	PrisNeedEvery = 10 * 60.0 // un besoin toutes les 10 minutes (8 en tout)
	PrisActTime   = 5 * 60.0  // une activité dure 5 minutes
	PrisVanTrip   = 2 * 60.0  // aller (1 min) et retour (1 min)
	PrisRecid     = 100       // points de récidive à l'arrivée
	PrisFail      = 10        // un besoin qu'il ne peut pas remplir
	PrisLvlMax    = 10
	PrisCabineN   = 10
	PrisAgresN    = 8
	PrisFamilleN  = 5
	PrisPosteN    = 10
	PrisTableN    = 3
	prisSpeed     = 170.0
)

// PrisAgres : les équipements de la salle de sport (et combien de prisonniers chacun accueille).
var PrisAgres = []struct {
	Name string
	Cap  int
}{{"Tapis de course", 1}, {"Vélo", 1}, {"Banc de musculation", 1}, {"Haltères", 1}, {"Sac de frappe", 1}, {"Panier de basket", 2}, {"Table de ping-pong", 2}, {"Corde à sauter", 1}}

// PrisTables : les places de chaque table de la cantine.
var PrisTables = []int{4, 4, 2}

// CellItems : ce qu'on ajoute dans une cellule, niveau après niveau.
var CellItems = []string{"Lit confortable", "Lavabo", "Toilettes", "Armoire", "Miroir", "Tapis", "Table", "Chaise", "Télévision", "Console de jeu"}

// Prison : ce qui vit à la prison.
type Prison struct {
	Pris []*Prisoner
	Vans [PrisVanN]Van
}

// Van : un fourgon pénitentiaire (T : le temps restant de son voyage ; 0 : au garage).
type Van struct {
	T     float64
	Carry bool // il ramène un voleur
	Cell  int  // la cellule qui l'attend
}

// Prisoner : un détenu.
type Prisoner struct {
	Pos, Face V
	WalkT     float64
	Moving    bool
	Look      int
	Cell      int
	Recid     int
	T         float64 // temps passé en prison
	Done      [NeedN]bool
	Asked     int // besoins déjà réclamés
	Need      int // le besoin en cours (-1 : aucun)
	Slot      int
	State     int
	Left      float64
	path      []V
}

const (
	psCell = iota // dans sa cellule
	psGo          // il va à son activité
	psDo          // il y est
	psBack        // il revient
)

// ---------------------------------------------------------------- géométrie

const (
	PrisCellW    = 170.0
	PrisCellH    = 250.0
	PrisCellGap  = 34.0
	PrisCorH     = 180.0 // le couloir d'un quartier de cellules
	PrisPassW    = 130.0 // le passage d'un quartier vers le grand couloir
	PrisCenterW  = 640.0
	PrisMainH    = 220.0 // le grand couloir
	PrisRoomH    = 380.0
	PrisRoomGap  = 60.0
	PrisBldY     = 100.0
	PrisParkY    = roadY + 240 // sous la route : la zone du joueur, puis le garage des fourgons
	PrisYardX    = 300.0
	PrisGarageX0 = PrisYardX + 2150 + 200
	PrisBayW     = 300.0
	PrisBayGap   = 140.0
	PrisGarageM  = 270.0
	PrisGarageH  = 470.0
)

// les largeurs des pièces du bas, de gauche à droite : cantine, salle de bain, salles des
// familles, salle de sport, salle de travail
var prisRoomW = []float64{760, 620, 5*150 + 4*20, 760, 640}

func prisRowW() float64 {
	w := 0.0
	for _, x := range prisRoomW {
		w += x
	}
	return w + float64(len(prisRoomW)-1)*PrisRoomGap
}

// PrisBldW, PrisBldH : le bâtiment.
var PrisBldW = prisRowW() + 120
var PrisBldH = 60 + 2*PrisCellH + PrisCorH + PrisMainH + PrisRoomH + 60

// PrisW : la largeur de la prison.
var PrisW = math.Max(PrisBldW+600, PrisGarageX0+2*PrisGarageM+PrisVanN*(PrisBayW+PrisBayGap)-PrisBayGap+300)

// PrisBldX : le bord gauche du bâtiment (centré).
var PrisBldX = (PrisW - PrisBldW) / 2

func prisCX() float64   { return PrisBldX + PrisBldW/2 }
func prisTopY() float64 { return PrisBldY + 60 }

// PrisMainY : le haut du grand couloir.
func PrisMainY() float64 { return prisTopY() + 2*PrisCellH + PrisCorH }

func prisZoneW() float64 { return 5*(PrisCellW+PrisCellGap) - PrisCellGap }

// PrisZone : le quartier de cellules z (0 : à gauche, 1 : à droite).
func PrisZone(z int) Rect {
	x := PrisBldX + 60
	if z == 1 {
		x = PrisBldX + PrisBldW - 60 - prisZoneW()
	}
	return Rect{x, prisTopY(), prisZoneW(), 2*PrisCellH + PrisCorH}
}

// PrisZoneCor : le couloir d'un quartier.
func PrisZoneCor(z int) Rect { r := PrisZone(z); return Rect{r.X, r.Y + PrisCellH, r.W, PrisCorH} }

// PrisPass : le passage d'un quartier vers le grand couloir (côté cour).
func PrisPass(z int) Rect {
	r := PrisZone(z)
	x := r.X + r.W + 20
	if z == 1 {
		x = r.X - 20 - PrisPassW
	}
	return Rect{x, r.Y + PrisCellH, PrisPassW, PrisMainY() - r.Y - PrisCellH}
}

// PrisCell : la cellule n° n (1 à 20) : deux par colonne, face à face (le surveillant au milieu).
func PrisCell(n int) Rect {
	z := (n - 1) / 10
	i := (n - 1) % 10
	c, row := i/2, i%2
	if z == 0 {
		c = 4 - c // (le quartier de gauche se remplit depuis la cour, au centre)
	}
	r := PrisZone(z)
	return Rect{r.X + float64(c)*(PrisCellW+PrisCellGap), r.Y + float64(row)*(PrisCellH+PrisCorH), PrisCellW, PrisCellH}
}

// PrisGuardPos : où se tient le surveillant n° k (devant ses deux cellules).
func PrisGuardPos(k int) V { c := PrisCell(2*k - 1); return V{c.C().X, c.Y + PrisCellH + PrisCorH/2} }

// PrisDirector, PrisPsy, PrisAS : la colonne du centre (le directeur en haut).
func PrisDirector() Rect {
	return Rect{prisCX() - PrisCenterW/2, prisTopY(), PrisCenterW, PrisCellH + 60}
}
func PrisPsy() Rect {
	d := PrisDirector()
	return Rect{d.X, d.Y + d.H + 20, PrisCenterW/2 - 10, PrisMainY() - d.Y - d.H - 20}
}
func PrisAS() Rect { p := PrisPsy(); return Rect{p.X + p.W + 20, p.Y, p.W, p.H} }

// PrisRoom : la pièce du bas n° i (0 cantine, 1 salle de bain, 2 familles, 3 sport, 4 travail).
func PrisRoom(i int) Rect {
	x := PrisBldX + 60
	for k := 0; k < i; k++ {
		x += prisRoomW[k] + PrisRoomGap
	}
	return Rect{x, PrisMainY() + PrisMainH, prisRoomW[i], PrisRoomH}
}

// PrisFamily : la salle des familles n° k (1 à 5).
func PrisFamily(k int) Rect {
	r := PrisRoom(2)
	return Rect{r.X + float64(k-1)*170, r.Y, 150, r.H}
}

// PrisCabin : la cabine n° k (1 à 10) de la salle de bain.
func PrisCabin(k int) Rect {
	r := PrisRoom(1)
	c, row := (k-1)%5, (k-1)/5
	return Rect{r.X + 15 + float64(c)*120, r.Y + 30 + float64(row)*180, 110, 150}
}

// PrisAgresAt : l'équipement n° k (1 à 8) de la salle de sport.
func PrisAgresAt(k int) Rect {
	r := PrisRoom(3)
	c, row := (k-1)%4, (k-1)/4
	return Rect{r.X + 20 + float64(c)*185, r.Y + 30 + float64(row)*180, 165, 150}
}

// PrisPoste : le poste de travail n° k (1 à 10).
func PrisPoste(k int) Rect {
	r := PrisRoom(4)
	c, row := (k-1)%5, (k-1)/5
	return Rect{r.X + 15 + float64(c)*124, r.Y + 30 + float64(row)*180, 110, 150}
}

// PrisTable : la table n° k (1 à 3) de la cantine.
func PrisTable(k int) Rect {
	r := PrisRoom(0)
	w := 200.0
	if k == 3 {
		w = 120
	}
	x := r.X + 40 + float64(k-1)*250
	return Rect{x, r.Y + 120, w, 140}
}

// PrisUpZone : la case du grand couloir qui ouvre le menu des améliorations.
func PrisUpZone() Rect { return Rect{prisCX() - 120, PrisMainY() + 35, 240, 150} }

// PrisBay : la place du fourgon n° k au garage (sous la route).
func PrisBay(k int) Rect {
	return Rect{PrisGarageX0 + PrisGarageM + float64(k-1)*(PrisBayW+PrisBayGap), PrisParkY + 60, PrisBayW, 250}
}

// ---------------------------------------------------------------- achats

func prisMoney(base float64, mult float64, n int) float64 {
	return niceRound(base * math.Pow(mult, float64(n-1)))
}

func newPrison(g *Game) *Place {
	W := PrisW
	p := &Place{ID: "prison", Name: "La prison", W: W, H: PrisParkY + PrisGarageH + 130, Top: PrisBldY - 150, g: g,
		rng: rand.New(rand.NewSource(211)), Rep: 3, Price: 0, SellMult: 1, MaxCust: 0}
	p.Pris = &Prison{}
	door := V{prisCX(), PrisBldY + PrisBldH}
	p.Gate = V{door.X, roadY + 40}
	p.SpawnL = V{350, roadY + 40}
	p.SpawnR = V{W - 350, roadY + 40}
	p.Inside = V{door.X, PrisMainY() + PrisMainH/2}
	p.RestSpot = p.Inside
	k := p.key
	up := func(id, name string, cost float64, pts string, desc string) {
		p.ups = append(p.ups, &Upgrade{Key: k(id), Name: name, Max: 1, Costs: []float64{cost}, Pts: pts,
			Desc: func(int) string { return desc }})
	}
	spot := func(id, label string, z Rect, done string) {
		p.Spots = append(p.Spots, &Spot{ID: id, Label: label, Zone: Rect{z.C().X - 70, z.C().Y - 65, 140, 130}, Done: func(int) string { return done }})
	}
	d := PrisDirector()
	up("directeur", "Directeur", 100, "formation", "Sans lui, la prison ne fonctionne pas. Il coûte 100 points d'université (l'école).")
	p.Spots = append(p.Spots, &Spot{ID: "directeur", Label: "Directeur", Zone: Rect{d.X + d.W - 180, d.Y + d.H - 150, 140, 130},
		Done: func(int) string { return "Le directeur est arrivé : la prison peut ouvrir !" }})
	for n := 1; n <= PrisCellN/2; n++ {
		up("surveillant_"+itoa(n), "Surveillant", float64(10*int(math.Pow(2, float64(n-1)))), "comptable",
			"Formé avec des points du lycée, il garde deux cellules : pas de cellule sans son surveillant.")
		gp := PrisGuardPos(n)
		spot("surveillant_"+itoa(n), "Surveillant", Rect{gp.X - 1, gp.Y - 1, 2, 2}, tr.F("Surveillant n° {n} formé !", "n", itoa(n)))
	}
	for n := 1; n <= PrisCellN; n++ {
		up("cellule_"+itoa(n), "Cellule", prisMoney(8e6, 1.25, n), "", "Une cellule pour un prisonnier.")
		spot("cellule_"+itoa(n), "Cellule", PrisCell(n), "Nouvelle cellule !")
	}
	for n := 1; n <= PrisVanN; n++ {
		up("fourgon_"+itoa(n), "Fourgon pénitentiaire", prisMoney(8e6, 2, n), "",
			"Il va chercher un voleur en garde à vue s'il y a une cellule libre (1 minute à l'aller, 1 minute au retour).")
		spot("fourgon_"+itoa(n), "Fourgon", PrisBay(n), "Nouveau fourgon !")
	}
	up("cantine", "Cantine", 64e6, "", "Les prisonniers y mangent (achète ensuite ses tables).")
	spot("cantine", "Cantine", PrisRoom(0), "La cantine ouvre !")
	for n := 1; n <= PrisTableN; n++ {
		up("table_"+itoa(n), "Table", []float64{16e6, 32e6, 64e6}[n-1], "", tr.F("Une table de {n} places.", "n", itoa(PrisTables[n-1])))
		spot("table_"+itoa(n), "Table", PrisTable(n), "Nouvelle table !")
	}
	up("bains", "Salle de bain", 64e6, "", "Les prisonniers s'y lavent (achète ensuite les douches et lavabos).")
	spot("bains", "Salle de bain", PrisRoom(1), "La salle de bain ouvre !")
	for n := 1; n <= PrisCabineN; n++ {
		up("cabine_"+itoa(n), "Douche et lavabo", prisMoney(8e6, 1.25, n), "", "Une cabine de plus.")
		spot("cabine_"+itoa(n), "Cabine", PrisCabin(n), "Nouvelle cabine !")
	}
	up("sport", "Salle de sport", 128e6, "", "Les prisonniers y font du sport (achète ensuite les équipements).")
	spot("sport", "Salle de sport", PrisRoom(3), "La salle de sport ouvre !")
	for n := 1; n <= PrisAgresN; n++ {
		a := PrisAgres[n-1]
		up("agres_"+itoa(n), a.Name, prisMoney(16e6, 1.25, n), "", tr.F("{n} place(s) de plus dans la salle de sport.", "n", itoa(a.Cap)))
		spot("agres_"+itoa(n), a.Name, PrisAgresAt(n), a.Name+" !")
	}
	for n := 1; n <= PrisFamilleN; n++ {
		up("famille_"+itoa(n), "Salle des familles", prisMoney(32e6, 2, n), "", "Un prisonnier y reçoit sa famille.")
		spot("famille_"+itoa(n), "Familles", PrisFamily(n), "Nouvelle salle des familles !")
	}
	up("travail", "Salle de travail", 256e6, "", "Les prisonniers y apprennent un métier (achète ensuite les postes).")
	spot("travail", "Salle de travail", PrisRoom(4), "La salle de travail ouvre !")
	for n := 1; n <= PrisPosteN; n++ {
		up("poste_"+itoa(n), "Poste de travail", prisMoney(16e6, 1.25, n), "", "Un poste de travail de plus.")
		spot("poste_"+itoa(n), "Poste", PrisPoste(n), "Nouveau poste de travail !")
	}
	up("psy", "Psychologue", 1000, "formation", "Elle reçoit un prisonnier à la fois (1 000 points d'université).")
	spot("psy", "Psy", PrisPsy(), "La psy est arrivée !")
	up("as", "Assistant social", 1000, "formation", "Il aide un prisonnier à la fois à préparer sa sortie (1 000 points d'université).")
	spot("as", "Assistant social", PrisAS(), "L'assistant social est arrivé !")
	p.setupYardH(PrisYardX+650, PrisParkY)
	return p
}

// prisLv : un niveau de la prison.
func (g *Game) prisLv(key string) int { return g.Levels["prison."+key] }

// HasWarden : le directeur est-il là ?
func (g *Game) HasWarden() bool { return g.prisLv("directeur") > 0 }

// prisSpotOK : cette case peut-elle s'afficher ? Une fois le directeur là, tout s'achète, dans
// l'ordre qu'on veut (demande) : une cellule demande seulement son surveillant, l'équipement d'une
// salle demande la salle.
func (p *Place) prisSpotOK(id string) bool {
	g := p.g
	if id == "directeur" {
		return true
	}
	if !g.HasWarden() {
		return false
	}
	var n int
	switch {
	case scanN(id, "cellule_", &n):
		return g.prisLv("surveillant_"+itoa((n+1)/2)) > 0
	case scanN(id, "table_", &n):
		return g.prisLv("cantine") > 0
	case scanN(id, "cabine_", &n):
		return g.prisLv("bains") > 0
	case scanN(id, "agres_", &n):
		return g.prisLv("sport") > 0
	case scanN(id, "poste_", &n):
		return g.prisLv("travail") > 0
	}
	return true
}

// ---------------------------------------------------------------- niveaux (points d'école)

// PrisImp : une amélioration (points d'école) : une cellule, une cabine, la cantine…
type PrisImp struct {
	Key  string // prison.niv_xxx
	Name string
	Type string // type de points
	N    int    // numéro (cellule, cabine, salle des familles), 0 sinon
}

// PrisImps : tout ce qui s'améliore (ce qui est acheté).
func (g *Game) PrisImps() []PrisImp {
	var out []PrisImp
	for n := 1; n <= PrisCellN; n++ {
		if g.prisLv("cellule_"+itoa(n)) > 0 {
			out = append(out, PrisImp{"niv_cellule_" + itoa(n), "Cellule", "ouvrier", n})
		}
	}
	for n := 1; n <= PrisCabineN; n++ {
		if g.prisLv("cabine_"+itoa(n)) > 0 {
			out = append(out, PrisImp{"niv_cabine_" + itoa(n), "Cabine", "ouvrier", n})
		}
	}
	if g.prisLv("cantine") > 0 {
		out = append(out, PrisImp{"niv_cantine", "Cantine", "marketing", 0})
	}
	if g.prisLv("sport") > 0 {
		out = append(out, PrisImp{"niv_sport", "Salle de sport", "marketing", 0})
	}
	for n := 1; n <= PrisFamilleN; n++ {
		if g.prisLv("famille_"+itoa(n)) > 0 {
			out = append(out, PrisImp{"niv_famille_" + itoa(n), "Salle des familles", "comptable", n})
		}
	}
	if g.prisLv("travail") > 0 {
		out = append(out, PrisImp{"niv_travail", "Salle de travail", "comptable", 0})
	}
	return out
}

// PrisImpCost : les points pour le niveau suivant (10 × n le niveau n ; 0 : au maximum).
func (g *Game) PrisImpCost(key string) int {
	l := g.Levels["prison."+key]
	if l >= PrisLvlMax {
		return 0
	}
	return 10 * (l + 1)
}

// BuyPrisImp : un niveau de plus.
func (g *Game) BuyPrisImp(im PrisImp) bool {
	c := g.PrisImpCost(im.Key)
	if c == 0 || g.Points(im.Type) < c {
		return false
	}
	g.Levels["pts."+im.Type] -= c
	g.Levels["prison."+im.Key]++
	return true
}

// ---------------------------------------------------------------- les besoins

// needSlots : les places où remplir un besoin (leur position, leur niveau).
func (p *Place) needSlots(need int, pr *Prisoner) (pos []V, lvl []int) {
	g := p.g
	add := func(v V, l int) { pos = append(pos, v); lvl = append(lvl, l) }
	switch need {
	case NeedDormir:
		add(PrisCell(pr.Cell).C(), g.prisLv("niv_cellule_"+itoa(pr.Cell)))
	case NeedRepas:
		if g.prisLv("cantine") > 0 {
			for t := 1; t <= PrisTableN; t++ {
				if g.prisLv("table_"+itoa(t)) == 0 {
					continue
				}
				r := PrisTable(t)
				n := PrisTables[t-1]
				for s := 0; s < n; s++ {
					x := r.X + 40 + float64(s%(n/2))*(r.W-80)/math.Max(1, float64(n/2-1))
					if n/2 == 1 {
						x = r.C().X
					}
					y := r.Y - 10
					if s >= n/2 {
						y = r.Y + r.H + 40
					}
					add(V{x, y}, g.prisLv("niv_cantine"))
				}
			}
		}
	case NeedBain:
		if g.prisLv("bains") > 0 {
			for n := 1; n <= PrisCabineN; n++ {
				if g.prisLv("cabine_"+itoa(n)) > 0 {
					add(PrisCabin(n).C().Add(V{0, 40}), g.prisLv("niv_cabine_"+itoa(n)))
				}
			}
		}
	case NeedSport:
		if g.prisLv("sport") > 0 {
			for n := 1; n <= PrisAgresN; n++ {
				if g.prisLv("agres_"+itoa(n)) > 0 {
					r := PrisAgresAt(n)
					for c := 0; c < PrisAgres[n-1].Cap; c++ {
						add(V{r.X + 50 + float64(c)*70, r.Y + r.H - 20}, g.prisLv("niv_sport"))
					}
				}
			}
		}
	case NeedFamille:
		for n := 1; n <= PrisFamilleN; n++ {
			if g.prisLv("famille_"+itoa(n)) > 0 {
				r := PrisFamily(n)
				add(V{r.C().X, r.Y + 120}, g.prisLv("niv_famille_"+itoa(n)))
			}
		}
	case NeedTravail:
		if g.prisLv("travail") > 0 {
			for n := 1; n <= PrisPosteN; n++ {
				if g.prisLv("poste_"+itoa(n)) > 0 {
					add(PrisPoste(n).C().Add(V{0, 40}), g.prisLv("niv_travail"))
				}
			}
		}
	case NeedPsy:
		if g.prisLv("psy") > 0 {
			add(PrisPsy().C().Add(V{60, 40}), PrisLvlMax)
		}
	case NeedAS:
		if g.prisLv("as") > 0 {
			add(PrisAS().C().Add(V{60, 40}), PrisLvlMax)
		}
	}
	return
}

// freeSlot : une place libre pour ce besoin (-1 : aucune).
func (p *Place) freeSlot(need int, pr *Prisoner) (int, V, int) {
	pos, lvl := p.needSlots(need, pr)
	for s := range pos {
		used := false
		for _, o := range p.Pris.Pris {
			if o != pr && o.Need == need && o.Slot == s && o.State != psCell && need != NeedDormir {
				used = true
				break
			}
		}
		if !used {
			return s, pos[s], lvl[s]
		}
	}
	return -1, V{}, 0
}

// ---------------------------------------------------------------- simulation

// PrisonHeld : les voleurs pris en charge par la prison (dans un fourgon ou en cellule).
func (g *Game) PrisonHeld() int {
	p := g.Place("prison")
	if p == nil || p.Pris == nil {
		return 0
	}
	n := len(p.Pris.Pris)
	for _, v := range p.Pris.Vans {
		if v.T > 0 && v.Carry {
			n++
		}
	}
	return n
}

// openPrison : la prison s'ouvre (offerte) dès qu'un premier groupe de voleurs est en garde à vue.
func (g *Game) openPrison() {
	p := g.Place("prison")
	if p == nil || p.Owned || g.Levels["commissariat.arrestations"] == 0 {
		return
	}
	p.Owned = true
	g.emit(Event{Kind: EvLevel, Text: "La prison ouvre ses portes (offerte) ! Il lui faut un directeur."})
}

// cellFree : la cellule n° n peut-elle recevoir un prisonnier (achetée, gardée, vide, pas attendue) ?
func (p *Place) cellFree(n int) bool {
	g := p.g
	if g.prisLv("cellule_"+itoa(n)) == 0 || g.prisLv("surveillant_"+itoa((n+1)/2)) == 0 {
		return false
	}
	for _, pr := range p.Pris.Pris {
		if pr.Cell == n {
			return false
		}
	}
	for _, v := range p.Pris.Vans {
		if v.T > 0 && v.Cell == n {
			return false
		}
	}
	return true
}

// takeCustody : un voleur quitte la garde à vue (le fourgon l'emmène).
func (g *Game) takeCustody() bool {
	for i, c := range g.Custody {
		if c.N > 0 {
			c.N--
			if c.N == 0 {
				g.Custody = append(g.Custody[:i], g.Custody[i+1:]...)
			}
			return true
		}
	}
	return false
}

func (p *Place) stepPrison(dt float64, active bool) {
	g := p.g
	if active && g.Player != nil {
		p.interact(g.Player, dt)
	}
	if !g.HasWarden() {
		return
	}
	pz := p.Pris
	// les fourgons
	for k := range pz.Vans {
		v := &pz.Vans[k]
		if g.prisLv("fourgon_"+itoa(k+1)) == 0 {
			continue
		}
		if v.T > 0 {
			v.T -= dt
			if v.T <= 0 {
				v.T = 0
				if v.Carry { // il arrive : le voleur entre dans sa cellule
					c := PrisCell(v.Cell)
					pr := &Prisoner{Pos: c.C(), Face: V{0, 1}, Look: 600 + 41*v.Cell + p.rng.Intn(40), Cell: v.Cell, Recid: PrisRecid, Need: -1}
					pz.Pris = append(pz.Pris, pr)
					g.emit(Event{Kind: EvLevel, Place: p.ID, Pos: c.C(), Text: tr.F("Un prisonnier arrive en cellule n° {n}.", "n", itoa(v.Cell))})
				}
				v.Carry = false
			}
			continue
		}
		if g.InCustody() == 0 {
			continue
		}
		for n := 1; n <= PrisCellN; n++ {
			if p.cellFree(n) && g.takeCustody() {
				v.T, v.Carry, v.Cell = PrisVanTrip, true, n
				break
			}
		}
	}
	// les prisonniers
	keep := pz.Pris[:0]
	for _, pr := range pz.Pris {
		if p.stepPrisoner(pr, dt) {
			keep = append(keep, pr)
		}
	}
	pz.Pris = keep
}

// stepPrisoner : faux quand il sort.
func (p *Place) stepPrisoner(pr *Prisoner, dt float64) bool {
	g := p.g
	pr.T += dt
	switch pr.State {
	case psCell:
		pr.Moving = false
		if pr.T >= PrisStay { // la sortie : réinséré, ou de retour parmi les voleurs
			roll := 1 + p.rng.Intn(100)
			if roll > pr.Recid {
				g.Levels["prison.reinseres"]++
				g.emit(Event{Kind: EvLevel, Place: p.ID, Pos: pr.Pos, Text: tr.F("Un prisonnier sort réinséré ({r} points de récidive) : un voleur de moins en ville !", "r", itoa(max(0, pr.Recid)))})
			} else {
				g.Levels["prison.recidivistes"]++
				g.emit(Event{Kind: EvLevel, Place: p.ID, Pos: pr.Pos, Text: tr.F("Un prisonnier sort ({r} points de récidive) : il retourne voler.", "r", itoa(pr.Recid))})
			}
			return false
		}
		if pr.Asked < NeedN && pr.T >= float64(pr.Asked+1)*PrisNeedEvery {
			pr.Asked++
			p.askNeed(pr)
		}
	case psGo, psBack:
		if p.walkPath(pr, dt) {
			if pr.State == psGo {
				pr.State, pr.Left = psDo, PrisActTime
			} else {
				pr.State, pr.Need = psCell, -1
			}
		}
	case psDo:
		pr.Moving = false
		if pr.Left -= dt; pr.Left <= 0 {
			pr.State = psBack
			pr.path = p.prisPath(pr.Pos, PrisCell(pr.Cell).C())
		}
	}
	return true
}

// needOrder : l'ordre où le prisonnier essaie ses besoins : la psy et l'assistant social d'abord
// (les plus rares), puis du besoin qui a le moins de places à celui qui en a le plus (selon ce que
// le joueur a acheté) ; dormir en dernier (sa cellule est toujours là).
func (p *Place) needOrder(pr *Prisoner) []int {
	order := []int{NeedPsy, NeedAS}
	rest := []int{NeedFamille, NeedTravail, NeedSport, NeedBain, NeedRepas}
	places := map[int]int{}
	for _, n := range rest {
		pos, _ := p.needSlots(n, pr)
		places[n] = len(pos)
	}
	for i := 1; i < len(rest); i++ { // (tri stable : à égalité, l'ordre de départ)
		for j := i; j > 0 && places[rest[j]] < places[rest[j-1]]; j-- {
			rest[j], rest[j-1] = rest[j-1], rest[j]
		}
	}
	return append(append(order, rest...), NeedDormir)
}

// askNeed : le prisonnier essaie ses besoins pas encore remplis, du plus rare au plus facile, et
// va au premier qui a une place libre ; s'il n'y en a aucune, ce besoin-là reste sans réponse.
func (p *Place) askNeed(pr *Prisoner) {
	first := -1
	for _, n := range p.needOrder(pr) {
		if pr.Done[n] {
			continue
		}
		if first < 0 {
			first = n
		}
		s, at, lvl := p.freeSlot(n, pr)
		if s < 0 {
			continue
		}
		pr.Done[n] = true
		pr.Recid -= p.g.reinsertMult(lvl) // (le cadeau de la prison : ×2)
		pr.Need, pr.Slot = n, s
		if n == NeedDormir { // (il reste dans sa cellule)
			pr.State, pr.Left = psDo, PrisActTime
			return
		}
		pr.State = psGo
		pr.path = p.prisPath(pr.Pos, at)
		return
	}
	if first >= 0 {
		pr.Done[first] = true
		pr.Recid += PrisFail
	}
}

// prisPath : de la cellule au grand couloir, puis à la pièce (et inversement).
func (p *Place) prisPath(from, to V) []V {
	main := PrisMainY() + PrisMainH/2
	var path []V
	exit := func(v V) []V { // d'une cellule (ou d'un quartier) jusqu'au grand couloir
		for z := 0; z < 2; z++ {
			if PrisZone(z).Has(v) {
				cor := PrisZoneCor(z)
				pass := PrisPass(z)
				return []V{{v.X, cor.C().Y}, {pass.C().X, cor.C().Y}, {pass.C().X, main}}
			}
		}
		return []V{{v.X, main}}
	}
	a := exit(from)
	b := exit(to)
	path = append(path, a...)
	for i := len(b) - 1; i >= 0; i-- {
		path = append(path, b[i])
	}
	return append(path, to)
}

func (p *Place) walkPath(pr *Prisoner, dt float64) bool {
	step := prisSpeed * dt
	for step > 0 && len(pr.path) > 0 {
		d := pr.path[0].Sub(pr.Pos)
		if d.Len() <= step {
			pr.Pos = pr.path[0]
			pr.path = pr.path[1:]
			step -= d.Len()
			continue
		}
		pr.Face = d.Norm()
		pr.Pos = pr.Pos.Add(pr.Face.Mul(step))
		step = 0
	}
	pr.Moving = len(pr.path) > 0
	pr.WalkT += dt
	return len(pr.path) == 0
}

// ---------------------------------------------------------------- sauvegarde

// PrisonSave : la prison dans la sauvegarde (elle reprend où elle en était : rien ne tourne jeu fermé).
type PrisonSave struct {
	Detenus  []DetenuSave `json:",omitempty"`
	Fourgons []VanSave    `json:",omitempty"`
}

// DetenuSave : un prisonnier.
type DetenuSave struct {
	Cellule, Recidive, Reclames int
	Temps                       float64
	Faits                       []int `json:",omitempty"`
}

// VanSave : un fourgon en route.
type VanSave struct {
	N, Cellule int
	Reste      float64
	Charge     bool
}

func (g *Game) prisonSave() *PrisonSave {
	p := g.Place("prison")
	if p == nil || p.Pris == nil {
		return nil
	}
	s := &PrisonSave{}
	for _, pr := range p.Pris.Pris {
		d := DetenuSave{Cellule: pr.Cell, Recidive: pr.Recid, Reclames: pr.Asked, Temps: pr.T}
		for n := 0; n < NeedN; n++ {
			if pr.Done[n] {
				d.Faits = append(d.Faits, n)
			}
		}
		s.Detenus = append(s.Detenus, d)
	}
	for k, v := range p.Pris.Vans {
		if v.T > 0 {
			s.Fourgons = append(s.Fourgons, VanSave{N: k + 1, Cellule: v.Cell, Reste: v.T, Charge: v.Carry})
		}
	}
	if len(s.Detenus) == 0 && len(s.Fourgons) == 0 {
		return nil
	}
	return s
}

func (g *Game) prisonRestore(s *PrisonSave) {
	p := g.Place("prison")
	if p == nil || p.Pris == nil || s == nil {
		return
	}
	for _, d := range s.Detenus {
		if d.Cellule < 1 || d.Cellule > PrisCellN {
			continue
		}
		pr := &Prisoner{Pos: PrisCell(d.Cellule).C(), Face: V{0, 1}, Look: 600 + 41*d.Cellule, Cell: d.Cellule, Recid: d.Recidive, Asked: d.Reclames, T: d.Temps, Need: -1}
		for _, n := range d.Faits {
			if n >= 0 && n < NeedN {
				pr.Done[n] = true
			}
		}
		p.Pris.Pris = append(p.Pris.Pris, pr)
	}
	for _, v := range s.Fourgons {
		if v.N >= 1 && v.N <= PrisVanN {
			p.Pris.Vans[v.N-1] = Van{T: v.Reste, Carry: v.Charge, Cell: v.Cellule}
		}
	}
}

// PrisonStats : les détenus, les réinsérés, les récidivistes.
func (g *Game) PrisonStats() (held, reins, recid int) {
	p := g.Place("prison")
	if p != nil && p.Pris != nil {
		held = len(p.Pris.Pris)
	}
	return held, g.Levels["prison.reinseres"], g.Levels["prison.recidivistes"]
}
