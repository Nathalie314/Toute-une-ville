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
			"La ferme exotique fait pousser ananas, bananes, noix de coco, mangues, avocats, citrons, oranges, canne à sucre, cacao, café, riz, vanille, thé, cannelle et poivre, et en fait du sucre, du chocolat, du café, du thé et du poivre séché.",
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
			"Le second ponton, plus au sud, a huit bassins : bigorneaux, palourdes, crabes, tourteaux, araignées de mer, langoustines, homards et langoustes, chacun avec sa drague.",
			"Au nord de la criée, le marais salant : dans ses bassins, l'eau de mer s'évapore jusqu'à ce qu'ils soient tout blancs ; le râteau mécanique ramasse le sel. On achète les bassins depuis le bord.",
			"Tous ses produits partent au supermarché.",
			cmd,
			"Conseil : impossible d'aller pêcher soi-même, alors mieux vaut augmenter la capacité de pêche des bateaux et des dragues !",
		}
	case "elevage":
		return []string{
			"La ferme d'élevage élève volailles, bovins, ovins et porcs, puis les abat. Elle ne reçoit pas de clients.",
			"Elle reçoit le maïs, le foin et les pommes de terre de la ferme.",
			"Toute sa viande part en camion frigorifique à la boucherie.",
			"Conseil : améliorer les drones et les camions pour que les livraisons à la boucherie soient régulières.",
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
	case "creperie":
		return []string{
			"La crêperie est un restaurant : en haut la cuisine, en bas la salle (puis la terrasse, dehors). Toi seul passes de l'une à l'autre.",
			"Les clients arrivent en voiture et s'assoient seuls à une table propre ; sans table, ils attendent devant la porte, puis repartent fâchés.",
			"Chacun commande une crêpe salée, une crêpe sucrée et une salade ; une table de 2 boit une bouteille de cidre, une table de 4 deux. Il n'y a pas de caisse : la table paie ce qu'elle a reçu en partant.",
			"Un serveur (ou toi, arrêté près de la table) prend la commande : elle part en cuisine. Il apporte les crêpes du comptoir, le cidre et les salades de leurs frigos, puis nettoie la table.",
			"En cuisine, le chef des crêpes nature en fait sans arrêt. Les autres chefs ne cuisinent que ce qui est commandé, et attendent qu'un commis emporte leur plat au comptoir.",
			"Pas d'étoiles ici : une table contente laisse un pourboire à côté de la table (des pièces qui brillent), et toi seul peux le ramasser en passant.",
			"La crêperie n'est ouverte que quand tu joues : en ton absence, elle ne rapporte rien, et ses clients attablés t'attendent.",
			"Les ingrédients arrivent de la ferme, de la crèmerie, de la ferme exotique, du glacier, du port, de la boucherie et du domaine.",
			"Commandes : seulement des crêpes nature, pour la cantine de l'école (une par élève du primaire, deux pour les autres) et pour le stand de crêpes du parc d'attractions.",
		}
	case "patisserie":
		return []string{
			"La pâtisserie fait des tartes aux pommes, du riz au lait, des clafoutis, des gâteaux au chocolat, des éclairs au café, des tartes au citron meringuées et des fraisiers.",
			"Elle reçoit les pommes, les cerises, les fraises, la farine de blé, le lait et les œufs de la ferme, le beurre de la crèmerie, le riz, la vanille, le chocolat et le café de la ferme exotique.",
			"Tous ses produits partent au supermarché.",
			cmd,
		}
	case "commissariat":
		return []string{
			"Le commissariat protège la ville des voleurs. Il est offert avec le supermarché, mais rien n'y fonctionne sans le commissaire : il coûte 10 points d'université (sa case est dans son bureau).",
			"La première voiture de police est offerte avec le commissaire ; les suivantes s'achètent au parking (la 2e coûte le prix du 4e lieu de la carte, la 3e celui du 8e, la 4e celui du 12e…). Une nouvelle voiture arrive tous les quatre lieux de la rangée du haut.",
			"Chaque voiture a besoin de son lieutenant, formé avec des points du collège sur la case sous la voiture (10, puis 50, puis 250…). Un lieutenant sans patrouille erre sur le parking en sifflotant.",
			"La case « Patrouilles » du parking (ou un clic sur le parking) ouvre le menu des voitures : une voiture qui patrouille dans un lieu le protège, les voleurs n'y viennent plus. On l'y voit passer de temps en temps.",
			"Des mécaniciens (points du primaire : 100 fois les points de son lieutenant) permettent à une voiture de protéger un deuxième lieu.",
			"Commandes du commissariat, selon les rangs : à la boulangerie (brioche du commissaire, pains au chocolat des capitaines, croissants des lieutenants, baguettes des voleurs en garde à vue), à la pâtisserie (fraisier, tartes au citron meringuées, éclairs au café) et à la ferme exotique (un café moulu par policier). Livrée : les policiers coûtent 2 fois moins de points d'école ; ratée : 2 fois plus.",
			"À gauche du bâtiment, les bureaux des capitaines : chacun coûte le prix d'un lieu de la carte (le 3e, puis le 6e, le 7e, le 10e…). Le capitaine se forme sur la case de son bureau, avec des points du lycée (50, puis 500, puis 5 000…).",
			"La case « Vidéosurveillance » du hall ouvre le menu des capitaines : chacun surveille un lieu. Quand des voleurs y attaquent, il a une chance de les repérer (1 %, puis +1 % tous les 10 points du primaire placés dans son bureau) : ils volent quand même, mais sont arrêtés en partant.",
			"Les voleurs arrêtés vont en garde à vue (à droite du bâtiment ; 5 par cellule, 10 cellules au plus) pendant 30 minutes, même jeu fermé : pendant ce temps, ils n'attaquent plus. Sans place libre, ils restent libres.",
		}
	case "caserne":
		return []string{
			"La caserne s'ouvre (offerte) avec la boulangerie, en même temps que les incendies. Rien n'y fonctionne sans le commandant : 50 points d'université (sa case est dans son bureau). Comme le commissariat, elle ne tourne que pendant que tu joues.",
			"Un lieu qui a des zones chaudes (fours, plaques de cuisson) peut prendre feu : la boulangerie, la pâtisserie, la crêperie, la boucherie (rôtisserie, foie gras, jambons, pâtés) et le domaine (distilleries, alambic). Plus il en a, plus le risque est grand.",
			"Un incendie dure 1 minute : tout le monde fuit, toi aussi. À la fin, tout ce que les pompiers n'ont pas sauvé est détruit : machines, tables, stocks, drones, employés. Le lieu revient à son point de départ (il reste à toi, ses points d'école aussi).",
			"La prévention (à gauche) : des bureaux, leurs chargés de prévention (points du lycée). Chacun rapporte 1 point de prévention incendie (PPI) toutes les 2 heures, plus vite avec des points d'université placés sur la caserne. La case « Prévention » du hall répartit les PPI : chaque lieu part de 100 % de risque, −1 % tous les 10 PPI, 10 % au moins.",
			"L'action (à droite) : jusqu'à 5 camions, chacun avec son pompier (points du collège). Pendant un incendie, chaque camion sauve 1 élément (+1 tous les 1 000 points du primaire, case « Camions » du hall) : la caissière, puis les employés, puis les objets du plus cher au moins cher.",
		}
	case "prison":
		return []string{
			"La prison s'ouvre (offerte) dès qu'un premier groupe de voleurs est mis en garde à vue. Rien n'y fonctionne sans son directeur : 100 points d'université. Comme le commissariat, elle ne tourne que pendant que tu joues.",
			"Commandes de la prison (détenus, personnel, directeur) : à la boulangerie (baguettes, croissants, pain au chocolat), à la crèmerie (yaourts, fromages de vache, fromage de brebis) ; avec la cantine, à la boucherie (steaks hachés, rôtis de bœuf, foie gras) et à la ferme (pommes de terre, salades, fraises). Livrée : chaque besoin rempli retire 2 fois plus de points de récidive ; ratée : 2 fois moins.",
			"Un fourgon (au garage, sous la route ; 5 au plus) va chercher un voleur en garde à vue dès qu'une cellule est libre : 1 minute à l'aller, 1 minute au retour.",
			"Un prisonnier par cellule (20 au plus). Un surveillant (points du lycée) garde deux cellules : pas de cellule sans son surveillant.",
			"Un prisonnier reste 1 h 30 et arrive avec 100 points de récidive. Aux minutes 10, 20… 80, il essaie de remplir un de ses besoins, du plus rare au plus facile : la psy, l'assistant social, sa famille, le travail, le sport, se laver, manger, dormir (sa cellule).",
			"S'il y a une place, il y va 5 minutes et perd autant de points que le niveau de la pièce ; sinon, il en gagne 10. À sa sortie, un tirage de 1 à 100 : au-dessus de ses points, il est réinséré et la ville compte un voleur de moins pour toujours ; sinon, il retourne voler.",
			"La case « Améliorations » du grand couloir ouvre le menu des niveaux (10 au plus) : chaque cellule et chaque cabine (primaire), la cantine et la salle de sport (collège), chaque salle des familles et la salle de travail (lycée). La psy et l'assistant social (1 000 points d'université chacun) comptent au niveau 10.",
		}
	case "huilerie":
		return []string{
			"L'huilerie ne reçoit pas de clients et ne prend pas de commandes : tout part en camion.",
			"Six cultures, chacune dans son espace avec ses ouvriers : tournesol, colza et lin (des champs), oliviers, noisetiers et noyers (trois colonnes d'arbres chacun).",
			"Sous chaque culture, sa presse (le moulin pour les olives) fait l'huile ; les graines ne se vendent pas.",
			"Les six huiles partent au supermarché (espace « Huiles »), les olives au rayon Fruits, les noisettes et les noix à l'Épicerie. L'huile d'olive ira aussi à la future pizzeria.",
			"Rien n'est offert : même le premier champ s'achète.",
		}
	case "parc":
		return []string{
			"Le parc d'attractions ne consomme rien et ne livre rien : les visiteurs paient l'entrée (plus chère à chaque nouvelle attraction) et viennent pour les attractions, qui grandissent avec leur taille.",
			"Un employé (ou toi) fait monter et descendre les visiteurs ; l'attente fait baisser leur bonheur.",
			"Les stands (ballons, glaces, pop-corn…) n'ont besoin de personne : les visiteurs heureux qui passent devant y dépensent un peu.",
			"Il passe des commandes à la ferme (maïs), à la ferme exotique (sucre en poudre), au glacier (glaces) et à la crêperie (crêpes nature) : livrées, ses gains sont multipliés ; ratées, ils sont divisés par deux pendant 5 minutes.",
			"Au parc, tu roules en trottinette.",
			"Le parc n'est ouvert que quand tu joues : en ton absence, il ne rapporte rien, et ses visiteurs reprennent leur visite à ton retour.",
		}
	}
	return []string{placeDesc(id)}
}
