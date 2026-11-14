package tr

import "testing"

func TestTraduction(t *testing.T) {
	Register("xx", "@nom = Test\n# commentaire\nBonjour = Hello\nIl te faut {prix}. = You need {prix}.\n1 objet \\= 3 = 1 item \\= 3\n")
	Set("xx")
	defer Set("fr")
	if T("Bonjour") != "Hello" || T(" Bonjour ") != " Hello " || T("Inconnu") != "Inconnu" {
		t.Fatal("T")
	}
	if F("Il te faut {prix}.", "prix", "$5") != "You need $5." {
		t.Fatal("F", F("Il te faut {prix}.", "prix", "$5"))
	}
	if T("1 objet = 3") != "1 item = 3" {
		t.Fatal("égal", T("1 objet = 3"))
	}
	Set("fr")
	if T("Bonjour") != "Bonjour" {
		t.Fatal("fr")
	}
}

func TestModeles(t *testing.T) {
	Register("yy", "Étal : {produit} = Stand: {produit}\nMaïs = Corn\n{nom} : {desc} = {nom}: {desc}\nDrone acheté pour l'étal « {p} » ! = Drone bought for the “{p}” stand!\n")
	Set("yy")
	defer Set("fr")
	if v := T("Étal : Maïs"); v != "Stand: Corn" {
		t.Fatal(v)
	}
	if v := T("Drone acheté pour l'étal « Étal : Maïs » !"); v != "Drone bought for the “Stand: Corn” stand!" {
		t.Fatal(v)
	}
	if v := T("Truc : Maïs"); v != "Truc: Corn" {
		t.Fatal(v)
	}
}

func TestNombres(t *testing.T) {
	Register("zz", "Servir {n} clients : {lieu}. = Serve {n} customers: {lieu}.\nport = harbor\nNiveau {n} : {p} plus rapide. = Level {n}: {p} faster.\n")
	Set("zz")
	defer Set("fr")
	if v := T("Servir 1 000 clients : port."); v != "Serve 1,000 customers: harbor." {
		t.Fatal(v)
	}
	if v := T("Niveau 2 : 20 % plus rapide."); v != "Level 2: 20% faster." {
		t.Fatal(v)
	}
}
