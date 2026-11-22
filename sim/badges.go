package sim

import "strings"

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
	{ID: "f_serre", Place: "ferme", Name: "Main verte", Desc: "Ouvrir les 7 colonnes de la serre.", Icon: IconVeg, Check: columnsBuilt(SerreIDs)},
	{ID: "f_serre_pleine", Place: "ferme", Name: "Serre pleine", Desc: "Acheter les 8 cases de chaque légume de la serre.", Icon: IconVeg, Check: allFull(SerreIDs)},
	{ID: "f_replante", Place: "ferme", Name: "Jardinier", Desc: "Ouvrir les six colonnes de la serre.", Icon: IconVeg,
		Check: func(g *Game) bool { return g.Stats.Replanted >= 100 || g.farm().columnsBuiltOf("serre") >= 6 }},
	{ID: "f_verger_plein", Place: "ferme", Name: "Verger plein", Desc: "Acheter les 8 cases de chaque arbre du verger.", Icon: IconFruit, Check: allFull(VergerIDs)},
	{ID: "f_fleurs", Place: "ferme", Name: "Mer de fleurs", Desc: "Planter les 9 pousses de fleurs.", Icon: IconFlower, Check: fieldFull("fleurs")},
	{ID: "f_caissiere", Place: "ferme", Name: "Patron", Desc: "Embaucher une caissière.", Icon: IconCashier, Check: lvl("ferme.caissiere", 1)},
	{ID: "f_ouvrier", Place: "ferme", Name: "Premier ouvrier", Desc: "Embaucher un ouvrier agricole.", Icon: IconWorker, Check: lvl("ferme.ouvriers", 1)},
	{ID: "f_equipe", Place: "ferme", Name: "Équipe au complet", Desc: "Embaucher tous les ouvriers de la ferme.", Icon: IconWorker, Check: lvl("ferme.ouvriers", len(HireOrder))},
	{ID: "f_moulin", Place: "ferme", Name: "Moulin à fond", Desc: "Agrandir au maximum le stock du moulin.", Icon: IconMill, Check: stationMax("moulin")},
	{ID: "f_moulin_mais", Place: "ferme", Name: "Farine dorée", Desc: "Agrandir au maximum le stock du moulin à maïs.", Icon: IconCorn, Check: stationMax("moulin_mais")},
	{ID: "f_chevre", Place: "ferme", Name: "Chevrier", Desc: "Agrandir au maximum le stock de la chèvrerie.", Icon: IconMilk, Check: stationMax("chevrerie")},
	{ID: "f_verger", Place: "ferme", Name: "Verger en fleurs", Desc: "Planter les 6 arbres du verger.", Icon: IconFruit, Check: columnsBuilt(VergerIDs)},
	{ID: "f_poules", Place: "ferme", Name: "Poules en forme", Desc: "Agrandir au maximum le stock du poulailler.", Icon: IconEgg, Check: stationMax("poulailler")},
	{ID: "f_etable", Place: "ferme", Name: "Vaches heureuses", Desc: "Agrandir au maximum le stock de l'étable.", Icon: IconMilk, Check: stationMax("etable")},
	{ID: "f_ruche", Place: "ferme", Name: "Reine des abeilles", Desc: "Agrandir au maximum le stock des ruches.", Icon: IconHoney, Check: stationMax("ruche")},
	{ID: "f_brebis", Place: "ferme", Name: "Berger", Desc: "Agrandir au maximum le stock de la bergerie.", Icon: IconMilk, Check: stationMax("bergerie")},
	{ID: "f_pancarte", Place: "ferme", Name: "Bien en vue", Desc: "50 points de collégiens placés à la ferme.", Icon: IconSign, Check: func(g *Game) bool {
		return g.Levels["ferme.pts_marketing"]+g.Levels["ferme.pts_panier"]+g.Levels["ferme.pts_fidele"]+g.Levels["ferme.pts_reclame"] >= 50
	}},
	{ID: "f_bio", Place: "ferme", Name: "Bons comptes", Desc: "Placer 70 points de comptable à la ferme.", Icon: IconBio, Check: lvl("ferme.pts_comptable", 70)},
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
	{ID: "b_recettes", Place: "boulangerie", Name: "Toutes les recettes", Desc: "Construire tous les fours.", Icon: IconStar,
		Check: func(g *Game) bool { return allBuilt(g.Place("boulangerie")) }},
	{ID: "b_equipe", Place: "boulangerie", Name: "Brigade complète", Desc: "Embaucher tous les boulangers.", Icon: IconWorker,
		Check: func(g *Game) bool { return staffFull(g.Place("boulangerie")) }},
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
	{ID: "i_vanille", Place: "glacier", Name: "Cent boules", Desc: "Faire 100 glaces à la vanille.", Icon: IconIce,
		Check: func(g *Game) bool {
			return g.Stats.Made["Glace à la vanille"]+g.Stats.Made["Glace nature"]+g.Stats.Made["Glace vanille"] >= 100
		}},
	{ID: "i_choco", Place: "glacier", Name: "Gourmandise", Desc: "Faire une glace au chocolat.", Icon: IconChoco,
		Check: func(g *Game) bool { return g.Stats.Made["Glace au chocolat"] >= 1 }},
	{ID: "i_recettes", Place: "glacier", Name: "Tous les parfums", Desc: "Construire toutes les fabriques du glacier.", Icon: IconStar,
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
	{ID: "s_fidelite", Place: "super", Name: "Clients fidèles", Desc: "Placer 100 points « Clients fidèles » au supermarché.", Icon: IconBio, Check: lvl("super.pts_fidele", 100)},

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
			for _, k := range []string{"pts_marketing", "pts_patience", "pts_panier", "pts_fidele", "pts_reclame", "pts_bonheur"} {
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

// ---------------------------------------------------------------- lieux récents
// ⚠️ Règle : chaque nouveau lieu reçoit ses badges (placeBadges + quelques badges
// propres à ce qu'il fabrique). La conceptrice n'y tient pas, mais des joueurs adorent.

// placeBadges : les badges de base d'un lieu (ouvert, premier client, tout
// débloqué, clients servis, 5 étoiles, équipe au complet).
func placeBadges(id, short, open, all, full, star string, clients int, icon int) []Badge {
	p := func(g *Game) *Place { return g.Place(id) }
	bs := []Badge{
		{ID: id + "_ouvert", Place: id, Name: open, Desc: "Acheter " + placeName(id) + ".", Icon: icon,
			Check: func(g *Game) bool { return p(g).Owned }},
		{ID: id + "_tout", Place: id, Name: all, Desc: "Tout débloquer : " + short + ".", Icon: IconStar,
			Check: func(g *Game) bool { return allUnlocked(p(g)) }},
		{ID: id + "_equipe", Place: id, Name: full, Desc: "Embaucher tous les employés : " + short + ".", Icon: IconWorker,
			Check: func(g *Game) bool { return staffFull(p(g)) }},
	}
	if clients > 0 {
		bs = append(bs,
			Badge{ID: id + "_client", Place: id, Name: "Premier client", Desc: "Encaisser un client : " + short + ".", Icon: IconCoin,
				Check: func(g *Game) bool { return p(g).Served >= 1 }},
			Badge{ID: id + "_clients", Place: id, Name: "Bonne adresse", Desc: "Servir " + groupThousands(clients) + " clients : " + short + ".", Icon: IconHeart,
				Check: func(g *Game) bool { return p(g).Served >= clients }},
			Badge{ID: id + "_rep", Place: id, Name: star, Desc: "Atteindre 5 étoiles : " + short + ".", Icon: IconStar,
				Check: func(g *Game) bool { q := p(g); return q.Owned && q.Rep >= 4.95 }})
	}
	return bs
}

// madeBadge : fabriquer n fois un produit.
func madeBadge(id, place, name, desc string, it Item, n, icon int) Badge {
	return Badge{ID: id, Place: place, Name: name, Desc: desc, Icon: icon,
		Check: func(g *Game) bool { return g.Stats.Made[Items[it].Name] >= n }}
}

// allUnlocked : le lieu est acheté et tous ses déblocages sont faits.
func allUnlocked(p *Place) bool {
	if p == nil || !p.Owned {
		return false
	}
	for _, u := range p.Unlocks {
		if !u.Done {
			return false
		}
	}
	return true
}

// staffFull : tous les employés du lieu sont embauchés (par espace s'il en a).
func staffFull(p *Place) bool {
	if p == nil || !p.Owned {
		return false
	}
	g := p.g
	for _, u := range p.PlaceUpgrades() {
		if p.Parc != nil && (strings.HasPrefix(u.Key, p.key("caissiere")) || strings.HasPrefix(u.Key, p.key("operateur_"))) && g.Levels[u.Key] < u.Max {
			return false // (au parc : toutes les caissières et tous les employés des attractions)
		}
		if u.Key == p.key("ouvriers") || (teamZonesOf(p.ID) != nil && strings.HasPrefix(u.Key, p.key("ouvriers_"))) {
			if g.Levels[u.Key] < u.Max {
				return false
			}
		}
	}
	return true
}

func groupThousands(n int) string {
	s := itoa(n)
	if len(s) > 3 {
		return s[:len(s)-3] + " " + s[len(s)-3:]
	}
	return s
}

func init() {
	B := func(bs ...Badge) { Badges = append(Badges, bs...) }
	// la ferme exotique
	B(placeBadges("exotique", "ferme exotique", "Sous les tropiques", "Paradis tropical", "Toute la jungle", "Perle des tropiques", 1000, IconFruit)...)
	B(madeBadge("exotique_sucre", "exotique", "Doux comme le sucre", "Faire 100 sachets de sucre en poudre.", SucrePoudre, 100, IconCane),
		madeBadge("exotique_choco", "exotique", "Chocolatier", "Faire une tablette de chocolat.", Chocolat, 1, IconChoco),
		madeBadge("exotique_cafe", "exotique", "Petit noir", "Faire 100 paquets de café moulu.", CafeMoulu, 100, IconStar),
		madeBadge("exotique_poivre", "exotique", "Poivre et piment", "Faire 200 sachets de poivre.", Poivre, 200, IconStar))
	// la ferme d'élevage (pas de clients)
	B(placeBadges("elevage", "ferme d'élevage", "Les pieds dans le foin", "Tout le troupeau", "Éleveurs au complet", "", 0, IconFarm)...)
	B(madeBadge("elevage_poulets", "elevage", "Basse-cour", "Élever 100 poulets.", Poulet, 100, IconEgg),
		madeBadge("elevage_vaches", "elevage", "Vaches à lait", "Élever 100 vaches.", Vache, 100, IconMilk))
	// la boucherie-charcuterie (peu de clients, mais chers)
	B(placeBadges("boucherie", "boucherie", "Chez le boucher", "Toute la charcuterie", "Bouchers au complet", "Boucherie étoilée", 300, IconSausage)...)
	B(madeBadge("boucherie_steaks", "boucherie", "Bien saignant", "Faire 100 steaks.", Steak, 100, IconSausage),
		madeBadge("boucherie_foie", "boucherie", "Repas de fête", "Faire un foie gras.", FoieGras, 1, IconStar))
	// le port
	B(placeBadges("port", "port", "Moussaillon", "Toute la criée", "Équipage au complet", "Port étoilé", 1000, IconTruck)...)
	B(madeBadge("port_sardines", "port", "Serrés comme des sardines", "Pêcher 100 sardines.", Sardine, 100, IconStar),
		madeBadge("port_huitres", "port", "Perle rare", "Récolter 100 huîtres.", Huitre, 100, IconStar),
		madeBadge("port_sel", "port", "Fleur de sel", "Récolter 500 sel au marais salant.", Sel, 500, IconStar),
		madeBadge("port_crabes", "port", "Pince-mi, pince-moi", "Récolter 200 crabes.", Crabe, 200, IconStar),
		madeBadge("port_homards", "port", "Homard bleu", "Récolter 100 homards.", Homard, 100, IconStar))
	// le domaine
	B(placeBadges("domaine", "domaine", "Bienvenue au domaine", "Grand cru", "Domaine au complet", "Domaine étoilé", 1000, IconFruit)...)
	B(madeBadge("domaine_vin", "domaine", "Vendanges", "Faire 100 bouteilles de vin rouge.", VinRouge, 100, IconFruit),
		madeBadge("domaine_cidre", "domaine", "Cidre bouché", "Faire 100 bouteilles de cidre.", Cidre, 100, IconFruit))
	// la crêperie (pas d'étoiles : pas de badge de réputation)
	crep := func(g *Game) *Place { return g.Place("creperie") }
	B(placeBadges("creperie", "crêperie", "À table !", "Toute la carte", "Brigade au complet", "", 0, IconCake)...)
	B(Badge{ID: "creperie_client", Place: "creperie", Name: "Premier couvert", Desc: "Servir une première table : crêperie.", Icon: IconCoin,
		Check: func(g *Game) bool { return crep(g).Served >= 1 }},
		Badge{ID: "creperie_clients", Place: "creperie", Name: "Bonne table", Desc: "Servir 1 000 clients : crêperie.", Icon: IconHeart,
			Check: func(g *Game) bool { return crep(g).Served >= 1000 }},
		Badge{ID: "creperie_pourboire", Place: "creperie", Name: "Merci pour le service", Desc: "Ramasser 100 000 € de pourboires.", Icon: IconCoin,
			Check: func(g *Game) bool { return g.Stats.Tips >= 100000 }},
		madeBadge("creperie_sourire", "creperie", "Crêpe qui sourit", "Faire 100 crêpes saucisse-œufs.", CrepeSaucisseOeuf, 100, IconEgg),
		madeBadge("creperie_flambee", "creperie", "Flambée !", "Faire une crêpe au canard flambé au whisky.", CrepeCanard, 1, IconStar))
	// la pâtisserie
	B(placeBadges("patisserie", "pâtisserie", "Bec sucré", "Toute la vitrine", "Pâtissiers au complet", "Pâtisserie étoilée", 1000, IconCake)...)
	B(madeBadge("patisserie_tartes", "patisserie", "Tarte aux pommes", "Faire 100 tartes aux pommes.", TartePommes, 100, IconCake),
		madeBadge("patisserie_fraisier", "patisserie", "Cerise sur le gâteau", "Faire un fraisier.", Fraisier, 1, IconCake))
	// l'huilerie
	B(placeBadges("huilerie", "huilerie", "Tout baigne", "Toutes les huiles", "Huiliers au complet", "", 0, IconFruit)...)
	B(madeBadge("huilerie_olive", "huilerie", "Huile vierge", "Faire 100 bouteilles d'huile d'olive.", HuileOlive, 100, IconFruit),
		madeBadge("huilerie_tournesol", "huilerie", "Plein soleil", "Faire 500 bouteilles d'huile de tournesol.", HuileTournesol, 500, IconFlower),
		Badge{ID: "huilerie_vergers", Place: "huilerie", Name: "Bosquet", Desc: "Planter les neuf colonnes d'arbres de l'huilerie.", Icon: IconFruit,
			Check: func(g *Game) bool {
				q := g.Place("huilerie")
				for _, c := range HuilCultures {
					for _, id := range []string{c.ID, c.ID + "_2", c.ID + "_3"} {
						if s := q.Station(id); c.Tree && (s == nil || !s.Built) {
							return false
						}
					}
				}
				return q.Owned
			}})
	// le commissariat (comme placeBadges : ouvert, tout acheté, équipe au complet ; puis les patrouilles)
	polAll := func(g *Game) bool {
		for k := 1; k <= PolCars(); k++ {
			if !g.PolCarOwned(k) {
				return false
			}
		}
		return g.HasChief()
	}
	B(Badge{ID: "commissariat_ouvert", Place: "commissariat", Name: "Sous bonne garde", Desc: "Ouvrir le commissariat (il est offert avec le supermarché).", Icon: IconStar,
		Check: func(g *Game) bool { return g.Place("commissariat").Owned }},
		Badge{ID: "commissariat_chef", Place: "commissariat", Name: "Monsieur le commissaire", Desc: "Faire venir le commissaire.", Icon: IconWorker,
			Check: func(g *Game) bool { return g.HasChief() }},
		Badge{ID: "commissariat_tout", Place: "commissariat", Name: "Toute la flotte", Desc: "Acheter toutes les voitures de police.", Icon: IconTruck, Check: polAll},
		Badge{ID: "commissariat_equipe", Place: "commissariat", Name: "Toute la brigade", Desc: "Former tous les lieutenants.", Icon: IconWorker,
			Check: func(g *Game) bool { return g.PolLieutenants() >= PolCars() }},
		Badge{ID: "commissariat_patrouille", Place: "commissariat", Name: "En patrouille", Desc: "Envoyer une voiture de police en patrouille.", Icon: IconTruck,
			Check: func(g *Game) bool {
				for k := 1; k <= PolCars(); k++ {
					if g.PolOnPatrol(k) {
						return true
					}
				}
				return false
			}},
		Badge{ID: "commissariat_double", Place: "commissariat", Name: "Deux lieux à la fois", Desc: "Une voiture protège deux lieux (mécaniciens).", Icon: IconStar,
			Check: func(g *Game) bool {
				for k := 1; k <= PolCars(); k++ {
					if g.PolDriven(k) && g.PolPatrol(k, 0) != "" && g.PolPatrol(k, 1) != "" {
						return true
					}
				}
				return false
			}})
	// le parc d'attractions
	parc := func(g *Game) *Place { return g.Place("parc") }
	B(placeBadges("parc", "parc d'attractions", "Ticket d'entrée", "Tout le parc", "Équipe au complet", "Parc étoilé", 1000, IconHeart)...)
	B(Badge{ID: "parc_frisson", Place: "parc", Name: "Grand frisson", Desc: "Ouvrir le grand huit.", Icon: IconStar,
		Check: func(g *Game) bool { st := parc(g).Station("attr_grandhuit"); return st != nil && st.Built }},
		Badge{ID: "parc_stands", Place: "parc", Name: "Barbe à papa et pop-corn", Desc: "Vendre 1 000 fois aux stands du parc.", Icon: IconCoin,
			Check: func(g *Game) bool { return parc(g).Parc.StandSales >= 1000 }},
		Badge{ID: "parc_geant", Place: "parc", Name: "Toujours plus grand", Desc: "Mettre une attraction à la taille 10.", Icon: IconStar,
			Check: func(g *Game) bool {
				q := parc(g)
				for _, r := range q.Parc.Rides {
					if q.RideBuilt(r) && q.RideLevel(r) >= 10 {
						return true
					}
				}
				return false
			}})
}
