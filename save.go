package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"ville/sim"
)

// La partie est enregistrée à côté du jeu (ToutUneVille_sauvegarde.json, dans
// le même dossier que l'exe) : on peut emporter le jeu sur une clé USB avec sa
// partie. Si ce dossier n'est pas modifiable, on se rabat sur
// %AppData%\ToutUneVille\sauvegarde.json.

type saveFile struct {
	Jeu  *sim.Save `json:"jeu"`
	Muet bool      `json:"muet"`
	Zoom float64   `json:"zoom"`
	Aide bool      `json:"aide_vue"`
}

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

func (g *game) save() {
	g.saveT = 0
	if g.shots != nil {
		return
	}
	path := savePath()
	if path == "" {
		return
	}
	f := saveFile{Jeu: g.sim.Snapshot(time.Now().Unix()), Muet: g.snd != nil && g.snd.muted, Zoom: g.zoom, Aide: g.helpSeen}
	b, err := json.MarshalIndent(f, "", " ")
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	tmp := path + ".tmp"
	if os.WriteFile(tmp, b, 0o644) == nil {
		_ = os.Rename(tmp, path)
	}
}

func (g *game) load() {
	b, err := os.ReadFile(savePath())
	if err != nil && localPath != "" {
		// pas encore de sauvegarde à côté du jeu : on reprend l'ancienne (AppData)
		b, err = os.ReadFile(appDataPath())
	}
	if err != nil {
		// première partie : on explique les commandes
		g.openPanel(panelHelp)
		g.helpSeen = true
		return
	}
	var f saveFile
	if json.Unmarshal(b, &f) != nil || f.Jeu == nil {
		g.openPanel(panelHelp)
		return
	}
	if f.Jeu.Version < sim.SaveVersion {
		// partie d'avant la version 4.0 : on repart de zéro
		g.openPanel(panelHelp)
		g.helpSeen = true
		g.toast("Version 4.0 : nouvelle ferme, nouvelle partie !", "")
		return
	}
	amount, secs := g.sim.Restore(f.Jeu, time.Now().Unix())
	if g.snd != nil {
		g.snd.muted = f.Muet
	}
	if f.Zoom > 0 {
		g.zoom = f.Zoom
	}
	g.helpSeen = true
	g.sim.Enter(g.sim.Places[0])
	g.camPos = g.sim.Player.Pos
	g.welcome.points = g.sim.Levels["ecole.absence"]
	delete(g.sim.Levels, "ecole.absence")
	if secs > 120 {
		g.welcome.amount, g.welcome.secs = amount, secs
		g.openPanel(panelWelcome)
	}
}
