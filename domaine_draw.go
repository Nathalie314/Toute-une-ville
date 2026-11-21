package main

import (
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// Le domaine : le sol (chemin, terrasse de la boutique), la route verticale des
// camions à droite, les fabriques (presses et foudres, brasserie, alambics,
// pressoirs), le houblon et la vigne, les icônes des boissons et l'avertissement
// sur l'alcool dans la zone du joueur.

func init() {
	cropCol[sim.Houblon] = 0x8fcf4a
	cropCol[sim.Raisin] = 0x6a2a6e
}

// domaineGround : herbe, chemin, terrasse en bois, route du bas et route verticale.
func (g *game) domaineGround(dst *ebiten.Image, p *sim.Place, tl, br V) {
	path := rgb(0xd9c08f)
	// le chemin sous les champs, et les allées vers la boutique et la distillerie
	roundRect(dst, 300, sim.DomPathY, sim.DomColX-460, 70, 35, path)
	roundRect(dst, 1180, sim.DomPathY, 140, 560, 50, path)
	roundRect(dst, sim.DomColX-560, sim.DomColY0-40, 160, 1080, 60, path)
	// les espaces : leur nom sur le chemin
	for _, z := range sim.TeamZones(p.ID) {
		a := 0.85
		if !p.ZoneOpen(z.ID) {
			a = 0.35
		}
		cx := (z.X0 + z.X1) / 2
		if z.ID == "distillerie" {
			roundRect(dst, sim.DomColX-560, sim.DomColY0-80, 220, 30, 10, rgba(0x6a2a4a, a))
			say(dst, "DISTILLERIE", sim.DomColX-450, sim.DomColY0-65, 17, true, 0, rgba(0xffffff, a))
			continue
		}
		roundRect(dst, cx-90, sim.DomPathY+20, 180, 30, 10, rgba(0x6a2a4a, a))
		say(dst, strings.ToUpper(z.Name), cx, sim.DomPathY+35, 17, true, 0, rgba(0xffffff, a))
	}
	// la boutique : une terrasse en bois sombre (une cave)
	roundRect(dst, 110, 1030, 3150, 480, 30, rgb(0x8a5a3a))
	roundRect(dst, 120, 1040, 3130, 460, 26, rgb(0xd9b98a))
	for x := 120.0; x < 3240; x += 70 {
		line(dst, x, 1042, x, 1498, 1.5, rgba(0x8a5a3a, 0.35))
	}
	// la réserve : une dalle grise contre la route
	roundRect(dst, sim.DomPalX-30, sim.DomColY0, 340, 4.8*sim.DomColStep+60, 20, rgba(0xb9b3a8, 0.9))
	say(dst, "RÉSERVE (camions de ma ferme)", sim.DomPalX+140, sim.DomColY0-18, 16, true, 0, rgba(0x4a4a4a, 0.85))
	// les quais, au nord de la réserve
	roundRect(dst, sim.DomPalX-30, sim.DomDockY0-100, 340, 3*sim.DomDockStep+40, 20, rgba(0xb9b3a8, 0.6))
	// les réserves de jus, à gauche du quai du supermarché (une dalle qui le rejoint)
	roundRect(dst, sim.DomJusX-40, sim.DomJusY-120, sim.DomPalX-sim.DomJusX+60, 340, 20, rgba(0xb9b3a8, 0.6))
	// --- la route du bas, puis la route verticale des camions (elle la rejoint)
	road(dst, tl, br, 1640)
	x0 := sim.DomRoadX
	top := math.Min(tl.Y-20, sim.DomTop-200)
	rect(dst, x0-12, top, 12, 1640-top, rgb(0xb0b0a8))
	rect(dst, x0, top, sim.DomRoadW, 1640-top+150, rgb(0x6b6f74))
	rect(dst, x0+sim.DomRoadW, top, 12, 1640-top, rgb(0xb0b0a8))
	for y := math.Floor(top/160) * 160; y < 1620; y += 160 {
		rect(dst, x0+sim.DomRoadW/2-4, y, 8, 80, rgb(0xf2f2e6))
	}
	say(dst, "CAMIONS", x0+sim.DomRoadW/2, 1560, 16, true, 0, rgba(0xffffff, 0.8))
	rect(dst, 1180, 1500, 140, 140, path)
	// la clôture (avec le portail) et le panneau d'entrée
	fence := func(x1, x2 float64) {
		for x := x1; x <= x2; x += 40 {
			rect(dst, x-5, 1532, 10, 40, rgb(0x6e4322))
			circle(dst, x, 1532, 6, rgb(0x8a5a33))
		}
		rect(dst, x1, 1542, x2-x1, 7, rgb(0x8a5a33))
		rect(dst, x1, 1558, x2-x1, 7, rgb(0x8a5a33))
	}
	fence(60, 1170)
	fence(1330, x0-40)
	roundRect(dst, 1130, 1475, 240, 46, 8, rgb(0x6a2a4a))
	say(dst, "LE DOMAINE", 1250, 1498, 24, true, 0, rgb(0xffe7a8))
	g.backTruckV(dst, p)
}

// drawAlcoholNotice : l'avertissement, bien lisible, dans la zone du joueur
// (à droite des embauches, au-dessus du vestiaire : rien d'autre à cet endroit).
func drawAlcoholNotice(dst *ebiten.Image, p *sim.Place) {
	x, y, w, h := p.Yard.X+440, 1030.0, 370.0, 130.0
	roundRect(dst, x, y, w, h, 12, rgba(0xffffff, 0.95))
	roundRect(dst, x, y, 10, h, 4, rgb(0xc0392b))
	sayFit(dst, "L'alcool est dangereux", x+w/2+5, y+30, 22, true, 0, w-30, rgb(0x2b2b2b))
	sayFit(dst, "pour la santé.", x+w/2+5, y+62, 22, true, 0, w-30, rgb(0x2b2b2b))
	sayFit(dst, "À consommer avec modération.", x+w/2+5, y+98, 20, false, 0, w-30, rgb(0x2b2b2b))
}

// backTruckV : le camion de ma ferme monte la route verticale jusqu'à la réserve.
func (g *game) backTruckV(dst *ebiten.Image, p *sim.Place) {
	best, every := 1e9, 25.0
	var from, fromID string
	for _, q := range g.sim.Places {
		for _, d := range q.Stations {
			if d.Kind == sim.Dock && d.To == p.ID && g.sim.DockActive(d) && q.Owned {
				if l := g.sim.TruckLeft(d); l < best {
					best, from, every, fromID = l, q.Name, g.sim.DockEvery(d), q.ID
				}
			}
		}
	}
	if from == "" {
		return
	}
	e := every - best
	truckFrom = fromID
	defer func() { truckFrom = "ferme" }()
	x := sim.DomRoadX + sim.DomRoadW/2 + 30
	stop := sim.DomColY0 + 2*sim.DomColStep + 60 // au milieu de la réserve
	switch {
	case best < 3: // il monte depuis la route du bas
		f := (3 - best) / 3
		drawTruckTop(dst, x, stop+1300*(1-f)*(1-f), true)
	case e < 2: // il décharge
		drawTruckTop(dst, x, stop, true)
	case e < 4.5: // il continue vers le nord
		f := (e - 2) / 2.5
		drawTruckTop(dst, x, stop-2400*f*f, false)
	}
	sayOutlined(dst, "Prochaine livraison ("+strings.ToLower(from)+") : "+itoa(int(best))+" s", sim.DomPalX-20, sim.DomColY0-52, 18, -1, rgb(0xffffff), rgb(0x2b2b2b))
}

// drawDockV : un quai du domaine, son camion garé sur la route verticale.
func (g *game) drawDockV(dst *ebiten.Image, s *sim.Station) {
	b := s.Body
	roundRect(dst, b.X, b.Y, b.W, b.H, 6, rgb(0x9a6b3e))
	k := 0
	for it := sim.Item(0); it < sim.NItems; it++ {
		for n := 0; n < s.Buf[it] && k < 22; n++ {
			drawItem(dst, it, b.X+16+float64(k%11)*17, b.Y+14+float64(k/11)*14, 11)
			k++
		}
	}
	left := g.sim.TruckLeft(s)
	every := g.sim.DockEvery(s)
	truckFrom = "domaine"
	defer func() { truckFrom = "ferme" }()
	e := every - left
	x := sim.DomRoadX + sim.DomRoadW/2 + 30
	if s.PlaceID() == "port" {
		g.drawPortTruck(dst, s, left, every)
		return
	}
	y0 := b.C().Y + 10
	truck := func(y float64) { scaled(x, y, 0.58, func() { drawTruckTop(dst, x, y, true) }) }
	if e < 2.5 { // il vient de partir vers le nord
		f := e / 2.5
		truck(y0 - 1800*f*f)
	}
	switch {
	case e >= 4.5:
		scaled(x, y0, 0.58, func() { drawTruckTop(dst, x, y0, left < 2) })
		sayOutlined(dst, "Départ dans "+itoa(int(left))+" s", x, y0+50, 15, 0, rgb(0xffffff), rgb(0x2b2b2b))
	case e >= 1.5: // le suivant monte depuis le sud
		f := (e - 1.5) / 3
		truck(y0 + 1800*(1-f)*(1-f))
	}
}

// drawDomaineMachine : les fabriques du domaine.
func drawDomaineMachine(dst *ebiten.Image, s *sim.Station, t float64) {
	b := s.Body
	ellipse(dst, b.C().X, b.Y+b.H, b.W*0.55, 16, rgba(0x000000, 0.18))
	work := s.Working
	switch s.ID {
	case "vin_blanc", "vin_rouge":
		// une grosse bassine de raisin, la presse au-dessus, le foudre (grand tonneau debout) à côté
		wine := uint32(0xe8d77a)
		if s.ID == "vin_rouge" {
			wine = 0x8a1f3a
		}
		bx := b.X + 10
		ellipse(dst, bx+70, b.Y+b.H-30, 70, 26, rgb(0x8a5a33))
		ellipse(dst, bx+70, b.Y+b.H-46, 64, 18, rgb(0x6a2a6e))
		for k := 0; k < 6; k++ {
			circle(dst, bx+30+float64(k)*15, b.Y+b.H-48+float64(k%2)*4, 6, rgb(0x8a3a8e))
		}
		// la presse : une vis et un plateau
		py := b.Y + 40 + math.Abs(math.Sin(t*1.5))*12*boolF(work)
		rect(dst, bx+66, b.Y+10, 8, py-b.Y, rgb(0x5a5a5a))
		roundRect(dst, bx+20, py, 100, 14, 4, rgb(0x6e4322))
		rect(dst, bx+10, b.Y+6, 120, 10, rgb(0x6e4322))
		rect(dst, bx+10, b.Y+6, 8, b.H-40, rgb(0x6e4322))
		rect(dst, bx+122, b.Y+6, 8, b.H-40, rgb(0x6e4322))
		// le tuyau vers le foudre
		line(dst, bx+130, b.Y+b.H-40, b.X+b.W-150, b.Y+b.H-60, 6, rgb(0x9a9a9a))
		// le foudre : un grand tonneau debout
		fx, top, bot := b.X+b.W-85, b.Y-20, b.Y+b.H-8
		ellipse(dst, fx+20, bot+4, 70, 14, rgba(0x000000, 0.2))
		mid := (top + bot) / 2
		hh := (bot - top) / 2
		polygon(dst, []V{{X: fx - 52, Y: top + 14}, {X: fx - 64, Y: mid}, {X: fx - 52, Y: bot - 6}, {X: fx + 52, Y: bot - 6}, {X: fx + 64, Y: mid}, {X: fx + 52, Y: top + 14}}, rgb(0xa86a3a))
		polygon(dst, []V{{X: fx - 52, Y: top + 14}, {X: fx - 64, Y: mid}, {X: fx - 52, Y: bot - 6}, {X: fx - 30, Y: bot - 6}, {X: fx - 38, Y: mid}, {X: fx - 30, Y: top + 14}}, rgb(0xc0824d))
		for k := -2; k <= 2; k++ { // les douelles
			x := fx + float64(k)*22
			line(dst, x, top+16, x+float64(k)*3, mid, 2, rgba(0x5a3a22, 0.5))
			line(dst, x+float64(k)*3, mid, x, bot-8, 2, rgba(0x5a3a22, 0.5))
		}
		for _, y := range []float64{top + 26, mid - hh*0.35, mid + hh*0.35, bot - 18} { // les cercles de fer
			w := 52 + 12*(1-math.Abs(y-mid)/hh)
			line(dst, fx-w, y, fx+w, y, 5, rgb(0x4a4a4a))
		}
		ellipse(dst, fx, top+14, 52, 12, rgb(0x8a5a33)) // le dessus
		ellipse(dst, fx, top+14, 44, 8, rgb(0x6e4322))
		roundRect(dst, fx-20, mid-8, 40, 34, 8, rgb(wine)) // l'étiquette du vin
		circle(dst, fx, bot-30, 5, rgb(0x5a3a22))          // le robinet
	case "biere":
		// deux cuves de cuivre et une cheminée
		for k, x := range []float64{b.X + 70, b.X + 200} {
			r := 60.0 - float64(k)*8
			rect(dst, x-r, b.Y+b.H-90, 2*r, 80, rgb(0xc77a3a))
			ellipse(dst, x, b.Y+b.H-90, r, 22, rgb(0xd98a4a))
			ellipse(dst, x, b.Y+b.H-10, r, 14, rgb(0x9a5a2a))
			line(dst, x, b.Y+b.H-110, x, b.Y+20, 6, rgb(0xb06a30))
		}
		line(dst, b.X+70, b.Y+20, b.X+200, b.Y+20, 6, rgb(0xb06a30))
		if work {
			for k := 0; k < 3; k++ {
				ph := math.Mod(t*0.6+float64(k)/3, 1)
				circle(dst, b.X+200, b.Y+10-ph*60, 8+ph*10, rgba(0xeeeeee, 0.6*(1-ph)))
			}
		}
	case "whisky", "vodka", "alcool_poire":
		// un alambic de cuivre : la chaudière, le chapiteau, le col de cygne, le serpentin
		ellipse(dst, b.X+80, b.Y+b.H-40, 62, 36, rgb(0xc77a3a))
		ellipse(dst, b.X+80, b.Y+b.H-60, 44, 20, rgb(0xe09a5a))
		circle(dst, b.X+80, b.Y+50, 26, rgb(0xd98a4a))
		line(dst, b.X+80, b.Y+70, b.X+80, b.Y+b.H-70, 10, rgb(0xb06a30))
		line(dst, b.X+100, b.Y+40, b.X+180, b.Y+60, 7, rgb(0xb06a30))
		rect(dst, b.X+170, b.Y+60, 70, 70, rgb(0x7a8a96))
		for k := 0; k < 4; k++ {
			line(dst, b.X+175, b.Y+70+float64(k)*15, b.X+235, b.Y+76+float64(k)*15, 3, rgb(0xc77a3a))
		}
		roundRect(dst, b.X+186, b.Y+b.H-24, 40, 22, 4, rgb(0x5a3a22)) // le tonnelet
		if work {
			circle(dst, b.X+80, b.Y+b.H-14, 10+3*math.Sin(t*8), rgba(0xff8a2a, 0.8)) // le feu
		}
	case "cidre":
		// des tonneaux debout
		for k := 0; k < 3; k++ {
			x := b.X + 50 + float64(k)*80
			roundRect(dst, x-30, b.Y+40, 60, b.H-50, 18, rgb(0xa86a3a))
			line(dst, x-30, b.Y+70, x+30, b.Y+70, 4, rgb(0x5a3a22))
			line(dst, x-30, b.Y+b.H-40, x+30, b.Y+b.H-40, 4, rgb(0x5a3a22))
		}
		drawItem(dst, sim.Pomme, b.X+b.W-30, b.Y+30, 16)
	default: // pressoirs (jus de raisin, jus de pomme)
		fruit := sim.Raisin
		if s.ID == "jus_pomme" {
			fruit = sim.Pomme
		}
		roundRect(dst, b.X+30, b.Y+60, b.W-60, b.H-70, 14, rgb(0x8a5a33))
		for k := 0; k < 5; k++ {
			drawItem(dst, fruit, b.X+60+float64(k)*36, b.Y+70, 15)
		}
		py := b.Y + 20 + math.Abs(math.Sin(t*2))*10*boolF(work)
		rect(dst, b.C().X-4, b.Y, 8, py-b.Y+10, rgb(0x5a5a5a))
		roundRect(dst, b.X+50, py+10, b.W-100, 14, 4, rgb(0x6e4322))
		drawItem(dst, s.Item, b.X+b.W-40, b.Y+b.H-30, 18)
	}
}

func boolF(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

// drawDomainePlant : le houblon (une perche, une liane, des cônes) et la vigne
// (un piquet, des feuilles, des grappes) ; faux si ce n'est pas l'un d'eux.
func drawDomainePlant(dst *ebiten.Image, it sim.Item, x, y, gr, t float64) bool {
	ripe := gr >= 1
	sway := 0.0
	if ripe {
		sway = math.Sin(t*2) * 2
	}
	switch it {
	case sim.Houblon:
		h := 20 + 50*gr
		line(dst, x, y, x, y-h-10, 3, rgb(0x9a6b3e))
		for k := 0; k < 5; k++ {
			yy := y - float64(k)*h/5
			ellipse(dst, x+float64(k%2*2-1)*7+sway, yy-6, 8, 5, rgb(0x3f9b35))
		}
		if gr > 0.5 {
			c := lerpCol(0x9fdc6a, 0xd8e86a, gr)
			for k := 0; k < 3; k++ {
				ellipse(dst, x+float64(k%2*2-1)*9+sway, y-h*0.4-float64(k)*h*0.2, 5, 7, c)
			}
		}
	case sim.Raisin:
		h := 16 + 30*gr
		rect(dst, x-2, y-h-6, 4, h+6, rgb(0x6e4322))
		line(dst, x-24, y-h, x+24, y-h, 2, rgb(0x8a7a6a)) // le fil
		for k := -1; k <= 1; k++ {
			circle(dst, x+float64(k)*14+sway, y-h-4, 9, rgb(0x4a9a3a))
		}
		if gr > 0.4 {
			c := lerpCol(0xa8d86a, 0x6a2a6e, (gr-0.4)/0.6)
			for k := 0; k < 6; k++ {
				circle(dst, x-6+float64(k%3)*6+sway, y-h+8+float64(k/3)*7, 4, c)
			}
			circle(dst, x+sway, y-h+20, 4, c)
		}
	default:
		return false
	}
	if ripe {
		a := math.Max(0, math.Sin(t*3))
		circle(dst, x+12, y-50, 3*a, rgba(0xffffff, 0.8*a))
	}
	return true
}

// drawBottle : une bouteille (verre, étiquette, bouchon).
func drawBottle(dst *ebiten.Image, x, y, r float64, glass, label, cap uint32) {
	roundRect(dst, x-r*0.32, y-r*0.35, r*0.64, r*1.25, r*0.15, rgb(glass))
	rect(dst, x-r*0.12, y-r*0.85, r*0.24, r*0.55, rgb(glass))
	rect(dst, x-r*0.14, y-r*0.98, r*0.28, r*0.18, rgb(cap))
	roundRect(dst, x-r*0.28, y+r*0.05, r*0.56, r*0.4, r*0.06, rgb(label))
	line(dst, x-r*0.2, y-r*0.25, x-r*0.2, y+r*0.75, r*0.06, rgba(0xffffff, 0.35))
}

// drawItemDomaine : les icônes du domaine (faux si ce n'en est pas une).
func drawItemDomaine(dst *ebiten.Image, it sim.Item, x, y, r float64) bool {
	switch it {
	case sim.Houblon:
		ellipse(dst, x, y, r*0.45, r*0.7, rgb(0x9fdc6a))
		for k := 0; k < 3; k++ {
			line(dst, x-r*0.4, y-r*0.4+float64(k)*r*0.35, x+r*0.4, y-r*0.3+float64(k)*r*0.35, r*0.06, rgb(0x6aa83f))
		}
		line(dst, x, y-r*0.7, x+r*0.2, y-r, r*0.08, rgb(0x4a7a2a))
	case sim.Raisin:
		line(dst, x, y-r*0.6, x+r*0.15, y-r*0.95, r*0.08, rgb(0x6e4322))
		ellipse(dst, x+r*0.35, y-r*0.65, r*0.3, r*0.16, rgb(0x4a9a3a))
		for k := 0; k < 9; k++ {
			row := []int{0, 0, 0, 1, 1, 1, 2, 2, 3}[k]
			col := []float64{-1, 0, 1, -0.5, 0.5, 1.5, 0, 1, 0.5}[k]
			circle(dst, x+(col-0.5)*r*0.32, y-r*0.4+float64(row)*r*0.3, r*0.18, rgb(0x6a2a6e))
		}
	case sim.JusPomme:
		// une bouteille en verre, couleur du jus (jaune orangé), une pomme sur l'étiquette
		drawBottle(dst, x, y, r, 0xf0a530, 0xfff6e0, 0x2f8a4a)
		drawItem(dst, sim.Pomme, x, y+r*0.25, r*0.22)
	case sim.JusRaisin:
		c := uint32(0x6a1f4e)
		// une brique de jus
		roundRect(dst, x-r*0.45, y-r*0.6, r*0.9, r*1.4, r*0.08, rgb(0xffffff))
		rect(dst, x-r*0.45, y-r*0.1, r*0.9, r*0.5, rgb(c))
		polygon(dst, []V{{X: x - r*0.45, Y: y - r*0.6}, {X: x, Y: y - r*0.9}, {X: x + r*0.45, Y: y - r*0.6}}, rgb(0xeeeeee))
		drawItem(dst, sim.Raisin, x, y+r*0.15, r*0.35)
	case sim.VinBlanc:
		drawBottle(dst, x, y, r, 0x6f9a4a, 0xf3ead8, 0xd9a441)
	case sim.VinRouge:
		drawBottle(dst, x, y, r, 0x3a1f2a, 0xf3ead8, 0x8a1f3a)
	case sim.Biere:
		// une chope mousseuse
		roundRect(dst, x-r*0.45, y-r*0.4, r*0.8, r*1.2, r*0.1, rgb(0xf0b030))
		circle(dst, x-r*0.25, y-r*0.45, r*0.25, rgb(0xffffff))
		circle(dst, x+r*0.05, y-r*0.5, r*0.27, rgb(0xffffff))
		circle(dst, x+r*0.25, y-r*0.42, r*0.2, rgb(0xffffff))
		roundRect(dst, x+r*0.32, y-r*0.15, r*0.3, r*0.6, r*0.12, rgb(0xd9a441))
	case sim.Cidre:
		// une bouteille de cidre bouché : col vert, bouchon de liège et son muselet, le cidre
		// brun foncé dans le ventre, une étiquette orange avec une hermine
		glass, cider, dark := rgb(0x4e6b3a), rgb(0x45361a), rgb(0x332812)
		shape := []V{{X: -0.12, Y: -0.84}, {X: 0.12, Y: -0.84}, {X: 0.13, Y: -0.45}, {X: 0.2, Y: -0.25}, {X: 0.31, Y: 0},
			{X: 0.33, Y: 0.2}, {X: 0.33, Y: 0.86}, {X: 0.29, Y: 0.95}, {X: -0.29, Y: 0.95}, {X: -0.33, Y: 0.86},
			{X: -0.33, Y: 0.2}, {X: -0.31, Y: 0}, {X: -0.2, Y: -0.25}, {X: -0.13, Y: -0.45}}
		var ps []V
		for _, q := range shape {
			ps = append(ps, V{X: x + q.X*r, Y: y + q.Y*r})
		}
		polygon(dst, ps, cider)
		polygon(dst, []V{{X: x + r*0.12, Y: y - r*0.84}, {X: x + r*0.13, Y: y - r*0.45}, {X: x + r*0.2, Y: y - r*0.25},
			{X: x + r*0.31, Y: y}, {X: x + r*0.33, Y: y + r*0.2}, {X: x + r*0.33, Y: y + r*0.86}, {X: x + r*0.29, Y: y + r*0.95},
			{X: x + r*0.2, Y: y + r*0.95}, {X: x + r*0.22, Y: y + r*0.2}, {X: x + r*0.2, Y: y}, {X: x + r*0.08, Y: y - r*0.4}}, dark) // l'ombre
		polygon(dst, []V{{X: x - r*0.12, Y: y - r*0.84}, {X: x + r*0.12, Y: y - r*0.84}, {X: x + r*0.13, Y: y - r*0.45},
			{X: x + r*0.17, Y: y - r*0.32}, {X: x - r*0.17, Y: y - r*0.32}, {X: x - r*0.13, Y: y - r*0.45}}, glass) // le col, vert
		roundRect(dst, x-r*0.1, y-r*0.99, r*0.2, r*0.17, r*0.03, rgb(0xc9a86a)) // le bouchon
		rect(dst, x-r*0.14, y-r*0.84, r*0.28, r*0.05, rgb(0x3a5a2e))            // la bague
		line(dst, x-r*0.08, y-r*0.97, x-r*0.1, y-r*0.72, r*0.02, rgb(0xd8d8d8)) // le muselet
		line(dst, x+r*0.08, y-r*0.97, x+r*0.1, y-r*0.72, r*0.02, rgb(0xd8d8d8))
		line(dst, x-r*0.12, y-r*0.72, x+r*0.12, y-r*0.72, r*0.02, rgb(0xd8d8d8))
		ring(dst, x, y-r*0.78, r*0.035, r*0.015, rgb(0xd8d8d8))
		line(dst, x-r*0.05, y-r*0.75, x-r*0.05, y-r*0.5, r*0.035, rgba(0xc8e0b0, 0.6)) // reflets
		line(dst, x-r*0.2, y-r*0.05, x-r*0.24, y+r*0.35, r*0.045, rgba(0xb8d0a0, 0.55))
		rect(dst, x-r*0.33, y+r*0.42, r*0.66, r*0.36, rgb(0xd98a48)) // l'étiquette
		rect(dst, x+r*0.22, y+r*0.42, r*0.11, r*0.36, rgb(0xc07838))
		line(dst, x-r*0.2, y+r*0.5, x+r*0.2, y+r*0.5, r*0.045, rgb(0x5a2418)) // « CIDRE »
		hx, hy := x, y+r*0.65                                                 // l'hermine
		line(dst, hx, hy-r*0.08, hx, hy+r*0.09, r*0.035, rgb(0x3a1a10))
		line(dst, hx-r*0.07, hy-r*0.01, hx+r*0.07, hy-r*0.01, r*0.035, rgb(0x3a1a10))
		polygon(dst, []V{{X: hx - r*0.05, Y: hy + r*0.11}, {X: hx + r*0.05, Y: hy + r*0.11}, {X: hx, Y: hy + r*0.03}}, rgb(0x3a1a10))
	case sim.AlcoolPoire:
		drawBottle(dst, x, y, r, 0xdff0f4, 0xf3ead8, 0x2b2b2b)
		ellipse(dst, x, y+r*0.25, r*0.16, r*0.22, rgb(0xc9d24a)) // la poire dans la bouteille
	case sim.Vodka:
		drawBottle(dst, x, y, r, 0xeef6fb, 0x2b6cd6, 0xc0c0c0)
	case sim.Whisky:
		// une bouteille carrée, ambrée
		roundRect(dst, x-r*0.4, y-r*0.3, r*0.8, r*1.2, r*0.06, rgb(0xb5651d))
		rect(dst, x-r*0.12, y-r*0.8, r*0.24, r*0.5, rgb(0xb5651d))
		rect(dst, x-r*0.15, y-r*0.95, r*0.3, r*0.18, rgb(0x2b2b2b))
		roundRect(dst, x-r*0.32, y+r*0.05, r*0.64, r*0.38, r*0.04, rgb(0x2b2b2b))
		line(dst, x-r*0.22, y+r*0.24, x+r*0.22, y+r*0.24, r*0.05, rgb(0xd9a441))
	default:
		return false
	}
	return true
}

// drawTruckTop : un camion vu du dessus, qui roule vers le nord (route verticale) :
// la cabine devant, la remorque derrière ; d'où il vient est écrit sur son toit.
func drawTruckTop(dst *ebiten.Image, cx, cy float64, busy bool) {
	drawTruckTopDir(dst, cx, cy, busy, false)
}

// drawTruckTopDir : le camion vu de dessus ; down : il roule vers le sud (la cabine en bas).
func drawTruckTopDir(dst *ebiten.Image, cx, cy float64, busy, down bool) {
	label := map[string]string{"ferme": "MA FERME", "domaine": "LE DOMAINE", "elevage": "ÉLEVAGE", "port": "LE PORT"}[truckFrom]
	if label == "" {
		label = "MA FERME"
	}
	// (y, hauteur) d'un morceau du camion, retourné s'il roule vers le sud
	Y := func(dy, h float64) float64 {
		if down {
			return cy - dy - h
		}
		return cy + dy
	}
	roundRect(dst, cx-44, Y(-118, 250), 104, 250, 12, rgba(0x000000, 0.22)) // ombre
	for _, wy := range []float64{-80, 50, 92} {                             // les roues, qui dépassent un peu
		roundRect(dst, cx-56, Y(wy, 30), 12, 30, 4, rgb(0x2b2b2b))
		roundRect(dst, cx+44, Y(wy, 30), 12, 30, 4, rgb(0x2b2b2b))
	}
	// la remorque et son toit
	roundRect(dst, cx-48, Y(-50, 180), 96, 180, 8, rgb(0xd9ccb2))
	roundRect(dst, cx-44, Y(-46, 172), 88, 172, 6, rgb(0xf3ead8))
	for y := -30.0; y < 120; y += 24 {
		yy := Y(y, 0)
		line(dst, cx-40, yy, cx+40, yy, 1.5, rgba(0xb9ab90, 0.6))
	}
	roundRect(dst, cx-42, Y(22, 34), 84, 34, 6, rgba(0xffffff, 0.9))
	sayFit(dst, label, cx, Y(22, 34)+17, 16, true, 0, 80, rgb(0x2fa86b))
	// la cabine, devant
	roundRect(dst, cx-44, Y(-112, 58), 88, 58, 12, rgb(0x2fa86b))
	roundRect(dst, cx-36, Y(-108, 18), 72, 18, 6, rgb(0x9fd3ff)) // le pare-brise
	rect(dst, cx-44, Y(-112, 58), 14, 58, rgb(0x46c283))
	if busy {
		circle(dst, cx, Y(136, 0), 7, rgb(0xffd23f))
	}
}
