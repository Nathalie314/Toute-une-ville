//go:build !js

package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
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
	if runFromTemp {
		return "" // (dossier temporaire : on enregistre dans AppData, qui reste)
	}
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

// backupPath : la copie de secours (la sauvegarde d'avant).
func backupPath(path string) string {
	return strings.TrimSuffix(path, ".json") + "_secours.json"
}

func writeSaveData(b []byte) {
	path := savePath()
	if path == "" {
		return
	}
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	tmp := path + ".tmp"
	if os.WriteFile(tmp, b, 0o644) != nil {
		return
	}
	// l'ancienne sauvegarde devient la copie de secours, puis la nouvelle prend sa place
	// (si le jeu est coupé entre les deux, la copie de secours est relue au lancement)
	if _, err := os.Stat(path); err == nil {
		_ = os.Remove(backupPath(path))
		_ = os.Rename(path, backupPath(path))
	}
	_ = os.Rename(tmp, path)
}

// readBackupData : la copie de secours.
func readBackupData() ([]byte, error) { return os.ReadFile(backupPath(savePath())) }

// keepBrokenSave : une sauvegarde illisible est gardée de côté, sous un autre nom.
func keepBrokenSave() {
	path := savePath()
	_ = os.Rename(path, strings.TrimSuffix(path, ".json")+"_abimee_"+time.Now().Format("2006-01-02_15h04")+".json")
}

// runFromTemp : le jeu a été lancé depuis un dossier temporaire (double clic sur l'exe
// dans le zip, sans l'extraire) : ce dossier est effacé, la partie serait perdue. On
// enregistre alors dans le dossier de l'utilisateur (AppData), et on prévient.
var runFromTemp = func() bool {
	exe, err := os.Executable()
	if err != nil || runtime.GOOS != "windows" {
		return false
	}
	d := strings.ToLower(filepath.Dir(exe))
	t := strings.ToLower(os.TempDir())
	return (t != "" && strings.HasPrefix(d, t)) || strings.Contains(d, `\temp\`) || strings.Contains(d, `\appdata\local\temp`) || strings.Contains(d, ".zip")
}()

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

// (pas de navigateur : pas de conseil « plein écran »)
func screenTipUseful() bool { return false }
func showScreenTip()        {}
