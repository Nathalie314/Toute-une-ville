package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"ville/sim"
)

// L'hélicoptère : dans la zone du joueur de chaque lieu, à côté de l'hélisurface,
// une case (téléporteur) par destination (la carte et les autres lieux achetés).
// On reste une seconde dessus pour partir. Raccourcis : 1 à 9 (lieux), 0 (carte),
// V (mairie), P (personnage). Le vestiaire ouvre aussi la page « Personnage ».

var busOrder = []string{"ferme", "ecole", "super", "cremerie", "exotique", "boulangerie", "glacier", "elevage", "boucherie", "port", "domaine", "patisserie"} // (le port : touche 0 ; le domaine : sans touche au départ)

var busNames = map[string]string{"ferme": "Ma ferme", "ecole": "École", "super": "Supermarché", "boulangerie": "Boulangerie", "cremerie": "Crèmerie", "glacier": "Glacier", "elevage": "Élevage", "boucherie": "Boucherie", "exotique": "Exotique", "port": "Port", "domaine": "Domaine", "patisserie": "Pâtisserie"}
var busCols = map[string]uint32{"ferme": 0xc0623a, "ecole": 0xd4a017, "super": 0x2fa86b, "boulangerie": 0xa0522d, "cremerie": 0x7fb3d5, "glacier": 0x5aa0e0, "elevage": 0x8a5a2b, "boucherie": 0xb03a3a, "exotique": 0x1f8a5a, "port": 0x2b6cd6, "domaine": 0x7a2a4a, "patisserie": 0xe0567f, "": 0x5b7fd6}

type busDest struct {
	id   string // "" : la carte
	zone sim.Rect
}

func (g *game) busDests(p *sim.Place) []busDest {
	ids := []string{""}
	for _, id := range busOrder {
		if q := g.sim.Place(id); q != nil && q.Owned && q != p {
			ids = append(ids, id)
		}
	}
	var out []busDest
	for i, id := range ids {
		out = append(out, busDest{id: id, zone: p.YardTele(i)})
	}
	return out
}

func (g *game) busGo(id string) {
	g.busT = 0
	if id == "" {
		g.goMap()
		return
	}
	if q := g.sim.Place(id); q != nil && q.Owned {
		g.snd.play(sClick)
		g.goPlace(q)
	}
}

// updateBus : attente sur une case de l'arrêt, et raccourcis clavier.
func (g *game) updateBus(dt float64) {
	if g.panel != panelNone {
		return
	}
	keys := []ebiten.Key{ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4, ebiten.Key5, ebiten.Key6, ebiten.Key7, ebiten.Key8, ebiten.Key9, ebiten.Key0}
	pad := []ebiten.Key{ebiten.KeyNumpad1, ebiten.KeyNumpad2, ebiten.KeyNumpad3, ebiten.KeyNumpad4, ebiten.KeyNumpad5, ebiten.KeyNumpad6, ebiten.KeyNumpad7, ebiten.KeyNumpad8, ebiten.KeyNumpad9, ebiten.KeyNumpad0}
	for i := range keys {
		if inpututil.IsKeyJustPressed(keys[i]) || inpututil.IsKeyJustPressed(pad[i]) {
			k := itoa((i + 1) % 10)
			if g.scene == sceneMap && g.mapHover != "" {
				// sur la carte, un chiffre sur un lieu : il devient un favori
				g.sim.SetFav(k, g.mapHover)
				g.snd.play(sClick)
				g.toast("Favori : "+busNames[g.mapHover]+" sur la touche "+k, "")
				return
			}
			id := g.sim.Fav[k]
			if id == "" {
				g.toast("Touche "+k+" : aucun lieu (sur la carte, mets la main sur un lieu et appuie sur "+k+")", "")
			} else if q := g.sim.Place(id); q != nil && q.Owned {
				g.busGo(id)
			} else {
				g.toast(busNames[id]+" : pas encore acheté", "")
			}
			return
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyV) {
		g.openPanel(panelMairie)
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.openPanel(panelWardrobe)
		return
	}
	p := g.sim.Cur
	if g.scene != scenePlace || p == nil {
		g.busT = 0
		return
	}
	pl := g.sim.Player
	// vestiaire : on s'arrête sur son paillasson (une fois par passage)
	if p.Vestiaire.Has(pl.Pos) {
		if !pl.Moving && !g.vestDone {
			g.vestDone = true
			g.openPanel(panelWardrobe)
		}
	} else {
		g.vestDone = false
	}
	for _, d := range g.busDests(p) {
		if d.zone.Has(pl.Pos) && !pl.Moving {
			g.busT += dt
			if g.busT >= 1 {
				g.busGo(d.id)
			}
			return
		}
	}
	g.busT = 0
}

// drawBus : l'arrêt et ses cases.
func (g *game) drawBus(dst *ebiten.Image, p *sim.Place) {
	ds := g.busDests(p)
	if len(ds) == 0 {
		return
	}
	pl := g.sim.Player
	for _, d := range ds {
		r := d.zone
		c := busCols[d.id]
		roundRect(dst, r.X, r.Y, r.W, r.H, 14, rgba(0x000000, 0.3))
		dashedRect(dst, r.X+3, r.Y+3, r.W-6, r.H-6, 4, 10, g.t*25, rgb(c))
		name := "Carte"
		if d.id != "" {
			name = busNames[d.id]
		}
		roundRect(dst, r.X+10, r.Y+12, r.W-20, 34, 8, rgb(c))
		sayFit(dst, "→ "+name, r.C().X, r.Y+29, 16, true, 0, r.W-24, rgb(0xffffff))
		key := "Tab"
		if d.id != "" {
			key = g.sim.FavKey(d.id)
		}
		if key != "" {
			say(dst, "touche "+key, r.C().X, r.Y+r.H-26, 14, false, 0, rgb(0xffffff))
		}
		if o := g.sim.Order; o != nil && d.id != "" && o.Place() == d.id {
			badge(dst, r.X+r.W-6, r.Y+6, g.t) // la commande est à livrer là-bas
		}
		if r.Has(pl.Pos) && g.busT > 0 {
			f := math.Min(1, g.busT)
			roundRect(dst, r.X, r.Y+r.H*(1-f), r.W, r.H*f, 14, rgba(c, 0.45))
		}
	}
}
