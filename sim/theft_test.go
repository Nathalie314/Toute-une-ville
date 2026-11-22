package sim

import (
	"math"
	"testing"
)

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

// TestVolSauvegarde : quitter le jeu ne chasse pas les voleurs, mais leur temps s'écoule jeu fermé.
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
	g2.Restore(s, 1000+10) // (dix secondes plus tard : ils sont toujours là)
	if g2.Theft == nil || g2.Theft.Place != "ferme" || math.Abs(g2.Theft.Left-(left-10)) > 0.01 {
		t.Fatalf("le vol n'a pas repris : %+v (il restait %v s)", g2.Theft, left)
	}
	s = g.Snapshot(1000)
	g3 := NewGame()
	g3.Restore(s, 1000+3600) // (une heure plus tard : ils sont partis)
	if g3.Theft != nil {
		t.Fatalf("les voleurs sont restés une heure : %+v", g3.Theft)
	}
}

// TestVolHorsLigneRetour : absent plus longtemps que le vol, le lieu rapporte après leur départ.
func TestVolHorsLigneRetour(t *testing.T) {
	gain := func(vol bool) float64 {
		g := NewGame()
		g.Money = 1e6
		g.BuyPlace(g.sup())
		g.Levels["super.caissiere"] = 1
		for g.T < 600 {
			g.Step(1.0 / 30)
		}
		g.earn(30000, V{}, "ferme")
		g.earn(30000, V{}, "super")
		if vol {
			g.StartTheft("ferme")
		}
		s := g.Snapshot(1000)
		off, _ := NewGame().Restore(s, 1000+7200)
		return off
	}
	sans, avec := gain(false), gain(true)
	if avec >= sans || avec < sans*0.95 {
		t.Fatalf("deux heures d'absence avec un vol de %d min : %.0f (sans vol : %.0f)", 2, avec, sans)
	}
}

// TestVolHorsLigne : pendant un vol, le lieu attaqué ne rapporte rien hors ligne.
func TestVolHorsLigne(t *testing.T) {
	g := NewGame()
	g.Place("super").Owned = true
	g.Enter(g.Place("ferme"))
	g.T = 600
	g.earn(30000, V{}, "ferme") // (la ferme a gagné autant que le supermarché ces dernières minutes)
	g.earn(30000, V{}, "super")
	avant := g.RateOffline()
	g.StartTheft("ferme")
	if apres := g.RateOffline(); avant <= 0 || apres > avant*0.51 {
		t.Fatalf("gains hors ligne pendant un vol à la ferme : %.2f €/s (avant : %.2f)", apres, avant)
	}
}

// TestVolPourboires : à la crêperie, les voleurs raflent les pourboires des tables.
func TestVolPourboires(t *testing.T) {
	g := NewGame()
	g.Place("super").Owned = true
	p := g.Place("creperie")
	p.Owned = true
	g.Enter(p)
	if len(p.Crep.Tables) == 0 {
		t.Skip("pas de table")
	}
	for _, tb := range p.Crep.Tables {
		if tb.St != nil {
			tb.St.Built = true
		}
		tb.Tips = 10
	}
	g.StartTheft("creperie")
	for i := 0; i < 30*120 && g.Theft != nil; i++ {
		g.Step(1.0 / 30)
	}
	reste := 0.0
	for _, tb := range p.Crep.Tables {
		reste += tb.Tips
	}
	if reste > 0 {
		t.Fatalf("il reste %.0f de pourboires sur les tables", reste)
	}
}
