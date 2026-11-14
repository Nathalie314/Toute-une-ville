//go:build !js

package main

import (
	"os"
	"path/filepath"
)

// La partie est enregistrée à côté du jeu (ToutUneVille_sauvegarde.json, dans
// le même dossier que l'exe) : on peut emporter le jeu sur une clé USB avec sa
// partie. Si ce dossier n'est pas modifiable, on se rabat sur
// %AppData%\ToutUneVille\sauvegarde.json.

// isWeb : le jeu tourne-t-il dans un navigateur ?
const isWeb = false

// appDataPath : l'ancien emplacement (et le secours).
func appDataPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "ToutUneVille", "sauvegarde.json")
}

// localPath : la sauvegarde à côté de l'exe ("" si ce dossier n'est pas modifiable).
var localPath = func() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	if r, err := filepath.EvalSymlinks(exe); err == nil {
		exe = r
	}
	dir := filepath.Dir(exe)
	test := filepath.Join(dir, ".ecriture_test")
	if os.WriteFile(test, []byte("ok"), 0o644) != nil {
		return ""
	}
	_ = os.Remove(test)
	return filepath.Join(dir, "ToutUneVille_sauvegarde.json")
}()

func savePath() string {
	if localPath != "" {
		return localPath
	}
	return appDataPath()
}

// onPageHide : seulement pour la version navigateur.
func onPageHide(save func()) {}

func readSaveData() ([]byte, error) {
	b, err := os.ReadFile(savePath())
	if err != nil && localPath != "" {
		// pas encore de sauvegarde à côté du jeu : on reprend l'ancienne (AppData)
		b, err = os.ReadFile(appDataPath())
	}
	return b, err
}

func writeSaveData(b []byte) {
	path := savePath()
	if path == "" {
		return
	}
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	tmp := path + ".tmp"
	if os.WriteFile(tmp, b, 0o644) == nil {
		_ = os.Rename(tmp, path)
	}
}

// extraLanguageFiles : les fichiers de langue d'un dossier « langues » posé à côté de l'exe.
func extraLanguageFiles() map[string][]byte {
	out := map[string][]byte{}
	exe, err := os.Executable()
	if err != nil {
		return out
	}
	files, _ := filepath.Glob(filepath.Join(filepath.Dir(exe), "langues", "*.txt"))
	for _, f := range files {
		if b, err := os.ReadFile(f); err == nil {
			out[filepath.Base(f)] = b
		}
	}
	return out
}
