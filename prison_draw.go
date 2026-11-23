package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
	"ville/tr"
)

// La prison : deux quartiers de cellules (à gauche et à droite), au centre la direction, la psy et
// l'assistant social ; sous le grand couloir, la cantine, la salle de bain, les salles des
// familles, la salle de sport et la salle de travail. Sous la route : la zone du joueur et le
// garage des fourgons.

const prisOrange = 0xf28c28

// prisonerLook : la combinaison orange des détenus.
func prisonerLook(n int) look {
	lk := lookFrom(n)
	lk.shirt, lk.pants, lk.skirt = prisOrange, prisOrange, 0
	lk.hat = 0
	return lk
}

// guardLook : les surveillants (chemise bleu gris, casquette).
func guardLook(n int) look {
	lk := lookFrom(800 + n*13)
	lk.shirt, lk.pants, lk.hat, lk.hatC, lk.badge = 0x5f6f86, 0x2b3240, 2, 0x2b3240, true
	lk.skirt = 0
	return lk
}

func (g *game) prisonGround(dst *ebiten.Image, p *sim.Place, tl, br V) {
	gs := g.sim
	lv := func(k string) int { return gs.Levels["prison."+k] }
	road(dst, tl, br, 1640)
	door := V{X: p.Gate.X, Y: sim.PrisBldY + sim.PrisBldH}
	rect(dst, door.X-90, door.Y, 180, 1624-door.Y, rgb(0xd6d2c8))
	// le garage des fourgons, sous la route
	b1, bn := sim.PrisBay(1), sim.PrisBay(sim.PrisVanN)
	gx0, gx1 := b1.X-sim.PrisGarageM, bn.X+bn.W+sim.PrisGarageM
	rect(dst, (gx0+gx1)/2-110, 1790, 220, sim.PrisParkY-1790, rgb(0x55595e))
	slab(dst, gx0, sim.PrisParkY, gx1-gx0, sim.PrisGarageH, 14, 6, 0x5d6166)
	for k := 1; k <= sim.PrisVanN; k++ {
		b := sim.PrisBay(k)
		rect(dst, b.X-4, b.Y, 8, b.H, rgb(0xf2f2e6))
		rect(dst, b.X+b.W-4, b.Y, 8, b.H, rgb(0xf2f2e6))
		say(dst, "FOURGON", b.C().X, b.Y+b.H-22, 18, true, 0, rgba(0xf2f2e6, 0.8))
		if v := g.prisVan(p, k); v != nil && v.T > 0 {
			say(dst, "En route", b.C().X, b.C().Y, 20, true, 0, rgba(0xffffff, 0.85))
		}
	}
	// le bâtiment, entouré de son mur
	bx, by, bw, bh := sim.PrisBldX, sim.PrisBldY, sim.PrisBldW, sim.PrisBldH
	slab(dst, bx-30, by-30, bw+60, bh+60, 10, 18, 0x8a8478) // (le mur d'enceinte)
	slab(dst, bx, by, bw, bh, 8, 12, 0x9a9a92)
	// la cour, entre les quartiers et la direction
	for z := 0; z < 2; z++ {
		zr := sim.PrisZone(z)
		d := sim.PrisDirector()
		x0, x1 := zr.X+zr.W+20+sim.PrisPassW+30, d.X-30
		if z == 1 {
			x0, x1 = d.X+d.W+30, zr.X-20-sim.PrisPassW-30
		}
		if x1 > x0 {
			roundRect(dst, x0, zr.Y, x1-x0, sim.PrisMainY()-zr.Y-30, 20, rgb(0xd8c99a))
			say(dst, "COUR", (x0+x1)/2, zr.Y+zr.H/2, 22, true, 0, rgba(0x6e5a2a, 0.6))
		}
	}
	// le grand couloir, les couloirs des quartiers, leurs passages
	floor := rgb(0xd9d6ce)
	rect(dst, bx+40, sim.PrisMainY(), bw-80, sim.PrisMainH, floor)
	for z := 0; z < 2; z++ {
		c := sim.PrisZoneCor(z)
		rect(dst, c.X-10, c.Y+10, c.W+20, c.H-20, floor)
		ps := sim.PrisPass(z)
		rect(dst, ps.X, ps.Y+10, ps.W, ps.H, floor)
		x0, x1 := math.Min(c.X+c.W, ps.X), math.Max(c.X, ps.X+ps.W)
		rect(dst, math.Min(x0, x1)-10, c.Y+10, math.Abs(x1-x0)+20, c.H-20, floor)
		lbl := "QUARTIER A"
		if z == 1 {
			lbl = "QUARTIER B"
		}
		say(dst, lbl, c.C().X, c.C().Y-60, 20, true, 0, rgba(0x4a4a4a, 0.5))
	}
	// les cellules
	for n := 1; n <= sim.PrisCellN; n++ {
		r := sim.PrisCell(n)
		c := rgb(0xc4c0b4)
		if lv("cellule_"+itoa(n)) == 0 {
			c = rgba(0xc4c0b4, 0.35)
		}
		rect(dst, r.X, r.Y, r.W, r.H, c)
	}
	// la colonne du centre
	d, psy, as := sim.PrisDirector(), sim.PrisPsy(), sim.PrisAS()
	rect(dst, d.X, d.Y, d.W, d.H, rgb(0xc9a477))
	for y := d.Y; y < d.Y+d.H; y += 36 {
		rect(dst, d.X, y, d.W, 2, rgba(0x8a6a44, 0.35))
	}
	say(dst, "DIRECTION", d.C().X, d.Y+24, 20, true, 0, rgba(0x4a3a22, 0.75))
	for i, r := range []sim.Rect{psy, as} {
		c := rgb(0xb7c9b0)
		if (i == 0 && lv("psy") == 0) || (i == 1 && lv("as") == 0) {
			c = rgba(0xb7c9b0, 0.45)
		}
		rect(dst, r.X, r.Y, r.W, r.H, c)
		say(dst, []string{"PSY", "ASSISTANT SOCIAL"}[i], r.C().X, r.Y+24, 18, true, 0, rgba(0x2b4a2b, 0.7))
	}
	// les pièces du bas
	names := []string{"CANTINE", "SALLE DE BAIN", "FAMILLES", "SALLE DE SPORT", "SALLE DE TRAVAIL"}
	keys := []string{"cantine", "bains", "", "sport", "travail"}
	cols := []uint32{0xe8d9b5, 0xcfe3ef, 0xe9d3c7, 0xc9d8c0, 0xd8d0c0}
	for i := range names {
		r := sim.PrisRoom(i)
		c := rgb(cols[i])
		if keys[i] != "" && lv(keys[i]) == 0 {
			c = rgba(cols[i], 0.35)
		}
		if i == 2 { // les salles des familles : chacune la sienne
			for k := 1; k <= sim.PrisFamilleN; k++ {
				f := sim.PrisFamily(k)
				fc := rgb(cols[i])
				if lv("famille_"+itoa(k)) == 0 {
					fc = rgba(cols[i], 0.35)
				}
				rect(dst, f.X, f.Y, f.W, f.H, fc)
			}
		} else {
			rect(dst, r.X, r.Y, r.W, r.H, c)
		}
		say(dst, names[i], r.C().X, r.Y+r.H-22, 18, true, 0, rgba(0x3a3a3a, 0.6))
	}
	// la case des améliorations (au milieu du grand couloir)
	if gs.HasWarden() {
		r := sim.PrisUpZone()
		in := gs.Cur == p && r.Has(gs.Player.Pos)
		zoneBox(dst, r, in, g.t, 0x9fd3ff)
		say(dst, "AMÉLIORATIONS", r.C().X, r.C().Y-14, 18, true, 0, rgb(0x2b3a5a))
		say(dst, "Reste ici", r.C().X, r.C().Y+18, 15, false, 0, rgba(0x2b3a5a, 0.85))
	}
	// l'enseigne
	tw := textWidth("PRISON", 26, true) + 60
	roundRect(dst, door.X-tw/2, door.Y+36, tw, 42, 10, rgb(0x5d6d7e))
	say(dst, "PRISON", door.X, door.Y+57, 26, true, 0, rgb(0xffffff))
}

func (g *game) prisVan(p *sim.Place, k int) *sim.Van {
	if p.Pris == nil || k < 1 || k > sim.PrisVanN {
		return nil
	}
	return &p.Pris.Vans[k-1]
}

// prisonDrawables : le mobilier, le personnel, les détenus, les fourgons.
func (g *game) prisonDrawables(dst *ebiten.Image, p *sim.Place, ds []drawable) []drawable {
	gs := g.sim
	t := g.t
	lv := func(k string) int { return gs.Levels["prison."+k] }
	add := func(x, y float64, fn func()) { ds = append(ds, drawable{x: x, y: y, fn: fn}) }
	person := func(pos V, lk look) { add(pos.X, pos.Y, func() { drawPerson(dst, pos, V{Y: 1}, 0, false, lk, nil) }) }
	desk := func(x, y float64) {
		add(x, y+60, func() {
			roundRect(dst, x-110, y, 220, 60, 6, rgb(0x6e4322))
			roundRect(dst, x-80, y+10, 60, 36, 4, rgb(0xf3ead8))
		})
	}
	// la direction, la psy, l'assistant social
	d := sim.PrisDirector()
	dx, dy := d.C().X-60, d.Y+160
	desk(dx, dy)
	if gs.HasWarden() {
		lk := chiefLook()
		lk.vest, lk.shirt = 0x4a4a4a, 0x6a6a6a
		person(V{X: dx, Y: dy - 12}, lk)
	}
	for i, r := range []sim.Rect{sim.PrisPsy(), sim.PrisAS()} {
		x, y := r.C().X-50, r.Y+120
		desk(x, y)
		if (i == 0 && lv("psy") > 0) || (i == 1 && lv("as") > 0) {
			lk := lookFrom(5000 + i*77)
			lk.woman = i == 0
			lk.shirt = []uint32{0xb07ab8, 0x5a8a6a}[i]
			person(V{X: x, Y: y - 12}, lk)
		}
	}
	// les cellules : barreaux côté couloir ; le mobilier selon le niveau
	for n := 1; n <= sim.PrisCellN; n++ {
		r := sim.PrisCell(n)
		if lv("cellule_"+itoa(n)) == 0 {
			continue
		}
		top := ((n-1)%10)%2 == 0 // (la cellule du haut de sa colonne : barreaux en bas)
		l := lv("niv_cellule_" + itoa(n))
		add(r.C().X, r.Y+10, func() {
			by := r.Y + 14
			if !top {
				by = r.Y + r.H - 70
			}
			// le lit (plus beau avec le 1er niveau), puis le reste
			bc := uint32(0x8a8a8a)
			if l >= 1 {
				bc = 0x4a78c2
			}
			roundRect(dst, r.X+12, by, 60, 56, 6, rgb(bc))
			rect(dst, r.X+16, by+4, 52, 14, rgb(0xf2f2f2))
			items := []func(x, y float64){
				nil,
				func(x, y float64) { roundRect(dst, x, y, 24, 18, 6, rgb(0xe8eef2)) },                                     // lavabo
				func(x, y float64) { circle(dst, x+12, y+10, 11, rgb(0xf2f2f2)) },                                         // toilettes
				func(x, y float64) { rect(dst, x, y, 26, 40, rgb(0x8a5a33)) },                                             // armoire
				func(x, y float64) { roundRect(dst, x, y, 18, 26, 4, rgb(0xbfe3f5)) },                                     // miroir
				func(x, y float64) { roundRect(dst, x-6, y, 40, 22, 6, rgb(0xc0392b)) },                                   // tapis
				func(x, y float64) { rect(dst, x, y, 30, 22, rgb(0xa47447)) },                                             // table
				func(x, y float64) { rect(dst, x+6, y, 16, 16, rgb(0x6e4322)) },                                           // chaise
				func(x, y float64) { rect(dst, x, y, 30, 20, rgb(0x2b2b2b)); rect(dst, x+3, y+3, 24, 14, rgb(0x5b9bd5)) }, // télé
				func(x, y float64) { roundRect(dst, x, y, 26, 14, 5, rgb(0x3a3a3a)) },                                     // console
			}
			for k := 1; k < len(items) && k <= l; k++ {
				x := r.X + 84 + float64((k-1)%3)*28
				y := by + float64((k-1)/3)*22 - 6
				if !top {
					y = by - float64((k-1)/3)*22 + 30
				}
				items[k](x, y)
			}
		})
		add(r.C().X, r.Y+r.H, func() { // les barreaux
			y := r.Y + r.H - 8
			if !top {
				y = r.Y
			}
			rect(dst, r.X, y, r.W, 8, rgb(0x3a3f47))
			for x := r.X + 8; x < r.X+r.W; x += 20 {
				rect(dst, x, y-46+map[bool]float64{true: 0, false: 54}[!top], 4, 46, rgb(0x3a3f47))
			}
		})
	}
	// les surveillants, chacun devant ses deux cellules
	for k := 1; k <= sim.PrisCellN/2; k++ {
		if lv("surveillant_"+itoa(k)) > 0 {
			person(sim.PrisGuardPos(k), guardLook(k))
		}
	}
	// la cantine
	for k := 1; k <= sim.PrisTableN; k++ {
		if lv("table_"+itoa(k)) > 0 {
			r := sim.PrisTable(k)
			add(r.C().X, r.Y+r.H, func() { roundRect(dst, r.X, r.Y+30, r.W, r.H-60, 10, rgb(0xa47447)) })
		}
	}
	// la salle de bain : les cabines
	for k := 1; k <= sim.PrisCabineN; k++ {
		if lv("cabine_"+itoa(k)) > 0 {
			r := sim.PrisCabin(k)
			add(r.C().X, r.Y+40, func() {
				roundRect(dst, r.X, r.Y, r.W, 40, 6, rgb(0xe8eef2))
				circle(dst, r.C().X, r.Y+20, 10, rgb(0x9fc9e0))
				rect(dst, r.X, r.Y, 4, r.H, rgba(0x7a9ab0, 0.8))
			})
		}
	}
	// la salle de sport
	for k := 1; k <= sim.PrisAgresN; k++ {
		if lv("agres_"+itoa(k)) > 0 {
			r := sim.PrisAgresAt(k)
			k := k
			add(r.C().X, r.Y+70, func() { drawAgres(dst, k, r) })
		}
	}
	// les salles des familles : une table, des chaises
	for k := 1; k <= sim.PrisFamilleN; k++ {
		if lv("famille_"+itoa(k)) > 0 {
			r := sim.PrisFamily(k)
			add(r.C().X, r.Y+200, func() { roundRect(dst, r.C().X-45, r.Y+150, 90, 50, 8, rgb(0xa47447)) })
		}
	}
	// la salle de travail : des établis
	for k := 1; k <= sim.PrisPosteN; k++ {
		if lv("poste_"+itoa(k)) > 0 {
			r := sim.PrisPoste(k)
			add(r.C().X, r.Y+60, func() {
				roundRect(dst, r.X+5, r.Y+10, r.W-10, 50, 6, rgb(0x8a6a44))
				rect(dst, r.X+20, r.Y+20, 30, 14, rgb(0x7a7a7a))
			})
		}
	}
	// les détenus (et la famille qui leur rend visite)
	if p.Pris != nil {
		for i, pr := range p.Pris.Pris {
			pr, i := pr, i
			add(pr.Pos.X, pr.Pos.Y, func() {
				drawPerson(dst, pr.Pos, pr.Face, pr.WalkT, pr.Moving, prisonerLook(pr.Look), nil)
			})
			if pr.Need == sim.NeedFamille && pr.State == 2 {
				for j := 0; j < 2; j++ {
					v := pr.Pos.Add(V{X: float64(j*2-1) * 40, Y: 100})
					add(v.X, v.Y, func() { drawPerson(dst, v, V{Y: -1}, 0, false, lookFrom(7000+i*11+j), nil) })
				}
			}
		}
		// les fourgons : au garage ; partis, on les voit quitter le garage puis revenir
		for k := 1; k <= sim.PrisVanN; k++ {
			if lv("fourgon_"+itoa(k)) == 0 {
				continue
			}
			v := p.Pris.Vans[k-1]
			b := sim.PrisBay(k)
			if v.T <= 0 {
				c := b.C()
				add(c.X, c.Y+90, func() { drawPrisonVanTop(dst, c.X, c.Y) })
				continue
			}
			el := sim.PrisVanTrip - v.T
			const show = 12.0
			var x float64
			switch {
			case el < show: // il part vers la gauche, sur la route
				x = b.C().X - el*450
			case v.T < show: // il revient
				x = b.C().X - v.T*450
			default:
				continue
			}
			y := 1700.0
			add(x, y+20, func() { drawPrisonVanSide(dst, x, y, -1) })
		}
	}
	_ = t
	return ds
}

func drawAgres(dst *ebiten.Image, k int, r sim.Rect) {
	cx, y := r.C().X, r.Y+20
	switch k {
	case 1: // tapis de course
		roundRect(dst, cx-60, y, 120, 40, 8, rgb(0x2b2b2b))
		rect(dst, cx-50, y+8, 100, 24, rgb(0x4a4a4a))
	case 2: // vélo
		circle(dst, cx-30, y+28, 16, rgb(0x2b2b2b))
		circle(dst, cx+30, y+28, 16, rgb(0x2b2b2b))
		line(dst, cx-30, y+28, cx+30, y+28, 5, rgb(0xc0392b))
	case 3: // banc de musculation
		roundRect(dst, cx-60, y+14, 120, 22, 6, rgb(0x2b4f8a))
		rect(dst, cx-70, y, 6, 50, rgb(0x7a7a7a))
		rect(dst, cx+64, y, 6, 50, rgb(0x7a7a7a))
	case 4: // haltères
		for i := 0; i < 3; i++ {
			x := cx - 50 + float64(i)*50
			rect(dst, x-14, y+20, 28, 6, rgb(0x7a7a7a))
			circle(dst, x-14, y+23, 8, rgb(0x2b2b2b))
			circle(dst, x+14, y+23, 8, rgb(0x2b2b2b))
		}
	case 5: // sac de frappe
		roundRect(dst, cx-18, y, 36, 70, 14, rgb(0xb03a3a))
	case 6: // panier de basket
		rect(dst, cx-4, y, 8, 50, rgb(0x7a7a7a))
		ring(dst, cx, y+6, 16, 4, rgb(0xe67e22))
	case 7: // table de ping-pong
		roundRect(dst, cx-70, y, 140, 70, 6, rgb(0x2a7a4a))
		rect(dst, cx-2, y, 4, 70, rgb(0xffffff))
	default: // corde à sauter
		ring(dst, cx, y+30, 26, 3, rgb(0xe67e22))
	}
}

// drawPrisonVanTop : le fourgon pénitentiaire garé, vu de dessus (l'avant en haut).
func drawPrisonVanTop(dst *ebiten.Image, x, y float64) {
	ellipse(dst, x+6, y+10, 72, 118, rgba(0x000000, 0.22))
	roundRect(dst, x-60, y-110, 120, 220, 18, rgb(0xe8eaee))
	roundRect(dst, x-60, y-20, 120, 40, 4, rgb(0x5d6d7e))
	roundRect(dst, x-46, y-100, 92, 30, 8, rgb(0x2b3a4a))
	for i := 0; i < 3; i++ {
		rect(dst, x-40+float64(i)*30, y+40, 18, 50, rgba(0x2b3a4a, 0.5)) // (les grilles des fenêtres)
	}
}

// drawPrisonVanSide : le fourgon sur la route, vu de profil (fx : 1 vers la droite).
func drawPrisonVanSide(dst *ebiten.Image, x, y, fx float64) {
	ellipse(dst, x, y+2, 80, 10, rgba(0x000000, 0.25))
	roundRect(dst, x-78, y-76, 156, 68, 10, rgb(0xe8eaee))
	rect(dst, x-78, y-36, 156, 12, rgb(0x5d6d7e))
	roundRect(dst, x+fx*44-18, y-68, 34, 24, 5, rgb(0xbfe3f5))
	for i := 0; i < 3; i++ {
		rect(dst, x-fx*10-30+float64(i)*20, y-66, 12, 20, rgba(0x2b3a4a, 0.6))
	}
	for _, wx := range []float64{-46, 46} {
		circle(dst, x+wx, y-8, 14, rgb(0x2b2b2b))
		circle(dst, x+wx, y-8, 6, rgb(0xb0b0b0))
	}
	say(dst, "PRISON", x-fx*10, y-18, 14, true, 0, rgb(0x5d6d7e))
}

// prisonHUD : la case des améliorations ouvre le menu.
func (g *game) prisonHUD(dst *ebiten.Image) {
	p := g.sim.Cur
	if p == nil || p.Pris == nil || !g.sim.HasWarden() {
		return
	}
	pl := g.sim.Player
	in := sim.PrisUpZone().Has(pl.Pos)
	if in && !pl.Moving && !g.prisZoneIn && g.panel == panelNone {
		g.prisZoneIn = true
		g.openPanel(panelPrison)
	}
	if !in {
		g.prisZoneIn = false
	}
}

// drawPrisonPanel : les améliorations (points d'école), une par une ; et le bilan de la prison.
func (g *game) drawPrisonPanel(dst *ebiten.Image) {
	gs := g.sim
	w, h := math.Min(1500, scr.VW-60), 960.0
	x, y := g.panelFrame(dst, "La prison", w, h)
	held, reins, recid := gs.PrisonStats()
	sayFit(dst, tr.F("Détenus : {n}. Réinsérés : {r}. Récidivistes : {c}. Voleurs en liberté : {v}.", "n", itoa(held), "r", itoa(reins), "c", itoa(recid), "v", itoa(gs.Thieves())),
		x+w/2, y+30, 20, true, 0, w-60, rgb(0x2b3a5a))
	say(dst, "Chaque niveau retire 1 point de récidive à qui y remplit son besoin (10 niveaux au plus).", x+w/2, y+64, 18, false, 0, rgb(0x5a4a3a))
	imps := gs.PrisImps()
	if len(imps) == 0 {
		say(dst, "Rien à améliorer pour l'instant : achète d'abord des cellules ou des salles.", x+w/2, y+200, 20, false, 0, rgb(0x5a4a3a))
		return
	}
	cols := 5
	bw := (w - 60 - float64(cols-1)*12) / float64(cols)
	bh := 96.0
	for i, im := range imps {
		im := im
		bx := x + 30 + float64(i%cols)*(bw+12)
		by := y + 100 + float64(i/cols)*(bh+10)
		if by+bh > y+h-110 {
			break
		}
		l := gs.Levels["prison."+im.Key]
		name := tr.T(im.Name)
		if im.N > 0 {
			name = tr.F("{nom} n° {n}", "nom", tr.T(im.Name), "n", itoa(im.N))
		}
		roundRect(dst, bx, by, bw, bh, 14, rgb(0xf1e6d2))
		sayFit(dst, name, bx+bw/2, by+20, 17, true, 0, bw-16, rgb(0x2b3a5a))
		for k := 0; k < sim.PrisLvlMax; k++ {
			c := rgba(0x2b3a5a, 0.2)
			if k < l {
				c = rgb(0x5b7fd6)
			}
			roundRect(dst, bx+12+float64(k)*(bw-24)/10, by+36, (bw-24)/10-3, 10, 3, c)
		}
		c := gs.PrisImpCost(im.Key)
		if c == 0 {
			say(dst, "Maximum", bx+bw/2, by+72, 18, true, 0, rgb(0x2fa86b))
			continue
		}
		g.btn(dst, ptsPrice(im.Type, c), bx+10, by+52, bw-20, 38, 0x2fa86b, gs.Points(im.Type) >= c, func() {
			if gs.BuyPrisImp(im) {
				g.snd.play(sBuy)
			}
		})
	}
}
