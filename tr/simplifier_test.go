package tr

import (
	"os"
	"strings"
	"testing"
)

// Outil : TR_SIMPLIFIER=../langues/en.txt go test ./tr -run Simplifier
// écrit dans TR_SIMPLIFIER.inutiles les lignes qui se déduisent d'autres lignes
// (même traduction, à la majuscule près, sans elles).
func TestSimplifier(t *testing.T) {
	f := os.Getenv("TR_SIMPLIFIER")
	if f == "" {
		t.Skip()
	}
	b, err := os.ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}
	Register("simpl", string(b))
	l := langs["simpl"]
	var keys []string
	for k := range l.dict {
		if !strings.Contains(k, "{") {
			keys = append(keys, k)
		}
	}
	// les textes longs d'abord : on garde de préférence les lignes courtes (« crèmerie »)
	sortByLen(keys)
	var out []string
	for _, k := range keys {
		v := l.dict[k]
		delete(l.dict, k)
		if got := sameCapital(k, l.tr(k, 0)); strings.EqualFold(got, v) && got != k {
			out = append(out, k)
			continue
		}
		l.dict[k] = v
	}
	_ = os.WriteFile(f+".inutiles", []byte(strings.Join(out, "\n")+"\n"), 0o644)
	t.Log(len(out), "lignes déductibles")
}

func sortByLen(ks []string) {
	for i := 1; i < len(ks); i++ {
		for j := i; j > 0 && (len(ks[j]) > len(ks[j-1]) || len(ks[j]) == len(ks[j-1]) && ks[j] < ks[j-1]); j-- {
			ks[j], ks[j-1] = ks[j-1], ks[j]
		}
	}
}
