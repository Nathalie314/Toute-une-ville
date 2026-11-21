package sim

import (
	"container/heap"
	"math"
)

// Recherche de chemin (A*) sur une grille, pour que les PNJ contournent
// les bâtiments au lieu de buter dedans.

const navCell = 30.0

type navGrid struct {
	w, h    int
	blocked []bool
	built   int     // nombre de stations construites quand la grille a été faite
	oy      float64 // haut de la grille (la ferme monte au-dessus de 0)
	// tableaux de travail de A*, gardés d'une recherche à l'autre (avant : trois grands
	// tableaux neufs à chaque trajet, ce qui faisait travailler le ramasse-miettes et ramer
	// le jeu) ; une case n'est valable que si son tampon est celui de la recherche en cours
	gs   []float64
	prev []int32
	seen []uint32 // tampon : gs et prev valables
	shut []uint32 // tampon : case fermée
	gen  uint32
	open navHeap
}

func (p *Place) builtCount() int {
	n := 0
	for _, s := range p.Stations {
		if s.Built {
			n++
		}
	}
	return n
}

func (p *Place) grid() *navGrid { return p.gridFor(false) }

// gridFor : la grille des PNJ ; cust : celle des clients (sans les zones interdites).
func (p *Place) gridFor(cust bool) *navGrid {
	if cust && len(p.NoCust) == 0 {
		cust = false
	}
	if !cust && p.nav != nil && p.nav.built == p.builtCount() {
		return p.nav
	}
	if cust && p.navC != nil && p.navC.built == p.builtCount() {
		return p.navC
	}
	g := &navGrid{w: int(p.W/navCell) + 1, h: int((p.H-p.Top)/navCell) + 1, built: p.builtCount(), oy: p.Top}
	g.blocked = make([]bool, g.w*g.h)
	for y := 0; y < g.h; y++ {
		for x := 0; x < g.w; x++ {
			c := V{(float64(x) + 0.5) * navCell, g.oy + (float64(y)+0.5)*navCell}
			g.blocked[y*g.w+x] = p.solidAt(c, 20) || (cust && p.noCust(c, 20))
		}
	}
	if cust {
		p.navC = g
	} else {
		p.nav = g
	}
	return g
}

func (p *Place) noCust(q V, r float64) bool {
	for _, z := range p.NoCust {
		if z.Grow(r).Has(q) {
			return true
		}
	}
	return false
}

// clearC, routeC : pareil, pour un client (il évite aussi les zones interdites).
func (p *Place) clearC(a, b V) bool {
	if !p.clear(a, b) {
		return false
	}
	d := b.Sub(a)
	n := int(d.Len()/12) + 1
	for i := 1; i <= n; i++ {
		if p.noCust(a.Add(d.Mul(float64(i)/float64(n))), 17) {
			return false
		}
	}
	return true
}

// clear : peut-on aller tout droit de a à b sans toucher de bâtiment ?
func (p *Place) clear(a, b V) bool {
	d := b.Sub(a)
	n := int(d.Len()/12) + 1
	for i := 1; i <= n; i++ {
		q := a.Add(d.Mul(float64(i) / float64(n)))
		if p.solidAt(q, 17) {
			return false
		}
	}
	return true
}

type navNode struct {
	i int32
	f float64
}
type navHeap []navNode

func (h navHeap) Len() int            { return len(h) }
func (h navHeap) Less(i, j int) bool  { return h[i].f < h[j].f }
func (h navHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *navHeap) Push(x interface{}) { *h = append(*h, x.(navNode)) }
func (h *navHeap) Pop() interface{} {
	old := *h
	n := old[len(old)-1]
	*h = old[:len(old)-1]
	return n
}

// route renvoie une suite de points de passage de from à to (sans from).
func (p *Place) route(from, to V) []V  { return p.routeOn(p.grid(), from, to, p.clear) }
func (p *Place) routeC(from, to V) []V { return p.routeOn(p.gridFor(true), from, to, p.clearC) }

func (p *Place) routeOn(g *navGrid, from, to V, clear func(a, b V) bool) []V {
	cell := func(v V) (int, int) {
		x := int(v.X / navCell)
		y := int((v.Y - g.oy) / navCell)
		return max(0, min(g.w-1, x)), max(0, min(g.h-1, y))
	}
	free := func(x, y int) bool {
		return x >= 0 && y >= 0 && x < g.w && y < g.h && !g.blocked[y*g.w+x]
	}
	// départ / arrivée : la case libre la plus proche
	nearest := func(v V) (int, int) {
		x, y := cell(v)
		if free(x, y) {
			return x, y
		}
		for r := 1; r < 8; r++ {
			for dy := -r; dy <= r; dy++ {
				for dx := -r; dx <= r; dx++ {
					if free(x+dx, y+dy) {
						return x + dx, y + dy
					}
				}
			}
		}
		return x, y
	}
	sx, sy := nearest(from)
	tx, ty := nearest(to)
	start, goal := sy*g.w+sx, ty*g.w+tx
	n := g.w * g.h
	if len(g.gs) != n {
		g.gs, g.prev = make([]float64, n), make([]int32, n)
		g.seen, g.shut = make([]uint32, n), make([]uint32, n)
		g.gen = 0
	}
	g.gen++
	if g.gen == 0 { // (le compteur a fait le tour : on repart de zéro)
		for i := range g.seen {
			g.seen[i], g.shut[i] = 0, 0
		}
		g.gen = 1
	}
	gen := g.gen
	gsAt := func(i int) float64 {
		if g.seen[i] != gen {
			return math.Inf(1)
		}
		return g.gs[i]
	}
	hfn := func(i int) float64 {
		dx := math.Abs(float64(i%g.w - tx))
		dy := math.Abs(float64(i/g.w - ty))
		return math.Max(dx, dy) + 0.414*math.Min(dx, dy)
	}
	open := &g.open
	*open = (*open)[:0]
	g.gs[start], g.prev[start], g.seen[start] = 0, -1, gen
	heap.Push(open, navNode{i: int32(start), f: hfn(start)})
	found := false
	for open.Len() > 0 {
		cur := int(heap.Pop(open).(navNode).i)
		if g.shut[cur] == gen {
			continue
		}
		if cur == goal {
			found = true
			break
		}
		g.shut[cur] = gen
		cx, cy := cur%g.w, cur/g.w
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}
				nx, ny := cx+dx, cy+dy
				if !free(nx, ny) {
					continue
				}
				// pas de coin coupé
				if dx != 0 && dy != 0 && (!free(cx+dx, cy) || !free(cx, cy+dy)) {
					continue
				}
				ni := ny*g.w + nx
				cost := 1.0
				if dx != 0 && dy != 0 {
					cost = 1.414
				}
				if v := g.gs[cur] + cost; v < gsAt(ni) {
					g.gs[ni], g.prev[ni], g.seen[ni] = v, int32(cur), gen
					heap.Push(open, navNode{i: int32(ni), f: v + hfn(ni)})
				}
			}
		}
	}
	if !found {
		return []V{to}
	}
	var pts []V
	for i := goal; i != -1 && i != start; i = int(g.prev[i]) {
		pts = append(pts, V{(float64(i%g.w) + 0.5) * navCell, g.oy + (float64(i/g.w)+0.5)*navCell})
	}
	// à l'endroit
	for i, j := 0, len(pts)-1; i < j; i, j = i+1, j-1 {
		pts[i], pts[j] = pts[j], pts[i]
	}
	pts = append(pts, to)
	// on lisse : on va tout droit tant que la ligne droite reste libre (avant : on essayait
	// chaque point depuis le bout du chemin, ce qui coûtait le carré de sa longueur)
	var out []V
	cur := from
	for i := 0; i < len(pts); {
		j := i
		for j+1 < len(pts) && clear(cur, pts[j+1]) {
			j++
		}
		out = append(out, pts[j])
		cur = pts[j]
		i = j + 1
	}
	return out
}
