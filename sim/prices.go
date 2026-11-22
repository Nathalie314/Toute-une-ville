package sim

import (
	"math"
	"strings"
)

// La courbe des prix : un lieu coûte d'autant plus cher qu'il est loin sur la
// carte (MapOrder), à l'achat comme pour tout ce qu'on y achète (déblocages,
// employés, étals, drones, stocks, pousses…). Le but : inciter à ouvrir les
// lieux dans l'ordre de la carte, et qu'un nouveau lieu demande d'attendre.
//
// Les quatre premiers lieux (ferme, école, supermarché, crèmerie) gardent leurs
// prix : le début reste rapide. Ensuite, chaque lieu coûte 2 fois plus que le
// précédent : son prix d'achat (PlacePrice) et le prix total de son contenu
// (ContentTarget). Pour y arriver, tous les prix du lieu sont multipliés par un
// même coefficient (Place.Scale), calculé au lancement : l'équilibre interne
// du lieu (ce qui est cher, ce qui ne l'est pas) ne change pas.
//
// Pour que ça tienne avec les futurs lieux, ceux-ci devront rapporter plus
// (produits plus chers) : chaque lieu doit à peu près doubler les revenus.

const (
	scaleFrom   = 5      // premier rang dont les prix suivent la courbe (la ferme exotique)
	priceFrom   = 500000 // prix d'achat du lieu de rang scaleFrom
	scaleFactor = 2.0    // d'un lieu au suivant
	// refContent : le prix total du contenu de la crèmerie (lieu de rang scaleFrom-1) quand la
	// courbe a été équilibrée. Il est figé : ajouter quelque chose à la crèmerie (ses vitrines à
	// 100 produits avaient multiplié ce total par 85, et tous les prix à partir de la ferme
	// exotique avec lui) ne doit plus changer les prix des autres lieux.
	refContent = 15881319.0
)

// PlacePrice : prix d'achat d'un lieu de ce rang (0 : on garde son prix).
func PlacePrice(rank int) float64 {
	if rank < scaleFrom {
		return 0
	}
	return priceFrom * math.Pow(scaleFactor, float64(rank-scaleFrom))
}

// ContentTarget : prix total du contenu d'un lieu de ce rang, d'après celui du
// lieu de rang scaleFrom-1 (la crèmerie).
func ContentTarget(rank int, ref float64) float64 {
	return ref * math.Pow(scaleFactor, float64(rank-scaleFrom+1))
}

// LA RÈGLE DES PALIERS (21/11) : un joueur qui arrive dans un lieu doit pouvoir tout y
// débloquer avant de passer au lieu suivant, et pas beaucoup plus tôt. Pour chaque lieu à
// partir de la crèmerie, le total de ses déblocages (champs, machines, bateaux, bassins…)
// vaut palier × le prix du lieu suivant sur la carte ; le coefficient du lieu (Place.Scale) en
// découle, et il s'applique à tout ce qu'on y achète (employés, cases, drones, étals…).
//
// Avant, le coefficient venait du contenu total du lieu (tous les niveaux de tout), ce qui
// donnait des lieux 10 à 40 fois trop chers (la ferme exotique, le port) ou trop bon marché.
//
// Ajouter des marchandises à un lieu existant : les nouvelles doivent coûter plus cher, sans
// faire baisser les anciennes. On monte alors le budget du lieu (extension[id]) pour que son
// coefficient ne bouge pas ; TestPaliers (prices_test.go) vérifie les coefficients et que chaque
// lieu reste dans la norme (ce qui arriverait sinon : la règle ferait baisser les anciens prix).
const palier = 1.2

// palierFrom : premier rang soumis à la règle (la crèmerie ; la ferme et l'école, lieux de départ
// avec le tutoriel, gardent leurs prix ; le supermarché suit les lieux qui le fournissent :
// voir superPrices).
const palierFrom = 4

// extension : budget en plus d'un lieu agrandi après coup (1 : aucun).
var extension = map[string]float64{}

// nextPrice : le prix du lieu qui suit sur la carte (pour le dernier : deux fois le sien).
func nextPrice(r int) float64 { return PlacePrice(r + 1) }

// niceRound : un prix arrondi à 3 chiffres (51 234 → 51 200).
func niceRound(v float64) float64 {
	if v < 1000 {
		return math.Round(v)
	}
	e := math.Pow(10, math.Floor(math.Log10(v))-2)
	return math.Round(v/e) * e
}

// scaleOf : coefficient des prix d'un lieu (1 si rien).
func (g *Game) scaleOf(id string) float64 {
	if p := g.Place(id); p != nil && p.Scale > 0 {
		return p.Scale
	}
	return 1
}

// keyScale : coefficient des prix d'une amélioration, d'après le lieu de sa clé
// (« exotique.ouvriers_cafe » → la ferme exotique).
func (g *Game) keyScale(key string) float64 {
	id, _, ok := strings.Cut(key, ".")
	if !ok {
		return 1
	}
	return g.scaleOf(id)
}

// rawCost : prix d'un niveau sans aucun coefficient.
func (u *Upgrade) rawCost(lvl int) float64 {
	if lvl < len(u.Costs) {
		return u.Costs[lvl]
	}
	return math.Round(u.Base * math.Pow(u.Mult, float64(lvl)))
}

// rawPlantCost : prix de la pousse n° k+1 d'un champ, sans coefficient.
func (s *Station) rawPlantCost(k int) float64 {
	if s.Column {
		return math.Round(Items[s.Item].Price * 12 * math.Pow(1.4, float64(k-1)))
	}
	return math.Round(Items[s.Item].Price * math.Pow(1.18, float64(k-1)))
}

// contentCost : prix de tout ce qu'on peut acheter dans un lieu (sans coefficient) :
// déblocages, améliorations posées au sol (tous les niveaux), niveaux et stocks
// des bâtiments, pousses des champs.
func (p *Place) contentCost() float64 {
	sum := 0.0
	for _, u := range p.Unlocks {
		sum += u.Cost
	}
	seen := map[string]bool{}
	var ups []*Upgrade
	for _, sp := range p.Spots {
		if u := p.SpotUpgrade(sp); u != nil && !seen[u.Key] {
			seen[u.Key] = true
			ups = append(ups, u)
		}
	}
	for _, s := range p.Stations {
		if s.Upgradable() {
			ups = append(ups, p.StationUpgrade(s))
		}
		if s.Stockable() {
			ups = append(ups, p.StockUpgrade(s))
		}
		if s.Kind == Field {
			for k := 1; k < len(s.Slots); k++ {
				sum += s.rawPlantCost(k)
			}
		}
	}
	for _, u := range ups {
		for l := 0; l < u.Max && l < 60; l++ {
			sum += u.rawCost(l)
		}
	}
	return sum
}

// applyMapPrices : fixe le coefficient, le prix d'achat et les déblocages de
// chaque lieu selon son rang sur la carte (après fixRayonCosts : les rayons du
// supermarché, 3e lieu, gardent leurs prix).
func (g *Game) applyMapPrices() {
	ref := refContent
	for _, p := range g.Places {
		r := MapRank(p.ID)
		if r < palierFrom || ref <= 0 {
			continue
		}
		raw := 0.0 // les déblocages, sans coefficient
		for _, u := range p.Unlocks {
			raw += u.Cost
		}
		if raw > 0 {
			k := palier
			if e, ok := extension[p.ID]; ok {
				k *= e
			}
			p.Scale = k * nextPrice(r) / raw
		} else if c := p.contentCost(); c > 0 {
			p.Scale = ContentTarget(r, ref) / c
		} else {
			continue
		}
		if r >= scaleFrom {
			p.Price = PlacePrice(r) // (la crèmerie garde son prix d'achat)
		}
		for _, u := range p.Unlocks {
			u.Cost = niceRound(u.Cost * p.Scale)
		}
		for _, s := range p.Stations {
			s.scale = p.Scale
		}
	}
}

// scale1 : coefficient des prix du lieu (1 si rien).
func (p *Place) scale1() float64 {
	if p.Scale > 0 {
		return p.Scale
	}
	return 1
}
