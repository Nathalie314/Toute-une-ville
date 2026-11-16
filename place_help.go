package main

// placeHelp : la page « ? » d'un lieu : ce qu'il fait, ce qu'il reçoit, où partent
// ses produits, ses commandes, et un conseil (quand il y en a un). Un paragraphe par ligne.
func placeHelp(id string) []string {
	cmd := "Des habitants lui passent des commandes."
	cantine := "Des habitants et la cantine de l'école lui passent des commandes."
	switch id {
	case "ferme":
		return []string{
			"Lieu de départ. La ferme est le lieu de production de base : champs, vergers, animaux et moulins.",
			"Tous les produits de ses étals partent au supermarché.",
			"Ses trois laits (vache, chèvre, brebis) partent à la crèmerie.",
			"Ses deux farines (blé et maïs), ses œufs et son lait de vache partent à la boulangerie.",
			"Ses fraises, ses poires et ses œufs partent chez le glacier.",
			"Son foin, son maïs et ses pommes de terre partent à la ferme d'élevage.",
			"Son blé, ses pommes, ses poires et ses pommes de terre partent au domaine.",
			"Ses pommes, ses cerises, ses fraises, sa farine de blé, son lait et ses œufs partent à la pâtisserie.",
			"Des habitants et le parc d'attractions lui passent des commandes.",
			"Conseil : maximiser la production. Commencer par la vitesse des champs, puis la vitesse des animaux (pour les machines, attendre d'avoir débloqué la boulangerie). Augmenter ce que portent les ouvriers et leur vitesse. Dès le supermarché débloqué, améliorer les drones et les camions. Ne pas améliorer les gains liés à la ferme : cela ne sert à rien, ils sont minimes de toute façon.",
		}
	case "super":
		return []string{
			"L'immense supermarché vend les produits de tous tes lieux, espace par espace.",
			"Il reçoit en camion les produits de la ferme, de la crèmerie, de la ferme exotique, de la boulangerie, du glacier, du port, de la boucherie, du domaine et de la pâtisserie (1 objet livré = 3 en réserve).",
			"Conseil : penser à agrandir le parking et le nombre de caddies. Ne JAMAIS laisser une caisse sans caissière. Compter au départ un magasinier pour deux rayons.",
		}
	case "cremerie":
		return []string{
			"La crèmerie transforme les laits de la ferme en crème, beurre, fromages et yaourts.",
			"Elle reçoit les trois laits (vache, chèvre, brebis) de la ferme.",
			"Tous ses produits partent au supermarché.",
			"Son beurre part aussi à la boulangerie et à la pâtisserie, sa crème chez le glacier.",
			cantine,
			"Conseil : la crème et le beurre seront très demandés. Augmenter la vitesse de production des machines mais aussi leur rentabilité est nécessaire. Augmenter le temps pour réaliser les commandes, car les commandes de l'école sont énormes.",
		}
	case "exotique":
		return []string{
			"La ferme exotique fait pousser ananas, bananes, noix de coco, mangues, canne à sucre, cacao, café, riz, vanille, thé et cannelle, et en fait du sucre, du chocolat et du café.",
			"Tous ses produits partent au supermarché.",
			"Son sucre en poudre et son chocolat partent à la boulangerie.",
			"Ses mangues, ses ananas, ses noix de coco, sa vanille, son café moulu et son chocolat partent chez le glacier.",
			"Son riz, sa vanille, son chocolat et son café moulu partent à la pâtisserie.",
			"Des habitants et le parc d'attractions lui passent des commandes.",
			"Conseil : augmenter à la fois les champs et la vitesse des machines, car le sucre sera très demandé. Augmenter le temps pour réaliser les commandes, car les commandes du parc d'attractions sont énormes.",
		}
	case "boulangerie":
		return []string{
			"La boulangerie fait du pain et des viennoiseries : baguettes, pain de mie, pain blé-maïs, brioches, croissants et pains au chocolat.",
			"Elle reçoit les farines, les œufs et le lait de la ferme, le beurre de la crèmerie, le sucre et le chocolat de la ferme exotique.",
			"Tous ses produits partent au supermarché.",
			cantine,
			"Conseil : augmenter la vitesse des machines, mais aussi le temps pour réaliser les commandes, car les commandes de l'école sont énormes.",
		}
	case "glacier":
		return []string{
			"Le glacier fait des sorbets (fraise, mangue, ananas) et des glaces (vanille, café, noix de coco, chocolat, poire-chocolat).",
			"Il reçoit les fraises, les poires et les œufs de la ferme, la crème de la crèmerie, les fruits, la vanille, le café et le chocolat de la ferme exotique.",
			"Tous ses produits partent au supermarché.",
			"Des habitants, la cantine de l'école et le parc d'attractions lui passent des commandes.",
			"Conseil : les sorbets sont faciles à réaliser, mais les glaces demandent de la crème : il faut donc augmenter la crèmerie, ainsi que ses drones et ses camions, pour assurer l'approvisionnement. Augmenter le temps pour réaliser les commandes, car les commandes du parc d'attractions sont énormes.",
		}
	case "port":
		return []string{
			"Le port pêche des poissons (sardines, cabillaud, thon, saumon) avec ses bateaux, et ramasse des coquillages et des crustacés (huîtres, moules, saint-jacques, gambas) avec ses dragues.",
			"Tous ses produits partent au supermarché.",
			cmd,
			"Conseil : impossible d'aller pêcher soi-même, alors mieux vaut augmenter la capacité de pêche des bateaux et des dragues !",
		}
	case "elevage":
		return []string{
			"La ferme d'élevage élève volailles, bovins, ovins et porcs, puis les abat. Elle ne reçoit pas de clients.",
			"Elle reçoit le maïs, le foin et les pommes de terre de la ferme.",
			"Toute sa viande part en camion frigorifique à la boucherie.",
			"Chaque espace a ses éleveurs.",
		}
	case "boucherie":
		return []string{
			"La boucherie-charcuterie découpe la viande de la ferme d'élevage : poulet, canard, bœuf, veau, mouton, agneau, porc. Ses clients sont rares mais achètent cher.",
			"Elle reçoit les volailles et les carcasses de la ferme d'élevage.",
			"Tous ses produits partent au supermarché.",
			"Chaque espace a ses bouchers.",
		}
	case "domaine":
		return []string{
			"Le domaine cultive la vigne et le houblon, et fait des jus, des vins, de la bière, du cidre et des alcools.",
			"Il reçoit le blé, les pommes, les poires et les pommes de terre de la ferme.",
			"Son raisin et ses jus partent au supermarché ; ses vins, sa bière, son cidre et ses alcools se vendent sur place et partent aussi au supermarché.",
			"Chaque espace a ses ouvriers.",
			cmd,
			"L'alcool est dangereux pour la santé : à consommer avec modération.",
		}
	case "patisserie":
		return []string{
			"La pâtisserie fait des tartes aux pommes, du riz au lait, des clafoutis, des gâteaux au chocolat, des éclairs au café et des fraisiers.",
			"Elle reçoit les pommes, les cerises, les fraises, la farine de blé, le lait et les œufs de la ferme, le beurre de la crèmerie, le riz, la vanille, le chocolat et le café de la ferme exotique.",
			"Tous ses produits partent au supermarché.",
			cmd,
		}
	case "parc":
		return []string{
			"Le parc d'attractions ne consomme rien et ne livre rien : les visiteurs paient l'entrée (plus chère à chaque nouvelle attraction) et viennent pour les attractions, qui grandissent avec leur taille.",
			"Un employé (ou toi) fait monter et descendre les visiteurs ; l'attente fait baisser leur bonheur.",
			"Les stands (ballons, glaces, pop-corn…) n'ont besoin de personne : les visiteurs heureux qui passent devant y dépensent un peu.",
			"Il passe des commandes à la ferme (maïs), à la ferme exotique (sucre en poudre) et au glacier (glaces) : livrées, ses gains sont multipliés ; ratées, ils sont divisés par deux pendant 5 minutes.",
			"Au parc, tu roules en trottinette.",
		}
	}
	return []string{placeDesc(id)}
}
