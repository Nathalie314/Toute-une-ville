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
	"regexp"
	"sort"
	"strings"
)

// Language : une langue disponible.
type Language struct {
	Code string // « en » (nom du fichier sans .txt)
	Name string // « English » (ligne @nom)
	dict map[string]string
	pats []pattern         // les textes à trous, essayés comme modèles
	memo map[string]string // textes déjà traduits par un modèle
}

// pattern : un texte à trous (« Étal : {produit} ») vu comme un modèle : il
// traduit aussi « Étal : Maïs » (le morceau « Maïs » est traduit à son tour).
type pattern struct {
	re    *regexp.Regexp
	names []string
	out   string
	lit   int // nombre de lettres fixes (les modèles les plus précis d'abord)
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
	l.compile()
}

var numRe = regexp.MustCompile(`^[+−-]?[0-9][0-9 \x{202F}\x{00A0}]*(,[0-9]+)?( ?%)?$`)

// englishNumber : « 1 000 » → « 1,000 », « 2,5 » → « 2.5 », « 20 % » → « 20% ».
func englishNumber(s string) string {
	s = strings.NewReplacer(",", ".", " %", "%", "\u202f%", "%", "\u00a0%", "%").Replace(s)
	return strings.NewReplacer(" ", ",", "\u202f", ",", "\u00a0", ",").Replace(s)
}

var holeRe = regexp.MustCompile(`\{[^{}]+\}`)

// compile prépare les modèles (les textes qui contiennent des {trous}).
func (l *Language) compile() {
	l.pats, l.memo = nil, map[string]string{}
	for k, v := range l.dict {
		if !strings.Contains(k, "{") {
			continue
		}
		var names []string
		expr := "^"
		last := 0
		for _, m := range holeRe.FindAllStringIndex(k, -1) {
			expr += regexp.QuoteMeta(k[last:m[0]]) + "(.+?)"
			names = append(names, k[m[0]:m[1]])
			last = m[1]
		}
		expr += regexp.QuoteMeta(k[last:]) + "$"
		lit := len([]rune(holeRe.ReplaceAllString(k, "")))
		if len(names) == 0 || lit < 2 {
			continue
		}
		re, err := regexp.Compile(expr)
		if err != nil {
			continue
		}
		l.pats = append(l.pats, pattern{re: re, names: names, out: v, lit: lit})
	}
	sort.Slice(l.pats, func(i, j int) bool {
		if l.pats[i].lit != l.pats[j].lit {
			return l.pats[i].lit > l.pats[j].lit
		}
		return l.pats[i].re.String() < l.pats[j].re.String()
	})
}

// byPattern traduit s avec le premier modèle qui lui correspond.
func (l *Language) byPattern(s string, depth int) (string, bool) {
	if depth > 6 {
		return "", false
	}
	for _, p := range l.pats {
		m := p.re.FindStringSubmatch(s)
		if m == nil {
			continue
		}
		out := p.out
		for i, name := range p.names {
			out = strings.ReplaceAll(out, name, l.tr(m[i+1], depth+1))
		}
		return out, true
	}
	return "", false
}

// tr : traduction exacte, sinon par modèle, sinon le texte tel quel.
func (l *Language) tr(s string, depth int) string {
	if v, ok := l.dict[s]; ok {
		return v
	}
	if numRe.MatchString(s) { // un nombre à la française (1 000 ; 2,5 ; 20 %) : à l'anglaise
		return englishNumber(s)
	}
	if v, ok := l.byPattern(s, depth); ok {
		return v
	}
	if missing != nil && hasLetters(s) && !translated[strings.TrimSpace(s)] {
		missing[strings.TrimSpace(s)] = true // (un morceau sans traduction)
	}
	return s
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
	if missing != nil && translated[core] { // (outil : un texte déjà traduit)
		return s
	}
	if v, ok := cur.memo[s]; ok {
		return v
	}
	if v, ok := cur.byPattern(s, 0); ok {
		if len(cur.memo) > 20000 { // (les textes avec des nombres changent sans cesse)
			cur.memo = map[string]string{}
		}
		cur.memo[s] = v
		if missing != nil {
			translated[v] = true
		}
		return v
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
	Seen(s)
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
