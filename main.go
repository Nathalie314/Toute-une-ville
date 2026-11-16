package main

import (
	"errors"
	"flag"
	"image"
	"image/color"
	"math"
	"os"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"ville/sim"
	"ville/tr"
)

// Toute une Ville — un jeu de gestion où l'on construit Clairbourg,
// lieu après lieu, en commençant par la ferme.

var errQuit = errors.New("quitter")

// scr : taille de l'écran. L'interface fait toujours 1080 unités de haut.
var scr struct {
	W, H   float64 // pixels
	S      float64 // pixels par unité
	K      float64 // sur-échantillonnage du dessin (anticrénelage)
	VW, VH float64 // taille en unités
	Small  bool    // petit écran (téléphone dans le navigateur) : textes d'aide plus gros
}

type scene int

const (
	sceneMap scene = iota
	scenePlace
)

type panel int

const (
	panelNone panel = iota
	panelMairie
	panelWardrobe
	panelWelcome
	panelHelp
	panelReset
	panelBadges
	panelBuy
	panelPoints
	panelLang
	panelGains
	panelPlaceHelp // le « ? » d'un lieu : comment fonctionne ce lieu
)

type button struct {
	x, y, w, h float64
	fn         func()
	danger     bool // pas d'activation « en restant dessus » (quitter, tout effacer)
}

type game struct {
	wheelAcc  float64    // crans de molette pas encore utilisés par une liste
	lastPlace *sim.Place // le dernier lieu visité (Tab sur la carte y ramène)
	mapHover  string     // carte : le lieu sous la main ou la souris (un chiffre en fait un favori)
	sim       *sim.Game
	scene     scene
	panel     panel
	helpPage  int       // page de l'aide (petits écrans : l'aide tient sur plusieurs pages)
	gains     gainsView // la page des gains par lieu (calculée à son ouverture)
	zoom      float64
	camPos    V
	t         float64

	parts  []particle
	toasts []toastMsg
	snd    *soundBank
	saveT  float64
	scroll float64

	btns    []button
	mouse   V // position de la souris en unités d'interface
	pressUI bool
	welcome struct {
		amount, secs float64
		points       int
	}
	pending            map[string]int // points de l'école choisis mais pas encore validés
	ptsTab             int            // onglet des points (primaire…)
	ptsPlace           string         // lieu choisi dans l'onglet
	ptsScroll          int            // ascenseur de la liste des lieux (points)
	ambiT              float64        // écran Personnage : temps appuyé sur « Ambidextre »
	seenPark, seenCart int            // supermarché : clients refusés déjà signalés
	refusedAt          float64
	refusedWhat        string
	busT               float64   // attente sur un téléporteur de l'hélicoptère
	vestDone           bool      // la page « Personnage » s'est ouverte au vestiaire (on attend qu'il en sorte)
	hand               handState // la main (menus au clavier)
	panelBtn0          int       // premier bouton de la fenêtre ouverte
	lastPay            float64
	coinSnd            float64
	shots              *shotScript
	typed              []rune // caractères tapés à cette image
	saveShots          *saveShots
	capName            string
	helpSeen           bool
	langThenHelp       bool // premier lancement : après le choix de la langue, l'aide
	quit               bool
	off                *ebiten.Image
	buying             *sim.Place
}

func newGame() *game {
	g := &game{sim: sim.NewGame(), zoom: 1.15, scene: scenePlace}
	g.camPos = g.sim.Player.Pos
	return g
}

// ---------------------------------------------------------------- boucle

func (g *game) Update() error {
	const dt = 1.0 / 60
	if g.saveShots != nil {
		return g.saveShots.update(g)
	}
	if g.shots != nil {
		return g.shots.update(g)
	}
	if g.quit {
		return errQuit
	}
	if ebiten.IsWindowBeingClosed() {
		g.save()
		return errQuit
	}
	g.t += dt
	g.input(dt)
	g.updateBus(dt)
	g.sim.Step(dt)
	g.handleEvents()
	g.updateParticles(dt)
	if g.scene == scenePlace && g.sim.Cur != nil {
		g.updateCamera(g.sim.Cur, dt)
	}
	g.saveT += dt
	if g.saveT > 20 {
		g.save()
	}
	return nil
}

func (g *game) input(dt float64) {
	pollPointer() // (souris ou doigt)
	mx, my := pointerPos()
	g.mouse = V{X: float64(mx) / scr.S, Y: float64(my) / scr.S}
	g.updateHand(dt)
	// plein écran
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) || (inpututil.IsKeyJustPressed(ebiten.KeyEnter) && ebiten.IsKeyPressed(ebiten.KeyAlt)) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	// les caractères tapés (et non la position des touches : sur un clavier AZERTY,
	// la touche « , ? » est à la place du M d'un clavier anglais)
	g.typed = ebiten.AppendInputChars(g.typed[:0])
	if g.typedAny('m', 'M') {
		g.toggleMute()
	}
	// $ : la gestion de l'argent (le tableau des gains, comme un clic sur l'argent)
	if g.typedAny('$') && g.panel == panelNone {
		g.openGains()
	}
	// clics sur les boutons (ceux dessinés à l'image précédente)
	if pointerJustPressed() {
		g.pressUI = false
		for i := len(g.btns) - 1; i >= 0; i-- {
			b := g.btns[i]
			if g.mouse.X >= b.x && g.mouse.X <= b.x+b.w && g.mouse.Y >= b.y && g.mouse.Y <= b.y+b.h {
				g.pressUI = true
				g.snd.play(sClick)
				if b.fn != nil {
					b.fn()
				}
				break
			}
		}
		if g.panel != panelNone {
			g.pressUI = true
		}
	}
	_, wy := ebiten.Wheel()
	wy += pinchZoom() // (deux doigts : on zoome en pinçant)
	g.wheelAcc += wy  // (pour les listes des fenêtres, lues au dessin)
	// Tab : la carte de la ville (et retour au lieu) ; Échap : fermer une fenêtre
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		switch {
		case g.panel != panelNone:
			g.closePanel()
		case g.scene == scenePlace:
			g.goMap()
		case g.scene == sceneMap && g.lastPlace != nil && g.lastPlace.Owned:
			g.goPlace(g.lastPlace)
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyEscape) && g.panel != panelNone {
		g.closePanel()
	}
	// l'aide : la touche « , ? » (avec ou sans majuscule), aussi sur la carte
	if g.panel == panelNone && g.typedAny(',', '?') {
		g.openHelp()
	}
	if g.panel != panelNone {
		g.scroll -= wy * 60
		g.sim.MovePlayer(V{}, dt)
		return
	}
	if g.scene == scenePlace {
		if inpututil.IsKeyJustPressed(ebiten.KeyB) {
			g.openPanel(panelBadges)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyH) || inpututil.IsKeyJustPressed(ebiten.KeyF1) {
			g.openHelp()
		}
		// zoom : molette, ou + et - au clavier
		if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadAdd) {
			wy++
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadSubtract) || inpututil.IsKeyJustPressed(ebiten.Key6) {
			wy--
		}
		minZ := 0.6
		if g.sim.Cur != nil && g.sim.Cur.ID == "parc" {
			minZ = 0.2 // (le parc est immense : on peut reculer bien plus)
		}
		g.zoom = math.Max(minZ, math.Min(1.9, g.zoom*math.Pow(1.1, wy)))
		dir := V{}
		k := ebiten.IsKeyPressed
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
		if dir.Len() == 0 && pointerPressed() && !g.pressUI {
			// on va vers la souris (ou vers le doigt)
			v := g.worldView(1)
			mx, my := pointerPos()
			target := v.toLogical(float64(mx), float64(my))
			d := target.Sub(g.sim.Player.Pos)
			if d.Len() > 12 {
				dir = d.Mul(1 / math.Max(d.Len(), 60))
			}
		}
		if dir.Len() > 1 {
			dir = dir.Norm()
		}
		if g.isoHere() && dir.Len() > 0 { // flèche vers le haut de l'écran : vers le haut du losange
			l := dir.Len()
			x, y := isoInv(dir.X, dir.Y)
			dir = V{X: x, Y: y}.Norm().Mul(l)
		}
		g.sim.MovePlayer(dir, dt)
	}
}

// openHelp : dans un lieu, le « ? » explique ce lieu ; sur la carte de la ville, comment jouer.
func (g *game) openHelp() {
	if g.scene == scenePlace && g.sim.Cur != nil {
		g.openPanel(panelPlaceHelp)
		return
	}
	g.openPanel(panelHelp)
}

func (g *game) openPanel(p panel) {
	g.panel = p
	g.scroll = 0
	g.helpPage = 0
}

func (g *game) closePanel() {
	if g.panel == panelLang && g.langThenHelp {
		g.langThenHelp = false
		g.openPanel(panelHelp)
		return
	}
	g.panel = panelNone
	g.pending = nil
	g.save()
}

func (g *game) goMap() {
	if g.sim.Cur != nil {
		g.lastPlace = g.sim.Cur // (Tab sur la carte : on y retourne)
	}
	g.sim.Enter(nil)
	g.scene = sceneMap
	g.save()
}

func (g *game) goPlace(p *sim.Place) {
	g.sim.Enter(p)
	g.scene = scenePlace
	g.camPos = g.sim.Player.Pos
}

func (g *game) toggleMute() {
	g.snd.muted = !g.snd.muted
}

// ---------------------------------------------------------------- dessin

// Draw dessine tout dans une image deux fois plus grande que l'écran puis la
// réduit : c'est l'anticrénelage, bien plus rapide que de lisser chaque forme.
func (g *game) Draw(screen *ebiten.Image) {
	k := 2.0
	if scr.H > 1500 {
		k = 1.5
	}
	if scr.H > 2200 {
		k = 1
	}
	scr.K = k
	w, h := int(math.Ceil(scr.W*k)), int(math.Ceil(scr.H*k))
	if g.off == nil || g.off.Bounds().Dx() != w || g.off.Bounds().Dy() != h {
		if g.off != nil {
			g.off.Deallocate()
		}
		g.off = ebiten.NewImageWithOptions(image.Rect(0, 0, w, h), &ebiten.NewImageOptions{Unmanaged: true})
	}
	dst := g.off
	dst.Clear()
	g.btns = g.btns[:0]
	g.panelBtn0 = 0
	ui := view{s: scr.S * k}
	if g.scene == scenePlace && g.sim.Cur != nil {
		g.drawWorld(dst, g.sim.Cur)
		cam = ui
		g.drawHUD(dst)
	} else {
		cam = ui
		g.drawMap(dst)
	}
	cam = ui
	g.drawToasts(dst)
	switch g.panel {
	case panelMairie:
		g.drawMairiePanel(dst)
	case panelWardrobe:
		g.drawWardrobePanel(dst)
	case panelWelcome:
		g.drawWelcome(dst)
	case panelHelp:
		g.drawHelp(dst)
	case panelPlaceHelp:
		g.drawPlaceHelp(dst)
	case panelReset:
		g.drawReset(dst)
	case panelBadges:
		g.drawBadges(dst)
	case panelBuy:
		g.drawBuyPlace(dst)
	case panelPoints:
		g.drawPointsPanel(dst)
	case panelLang:
		g.drawLangPanel(dst)
	case panelGains:
		g.drawGains(dst)
	}
	g.drawHand(dst)
	g.wheelAcc = 0 // (un cran de molette ne sert qu'une fois)
	flush()
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Scale(1/k, 1/k)
	screen.DrawImage(dst, op)
	if g.capName != "" {
		saveScreenshot(screen, g.capName)
		g.capName = ""
	}
}

func (g *game) Layout(ow, oh int) (int, int) {
	return ow, oh
}

func (g *game) LayoutF(ow, oh float64) (float64, float64) {
	s := ebiten.Monitor().DeviceScaleFactor()
	w, h := ow*s, oh*s
	scr.W, scr.H = w, h
	scr.S = h / 1080
	scr.VW, scr.VH = w/scr.S, 1080
	scr.Small = isWeb && oh < 600
	return w, h
}

func (g *game) addButton(x, y, w, h float64, fn func()) {
	g.btns = append(g.btns, button{x: x, y: y, w: w, h: h, fn: fn})
}

func (g *game) hover(x, y, w, h float64) bool {
	return g.mouse.X >= x && g.mouse.X <= x+w && g.mouse.Y >= y && g.mouse.Y <= y+h
}

// ---------------------------------------------------------------- lancement

func main() {
	capDir := flag.String("captures", "", "dossier de captures (développement)")
	windowed := flag.Bool("fenetre", false, "démarrer en fenêtre")
	flag.Parse()

	g := newGame()
	g.snd = newSoundBank()
	if *capDir == "" {
		g.load()
	} else {
		g.shots = newShotScript(*capDir)
		if f := os.Getenv("CAPSAVE"); f != "" {
			g.saveShots = newSaveShots(*capDir, f)
		}
	}

	ebiten.SetWindowTitle("Toute une Ville")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowIcon([]image.Image{appIcon()})
	ebiten.SetWindowClosingHandled(true)
	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetTPS(60)
	ebiten.SetWindowSize(1280, 720)
	if isWeb {
		onPageHide(g.save) // (navigateur : on enregistre quand on quitte la page)
	} else if *capDir == "" && !*windowed {
		ebiten.SetFullscreen(true)
	} else {
		ebiten.SetWindowSize(1600, 900)
	}
	if os.Getenv("TR_LANGUE") != "" { // outil de traduction : jouer dans une langue et noter ce qui manque
		tr.Set(os.Getenv("TR_LANGUE"))
		tr.Track()
	}
	err := ebiten.RunGame(g)
	if os.Getenv("TR_LANGUE") != "" {
		if f := os.Getenv("TR_MANQUANTS"); f != "" {
			_ = os.WriteFile(f, []byte(strings.Join(tr.Missing(), "\n")+"\n"), 0o644)
		}
	}
	if err != nil && err != errQuit && err != errDone {
		os.Exit(1)
	}
}

// appIcon : une petite maison sur fond vert.
func appIcon() image.Image {
	const n = 64
	img := image.NewRGBA(image.Rect(0, 0, n, n))
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			c := color.RGBA{0x6c, 0xb8, 0x4a, 255}
			dx, dy := float64(x)-31.5, float64(y)-31.5
			if math.Hypot(dx, dy) > 31 {
				c = color.RGBA{}
			}
			// toit
			if y >= 12 && y < 32 && math.Abs(dx) < float64(y-12)*1.1+2 {
				c = color.RGBA{0xc0, 0x39, 0x2b, 255}
			}
			// murs
			if y >= 32 && y < 52 && math.Abs(dx) < 18 {
				c = color.RGBA{0xf3, 0xea, 0xd8, 255}
				if y >= 38 && math.Abs(dx) < 5 {
					c = color.RGBA{0x8a, 0x5a, 0x33, 255}
				}
			}
			img.Set(x, y, c)
		}
	}
	return img
}

var _ = time.Now

// typedAny : un de ces caractères a-t-il été tapé à cette image ?
func (g *game) typedAny(rs ...rune) bool {
	for _, t := range g.typed {
		for _, r := range rs {
			if t == r {
				return true
			}
		}
	}
	return false
}
