package sim

import (
	"math"
	"sort"
	"strings"
)

// Rayons vides du supermarché : on cherche quel maillon de la chaîne bloque
// (production → étal du lieu → drone → quai → camion → entrepôt → magasinier → rayon).

// Block : la cause d'un rayon presque vide.
type Block int

const (
	BlockNone  Block = iota
	BlockStaff       // la palette de l'entrepôt est pleine : les magasiniers ne suivent pas
	BlockTruck       // le quai du lieu d'origine déborde : le camion n'emporte pas tout
	BlockDrone       // l'étal du lieu d'origine est plein : le drone ne suit pas (ou il n'y en a pas)
	BlockProd        // le lieu d'origine ne produit pas assez
	BlockLock        // le produit n'est plus fabriqué nulle part
	NBlocks
)

// rayonLow : un rayon est « presque vide » sous le quart de sa contenance.
func rayonLow(s *Station) bool { return s.Stock*4 < s.StockMax }

// rayonSource : le lieu qui envoie ce produit au supermarché, son étal et son quai.
func (g *Game) rayonSource(it Item) (*Place, *Station, *Station) {
	for _, p := range g.Places {
		if p.ID == "super" || !p.Owned {
			continue
		}
		et := p.shelfFor(it)
		if et == nil {
			continue
		}
		for _, q := range p.Stations {
			if q.Kind == Dock && q.To == "super" {
				return p, et, q
			}
		}
	}
	return nil, nil, nil
}

// rayonNow : ce qui bloque ce rayon en ce moment.
func (g *Game) rayonNow(sp *Place, s *Station) Block {
	if !rayonLow(s) {
		return BlockNone
	}
	if pal := sp.pallet(s.Item); pal != nil && pal.Pile >= 10 {
		return BlockStaff
	}
	from, et, dk := g.rayonSource(s.Item)
	if from == nil {
		return BlockLock
	}
	if dk.Buf[s.Item] > 0 && (dk.lastFull || dk.Buf[s.Item]*2 >= dk.BufMax) {
		return BlockTruck
	}
	if g.Levels[from.key(droneKey(s.Item))] == 0 {
		if et.Stock > 0 {
			return BlockDrone
		}
		return BlockProd
	}
	capa := DroneCap(g.Levels[from.key("pts_dr_cap")])
	if et.Stock >= max(3, min(capa, et.StockMax/2)) {
		return BlockDrone
	}
	return BlockProd
}

// stepRayonDiag : on retient ce qui a bloqué chaque rayon pendant les ~30 dernières
// secondes (sinon l'icône changerait sans arrêt).
func (g *Game) stepRayonDiag(dt float64) {
	sp := g.Place("super")
	if sp == nil || !sp.Owned {
		return
	}
	g.diagT += dt
	if g.diagT < 0.5 {
		return
	}
	step := g.diagT
	g.diagT = 0
	k := math.Exp(-step / 30)
	for _, s := range sp.Stations {
		if s.Kind != Shelf || !s.Built {
			continue
		}
		for i := range s.diag {
			s.diag[i] *= k
		}
		s.diag[g.rayonNow(sp, s)] += step
	}
}

// RayonBlock : pourquoi ce rayon est presque vide (BlockNone s'il ne l'est pas).
func (g *Game) RayonBlock(s *Station) Block {
	if s.Kind != Shelf || !s.Built || !rayonLow(s) || !strings.HasPrefix(s.ID, "rayon_") {
		return BlockNone // (seulement les rayons du supermarché)
	}
	best, bv := BlockNone, 0.0
	for b := BlockStaff; b < NBlocks; b++ {
		if s.diag[b] > bv {
			best, bv = b, s.diag[b]
		}
	}
	if best == BlockNone {
		best = g.rayonNow(g.Place("super"), s)
	}
	return best
}

// placeThe : « la ferme », « le glacier », « le port »…
func placeThe(id string) string {
	t := placeTo(id)
	switch {
	case len(t) > 3 && t[:3] == "au ":
		return "le " + t[3:]
	case len(t) > 5 && t[:5] == "chez ":
		return t[5:]
	}
	return t
}

// missingInput : l'ingrédient qui manque à la machine qui fabrique ce produit
// (-1 : rien ne manque, ou ce n'est pas une machine), et le lieu d'où il vient.
func (g *Game) missingInput(from *Place, it Item) (Item, *Place) {
	if from == nil {
		return -1, nil
	}
	for _, m := range from.Stations {
		if m.Kind != Machine || !m.Built || m.Item != it {
			continue
		}
		for _, in := range m.Inputs {
			if m.Buf[in] > 0 {
				continue
			}
			pal := from.pallet(in)
			if pal != nil && pal.Pile > 0 {
				continue
			}
			if pal == nil {
				return in, from // (fait sur place : la crème du beurre)
			}
			for _, q := range g.Places {
				if !q.Owned || q.shelfFor(in) == nil {
					continue
				}
				for _, d := range q.Stations {
					if d.Kind == Dock && d.To == from.ID {
						return in, q
					}
				}
			}
			return in, nil
		}
	}
	return -1, nil
}

// RayonWhy : la bulle d'un rayon presque vide (deux lignes : le constat, puis le remède).
func (g *Game) RayonWhy(s *Station) (string, string) {
	b := g.RayonBlock(s)
	name := Items[s.Item].Name
	from, _, _ := g.rayonSource(s.Item)
	id := ""
	if from != nil {
		id = from.ID
	}
	switch b {
	case BlockStaff:
		z := superZone(SuperZoneOf(s.Item))
		if z != nil && g.Levels["super."+teamKey(z.ID)] < len(z.Items) {
			return name + " : la palette est pleine, mais personne ne remplit le rayon.",
				"→ Embauche un magasinier de plus (" + z.Name + ")."
		}
		return name + " : la palette est pleine, mais les magasiniers ne suivent pas.",
			"→ Points d'école « Entraînement » au supermarché (marche, mains expertes, charge)."
	case BlockTruck:
		return name + " : le quai " + placeOf(id) + " déborde, le camion n'emporte pas tout.",
			"→ Points de mécaniciens « Camions : charge » ou « Camions : départs » " + placeAt(id) + "."
	case BlockDrone:
		if from != nil && g.Levels[from.key(droneKey(s.Item))] == 0 {
			return name + " : l'étal " + placeOf(id) + " en a, mais il n'a pas de drone.",
				"→ Achète le drone de cet étal " + placeAt(id) + "."
		}
		return name + " : l'étal " + placeOf(id) + " en a, mais le drone ne suit pas.",
			"→ Points de mécaniciens « Drones : charge » ou « Drones : vitesse » " + placeAt(id) + "."
	case BlockProd:
		if in, src := g.missingInput(from, s.Item); in >= 0 {
			what := name + " : " + placeThe(id) + " attend son ingrédient « " + Items[in].Name + " »."
			switch {
			case src == from:
				return what, "→ Il se fabrique sur place : développe sa machine " + placeAt(id) + "."
			case src != nil:
				return what, "→ Il vient " + placeOf(src.ID) + " : sa production, son drone, son camion vers " + placeThe(id) + "."
			}
			return what, "→ " + whereLocked(in) + "."
		}
		return name + " : " + placeThe(id) + " n'en produit pas assez.",
			"→ Développe sa production " + placeAt(id) + " (niveaux, ouvriers, points d'école)."
	case BlockLock:
		return name + " : plus personne n'en fabrique.", "→ " + whereLocked(s.Item) + "."
	}
	return "", ""
}

// RayonSummary : les rayons presque vides, regroupés par cause (les deux plus nombreuses).
func (g *Game) RayonSummary() string {
	sp := g.Place("super")
	if sp == nil || !sp.Owned {
		return ""
	}
	type grp struct {
		b    Block
		from string
		in   bool // il manque un ingrédient
		n    int
	}
	var gs []*grp
	for _, s := range sp.Stations {
		b := g.RayonBlock(s)
		if b == BlockNone || b == BlockLock {
			continue
		}
		from, in := "", false
		if b != BlockStaff {
			if p, _, _ := g.rayonSource(s.Item); p != nil {
				from = p.ID
				if b == BlockProd {
					m, _ := g.missingInput(p, s.Item)
					in = m >= 0
				}
			}
		}
		var hit *grp
		for _, x := range gs {
			if x.b == b && x.from == from && x.in == in {
				hit = x
			}
		}
		if hit == nil {
			hit = &grp{b: b, from: from, in: in}
			gs = append(gs, hit)
		}
		hit.n++
	}
	if len(gs) == 0 {
		return ""
	}
	sort.SliceStable(gs, func(i, j int) bool { return gs[i].n > gs[j].n })
	txt := "Rayons vides : "
	for i, x := range gs {
		if i == 2 {
			break
		}
		if i > 0 {
			txt += " ; "
		}
		n, v := itoa(x.n), " attendent "
		if x.n == 1 {
			v = " attend "
		}
		if i > 0 && gs[0].b != BlockStaff {
			v = " " // (« 3 attendent les drones de… ; 2 le camion de… »)
		}
		switch x.b {
		case BlockStaff:
			txt += n + " faute de magasiniers"
		case BlockTruck:
			txt += n + v + "le camion " + placeOf(x.from)
		case BlockDrone:
			txt += n + v + "les drones " + placeOf(x.from)
		case BlockProd:
			if x.in {
				txt += n + v + "des ingrédients " + placeAt(x.from)
			} else {
				txt += n + v + "la production " + placeOf(x.from)
			}
		}
	}
	return txt + "."
}

// RayonMissing : pour l'icône, l'ingrédient qui manque en amont (-1 : aucun).
func (g *Game) RayonMissing(s *Station) Item {
	if g.RayonBlock(s) != BlockProd {
		return -1
	}
	from, _, _ := g.rayonSource(s.Item)
	in, _ := g.missingInput(from, s.Item)
	return in
}
