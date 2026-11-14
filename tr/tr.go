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
	// débuts : « @début la = the » : un texte qui commence par « la » se traduit
	// par « the » suivi de la traduction du reste (les plus longs d'abord)
	starts []start
}

type start struct{ fr, out string }

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
		if fr, ok := strings.CutPrefix(k, "@début "); ok {
			l.starts = append(l.starts, start{fr: fr, out: v})
			sort.SliceStable(l.starts, func(i, j int) bool { return len(l.starts[i].fr) > len(l.starts[j].fr) })
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
	alt := flipFirst(s) // (« en main : … » se traduit aussi avec « En main : … »)
	for _, p := range l.pats {
		m := p.re.FindStringSubmatch(s)
		if m == nil && alt != "" {
			m = p.re.FindStringSubmatch(alt)
		}
		if m == nil {
			continue
		}
		out := p.out
		for i, name := range p.names {
			out = strings.ReplaceAll(out, name, l.tr(m[i+1], depth+1))
		}
		return sameCase(s, out), true
	}
	return "", false
}

// flipFirst : le texte avec sa première lettre dans l'autre casse ("" si ce n'est pas une lettre).
func flipFirst(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return ""
	}
	f := string(r[0])
	o := strings.ToLower(f)
	if o == f {
		o = strings.ToUpper(f)
	}
	if o == f {
		return ""
	}
	return o + string(r[1:])
}

// sameCase : la traduction commence par une majuscule si le texte d'origine aussi (et inversement).
func sameCase(src, v string) string {
	sr, vr := []rune(src), []rune(v)
	if len(sr) == 0 || len(vr) == 0 {
		return v
	}
	f := string(sr[0])
	switch {
	case strings.ToUpper(f) == f && strings.ToLower(f) != f:
		return strings.ToUpper(string(vr[0])) + string(vr[1:])
	case strings.ToLower(f) == f && strings.ToUpper(f) != f:
		return strings.ToLower(string(vr[0])) + string(vr[1:])
	}
	return v
}

// tr : traduction exacte, sinon par modèle, sinon le texte tel quel.
func (l *Language) tr(s string, depth int) string {
	if v, ok := l.dict[s]; ok {
		return v
	}
	if outs[s] {
		return s
	}
	if v, ok := l.derive(s, false); ok {
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

// derive : un texte absent du fichier, mais déduit d'une ligne voisine :
// la même avec une majuscule (ou sans), ou précédée d'un article (« @début »).
func (l *Language) derive(s string, _ bool) (string, bool) {
	if v, ok := l.withStart(s); ok {
		return v, true
	}
	r := []rune(s)
	if len(r) < 2 {
		return "", false
	}
	first := string(r[0])
	other := strings.ToLower(first)
	if other == first {
		other = strings.ToUpper(first)
	}
	if other == first {
		return "", false
	}
	alt := other + string(r[1:])
	v, ok := l.dict[alt]
	if !ok {
		v, ok = l.withStart(alt)
	}

	if !ok || v == "" {
		return "", false
	}
	vr := []rune(v)
	if strings.ToUpper(first) == first {
		return strings.ToUpper(string(vr[0])) + string(vr[1:]), true
	}
	return strings.ToLower(string(vr[0])) + string(vr[1:]), true
}

// sameCapital : une phrase qui commence par une majuscule garde sa majuscule
// (« Ouvrier du maïs » → « {zone} worker » → « Corn worker »).
func sameCapital(src, v string) string {
	sr, vr := []rune(src), []rune(v)
	if len(sr) == 0 || len(vr) == 0 || strings.ToUpper(string(sr[0])) != string(sr[0]) || strings.ToLower(string(sr[0])) == string(sr[0]) {
		return v
	}
	return strings.ToUpper(string(vr[0])) + string(vr[1:])
}

// withStart : « la crèmerie » → « the » + « crèmerie » traduit.
func (l *Language) withStart(s string) (string, bool) {
	for _, st := range l.starts {
		rest, ok := strings.CutPrefix(s, st.fr)
		if !ok || rest == "" {
			continue
		}
		if !strings.HasSuffix(st.fr, "'") {
			if rest[0] != ' ' {
				continue
			}
			rest = rest[1:]
		}
		v, ok := l.dict[rest]
		if !ok {
			v, ok = l.derive(rest, false)
		}
		if !ok {
			continue
		}
		if st.out == "" {
			return v, true
		}
		return st.out + " " + v, true
	}
	return "", false
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
	outs = map[string]bool{}
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
		markOut(v)
		if missing != nil {
			translated[v] = true
		}
		return v
	}
	if outs[s] {
		return s
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
	if numRe.MatchString(core) { // un nombre seul (« +20 % ») : à l'anglaise
		v := strings.Replace(s, core, englishNumber(core), 1)
		markOut(v)
		return v
	}
	if v, ok := cur.derive(s, false); ok {
		cur.memo[s] = v
		markOut(v)
		return v
	}
	if v, ok := cur.byPattern(s, 0); ok {
		markOut(v)
		v = sameCapital(s, v)
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
	if cur.dict == nil {
		return
	}
	markOut(s)
	if missing != nil {
		translated[strings.TrimSpace(s)] = true
	}
}

// outs : les textes déjà traduits (ou déjà dans la bonne langue) : on ne les
// retraduit pas (un modèle court comme « {h} h {m} » abîmerait une phrase anglaise).
var outs = map[string]bool{}

func markOut(v string) {
	if len(outs) > 20000 {
		outs = map[string]bool{}
	}
	outs[v] = true
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
