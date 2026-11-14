// Joue la ferme en accéléré avec un joueur automatique et affiche le rythme.
//
//	go run ./cmd/balance [heures]              joueur automatique, sans pause
//	go run ./cmd/balance -humain [jours]       joueuse « réelle » : séances de 30 min
//	                                           entre 8 h et 23 h, pauses de 2 h, nuits ;
//	                                           gains hors ligne par la vraie sauvegarde
//	options du mode humain : -seance MIN, -pause MIN
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"ville/sim"
)

func main() {
	human := flag.Bool("humain", false, "séances courtes et absences")
	session := flag.Float64("seance", 30, "durée d'une séance (minutes)")
	pause := flag.Float64("pause", 120, "pause entre deux séances dans la journée (minutes)")
	flag.Parse()
	if *human {
		days := 7.0
		if flag.NArg() > 0 {
			days, _ = strconv.ParseFloat(flag.Arg(0), 64)
		}
		runHuman(days, *session*60, *pause*60)
		return
	}
	hours := 3.0
	if flag.NArg() > 0 {
		hours, _ = strconv.ParseFloat(flag.Arg(0), 64)
	}
	g := sim.NewGame()
	b := &sim.Bot{}
	const dt = 1.0 / 30
	last := 0
	for g.T < hours*3600 {
		b.Step(g, dt)
		g.Step(dt)
		g.Events = g.Events[:0]
		if int(g.T)/600 != last {
			last = int(g.T) / 600
			fmt.Printf("  --- %3d min : %8.0f € en poche, %6.1f €/s, servis %d, ratés %d, réputation %.2f / %.2f, badges %d\n", last*10, g.Money, g.Rate(), g.Stats.Served, g.Stats.Missed, g.Places[0].Rep, g.Place("super").Rep, len(g.Badges))
			if last%3 == 0 {
				fmt.Printf("        fabriqués : %v\n", g.Stats.Made)
			}
		}
		for _, l := range b.Log {
			fmt.Println(l)
		}
		b.Log = b.Log[:0]
	}
}

var jours = []string{"dim", "lun", "mar", "mer", "jeu", "ven", "sam"}

// clock : « dim 15:00 » pour un instant (en secondes) depuis dimanche 0 h.
func clock(t float64) string {
	m := int(t / 60)
	return fmt.Sprintf("%s %02d:%02d", jours[m/1440%7], m/60%24, m%60)
}

// runHuman : la partie commence un dimanche à 15 h ; la joueuse joue des séances
// courtes entre 8 h et 23 h ; entre deux séances, le jeu est fermé (sauvegarde
// écrite, relue à la séance suivante : gains hors ligne compris).
func runHuman(days, session, pause float64) {
	g := sim.NewGame()
	b := &sim.Bot{Human: true}
	const dt = 1.0 / 30
	now := 15 * 3600.0 // dimanche 15 h (horloge réelle, en secondes)
	end := now + days*86400
	owned := map[string]bool{}
	nSess := 0
	offTotal := 0.0
	for now < end {
		nSess++
		start := g.T
		for g.T-start < session {
			b.Step(g, dt)
			g.Step(dt)
			g.Events = g.Events[:0]
			for _, l := range b.Log {
				_ = l // (le détail des achats est trop long pour des jours de jeu)
			}
			b.Log = b.Log[:0]
			for _, p := range g.Places {
				if p.Owned && !owned[p.ID] {
					owned[p.ID] = true
					fmt.Printf("  >> %s (séance %d, %s de jeu) : %s\n", clock(now+g.T-start), nSess, hm(g.Stats.Played), p.Name)
				}
			}
		}
		now += session
		// le jeu se ferme : prochaine séance après la pause, ou le lendemain à 8 h
		next := now + pause
		if h := hourOf(next); h >= 23 || h < 8 {
			d := int(next / 86400)
			if hourOf(next) >= 23 {
				d++
			}
			next = float64(d)*86400 + 8*3600
		}
		sv := g.Snapshot(int64(now))
		raw, _ := json.Marshal(sv)
		var back sim.Save
		if err := json.Unmarshal(raw, &back); err != nil {
			panic(err)
		}
		g2 := sim.NewGame()
		off, secs := g2.Restore(&back, int64(next))
		g = g2
		b = &sim.Bot{Human: true} // (le bot repart de zéro sur la partie rechargée)
		offTotal += off
		fmt.Printf("  %s  séance %2d finie : %s de jeu, %9.0f € en poche, %7.1f €/s · absence %s → +%.0f € (%.0f min comptées)\n",
			clock(now), nSess, hm(g.Stats.Played), sv.Money, sv.Rate, hm(next-now), off, secs/60)
		now = next
	}
	if out := os.Getenv("BALANCE_SAVE"); out != "" { // (pour comparer avec une vraie sauvegarde)
		raw, _ := json.MarshalIndent(map[string]any{"jeu": g.Snapshot(int64(now))}, "", " ")
		os.WriteFile(out, raw, 0o644)
	}
	fmt.Printf("\n  Bilan : %d séances, %s de jeu, %.0f € gagnés dont %.0f € hors ligne (%.0f %%)\n",
		nSess, hm(g.Stats.Played), g.Stats.Earned, offTotal, 100*offTotal/max(1, g.Stats.Earned))
}

func hourOf(t float64) int { return int(t/3600) % 24 }

func hm(s float64) string {
	m := int(s / 60)
	return fmt.Sprintf("%dh%02d", m/60, m%60)
}
