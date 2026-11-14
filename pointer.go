package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Le pointeur : la souris, ou le doigt sur un écran tactile (iPad, version
// navigateur). Un doigt posé = clic maintenu (on avance vers lui, on appuie
// sur les boutons) ; deux doigts qu'on écarte ou rapproche = zoom.

var (
	touchIDs    []ebiten.TouchID
	justTouched []ebiten.TouchID
	pinchDist   float64 // écart entre deux doigts à l'image précédente (0 : pas de pincement)
)

// pollPointer : à appeler une fois par image, avant les autres fonctions.
func pollPointer() {
	touchIDs = ebiten.AppendTouchIDs(touchIDs[:0])
	justTouched = inpututil.AppendJustPressedTouchIDs(justTouched[:0])
}

// pointerPos : où est le pointeur (en pixels d'écran) ?
func pointerPos() (int, int) {
	if len(touchIDs) > 0 {
		return ebiten.TouchPosition(touchIDs[0])
	}
	return ebiten.CursorPosition()
}

// pointerPressed : bouton de la souris enfoncé, ou un doigt posé.
func pointerPressed() bool {
	return len(touchIDs) == 1 || ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
}

// pointerJustPressed : on vient de cliquer, ou de poser un doigt.
func pointerJustPressed() bool {
	return (len(justTouched) > 0 && len(touchIDs) == 1) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
}

// pinchZoom : de combien de crans zoomer en pinçant à deux doigts (comme la molette).
func pinchZoom() float64 {
	if len(touchIDs) < 2 {
		pinchDist = 0
		return 0
	}
	x0, y0 := ebiten.TouchPosition(touchIDs[0])
	x1, y1 := ebiten.TouchPosition(touchIDs[1])
	d := math.Hypot(float64(x1-x0), float64(y1-y0))
	z := 0.0
	if pinchDist > 0 && d > 0 {
		z = math.Log(d/pinchDist) / math.Log(1.1) // (un cran de molette : ×1,1)
	}
	pinchDist = d
	return z
}
