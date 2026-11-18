package sim

import (
	"math"
	"ville/tr"
)

// Les commandes spéciales : de temps en temps, un habitant commande plusieurs
// produits de la ferme, à livrer dans la caisse des commandes avant la fin du
// temps imparti. La récompense vaut bien plus que le prix normal.

// OrderLine : une ligne de commande.
type OrderLine struct {
	It  Item
	N   int
	Got int
}

// Order : une commande en cours.
type Order struct {
	Who    string
	At     string `json:",omitempty"` // lieu où livrer ("" : la ferme)
	Lines  []OrderLine
	Left   float64 // secondes restantes
	Total  float64 // durée totale
	Reward float64
	Fast   float64 `json:",omitempty"` // livrée en moins de Fast secondes : cadeau doublé
	Gift   string  `json:",omitempty"` // cadeau (voir gifts.go ; "" : ancienne commande payée en argent)
	Dur    float64 `json:",omitempty"` // durée du cadeau (réclame comprise)
}

// FastLeft : secondes qui restent pour la livraison express (gain doublé), 0 si c'est trop tard.
func (o *Order) FastLeft() float64 { return math.Max(0, o.Fast-(o.Total-o.Left)) }

// Place : lieu où livrer la commande.
func (o *Order) Place() string {
	if o.At == "" {
		return "ferme"
	}
	return o.At
}

// sig : la commande en quelques mots (qui + quoi), pour éviter les répétitions.
func (o *Order) sig() string {
	s := o.Who
	for _, l := range o.Lines {
		s += "|" + itoa(int(l.It))
	}
	return s
}

// Need : combien il manque de ce produit.
func (o *Order) Need(it Item) int {
	for _, l := range o.Lines {
		if l.It == it {
			return l.N - l.Got
		}
	}
	return 0
}

func (o *Order) add(it Item) {
	for i := range o.Lines {
		if o.Lines[i].It == it && o.Lines[i].Got < o.Lines[i].N {
			o.Lines[i].Got++
			return
		}
	}
}

// Done : tout est livré.
func (o *Order) Done() bool {
	for _, l := range o.Lines {
		if l.Got < l.N {
			return false
		}
	}
	return true
}

var orderWho = []string{"Mme Martin", "M. Dupont", "la famille Leroy", "Mamie Jeanne", "le voisin Paul", "Mme Petit", "M. Garnier",
	"les jumeaux Morel", "Mlle Rousseau", "le facteur", "le maire", "Mme Bernard", "M. Lefèvre", "la famille Moreau", "Papi Lucien",
	"le docteur Simon", "Mme Laurent", "M. Michel", "la famille Fontaine", "Tante Odette", "le curé", "Mme Girard", "M. Bonnet",
	"les sœurs Dubois", "le gendarme", "Mme Mercier", "M. Blanc", "la famille Faure", "le pharmacien", "Mme Chevalier", "M. Perrin",
	"la maîtresse d'école", "le coiffeur", "Mme Robin", "M. Gauthier", "la famille Roux", "le boulanger d'en face", "Mlle Colin", "M. Vincent", "la fleuriste",
	"le garagiste", "le vétérinaire", "le marin"}

const orderTime = 300.0

// orderBox : la caisse des commandes du lieu où livrer.
func (g *Game) orderBox(o *Order) (*Place, *Station) {
	p := g.Place(o.Place())
	if p == nil {
		p = g.farm()
	}
	return p, p.Station("commande")
}

// Clients qui passent des commandes : la ferme (des habitants) et, dès que
// l'école et la boulangerie sont ouvertes, la cantine de l'école. Chaque
// nouvelle commande vient d'un client tiré au hasard, à parts égales.
// (Plus tard : hôpital, prison…)
func (g *Game) orderSources() []func() *Order {
	src := []func() *Order{g.farmOrder}
	// boulangerie, crèmerie, glacier : des particuliers ou la cantine de l'école, au hasard
	for _, id := range []string{"boulangerie", "cremerie", "glacier"} {
		if c := g.Place(id); c != nil && c.Owned {
			id := id
			src = append(src, func() *Order {
				if sc := g.Place("ecole"); sc != nil && sc.Owned && g.rng.Float64() < 0.5 {
					if o := g.schoolOrder(id); o != nil {
						return o
					}
				}
				return g.shopOrder(id)
			})
		}
	}
	if c := g.Place("exotique"); c != nil && c.Owned {
		src = append(src, g.exoticOrder)
	}
	if c := g.Place("port"); c != nil && c.Owned {
		src = append(src, g.portOrder)
	}
	if c := g.Place("domaine"); c != nil && c.Owned {
		src = append(src, g.domaineOrder)
	}
	if c := g.Place("patisserie"); c != nil && c.Owned {
		src = append(src, func() *Order { return g.shopOrder("patisserie") })
	}
	// la crêperie : seulement des crêpes nature, et seulement pour la cantine de l'école (le parc,
	// lui, en commande avec ses autres commandes)
	if c := g.Place("creperie"); c != nil && c.Owned {
		if sc := g.Place("ecole"); sc != nil && sc.Owned {
			src = append(src, func() *Order { return g.schoolOrder("creperie") })
		}
	}
	if c := g.Place("parc"); c != nil && c.Owned {
		src = append(src, g.parcOrder)
	}
	return src
}

func (g *Game) stepOrders(dt float64) {
	if g.Order != nil {
		_, box := g.orderBox(g.Order)
		g.Order.Left -= dt
		if g.Order.Left <= 0 {
			// temps écoulé : ce qui a été livré est payé au prix normal (sans bonus)
			o := g.Order
			paid := 0.0
			for _, l := range o.Lines {
				paid += Items[l.It].Price * float64(l.Got)
			}
			at := box.Body.C()
			txt := tr.F("Commande : {qui} — temps écoulé…", "qui", tr.T(o.Who))
			if paid > 0 {
				g.earnBonus(paid, at, o.Place())
				txt += " " + tr.F("Ce qui a été livré est payé : +{prix}", "prix", moneyText(paid))
			}
			// commande ratée : le lieu perd une étoile de réputation ; celles du parc et de la
			// cantine ont leur propre pénalité (5 min, jamais allongées), sans double peine
			switch o.Gift {
			case "parc", "ecole":
				g.Boost = &Boost{Kind: o.Gift + "_malus", Left: BoostTime, Total: BoostTime}
				txt += " " + tr.F("(pénalité : {quoi} pendant {duree})", "quoi", lowerFirst(g.Boost.Text()), "duree", durText(BoostTime))
			default:
				if p := g.Place(o.Place()); p != nil {
					p.reputation(-1)
					txt += " " + tr.T("(−1 étoile de réputation)")
				}
			}
			g.emit(Event{Kind: EvLevel, Pos: at, Text: txt})
			g.emit(Event{Kind: EvAngry, Pos: at})
			g.Order = nil
			g.orderT = 180 + 120*g.rng.Float64()
		}
		return
	}
	if len(g.farmItems()) < 3 {
		return // pas encore de commandes
	}
	if g.Boost != nil {
		return // (un seul cadeau à la fois : la prochaine commande attend)
	}
	if g.orderT == 0 {
		g.orderT = 120 // la première arrive vite
	}
	g.orderT -= dt
	if g.orderT > 0 {
		return
	}
	src := g.orderSources()
	var o *Order
	for try := 0; try < 6; try++ {
		o = src[g.rng.Intn(len(src))]()
		if o == nil {
			o = g.farmOrder()
		}
		if o != nil && o.sig() != g.lastOrder {
			break // pas deux fois la même commande de suite
		}
	}
	g.lastOrder = o.sig()
	if q := g.Place(o.Place()); q != nil {
		m := 1 + q.ptsF("pts_logist") // logistique (lycée)
		o.Total *= m
		o.Left = o.Total
	}
	o.Gift, o.Reward = giftOf(o.Who), 0
	if p := g.Place(o.Place()); p != nil {
		o.Dur = BoostTime * (1 + p.ptsF("pts_reclame")) // réclame (points du collège)
	}
	g.Order = o
	p, box := g.orderBox(o)
	txt := tr.F("Nouvelle commande : {qui} !", "qui", tr.T(o.Who))
	if p.ID != "ferme" {
		txt += " " + tr.F("À livrer {alieu}.", "alieu", placeAt(p.ID), "lieu", theName(p.ID))
	}
	txt += GiftSep + tr.F("Cadeau : {cadeau} !", "cadeau", o.GiftText())
	g.emit(Event{Kind: EvOrder, Pos: box.Body.C(), Text: txt}) // (sonnerie de téléphone)
}

// placeTo : « à la boulangerie »…
func placeTo(id string) string {
	switch id {
	case "boulangerie":
		return "la boulangerie"
	case "cremerie":
		return "la crèmerie"
	case "exotique":
		return "la ferme exotique"
	case "glacier":
		return "chez le glacier"
	case "super":
		return "au supermarché"
	case "elevage":
		return "la ferme d'élevage"
	case "boucherie":
		return "la boucherie"
	case "ecole":
		return "l'école"
	case "port":
		return "au port"
	case "domaine":
		return "au domaine"
	case "patisserie":
		return "la pâtisserie"
	}
	return "la ferme"
}

// farmItems : produits proposés à la ferme (ceux qui ont un étal).
func (g *Game) farmItems() []Item {
	var its []Item
	for _, s := range g.farm().Stations {
		if s.Kind == Shelf && s.Built {
			its = append(its, s.Item)
		}
	}
	return its
}

// farmOrder : un habitant commande deux ou trois produits de la ferme.
func (g *Game) farmOrder() *Order {
	its := g.farmItems()
	if len(its) < 3 {
		return nil
	}
	// deux produits au hasard, parmi les plus récents de préférence
	who := g.orderPerson("ferme")
	o := &Order{Who: who, Left: orderTime, Total: orderTime, Fast: 60}
	k := 2
	if len(its) >= 6 && g.rng.Float64() < 0.4 {
		k = 3
	}
	used := map[Item]bool{}
	value := 0.0
	for len(o.Lines) < k {
		// biais vers la fin de la liste (produits plus élaborés)
		i := len(its) - 1 - int(math.Floor(math.Pow(g.rng.Float64(), 1.6)*float64(len(its))))
		it := its[max(0, min(len(its)-1, i))]
		if used[it] {
			continue
		}
		used[it] = true
		n := 5 + g.rng.Intn(8)
		o.Lines = append(o.Lines, OrderLine{It: it, N: n})
		value += Items[it].Price * float64(n)
	}
	o.Reward = math.Round(value*3/10) * 10
	return o
}

// shopOrder : un particulier commande deux ou trois produits d'un lieu, à livrer
// dans ce lieu (nil : pas encore assez de produits).
func (g *Game) shopOrder(at string) *Order {
	c := g.Place(at)
	var its []Item
	for _, s := range c.Stations {
		if s.Kind == Shelf && s.Built && !s.NoSale {
			its = append(its, s.Item)
		}
	}
	if len(its) < 2 {
		return nil
	}
	who := g.orderPerson(at)
	o := &Order{Who: who, At: at, Left: orderTime, Total: orderTime, Fast: 60}
	k := 2
	if len(its) >= 4 && g.rng.Float64() < 0.4 {
		k = 3
	}
	used := map[Item]bool{}
	value := 0.0
	for len(o.Lines) < k {
		it := its[g.rng.Intn(len(its))]
		if used[it] {
			continue
		}
		used[it] = true
		n := 4 + g.rng.Intn(6)
		o.Lines = append(o.Lines, OrderLine{It: it, N: n})
		value += Items[it].Price * float64(n)
	}
	o.Reward = math.Round(value*3/10) * 10
	return o
}

// schoolTime : la cantine laisse plus de temps (beaucoup de pains).
const schoolTime = 600.0

// schoolOrder : la cantine de l'école commande, dans cet atelier (boulangerie,
// crèmerie ou glacier), un produit pour chaque personne présente à l'école.
func (g *Game) schoolOrder(at string) *Order {
	// un produit par personne : primaire, collège, lycée, université, puis les adultes
	// (même si l'atelier ne sait pas encore le faire : la commande pousse à ouvrir la recette)
	if at == "creperie" {
		return g.schoolCrepes()
	}
	menu := map[string][]Item{
		"boulangerie": {Pain, Brioche, Croissant, PainChoco, PainMele},
		"cremerie":    {Lait, FromageBlanc, Yaourt, Fromage, Beurre},
		"glacier":     {SorbetFraise, SorbetMangue, SorbetAnanas, GlaceVanille, GlaceCafe},
	}[at]
	sc := g.Place("ecole")
	n := map[Item]int{}
	for _, st := range sc.Students {
		if st.Room == nil {
			continue
		}
		n[menu[st.Room.Level]]++
	}
	adult := menu[4]
	for _, r := range sc.Rooms {
		if r.built && sc.HasProf(r) {
			n[adult]++ // les profs
		}
	}
	n[adult] += len(sc.Workers) // les surveillants
	if g.Levels[sc.key("secretaire")] > 0 {
		n[adult]++ // la secrétaire
	}
	o := &Order{Who: "la cantine de l'école", At: at, Left: schoolTime, Total: schoolTime, Fast: 120}
	value := 0.0
	for _, it := range menu {
		if n[it] > 0 {
			o.Lines = append(o.Lines, OrderLine{It: it, N: n[it]})
			value += Items[it].Price * float64(n[it])
		}
	}
	if len(o.Lines) == 0 {
		return nil
	}
	o.Reward = math.Round(value*3/10) * 10
	return o
}

// parcOrder : le parc d'attractions commande de quoi faire tourner ses stands : du
// maïs à la ferme (le pop-corn), du sucre en poudre à la ferme exotique (la barbe à
// papa), des glaces au glacier (le marchand de glaces). Livrée : ses gains sont
// multipliés ; ratée : divisés par deux pendant 5 min.
func (g *Game) parcOrder() *Order {
	type cmd struct {
		at    string
		lines []OrderLine
	}
	cmds := []cmd{{"ferme", []OrderLine{{It: Mais, N: 200}}}}
	if p := g.Place("exotique"); p != nil && p.Owned {
		cmds = append(cmds, cmd{"exotique", []OrderLine{{It: SucrePoudre, N: 50}}})
	}
	if p := g.Place("glacier"); p != nil && p.Owned {
		var ls []OrderLine
		for _, s := range p.Stations {
			if s.Kind == Shelf && !s.NoSale {
				ls = append(ls, OrderLine{It: s.Item, N: 10}) // (dix de chaque parfum)
			}
		}
		cmds = append(cmds, cmd{"glacier", ls})
	}
	if p := g.Place("creperie"); p != nil && p.Owned {
		cmds = append(cmds, cmd{"creperie", []OrderLine{{It: CrepeNature, N: 100}}}) // (le stand de crêpes)
	}
	c := cmds[g.rng.Intn(len(cmds))]
	return &Order{Who: "le parc d'attractions", At: c.at, Lines: c.lines, Left: schoolTime, Total: schoolTime, Fast: 180}
}

func (g *Game) finishOrder(at V) {
	o := g.Order
	g.Order = nil
	g.orderT = 180 + 120*g.rng.Float64()
	g.Stats.Orders++
	if o.Gift != "" {
		express := o.FastLeft() > 0
		gift := g.giveGift(o, express)
		txt := tr.F("Commande livrée : {qui} !", "qui", tr.T(o.Who)) + GiftSep + tr.F("Cadeau : {cadeau}", "cadeau", gift)
		if express {
			txt = tr.F("Livraison express : {qui} !", "qui", tr.T(o.Who)) + GiftSep + tr.F("Cadeau doublé : {cadeau}", "cadeau", gift)
		}
		g.emit(Event{Kind: EvLevel, Pos: at, Text: txt})
		return
	}
	gain, txt := o.Reward, "" // (ancienne commande d'une sauvegarde : payée en argent)
	if o.FastLeft() > 0 {
		gain *= 2
		txt = " " + tr.T("Livraison express : gain doublé !")
	}
	g.earnBonus(gain, at, o.Place())
	g.emit(Event{Kind: EvLevel, Pos: at, Text: tr.F("Commande livrée ({qui}) : +{prix} !", "qui", tr.T(o.Who), "prix", moneyText(gain)) + txt})
}

// schoolCrepes : la cantine commande des crêpes nature à la crêperie : une par élève
// du primaire, deux pour tous les autres (collège, lycée, université et adultes).
func (g *Game) schoolCrepes() *Order {
	sc := g.Place("ecole")
	n := 0
	for _, st := range sc.Students {
		if st.Room == nil {
			continue
		}
		if st.Room.Level == 0 {
			n++
		} else {
			n += 2
		}
	}
	for _, r := range sc.Rooms {
		if r.built && sc.HasProf(r) {
			n += 2 // les profs
		}
	}
	n += 2 * len(sc.Workers) // les surveillants
	if g.Levels[sc.key("secretaire")] > 0 {
		n += 2 // la secrétaire
	}
	if n == 0 {
		return nil
	}
	o := &Order{Who: "la cantine de l'école", At: "creperie", Left: schoolTime, Total: schoolTime, Fast: 120,
		Lines: []OrderLine{{It: CrepeNature, N: n}}}
	o.Reward = math.Round(Items[CrepeNature].Price*float64(n)*3/10) * 10
	return o
}
