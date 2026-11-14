package main

import (
	"embed"
	"path"
	"strings"

	"ville/tr"
)

// Les langues du jeu : un fichier texte par langue dans le dossier langues/
// (voir tr/tr.go et langues/LISEZMOI.txt). Ils sont intégrés au jeu ; sous
// Windows, un dossier « langues » posé à côté de l'exe est lu en plus (pour
// essayer une traduction sans recompiler).

//go:embed langues/*.txt
var langFiles embed.FS

func init() {
	entries, _ := langFiles.ReadDir("langues")
	for _, e := range entries {
		addLanguageFile(e.Name(), func() ([]byte, error) { return langFiles.ReadFile("langues/" + e.Name()) })
	}
	for name, b := range extraLanguageFiles() {
		addLanguageFile(name, func() ([]byte, error) { return b, nil })
	}
}

func addLanguageFile(name string, read func() ([]byte, error)) {
	code := strings.TrimSuffix(path.Base(name), ".txt")
	if code == name || code == "fr" || strings.EqualFold(code, "lisezmoi") || strings.EqualFold(code, "readme") {
		return
	}
	if b, err := read(); err == nil {
		tr.Register(code, string(b))
	}
}
