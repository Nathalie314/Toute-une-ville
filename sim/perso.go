package sim

import "math"

// Le personnage du joueur : des équipements à niveaux (clés « perso.xxx » dans
// Levels), un porté par zone (tête, dos, pieds ; une ou deux mains). Les zones
// se cumulent. L'ensemble du diplômé n'agit que dans la classe où l'on se trouve.

// Gear : un équipement à niveaux.
type Gear struct {
	ID    string   // clé : perso.<ID>
	Slot  string   // tete, dos, pieds, main
	Names []string // nom de chaque niveau (1 à n)
	Costs []float64
	Desc  func(l int) string
}

func (g *Gear) Key() string { return "perso." + g.ID }
func (g *Gear) Max() int    { return len(g.Costs) }

// Upgrade : l'équipement vu comme une amélioration ordinaire.
func (gr *Gear) Upgrade() *Upgrade {
	return &Upgrade{Key: gr.Key(), Name: gr.Names[0], Max: gr.Max(), Costs: gr.Costs, Desc: gr.Desc}
}

// bonus par niveau
var shoeSpeed = []float64{0, 0.10, 0.20, 0.30, 0.45, 0.60, 0.75, 0.90, 1.10, 1.30, 1.55, 1.80, 2.10}
var bagCap = []int{0, 2, 4, 6, 8, 11, 14, 17, 20, 24, 28, 33, 40}

func toqueF(l int) float64   { return 1 + 0.25*float64(l) }
func toqueR(l int) float64   { return 250 + 30*float64(l) }
func waterF(l int) float64   { return 1 + 0.3*float64(l) }
func waterR(l int) float64   { return 250 + 40*float64(l) }
func whistleF(l int) float64 { return 1 + 0.2*float64(l) }
func whistleR(l int) float64 { return 400 + 50*float64(l) }
func megaF(l int) float64    { return 1 + 0.15*float64(l) }

// diplomeF : bonus de formation d'une pièce de l'ensemble du diplômé (+20 % par niveau).
func diplomeF(l int) float64 { return 0.2 * float64(l) }

// Diplome : l'ensemble du diplômé.
var Diplome = []string{"chapeau", "toge", "livre", "mocassins"}

func diplomeDesc(l int) string {
	return "Dans la classe où tu te trouves : formation " + pct(100*diplomeF(l)) + " plus rapide."
}

// Gears : tous les équipements.
var Gears = []*Gear{
	{ID: "chaussures", Slot: "pieds",
		Names: []string{"Sandales", "Baskets", "Chaussures de course", "Chaussures à crampons", "Rollers", "Bottes de randonnée", "Bottes de sept lieues", "Chaussures ailées",
			"Rollers à réaction", "Patins à fusée", "Bottes de comète", "Sandales de l'éclair"},
		Costs: []float64{500, 1500, 4000, 10000, 25000, 60000, 140000, 300000, 650000, 1.4e6, 3e6, 6.5e6},
		Desc:  func(l int) string { return "Tu marches " + pct(100*shoeSpeed[l]) + " plus vite." }},
	{ID: "sac", Slot: "dos",
		Names: []string{"Cartable", "Sac à dos", "Sac de randonnée", "Grand sac", "Hotte", "Sac de montagne", "Sac de voyage", "Sac sans fond",
			"Sac de déménageur", "Malle sur roulettes", "Sac de géant", "Sac magique"},
		Costs: []float64{250, 800, 2500, 7000, 18000, 45000, 110000, 260000, 600000, 1.3e6, 2.8e6, 6e6},
		Desc:  func(l int) string { return "Tu portes " + itoa(bagCap[l]) + " objets de plus." }},
	{ID: "mocassins", Slot: "pieds",
		Names: []string{"Mocassins", "Mocassins cirés", "Mocassins à glands", "Mocassins vernis", "Mocassins d'honneur", "Mocassins de recteur", "Mocassins d'académicien", "Mocassins de prix Nobel", "Mocassins légendaires"},
		Costs: []float64{2500, 8000, 24000, 70000, 180000, 450000, 1.1e6, 2.6e6, 6e6}, Desc: diplomeDesc},
	{ID: "toge", Slot: "dos",
		Names: []string{"Toge d'étudiant", "Toge de licence", "Toge de master", "Toge de docteur", "Toge d'honneur", "Toge de recteur", "Toge d'académicien", "Toge de prix Nobel", "Toge légendaire"},
		Costs: []float64{3000, 9000, 27000, 80000, 200000, 500000, 1.2e6, 2.8e6, 6.5e6}, Desc: diplomeDesc},
	{ID: "toque", Slot: "tete",
		Names: []string{"Toque de commis", "Toque de cuisinier", "Toque de second", "Toque de chef", "Toque de grand chef", "Toque étoilée", "Toque deux étoiles", "Toque trois étoiles", "Toque de meilleur ouvrier", "Toque d'or"},
		Costs: []float64{2000, 6000, 15000, 40000, 100000, 250000, 600000, 1.4e6, 3.2e6, 7e6},
		Desc: func(l int) string {
			return "Les bâtiments près de toi vont " + pct(100*(toqueF(l)-1)) + " plus vite (rayon " + itoa(int(toqueR(l))) + ")."
		}},
	{ID: "chapeau", Slot: "tete",
		Names: []string{"Chapeau en carton", "Chapeau de diplômé", "Chapeau à pompon", "Chapeau doré", "Chapeau d'honneur", "Chapeau de recteur", "Chapeau d'académicien", "Chapeau de prix Nobel", "Chapeau légendaire"},
		Costs: []float64{3500, 10000, 30000, 90000, 220000, 550000, 1.3e6, 3e6, 7e6}, Desc: diplomeDesc},
	{ID: "arrosoir", Slot: "main",
		Names: []string{"Petit arrosoir", "Arrosoir", "Grand arrosoir", "Tuyau d'arrosage", "Arroseur automatique", "Brumisateur", "Nuage portatif", "Pluie de printemps", "Arrosoir magique"},
		Costs: []float64{1500, 5000, 15000, 45000, 120000, 300000, 750000, 1.8e6, 4.2e6},
		Desc: func(l int) string {
			return "En main : les pousses près de toi poussent " + pct(100*(waterF(l)-1)) + " plus vite."
		}},
	{ID: "sifflet", Slot: "main",
		Names: []string{"Sifflet en plastique", "Sifflet", "Sifflet d'arbitre", "Sifflet en argent", "Sifflet en or", "Sifflet en platine", "Sifflet à ultrasons", "Corne de brume", "Sifflet magique"},
		Costs: []float64{2500, 8000, 25000, 70000, 180000, 450000, 1.1e6, 2.6e6, 6e6},
		Desc: func(l int) string {
			return "En main : les ouvriers près de toi vont " + pct(100*(whistleF(l)-1)) + " plus vite."
		}},
	{ID: "megaphone", Slot: "main",
		Names: []string{"Cornet en carton", "Porte-voix", "Mégaphone", "Grand mégaphone", "Haut-parleur géant", "Sono de concert", "Écran géant", "Spot publicitaire", "Star de la télé"},
		Costs: []float64{4000, 12000, 35000, 90000, 220000, 550000, 1.3e6, 3e6, 7e6},
		Desc: func(l int) string {
			return "En main : " + pct(100*(megaF(l)-1)) + " de clients en plus dans le lieu où tu es."
		}},
	{ID: "livre", Slot: "main",
		Names: []string{"Livre de poche", "Manuel scolaire", "Gros dictionnaire", "Encyclopédie", "Livre d'or", "Grimoire", "Bibliothèque de poche", "Savoir universel", "Livre légendaire"},
		Costs: []float64{2000, 7000, 20000, 60000, 160000, 400000, 1e6, 2.4e6, 5.5e6}, Desc: diplomeDesc},
}

// Slots : les zones du personnage, dans l'ordre de la page.
var Slots = []string{"tete", "dos", "main", "pieds"}

// SlotGears : les équipements d'une zone.
func SlotGears(slot string) []*Gear {
	var out []*Gear
	for _, g := range Gears {
		if g.Slot == slot {
			out = append(out, g)
		}
	}
	return out
}

// Worn : l'équipement est-il porté (ou tenu en main) ?
func (g *Game) Worn(id string) bool {
	gr := GearByID(id)
	if gr == nil || g.GearLevel(id) == 0 {
		return false
	}
	if gr.Slot == "main" {
		return g.Holding(id)
	}
	return g.Levels["porte."+id] > 0
}

// Wear : porter un équipement (on enlève l'autre de la même zone), ou l'enlever s'il est déjà porté.
func (g *Game) Wear(id string) {
	gr := GearByID(id)
	if gr == nil || g.GearLevel(id) == 0 {
		return
	}
	if gr.Slot == "main" {
		g.ToggleHold(id)
		return
	}
	if g.Worn(id) {
		delete(g.Levels, "porte."+id)
		return
	}
	for _, o := range SlotGears(gr.Slot) {
		delete(g.Levels, "porte."+o.ID)
	}
	g.Levels["porte."+id] = 1
}

// WornIn : l'équipement porté dans une zone (hors mains).
func (g *Game) WornIn(slot string) string {
	for _, o := range SlotGears(slot) {
		if g.Worn(o.ID) {
			return o.ID
		}
	}
	return ""
}

// autoWear : le premier niveau acheté se porte tout seul si la zone est libre.
func (g *Game) autoWear(key string) {
	if len(key) < 7 || key[:6] != "perso." {
		return
	}
	gr := GearByID(key[6:])
	if gr == nil || g.GearLevel(gr.ID) != 1 {
		return
	}
	if gr.Slot == "main" {
		if len(g.Held()) < g.Hands() {
			g.ToggleHold(gr.ID)
		}
		return
	}
	if g.WornIn(gr.Slot) == "" {
		g.Wear(gr.ID)
	}
}

// ClassBoost : bonus de formation de l'ensemble du diplômé dans une classe
// (seulement si le joueur s'y trouve).
func (p *Place) ClassBoost(r *Room) float64 {
	g := p.g
	if g.Cur != p || g.Player == nil || !r.Body.Has(g.Player.Pos) {
		return 1
	}
	f := 1.0
	for _, id := range Diplome {
		if g.Worn(id) {
			f += diplomeF(g.GearLevel(id))
		}
	}
	return f
}

// Ambidextre : deuxième main.
var Ambidextre = &Upgrade{Key: "perso.ambidextre", Name: "Ambidextre", Max: 1, Costs: []float64{150000},
	Desc: func(int) string { return "Tu tiens deux objets à la fois." }}

// GearByID : un équipement.
func GearByID(id string) *Gear {
	for _, g := range Gears {
		if g.ID == id {
			return g
		}
	}
	return nil
}

// GearLevel : niveau d'un équipement (0 : pas acheté).
func (g *Game) GearLevel(id string) int { return g.Levels["perso."+id] }

// GearName : nom actuel d'un équipement acheté.
func (g *Game) GearName(id string) string {
	gr := GearByID(id)
	l := g.GearLevel(id)
	if gr == nil || l == 0 {
		return ""
	}
	return gr.Names[l-1]
}

// HandItems : les objets de main.
var HandItems = []string{"arrosoir", "sifflet", "megaphone", "livre"}

// Holding : l'objet est-il en main ?
func (g *Game) Holding(id string) bool {
	return g.GearLevel(id) > 0 && (g.Levels["main1."+id] > 0 || g.Levels["main2."+id] > 0)
}

// Hands : nombre de mains disponibles.
func (g *Game) Hands() int { return 1 + g.Levels["perso.ambidextre"] }

// Held : objets tenus.
func (g *Game) Held() []string {
	var out []string
	for _, id := range HandItems {
		if g.Holding(id) {
			out = append(out, id)
		}
	}
	return out
}

// ToggleHold : prendre ou poser un objet (si les mains sont pleines, on pose le plus ancien).
func (g *Game) ToggleHold(id string) {
	if g.GearLevel(id) == 0 {
		return
	}
	if g.Holding(id) {
		delete(g.Levels, "main1."+id)
		delete(g.Levels, "main2."+id)
		return
	}
	held := g.Held()
	if len(held) >= g.Hands() {
		// on pose le premier
		delete(g.Levels, "main1."+held[0])
		delete(g.Levels, "main2."+held[0])
	}
	g.Levels["main1."+id] = 1
}

// BuyGear : niveau suivant d'un équipement. Le premier objet de main acheté est pris en main.
func (g *Game) BuyGear(gr *Gear) bool {
	return g.Buy(gr.Upgrade()) // (Buy le fait porter s'il est neuf et que la zone est libre)
}

// ---------------------------------------------------------------- effets

func (g *Game) playerSpeedF() float64 {
	if !g.Worn("chaussures") {
		return 1
	}
	return 1 + shoeSpeed[min(g.GearLevel("chaussures"), len(shoeSpeed)-1)]
}
func (g *Game) playerCapBonus() int {
	if !g.Worn("sac") {
		return 0
	}
	return bagCap[min(g.GearLevel("sac"), len(bagCap)-1)]
}

// toqueBoost : multiplicateur d'un bâtiment près du joueur.
func (p *Place) toqueBoost(s *Station) float64 {
	g := p.g
	l := g.GearLevel("toque")
	if l == 0 || !g.Worn("toque") || g.Cur != p || g.Player == nil || g.Player.Pos.Dist(s.Body.C()) > toqueR(l) {
		return 1
	}
	return toqueF(l)
}

// waterBoost : multiplicateur de pousse près du joueur (arrosoir en main).
func (p *Place) waterBoost(at V) float64 {
	g := p.g
	if !g.Holding("arrosoir") || g.Cur != p || g.Player == nil {
		return 1
	}
	l := g.GearLevel("arrosoir")
	if g.Player.Pos.Dist(at) > waterR(l) {
		return 1
	}
	return waterF(l)
}

// whistleBoost : multiplicateur de vitesse d'un ouvrier près du joueur (sifflet en main).
func (p *Place) whistleBoost(a *Agent) float64 {
	g := p.g
	if !g.Holding("sifflet") || g.Cur != p || g.Player == nil {
		return 1
	}
	l := g.GearLevel("sifflet")
	if g.Player.Pos.Dist(a.Pos) > whistleR(l) {
		return 1
	}
	return whistleF(l)
}

// megaBoost : clients en plus dans le lieu du joueur (mégaphone en main).
func (p *Place) megaBoost() float64 {
	g := p.g
	if !g.Holding("megaphone") || g.Cur != p {
		return 1
	}
	return megaF(g.GearLevel("megaphone"))
}

// GearRadius : rayon d'effet d'un objet tenu (0 si sans rayon), pour l'affichage.
func (g *Game) GearRadius(id string) float64 {
	l := g.GearLevel(id)
	switch id {
	case "toque":
		return toqueR(l)
	case "arrosoir":
		return waterR(l)
	case "sifflet":
		return whistleR(l)
	}
	return 0
}

var _ = math.Abs
