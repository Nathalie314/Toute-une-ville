package sim

import (
	"fmt"
	"math"
	"testing"
)

// TestPaliers : la règle des paliers (prices.go). Si ce test échoue après un ajout :
//   - un nouveau lieu : ajouter son coefficient à attendus (il est calculé par la règle) ;
//   - des marchandises ajoutées à un lieu existant : les anciens prix ne doivent pas baisser.
//     Monter extension[lieu] (prices.go) jusqu'à retrouver le coefficient attendu ; le test
//     vérifie aussi que le lieu reste dans la norme (tout débloquer coûte au plus 3 fois le
//     lieu suivant). S'il en sort, les nouvelles marchandises sont trop chères : baisser leurs prix.
func TestPaliers(t *testing.T) {
	attendus := map[string]float64{
		"cremerie": 2.857, "exotique": 0.9877, "boulangerie": 22.22, "glacier": 14.91, "port": 1.284, "elevage": 49.36,
		"boucherie": 43.15, "domaine": 90.25, "creperie": 76, "patisserie": 1506, "parc": 14.5, "huilerie": 933.7,
	}
	g := NewGame()
	for _, p := range g.Places {
		r := MapRank(p.ID)
		if r < palierFrom {
			continue
		}
		want, ok := attendus[p.ID]
		if !ok {
			t.Errorf("%s : nouveau lieu, coefficient %.4g à ajouter à attendus", p.ID, p.Scale)
			continue
		}
		if math.Abs(p.Scale/want-1) > 0.01 {
			t.Errorf("%s : coefficient %.4g au lieu de %.4g (les prix du lieu ont bougé : voir extension, prices.go)", p.ID, p.Scale, want)
		}
		sum := 0.0
		for _, u := range p.Unlocks {
			sum += u.Cost
		}
		if k := sum / nextPrice(r); k > 3*palier/1.2 {
			t.Errorf("%s : tout débloquer coûte %.1f fois le lieu suivant (norme : %.1f) : nouvelles marchandises trop chères", p.ID, k, palier)
		}
	}
}

// TestSuperPaliers : au supermarché, les rayons des produits d'un lieu coûtent ensemble le
// quart de son prix d'achat, ses magasiniers la moitié (superPrices). Un nouveau produit
// vendu au supermarché doit avoir un lieu d'origine (SourceOf), sinon ses prix ne suivent pas.
func TestSuperPaliers(t *testing.T) {
	g := NewGame()
	sup := g.Place("super")
	ray, mag := map[string]float64{}, map[string]float64{}
	for _, u := range sup.Unlocks {
		var it Item
		if _, err := fmt.Sscanf(u.ID, "r_%d", &it); err == nil {
			id := g.SourceOf(it)
			if id == "" {
				t.Errorf("rayon %s : produit sans lieu d'origine", Items[it].Name)
			}
			ray[id] += u.Cost
		}
	}
	for _, z := range SuperZones {
		for _, u := range sup.ups {
			if u.Key != sup.key(teamKey(z.ID)) {
				continue
			}
			for l, it := range z.Items {
				if l < len(u.Costs) {
					mag[g.SourceOf(it)] += u.Costs[l]
				}
			}
		}
	}
	for id, r := range ray {
		q := g.Place(id)
		if q == nil || MapRank(id) < palierFrom {
			continue
		}
		if math.Abs(r/(superRayons*q.Price)-1) > 0.05 {
			t.Errorf("%s : ses rayons coûtent %.3g au lieu de %.3g", id, r, superRayons*q.Price)
		}
		if math.Abs(mag[id]/(superMagas*q.Price)-1) > 0.05 {
			t.Errorf("%s : ses magasiniers coûtent %.3g au lieu de %.3g", id, mag[id], superMagas*q.Price)
		}
	}
}
