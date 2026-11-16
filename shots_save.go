package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"ville/sim"
)

// Captures d'une vraie sauvegarde (développement) : CAPSAVE=fichier.json et
// -captures dossier. On charge la partie, on la laisse tourner un peu
// (CAPSEC secondes, 20 par défaut), puis on photographie chaque lieu : en
// entier, puis par morceaux (zoom CAPZOOM, 0,55 par défaut) de gauche à droite.
// CAPPLACES=boulangerie,glacier : seulement ces lieux.

type saveShots struct {
	dir   string
	queue []func(g *game) string
}

func newSaveShots(dir, file string) *saveShots {
	s := &saveShots{dir: dir}
	s.queue = append(s.queue, func(g *game) string {
		b, err := os.ReadFile(file)
		if err == nil {
			var f struct {
				Jeu *sim.Save `json:"jeu"`
			}
			if json.Unmarshal(b, &f) == nil && f.Jeu != nil {
				g.sim.Restore(f.Jeu, f.Jeu.Unix)
			}
		}
		for _, p := range g.sim.Places {
			g.sim.Enter(p)
			for t := 0.0; t < envF("CAPSEC", 20); t += 1.0 / 30 {
				g.sim.Step(1.0 / 30)
			}
			g.sim.Events = g.sim.Events[:0]
		}
		if os.Getenv("CAPMAP") != "" { // CAPMAP=1 : la carte de la ville d'abord
			s.queue = append(s.queue, func(g *game) string {
				g.goMap()
				return "carte"
			})
		}
		for _, p := range g.sim.Places {
			p := p
			if !p.Owned || (os.Getenv("CAPPLACES") != "" && !strings.Contains(","+os.Getenv("CAPPLACES")+",", ","+p.ID+",")) {
				continue
			}
			if os.Getenv("CAPPOINTS") != "" { // CAPPOINTS=1 : la page des points de l'école, pour ce lieu
				s.queue = append(s.queue, func(g *game) string {
					g.goPlace(p)
					g.openPanel(panelPoints)
					g.ptsTab, g.ptsPlace = 0, p.ID
					return p.ID + "_points"
				}, func(g *game) string { g.closePanel(); return "" })
			}
			if os.Getenv("CAPHELP") != "" { // CAPHELP=1 : le « ? » de chaque lieu
				s.queue = append(s.queue, func(g *game) string {
					g.goPlace(p)
					g.openPanel(panelPlaceHelp)
					return p.ID + "_aide"
				}, func(g *game) string { g.closePanel(); return "" })
			}
			if at := os.Getenv("CAPAT"); at != "" { // CAPAT=x,y : une seule photo, centrée sur ce point
				var x, y float64
				fmt.Sscanf(at, "%f,%f", &x, &y)
				for k := 0; k < int(envF("CAPN", 1)); k++ {
					k := k
					s.queue = append(s.queue, func(g *game) string {
						g.goPlace(p)
						if k > 0 {
							for t := 0.0; t < envF("CAPDT", 3); t += 1.0 / 30 {
								g.sim.Step(1.0 / 30)
							}
						}
						g.zoom = envF("CAPZOOM", 0.55)
						g.sim.Player.Pos = sim.V{X: x, Y: y}
						for i := 0; i < 60; i++ {
							g.updateCamera(p, 1)
						}
						return p.ID + "_vise_" + itoa2(k)
					})
				}
				continue
			}
			s.queue = append(s.queue, func(g *game) string {
				g.goPlace(p)
				w, h := p.MaxX()-p.MinX(), p.H-p.Top+300
				g.zoom = min(scr.VW/w, scr.VH/h)
				g.sim.Player.Pos = sim.V{X: (p.MinX() + p.MaxX()) / 2, Y: (p.Top + p.H) / 2}
				for i := 0; i < 60; i++ {
					g.updateCamera(p, 1)
				}
				return p.ID + "_0"
			})
			z := envF("CAPZOOM", 0.55)
			step := scr.VW / z * 0.8
			n := 1
			for x := p.MinX() + scr.VW/z/2; x-scr.VW/z/2 < p.MaxX(); x += step {
				x, k := x, n
				for _, y := range []float64{p.Top + scr.VH/z/2, p.H - scr.VH/z/2 + 200} {
					y := y
					s.queue = append(s.queue, func(g *game) string {
						g.goPlace(p)
						g.zoom = z
						g.sim.Player.Pos = sim.V{X: x, Y: y}
						for i := 0; i < 60; i++ {
							g.updateCamera(p, 1)
						}
						if y > p.Top+scr.VH/z/2 {
							return p.ID + "_" + itoa2(k) + "b"
						}
						return p.ID + "_" + itoa2(k) + "a"
					})
				}
				n++
			}
		}
		return ""
	})
	return s
}

func itoa2(n int) string { return string(rune('0'+n/10)) + string(rune('0'+n%10)) }

func (s *saveShots) update(g *game) error {
	if g.capName != "" {
		return nil
	}
	if len(s.queue) == 0 {
		return errDone
	}
	f := s.queue[0]
	s.queue = s.queue[1:]
	if n := f(g); n != "" {
		g.capName = s.dir + "/" + n + ".png"
	}
	return nil
}
