package sim

import "math"

// ---------------------------------------------------------------- employés

// need : combien une destination peut encore accepter d'un produit.
func (p *Place) need(s *Station, it Item) int {
	switch s.Kind {
	case Shelf, Store:
		if s.Item == it {
			return s.StockMax - s.Stock - s.claim[it]
		}
	case Field:
		if s.Replant && it == s.Seed {
			n := 0
			for _, pl := range s.Plants {
				if pl.Empty() {
					n++
				}
			}
			return n - s.claim[it]
		}
	case Machine:
		for _, in := range s.Inputs {
			if in == it {
				return s.BufMax - s.Buf[it] - s.claim[it]
			}
		}
	case Dock:
		if p.g.dockAccepts(s, it) {
			return s.BufMax - s.Buf[it] - s.claim[it]
		}
	}
	return 0
}

// SoonRipe : un ouvrier attend les pousses mûres dans moins de ce nombre de secondes
// (il ramasse toute la rangée d'un coup au lieu de faire un aller-retour par légume).
const SoonRipe = 5.0

// soon : la pousse sera-t-elle mûre bientôt ?
func (p *Place) soon(s *Station, pl Plant) bool {
	if pl.G < 0 || pl.G >= 1 {
		return false
	}
	f, _ := p.speedOf(s)
	return (1-pl.G)*s.GrowTime()/math.Max(0.1, f) <= SoonRipe
}

// soonCount : pousses bientôt mûres (pas encore cueillables).
func (p *Place) soonCount(s *Station) int {
	n := 0
	for _, pl := range s.Plants {
		if p.soon(s, pl) {
			n++
		}
	}
	return n
}

func (p *Place) needMax(s *Station) int {
	if s.Kind == Shelf || s.Kind == Store {
		return s.StockMax
	}
	if s.Kind == Field {
		return max(1, len(s.Plants))
	}
	return s.BufMax
}

// available : ce qu'une source peut donner tout de suite.
func (p *Place) available(s *Station) int {
	switch s.Kind {
	case Field:
		n := 0
		for _, pl := range s.Plants {
			if pl.G >= 1 {
				n++
			}
		}
		return n*FieldYield(p.Level(s)) - s.claim[s.Item]
	case Animal, Machine, Pallet:
		return s.Pile - s.claim[s.Item]
	}
	return 0
}

func (p *Place) pickJob(w *Agent) *job {
	capa := p.carryCap(w)
	var best *job
	bestScore := 0.0
	for _, dst := range p.Stations {
		if !dst.Built || (dst.Kind != Shelf && dst.Kind != Machine && dst.Kind != Dock && dst.Kind != Store && !(dst.Kind == Field && dst.Replant)) {
			continue
		}
		var items []Item
		switch dst.Kind {
		case Field:
			items = []Item{dst.Seed} // replanter
		case Shelf, Store:
			items = []Item{dst.Item}
		case Machine:
			items = dst.Inputs
		default:
			for it := Item(0); it < NItems; it++ {
				if p.g.dockAccepts(dst, it) {
					items = append(items, it)
				}
			}
		}
		for _, it := range items {
			if !teamAllows(w, dst, it) {
				continue
			}
			n := p.need(dst, it)
			if n <= 0 {
				continue
			}
			for _, src := range p.Stations {
				if !src.Built || src.Dredged || (src.Item != it && !(src.Two && src.Item2 == it)) || (src.Kind != Field && src.Kind != Animal && src.Kind != Machine && src.Kind != Pallet) {
					continue
				}
				if src.Team != "" && w.Team != "" && src.Team != w.Team && !src.Shared {
					continue // (on ne se sert pas dans l'espace des autres)
				}
				av := p.available(src)
				if src.Two && it == src.Item2 {
					av = src.Pile2 - src.claim[it]
				}
				if src.Kind == Field && av > 0 && av < capa-w.load() {
					av += p.soonCount(src) * FieldYield(p.Level(src)) // il attendra celles qui vont mûrir
				}
				if src.Kind == Field && src.Column && av < min(3, len(src.Plants)) {
					continue // (pas tout en haut de la colonne pour un seul légume : il ira ailleurs ou attendra)
				}
				if av <= 0 {
					continue
				}
				want := min(capa-w.load(), n, av) // (les places qui restent dans ses bras)
				// on privilégie ce qui manque le plus et ce qui rapporte le plus
				score := float64(n) / float64(p.needMax(dst)) * (1 + math.Sqrt(Items[it].Price)/4)
				if dst.Kind == Shelf && dst.Stock == 0 {
					score *= 1.6
				}
				if dst.Kind == Machine && dst.Buf[it] == 0 {
					score *= 1.5 // une machine à l'arrêt passe avant
				}
				if dst.Kind == Dock {
					score *= 0.6 // la boutique de la ferme d'abord
				}
				if dst.Kind == Store {
					score *= 0.12 // la réserve : seulement le surplus
				}
				if dst.Kind == Field {
					score = 2.2 * float64(n) / float64(p.needMax(dst)) // replanter passe avant presque tout
				}
				if p.Crep != nil && ((dst.Kind == Machine && dst.Item != CrepeNature && p.crepDemand(dst.Item) > 0) || dst.Kind == Shelf) {
					score *= 6 // crêperie : ce qu'une table attend passe avant les réserves
				}
				score *= 0.85 + 0.3*p.rand()
				score /= 1 + w.Pos.Dist(src.Body.C())/2500
				if score > bestScore {
					bestScore = score
					best = &job{src: src, dst: dst, it: it, want: want}
				}
			}
		}
	}
	if best != nil {
		best.dst.claim[best.it] += best.want
		best.src.claim[best.it] += best.want
	}
	return best
}

// chainJob : une autre colonne de son équipe où cueillir avant d'aller aux étals
// (au moins 2 fruits ou légumes, mûrs ou presque ; la plus proche).
func (p *Place) chainJob(w *Agent) *job {
	room := p.carryCap(w) - w.load()
	var best *job
	bd := math.Inf(1)
	for _, src := range p.Stations {
		if !src.Built || src.Kind != Field || !src.Column || (src.Team != "" && src.Team != w.Team) {
			continue
		}
		if p.ID == "ferme" && teamOf(src.Item) != w.Team {
			continue
		}
		av := p.available(src) + p.soonCount(src)*FieldYield(p.Level(src))
		if av < min(2, room) {
			continue
		}
		var dst *Station
		for _, s := range p.Stations {
			if s.Built && s.Kind == Shelf && s.Item == src.Item && teamAllows(w, s, src.Item) && p.need(s, src.Item) > 0 {
				dst = s
				break
			}
		}
		if dst == nil {
			continue
		}
		want := min(room, av, p.need(dst, src.Item))
		if d := w.Pos.Dist(src.Body.C()); want > 0 && d < bd {
			bd = d
			best = &job{src: src, dst: dst, it: src.Item, want: want}
		}
	}
	if best != nil {
		best.dst.claim[best.it] += best.want
		best.src.claim[best.it] += best.want
	}
	return best
}

func (p *Place) endJob(w *Agent) {
	j := w.job
	if j == nil {
		return
	}
	j.dst.claim[j.it] -= j.want
	if j.dst.claim[j.it] < 0 {
		j.dst.claim[j.it] = 0
	}
	if j.phase == 0 {
		p.releaseSrc(j)
	}
	w.job = nil
}

func (p *Place) releaseSrc(j *job) {
	if j.src != nil {
		j.src.claim[j.it] -= j.want
		if j.src.claim[j.it] < 0 {
			j.src.claim[j.it] = 0
		}
		j.src = nil
	}
}

// findDest cherche où livrer ce que l'employé porte déjà.
func (p *Place) findDest(w *Agent) (*Station, Item) {
	for _, it := range w.Carry {
		for _, s := range p.Stations {
			if s.Built && teamAllows(w, s, it) && p.need(s, it) > 0 {
				return s, it
			}
		}
	}
	return nil, 0
}

// BoredAfter : secondes sans travail avant que l'ouvrier s'installe.
const BoredAfter = 8.0

// CashierBoredAfter : secondes sans client avant que la caissière sifflote.
const CashierBoredAfter = 5.0

func (p *Place) workerAI(w *Agent, dt float64) {
	speed := p.agentSpeed(w)
	if w.job == nil {
		// des sachets de graines en trop (plus de case vide) : il les remet dans sa
		// poche (ils sont gratuits), au lieu de se promener avec
		keep := w.Carry[:0]
		for _, it := range w.Carry {
			if IsSeed(it) {
				useful := false
				for _, s := range p.Stations {
					if s.Kind == Field && s.Replant && s.Seed == it && p.need(s, it) > 0 {
						useful = true
					}
				}
				if !useful {
					continue
				}
			}
			keep = append(keep, it)
		}
		w.Carry = keep
		if len(w.Carry) > 0 {
			dst, it := p.findDest(w)
			if dst != nil {
				n := min(w.count(it), p.need(dst, it))
				w.job = &job{dst: dst, it: it, want: n, phase: 1}
				dst.claim[it] += n
				w.noDestT = 0
			} else {
				w.noDestT += dt
				// il garde ce qu'il porte (ça servira plus tard) et utilise les
				// places qui lui restent pour un autre travail
				if w.load() < p.carryCap(w) {
					w.idleT -= dt
					if w.idleT <= 0 {
						w.idleT = 0.4
						if j := p.pickJob(w); j != nil {
							w.job = j
							w.noDestT = 0
						}
					}
				}
				if w.job == nil && w.load() >= p.carryCap(w) && w.noDestT > 30 {
					// les bras pleins de choses inutiles depuis longtemps : direction le compost
					for _, s := range p.Stations {
						if s.Kind == Trash && s.Zone.W > 0 { // (le hangar et le local sont des décors)
							if p.walk(w, s.Zone.C(), speed, dt) {
								w.noDestT = 30
							}
							return
						}
					}
				}
			}
		} else {
			w.idleT -= dt
			if w.idleT <= 0 {
				w.job = p.pickJob(w)
				w.idleT = 0.4
			}
		}
		if w.job == nil {
			if len(w.Carry) == 0 {
				w.Bored += dt
			}
			if w.Settled {
				w.Moving = false
				return
			}
			// flâne près du point de repos
			w.wanderT -= dt
			if w.wanderT <= 0 || !w.hasT {
				rest, dy := p.RestSpot, 120.0
				if w.Team != "" {
					rest, dy = p.teamRest(w.Team), 50
				}
				w.target = rest.Add(V{(p.rand() - 0.5) * 260, (p.rand() - 0.5) * dy})
				w.hasT = true
				w.wanderT = 3 + 3*p.rand()
			}
			if p.walk(w, w.target, speed*0.6, dt) && w.Bored > BoredAfter {
				// plus rien à faire depuis un moment : il s'installe
				w.Settled = true
				w.Mood = int(p.rand() * 3)
				w.Face = V{0, 1}
			}
			return
		}
	}
	w.Bored, w.Settled = 0, false
	j := w.job
	capa := p.carryCap(w)
	if j.phase == 0 {
		src := j.src
		// serre : il prend d'abord les sachets de graines (au pied de la colonne),
		// pour replanter chaque case dès qu'il l'a cueillie
		if src.Kind == Field && src.Replant && !j.seeded {
			crate := p.Station("graines_" + src.ID)
			y := max(1, FieldYield(p.Level(src)))
			need := (j.want+y-1)/y - w.count(src.Seed)
			if need <= 0 || crate == nil || !crate.Built {
				j.seeded = true
			} else {
				if p.walk(w, crate.Zone.C(), speed, dt) || crate.Zone.Has(w.Pos) {
					for k := 0; k < need; k++ {
						w.Carry = append(w.Carry, src.Seed)
					}
					j.seeded = true
					p.g.emit(Event{Kind: EvTake, Pos: crate.Zone.C(), It: src.Seed})
				}
				return
			}
		}
		var tgt V
		inZone, fieldWait := false, false
		switch src.Kind {
		case Field:
			// vers le plant mûr le plus proche
			best, bd := -1, 1e9
			for i, pl := range src.Plants {
				if pl.G >= 1 {
					if d := pl.P.Dist(w.Pos); d < bd {
						best, bd = i, d
					}
				}
			}
			soon := false
			if best < 0 && w.load() < capa {
				// rien de mûr : il attend devant la pousse qui va mûrir le plus tôt
				bg := -1.0
				for i, pl := range src.Plants {
					if p.soon(src, pl) && pl.G > bg {
						best, bg = i, pl.G
					}
				}
				soon = best >= 0
			}
			if best >= 0 {
				tgt = src.Plants[best].P
			} else {
				tgt = src.Zone.C()
			}
			inZone = src.Zone.Has(w.Pos) && (best < 0 || soon)
			if soon {
				fieldWait = true
			}
		case Machine:
			tgt = src.OutZone.C()
			inZone = src.OutZone.Has(w.Pos)
			if src.Two && j.it == src.Item2 {
				tgt = src.OutZone2.C()
				inZone = src.OutZone2.Has(w.Pos)
			}
		default:
			tgt = src.Zone.C()
			inZone = src.Zone.Has(w.Pos)
		}
		arrived := p.walk(w, tgt, speed, dt)
		if arrived || inZone {
			j.waitT += dt
		}
		full := w.load() >= capa || j.got >= j.want
		pile := src.Pile
		if src.Two && j.it == src.Item2 {
			pile = src.Pile2
		}
		empty := (src.Kind == Field && inZone && !fieldWait) || ((src.Kind == Animal || src.Kind == Machine || src.Kind == Pallet) && pile == 0 && inZone)
		limit := 6.0
		if fieldWait {
			limit = SoonRipe + 4 // (il attend la fin de la rangée)
		}
		if full || (empty && j.waitT > 0.8) || j.waitT > limit {
			// serre : il replante d'abord les cases qu'il vient de cueillir (il a les graines)
			if src.Kind == Field && src.Replant && w.count(src.Seed) > 0 && j.waitT < limit+6 {
				bd, to := 1e9, V{}
				for _, pl := range src.Plants {
					if pl.Empty() {
						if d := pl.P.Dist(w.Pos); d < bd {
							bd, to = d, pl.P
						}
					}
				}
				if bd < 1e9 {
					p.walk(w, to, speed, dt)
					j.waitT += dt
					return
				}
			}
			p.releaseSrc(j)
			if w.count(j.it) == 0 {
				p.endJob(w)
				return
			}
			j.phase = 1
			// serre, verger, fruits exotiques : il reste de la place dans ses bras ?
			// il remplit d'abord sa cagette dans une autre colonne, puis il livre tout
			if src.Kind == Field && src.Column && w.load() < p.carryCap(w) {
				if nj := p.chainJob(w); nj != nil {
					p.endJob(w)
					w.job = nj
				}
			}
		}
		return
	}
	// livraison
	dst := j.dst
	to := dst.Zone.C()
	if dst.Kind == Field {
		// replanter : vers la case vide la plus proche
		bd := 1e9
		for _, pl := range dst.Plants {
			if pl.Empty() {
				if d := pl.P.Dist(w.Pos); d < bd {
					to, bd = pl.P, d
				}
			}
		}
		if bd > 1e8 {
			p.endJob(w)
			return
		}
		p.walk(w, to, speed, dt)
		j.waitT += dt
		if w.count(j.it) == 0 || j.waitT > 30 {
			p.endJob(w)
		}
		return
	}
	if p.walk(w, to, speed, dt) || dst.Zone.Has(w.Pos) {
		if w.count(j.it) == 0 || p.need(dst, j.it)+dst.claim[j.it] <= 0 {
			p.endJob(w)
		} else {
			j.waitT += dt
			if j.waitT > 20 {
				p.endJob(w)
			}
		}
	}
}

// ---------------------------------------------------------------- clients

func (p *Place) spawnCustomer() {
	var shelves []*Station
	for _, s := range p.Stations {
		if s.Kind == Shelf && s.Built && !s.NoSale && (!p.OnlyStocked || s.Stock-s.booked > 0) {
			shelves = append(shelves, s)
		}
	}
	if len(shelves) == 0 {
		return
	}
	c := &Customer{Look: p.rng.Intn(1 << 20)}
	from := p.SpawnL
	if p.rng.Intn(2) == 0 {
		from = p.SpawnR
	}
	c.Pos = from
	k := 1
	if len(shelves) > 1 && p.rand() < 0.55 {
		k = 2
	}
	if len(shelves) > 3 && p.rand() < 0.3 {
		k = 3
	}
	for k < p.WantMax && k < len(shelves) && p.rand() < 0.4 {
		k++
	}
	k = min(k, 3)                 // petit magasin : 3 produits différents au plus…
	extra := p.ptsF("pts_panier") // (panier moyen, points du collège)
	left := 10 + int(extra*10)    // … et 10 objets en tout
	p.rng.Shuffle(len(shelves), func(i, j int) { shelves[i], shelves[j] = shelves[j], shelves[i] })
	for _, s := range shelves[:k] {
		n := 1 + p.rng.Intn(4)
		if Items[s.Item].Price > 60 {
			n = 1 + p.rng.Intn(2)
		}
		n += int(extra) // panier moyen : un produit de plus tous les 100 points…
		if p.rand() < extra-math.Floor(extra) {
			n++ // … et le reste en chance
		}
		n = min(n, left)
		if n <= 0 {
			break
		}
		left -= n
		b := 0
		if p.OnlyStocked {
			n = min(n, s.Stock-s.booked)
			s.booked += n
			b = n
		}
		c.Wants = append(c.Wants, Want{It: s.Item, N: n, Booked: b})
	}
	// on visite les rayons de gauche à droite
	for i := range c.Wants {
		for j := i + 1; j < len(c.Wants); j++ {
			if p.shelfFor(c.Wants[j].It).Spot.X < p.shelfFor(c.Wants[i].It).Spot.X {
				c.Wants[i], c.Wants[j] = c.Wants[j], c.Wants[i]
			}
		}
	}
	c.path = []V{p.Gate.Add(V{(p.rand() - 0.5) * 60, 0}), p.Inside.Add(V{(p.rand() - 0.5) * 80, 0})}
	p.cartIn(c)
	p.arriveBy(c)
	p.Customers = append(p.Customers, c)
}

func (c *Customer) walkTo(to V, speed, dt float64) bool {
	d := to.Sub(c.Pos)
	if d.Len() < 5 {
		c.Moving = false
		return true
	}
	c.Moving = true
	c.WalkT += dt
	c.Face = d.Norm()
	c.Pos = c.Pos.Add(c.Face.Mul(math.Min(d.Len(), speed*dt)))
	return false
}

func (p *Place) queueSlot(r *Station, i int) V {
	h := r.QueueHead
	if r.QueueLane {
		// trois rangées de six, en serpentin, sous la caisse ; au-delà, vers la droite
		if i < 18 {
			row, col := i/6, i%6
			if row%2 == 1 {
				col = 5 - col
			}
			return h.Add(V{58*float64(col) - 58, 54 * float64(row)})
		}
		return h.Add(V{58*float64(i-18) + 290, 108})
	}
	d := r.QueueDown
	if i < d {
		return h.Add(V{0, 58 * float64(i)})
	}
	return h.Add(V{58 * float64(i-d+1), 58 * float64(d-1)})
}

// pickRegister : la caisse construite où la file est la plus courte
// (on préfère une caisse où quelqu'un encaisse).
func (p *Place) pickRegister() *Station {
	var best *Station
	bs := 1e9
	for _, r := range p.Registers() {
		sc := float64(p.QueueAt(r))
		if !p.manned(r) {
			sc += 3
		}
		if sc < bs {
			best, bs = r, sc
		}
	}
	return best
}

func (p *Place) customerStep(c *Customer, dt float64) {
	speed := 165.0
	if p.ID == "super" {
		speed *= p.superCustSpeed()
	}
	pat := 1 + p.ptsF("pts_patience") // patience (collège)
	if c.InCar() {
		return // encore en route vers le parking
	}
	if c.State == cLeave {
		p.cartOut(c)
		p.leaveByCar(c)
	}
	if len(c.path) > 0 {
		if p.cWalk(c, c.path[0], speed, dt) {
			p.atCartSpot(c)
			c.path = c.path[1:]
			if len(c.path) == 0 && c.State == cLeave {
				c.State = cGone
				if c.veh != nil && c.veh.State == 1 {
					p.depart(c.veh)
				}
			}
			if len(c.path) == 0 && c.State == cEnter {
				c.State = cShop
			}
		}
		return
	}
	switch c.State {
	case cShop:
		if c.idx >= len(c.Wants) {
			got := 0
			for _, w := range c.Wants {
				got += w.Got
			}
			if got == 0 {
				c.Mood = 2
				c.State = cLeave
				c.path = []V{p.Inside, p.Gate, p.SpawnL}
				if c.Pos.X > p.Gate.X {
					c.path[2] = p.SpawnR
				}
				if p.ID == "super" {
					c.path = p.superLeavePath(c)
				}
				p.g.emit(Event{Kind: EvAngry, Pos: c.Pos})
				p.g.Stats.Angry++
				p.reputation(-0.12)
				return
			}
			c.State = cQueue
			c.reg = p.pickRegister()
			if p.ID == "super" {
				c.reg = p.superRegister(c)
			}
			p.queueSeq++
			c.ticket = p.queueSeq // ordre d'arrivée à la caisse
			c.Wait = 0
			return
		}
		w := &c.Wants[c.idx]
		s := p.shelfFor(w.It)
		if s == nil {
			c.idx++
			return
		}
		spot := s.Spot.Add(V{float64((c.Look%5)-2) * 10, 0})
		if !p.cWalk(c, spot, speed, dt) {
			return
		}
		c.Face = V{0, -1}
		c.tick -= dt
		mine := w.Got < w.Booked // un article qu'il a réservé (ou pas)
		// (on ne fait pas attendre quelqu'un devant un rayon plein : s'il reste des
		// articles, il se sert, même s'ils étaient réservés par d'autres, qui
		// attendront à leur tour le magasinier)
		if s.Stock > 0 {
			if c.tick <= 0 {
				s.Stock--
				if p.OnlyStocked && mine && s.booked > 0 {
					s.booked--
				}
				if s.booked > s.Stock {
					s.booked = s.Stock // (pas plus de réservations que d'articles)
				}
				w.Got++
				c.Carry = append(c.Carry, w.It)
				c.tick = 0.35
				if c.Mood == 1 {
					c.Mood = 0
				}
			}
		} else {
			c.Wait += dt
			if c.Wait > 7*pat {
				c.Mood = 1
			}
		}
		// il en a déjà un peu et ça tarde : il se contente de ce qu'il a
		// (au supermarché : dès qu'il a au moins 3 articles dans son caddie, sans râler)
		if (w.Got > 0 || (p.ID == "super" && c.gotAll() >= 3)) && c.Wait > 4*pat {
			p.unbook(s, w.Booked-w.Got)
			w.N = w.Got
		}
		if w.Got >= w.N || c.Wait > 15*pat {
			if w.Got < w.N {
				p.unbook(s, w.Booked-w.Got)
			}
			if c.Wait > 15*pat && w.Got < w.N {
				// rayon vide trop longtemps : il passe au suivant, mécontent
				c.Mood = max(c.Mood, 1)
				p.g.Stats.Missed++
			}
			c.idx++
			c.Wait = 0
		}
	case cQueue:
		// place dans la file = nombre de clients en file devant
		reg := c.reg
		pos := 0
		for _, o := range p.Customers {
			if o != c && o.State == cQueue && o.reg == reg && o.ticket < c.ticket {
				pos++
			}
		}
		c.slot = pos
		if !p.cWalk(c, p.queueSlot(reg, pos), speed, dt) {
			return
		}
		c.Face = V{0, -1}
		c.Wait += dt
		if c.Wait > 25*pat && c.Mood == 0 { // (on patiente plus longtemps en caisse)
			c.Mood = 1
		}
		if pos == 0 && p.manned(reg) {
			if c.CheckN == 0 {
				n := 0
				for _, w := range c.Wants {
					n += w.Got
				}
				c.CheckN = float64(max(1, n))
				c.CheckT = c.CheckN
			}
			// chaque objet prend un peu de temps à passer
			c.CheckT -= dt / p.CashTime(reg)
			c.Wait -= dt * 0.8 // on patiente mieux quand on est servi
			if c.CheckT <= 0 {
				// paiement
				total := 0.0
				for _, w := range c.Wants {
					total += Items[w.It].Price * float64(w.Got)
				}
				total *= p.g.priceMult(p) * p.SellMult
				p.g.earn(total, reg.Body.C().Add(V{0, -40}), p.ID)
				p.g.Stats.Served++
				p.Served++
				switch c.Mood {
				case 0:
					p.reputation(p.RepGain)
					p.g.Stats.Happy++
				case 1:
					p.reputation(0.002)
				default:
					p.reputation(-0.02)
				}
				c.Paid = total
				c.Carry = nil
				c.State = cLeave
				c.path = p.leavePath(c)
				return
			}
		}
		if c.Wait > 90*pat {
			// personne pour encaisser : il repose tout et s'en va
			c.Mood = 2
			c.State = cLeave
			c.Carry = nil
			p.g.Stats.Missed++
			p.g.Stats.Angry++
			p.reputation(-0.15)
			c.path = p.leavePath(c)
			p.g.emit(Event{Kind: EvAngry, Pos: c.Pos})
		}
	}
}

// leavePath : chemin de sortie, par le portail, du côté le plus proche.
func (p *Place) leavePath(c *Customer) []V {
	if p.ID == "super" {
		return p.superLeavePath(c)
	}
	end := p.SpawnL
	if p.rng.Intn(2) == 0 {
		end = p.SpawnR
	}
	return []V{c.Pos.Add(p.LeaveOff), p.Inside.Add(V{(p.rand() - 0.5) * 80, 60}), p.Gate, end}
}

// gotAll : articles déjà dans son panier (ou son caddie).
func (c *Customer) gotAll() int {
	n := 0
	for _, w := range c.Wants {
		n += w.Got
	}
	return n
}

func (p *Place) unbook(s *Station, n int) {
	if p.OnlyStocked && n > 0 { // (n < 0 : il a pris plus que sa réservation)
		s.booked = max(0, s.booked-n)
	}
}
