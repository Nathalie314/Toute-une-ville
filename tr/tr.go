// Package tr traduit les textes du jeu.
//
// Le français est la langue d'origine : chaque texte du jeu est écrit en
// français dans le code, et sert de « clé ». Une autre langue est un simple
// fichier texte (dossier langues/ à la racine du projet, par exemple en.txt) :
//
//	@nom = English
//	# commentaire
//	Champ de maïs = Corn field
//	Il te faut {prix}. = You need {prix}.
//
// À gauche du premier « = » : le texte français exact ; à droite : sa
// traduction. Les {mots} entre accolades sont remplacés par le jeu (un prix,
// un nom…) et doivent être recopiés tels quels. « \n » : passage à la ligne.
// Un signe « = » à l'intérieur d'un texte s'écrit « \= ».
// Un texte absent du fichier reste en français.
package tr

import (
	"sort"
	"strings"
)

// Language : une langue disponible.
type Language struct {
	Code string // « en » (nom du fichier sans .txt)
	Name string // « English » (ligne @nom)
	dict map[string]string
}

var (
	langs = map[string]*Language{"fr": {Code: "fr", Name: "Français"}}
	cur   = langs["fr"]
)

// Register ajoute (ou complète) une langue à partir du contenu de son fichier.
// Les lignes d'un second fichier de la même langue remplacent celles du premier.
func Register(code, content string) {
	l := langs[code]
	if l == nil {
		l = &Language{Code: code, Name: code}
		langs[code] = l
	}
	if l.dict == nil {
		l.dict = map[string]string{}
	}
	for _, line := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "\uFEFF"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// (un « = » dans un texte s'écrit « \= », pour ne pas le confondre avec le séparateur)
		k, v, ok := strings.Cut(strings.ReplaceAll(line, `\=`, "\x00"), " = ")
		if !ok {
			continue
		}
		k, v = strings.ReplaceAll(k, "\x00", "="), strings.ReplaceAll(v, "\x00", "=")
		k, v = unescape(strings.TrimSpace(k)), unescape(strings.TrimSpace(v))
		if k == "@nom" {
			l.Name = v
			continue
		}
		if v != "" {
			l.dict[k] = v
		}
	}
}

func unescape(s string) string { return strings.ReplaceAll(s, `\n`, "\n") }

// Languages : les langues disponibles, le français d'abord.
func Languages() []*Language {
	var out []*Language
	for _, l := range langs {
		out = append(out, l)
	}
	sort.Slice(out, func(i, j int) bool {
		if (out[i].Code == "fr") != (out[j].Code == "fr") {
			return out[i].Code == "fr"
		}
		return out[i].Name < out[j].Name
	})
	return out
}

// Set choisit la langue (code inconnu : français).
func Set(code string) {
	if l := langs[code]; l != nil {
		cur = l
	} else {
		cur = langs["fr"]
	}
}

// Lang : le code de la langue choisie (« fr », « en »…).
func Lang() string { return cur.Code }

// English : vrai si les nombres s'écrivent à l'anglaise ($1,234.5).
func English() bool { return cur.Code != "fr" }

// T traduit un texte. Les espaces au début et à la fin sont gardés
// (« 3 » + T(" récoltes") …), le reste doit correspondre exactement.
func T(s string) string {
	if cur.dict == nil || s == "" {
		return s
	}
	if v, ok := cur.dict[s]; ok {
		if missing != nil {
			translated[v] = true
		}
		return v
	}
	core := strings.TrimSpace(s)
	if core != s && core != "" {
		if v, ok := cur.dict[core]; ok {
			i := strings.Index(s, core)
			return s[:i] + v + s[i+len(core):]
		}
	}
	if missing != nil && hasLetters(core) && !translated[core] {
		missing[core] = true
	}
	return s
}

// Seen : ce texte est déjà passé par T (ligne coupée par le jeu, traduction) :
// l'outil ne le note pas comme manquant.
func Seen(s string) {
	if missing != nil {
		translated[strings.TrimSpace(s)] = true
	}
}

// translated : les traductions déjà produites (pour ne pas les noter comme manquantes).
var translated = map[string]bool{}

func hasLetters(s string) bool {
	n := 0
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r > 0xbf && r < 0x250 {
			n++
		}
	}
	return n >= 2
}

// F traduit un texte à trous puis remplit les trous : F("Il te faut {prix}.", "prix", "50 €").
func F(s string, kv ...string) string {
	s = T(s)
	for i := 0; i+1 < len(kv); i += 2 {
		s = strings.ReplaceAll(s, "{"+kv[i]+"}", kv[i+1])
	}
	return s
}

// Has : vrai si ce texte a une traduction dans la langue choisie.
func Has(s string) bool {
	_, ok := cur.dict[s]
	return ok
}

// ---------------------------------------------------------------- outil de traduction

var missing map[string]bool

// Track : note les textes affichés qui n'ont pas de traduction (pour les traducteurs).
func Track() { missing = map[string]bool{} }

// Missing : les textes notés par Track, triés.
func Missing() []string {
	var out []string
	for k := range missing {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
