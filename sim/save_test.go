package sim

import (
	"encoding/json"
	"testing"
)

func TestSaveRestore(t *testing.T) {
	g := NewGame()
	b := &Bot{}
	for g.T < 40*60 {
		b.Step(g, 1.0/30)
		g.Step(1.0 / 30)
		g.Events = g.Events[:0]
	}
	s := g.Snapshot(1000)
	js, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var s2 Save
	if err := json.Unmarshal(js, &s2); err != nil {
		t.Fatal(err)
	}
	h := NewGame()
	gain, secs := h.Restore(&s2, 1000+5*3600)
	t.Logf("argent %.0f, absence prise en compte %.0f s, gain %.0f, ouvriers %d", g.Money, secs, gain, len(h.Places[0].Workers))
	if h.Money < g.Money || secs != NightHours[g.Levels["mairie.veilleur"]]*3600 {
		t.Fatal("restauration incorrecte")
	}
	built := 0
	for _, st := range h.Places[0].Stations {
		if st.Built {
			built++
		}
	}
	built0 := 0
	for _, st := range g.Places[0].Stations {
		if st.Built {
			built0++
		}
	}
	if built != built0 {
		t.Fatalf("stations construites : %d au lieu de %d", built, built0)
	}
	// la partie restaurée continue de tourner
	for h.T < 120 {
		h.Step(1.0 / 30)
	}
}

func TestSaveSuper(t *testing.T) {
	g := NewGame()
	g.Money = 1e6
	g.BuyPlace(g.sup())
	g.Levels["super.caissiere"] = 1
	s := g.sup()
	s.complete(s.Unlocks[1])
	g.farm().Station("quai").Buf[Mais] = 5
	for g.T < 60 {
		g.Step(1.0 / 30)
	}
	sv := g.Snapshot(1000)
	h := NewGame()
	h.Restore(sv, 1000)
	if !h.sup().Owned || !h.sup().Station("rayon_1").Built || h.sup().Register().Cashier == nil {
		t.Fatal("supermarché mal restauré")
	}
}
