package sim

import "testing"

// TestCommandesPolicePrison : les menus selon les rangs, et les cadeaux.
func TestCommandesPolicePrison(t *testing.T) {
	g := NewGame()
	lv := g.Levels
	if g.policeOrder("boulangerie") != nil || g.prisonOrder("boulangerie") != nil {
		t.Fatal("des commandes sans commissaire ni directeur")
	}
	lv["commissariat.commissaire"], lv["commissariat.lieutenant_1"], lv["commissariat.lieutenant_2"] = 1, 1, 1
	lv["commissariat.bureau_1"], lv["commissariat.capitaine_1"] = 1, 1
	g.Custody = []*Custody{{N: 3, Left: 100}}
	want := map[Item]int{Brioche: 1, PainChoco: 1, Croissant: 2, Pain: 3}
	o := g.policeOrder("boulangerie")
	for _, l := range o.Lines {
		if want[l.It] != l.N {
			t.Fatalf("boulangerie du commissariat : %v", o.Lines)
		}
	}
	if o := g.policeOrder("exotique"); len(o.Lines) != 1 || o.Lines[0].It != CafeMoulu || o.Lines[0].N != 4 {
		t.Fatalf("café : %v", o.Lines)
	}
	// le cadeau : les policiers 2 fois moins chers en points ; raté : 2 fois plus
	pol := g.Place("commissariat")
	u := pol.SpotUpgrade(pol.spot("lieutenant_3"))
	base := u.Cost(0)
	g.Boost = &Boost{Kind: "police", Left: 100, Total: 100}
	if u.Cost(0) != base/2 {
		t.Fatalf("cadeau du commissariat : %.0f (sans : %.0f)", u.Cost(0), base)
	}
	g.Boost = &Boost{Kind: "police_malus", Left: 100, Total: 100}
	if u.Cost(0) != base*2 {
		t.Fatalf("pénalité du commissariat : %.0f", u.Cost(0))
	}
	g.Boost = nil
	// la prison : le directeur à part ; la boucherie seulement avec la cantine
	lv["prison.directeur"], lv["prison.surveillant_1"], lv["prison.psy"] = 1, 1, 1
	g.Place("prison").Pris.Pris = []*Prisoner{{Cell: 1}, {Cell: 2}}
	if g.prisonOrder("boucherie") != nil {
		t.Fatal("la boucherie sans cantine")
	}
	lv["prison.cantine"] = 1
	want = map[Item]int{SteakHache: 2, RotiBoeuf: 2, FoieGras: 1}
	for _, l := range g.prisonOrder("boucherie").Lines {
		if want[l.It] != l.N {
			t.Fatalf("boucherie de la prison : %v", g.prisonOrder("boucherie").Lines)
		}
	}
	if giftOf(whoPolice) != "police" || giftOf(whoPrison) != "prison" {
		t.Fatal("cadeaux")
	}
	g.Boost = &Boost{Kind: "prison", Left: 100, Total: 100}
	if g.reinsertMult(3) != 6 {
		t.Fatal("cadeau de la prison")
	}
}
