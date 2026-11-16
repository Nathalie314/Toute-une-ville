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
	// les trois boutons du son : petits sons, téléphone, musique (vrai : coupé)
	SansPetits  bool    `json:"sans_petits,omitempty"`
	SansTel     bool    `json:"sans_tel,omitempty"`
	SansMusique bool    `json:"sans_musique,omitempty"`
	Zoom        float64 `json:"zoom"`
	Aide        bool    `json:"aide_vue"`
	// Langue : « fr », « en »… (vide : français, pour les parties d'avant le choix de la langue)
	Langue string `json:"langue,omitempty"`
}

func (g *game) save() {
	g.saveT = 0
	if g.shots != nil {
		return
	}
	f := saveFile{Jeu: g.sim.Snapshot(time.Now().Unix()), Muet: g.snd != nil && g.snd.muted, Zoom: g.zoom, Aide: g.helpSeen, Langue: tr.Lang()}
	if g.snd != nil {
		f.SansPetits, f.SansTel, f.SansMusique = g.snd.noSmall, g.snd.noPhone, g.snd.noMusic
	}
	b, err := json.MarshalIndent(f, "", " ")
	if err != nil {
		return
	}
	writeSaveData(b)
}

// parseSave : une sauvegarde lisible ?
func parseSave(b []byte) (saveFile, bool) {
	var f saveFile
	return f, json.Unmarshal(b, &f) == nil && f.Jeu != nil
}

func (g *game) load() {
	if runFromTemp { // lancé depuis le zip : on prévient (après le chargement)
		defer g.openPanel(panelTempWarn)
	}
	b, err := readSaveData()
	f, ok := parseSave(b)
	if err == nil && !ok {
		// la sauvegarde est abîmée : on la met de côté (elle n'est jamais écrasée)…
		keepBrokenSave()
	}
	if !ok {
		// … et on repart de la copie de secours (la sauvegarde précédente), s'il y en a une
		if bb, e := readBackupData(); e == nil {
			if fb, okb := parseSave(bb); okb {
				f, ok, err = fb, true, nil
				g.toast("Ta sauvegarde était abîmée : la partie reprend depuis la copie de secours.", "")
			}
		}
	}
	if err != nil && !ok {
		// première partie : on choisit la langue, puis on explique les commandes
		g.openPanel(panelLang)
		g.langThenHelp = true
		g.helpSeen = true
		return
	}
	if !ok {
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
		g.snd.noSmall, g.snd.noPhone, g.snd.noMusic = f.SansPetits, f.SansTel, f.SansMusique
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
