package main

import (
	"errors"
	"image/png"
	"os"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"

	"ville/sim"
)

// Mode « captures » (développement) : un joueur automatique joue en accéléré,
// puis on enregistre des images des différents écrans.

var errDone = errors.New("fini")

type shotScript struct {
	dir   string
	step  int
	bot   *sim.Bot
	frame int
}

func newShotScript(dir string) *shotScript {
	return &shotScript{dir: dir, bot: &sim.Bot{}}
}

func saveScreenshot(img *ebiten.Image, name string) {
	f, err := os.Create(name)
	if err != nil {
		return
	}
	defer f.Close()
	_ = png.Encode(f, img)
}

func envF(k string, d float64) float64 {
	if v, err := strconv.ParseFloat(os.Getenv(k), 64); err == nil {
		return v
	}
	return d
}

func (s *shotScript) update(g *game) error {
	if g.capName != "" {
		return nil
	}
	s.frame++
	cap := func(n string) { g.capName = s.dir + "/" + n + ".png" }
	switch s.step {
	case 0:
		// début de partie
		g.openPanel(panelHelp)
		cap("01_aide")
	case 1:
		g.closePanel()
		for i := 0; i < 90; i++ {
			g.sim.Step(1.0 / 60)
			g.handleEvents()
			g.updateParticles(1.0 / 60)
		}
		g.camPos = g.sim.Player.Pos
		cap("02_debut")
	case 2:
		// on joue en accéléré
		mins := envF("MINUTES", 25)
		if v := envF("SECONDES", 0); v > 0 {
			mins = v / 60
		}
		for g.sim.T < mins*60 {
			s.bot.Step(g.sim, 1.0/30)
			g.sim.Step(1.0 / 30)
			g.sim.Events = g.sim.Events[:0]
		}
		g.parts = nil
		for i := 0; i < 60; i++ {
			s.bot.Step(g.sim, 1.0/60)
			g.sim.Step(1.0 / 60)
			g.t += 1.0 / 60
			g.handleEvents()
			g.updateParticles(1.0 / 60)
			g.updateCamera(g.sim.Cur, 1)
		}
		cap("03_ferme")
	case 3:
		g.zoom = 0.6
		g.updateCamera(g.sim.Cur, 1)
		cap("04_ferme_large")
	case 4:
		g.zoom = 1.15
		g.openPanel(panelUpgrades)
		cap("05_ameliorations")
	case 5:
		g.closePanel()
		g.goMap()
		cap("06_carte")
	case 6:
		g.openPanel(panelMairie)
		cap("07_mairie")
	case 7:
		g.openPanel(panelWardrobe)
		cap("08_vetements")
	case 8:
		g.openPanel(panelReset)
		cap("10_remise_a_zero")
	case 9:
		g.welcome.amount, g.welcome.secs = 1234, 3*3600+720
		g.openPanel(panelWelcome)
		cap("09_retour")
	case 10:
		g.openPanel(panelBadges)
		cap("12_badges")
	case 11:
		g.closePanel()
		g.sim.Money += g.sim.Place("super").Price
		g.buying = g.sim.Place("super")
		g.openPanel(panelBuy)
		cap("12b_achat")
	case 12:
		// le supermarché
		g.closePanel()
		sup := g.sim.Place("super")
		g.sim.BuyPlace(sup)
		g.goPlace(g.sim.Places[0])
		end := g.sim.T + envF("SUPERMIN", 15)*60
		for g.sim.T < end {
			s.bot.Step(g.sim, 1.0/30)
			g.sim.Step(1.0 / 30)
			g.sim.Events = g.sim.Events[:0]
		}
		g.updateCamera(g.sim.Cur, 1)
		g.camPos = g.sim.Place("ferme").Station("quai").Body.C()
		cap("13_ferme_quai")
	case 13:
		g.goPlace(g.sim.Place("super"))
		for i := 0; i < 120; i++ {
			s.bot.Step(g.sim, 1.0/60)
			g.sim.Step(1.0 / 60)
			g.t += 1.0 / 60
			g.handleEvents()
			g.updateParticles(1.0 / 60)
			g.updateCamera(g.sim.Cur, 1)
		}
		cap("14_super")
	case 14:
		g.zoom = 0.6
		g.updateCamera(g.sim.Cur, 1)
		cap("15_super_large")
	case 15:
		g.zoom = 1.15
		g.goMap()
		cap("16_carte")
	case 16:
		for _, id := range []string{"ecole", "boulangerie", "glacier"} {
			pl := g.sim.Place(id)
			g.sim.Money += pl.Price
			g.sim.BuyPlace(pl)
		}
		end := g.sim.T + envF("SUPERMIN", 15)*60
		g.goPlace(g.sim.Places[0])
		for g.sim.T < end {
			s.bot.Step(g.sim, 1.0/30)
			g.sim.Step(1.0 / 30)
			g.sim.Events = g.sim.Events[:0]
		}
		g.zoom = 0.45
		g.updateCamera(g.sim.Cur, 1)
		cap("17_ferme_entiere")
	case 17:
		g.zoom = 0.6
		g.goPlace(g.sim.Place("boulangerie"))
		g.updateCamera(g.sim.Cur, 1)
		cap("18_boulangerie")
	case 18:
		g.goPlace(g.sim.Place("glacier"))
		g.updateCamera(g.sim.Cur, 1)
		cap("19_glacier")
	case 19:
		g.goPlace(g.sim.Place("ecole"))
		g.zoom = 0.55
		g.sim.Player.Pos = sim.V{X: sim.SchoolAccueil().C().X, Y: 1100}
		for i := 0; i < 8; i++ {
			g.updateCamera(g.sim.Cur, 1)
		}
		cap("19b_ecole")
	case 20:
		g.zoom = 1.15
		g.sim.Player.Pos = g.sim.Place("ecole").Rooms[0].Body.C()
		for i := 0; i < 8; i++ {
			g.updateCamera(g.sim.Cur, 1)
		}
		cap("19c_classe")
	case 21:
		g.sim.Levels["pts.ouvrier"] += 23
		g.sim.Levels["pts.marketing"] += 7
		g.sim.Levels["pts.comptable"] += 2
		g.goMap()
		g.openPanel(panelPoints)
		g.pending = map[string]int{"ferme.pts_agro": 10, "ferme.pts_cam_vit": 3, "super.pts_marketing": 1}
		cap("19d_points")
	case 22:
		g.closePanel()
		g.zoom = 1.15
		g.goMap()
		cap("20_carte")
	case 23:
		g.resetGame()
		cap("11_apres_remise")
	default:
		return errDone
	}
	s.step++
	return nil
}
