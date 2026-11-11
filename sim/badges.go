package sim

// Badges : de petits trophées, juste pour le plaisir (ils ne rapportent rien).

type Badge struct {
	ID, Place, Name, Desc string
	Icon                  int // indice d'icône (voir l'affichage)
	Check                 func(g *Game) bool
}

// icônes possibles
const (
	IconStar = iota
	IconWheat
	IconCorn
	IconVeg
	IconFlower
	IconCane
	IconMill
	IconBread
	IconEgg
	IconMilk
	IconHoney
	IconCake
	IconWorker
	IconCashier
	IconCoin
	IconHeart
	IconClock
	IconShirt
	IconMoon
	IconBag
	IconSign
	IconBio
	IconFarm
	IconCart
	IconCheese
	IconSausage
	IconFruit
	IconChoco
	IconIce
	IconTruck
)

func (g *Game) made(it Item) {
	if g.Stats.Made == nil {
		g.Stats.Made = map[string]int{}
	}
	g.Stats.Made[Items[it].Name]++
}

func (g *Game) farm() *Place { return g.Place("ferme") }
func (g *Game) sup() *Place  { return g.Place("super") }

// rayons : nombre de rayons ouverts.
func (p *Place) rayons() int {
	n := 0
	for _, s := range p.Stations {
		if s.Kind == Shelf && s.Built {
			n++
		}
	}
	return n
}

func fieldFull(id string) func(g *Game) bool {
	return func(g *Game) bool {
		s := g.farm().Station(id)
		return s.Built && len(s.Plants) >= len(s.Slots)
	}
}

func columnsBuilt(ids []string) func(g *Game) bool {
	return func(g *Game) bool {
		for _, id := range ids {
			if !g.farm().Station(id).Built {
				return false
			}
		}
		return true
	}
}

func allFull(ids []string) func(g *Game) bool {
	return func(g *Game) bool {
		for _, id := range ids {
			if !fieldFull(id)(g) {
				return false
			}
		}
		return true
	}
}

func stationMax(id string) func(g *Game) bool {
	return func(g *Game) bool {
		p := g.farm()
		s := p.Station(id)
		return s.Built && g.Levels[p.key(id+"_stock")] >= p.StockUpgrade(s).Max // (plus de niveaux : le stock)
	}
}

func lvl(key string, n int) func(g *Game) bool {
	return func(g *Game) bool { return g.Levels[key] >= n }
}

func built(id string) func(g *Game) bool {
	return func(g *Game) bool { return g.farm().Station(id).Built }
}

var Badges = []Badge{
	// --- la ferme
	{ID: "f_recolte", Place: "ferme", Name: "Première récolte", Desc: "Récolter quelque chose.", Icon: IconWheat,
		Check: func(g *Game) bool { return g.Stats.Harvested >= 1 }},
	{ID: "f_client", Place: "ferme", Name: "Premier client", Desc: "Encaisser un client.", Icon: IconCoin,
		Check: func(g *Game) bool { return g.farm().Served >= 1 }},
	{ID: "f_ble", Place: "ferme", Name: "Champ de blé complet", Desc: "Planter les 66 pousses de blé.", Icon: IconWheat, Check: fieldFull("ble")},
	{ID: "f_mais", Place: "ferme", Name: "Champ de maïs complet", Desc: "Planter les 30 pousses de maïs.", Icon: IconCorn, Check: fieldFull("mais")},
	{ID: "f_serre", Place: "ferme", Name: "Main verte", Desc: "Ouvrir les 6 colonnes de la serre.", Icon: IconVeg, Check: columnsBuilt(ColumnIDs[:6])},
	{ID: "f_serre_pleine", Place: "ferme", Name: "Serre pleine", Desc: "Acheter les 8 cases de chaque légume de la serre.", Icon: IconVeg, Check: allFull(ColumnIDs[:6])},
	{ID: "f_replante", Place: "ferme", Name: "Jardinier", Desc: "Ouvrir les six colonnes de la serre.", Icon: IconVeg,
		Check: func(g *Game) bool { return g.Stats.Replanted >= 100 || g.farm().columnsBuiltOf("serre") >= 6 }},
	{ID: "f_verger_plein", Place: "ferme", Name: "Verger plein", Desc: "Acheter les 8 cases de chaque arbre du verger.", Icon: IconFruit, Check: allFull(ColumnIDs[6:])},
	{ID: "f_fleurs", Place: "ferme", Name: "Mer de fleurs", Desc: "Planter les 9 pousses de fleurs.", Icon: IconFlower, Check: fieldFull("fleurs")},
	{ID: "f_caissiere", Place: "ferme", Name: "Patron", Desc: "Embaucher une caissière.", Icon: IconCashier, Check: lvl("ferme.caissiere", 1)},
	{ID: "f_ouvrier", Place: "ferme", Name: "Premier ouvrier", Desc: "Embaucher un ouvrier agricole.", Icon: IconWorker, Check: lvl("ferme.ouvriers", 1)},
	{ID: "f_equipe", Place: "ferme", Name: "Équipe au complet", Desc: "Embaucher tous les ouvriers de la ferme.", Icon: IconWorker, Check: lvl("ferme.ouvriers", len(HireOrder))},
	{ID: "f_moulin", Place: "ferme", Name: "Moulin à fond", Desc: "Agrandir au maximum le stock du moulin.", Icon: IconMill, Check: stationMax("moulin")},
	{ID: "f_moulin_mais", Place: "ferme", Name: "Farine dorée", Desc: "Agrandir au maximum le stock du moulin à maïs.", Icon: IconCorn, Check: stationMax("moulin_mais")},
	{ID: "f_chevre", Place: "ferme", Name: "Chevrier", Desc: "Agrandir au maximum le stock de la chèvrerie.", Icon: IconMilk, Check: stationMax("chevrerie")},
	{ID: "f_verger", Place: "ferme", Name: "Verger en fleurs", Desc: "Planter les 6 arbres du verger.", Icon: IconFruit, Check: columnsBuilt(ColumnIDs[6:])},
	{ID: "f_poules", Place: "ferme", Name: "Poules en forme", Desc: "Agrandir au maximum le stock du poulailler.", Icon: IconEgg, Check: stationMax("poulailler")},
	{ID: "f_etable", Place: "ferme", Name: "Vaches heureuses", Desc: "Agrandir au maximum le stock de l'étable.", Icon: IconMilk, Check: stationMax("etable")},
	{ID: "f_ruche", Place: "ferme", Name: "Reine des abeilles", Desc: "Agrandir au maximum le stock des ruches.", Icon: IconHoney, Check: stationMax("ruche")},
	{ID: "f_brebis", Place: "ferme", Name: "Berger", Desc: "Agrandir au maximum le stock de la bergerie.", Icon: IconMilk, Check: stationMax("bergerie")},
	{ID: "f_pancarte", Place: "ferme", Name: "Bien en vue", Desc: "50 points de collégiens placés à la ferme.", Icon: IconSign, Check: func(g *Game) bool {
		return g.Levels["ferme.pts_marketing"]+g.Levels["ferme.pts_panier"]+g.Levels["ferme.pts_fidele"]+g.Levels["ferme.pts_reclame"] >= 50
	}},
	{ID: "f_bio", Place: "ferme", Name: "100 % bio", Desc: "Label bio au niveau maximum.", Icon: IconBio, Check: lvl("ferme.bio", 6)},
	{ID: "f_tout", Place: "ferme", Name: "La ferme idéale", Desc: "Construire tous les bâtiments et champs de la ferme.", Icon: IconFarm,
		Check: func(g *Game) bool {
			for _, u := range g.farm().Unlocks {
				if !u.Done {
					return false
				}
			}
			return true
		}},
	{ID: "f_rep", Place: "ferme", Name: "Cinq étoiles", Desc: "Atteindre 5 étoiles de réputation à la ferme.", Icon: IconStar,
		Check: func(g *Game) bool { return g.farm().Rep >= 4.95 }},
	{ID: "f_100", Place: "ferme", Name: "Habitués", Desc: "Servir 100 clients.", Icon: IconHeart,
		Check: func(g *Game) bool { return g.farm().Served >= 100 }},
	{ID: "f_1000", Place: "ferme", Name: "Toute la région", Desc: "Servir 1 000 clients.", Icon: IconHeart,
		Check: func(g *Game) bool { return g.farm().Served >= 1000 }},
	{ID: "f_10000", Place: "ferme", Name: "Célèbre", Desc: "Servir 10 000 clients.", Icon: IconHeart,
		Check: func(g *Game) bool { return g.farm().Served >= 10000 }},

	// --- la boulangerie
	{ID: "b_ouvert", Place: "boulangerie", Name: "Bonjour, boulanger !", Desc: "Acheter la boulangerie.", Icon: IconBread,
		Check: func(g *Game) bool { return g.Place("boulangerie").Owned }},
	{ID: "b_baguettes", Place: "boulangerie", Name: "Cent baguettes", Desc: "Cuire 100 baguettes.", Icon: IconBread,
		Check: func(g *Game) bool { return g.Stats.Made["Baguette"] >= 100 }},
	{ID: "b_croissant", Place: "boulangerie", Name: "Pur beurre", Desc: "Faire un croissant.", Icon: IconBread,
		Check: func(g *Game) bool { return g.Stats.Made["Croissant"] >= 1 }},
	{ID: "b_recettes", Place: "boulangerie", Name: "Toutes les recettes", Desc: "Construire les 6 fours.", Icon: IconStar,
		Check: func(g *Game) bool { return allBuilt(g.Place("boulangerie")) }},
	{ID: "b_equipe", Place: "boulangerie", Name: "Brigade complète", Desc: "Embaucher les 6 boulangers.", Icon: IconWorker, Check: lvl("boulangerie.ouvriers", 6)},
	{ID: "b_1000", Place: "boulangerie", Name: "Le pain du quartier", Desc: "Servir 1 000 clients à la boulangerie.", Icon: IconHeart,
		Check: func(g *Game) bool { return g.Place("boulangerie").Served >= 1000 }},
	{ID: "b_rep", Place: "boulangerie", Name: "Meilleur ouvrier", Desc: "Atteindre 5 étoiles à la boulangerie.", Icon: IconStar,
		Check: func(g *Game) bool { p := g.Place("boulangerie"); return p.Owned && p.Rep >= 4.95 }},

	// --- la crèmerie
	{ID: "c_ouvert", Place: "cremerie", Name: "Crème fraîche", Desc: "Acheter la crèmerie.", Icon: IconMilk,
		Check: func(g *Game) bool { return g.Place("cremerie").Owned }},
	{ID: "c_beurre", Place: "cremerie", Name: "Comme du beurre", Desc: "Faire 100 plaquettes de beurre.", Icon: IconMilk,
		Check: func(g *Game) bool { return g.Stats.Made["Beurre"] >= 100 }},
	{ID: "c_fromages", Place: "cremerie", Name: "Plateau de fromages", Desc: "Faire un fromage de vache, de chèvre et de brebis.", Icon: IconCheese,
		Check: func(g *Game) bool {
			return g.Stats.Made["Fromage de vache"] >= 1 && g.Stats.Made["Fromage de chèvre"] >= 1 && g.Stats.Made["Fromage de brebis"] >= 1
		}},
	{ID: "c_recettes", Place: "cremerie", Name: "Toute la crèmerie", Desc: "Construire les 9 machines de la crèmerie.", Icon: IconStar,
		Check: func(g *Game) bool { return allBuilt(g.Place("cremerie")) }},
	{ID: "c_1000", Place: "cremerie", Name: "Bon appétit", Desc: "Servir 1 000 clients à la crèmerie.", Icon: IconHeart,
		Check: func(g *Game) bool { return g.Place("cremerie").Served >= 1000 }},
	{ID: "c_rep", Place: "cremerie", Name: "Crème de la crème", Desc: "Atteindre 5 étoiles à la crèmerie.", Icon: IconStar,
		Check: func(g *Game) bool { p := g.Place("cremerie"); return p.Owned && p.Rep >= 4.95 }},

	// --- le glacier
	{ID: "i_ouvert", Place: "glacier", Name: "Il fait chaud !", Desc: "Acheter le glacier.", Icon: IconIce,
		Check: func(g *Game) bool { return g.Place("glacier").Owned }},
	{ID: "i_vanille", Place: "glacier", Name: "Cent boules", Desc: "Faire 100 glaces nature.", Icon: IconIce,
		Check: func(g *Game) bool { return g.Stats.Made["Glace nature"]+g.Stats.Made["Glace vanille"] >= 100 }},
	{ID: "i_choco", Place: "glacier", Name: "Gourmandise", Desc: "Faire une glace au chocolat.", Icon: IconChoco,
		Check: func(g *Game) bool { return g.Stats.Made["Glace au chocolat"] >= 1 }},
	{ID: "i_recettes", Place: "glacier", Name: "Quatre parfums", Desc: "Construire les 4 machines du glacier.", Icon: IconStar,
		Check: func(g *Game) bool { return allBuilt(g.Place("glacier")) }},
	{ID: "i_1000", Place: "glacier", Name: "File d'attente", Desc: "Servir 1 000 clients au glacier.", Icon: IconHeart,
		Check: func(g *Game) bool { return g.Place("glacier").Served >= 1000 }},
	{ID: "i_rep", Place: "glacier", Name: "Glacier étoilé", Desc: "Atteindre 5 étoiles au glacier.", Icon: IconStar,
		Check: func(g *Game) bool { p := g.Place("glacier"); return p.Owned && p.Rep >= 4.95 }},

	// --- le supermarché
	{ID: "s_ouvert", Place: "super", Name: "Grande ouverture", Desc: "Acheter le supermarché.", Icon: IconCart,
		Check: func(g *Game) bool { return g.sup().Owned }},
	{ID: "s_camion", Place: "super", Name: "Première livraison", Desc: "Le camion apporte des produits de la ferme.", Icon: IconTruck,
		Check: func(g *Game) bool { return g.Stats.Delivered >= 1 }},
	{ID: "s_client", Place: "super", Name: "Premier caddie", Desc: "Encaisser un client au supermarché.", Icon: IconCoin,
		Check: func(g *Game) bool { return g.sup().Served >= 1 }},
	{ID: "s_5rayons", Place: "super", Name: "Bien achalandé", Desc: "Ouvrir 5 rayons.", Icon: IconCart,
		Check: func(g *Game) bool { return g.sup().rayons() >= 5 }},
	{ID: "s_rayons", Place: "super", Name: "Tout le catalogue", Desc: "Ouvrir tous les rayons.", Icon: IconCart,
		Check: func(g *Game) bool { return g.sup().rayons() >= len(superOrder) }},
	{ID: "s_caisses", Place: "super", Name: "Douze caisses", Desc: "Ouvrir les 12 caisses.", Icon: IconCashier,
		Check: func(g *Game) bool { return len(g.sup().Registers()) >= SuperRegs }},
	{ID: "s_caissieres", Place: "super", Name: "Pas d'attente", Desc: "Une caissière à chaque caisse.", Icon: IconCashier,
		Check: func(g *Game) bool {
			for i := 1; i <= SuperRegs; i++ {
				k := "super.caissiere"
				if i > 1 {
					k += itoa(i)
				}
				if g.Levels[k] == 0 {
					return false
				}
			}
			return true
		}},
	{ID: "s_equipe", Place: "super", Name: "Équipe de choc", Desc: "Embaucher 4 magasiniers dans un espace.", Icon: IconWorker,
		Check: func(g *Game) bool {
			for _, z := range SuperZones {
				if g.Levels["super."+teamKey(z.ID)] >= superStaff {
					return true
				}
			}
			return false
		}},
	{ID: "s_parking", Place: "super", Name: "Parking géant", Desc: "300 places de parking.", Icon: IconCart,
		Check: func(g *Game) bool { return g.sup().ParkTotal() >= SuperMaxPark }},
	{ID: "s_caddies", Place: "super", Name: "Armée de caddies", Desc: "300 caddies.", Icon: IconCart,
		Check: func(g *Game) bool { return g.sup().CartsTotal() >= SuperMaxPark }},
	{ID: "s_1000", Place: "super", Name: "Hypermarché", Desc: "Servir 1 000 clients au supermarché.", Icon: IconHeart,
		Check: func(g *Game) bool { return g.sup().Served >= 1000 }},
	{ID: "s_10000", Place: "super", Name: "Le rendez-vous de la ville", Desc: "Servir 10 000 clients au supermarché.", Icon: IconHeart,
		Check: func(g *Game) bool { return g.sup().Served >= 10000 }},
	{ID: "s_rep", Place: "super", Name: "Cinq étoiles en rayon", Desc: "Atteindre 5 étoiles au supermarché.", Icon: IconStar,
		Check: func(g *Game) bool { return g.sup().Owned && g.sup().Rep >= 4.95 }},
	{ID: "s_fidelite", Place: "super", Name: "Clients fidèles", Desc: "Carte de fidélité au niveau maximum.", Icon: IconBio, Check: lvl("super.bio", 6)},

	// --- Clairbourg (général)
	{ID: "g_camion", Place: "ville", Name: "Poids lourd", Desc: "100 points de mécaniciens dans un même lieu.", Icon: IconTruck, Check: func(g *Game) bool {
		for _, p := range g.Places {
			if g.Levels[p.key("pts_cam_vit")]+g.Levels[p.key("pts_cam_cap")] >= 100 {
				return true
			}
		}
		return false
	}},
	{ID: "g_100", Place: "ville", Name: "Tirelire", Desc: "Gagner 100 € au total.", Icon: IconCoin,
		Check: func(g *Game) bool { return g.Stats.Earned >= 100 }},
	{ID: "g_10k", Place: "ville", Name: "Bas de laine", Desc: "Gagner 10 000 € au total.", Icon: IconCoin,
		Check: func(g *Game) bool { return g.Stats.Earned >= 10000 }},
	{ID: "g_1m", Place: "ville", Name: "Millionnaire", Desc: "Gagner 1 000 000 € au total.", Icon: IconCoin,
		Check: func(g *Game) bool { return g.Stats.Earned >= 1e6 }},
	{ID: "g_tenue", Place: "ville", Name: "Nouveau look", Desc: "Acheter un équipement pour ton personnage.", Icon: IconShirt,
		Check: func(g *Game) bool {
			for _, gr := range Gears {
				if g.GearLevel(gr.ID) > 0 {
					return true
				}
			}
			return false
		}},
	{ID: "g_tenues", Place: "ville", Name: "Diplômé", Desc: "Porter l'ensemble du diplômé : chapeau, toge, livre et mocassins.", Icon: IconShirt,
		Check: func(g *Game) bool {
			for _, id := range Diplome {
				if !g.Worn(id) {
					return false
				}
			}
			return true
		}},
	{ID: "g_veilleur", Place: "ville", Name: "Nuits tranquilles", Desc: "Veilleur de nuit au niveau maximum.", Icon: IconMoon, Check: lvl("mairie.veilleur", 5)},
	{ID: "g_sac", Place: "ville", Name: "Gros bras", Desc: "Le sac sans fond.", Icon: IconBag, Check: lvl("perso.sac", 8)},
	// l'école
	{ID: "e_ecole", Place: "ecole", Name: "Rentrée des classes", Desc: "Ouvrir l'école (elle est offerte !).", Icon: IconStar, Check: func(g *Game) bool { return g.Place("ecole").Owned }},
	{ID: "e_premier", Place: "ecole", Name: "Premier diplômé", Desc: "Un élève quitte l'école.", Icon: IconStar, Check: lvl("ecole.diplomes", 1)},
	{ID: "e_cent", Place: "ecole", Name: "Cent élèves", Desc: "100 élèves formés.", Icon: IconWorker, Check: lvl("ecole.diplomes", 100)},
	{ID: "e_mille", Place: "ecole", Name: "Mille élèves", Desc: "1 000 élèves formés.", Icon: IconWorker, Check: lvl("ecole.diplomes", 1000)},
	{ID: "e_college", Place: "ecole", Name: "Au collège !", Desc: "Ouvrir une classe de collège.", Icon: IconSign, Check: roomBuilt(College)},
	{ID: "e_lycee", Place: "ecole", Name: "Au lycée !", Desc: "Ouvrir une classe de lycée.", Icon: IconSign, Check: roomBuilt(Lycee)},
	{ID: "e_univ", Place: "ecole", Name: "Université", Desc: "Ouvrir l'université.", Icon: IconStar, Check: roomBuilt(Universite)},
	{ID: "e_complet", Place: "ecole", Name: "Établissement complet", Desc: "Ouvrir les 13 classes.", Icon: IconStar, Check: func(g *Game) bool {
		for _, r := range g.Place("ecole").Rooms {
			if !r.built {
				return false
			}
		}
		return true
	}},
	{ID: "g_pub", Place: "ville", Name: "Star de la pub", Desc: "100 points de collégiens placés en tout.", Icon: IconSign, Check: func(g *Game) bool {
		n := 0
		for _, p := range g.Places {
			for _, k := range []string{"pts_marketing", "pts_patience", "pts_panier", "pts_fidele", "pts_reclame"} {
				n += g.Levels[p.key(k)]
			}
		}
		return n >= 100
	}},
	{ID: "g_1h", Place: "ville", Name: "Une heure à Clairbourg", Desc: "Jouer une heure.", Icon: IconClock,
		Check: func(g *Game) bool { return g.Stats.Played >= 3600 }},
	{ID: "g_10h", Place: "ville", Name: "Citoyen d'honneur", Desc: "Jouer dix heures.", Icon: IconClock,
		Check: func(g *Game) bool { return g.Stats.Played >= 36000 }},
	{ID: "g_heureux", Place: "ville", Name: "Sourires", Desc: "Rendre 500 clients contents.", Icon: IconHeart,
		Check: func(g *Game) bool { return g.Stats.Happy >= 500 }},
}

// checkBadges attribue les badges mérités (avec annonce si loud).
func (g *Game) checkBadges(loud bool) {
	if g.Badges == nil {
		g.Badges = map[string]bool{}
	}
	for _, b := range Badges {
		if !g.Badges[b.ID] && b.Check(g) {
			g.Badges[b.ID] = true
			if loud {
				g.emit(Event{Kind: EvBadge, Text: b.Name})
			}
		}
	}
}

// allBuilt : toutes les machines du lieu sont construites (et le lieu acheté).
func allBuilt(p *Place) bool {
	if !p.Owned {
		return false
	}
	for _, s := range p.Stations {
		if s.Kind == Machine && !s.Built {
			return false
		}
	}
	return true
}

func roomBuilt(level int) func(g *Game) bool {
	return func(g *Game) bool {
		for _, r := range g.Place("ecole").Rooms {
			if r.Level == level && r.built {
				return true
			}
		}
		return false
	}
}
