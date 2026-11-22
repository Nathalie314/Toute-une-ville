package sim

import "testing"

// TestCommissariat : il s'ouvre avec le supermarché ; commissaire (université), voiture offerte,
// lieutenant (collège), patrouille : le lieu protégé n'a plus de voleurs.
func TestCommissariat(t *testing.T) {
	g := NewGame()
	pol := g.Place("commissariat")
	g.Step(1.0 / 30)
	if pol.Owned {
		t.Fatal("le commissariat s'ouvre avant le supermarché")
	}
	g.Place("super").Owned = true
	g.Step(1.0 / 30)
	if !pol.Owned {
		t.Fatal("le commissariat ne s'ouvre pas avec le supermarché")
	}
	if PriceRank("commissariat") != 0 || len(TopRow()) != 14 || PolCars() != 4 {
		t.Fatalf("rangées de la carte : rang %d, %d lieux en haut, %d voitures", PriceRank("commissariat"), len(TopRow()), PolCars())
	}
	// le commissaire : 10 points d'université, sur sa case
	g.Enter(pol)
	sp := pol.spot("commissaire")
	g.Levels["pts.formation"] = 9
	g.Player.Pos = sp.Zone.C()
	for i := 0; i < 60; i++ {
		g.Step(1.0 / 30)
	}
	if g.HasChief() {
		t.Fatal("commissaire acheté avec 9 points")
	}
	g.Levels["pts.formation"] = 12
	for i := 0; i < 60; i++ {
		g.Step(1.0 / 30)
	}
	if !g.HasChief() || g.Points("formation") != 2 || !g.PolCarOwned(1) {
		t.Fatalf("commissaire : %v, points : %d, voiture offerte : %v", g.HasChief(), g.Points("formation"), g.PolCarOwned(1))
	}
	// un lieutenant (10 points du collège), puis la patrouille à la ferme
	g.Levels["pts.marketing"] = 10
	g.Player.Pos = pol.spot("lieutenants").Zone.C()
	for i := 0; i < 60; i++ {
		g.Step(1.0 / 30)
	}
	if g.PolLieutenants() != 1 || len(pol.Pol.Cops) != 1 {
		t.Fatalf("lieutenants : %d", g.PolLieutenants())
	}
	g.SetPatrol(1, 0, "ferme")
	if !g.Patrolled("ferme") {
		t.Fatal("la ferme n'est pas protégée")
	}
	for _, e := range g.eventGroups("ferme") {
		if e.vol {
			t.Fatal("des voleurs possibles à la ferme protégée")
		}
	}
	// une seule patrouille par lieu ; deux lieux avec les mécaniciens
	g.SetPatrol(1, 0, "super")
	if g.Patrolled("ferme") || !g.Patrolled("super") {
		t.Fatal("la voiture n'a pas changé de lieu")
	}
	if len(g.eventGroups("super")) != 0 {
		t.Fatal("au supermarché protégé, rien ne devrait pouvoir arriver (il n'a que des voleurs)")
	}
	g.Levels["pts.ouvrier"] = DoubleCost(1)
	if !g.BuyDouble(1) || g.Points("ouvrier") != 0 {
		t.Fatal("mécaniciens non achetés")
	}
	g.SetPatrol(1, 1, "ferme")
	if !g.Patrolled("ferme") || !g.Patrolled("super") {
		t.Fatal("la voiture ne protège pas deux lieux")
	}
	// sauvegarde
	s := g.Snapshot(1000)
	h := NewGame()
	h.Restore(s, 1000)
	if !h.Place("commissariat").Owned || !h.Patrolled("ferme") || !h.Patrolled("super") || h.PolLieutenants() != 1 {
		t.Fatal("commissariat mal restauré")
	}
	// les prix des voitures : le 4e lieu, le 8e lieu
	if c := pol.SpotUpgrade(pol.spot("voiture_2")).Cost(0); c != PlacePrice(4) {
		t.Fatalf("2e voiture : %.0f (4e lieu : %.0f)", c, PlacePrice(4))
	}
	if c := pol.SpotUpgrade(pol.spot("voiture_3")).Cost(0); c != PlacePrice(8) {
		t.Fatalf("3e voiture : %.0f (8e lieu : %.0f)", c, PlacePrice(8))
	}
	if c := pol.SpotUpgrade(pol.spot("voiture_4")).Cost(0); c != PlacePrice(12) {
		t.Fatalf("4e voiture : %.0f (12e lieu : %.0f)", c, PlacePrice(12))
	}
}
