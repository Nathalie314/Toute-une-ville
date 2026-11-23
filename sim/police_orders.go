package sim

import "math"

// Les commandes du commissariat et de la prison (comme la cantine de l'école) : selon les gens
// présents, chacun selon son rang.
//
// Le commissariat (dès qu'il a son commissaire) :
//   boulangerie : brioche (commissaire), pain au chocolat (capitaines), croissant (lieutenants),
//                 baguette (voleurs en garde à vue : leur sandwich)
//   pâtisserie  : fraisier (commissaire), tarte au citron meringuée (capitaines), éclair au café (lieutenants)
//   ferme exotique : un café moulu par policier (la machine à café du commissariat)
//   livrée : les policiers coûtent 2 fois moins de points d'école (pub auprès de leurs amis) ;
//   ratée : 2 fois plus, pendant le temps du cadeau.
// La prison (dès qu'elle a son directeur) :
//   boulangerie : baguette (détenus), croissant (personnel), pain au chocolat (directeur)
//   crèmerie    : yaourt de vache, fromage de vache, fromage de brebis
//   boucherie   : steak haché, rôti de bœuf, foie gras        } seulement si la cantine
//   ferme       : pommes de terre, salade, fraises            } de la prison est ouverte
//   livrée : chaque besoin rempli retire 2 fois plus de points de récidive ; ratée : 2 fois moins.

const (
	whoPolice = "le commissariat"
	whoPrison = "la prison"
)

// policeCrew : le commissaire, les capitaines, les lieutenants.
func (g *Game) policeCrew() (chief, capt, lieut int) {
	if g.HasChief() {
		chief = 1
	}
	for n := 1; n <= PolOffices(); n++ {
		if g.PolOfficeOwned(n) && g.PolCaptain(n) {
			capt++
		}
	}
	return chief, capt, g.PolLieutenants()
}

// prisonCrew : les détenus, le personnel (surveillants, psy, assistant social), le directeur.
func (g *Game) prisonCrew() (inmates, staff, warden int) {
	inmates, _, _ = g.PrisonStats()
	for k := 1; k <= PrisCellN/2; k++ {
		if g.prisLv("surveillant_"+itoa(k)) > 0 {
			staff++
		}
	}
	staff += g.prisLv("psy") + g.prisLv("as")
	if g.HasWarden() {
		warden = 1
	}
	return
}

// crewOrder : une commande à partir de « tant de ce produit ».
func (g *Game) crewOrder(who, at string, lines [][2]int) *Order {
	o := &Order{Who: who, At: at, Left: schoolTime, Total: schoolTime, Fast: 120}
	value := 0.0
	for _, l := range lines {
		it, n := Item(l[0]), l[1]
		if n <= 0 {
			continue
		}
		merged := false
		for i := range o.Lines {
			if o.Lines[i].It == it {
				o.Lines[i].N += n
				merged = true
			}
		}
		if !merged {
			o.Lines = append(o.Lines, OrderLine{It: it, N: n})
		}
		value += Items[it].Price * float64(n)
	}
	if len(o.Lines) == 0 {
		return nil
	}
	o.Reward = math.Round(value*3/10) * 10
	return o
}

// policeOrder : la commande du commissariat dans ce lieu (nil : rien à y commander).
func (g *Game) policeOrder(at string) *Order {
	c, k, l := g.policeCrew()
	if c == 0 {
		return nil
	}
	switch at {
	case "boulangerie":
		return g.crewOrder(whoPolice, at, [][2]int{{int(Brioche), c}, {int(PainChoco), k}, {int(Croissant), l}, {int(Pain), g.InCustody()}})
	case "patisserie":
		return g.crewOrder(whoPolice, at, [][2]int{{int(Fraisier), c}, {int(TarteCitron), k}, {int(Eclair), l}})
	case "exotique":
		return g.crewOrder(whoPolice, at, [][2]int{{int(CafeMoulu), c + k + l}})
	}
	return nil
}

// prisonOrder : la commande de la prison dans ce lieu (nil : rien à y commander).
func (g *Game) prisonOrder(at string) *Order {
	in, st, w := g.prisonCrew()
	if w == 0 {
		return nil
	}
	cantine := g.prisLv("cantine") > 0
	switch at {
	case "boulangerie":
		return g.crewOrder(whoPrison, at, [][2]int{{int(Pain), in}, {int(Croissant), st}, {int(PainChoco), w}})
	case "cremerie":
		return g.crewOrder(whoPrison, at, [][2]int{{int(Yaourt), in}, {int(Fromage), st}, {int(FromageBrebis), w}})
	case "boucherie":
		if cantine {
			return g.crewOrder(whoPrison, at, [][2]int{{int(SteakHache), in}, {int(RotiBoeuf), st}, {int(FoieGras), w}})
		}
	case "ferme":
		if cantine {
			return g.crewOrder(whoPrison, at, [][2]int{{int(PommeDeTerre), in}, {int(Chou), st}, {int(Fraise), w}})
		}
	}
	return nil
}

// PtsMult : le cadeau du commissariat sur le prix (en points d'école) des policiers.
func (g *Game) ptsMult(key string) float64 {
	if len(key) < len("commissariat.") || key[:len("commissariat.")] != "commissariat." {
		return 1
	}
	if g.Boosted("police", "") {
		return 1 / float64(g.Boost.Mul())
	}
	if g.Boosted("police_malus", "") {
		return 2
	}
	return 1
}

// reinsertMult : le cadeau de la prison sur ce que retire un besoin rempli.
func (g *Game) reinsertMult(lvl int) int {
	if g.Boosted("prison", "") {
		return lvl * g.Boost.Mul()
	}
	if g.Boosted("prison_malus", "") {
		return lvl / 2
	}
	return lvl
}
