package sim

import "testing"

// TestVol : un vol dure [lieux débloqués en haut] minutes ; pendant ce temps, les ouvriers se
// cachent et aucun client n'arrive ; à la fin, tout reprend.
func TestVol(t *testing.T) {
	g := NewGame()
	g.Place("super").Owned = true
	p := g.Place("ferme")
	g.Enter(p)
	for i := 0; i < 30*60; i++ {
		g.Step(1.0 / 30)
	}
	if g.Thieves() != thievesBase+thievesPerLieu*2 {
		t.Fatalf("voleurs : %d", g.Thieves())
	}
	g.StartTheft("ferme")
	if g.Theft == nil || g.Theft.Total != 120 {
		t.Fatalf("vol de %v s au lieu de 120 (ferme et supermarché)", g.Theft)
	}
	for i := 0; i < 30*60; i++ {
		g.Step(1.0 / 30)
	}
	for _, c := range p.Customers {
		if c.State != cLeave && c.State != cGone && !c.InCar() {
			t.Fatalf("un client est resté pendant le vol")
		}
	}
	for g.Theft != nil {
		g.Step(1.0 / 30)
	}
	for _, w := range p.Workers {
		if w.Hidden {
			t.Fatalf("un ouvrier est resté caché après le vol")
		}
	}
}

// TestVolSauvegarde : quitter le jeu ne chasse pas les voleurs (le vol reprend, en pause jeu fermé).
func TestVolSauvegarde(t *testing.T) {
	g := NewGame()
	g.Place("super").Owned = true
	g.Enter(g.Place("ferme"))
	g.StartTheft("ferme")
	for i := 0; i < 30*30; i++ {
		g.Step(1.0 / 30)
	}
	left := g.Theft.Left
	s := g.Snapshot(1000)
	g2 := NewGame()
	g2.Restore(s, 1000+3600) // (une heure plus tard)
	if g2.Theft == nil || g2.Theft.Place != "ferme" || g2.Theft.Left != left {
		t.Fatalf("le vol n'a pas repris : %+v (il restait %v s)", g2.Theft, left)
	}
}
