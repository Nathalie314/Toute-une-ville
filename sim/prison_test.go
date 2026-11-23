package sim

import "testing"

// TestPrison : elle s'ouvre avec la première garde à vue ; directeur, surveillant, cellule,
// fourgon : le voleur arrive en 2 minutes ; il dort à la 10e minute puis rate tous ses autres
// besoins (170 points : récidiviste) ; avec tout au niveau 10, il sort à 20 points.
func TestPrison(t *testing.T) {
	g := NewGame()
	g.Place("super").Owned = true
	g.Step(1.0 / 30)
	pr := g.Place("prison")
	if pr.Owned {
		t.Fatal("la prison s'ouvre trop tôt")
	}
	lv := g.Levels
	lv["commissariat.commissaire"], lv["commissariat.cellule_1"] = 1, 1
	g.arrest("ferme", 5, CustodyTime)
	g.Step(1.0 / 30)
	if !pr.Owned {
		t.Fatal("la prison ne s'ouvre pas avec la première garde à vue")
	}
	if pr.SpotVisible(pr.spot("cellule_1")) {
		t.Fatal("une cellule avant le directeur")
	}
	lv["prison.directeur"], lv["prison.surveillant_1"], lv["prison.cellule_1"], lv["prison.fourgon_1"] = 1, 1, 1, 1
	if !pr.SpotVisible(pr.spot("cellule_2")) || pr.SpotVisible(pr.spot("cellule_3")) || !pr.SpotVisible(pr.spot("surveillant_5")) || !pr.SpotVisible(pr.spot("famille_4")) {
		t.Fatal("cases : tout s'achète dans l'ordre qu'on veut, sauf une cellule sans son surveillant")
	}
	if c := pr.SpotUpgrade(pr.spot("cellule_7")).Cost(0); c != prisMoney(8e6, 1.25, 7) {
		t.Fatalf("cellule n° 7 : %.0f (prix fixe, selon son numéro)", c)
	}
	// la rareté : une seule table (2 places) et un sport à 3 places : manger passe avant le sport
	lv["prison.cantine"], lv["prison.table_3"], lv["prison.sport"], lv["prison.agres_6"], lv["prison.agres_1"] = 1, 1, 1, 1, 1
	o := pr.needOrder(&Prisoner{Cell: 1})
	ir, is := 0, 0
	for i, n := range o {
		if n == NeedRepas {
			ir = i
		}
		if n == NeedSport {
			is = i
		}
	}
	if o[0] != NeedPsy || o[1] != NeedAS || o[len(o)-1] != NeedDormir || ir > is {
		t.Fatalf("ordre des besoins : %v", o)
	}
	for _, k := range []string{"cantine", "table_3", "sport", "agres_6", "agres_1"} {
		delete(lv, "prison."+k)
	}
	avant := g.Thieves()
	steps := func(s float64) {
		for i := 0; i < int(s*4); i++ { // (la prison seule, à grands pas : le test reste rapide)
			pr.stepPrison(0.25, false)
		}
	}
	steps(1)
	if g.InCustody() != 4 || g.Thieves() != avant {
		t.Fatalf("le fourgon emmène un voleur : garde à vue %d, voleurs %d (avant %d)", g.InCustody(), g.Thieves(), avant)
	}
	steps(PrisVanTrip)
	if len(pr.Pris.Pris) != 1 {
		t.Fatalf("%d prisonnier(s) après le voyage du fourgon", len(pr.Pris.Pris))
	}
	d := pr.Pris.Pris[0]
	steps(PrisNeedEvery + 2)
	if !d.Done[NeedDormir] || d.Recid != PrisRecid {
		t.Fatalf("10e minute : il devait dormir (récidive %d)", d.Recid)
	}
	steps(PrisStay - PrisNeedEvery)
	if len(pr.Pris.Pris) != 0 || lv["prison.recidivistes"] != 1 || d.Recid != PrisRecid+7*PrisFail {
		t.Fatalf("sortie : %d détenus, récidivistes %d, points %d", len(pr.Pris.Pris), lv["prison.recidivistes"], d.Recid)
	}
	// tout au maximum : il sort à 20 points (80 % de chances d'être réinséré)
	for _, k := range []string{"cantine", "table_1", "bains", "cabine_1", "sport", "agres_1", "famille_1", "travail", "poste_1", "psy", "as"} {
		lv["prison."+k] = 1
	}
	for _, k := range []string{"niv_cellule_1", "niv_cantine", "niv_cabine_1", "niv_sport", "niv_famille_1", "niv_travail"} {
		lv["prison."+k] = PrisLvlMax
	}
	g.Custody, pr.Pris.Pris, pr.Pris.Vans = nil, nil, [PrisVanN]Van{} // (un seul prisonnier, celui du test)
	pr.Pris.Pris = append(pr.Pris.Pris, &Prisoner{Pos: PrisCell(1).C(), Cell: 1, Recid: PrisRecid, Need: -1})
	d = pr.Pris.Pris[0]
	steps(PrisStay - 60)
	if d.Recid != 20 {
		t.Fatalf("tout au niveau 10 : %d points (20 attendus)", d.Recid)
	}
	// la sauvegarde (rien ne tourne jeu fermé)
	s := g.Snapshot(1000)
	h := NewGame()
	h.Restore(s, 1000+7200)
	if hp := h.Place("prison"); !hp.Owned || len(hp.Pris.Pris) != 1 || hp.Pris.Pris[0].Recid != 20 || hp.Pris.Pris[0].T != d.T {
		t.Fatal("prison mal restaurée")
	}
}
