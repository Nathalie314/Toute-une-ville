package sim

import "math"

// V est un point ou un vecteur du monde (en pixels « logiques »).
type V struct{ X, Y float64 }

func (a V) Add(b V) V             { return V{a.X + b.X, a.Y + b.Y} }
func (a V) Sub(b V) V             { return V{a.X - b.X, a.Y - b.Y} }
func (a V) Mul(k float64) V       { return V{a.X * k, a.Y * k} }
func (a V) Len() float64          { return math.Hypot(a.X, a.Y) }
func (a V) Dist(b V) float64      { return a.Sub(b).Len() }
func (a V) Lerp(b V, t float64) V { return a.Add(b.Sub(a).Mul(t)) }
func (a V) Norm() V {
	l := a.Len()
	if l < 1e-9 {
		return V{}
	}
	return a.Mul(1 / l)
}

// Rect est un rectangle aligné sur les axes.
type Rect struct{ X, Y, W, H float64 }

func (r Rect) Has(p V) bool {
	return p.X >= r.X && p.X <= r.X+r.W && p.Y >= r.Y && p.Y <= r.Y+r.H
}
func (r Rect) C() V { return V{r.X + r.W/2, r.Y + r.H/2} }
func (r Rect) Grow(d float64) Rect {
	return Rect{r.X - d, r.Y - d, r.W + 2*d, r.H + 2*d}
}
