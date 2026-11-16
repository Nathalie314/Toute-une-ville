//go:build js

package main

import (
	"errors"
	"syscall/js"
)

// Version navigateur : la partie est gardée dans la mémoire du navigateur
// (localStorage), propre à chaque appareil et à chaque navigateur.

// isWeb : le jeu tourne-t-il dans un navigateur ?
const isWeb = true

const webSaveKey = "ToutUneVille_sauvegarde"

func readSaveData() ([]byte, error) {
	ls := js.Global().Get("localStorage")
	if ls.IsUndefined() || ls.IsNull() {
		return nil, errors.New("pas de localStorage")
	}
	v := ls.Call("getItem", webSaveKey)
	if v.IsNull() || v.IsUndefined() {
		return nil, errors.New("pas de sauvegarde")
	}
	return []byte(v.String()), nil
}

// onPageHide : enregistre la partie quand on quitte la page (onglet fermé,
// application changée sur l'iPad…), en plus de l'enregistrement régulier.
func onPageHide(save func()) {
	cb := js.FuncOf(func(this js.Value, args []js.Value) any {
		save()
		return nil
	})
	js.Global().Call("addEventListener", "pagehide", cb)
	js.Global().Get("document").Call("addEventListener", "visibilitychange", cb)
}

func writeSaveData(b []byte) {
	ls := js.Global().Get("localStorage")
	if ls.IsUndefined() || ls.IsNull() {
		return
	}
	// l'ancienne sauvegarde devient la copie de secours
	if v := ls.Call("getItem", webSaveKey); !v.IsNull() && !v.IsUndefined() {
		ls.Call("setItem", webSaveKey+"_secours", v.String())
	}
	ls.Call("setItem", webSaveKey, string(b))
}

// readBackupData : la copie de secours.
func readBackupData() ([]byte, error) {
	ls := js.Global().Get("localStorage")
	if ls.IsUndefined() || ls.IsNull() {
		return nil, errors.New("pas de localStorage")
	}
	v := ls.Call("getItem", webSaveKey+"_secours")
	if v.IsNull() || v.IsUndefined() {
		return nil, errors.New("pas de copie de secours")
	}
	return []byte(v.String()), nil
}

// keepBrokenSave : une sauvegarde illisible est gardée de côté.
func keepBrokenSave() {
	ls := js.Global().Get("localStorage")
	if ls.IsUndefined() || ls.IsNull() {
		return
	}
	if v := ls.Call("getItem", webSaveKey); !v.IsNull() && !v.IsUndefined() {
		ls.Call("setItem", webSaveKey+"_abimee", v.String())
		ls.Call("removeItem", webSaveKey)
	}
}

// runFromTemp : (seulement pour la version Windows)
const runFromTemp = false

// extraLanguageFiles : rien dans le navigateur (les langues sont intégrées au jeu).
func extraLanguageFiles() map[string][]byte { return nil }

// screenTipUseful : téléphone ou tablette, jeu ouvert dans le navigateur (pas depuis
// l'écran d'accueil) : le conseil « plein écran » a un sens (voir web/index.html).
func screenTipUseful() bool {
	f := js.Global().Get("ToutUneVilleConseilUtile")
	return f.Type() == js.TypeFunction && f.Invoke().Truthy()
}

// showScreenTip : réaffiche le conseil « jouer en plein écran ».
func showScreenTip() {
	if f := js.Global().Get("ToutUneVilleConseil"); f.Type() == js.TypeFunction {
		f.Invoke()
	}
}
