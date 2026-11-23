package sim

import "testing"

// TestIncendie : zones chaudes, prévention, ce que sauvent les camions, le lieu remis à zéro.
func TestIncendie(t *testing.T) {
	g := NewGame()
	b := g.Place("boulangerie")
	b.Owned = true
	g.Enter(b)
	g.Step(1.0 / 30)
	if !g.Place("caserne").Owned {
		t.Fatal("la caserne ne s'ouvre pas avec la boulangerie")
	}
	for _, id := range []string{"pain_mie", "brioches", "croissants"} {
		for _, u := range b.Unlocks {
			if u.ID == id {
				b.complete(u)
			}
		}
	}
	if n := g.HotZones("boulangerie"); n != 4 {
		t.Fatalf("zones chaudes : %d (4 fours)", n)
	}
	lv := g.Levels
	lv["boulangerie.caissiere"], lv["boulangerie.ouvriers"], lv["boulangerie.etals"] = 1, 2, 3
	b.Station("croissants").Pile = 5
	// la prévention
	lv["pts.ppi"] = 2000
	if !g.SpendPPI("boulangerie", 2000) || g.PPIOf("boulangerie") != PPIMax || g.FireRisk("boulangerie") != 0.1 {
		t.Fatalf("PPI : %d, risque %.2f", g.PPIOf("boulangerie"), g.FireRisk("boulangerie"))
	}
	// deux camions (l'un amélioré) : 3 éléments sauvés : la caissière, les deux ouvriers
	lv["caserne.commandant"], lv["caserne.camion_1"], lv["caserne.pompier_1"], lv["caserne.camion_2"], lv["caserne.pompier_2"] = 1, 1, 1, 1, 1
	lv["pts.ouvrier"] = TruckUpCost
	if !g.BuyTruckUp(1) || g.SaveCapacity() != 3 {
		t.Fatalf("capacité des camions : %d", g.SaveCapacity())
	}
	g.startFire("boulangerie")
	for g.Theft != nil {
		g.Step(1.0 / 30)
	}
	if lv["boulangerie.caissiere"] != 1 || lv["boulangerie.ouvriers"] != 2 || lv["boulangerie.etals"] != 0 {
		t.Fatalf("après le feu : caissière %d, ouvriers %d, étals %d", lv["boulangerie.caissiere"], lv["boulangerie.ouvriers"], lv["boulangerie.etals"])
	}
	if b.Station("croissants").Built || b.Station("croissants").Pile != 0 || !b.Station("four_baguette").Built || g.HotZones("boulangerie") != 1 {
		t.Fatal("les fours achetés devaient brûler (le four de départ reste)")
	}
	// un camion de plus : il sauve le four le plus cher
	for _, u := range b.Unlocks {
		if u.ID == "brioches" || u.ID == "pain_mie" {
			b.complete(u)
		}
	}
	lv["caserne.camion_3"], lv["caserne.pompier_3"] = 1, 1
	g.startFire("boulangerie")
	s := g.Snapshot(1000) // (on quitte pendant l'incendie : il finit jeu fermé)
	h := NewGame()
	h.Restore(s, 1000+600)
	hb := h.Place("boulangerie")
	if h.Theft != nil || !hb.Station("brioches").Built || hb.Station("pain_mie").Built {
		t.Fatalf("incendie fini jeu fermé : il fallait sauver les brioches (le plus cher) et perdre le pain de mie")
	}
}
