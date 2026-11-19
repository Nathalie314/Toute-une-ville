package main

import (
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// La crêperie : la cuisine (carrelage, postes avec leur chef derrière le plan de
// travail, la réserve), le comptoir dans le mur, la salle (parquet, tables, frigo
// du cidre), la terrasse dehors, les clients, les serveurs et leurs bulles.

// crepGround : le sol, les murs, les portes, la route des livraisons, la terrasse.
func (g *game) crepGround(dst *ebiten.Image, p *sim.Place, tl, br V) {
	// la route des livraisons, au-dessus de la cuisine
	ry := sim.CrepTop + 20
	rect(dst, tl.X-10, ry, br.X-tl.X+20, 130, rgb(0x6f6f6f))
	for x := math.Floor(tl.X/120) * 120; x < br.X; x += 120 {
		rect(dst, x, ry+62, 60, 6, rgb(0xf2f2f2))
	}
	// la cuisine : carrelage blanc et gris
	kx, ky := sim.CrepBX0, sim.CrepKY0
	kw, kh := sim.CrepBX1-sim.CrepBX0, sim.CrepWallY-sim.CrepKY0
	rect(dst, kx, ky, kw, kh, rgb(0xeef0f2))
	for y := ky; y < ky+kh; y += 80 {
		for x := kx; x < kx+kw; x += 80 {
			if (int((x-kx)/80)+int((y-ky)/80))%2 == 0 {
				rect(dst, x, y, math.Min(80, kx+kw-x), math.Min(80, ky+kh-y), rgb(0xdfe3e7))
			}
		}
	}
	say(dst, "RÉSERVE", kx+kw/2, sim.CrepPalY-22, 18, true, 0, rgba(0x4a4a4a, 0.7))
	say(dst, "CUISINE", kx+kw-160, sim.CrepWallY-40, 22, true, 0, rgba(0x4a4a4a, 0.45))
	// la salle : parquet
	sy := sim.CrepWallY + sim.CrepWallH
	rect(dst, kx, sy, kw, sim.CrepBotY-sy, rgb(0xc99a62))
	for y := sy; y < sim.CrepBotY; y += 46 {
		line(dst, kx, y, kx+kw, y, 1.5, rgba(0x8a5a33, 0.35))
		off := 0.0
		if int((y-sy)/46)%2 == 1 {
			off = 120
		}
		for x := kx + off; x < kx+kw; x += 240 {
			line(dst, x, y, x, math.Min(y+46, sim.CrepBotY), 1.2, rgba(0x8a5a33, 0.3))
		}
	}
	// un grand tapis breton sous les tables du milieu (pour faire joli)
	roundRect(dst, 900, 600, 1700, 420, 30, rgba(0x2b5fa8, 0.12))
	// les murs (les stations « mur_ ») : de la pierre grise, comme en Bretagne
	for _, s := range p.Stations {
		if strings.HasPrefix(s.ID, "mur_") {
			b := s.Body
			rect(dst, b.X, b.Y, b.W, b.H, rgb(0x8d8f93))
			rect(dst, b.X, b.Y, b.W, math.Min(10, b.H), rgb(0xb5b8bd))
			for x := b.X + 20; x < b.X+b.W-10; x += 60 {
				if b.H > 40 {
					line(dst, x, b.Y+12, x, b.Y+b.H-4, 1.5, rgba(0x5a5c60, 0.5))
				}
			}
		}
	}
	// portes : paillassons
	roundRect(dst, 260, sim.CrepWallY+6, 160, sim.CrepWallH-12, 10, rgb(0x6e4322)) // cuisine ↔ salle
	say(dst, "Cuisine ↕ Salle", 340, sim.CrepWallY+sim.CrepWallH/2, 15, true, 0, rgb(0xffeab0))
	roundRect(dst, sim.CrepDoorX-80, sim.CrepBotY-6, 160, 40, 10, rgb(0x6e4322)) // entrée
	roundRect(dst, sim.CrepBX1-6, sim.CrepTerDY-80, 40, 160, 10, rgb(0x6e4322))  // terrasse
	roundRect(dst, sim.CrepBX0-34, 1260, 40, 150, 10, rgb(0x6e4322))             // vers la zone du joueur
	// devant : le trottoir (on y attend une table) et l'allée jusqu'au passage piéton
	rect(dst, sim.CrepBX0-30, sim.CrepBotY+26, sim.CrepTerX1-sim.CrepBX0+60, 1640-sim.CrepBotY-26, rgb(0xd5cfc4))
	for x := sim.CrepBX0 - 30.0; x < sim.CrepTerX1+30; x += 90 {
		line(dst, x, sim.CrepBotY+26, x, 1640, 1.2, rgba(0x9a948a, 0.5))
	}
	// l'enseigne au-dessus de la porte
	roundRect(dst, sim.CrepDoorX-150, sim.CrepBotY-70, 300, 48, 10, rgb(0x1f4f8a))
	say(dst, "CRÊPERIE", sim.CrepDoorX, sim.CrepBotY-46, 26, true, 0, rgb(0xffffff))
	// la terrasse : un plancher de bois, une barrière basse (ouverte en bas à gauche)
	tx0, tx1 := sim.CrepTerX0-120, sim.CrepTerX1
	if p.TerraceOpen() {
		rect(dst, tx0, 300, tx1-tx0, sim.CrepBotY-280, rgb(0xb7895a))
		for y := 300.0; y < sim.CrepBotY+20; y += 40 {
			line(dst, tx0, y, tx1, y, 1.5, rgba(0x6e4322, 0.35))
		}
		fc := rgb(0xf2f2f2)
		line(dst, tx0, 300, tx1, 300, 5, fc)
		line(dst, tx1, 300, tx1, sim.CrepBotY+20, 5, fc)
		line(dst, tx0+120, sim.CrepBotY+20, tx1, sim.CrepBotY+20, 5, fc)
		line(dst, tx0, 300, tx0, sim.CrepTerDY-90, 5, fc)
		line(dst, tx0, sim.CrepTerDY+90, tx0, sim.CrepBotY-60, 5, fc)
		say(dst, "TERRASSE", (tx0+tx1)/2, 330, 20, true, 0, rgba(0x4a2a12, 0.6))
	} else {
		dashedRect(dst, tx0, 300, tx1-tx0, sim.CrepBotY-280, 4, 18, 0, rgba(0xffffff, 0.5))
	}
	road(dst, tl, br, 1640)
	g.backTruck(dst, p)
}

// drawCrepStation : les postes (un chef derrière son plan de travail), le comptoir,
// le frigo du cidre ; vrai si la station est dessinée ici.
func (g *game) drawCrepStation(dst *ebiten.Image, p *sim.Place, s *sim.Station) bool {
	b := s.Body
	t := g.t
	switch {
	case strings.HasPrefix(s.ID, "table_"):
		return true // (dessinées avec leurs clients)
	case s.Kind == sim.Machine:
		// le chef, derrière le plan de travail
		lk := lookFrom(len(s.ID)*131 + int(b.X))
		lk.shirt, lk.pants, lk.apron, lk.hat = 0xffffff, 0x2b2b2b, 0x1f4f8a, 3
		bob := 0.0
		if s.Working {
			bob = math.Sin(t*8) * 2
		}
		drawPerson(dst, V{X: b.C().X, Y: b.Y + 70 + bob}, V{X: 0, Y: 1}, 0, false, lk, nil)
		// le plan de travail
		roundRect(dst, b.X, b.Y+58, b.W, b.H-58, 10, rgb(0x9aa4ad))
		rect(dst, b.X, b.Y+58, b.W, 16, rgb(0xc8d0d6))
		// la (ou les) crêpière(s) : un disque noir, une crêpe dorée si ça cuit
		n := 1
		if s.Item == sim.CrepeNature {
			n = 2
		}
		for k := 0; k < n; k++ {
			cx := b.X + 70 + float64(k)*110
			cy := b.Y + 100
			ellipse(dst, cx, cy, 46, 24, rgb(0x2b2b2b))
			if s.Working || s.Pile > 0 {
				ellipse(dst, cx, cy, 38, 19, rgb(0xe8b85a))
				ellipse(dst, cx-8, cy-4, 10, 4, rgba(0xc98a2a, 0.6))
			}
		}
		// ce qu'il garnit (l'icône du plat)
		drawItem(dst, s.Item, b.X+b.W-46, b.Y+102, 20)
		if s.Pile > 0 && s.Item != sim.CrepeNature {
			say(dst, "Prêt !", b.X+b.W-46, b.Y+50, 16, true, 0, rgb(0x2fa86b)) // (le chef attend qu'on l'emporte)
		}
		sign(dst, b.C().X, b.Y+b.H-4, s.Name) // (en bas du plan de travail : plus sur la tête du chef)
		return true
	case s.Kind == sim.Shelf:
		// le comptoir (passe-plat) : une tablette d'inox dans le mur, les assiettes posées dessus
		roundRect(dst, b.X, b.Y, b.W, b.H, 8, rgb(0xb0b8c0))
		rect(dst, b.X, b.Y, b.W, 12, rgb(0xdfe5ea))
		for k := 0; k < s.Stock && k < 6; k++ {
			x := b.X + 24 + float64(k%3)*56
			y := b.Y + 30 + float64(k/3)*22
			ellipse(dst, x, y, 24, 11, rgb(0xffffff))
			drawItem(dst, s.Item, x, y-4, 11)
		}
		return true
	case s.Kind == sim.Pallet && s.Item == sim.Cidre:
		// le frigo du cidre : vitré, des bouteilles dedans
		roundRect(dst, b.X, b.Y, b.W, b.H, 8, rgb(0xeef3f7))
		roundRect(dst, b.X+10, b.Y+10, b.W-20, b.H-20, 6, rgb(0xbfe3f2))
		n := min(14, s.Pile)
		for k := 0; k < n; k++ {
			drawItem(dst, sim.Cidre, b.X+24+float64(k%7)*24, b.Y+30+float64(k/7)*28, 10)
		}
		sayOutlined(dst, itoa(s.Pile), b.X+b.W-20, b.Y+b.H-14, 16, 0, rgb(0xffffff), rgb(0x2b2b2b))
		sign(dst, b.C().X, b.Y-16, s.Name)
		return true
	}
	return false
}

// crepDrawables : les tables (avec chaises, assiettes, pourboires), les clients,
// les serveurs, et les bulles des tables.
func (g *game) crepDrawables(dst *ebiten.Image, p *sim.Place, ds []drawable) []drawable {
	t := g.t
	for _, tb := range p.Crep.Tables {
		tb := tb
		if !tb.St.Built {
			continue
		}
		b := tb.Body()
		ds = append(ds, drawable{x: tb.Pos.X, y: b.Y + b.H/2, fn: func() { drawCrepTable(dst, tb, t) }})
		if tb.Terrace && tb.Seats == 4 {
			ds = append(ds, drawable{x: tb.Pos.X, y: b.Y + b.H + 1, fn: func() { drawParasol(dst, tb.Pos, t) }})
		}
	}
	for _, cg := range p.Crep.Groups {
		cg := cg
		if !cg.Visible() {
			continue
		}
		for i, m := range cg.Members {
			m, i := m, i
			ds = append(ds, drawable{x: m.Pos.X, y: m.Pos.Y, fn: func() {
				lk := lookFrom(m.Look)
				if cg.Seated() {
					lk.sit, lk.crate = true, false
				}
				drawPerson(dst, m.Pos, m.Face, m.WalkT, m.Moving, lk, nil)
				if i == 0 && cg.InQueue() && cg.QueueT() > 45 {
					face(dst, m.Pos.X+24, m.Pos.Y-110, 0xf5c518, 1) // (on s'impatiente devant la porte)
				}
			}})
		}
		if cg.Seated() && cg.Table != nil {
			tb := cg.Table
			ds = append(ds, drawable{x: tb.Pos.X, y: tb.Pos.Y + 400, fn: func() { drawCrepBubble(dst, cg, tb, t) }})
		}
	}
	for _, w := range p.CrepWaiters() {
		w := w
		ds = append(ds, drawable{x: w.Pos.X, y: w.Pos.Y, fn: func() { drawAgent(dst, p.ID, w, t) }})
	}
	return ds
}

// drawCrepTable : la table, ses chaises, ce qui est servi, la vaisselle sale, les pourboires.
func drawCrepTable(dst *ebiten.Image, tb *sim.CrepTable, t float64) {
	b := tb.Body()
	for i := 0; i < tb.Seats; i++ {
		s := tb.Seat(i)
		roundRect(dst, s.X-22, s.Y-10, 44, 40, 8, rgb(0x6e4322)) // (la chaise)
		rect(dst, s.X-22, s.Y-34, 44, 12, rgb(0x5a3a1e))
	}
	ellipse(dst, b.C().X+6, b.Y+b.H+6, b.W*0.5, 12, rgba(0x000000, 0.18))
	roundRect(dst, b.X, b.Y, b.W, b.H, 16, rgb(0x8a5a33))
	roundRect(dst, b.X+6, b.Y+6, b.W-12, b.H-14, 12, rgb(0xf3ece0)) // (la nappe)
	if tb.Seats == 4 {
		line(dst, b.X+6, b.C().Y-2, b.X+b.W-6, b.C().Y-2, 2, rgba(0x2b5fa8, 0.4))
	}
	if cg := tb.Group; cg != nil {
		// ce qui est déjà servi (les assiettes devant chacun, le cidre au milieu)
		k := 0
		for _, it := range cg.Got {
			x := b.X + 30 + float64(k%4)*((b.W-60)/3)
			y := b.Y + 26 + float64(k/4)*30
			if it != sim.Cidre {
				ellipse(dst, x, y, 18, 9, rgb(0xffffff))
			}
			drawItem(dst, it, x, y-3, 10)
			k++
		}
	}
	if tb.Dirty {
		for k := 0; k < tb.Seats; k++ {
			x := b.X + 30 + float64(k%2)*(b.W-60)
			y := b.Y + 28 + float64(k/2)*30
			ellipse(dst, x, y, 18, 9, rgb(0xe0dcd2))
			circle(dst, x-4, y-1, 3, rgba(0x8a5a33, 0.6))
		}
		if p := tb.CleanProgress(); p > 0 {
			roundRect(dst, b.C().X-40, b.Y-24, 80, 10, 5, rgba(0x000000, 0.35))
			roundRect(dst, b.C().X-38, b.Y-22, 76*p, 6, 3, rgb(0x5aa0e0))
		}
	}
	if tb.Tips > 0 {
		// les pourboires : quelques pièces sur la table (seul le joueur les ramasse)
		for k := 0; k < 3; k++ {
			coin(dst, b.X+b.W-30-float64(k)*9, b.Y+b.H-26-float64(k)*5, 9)
		}
		sayOutlined(dst, money(tb.Tips), b.X+b.W-20, b.Y+b.H+20, 16, 0, rgb(0xffe066), rgb(0x3b2a10))
	}
}

// drawParasol : un parasol à rayures au-dessus des tables de 4 de la terrasse.
func drawParasol(dst *ebiten.Image, at V, t float64) {
	x, y := at.X, at.Y-150
	line(dst, at.X, at.Y-10, x, y, 5, rgb(0x6e4322))
	for k := 0; k < 8; k++ {
		a0 := math.Pi + float64(k)*math.Pi/8
		a1 := a0 + math.Pi/8
		c := rgb(0xffffff)
		if k%2 == 0 {
			c = rgb(0x1f4f8a)
		}
		polygon(dst, []V{{X: x, Y: y - 20}, {X: x + math.Cos(a0)*120, Y: y + 30 + math.Sin(a0)*14}, {X: x + math.Cos(a1)*120, Y: y + 30 + math.Sin(a1)*14}}, c)
	}
}

// drawCrepBubble : au-dessus de la table, ce qu'elle attend et son bonheur.
func drawCrepBubble(dst *ebiten.Image, cg *sim.CrepGroup, tb *sim.CrepTable, t float64) {
	b := tb.Body()
	x, y := tb.Pos.X, b.Y-70
	switch {
	case cg.Waiting():
		// elle attend qu'on prenne la commande : un carnet et la barre de la tablette
		roundRect(dst, x-34, y-30, 68, 54, 14, rgba(0xffffff, 0.95))
		roundRect(dst, x-16, y-20, 32, 36, 4, rgb(0xfff3c4))
		for k := 0; k < 3; k++ {
			line(dst, x-10, y-10+float64(k)*9, x+10, y-10+float64(k)*9, 2, rgb(0x9a8a6a))
		}
		if p := cg.OrderProgress(); p > 0 {
			roundRect(dst, x-30, y+28, 60, 10, 5, rgba(0x000000, 0.35))
			roundRect(dst, x-28, y+30, 56*p, 6, 3, rgb(0x2fa86b))
		}
	case cg.Eating():
		face(dst, x, y, 0x7ed957, 0)
	default:
		// ce qu'il reste à servir
		ws := cg.Wants
		n := min(len(ws), 6)
		if n == 0 {
			break
		}
		w := 20 + float64(n)*36
		roundRect(dst, x-w/2, y-26, w, 48, 16, rgba(0xffffff, 0.95))
		for k := 0; k < n; k++ {
			drawItem(dst, ws[k], x-w/2+28+float64(k)*36, y-2, 14)
		}
		if len(ws) > n {
			say(dst, "+"+itoa(len(ws)-n), x+w/2+16, y-2, 16, true, 0, rgb(0x5a4a3a))
		}
	}
	// la jauge de bonheur (elle décide du pourboire)
	gw := 70.0
	c := rgb(0x2fa86b)
	if cg.Happy < 0.35 {
		c = rgb(0xe74c3c)
	} else if cg.Happy < 0.7 {
		c = rgb(0xf5c518)
	}
	roundRect(dst, x-gw/2, y+44, gw, 10, 5, rgba(0x000000, 0.3))
	roundRect(dst, x-gw/2+2, y+46, (gw-4)*cg.Happy, 6, 3, c)
}

// drawItemCrepe : les icônes des crêpes (une crêpe pliée en triangle, sa garniture).
func drawItemCrepe(dst *ebiten.Image, it sim.Item, x, y, r float64) bool {
	crepe := func() {
		// une crêpe pliée en quatre : un quart de disque doré, le bord arrondi plus foncé
		cx, cy := x-r*0.7, y+r*0.6
		pts := []V{{X: cx, Y: cy}}
		for k := 0; k <= 12; k++ {
			a := -math.Pi/2 + float64(k)/12*math.Pi/2
			pts = append(pts, V{X: cx + math.Cos(a)*r*1.45, Y: cy + math.Sin(a)*r*1.3})
		}
		polygon(dst, pts, rgb(0xe8b85a))
		for k := 0; k < 12; k++ {
			a0 := -math.Pi/2 + float64(k)/12*math.Pi/2
			a1 := -math.Pi/2 + float64(k+1)/12*math.Pi/2
			line(dst, cx+math.Cos(a0)*r*1.4, cy+math.Sin(a0)*r*1.25, cx+math.Cos(a1)*r*1.4, cy+math.Sin(a1)*r*1.25, r*0.1, rgb(0xc98a2a))
		}
		circle(dst, cx+r*0.45, cy-r*0.5, r*0.07, rgba(0xa8681a, 0.7))
		circle(dst, cx+r*0.9, cy-r*0.25, r*0.06, rgba(0xa8681a, 0.7))
		circle(dst, cx+r*0.3, cy-r*0.95, r*0.06, rgba(0xa8681a, 0.7))
	}
	switch it {
	case sim.CrepeNature:
		// une pile de crêpes, à plat
		for k := 0; k < 3; k++ {
			ellipse(dst, x, y+r*0.35-float64(k)*r*0.22, r*0.85, r*0.32, rgb(0xe8b85a))
			ellipse(dst, x, y+r*0.35-float64(k)*r*0.22, r*0.85, r*0.32, rgba(0xc98a2a, 0.25))
		}
		ellipse(dst, x, y-r*0.1, r*0.8, r*0.28, rgb(0xf0c878))
	case sim.CrepeSucre:
		crepe()
		for k := 0; k < 6; k++ {
			circle(dst, x-r*0.35+float64(k%3)*r*0.35, y+r*0.05+float64(k/3)*r*0.25, r*0.06, rgb(0xffffff))
		}
	case sim.CrepeBananeChoco:
		crepe()
		line(dst, x-r*0.5, y+r*0.2, x+r*0.4, y-r*0.15, r*0.12, rgb(0x5a3220))
		ellipse(dst, x+r*0.1, y+r*0.3, r*0.25, r*0.12, rgb(0xf6d53b))
	case sim.CrepePommeCannelle:
		crepe()
		circle(dst, x-r*0.15, y+r*0.2, r*0.16, rgb(0xd9342b))
		line(dst, x+r*0.1, y+r*0.35, x+r*0.5, y+r*0.1, r*0.1, rgb(0xa0521d))
	case sim.CrepePoireVanille:
		crepe()
		ellipse(dst, x+r*0.35, y-r*0.25, r*0.3, r*0.24, rgb(0xfff6dc)) // la boule de glace
		ellipse(dst, x-r*0.2, y+r*0.2, r*0.15, r*0.2, rgb(0xb8d64a))
	case sim.CrepeSaucisseOeuf:
		// un visage qui sourit : deux œufs (les yeux) et une saucisse (la bouche)
		ellipse(dst, x, y, r*0.95, r*0.75, rgb(0xe8b85a))
		for _, dx := range []float64{-0.35, 0.35} {
			circle(dst, x+dx*r, y-r*0.2, r*0.22, rgb(0xffffff))
			circle(dst, x+dx*r, y-r*0.2, r*0.1, rgb(0xf5b51a))
		}
		for k := 0; k < 6; k++ {
			a := math.Pi*0.15 + float64(k)/5*math.Pi*0.7
			circle(dst, x+math.Cos(a)*r*0.45, y+r*0.05+math.Sin(a)*r*0.3, r*0.11, rgb(0xa8452a))
		}
		ellipse(dst, x-r*0.75, y+r*0.35, r*0.2, r*0.1, rgb(0x5aa83a)) // un peu de salade
	case sim.CrepeThonPoireau:
		crepe()
		ellipse(dst, x-r*0.1, y+r*0.2, r*0.3, r*0.14, rgb(0xc0d3e0))
		line(dst, x+r*0.15, y+r*0.35, x+r*0.55, y, r*0.1, rgb(0x7cbf4a))
		ellipse(dst, x-r*0.6, y+r*0.42, r*0.18, r*0.08, rgb(0x5aa83a))
	case sim.CrepeCanard:
		crepe()
		ellipse(dst, x, y+r*0.2, r*0.35, r*0.16, rgb(0x8a2a2a))
		for k := 0; k < 3; k++ { // les flammes du whisky
			fx := x - r*0.25 + float64(k)*r*0.25
			polygon(dst, []V{{X: fx - r*0.1, Y: y}, {X: fx + r*0.1, Y: y}, {X: fx, Y: y - r*0.45}}, rgb(0xff8a2a))
		}
	case sim.CrepeSteakBrebis:
		crepe()
		ellipse(dst, x, y+r*0.22, r*0.32, r*0.14, rgb(0x6e3a22))
		rect(dst, x-r*0.25, y+r*0.02, r*0.5, r*0.1, rgb(0xfff3c4))
		ellipse(dst, x-r*0.6, y+r*0.42, r*0.18, r*0.08, rgb(0x5aa83a))
	default:
		return false
	}
	return true
}
