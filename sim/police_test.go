package sim

import (
	"math"
	"testing"
)

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
	g.Player.Pos = pol.spot("lieutenant_1").Zone.C()
	for i := 0; i < 60; i++ {
		g.Step(1.0 / 30)
	}
	if g.PolLieutenants() != 1 || pol.Pol.Cops[0] == nil {
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

// TestVideosurveillance : bureaux (prix des lieux 3, 6, 7, 10…), capitaine (lycée), arrestation à la
// sortie, garde à vue (2 h, même jeu fermé), moins de voleurs en ville pendant ce temps.
func TestVideosurveillance(t *testing.T) {
	if r := []int{survRank(1), survRank(2), survRank(3), survRank(4), survRank(5), survRank(6), survRank(7)}; r[0] != 3 || r[1] != 6 || r[2] != 7 || r[3] != 10 || r[4] != 11 || r[5] != 14 || r[6] != 15 {
		t.Fatalf("rangs des bureaux : %v", r)
	}
	if PolOffices() != 6 {
		t.Fatalf("%d bureaux (14 lieux : 6)", PolOffices())
	}
	g := NewGame()
	g.Place("super").Owned = true
	g.Step(1.0 / 30)
	pol := g.Place("commissariat")
	if c := pol.SpotUpgrade(pol.spot("bureau_2")).Cost(0); c != PlacePrice(6) {
		t.Fatalf("bureau n° 2 : %.0f (lieu 6 : %.0f)", c, PlacePrice(6))
	}
	if c := pol.SpotUpgrade(pol.spot("bureau_1")).Cost(0); c != g.Place("cremerie").Price || c == 0 {
		t.Fatalf("bureau n° 1 : %.0f (crèmerie : %.0f)", c, g.Place("cremerie").Price)
	}
	if c := pol.SpotUpgrade(pol.spot("cellule_1")).Cost(0); c != math.Round(g.Place("cremerie").Price/2) {
		t.Fatalf("cellule n° 1 : %.0f", c)
	}
	lv := g.Levels
	lv["commissariat.commissaire"], lv["commissariat.bureau_1"], lv["commissariat.capitaine_1"] = 1, 1, 1
	g.SetPatrol(1, 0, "super")
	if g.SetCamera(1, "super"); g.PolCamera(1) != "" {
		t.Fatal("un capitaine sur un lieu déjà confié à une voiture")
	}
	g.SetCamera(1, "ferme")
	if g.SetPatrol(1, 0, "ferme"); g.PolCamera(1) != "" {
		t.Fatal("la voiture arrive : le capitaine aurait dû être libéré")
	}
	g.SetPatrol(1, 0, "super")
	g.SetCamera(1, "ferme")
	lv["pts.ouvrier"] = 2000
	if !g.BuyVideo(1, 2000) || g.CatchChance(1) != 1 || g.Points("ouvrier") != 2000-VideoMax {
		t.Fatalf("techniciens : chance %.2f, points %d", g.CatchChance(1), g.Points("ouvrier"))
	}
	avant := g.Thieves()
	// sans cellule : repérés, mais libres
	g.StartTheft("ferme")
	if !g.Theft.Spotted {
		t.Fatal("voleurs non repérés à 100 %")
	}
	for g.Theft != nil {
		g.Step(1.0 / 30)
	}
	if g.InCustody() != 0 {
		t.Fatal("arrêtés sans cellule")
	}
	// une cellule : 5 voleurs en garde à vue
	lv["commissariat.cellule_1"] = 1
	g.StartTheft("ferme")
	for g.Theft != nil {
		g.Step(1.0 / 30)
	}
	if g.InCustody() != 5 || g.Thieves() != avant-5 {
		t.Fatalf("garde à vue : %d, voleurs %d (avant %d)", g.InCustody(), g.Thieves(), avant)
	}
	// dix minutes plus tard (jeu fermé) : toujours là ; trois heures : relâchés
	s := g.Snapshot(1000)
	h := NewGame()
	h.Restore(s, 1000+600)
	if h.InCustody() != 5 {
		t.Fatalf("après dix minutes : %d en garde à vue", h.InCustody())
	}
	h2 := NewGame()
	h2.Restore(s, 1000+3*3600)
	if h2.InCustody() != 0 {
		t.Fatalf("après trois heures : %d en garde à vue", h2.InCustody())
	}
}
