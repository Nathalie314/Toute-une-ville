package sim

import (
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
		"exotique": 0.9877, "boulangerie": 22.22, "glacier": 14.91, "port": 1.284, "elevage": 49.36,
		"boucherie": 43.15, "domaine": 90.25, "creperie": 76, "patisserie": 1506, "parc": 14.5, "huilerie": 933.7,
	}
	g := NewGame()
	for _, p := range g.Places {
		r := MapRank(p.ID)
		if r < scaleFrom {
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
