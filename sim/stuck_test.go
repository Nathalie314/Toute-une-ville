package sim

import "testing"

// Un ouvrier doit relier le champ de blé au dépôt du moulin sans rester bloqué.
func TestWorkerReachesMill(t *testing.T) {
	g := NewGame()
	p := g.farm()
	for _, u := range p.Unlocks {
		if u.ID == "ble" || u.ID == "moulin" || u.ID == "moulin_mais" || u.ID == "poulailler" || u.ID == "etable" {
			p.complete(u)
		}
	}
	w := &Agent{Kind: WorkerAgent}
	for _, from := range []V{{360, 300}, {200, 420}, {900, 430}, {1200, 300}} {
		for _, to := range []Rect{p.Station("moulin").Zone, p.Station("moulin_mais").Zone, p.Station("etable").Zone} {
			w.Pos = from
			w.route = nil
			w.routeTo = V{}
			ok := false
			for i := 0; i < 60*20; i++ {
				if p.walk(w, to.C(), 170, 1.0/60) {
					ok = true
					break
				}
			}
			if !ok {
				t.Errorf("bloqué de %v vers %v (arrêté en %v)", from, to.C(), w.Pos)
			}
		}
	}
}
