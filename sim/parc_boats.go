package sim

import "math"

// Petits bateaux et grand splash : chaque bateau (rondin) est autonome. Au quai,
// il se remplit, part dès qu'il est plein (ou après une longue attente), puis le
// suivant s'avance, avec un petit temps entre deux départs. Au retour, il
// débarque ses passagers à la sortie et se range derrière le quai.

// RideBoat : un bateau (ou un rondin) d'une attraction sur l'eau.
type RideBoat struct {
	Riders []*Visitor
	State  int     // 0 attend derrière le quai, 1 au quai (embarquement), 2 en route, 3 débarquement
	T      float64 // temps depuis le départ
	Prog   float64 // avancement du tour (0 à 1)
	Line   int     // rang dans la file derrière le quai (0 : au quai)
	Col    int     // sa couleur (elle ne change jamais)
	// Back : où il est, en places derrière le quai (il avance doucement, jamais à reculons) ;
	// EndBack : où il finira son tour (au bout de la file qui attend).
	Back, EndBack float64
}

const (
	boatGap  = 3.0 // secondes entre deux départs
	boatIdle = 8.0 // un bateau pas plein part quand même après cette attente
)

// Indep : les wagons de cette attraction partent chacun de leur côté.
func (r *Ride) Indep() bool { return r.Kind.Water }

func (p *Place) stepBoats(r *Ride, dt float64) {
	manned := p.RideManned(r)
	board := 1.0 / (1 + p.ptsF("pts_embarq"))
	// autant de bateaux que d'achetés
	for len(r.Boats) < p.RideUnits(r) {
		n := len(r.Boats)
		r.Boats = append(r.Boats, &RideBoat{Col: n, Back: float64(n), EndBack: 1})
	}
	r.gapT += dt
	r.boardT += dt
	// la file derrière le quai, dans l'ordre d'arrivée
	var line []*RideBoat
	for _, b := range r.Boats {
		if b.State == 0 || b.State == 1 {
			line = append(line, b)
		}
	}
	for i, b := range line {
		b.Line = i
	}
	// ils glissent jusqu'à leur place (vers l'avant seulement)
	const glide = 1.3 // places par seconde
	for _, b := range r.Boats {
		target := float64(b.Line)
		switch b.State {
		case 2:
			b.EndBack += math.Max(-glide*dt, math.Min(glide*dt, float64(len(line))-b.EndBack))
			continue
		case 3:
			target = float64(len(line))
		}
		if b.Back > target {
			b.Back = math.Max(target, b.Back-glide*dt)
		} else {
			b.Back = target
		}
	}
	if len(line) > 0 && manned && r.gapT >= boatGap && line[0].Back < 0.05 {
		b := line[0]
		b.State = 1
		seats := r.Kind.Seats
		if len(r.Queue) > 0 && len(b.Riders) < seats {
			v := r.Queue[0]
			if v.Pos.Dist(r.QueuePos(0)) < 30 && r.boardT >= board {
				r.boardT = 0
				r.Queue = r.Queue[1:]
				b.Riders = append(b.Riders, v)
				v.State = vRiding
				v.ride = r
				v.Pos = r.Booth().C()
				r.idleT = 0
			}
		} else if len(b.Riders) > 0 {
			r.idleT += dt
		}
		if len(b.Riders) >= seats || (len(b.Riders) > 0 && r.idleT > boatIdle) {
			b.State, b.T, b.Prog = 2, 0, 0
			b.EndBack = float64(len(line) - 1)
			r.gapT, r.idleT = 0, 0
		}
	}
	for _, b := range r.Boats {
		switch b.State {
		case 2:
			b.T += dt
			b.Prog = math.Min(1, b.T/p.RideTime(r))
			if b.Prog >= 1 {
				b.State = 3
				b.Back = b.EndBack // (il s'arrête là où son tour l'a mené)
			}
		case 3:
			if !manned {
				for _, v := range b.Riders {
					v.Joy = math.Max(0, v.Joy-parcJoyWait*dt/60)
				}
				continue
			}
			if len(b.Riders) == 0 {
				b.State, b.Prog = 0, 0
				// (il se range au bout de la file : il passe après ceux qui attendent)
				for i, o := range r.Boats {
					if o == b {
						r.Boats = append(append(r.Boats[:i:i], r.Boats[i+1:]...), b)
						break
					}
				}
				return
			}
			b.T += dt
			if b.T < p.RideTime(r)+board*0.6 {
				continue
			}
			b.T = p.RideTime(r)
			v := b.Riders[0]
			b.Riders = b.Riders[1:]
			p.parcRideDone(r, v)
		}
	}
}
