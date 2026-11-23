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
		if os.Getenv("CAPTUTO") != "" { // CAPTUTO=1 : les écrans du tutoriel
			ferme, ecole := g.sim.Place("ferme"), g.sim.Place("ecole")
			s.queue = append(s.queue,
				func(g *game) string { g.newStart(); g.openPanel(panelStart); return "tuto_0_bienvenue" },
				func(g *game) string { g.panel = panelNone; g.setTuto(tutoMap); return "tuto_1_carte" },
				func(g *game) string { g.goPlace(ferme); g.zoom = 1.15; g.setTuto(tutoHarvest); return "tuto_2_recolte" },
				func(g *game) string { g.setTuto(tutoServe); return "tuto_4_client" },
				func(g *game) string { g.setTuto(tutoSchool); return "tuto_10_ecole" },
				func(g *game) string { g.goPlace(ecole); g.setTuto(tutoMairie); return "tuto_12_maison" },
				func(g *game) string {
					g.setTuto(tutoPoints)
					g.openPanel(panelMairie)
					g.ptsTab, g.ptsPlace = 0, "ferme"
					return "tuto_13_points"
				},
				func(g *game) string {
					g.closePanel()
					g.goPlace(ferme)
					g.setTuto(tutoVest)
					return "tuto_15_vestiaire"
				},
				func(g *game) string { g.setTuto(tutoWard); g.openPanel(panelWardrobe); return "tuto_16_perso" },
				func(g *game) string { g.closePanel(); g.setTuto(tutoBravo); return "tuto_17_bravo" },
			)
			return ""
		}
		if os.Getenv("CAPPATROL") != "" {
			patrolAt = envF("CAPPATROL", 0.5)
		}
		if v := os.Getenv("CAPPOL"); v != "" { // CAPPOL=1 : un commissariat équipé (2 : et son menu, 3 : le tutoriel)
			lv := g.sim.Levels
			lv["commissariat.commissaire"], lv["commissariat.lieutenant_1"], lv["commissariat.lieutenant_2"], lv["commissariat.voiture_2"] = 1, 1, 1, 1
			lv["commissariat.bureau_1"], lv["commissariat.capitaine_1"], lv["commissariat.bureau_2"], lv["commissariat.cellule_1"], lv["commissariat.cellule_2"], lv["commissariat.video_1"] = 1, 1, 1, 1, 1, 90
			g.sim.SetCamera(1, "ferme")
			g.sim.Custody = append(g.sim.Custody, &sim.Custody{N: 7, Left: 3000})
			if v != "3" {
				lv["commissariat.tuto"] = 1
			}
			g.sim.SetPatrol(1, 0, "super")
			if v == "2" {
				s.queue = append(s.queue, func(g *game) string {
					g.goPlace(g.sim.Place("commissariat"))
					g.openPanel(panelPolice)
					g.polTab = 1
					return "video"
				}, func(g *game) string { g.closePanel(); return "" })
			}
		}
		if v := os.Getenv("CAPCAS"); v != "" { // CAPCAS=1 : une caserne équipée (2 : et son menu)
			lv := g.sim.Levels
			for _, k := range []string{"commandant", "bureau_1", "bureau_2", "charge_1", "camion_1", "camion_2", "pompier_1", "pompier_2"} {
				lv["caserne."+k] = 1
			}
			lv["pts.ppi"] = 35
			if v == "2" {
				s.queue = append(s.queue, func(g *game) string {
					g.goPlace(g.sim.Place("caserne"))
					g.openPanel(panelCaserne)
					return "caserne_menu"
				}, func(g *game) string { g.casTab = 1; return "caserne_camions" }, func(g *game) string { g.closePanel(); return "" })
			}
		}
		if os.Getenv("CAPPRIS") != "" { // CAPPRIS=1 : une prison bien équipée, avec des détenus
			lv := g.sim.Levels
			lv["commissariat.arrestations"] = 5
			g.sim.Place("prison").Owned = true
			for _, k := range []string{"directeur", "surveillant_1", "surveillant_2", "surveillant_3", "cellule_1", "cellule_2", "cellule_3", "cellule_4", "cellule_5",
				"fourgon_1", "fourgon_2", "cantine", "table_1", "table_2", "bains", "cabine_1", "cabine_2", "cabine_3", "sport", "agres_1", "agres_2", "agres_6",
				"famille_1", "famille_2", "travail", "poste_1", "poste_2", "psy", "as"} {
				lv["prison."+k] = 1
			}
			lv["prison.niv_cellule_1"], lv["prison.niv_cellule_2"], lv["prison.niv_cantine"] = 10, 4, 3
			g.sim.Custody = append(g.sim.Custody, &sim.Custody{N: 5, Left: 1500})
			for i := 0; i < 30*60*12; i++ { // (12 minutes : les détenus vivent leur vie)
				g.sim.Step(1.0 / 30)
			}
			if os.Getenv("CAPPRIS") == "2" {
				s.queue = append(s.queue, func(g *game) string {
					g.goPlace(g.sim.Place("prison"))
					g.openPanel(panelPrison)
					return "prison_menu"
				}, func(g *game) string { g.closePanel(); return "" })
			}
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
			if os.Getenv("CAPQUIT") != "" { // CAPQUIT=1 : la fenêtre « Quitter le jeu ? »
				s.queue = append(s.queue, func(g *game) string {
					g.goPlace(p)
					g.openPanel(panelQuit)
					return p.ID + "_quitter"
				}, func(g *game) string { g.closePanel(); return "" })
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
						if k == 0 && os.Getenv("CAPVOL") != "" { // CAPVOL=1 : des voleurs attaquent le lieu
							g.sim.StartTheft(p.ID)
						}
						if k == 0 && os.Getenv("CAPFEU") != "" { // CAPFEU=1 : le lieu prend feu
							g.sim.StartFire(p.ID)
						}
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
