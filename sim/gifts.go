package sim

import (
	"math"
	"strings"
	"ville/tr"
)

// Les cadeaux des commandes : chaque commande livrée offre un cadeau qui
// dépend de la personne livrée (annoncé dès l'arrivée de la commande).
// La plupart des particuliers doublent les ventes du lieu pendant 5 min ; les
// personnes spéciales (rares, tirées au hasard comme les autres) offrent autre
// chose. Un seul cadeau à la fois : la commande suivante attend la fin du
// cadeau en cours. Le temps du cadeau ne s'écoule que jeu ouvert.

// BoostTime : durée d'un cadeau (doublée en livraison express).
const BoostTime = 300.0

// Boost : le cadeau en cours.
type Boost struct {
	Kind  string // ventes, ecole, maire, garage, fleurs, veto, marin
	Place string `json:",omitempty"` // lieu concerné (ventes, fleurs)
	Left  float64
	Total float64
	Mult  int `json:",omitempty"` // ×2 (×4 en livraison express ; 0 : ×2)
}

// giftOf : le cadeau qu'offre cette personne pour une commande livrée à « at ».
func giftOf(who string) string {
	switch who {
	case "la maîtresse d'école":
		return "universite"
	case "la cantine de l'école":
		return "ecole"
	case "le maire":
		return "maire"
	case "le garagiste":
		return "garage"
	case "la fleuriste":
		return "fleurs"
	case "le vétérinaire":
		return "veto"
	case "le marin":
		return "marin"
	}
	return "ventes"
}

// orderPerson : un habitant au hasard (pas le même que la dernière fois ; le
// marin seulement quand il y a un port). La réclame du lieu (points du
// collège) fait venir plus souvent une personne au cadeau spécial.
func (g *Game) orderPerson(at string) string {
	who := g.orderPerson0()
	if p := g.Place(at); p != nil && giftOf(who) == "ventes" && g.rng.Float64() < math.Min(0.5, p.ptsF("pts_reclame")/2) {
		for i := 0; i < 40; i++ { // (on retire jusqu'à tomber sur quelqu'un de spécial)
			if w := g.orderPerson0(); giftOf(w) != "ventes" {
				return w
			}
		}
	}
	return who
}

func (g *Game) orderPerson0() string {
	for {
		who := orderWho[g.rng.Intn(len(orderWho))]
		if strings.HasPrefix(g.lastOrder, who+"|") {
			continue
		}
		if who == "le marin" {
			if p := g.Place("port"); p == nil || !p.Owned {
				continue
			}
		}
		if who == "le garagiste" {
			// (camions et drones ×2 : pas avant le supermarché, sans camion ni drone ça ne sert à rien)
			if p := g.Place("super"); p == nil || !p.Owned {
				continue
			}
		}
		return who
	}
}

// giftDesc : le cadeau en toutes lettres, avec son multiplicateur (×2, ×4 en express).
func giftDesc(kind, place string, m int) string {
	switch kind {
	case "ventes":
		return "ventes ×" + itoa(m) + " " + placeAt(place)
	case "ecole":
		return "l'école forme " + itoa(m) + " fois plus vite"
	case "maire":
		if m == 2 {
			return "tous les lieux à moitié prix"
		}
		return "tous les lieux au quart du prix"
	case "garage":
		return "camions et drones " + itoa(m) + " fois plus rapides"
	case "fleurs":
		if m > 2 {
			return "grand embellissement " + placeOf(place)
		}
		return "embellissement " + placeOf(place)
	case "veto":
		return "animaux " + itoa(m) + " fois plus productifs dans toute la ville"
	case "marin":
		return "bateaux " + itoa(m) + " fois plus rapides"
	}
	return ""
}

// GiftText : le cadeau annoncé (« ventes ×2 à la crèmerie pendant 5 min »).
func (o *Order) GiftText() string {
	switch o.Gift {
	case "universite":
		return "1 point université"
	case "fleurs":
		return giftDesc(o.Gift, o.Place(), 2) + "…"
	}
	return giftDesc(o.Gift, o.Place(), 2) + " pendant " + durText(o.dur())
}

// GiftExpress : le cadeau doublé d'une livraison express (« ventes ×4 »).
func (o *Order) GiftExpress() string {
	if o.Gift == "universite" {
		return "2 points université"
	}
	return giftDesc(o.Gift, o.Place(), 4)
}

// Mul : le multiplicateur du cadeau (×2 ; ×4 en livraison express).
func (b *Boost) Mul() int {
	if b.Mult > 0 {
		return b.Mult
	}
	return 2
}

// BoostText : le cadeau en cours, en quelques mots (bandeau en haut de l'écran).
func (b *Boost) Text() string {
	m := b.Mul()
	switch b.Kind {
	case "ventes":
		return tr.F("Ventes ×{n} {lieu}", "n", itoa(m), "lieu", tr.T(placeAt(b.Place)))
	case "ecole":
		return tr.F("École ×{n}", "n", itoa(m))
	case "maire":
		if m == 2 {
			return tr.T("Lieux à moitié prix")
		}
		return tr.T("Lieux au quart du prix")
	case "garage":
		return tr.F("Camions et drones ×{n}", "n", itoa(m))
	case "fleurs":
		if m > 2 {
			return tr.F("Grand embellissement {lieu}", "lieu", tr.T(placeOf(b.Place)))
		}
		return tr.F("Embellissement {lieu}", "lieu", tr.T(placeOf(b.Place)))
	case "veto":
		return tr.F("Animaux ×{n}", "n", itoa(m))
	case "marin":
		return tr.F("Bateaux ×{n}", "n", itoa(m))
	}
	return ""
}

// placeAt : « à la crèmerie », « au supermarché »…
func placeAt(id string) string {
	t := placeTo(id)
	switch {
	case strings.HasPrefix(t, "au "), strings.HasPrefix(t, "chez "):
		return t
	}
	return "à " + t
}

// placeOf : « de la crèmerie », « du supermarché »…
func placeOf(id string) string {
	t := placeTo(id)
	switch {
	case strings.HasPrefix(t, "au "):
		return "du " + t[3:]
	case strings.HasPrefix(t, "chez le "):
		return "du " + t[8:]
	}
	return "de " + t
}

// giveGift : la commande est livrée, son cadeau commence (express : l'effet est
// doublé — ventes ×4, prix ÷4… —, pas la durée).
func (g *Game) giveGift(o *Order, express bool) string {
	if o.Gift == "universite" {
		if express {
			g.Levels["pts.formation"] += 2
			return "2 points université !"
		}
		g.Levels["pts.formation"]++
		return "1 point université !"
	}
	m := 2
	if express {
		m = 4
	}
	g.Boost = &Boost{Kind: o.Gift, Place: o.Place(), Left: o.dur(), Total: o.dur(), Mult: m}
	return giftDesc(o.Gift, o.Place(), m) + " pendant " + durText(o.dur()) + " !"
}

func (g *Game) stepBoost(dt float64) {
	b := g.Boost
	if b == nil {
		return
	}
	b.Left -= dt
	if b.Left <= 0 {
		g.Boost = nil
		g.emit(Event{Kind: EvLevel, Pos: g.Player.Pos, Text: "Fin du cadeau : " + lowerFirst(b.Text()) + "."})
	}
}

func lowerFirst(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return s
	}
	return strings.ToLower(string(r[:1])) + string(r[1:])
}

// Boosted : ce cadeau est-il en cours (place "" : n'importe quel lieu) ?
func (g *Game) Boosted(kind, place string) bool {
	b := g.Boost
	return b != nil && b.Kind == kind && (place == "" || b.Place == place)
}

// PriceOf : prix d'achat d'un lieu (moitié prix pendant le cadeau du maire).
func (g *Game) PriceOf(p *Place) float64 {
	if g.Boosted("maire", "") {
		return p.Price / float64(g.Boost.Mul())
	}
	return p.Price
}

// boostF : multiplicateur d'un cadeau (2 s'il est en cours, sinon 1).
func (g *Game) boostF(kind, place string) float64 {
	if g.Boosted(kind, place) {
		return float64(g.Boost.Mul())
	}
	return 1
}

// durText : « 5 min », « 6 min 30 ».
func durText(secs float64) string {
	m, sec := int(secs)/60, int(secs)%60
	if sec == 0 {
		return itoa(m) + " min"
	}
	if sec < 10 {
		return itoa(m) + " min 0" + itoa(sec)
	}
	return itoa(m) + " min " + itoa(sec)
}

// dur : durée du cadeau (fixée à l'annonce : la réclame du lieu l'allonge).
func (o *Order) dur() float64 {
	if o.Dur > 0 {
		return o.Dur
	}
	return BoostTime
}
