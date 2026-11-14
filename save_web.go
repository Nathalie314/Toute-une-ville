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
	ls.Call("setItem", webSaveKey, string(b))
}
