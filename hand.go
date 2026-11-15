package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// La main : sur la carte et dans les fenêtres, les flèches (ou ZQSD) déplacent
// une grosse main. Restée un moment sur un bouton, elle l'active ; Entrée ou
// Espace l'activent tout de suite. La souris reste utilisable : dès qu'elle
// bouge, la main disparaît.

const handDwell = 1.0

type handState struct {
	on        bool
	pos       V
	dwell     float64
	last      [4]float64 // bouton visé (son rectangle)
	lastMouse V
	hidden    bool
}

func (g *game) handContext() bool { return g.panel != panelNone || g.scene == sceneMap }

// handButtons : les boutons visables (ceux de la fenêtre ouverte, sinon tous).
func (g *game) handButtons() []button {
	if g.panel != panelNone && g.panelBtn0 <= len(g.btns) {
		return g.btns[g.panelBtn0:]
	}
	return g.btns
}

func (g *game) updateHand(dt float64) {
	h := &g.hand
	// g.mouse est ici la vraie position du curseur (calculée juste avant)
	if _, wy := ebiten.Wheel(); g.mouse.Sub(h.lastMouse).Len() > 2 || wy != 0 {
		h.on = false // la souris a bougé (ou sa molette a tourné) : elle reprend la main
	}
	h.lastMouse = g.mouse
	if !g.handContext() {
		h.on = false
		g.setCursorHidden(false)
		return
	}
	k := ebiten.IsKeyPressed
	dir := V{}
	if k(ebiten.KeyArrowUp) || k(ebiten.KeyW) || k(ebiten.KeyZ) {
		dir.Y--
	}
	if k(ebiten.KeyArrowDown) || k(ebiten.KeyS) {
		dir.Y++
	}
	if k(ebiten.KeyArrowLeft) || k(ebiten.KeyA) || k(ebiten.KeyQ) {
		dir.X--
	}
	if k(ebiten.KeyArrowRight) || k(ebiten.KeyD) {
		dir.X++
	}
	enter := inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter)
	if dir.Len() > 0 || enter {
		if !h.on {
			h.on = true
			if h.pos == (V{}) {
				h.pos = V{X: scr.VW / 2, Y: scr.VH / 2}
			}
		}
	}
	g.setCursorHidden(h.on)
	if !h.on {
		return
	}
	if dir.Len() > 0 {
		h.pos = h.pos.Add(dir.Mul(900 * dt / dir.Len()))
		h.pos.X = math.Max(10, math.Min(scr.VW-10, h.pos.X))
		h.pos.Y = math.Max(10, math.Min(scr.VH-10, h.pos.Y))
	}
	// défilement des listes quand la main touche le haut ou le bas
	if g.panel != panelNone {
		if h.pos.Y > scr.VH-70 {
			g.scroll += 700 * dt
		} else if h.pos.Y < 70 {
			g.scroll -= 700 * dt
		}
	}
	g.mouse = h.pos // pour l'effet de survol (h.lastMouse garde la vraie position du curseur)
	// bouton sous la main
	var hit *button
	bs := g.handButtons()
	for i := len(bs) - 1; i >= 0; i-- {
		b := bs[i]
		if h.pos.X >= b.x && h.pos.X <= b.x+b.w && h.pos.Y >= b.y && h.pos.Y <= b.y+b.h {
			hit = &bs[i]
			break
		}
	}
	if hit == nil {
		h.dwell, h.last = 0, [4]float64{}
		return
	}
	r := [4]float64{hit.x, hit.y, hit.w, hit.h}
	if r != h.last {
		h.dwell, h.last = 0, r
	}
	if enter {
		g.handPress(hit)
		return
	}
	if hit.danger {
		return // quitter ou tout effacer : seulement avec Entrée
	}
	h.dwell += dt
	if h.dwell >= handDwell {
		g.handPress(hit)
	}
}

func (g *game) handPress(b *button) {
	g.hand.dwell = -0.4 // on peut répéter (ex. « +1 ») en restant dessus
	g.snd.play(sClick)
	if b.fn != nil {
		b.fn()
	}
}

func (g *game) setCursorHidden(hide bool) {
	if g.hand.hidden == hide || g.shots != nil {
		return
	}
	g.hand.hidden = hide
	if hide {
		ebiten.SetCursorMode(ebiten.CursorModeHidden)
	} else {
		ebiten.SetCursorMode(ebiten.CursorModeVisible)
	}
}

// drawHand : la main et, au-dessus d'un bouton, le cercle qui se remplit.
func (g *game) drawHand(dst *ebiten.Image) {
	h := &g.hand
	if !h.on || !g.handContext() {
		return
	}
	x, y := h.pos.X, h.pos.Y
	if h.dwell > 0 && h.last != ([4]float64{}) {
		f := math.Min(1, h.dwell/handDwell)
		ring(dst, x, y, 46, 8, rgba(0x000000, 0.25))
		n := int(40 * f)
		for i := 0; i < n; i++ {
			a0 := -math.Pi/2 + 2*math.Pi*float64(i)/40
			a1 := -math.Pi/2 + 2*math.Pi*float64(i+1)/40
			line(dst, x+math.Cos(a0)*46, y+math.Sin(a0)*46, x+math.Cos(a1)*46, y+math.Sin(a1)*46, 8, rgb(0xffd23f))
		}
	}
	scaled(x, y, 1.7, func() {
		// la main (l'index pointe sur (x, y))
		skin, edge := rgb(0xffe0bd), rgb(0x8a5a33)
		shadow := rgba(0x000000, 0.25)
		roundRect(dst, x-14+6, y+18+6, 46, 44, 16, shadow)
		// paume et doigts repliés
		roundRect(dst, x-16, y+16, 50, 46, 16, edge)
		roundRect(dst, x-13, y+19, 44, 40, 14, skin)
		for k := 0; k < 3; k++ {
			roundRect(dst, x+2+float64(k)*11, y+12, 13, 22, 6, edge)
			roundRect(dst, x+4+float64(k)*11, y+14, 9, 18, 4, skin)
		}
		// index tendu
		roundRect(dst, x-8, y-4, 16, 34, 8, edge)
		roundRect(dst, x-5, y-1, 10, 30, 5, skin)
		// pouce
		roundRect(dst, x-26, y+26, 22, 14, 7, edge)
		roundRect(dst, x-23, y+29, 18, 9, 4, skin)
	})
}
