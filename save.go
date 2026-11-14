package main

import (
	"encoding/json"
	"time"

	"ville/sim"
	"ville/tr"
)

// La partie est enregistrée par writeSaveData / lue par readSaveData :
// version Windows (save_file.go) : un fichier à côté de l'exe ;
// version navigateur (save_web.go) : la mémoire du navigateur (localStorage).

type saveFile struct {
	Jeu  *sim.Save `json:"jeu"`
	Muet bool      `json:"muet"`
	Zoom float64   `json:"zoom"`
	Aide bool      `json:"aide_vue"`
	// Langue : « fr », « en »… (vide : français, pour les parties d'avant le choix de la langue)
	Langue string `json:"langue,omitempty"`
}

func (g *game) save() {
	g.saveT = 0
	if g.shots != nil {
		return
	}
	f := saveFile{Jeu: g.sim.Snapshot(time.Now().Unix()), Muet: g.snd != nil && g.snd.muted, Zoom: g.zoom, Aide: g.helpSeen, Langue: tr.Lang()}
	b, err := json.MarshalIndent(f, "", " ")
	if err != nil {
		return
	}
	writeSaveData(b)
}

func (g *game) load() {
	b, err := readSaveData()
	if err != nil {
		// première partie : on choisit la langue, puis on explique les commandes
		g.openPanel(panelLang)
		g.langThenHelp = true
		g.helpSeen = true
		return
	}
	var f saveFile
	if json.Unmarshal(b, &f) != nil || f.Jeu == nil {
		g.openPanel(panelLang)
		g.langThenHelp = true
		g.openPanel(panelHelp)
		return
	}
	tr.Set(f.Langue)
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
