package sim

import (
	"math"
	"sort"
	"strings"
	"ville/tr"
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
	name := tr.T(Items[s.Item].Name)
	from, _, _ := g.rayonSource(s.Item)
	id := ""
	if from != nil {
		id = from.ID
	}
	// {alieu} / {delieu} : « à la crèmerie » / « de la crèmerie » (français) ; {lieu} : « the dairy »
	at := func(k string, kv ...string) string {
		return tr.F(k, append([]string{"produit", name, "alieu", placeAt(id), "delieu", placeOf(id), "lelieu", placeThe(id), "lieu", theName(id)}, kv...)...)
	}
	switch b {
	case BlockStaff:
		z := superZone(SuperZoneOf(s.Item))
		if z != nil && g.Levels["super."+teamKey(z.ID)] < len(z.Items) {
			return at("{produit} : la palette est pleine, mais personne ne remplit le rayon."),
				at("→ Embauche un magasinier de plus ({zone}).", "zone", tr.T(z.Name))
		}
		return at("{produit} : la palette est pleine, mais les magasiniers ne suivent pas."),
			tr.T("→ Points d'école « Entraînement » au supermarché (marche, mains expertes, charge).")
	case BlockTruck:
		return at("{produit} : le quai {delieu} déborde, le camion n'emporte pas tout."),
			at("→ Points de mécaniciens « Camions : charge » ou « Camions : départs » {alieu}.")
	case BlockDrone:
		if from != nil && g.Levels[from.key(droneKey(s.Item))] == 0 {
			return at("{produit} : l'étal {delieu} en a, mais il n'a pas de drone."),
				at("→ Achète le drone de cet étal {alieu}.")
		}
		return at("{produit} : l'étal {delieu} en a, mais le drone ne suit pas."),
			at("→ Points de mécaniciens « Drones : charge » ou « Drones : vitesse » {alieu}.")
	case BlockProd:
		if in, src := g.missingInput(from, s.Item); in >= 0 {
			what := at("{produit} : {lelieu} attend son ingrédient « {ingredient} ».", "ingredient", tr.T(Items[in].Name))
			switch {
			case src == from:
				return what, at("→ Il se fabrique sur place : développe sa machine {alieu}.")
			case src != nil:
				return what, at("→ Il vient {desource} : sa production, son drone, son camion vers {lelieu}.", "desource", placeOf(src.ID), "source", theName(src.ID))
			}
			return what, "→ " + tr.T(whereLocked(in)) + "."
		}
		return at("{produit} : {lelieu} n'en produit pas assez."),
			at("→ Développe sa production {alieu} (niveaux, ouvriers, points d'école).")
	case BlockLock:
		return at("{produit} : plus personne n'en fabrique."), "→ " + tr.T(whereLocked(s.Item)) + "."
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
	var parts []string
	for i, x := range gs {
		if i == 2 {
			break
		}
		// (« attend » au singulier, « attendent » au pluriel : deux phrases à traduire)
		k := map[Block]string{BlockStaff: "{n} faute de magasiniers", BlockTruck: "{n} attendent le camion {delieu}",
			BlockDrone: "{n} attendent les drones {delieu}", BlockProd: "{n} attendent la production {delieu}"}[x.b]
		if x.b == BlockProd && x.in {
			k = "{n} attendent des ingrédients {alieu}"
		}
		if x.n == 1 {
			k = strings.Replace(k, "attendent", "attend", 1)
		}
		parts = append(parts, tr.F(k, "n", itoa(x.n), "alieu", placeAt(x.from), "delieu", placeOf(x.from), "lieu", theName(x.from)))
	}
	return tr.F("Rayons vides : {liste}.", "liste", strings.Join(parts, " ; "))
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
