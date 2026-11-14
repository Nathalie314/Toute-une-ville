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
