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

// frozenScale : le coefficient des lieux déjà construits, figé (21/11). Ajouter du contenu
// à un lieu (le second ponton du port…) ne doit plus faire baisser les prix de ce qui y
// est déjà : avant, le coefficient était recalculé pour que le contenu total garde son prix.
var frozenScale = map[string]float64{
	"exotique": 7.57, "boulangerie": 12.26, "glacier": 8.25, "port": 40.95, "elevage": 122.09,
	"boucherie": 42.36, "domaine": 334.49, "creperie": 48.38, "patisserie": 752.14, "parc": 22.57,
	"huilerie": 5239.05,
}

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
		if r < scaleFrom || ref <= 0 {
			continue
		}
		c := p.contentCost()
		if c <= 0 {
			continue
		}
		p.Scale = ContentTarget(r, ref) / c
		if f, ok := frozenScale[p.ID]; ok {
			p.Scale = f
		}
		p.Price = PlacePrice(r)
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
