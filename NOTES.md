# Toute une Ville — notes de développement

Jeu de gestion « idle arcade » pour PC (Windows, plein écran), en français.
On construit la ville de **Clairbourg** lieu après lieu. Version 1.0 : la ferme,
la carte de la ville, la mairie et la boutique de vêtements.

## Principe
- On dirige un personnage (ZQSD / flèches / clic maintenu). Molette : zoom. F11 : plein écran.
- Champs : on marche dedans pour récolter. Les objets s'empilent dans les bras.
- Machines (moulin, four, pâtisserie) : dépôt à gauche (bleu), sortie à droite.
- Animaux (poulailler, étable, ruches) : production automatique, on ramasse devant.
- Étals : on dépose sur la zone claire au-dessus ; les clients se servent puis paient à la caisse
  (il faut quelqu'un à la caisse : le joueur ou la caissière).
- Zones en pointillés jaunes : on reste dessus, l'argent s'y verse, le bâtiment apparaît.
- E : améliorations du lieu. Échap : carte de la ville.

## Version 1.1 — début plus rapide
- Chaque champ commence avec **une seule pousse**. Les suivantes s'achètent une par une sur la case « +1 »
  (prix = prix de vente × 1,18^(n-1) : 3 €, 4 €, 4 €, 5 €… ; champ de blé complet ≈ 150 €, maïs ≈ 260 €).
- On ne paie une zone (pousse ou bâtiment) que si l'on **s'arrête** dessus : traverser ne coûte rien.
- Caisse déplacée à gauche, près du blé ; clients plus rapides, arrivant plus près ; premier client tout de suite.
- Un client qui a déjà un peu d'un produit n'attend plus longtemps le reste.
- Blé : 5 s pour pousser.
- Rythme mesuré : 1re pousse ≈ 30 s, maïs ≈ 3 min, blé complet ≈ 5 min, moulin ≈ 7 min, caissière ≈ 8 min.

## Version 1.2
- Bouton « Améliorations » en surbrillance (halo + « ! ») quand un achat est possible ; idem sur la carte.
- Bouton « Recommencer à zéro » sur la carte (avec confirmation).
- Les niveaux des champs et bâtiments s'achètent sur une case « Niveau N » au sol, à côté de chacun
  (à droite des bâtiments, sous le coin droit des champs). Le panneau E ne garde que personnel et ventes.

## Version 1.3 — tout au sol
- Plus de bouton « Améliorations » ni de touche E : tout s'achète sur des cases au sol (on s'arrête dessus).
- Caissière : case à gauche de la caisse. Ouvriers, bottes, brouettes : cases devant le hangar (en bas à droite de la boutique ;
  bottes et brouettes apparaissent après le 1er ouvrier). Pancarte et label bio : au bord de la route, de part et d'autre du portail.
- Poubelle remontée un peu ; les ouvriers inoccupés attendent près du hangar.

## Version 1.4 — réputation et badges
- Réputation par lieu (1 à 5 étoiles, départ 3) affichée sous l'argent et sur la carte.
  Client content +0,012 ; impatient +0,002 ; servi incomplètement −0,02 ; reparti sans rien −0,12 ; abandon à la caisse −0,15.
  Fréquentation × (0,5 + réputation / 6) : de ×0,67 (1 étoile) à ×1,33 (5 étoiles).
- 36 badges (25 pour la ferme, 11 pour Clairbourg), purement décoratifs : bouton « Badges » (touche B) en jeu et sur la carte.
  Annonce + carillon quand on en gagne un. Les badges déjà mérités dans une ancienne partie sont attribués sans fanfare.

## Version 2.0 — le supermarché
- S'achète sur la carte (15 000 €). Vend les produits de la ferme 2 × plus cher.
- À la ferme : « Quai du camion » (en bas à gauche, au bord de la route). On y dépose les produits
  (10 max par produit, seulement ceux qui ont un rayon au supermarché). Les ouvriers de la ferme l'alimentent aussi.
- Camion : toutes les 25 s, 20 produits (mairie « Camion de livraison » : jusqu'à 7 s et 80 produits).
- Supermarché : réserve en haut (une palette par produit), 11 rayons (blé offert, les autres de 450 € à 16 500 €,
  seulement si le produit est débloqué à la ferme), 3 caisses (2e : 3 000 €, 3e : 9 000 €).
- Cases au sol : caissière par caisse (600 / 1 500 / 4 000 €), employés de rayon (6), rollers, chariots,
  enseigne lumineuse (clients), carte de fidélité (prix).
- Les clients du supermarché ne viennent chercher que ce qui est en rayon (articles réservés) ;
  réputation plus lente à monter (+0,004 par client content).
- 12 badges supermarché + « Poids lourd » (camion au maximum).
- Mesure (joueur automatique) : achat vers 3 h 20 ; revenus 17 → 37 €/s dans les 40 minutes suivantes.
- Plusieurs caisses par lieu (chaque caisse a sa file et sa caissière).

## Version 2.1
- Les PNJ calculent leur chemin (A* sur une grille de 30 px) et contournent les bâtiments.
- « Stock + » : case sous la case « Niveau » de chaque bâtiment (moulin, four, poulailler, étable, ruches, pâtisserie) :
  +5 en entrée et +6 en sortie par niveau (5 niveaux).
- « Étals + » (ferme, sous la caissière) et « Rayons + » (supermarché, sous les employés) : +4 produits par étal/rayon et par niveau (6 niveaux).

## Version 2.2 — fromage
- Ferme élargie (3 150 de large au lieu de 2 700) : 7ᵉ bâtiment « Fromagerie » (lait → fromage, 5 s), 9 000 €,
  débloquée après l'étable. Fromage vendu 45 €. 12ᵉ étal, rayon Fromage au supermarché (6 750 €), badge « Maître fromager ».
- À prévoir : rayon fromage au futur restaurant.
- Le quai passe après les étals de la ferme (priorité ×0,6) pour ne pas vider la boutique de la ferme.
- Mairie : le camion est verrouillé tant que le supermarché n'est pas acheté.
- Les clients passent en caisse dans leur ordre d'arrivée à la caisse (ticket).

## Version 2.3 — porcherie
- Ferme élargie à 3 600 : 8ᵉ bâtiment « Porcherie » (légumes → saucisses, 5,5 s), 15 000 €, après le champ de fleurs, avant les ruches.
  Saucisses vendues 50 €. 13ᵉ étal ; rayon Saucisses au supermarché (7 500 €) ; 13ᵉ palette ; badge « Charcutier ».
- Supermarché réorganisé : 7 rayons en haut, 6 en bas.

## Version 2.4
- Champ de blé double (10 colonnes, 30 pousses) ; les autres champs décalés vers la droite.
- Ordre des bâtiments : moulin, four, poulailler, étable, fromagerie, porcherie, ruches, pâtisserie.
- Étal du fromage juste après le lait. Ferme : jusqu'à 10 ouvriers (7ᵉ à 10ᵉ : 220 k, 450 k, 900 k, 1,8 M€).

## Ferme : déblocages (dans l'ordre, 2 visibles à la fois)
| Lieu | Prix | Produit (prix de vente) |
|---|---|---|
| Champ de blé | départ | Blé 3 € |
| Champ de maïs | 25 € | Maïs 5 € |
| Moulin | 100 € | Farine 10 € (1 blé) |
| Four à pain | 300 € | Pain 22 € (1 farine) |
| Poulailler | 900 € | Œufs 12 € |
| Potager | 2 800 € | Légumes 15 € |
| Étable | 6 000 € | Lait 18 € |
| Fromagerie | 9 000 € | Fromage 45 € (1 lait) |
| Champ de fleurs | 11 000 € | Fleurs 24 € (futur fleuriste) |
| Porcherie | 15 000 € | Saucisses 50 € (1 légume) |
| Ruches | 22 000 € | Miel 40 € |
| Pâtisserie | 50 000 € | Gâteau 110 € (farine + œufs + lait) |
| Canne à sucre | 400 000 € | Canne 45 € — volontairement hors de portée de la ferme seule |

## Améliorations
- Ferme : caissière (150 €), 6 ouvriers, bottes (vitesse), brouettes (capacité), pancarte (clients), label bio (prix),
  niveau de chaque champ / bâtiment (+35 % de vitesse par niveau, 9 niveaux) — case au sol à côté.
- Mairie (toute la ville) : veilleur de nuit (gains hors-ligne 1 h → 24 h), sac à dos (+2 objets / niveau), camion, points de l'école.
- Vêtements (joueur seulement) : survêtement, salopette à poches, tablier de chef, rollers, costume chic, super-héros.

## Hors-ligne
Gains = revenu moyen des 5 dernières minutes × durée (plafonnée par le veilleur de nuit)
× (4 % + 1 % par ouvrier, max 8 % ; avant la 4.3 : 15 % + 3 %, max 30 %). Rien sans caissière.

## Version 3.0 — la grande réorganisation
- Ferme (5450 de large) : chaque champ au-dessus des bâtiments qu'il nourrit.
  maïs → moulin à maïs, poulailler ; blé (invendable) → moulin à blé, étable → fromagerie, crèmerie ;
  potager → porcherie ; verger (fruits vendus tels quels : la « respiration » du milieu de partie) ;
  fleurs (invendables) → ruches ; canne (invendable) → sucrerie ; cacao (invendable) + sucre → chocolaterie.
- Ordre / prix : moulin à maïs 60, poulailler 200, blé 400, moulin 700, étable 1800 (améliorations ×0,3),
  fromagerie 3000, crèmerie 4500, potager 6000, porcherie 9000, verger 12000, fleurs 16000, ruches 22000,
  canne 32000, sucrerie 45000, cacao 70000, chocolaterie 100000.
- Trois quais à la ferme : → supermarché, → boulangerie, → glacier (un camion chacun).
- Boulangerie-pâtisserie (20 000 €) : baguette, baguette de maïs, pain blé-maïs, brioches, pain d'épices,
  gâteau aux fruits, tartes, gâteau au chocolat.
- Glacier (40 000 €) : glace vanille, sorbet aux fruits, glace au chocolat.
- Boulangerie et glacier ont leur propre quai → supermarché (25 rayons, prix ×2).
- Sauvegarde version 3 : l'ancienne partie est convertie, mais une remise à zéro est conseillée.

## Version 3.1 — un ouvrier par zone
- Les ouvriers de la ferme appartiennent à une zone (maïs ×2, blé ×2, potager, verger, fleurs, canne, cacao : 9 en tout).
  Ils ne servent que les bâtiments et étals de leur zone, se reposent sous leur champ, et ne remplissent pas les quais.
- Ordre conseillé par les prix : maïs, blé, maïs, blé, potager, verger, fleurs, canne, cacao
  (250, 1000, 3500, 8000, 9000, 14000, 22000, 40000, 85000 €). Case à gauche du « Niveau » du champ.
- Bottes et brouettes restent devant le hangar. Sauvegarde version 4 : les anciens ouvriers sont répartis dans l'ordre.
- Supermarché : 1 objet livré = 10 objets en réserve. Étable : 2 laits par ration de blé.

## Version 3.2 — les drones
- Les ouvriers ne chargent plus les camions. Un drone par étal de la ferme (base « H » sous l'étal ; même prix pour tous : 800 € à la ferme, 2 500 € à la boulangerie, 4 000 € au glacier),
  achetable dès qu'un lieu reçoit les camions. Il prend les produits de SON étal et les porte au quai qui en manque le plus.
- Au départ : 1 objet, aller-retour 25 s (= intervalle du camion) → un seul objet par livraison.
- Hangar (après le premier drone) : Soute (jusqu'à 10 objets, 9 niveaux) et Hélices (6 niveaux, aller-retour 25 s → 9 s).
- Le joueur peut toujours déposer lui-même sur les quais. Boulangerie et glacier aussi : un drone par vitrine,
  soute et hélices à côté des employés (prix ×2 à la boulangerie, ×3 au glacier). Aucun employé ne remplit de quai.

## Version 4.0 — l'école
- L'école (5 000 €, sur la carte avant le supermarché). Les élèves arrivent au portail ; on les inscrit à l'accueil
  (le joueur ou la secrétaire, 1 200 €). 13 salles : 6 primaire, 4 collège, 2 lycée, 1 université.
  Chaque salle : 1 à 6 tables (case « Table » à l'endroit de la prochaine), un prof (case sur son bureau).
  Une classe avance si elle a un prof, ou si le joueur s'y tient.
- Formation : primaire 10 min, collège 20, lycée 40, université 80 (÷ 1 + points de formation / 100).
  Fin : l'élève monte au niveau supérieur (table libre + prof), sinon il part et rapporte 1 point :
  primaire → ouvrier, collège → marketing, lycée → comptable, université → formation.
- Points répartis à la mairie (bouton « Points de l'école »), définitifs après « Valider ». Pas de plafond.
  Ouvriers/employés : +1 % vitesse, +1 objet porté / 20 points ; mécaniciens (drones) : pareil pour les drones ;
  marketing : +1 % de clients ; comptables : +1 % de gains ; formation (école) : +1 % de vitesse de formation.
- Pancarte et label bio (et équivalents) supprimés ; sauvegarde v5 : niveaux convertis (20 marketing / 12 comptable par niveau).
- Absence : les classes avec prof continuent à moitié vitesse (points crédités au retour).
- Prix : salles primaire 0 / 1 500 / 4 000 / 9 000 / 20 000 / 45 000 ; collège 6 000 → 80 000 ; lycée 30 000 / 90 000 ;
  université 200 000. Tables 300 / 2 000 / 8 000 / 30 000 (×1,5 par table), profs 800 / 6 000 / 25 000 / 80 000.

## Version 4.1 — champs plus lents
- Durée de pousse : la 1re pousse d'un champ pousse à la vitesse de base ; à partir de 11 pousses, 6 fois plus lentement
  (progression linéaire entre les deux). Mesure (joueur automatique) : avant, 90-97 % des pousses étaient mûres en
  permanence ; maintenant 17 % à 15 min, 40-60 % dans la première heure.
- Potager, verger, fleurs, canne, cacao : 9 pousses (au lieu de 15).
- Niveau d'un champ : à partir du niveau 5, chaque pousse donne 2 récoltes (3 au niveau 9).
- Embauche libre : la case d'ouvrier apparaît dès que le champ de la zone existe ; seuls les prix guident
  (maïs 250 / 3 500, blé 1 000 / 8 000, potager 9 000, verger 14 000, fleurs 22 000, canne 40 000, cacao 85 000).
- Mairie : panneaux publicitaires supprimés (doublon avec le marketing) ; sauvegarde v6 : chaque niveau devient
  15 points de marketing dans chaque lieu.
- Correctif : un ouvrier (ou le joueur automatique) qui cherchait le compost partait vers le hangar (décor).

## Version 4.2 — le personnage
- La tuile « Vêtements » devient « Personnage » ; carte : Mairie et Personnage à gauche (séparés), lieux à droite.
- Équipements cumulables (clés perso.*) :
  pieds (chaussures, 8 niv., +10 % → +110 % de vitesse, 500 → 300 000 €) ;
  dos (sac, 8 niv., +2 → +20 objets, 250 → 260 000 €, remplace le sac à dos de la mairie) ;
  tête (toque, 6 niv., bâtiments proches +25 %/niv., 2 000 → 250 000 €) ;
  main, au choix (1 objet en main, 2 avec « Ambidextre » 150 000 €) : arrosoir (pousses proches +30 %/niv.),
  sifflet (ouvriers proches +20 %/niv.), mégaphone (+15 %/niv. de clients dans le lieu où tu es).
- Torse : tenues purement décoratives (200 → 10 000 €) ; sauvegarde v7 : la différence de prix est remboursée.
- Le personnage montre sa toque, son sac, ses chaussures et l'objet en main ; cercle de portée autour de lui.
- En jeu : deux boutons ronds (mairie, personnage) à côté du « ? », qui brillent quand un achat est possible.

## Version 4.3 — le matériel scolaire
- Local des fournitures (école) : 8 étagères qui se remplissent seules (crayons, cahiers ; géométrie, atlas ;
  calculatrice, microscope ; ordinateur, blouse), achetées avec la 1re classe de leur niveau (sans niveaux : elles restaient toujours pleines ; sauvegarde v8 : remboursés).
- Chaque étagère achetée : +15 % de vitesse de formation pour les classes de son niveau (et ses demandes).
- Une classe occupée demande de temps en temps un matériel de son niveau (bulle près du tableau). L'apporter :
  vitesse ×2 pendant 1 min. Le joueur ne ramasse au local que ce qui est demandé.
- Surveillant (2 max, 3 000 / 15 000 €) : apporte le matériel demandé, un peu moins vite que le joueur.
- Gains hors ligne : 4 % + 1 % par ouvrier (max 8 %).
- Le « ! » des lieux sur la carte ne compte que les achats réellement possibles.
- Près d'un achat au sol (bâtiment, champ, salle, étagère…), la bande d'aide dit à quoi il sert.

## Version 4.4 — passage en caisse
- Chaque objet prend du temps à passer en caisse : caissière 1,2 s (÷ 1 + 0,5 × niveau), joueur 0,5 s.
  Barre de progression au-dessus du client servi. Améliorations (8 niv., case près de la caisse, après la caissière) :
  Tiroir-caisse (ferme, 300 €), Scanners (supermarché, 1 500 €), Caisse rapide (boulangerie, glacier).
- En file, impatience après 25 s (au lieu de 12) et départ après 90 s (au lieu de 60).
- 5 s par objet testé : la ferme (une seule caisse) s'effondrait (réputation à 1, revenus ÷ 10).
- Crème : 48 € (au lieu de 26).

## Version 4.5 — l'arrêt de bus
- Dans chaque lieu, à gauche du portail : un arrêt de bus avec une case par destination (carte + lieux achetés).
  On reste 1 s dessus pour partir. Raccourcis : 1 ferme, 2 école, 3 supermarché, 4 boulangerie, 5 glacier, 0 carte,
  V mairie, P personnage.
- Points de l'école : « +10 » place jusqu'à 10 points (ce qui reste), cumulable.

## Version 4.6 — moins de creux avant le supermarché
- Supermarché : 8 000 € (au lieu de 15 000). Joueur automatique : école ≈ 1 h, supermarché ≈ 2 h (au lieu de 2 h 35),
  ≈ 116 €/s à 4 h (au lieu de ≈ 60).
- Commandes spéciales (ferme) : dès 3 étals, un habitant commande 2 ou 3 produits (5 à 12 de chaque) ; 5 min pour les
  déposer dans la caisse des commandes (à droite des étals) ; récompense = 3 × leur prix. Une à la fois, une nouvelle
  3 à 5 min après. Affichée sous l'argent, partout. Sauvegardée (le temps s'arrête jeu fermé) ;
  si le temps est écoulé, ce qui a été livré est payé au prix normal. (Plus tard : commandes des lieux, ex. la cantine de l'école.)
- Prix des rayons du supermarché selon la progression : ferme 800 + ½ du prix du bâtiment/champ du produit
  (farine de maïs 850 … chocolat 50 800) ; boulangerie 3 000 + ½ recette ; glacier 5 000 + ½ recette.
- Bandeau « Objectif » en haut : le prochain lieu à acheter et la progression de l'argent.

## Version 4.7 — tout au clavier
- Sur la carte et dans les fenêtres, les flèches (ou ZQSD) déplacent une grosse main. Restée 1 s sur un bouton, elle
  l'active (cercle qui se remplit ; on peut répéter, ex. « +1 ») ; Entrée ou Espace : tout de suite.
  « Quitter » et « Oui, tout effacer » : seulement avec Entrée. La main ne vise que la fenêtre ouverte ;
  en haut/bas d'une liste, elle la fait défiler. La souris reste utilisable (la main disparaît dès qu'elle bouge).

## Version 4.8
- Camions animés : celui du quai repart vers la droite au départ, le suivant arrive par la gauche et attend ;
  à la réserve des lieux livrés, le camion arrive, décharge et repart.
- Drones : base → étal (ils y prennent la marchandise) → quai → base, à vitesse constante (80, +30 %/niveau
  d'hélices) ; les trajets n'ont plus tous la même durée ; petite pause aléatoire sur la base.
- École : les salles ont des murs (porte en bas) ; élèves, surveillants et joueur passent par la porte.
- Blé : 3e ouvrier (18 000 €, entre le verger et les fleurs). Équipe complète : 10 ouvriers.
- Gains hors ligne : calculés sur le revenu moyen des 15 dernières minutes, sans les commandes spéciales.
- Réserve pleine : un quai n'accepte plus un produit dont la palette d'arrivée serait pleine (en comptant ce qui
  attend sur les quais) ; le camion ne charge que ce qui rentre : plus rien ne se perd.
- Correctif : le « 1 objet = 10 » du supermarché n'était en fait pas appliqué par les camions ; il l'est maintenant.

## Rythme mesuré (joueur automatique, cmd/balance, v3.0)
moulin à maïs ≈ 4 min, blé ≈ 9 min, moulin ≈ 16 min, étable ≈ 42 min, fromagerie ≈ 1 h,
crèmerie ≈ 1 h 30, potager ≈ 2 h, verger ≈ 2 h 15, supermarché ≈ 2 h 45, boulangerie ≈ 7 h, ruches ≈ 8 h.
Un humain sera sans doute un peu plus lent.

## Technique
- Go 1.24 + Ebitengine 2.8. Tout est dessiné par le code (formes + police Go), sons synthétisés.
- Dessin : triangles regroupés en lots, rendu 2× puis réduit (anticrénelage).
- Interface : 1080 unités de haut, largeur selon l'écran.
- `sim/` : toute la logique (sans affichage) ; `cmd/balance` : simulation accélérée.
- Sauvegarde : ToutUneVille_sauvegarde.json à côté de l'exe (jeu portable sur clé USB), toutes les 20 s et en quittant.
  Secours si le dossier n'est pas modifiable : %AppData%\ToutUneVille\sauvegarde.json (aussi relue au premier lancement).
- Compilation Windows :
  `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-H windowsgui -s -w" -o dist/ToutUneVille.exe .`
- Options : `-fenetre` (démarrer en fenêtre), `-captures dossier` (captures automatiques, développement).

## Prochaines étapes prévues
- Fleuriste (vend les fleurs), restaurant (rayon fromage, fruits), hôtel, zoo, école, hôpital, prison, parc d'attractions…

- Affichage de l’argent : sous 1 M le montant entier (999 999 €), au-delà notation scientifique avec un chiffre après la virgule : 1,2 × 10⁶ €, 1,5 × 10⁹ €…
- Ouvriers qui s’ennuient : après 8 s sans travail, ils s’installent (assis sur une botte de foin et s’endorment après 40 s, bâillent en s’étirant, ou sifflotent). Ils repartent dès qu’il y a du travail.
- Tenues du personnel selon le lieu (staffLook) : ferme vert + casquette ; boulangerie veste blanche, tablier fariné, charlotte ; glacier tablier rayé rose + calot ; supermarché polo bleu, gilet orange, badge ; école gilet + sifflet. Caissières assorties. Hors ferme, l’ouvrier qui s’ennuie s’assoit sur une caisse.
- Parking sous la route de chaque boutique (ferme, supermarché, boulangerie, glacier) : sim/parking.go + parking_draw.go. Clients en voiture (≈64 %), vélo, moto (casque), trottinette ; ils se garent, traversent au passage piéton, reviennent au véhicule. Sans place libre : à pied. Les voitures s’arrêtent au passage piéton et derrière le véhicule de devant ; voitures de passage toutes les 6–16 s. Les clients encore en voiture ou qui retournent à leur véhicule ne comptent pas dans le plafond de clients (équilibre inchangé : école 1:04, super 2:01, boulangerie 4:05). Hauteur des lieux : ParkH.
- Clients : ils contournent les rayons et caisses (cWalk, A* de nav.go, comme les employés). Supermarché : abri à caddies (24) à droite de l’entrée ; on prend un caddie en entrant (sinon on porte à la main), les articles vont dedans (puis des sacs après la caisse) et on le rapporte avant de sortir. Équilibre inchangé.
- Commandes de la cantine : sources de commandes à parts égales (orderSources : ferme ; + école dès que école ET boulangerie ouvertes ; plus tard hôpital, prison…). La cantine commande, à livrer à la BOULANGERIE (caisse « Commandes » à droite des améliorations), 1 pain par personne présente : primaire = baguette, surveillants + secrétaire = baguette de maïs, profs = pain blé-maïs, collège = brioche, lycée = pain d’épices, université = gâteau aux fruits. 10 min, récompense ×3 comme à la ferme, livré partiel payé au prix normal. Order.At = lieu de livraison (sauvegardé). Bandeau hors du lieu : « Commande à livrer : Boulangerie » ; pastille sur la tuile de la carte et « ! » sur l’arrêt de bus.
- Réserve de fleurs (Kind Store, « reserve_fleurs ») à droite du champ de fleurs, construite avec lui : 40 fleurs. L’ouvrier des fleurs n’y porte que le surplus (priorité ×0,12, les ruches d’abord) ; le joueur peut y déposer. Base de drone prévue dessous (StorePad, en pointillés « Drone : avec le fleuriste ») pour le futur camion vers le fleuriste.
- Camions et mécaniciens (save v9) : plus d’amélioration « Camion » à la mairie (niveaux remboursés au chargement). Les camions d’un lieu s’améliorent avec ses points « mécaniciens » : intervalle 25 s/(1+pts/100), charge 20+pts/2, quai 10+pts/10 places par produit. Les mécaniciens n’agissent plus sur les drones (soutes et hélices seulement). Badge « Poids lourd » : 100 points de mécaniciens dans un lieu. Bot : avant le supermarché il économise au-delà de 3 000 € (reproduit l’ancien blocage), et met la moitié de ses points « ouvrier » en mécaniciens une fois le supermarché acheté. Équilibre : école 1:04, super 1:58, boulangerie 3:49, 203 €/s à 5 h.
- Entrées des machines (icône + quantité) dessinées par-dessus les personnages (machineInputs) : la caissière de la boulangerie cachait le miel du pain d’épices. 3 entrées : nombres en quinconce.
- Mairie = page des points de l’école (touche V, tuile, raccourci), avec le veilleur de nuit en bandeau en haut. Avant l’école : petite fenêtre avec le veilleur seul. Tuile : « Points de l’école, veilleur » ou « N point(s) à répartir ! ».
- Quais : 20 places par produit au départ (+1 tous les 10 points de mécaniciens). Soute des drones : +2 objets par niveau (19 au max), prix ×1,48 par niveau (dernier ≈ 35 000 € à la ferme).
- Commandes : le hasard est ré-amorcé à chaque chargement (graine fixe → même première commande à chaque lancement) ; jamais deux fois la même commande ni le même habitant de suite (lastOrder).
- Soute : 30 niveaux (61 objets), hélices : 15 niveaux (×1,9 par niveau). Quai = 20 + 2 × niveau de soute du lieu + points de mécaniciens/10. Camion : 60 produits par voyage + points/2.
- Points « ouvrier » : à la ferme = Agronomes, à la boulangerie et au glacier = Chefs : +1 % de vitesse de production (champs, animaux, machines) par point. Ils n’agissent plus sur les ouvriers (vitesse et charge : bottes/brouettes, baskets/plateaux seulement). Supermarché : pas de case (croix « Rien à fabriquer ici ») ; case mécaniciens vide : croix « Pas de camions ici ».
- Glacier : 4 machines. Glace nature (lait, gratuite, 42 €), Sorbets (fruits, 12 000, 55 €), Glace aux fruits (lait + fruits, 22 000, 90 € — nouveau produit GlaceFruitLait ajouté en fin de liste), Glace au chocolat (lait + chocolat, 40 000, 130 €). La crème ne sert plus au glacier. Badges : « Cent boules » (glaces nature), « Quatre parfums ».
- Commande ratée (temps écoulé) : −1 étoile de réputation au lieu de livraison (ferme ou boulangerie). Annoncé dans le message de la nouvelle commande.
- Ferme : on apparaît juste après le portail (1250, 1440), comme dans les autres lieux (avant : sur le chemin sous les champs).
- « FERME » (lu « fermé ») et « La ferme » (insulte) remplacés par « MA FERME » / « Ma ferme » : panneau d’entrée, pancarte, camions, bus, carte, badges, nom du lieu.

## Version 4.0 — nouvelle ferme, crèmerie, immense supermarché (sauvegarde v10 : les parties d'avant repartent de zéro)
- **Ferme** (6 900 px de large, 14 emplacements de bâtiments) :
  - maïs (16 colonnes, 48 pousses) → moulin à maïs, poulailler, **mare aux canards** (cuisses de canard, 2 500 €) ;
  - blé (22 colonnes, 66 pousses) → moulin, étable, **chèvrerie** (lait de chèvre, 3 500 €), **bergerie** (lait de brebis, 5 000 €) ;
    plus de fromagerie ni de crèmerie à la ferme ;
  - potager, verger, fleurs (+ réserve), canne, cacao, puis **caféiers** (café vert, non vendable, 150 000 €) →
    **torréfacteur** (café en grains, 220 000 €) → **moulin à café** (café moulu à partir du café torréfié, 320 000 €) ;
  - champs remplis **en serpent** (rangée du bas →, milieu ←, haut →) ; prix des pousses inchangé (1,18^n) ;
  - ouvriers : 3 au maïs, 4 au blé, 1 au café (13 au total) ; 16 étals ; caisse des commandes, hangar et cases déplacés à droite ;
  - quais : supermarché, boulangerie, **crèmerie**, glacier.
- **Boulangerie** : pains et viennoiseries seulement (baguette, baguette de maïs, pain blé-maïs, brioche,
  **croissant** = farine + beurre, **pain au chocolat / chocolatine** = farine + beurre + chocolat ; il faut la crèmerie).
  Tartes, gâteaux et pain d'épices iront à la future pâtisserie. Cantine : lycée = croissant, université = pain au chocolat.
- **Crèmerie** (35 000 €, entre la boulangerie et le glacier) : 3 laits ; crème, beurre (fait avec la crème,
  chaîne interne), fromage blanc, yaourts et fromages de vache, chèvre, brebis (9 machines, 5 par rangée, 3 400 px).
  Quais vers le supermarché et la boulangerie (le beurre).
- **Glacier** : 60 000 €. Ateliers paramétrables (largeur, machines par rangée, quais) ; recettes « needs » (lieu requis).
- **Supermarché immense** : 4 800 px, 3 allées de 13 rayons (39 places, 35 produits), 5 caisses, 10 employés,
  40 caddies, 45 clients, parking 3 000 px.
- Carte : 5 colonnes de tuiles (réduites à la taille de l'écran) ; crèmerie ; pâtisserie et crêperie « Bientôt ». Bus : touche 6.
- Camions : le nom du lieu de départ sur le flanc. Nouveaux badges (chèvrerie, bergerie, mare, café, crèmerie).
- Bot : laisse tomber les commandes sans pénalité ; au-delà de 30 pousses, il fait d'abord les bâtiments.
- Équilibre (bot) : école 1:09, super 2:02, boulangerie 4:02, crèmerie 5:26, glacier 7:22, chocolaterie 9:06 ; 590 €/s à 10 h.
- Crèmerie : 2e rangée de machines décalée d’un cran à droite (shift2), plus rien derrière la caissière.
- Mare aux canards : proposée juste après le poulailler (1 000 €, cuisses à 22 €), étal juste après les œufs.
- Ferme : une cour tout à gauche (près du portail, qui ne bouge pas) : le hangar en haut avec bottes, brouettes, soute, hélices ; la caisse des commandes sous le chemin central. Le reste de la ferme est décalé de FarmX0 = 650 (shiftFarm), largeur 7 550.
- Cour de gauche (réservée au joueur, dalle de béton avec liseré jaune « RÉSERVÉ AU FERMIER ») : hangar en haut (niveau des champs), dessous 1re colonne bottes/brouettes, 2e colonne soute/hélices ; caisse des commandes sous le chemin central. Le chemin du haut commence au champ de maïs.

## 4.1 — étape 1 : corrections en vrac
- Laits de vache, chèvre et brebis au même prix : 30 € (à 18 €, le revenu chutait de 40 % ; à rediscuter avec le supermarché). Dessins distincts : bouteille bleue (vache), brique verte (chèvre), pichet en terre (brebis).
- Moulins : les pales tournent dans l'autre sens. Veilleur : bouton d'achat vert.
- Commandes : plus d'annonce « −1 étoile » dans le bandeau ; l'encart montre « Gain : X € » et « Perte : 1 ★ ». Gain doublé si livrée en moins d'1 min (ferme) ou 2 min (cantine) : Order.Fast, FastLeft().
- Niveau des machines écrit en clair (« Niv. N ») au lieu d'étoiles ; mot « DÉPÔT » retiré.
- 40 noms de clients pour les commandes.
- Ouvriers : ce qu'ils portent sans destination est gardé, et les places libres servent à d'autres travaux ; compost seulement les bras pleins depuis 30 s.
- Camion vers le supermarché seulement s'il y a un rayon ouvert pour un produit du lieu (DockActive).
- Drone d'une vitrine proposé dès qu'un lieu acheté a une palette pour son produit (beurre → boulangerie).
- Croissants : il faut la baratte à beurre de la crèmerie ; pains au chocolat / chocolatines : beurre + chocolaterie (recipe.needs « lieu:station »).
- Ouvriers du blé : 6 (15 ouvriers à la ferme).
- Crèmerie plus haute (rangées 470 / 820, vitrines à 1190 ; file de 4). Boulangerie : 2e rangée décalée (rien derrière la caissière).
- Entrepôt (nouvelle amélioration, case dans la réserve) : palettes des ateliers 40 → 100 (+5 × 12), supermarché 300 → 600 (+25 × 12).
- Réserve de fleurs invisible tant que le fleuriste n'existe pas.
- Équilibre : école 1:12, super 2:13, boulangerie 4:37.

## 4.1 — étape 2 : école et points (sauvegarde v11)
- L'école est offerte dès le début (une table, un prof). Hors ligne, plus d'élèves : seulement de l'argent. Aide « ? » propre à l'école (avertissement en rouge, durée des formations, taux d'échec).
- Échec en fin de formation : 1 sur 6 (Pédagogie : 1 sur 7, 8…) ; l'élève repart avec un point du niveau d'en dessous (rien au primaire). Levels « ecole.echecs ».
- Points (sim/school.go, PointKind{Key, Type, Group, Name, Desc, Val}) :
  - Primaire (ouvrier) : Agronomes (champs), Soigneurs (animaux), Entraînement (vitesse, charge /10), Mécaniciens (drones vitesse/charge, camions départs/charge, machines de la ferme vitesse/rentabilité), Chefs (machines qui cuisinent, vitesse/rentabilité).
  - Collège (marketing) : Marketing, Patience (multiplie les seuils d'attente des clients).
  - Lycée (comptable) : Comptables, Négociation (améliorations ×1/(1+n/100)), Logistique (temps des commandes).
  - Université (formation) : Formation, Pédagogie.
  - Rentabilité : chance de ne pas consommer un ingrédient n/(n+100), au plus 75 %.
- Les champs n'ont plus de niveau (remplacé par les Agronomes).
- Mairie : 4 onglets (un par niveau), liste des lieux à gauche, cartes à droite, +1 / +10, Valider / Annuler.
- Bot : répartit chaque type de point là où il y en a le moins parmi ses usages préférés.
- Équilibre : super ~2:15 (inchangé).

## 4.1 — étape 3 : zones du joueur et personnage (sauvegarde v12)
- Chaque lieu a sa zone du joueur (sim/yard.go, yard_draw.go) : dalle de béton grise (même fond que la cour de la ferme),
  560 de large ; à gauche de la ferme, des ateliers et de l'école (x = -620 pour les ateliers et l'école), à droite du supermarché (près de son entrée).
  De haut en bas : hélicoptère + téléporteurs (2 colonnes × 3), « OUVRIERS » (embauche, vitesse, capacité),
  caisse des commandes, vestiaire. Le joueur arrive à l'hélicoptère.
  Le hangar de la ferme est supprimé. Le surveillant de l'école est aussi dans la zone.
- L'hélicoptère remplace le bus (même fonctionnement : 1 s sur une case, raccourcis 0-6) ; son rotor tourne pendant l'attente.
- Vestiaire : s'arrêter sur son paillasson ouvre la page « Personnage » (aussi touche P et la carte). Le bouton rond du HUD est supprimé.
- Personnage : tenues décoratives supprimées. Zones tête / dos / mains / pieds, un objet porté par zone (« porte.<id> » dans Levels),
  mains : 1 objet (2 avec Ambidextre, acheté en restant appuyé 1,2 s sur la 2e case). Le premier niveau acheté se porte tout seul si la zone est libre.
  Un objet rangé n'agit pas.
- Ensemble du diplômé (5 niveaux chacun) : chapeau (tête), toge (dos), craie (main), mocassins (pieds) :
  +20 % de vitesse de formation par niveau, seulement dans la classe où se trouve le joueur (jusqu'à ×5 avec les 4 au max).
  Prix 2 000 → 220 000 €. Badges « Nouveau look » (un équipement) et « Diplômé » (les 4 portés).
- Correctif : plus de Soute ni d'Hélices à acheter (zone du joueur) ; les drones ne s'améliorent qu'avec les points
  « Mécaniciens » de l'école : charge 1 objet + 1 tous les 2 points, vitesse +3 % par point. Le joueur automatique y met des points.
  Les commandes remontent sous les améliorations des ouvriers.

## 4.1 — étape 4a : la nouvelle ferme (sauvegarde v12)
- Plus de viande : mare aux canards, potager et porcherie supprimés (cuisses de canard, légumes, saucisses).
  Champ de maïs : 10 colonnes (au-dessus de ses 2 bâtiments), 2 ouvriers au lieu de 3.
- La canne, le cacao et le café (avec sucrerie, chocolaterie, torréfacteur, moulin à café) quittent la ferme : ils iront
  à la ferme exotique (tuile « Bientôt » sur la carte, entre la crèmerie et la boulangerie). En attendant, pas de sucre en poudre
  ni de chocolat : glaces et pains au chocolat sont bloqués (« ferme exotique, bientôt ») ; la cantine commande des croissants à la place.
- Serre (6 colonnes : carottes, tomates, courgettes, poireaux, choux, pommes de terre) et verger (fraisiers, pommier, poirier,
  cerisier, abricotier, figuier) : une colonne par produit, achetée à part (6 000 → 42 000 €, serre et verger mêlés).
  Chaque colonne a 1 case au départ ; on achète les suivantes vers le haut (8 max, prix = prix du produit × 12 × 1,4^n).
  La ferme monte donc au-dessus des champs (Place.Top = -420, grille de chemins décalée) ; au-dessus du maïs et du blé : un étang (décor).
- Serre : après chaque récolte la case est vide ; on prend des sachets de graines (gratuits, caisse au pied de la colonne ;
  on ne prend que ce qu'il faut) et on marche sur les cases vides. Les ouvriers (1 par colonne) replantent aussi.
- Petits marchés (étals à auvent) sous la serre et sous le verger ; les 7 autres étals s'élargissent jusqu'à la serre ; le miel sous les ruches.
- Fleurs 50 000 €, ruches 65 000 €. 21 ouvriers (2 maïs, 6 blé, 12 colonnes, 1 fleurs), 250 → 60 000 €.
- Glacier (6 machines) : sorbet à la fraise (gratuit), sorbet à la cerise ; glaces fraise, cerise, chocolat, poire-chocolat
  (lait + crème + sucre en poudre + œufs + fruit ou chocolat). La crèmerie envoie aussi un camion au glacier.
- Ordre des lieux : supermarché 7 000 €, crèmerie 25 000 €, (ferme exotique), boulangerie 45 000 €, glacier 100 000 €.
  Joueur automatique : supermarché ≈ 2 h 05, crèmerie ≈ 3 h 30 (boulangerie à 60 000 € : ≈ 6 h 05). Il économise pour le prochain lieu
  quand il est à ~10 min de recettes, et n'achète plus les niveaux de champs (que le joueur ne peut pas acheter).
- Supermarché : 14 rayons par allée (41 produits), palettes plus étroites. Nouveaux produits : sucre roux / en morceaux / en poudre (pour plus tard).
- Badges : potager, canne, mare, café, sucrerie, chocolaterie, cacao, porcherie retirés ; « Main verte », « Serre pleine »,
  « Jardinier » (100 sachets plantés), « Verger plein » ajoutés.

## 4.1 — étape 5 : le nouveau supermarché (sauvegarde v12)
- Plan (sim/super.go, sim/superpark.go, super_draw.go) : zone du joueur à gauche ; entrepôt (en haut) et magasin découpés en espaces
  (Produits de la ferme, Légumes, Fruits, Crèmerie, Boulangerie, Glaces, puis Boucherie-charcuterie, Pâtisserie, Alcools « Bientôt ») ;
  colonne des caddies et sas d'entrée/sortie ; route d'accès ; parking (168 places au sol) et parking à étages (70 places par étage).
  Route du haut : camions et voitures ; route du bas : voitures. Chaque camion s'arrête derrière l'espace de son lieu.
- Un espace est fermé (barrière, « Il faut la crèmerie »…) tant que son lieu n'est pas acheté (légumes : la serre ; fruits : le verger).
- Magasiniers par espace (4 chacun, « super.ouvriers_<espace> ») : ils ne remplissent que leurs rayons.
- Même prix que partout (SellMult 1) ; 1 objet livré = 2 en réserve (au lieu de 10).
- Clients : tous en voiture et avec un caddie. Au plus (rayons ouverts × 5) × (1 + marketing) en même temps (300 max).
  Chacun achète entre 3 et (rayons ouverts × 2) articles, de droite à gauche, puis passe à la caisse la plus avantageuse (file + distance).
  Pas de place ou pas de caddie : il repart aussitôt, mécontent (réputation −0,02) ; l'aide le signale.
- Places et caddies : 10 au départ, +10 par achat (case au bord du parking, case au-dessus des caddies), 300 au plus.
- 12 caisses (1 500 € → ~200 000 €) et leurs caissières ; caisses à scanner : 0,5 s par article (÷ points « Caisses »).
  Scanners supprimés. Points Mécaniciens du supermarché : Caisses (vitesse) et Caddies (vitesse de marche des clients).
- Petits magasins (ferme, boulangerie, crèmerie, glacier) : un client prend 3 produits différents au plus, 10 objets au plus ;
  10 clients en même temps au plus (+1 tous les 10 points de marketing).
- Performances : ~0,2 ms par pas de simulation avec 160 clients ; les clients et voitures hors de l'écran ne sont pas dessinés.
- Prix des lieux revus : supermarché 7 000 €, crèmerie 22 000 €, boulangerie 50 000 €, glacier 80 000 €.
  Joueur automatique : supermarché ≈ 2 h 10, crèmerie ≈ 2 h 45, boulangerie ≈ 4 h 45, glacier ≈ 7 h (estimé).
  Il économise aussi au supermarché (et ailleurs) pour le prochain lieu quand il est à ~10 min de recettes.

## 4.1 — l'école redessinée (d'après le plan)
- À gauche, la cour des petits (parcours de tricycles avec trois petits qui roulent, toboggan et balançoires, marelle).
- En haut, la cour de récréation : terrain de basket, deux tables de ping-pong, tables de pique-nique, bancs sous les arbres, joli parc avec bassin.
- Rangée du haut : P1, P2, P3, C1, C2, C3, L1, L2 (porte en bas, sur le couloir). Rangée du bas : P6, P5, P4 (porte en haut à droite),
  l'accueil (portes sur le couloir et sur la rue, en face du portail) et le local des fournitures (sous C2 à L2).
  À droite, sur les deux rangées : l'amphithéâtre de l'université (gradins en arc, 6 places, porte sur le couloir).
  Collège : 3 salles (au lieu de 4). Géométrie : SchoolX0, SchoolColW, SchoolRow1Y/2Y, SchoolAmphi(), SchoolAccueil(), SchoolFourn().
- Chaque salle a une case « Vidéoprojecteur » (prix du prof × 1,5) : formation 50 % plus rapide dans la salle (écran et faisceau quand on y fait cours).
- Personnage : la craie devient un livre (livre de poche → livre d'or), même effet.

## 4.1 — retouches après essai
- Drones : achetables dès que le rayon du supermarché peut être acheté (ou est ouvert), ou dès que le lieu qui s'en sert
  est acheté (ex. le beurre, dès la boulangerie), même sans sa machine.
- Personnage : plus de niveaux (chaussures et sac 12, toque 10, objets en main et ensemble du diplômé 9 ; jusqu'à ~7 M€).
- Boulangerie et glacier « hauts » comme la crèmerie (plus d'espace entre les rangées de machines et les vitrines).
  Ateliers : le bâtiment commence à x = 20 (les cases caissière, vitrines, caisse rapide sont dedans).
- Ferme : la case d'embauche est sous le milieu de son champ et dit « Ouvrier du blé 1/6 », « Ouvrier du maïs 1/2 »…
- Bandeau « Va accueillir les élèves à l'école (touche 2) ! » ailleurs qu'à l'école, quand des élèves attendent, qu'une table
  est libre et qu'il n'y a pas de secrétaire (après les 5 premiers clients de la ferme).
- Bottes, brouettes, baskets, plateaux, rollers, chariots supprimés partout : seuls les points d'Entraînement (primaire) font
  marcher et porter les employés (charge : +1 objet tous les 5 points, au lieu de 10).
  Joueur automatique : supermarché ≈ 2 h 12, crèmerie ≈ 2 h 47, boulangerie ≈ 5 h 05.
- Vitesse des caisses : seulement les points « Caisses » (Mécaniciens, +2 % par point) dans tous les lieux qui vendent ;
  tiroir-caisse et caisse rapide supprimés.
- Plus de niveaux pour les bâtiments (fours, moulins, animaux…) : seuls les points de l'école les accélèrent
  (soigneurs, mécaniciens, chefs : +3 % par point). La case « Stock + » reste, à la place de l'ancienne case « Niveau ».
  Badges « … au maximum » de la ferme : agrandir le stock au maximum.
- École : tables moins chères (primaire 60 € → 300 € pour la 1re salle) ; l'accueil n'a plus de murs.
  Joueur automatique : supermarché ≈ 2 h, crèmerie ≈ 2 h 55, boulangerie ≈ 4 h 55.

## 4.1 — étape 6 : la ferme d'élevage et la boucherie-charcuterie (sauvegarde v13)
- Deux nouveaux lieux après le glacier : la ferme d'élevage (90 000 €) et la boucherie-charcuterie (140 000 €).
  Raccourcis 7 et 8 ; l'hélicoptère a maintenant 4 rangées de téléporteurs (la rangée des employés descend à y = 640).
- Ferme : le grenier à foin (20 000 €, au-dessus du champ de blé, invisible tant que la ferme d'élevage n'est pas achetée)
  transforme le blé en foin, rangé dans la grange à foin. L'étang est plus petit. Nouveau quai « → ferme d'élevage ».
  Les ouvriers du blé s'occupent aussi du foin.
- Ferme d'élevage (pas de clients, ne gagne rien elle-même) : réserve maïs / foin / pommes de terre ; 4 espaces, chacun
  ses éleveurs (embauche dans la zone du joueur) :
  volaille (poulailler et mare : maïs → poulets, canards ; abattoir à deux entrées), bovins (grand pré : foin → vaches
  et veaux, avec le taureau), ovins (grand pré : foin → moutons et agneaux), porcs (porcherie : pommes de terre → cochons).
  Les carcasses vont dans les frigos (chambre froide) ; des drones de fret (5 × plus de charge) les portent au camion
  frigorifique vers la boucherie. Déblocages : mare 15 000, bovins 40 000 et ovins 60 000 (il faut le grenier),
  porcs 80 000 (il faut les pommes de terre). Points : soigneurs (bâtiments d'animaux), mécaniciens (abattoirs, drones, camions).
- Boucherie-charcuterie : une caissière, peu de clients (mais chers), 4 espaces avec leurs bouchers, chaque espace
  montre son prochain déblocage.
  Volaille : poulet cru entier (préparation), rôtisserie, cuisses, puis aiguillettes, magret, foie gras (canards).
  Bœuf et veau : steaks, hachoir, rôtis, morceaux de veau. Mouton et agneau : gigots, côtes. Charcuterie : saucisses,
  jambons, côtelettes, pâtés. Quai vers le supermarché, dont l'espace « Boucherie-charcuterie » ouvre (6 colonnes).
- Nouvelles mécaniques (sim) : machines à deux sorties (Two / Item2 / Pile2 / OutZone2, pour les prés), étals sans
  clients (NoSale : grange, frigos), stations d'un espace (Team) et ShopZone (elevageZones, boucherieZones).
- Joueur automatique : supermarché ≈ 2 h, crèmerie ≈ 3 h, boulangerie ≈ 4 h 55, glacier ≈ 7 h 05,
  ferme d'élevage ≈ 7 h 35, boucherie ≈ 9 h 20.

## 4.1 — retouches après l'étape 6
- Partout : les cases d'embauche disent « +1 » (« Ouvrier du maïs +1 », « Magasinier +1 », « Volaille +1 », « Surveillant +1 », « Table +1 »…) au lieu de « 1/2 » ; les messages aussi.
- Commandes : « Nouvelle commande : le coiffeur ! », « Commande livrée (le coiffeur) : +… € » (plus de « de le » / « de les »).
- École : les surveillants sans travail attendent dans le couloir, devant l'accueil et le local des fournitures
  (avant : dehors, en bas de l'école, hors de vue et cachés par le bandeau d'aide).
- Supermarché, taille des caddies : le client tire bien 3 à (rayons ouverts × 2) articles. Il réserve ce qui est
  libre en rayon ; le reste de sa liste, il le prend s'il le trouve à son passage, ou attend un peu le magasinier,
  puis repart sans râler s'il a au moins 3 articles. Un client ne vient que s'il y a au moins 3 articles libres.
  Rayons : 30 articles (au lieu de 16), +4 par niveau.
  Correction : un client qui prenait plus que sa réservation faisait « grossir » les réservations du rayon (fuite).

## 4.1 — petites modifications (commandes crèmerie, formation express…)
- Crèmerie : une caisse des commandes (zone du joueur) ; des particuliers y commandent 2 ou 3 de ses produits.
- École : le local des fournitures est plus haut (le nom des étagères ne touche plus le mur) et a 3 portes,
  une entre deux colonnes d'étagères. Panneau du local entre deux colonnes.
- École : formation express. Case « Formation express » dans chaque classe (en haut à gauche) ; le prix dépend de ce
  qu'il reste à faire aux élèves assis (formation complète : 200 € primaire, 800 € collège, 3 000 € lycée,
  12 000 € université, par élève). Un intervenant (costume, mallette) arrive par le portail, fait cours 10 s :
  toute la classe est formée (même sans prof), puis il repart. Il faut quitter la case pour en racheter une.
- Ferme : le maïs pousse plus vite (5,5 s au lieu de 7) pour un début de partie plus rapide.
- Supermarché : dans l'entrepôt, seulement le nombre sous chaque palette (le dessin est déjà sur la palette),
  et plus le nom des espaces (il est déjà sur le panneau des rayons).
- Sons : on n'entend que ce qui se passe dans le lieu où l'on est (les messages des autres lieux s'affichent quand même).
- Nouvelle commande : sonnerie de téléphone à l'ancienne (« drriing… drriing »), entendue partout.
- Graines : les ouvriers replantent sans bruit ; le joueur fait juste un tout petit « pof » très doux (le son joyeux reste pour l'achat d'une nouvelle case).
- Serre et verger, ouvriers plus malins : ils prennent d'abord leurs sachets de graines (au pied de la colonne),
  attendent les pousses qui mûrissent dans moins de 5 s pour cueillir toute la colonne d'un coup, replantent
  chaque case sur place, puis vont à l'étal. Ils ont une cagette (8 places : toute la colonne).
  Les sachets de graines ne prennent plus de place dans les bras (joueur compris).
  Mesure (une colonne de tomates de 8 cases, 1 ouvrier, étal vidé en continu, 20 min) : 99 → 225 tomates livrées.
- Verger : une case achetée est d'abord de la terre nue, puis l'arbre (ou le fraisier) pousse : petit, moyen, grand
  (Plant.Young, durée = 1,5 pousse de fruits). Ensuite les fruits : petits, moyens, gros (verts), puis gros et de la
  bonne couleur → on cueille ; seuls les fruits disparaissent, l'arbre reste. Dessin : drawOrchard (farm_draw.go),
  colonne dessinée du haut vers le bas (l'arbre de devant cache celui de derrière).
- Carte : « Port » (bientôt) après le glacier, « Caviste » (bientôt) après la boucherie.

## 4.1 — la ferme exotique (sauvegarde v14)
- Nouveau lieu après la crèmerie (35 000 €), touche 5 (l'hélicoptère a 5 rangées de téléporteurs, touches 1 à 9).
  Comme ma ferme : champs, bâtiments, étals, caissière, clients, commandes de particuliers, drones, camions
  (supermarché, boulangerie, glacier). Six espaces, chacun ses ouvriers (embauche sous le premier champ de l'espace) :
  fruits (ananas, bananier, cocotier, manguier : colonnes comme le verger, l'arbre pousse puis les fruits),
  sucre (canne → raffinerie : sucre en morceaux → moulin à sucre : sucre en poudre),
  cacao (cacaoyers → chocolaterie : cacao + sucre en poudre = tablettes), café (caféiers → torréfacteur : grains →
  moulin à café : moulu), riz (grande rizière, sachets), épices (vanille, théiers → séchoir : boîtes de thé, cannelle).
  Le moulin à sucre est « partagé » (Station.Shared) : la chocolaterie (espace cacao) s'y sert.
- Ordre et prix : ananas offert, canne 3 000, raffinerie 6 000, bananier 8 000, moulin à sucre 11 000, cocotier 15 000,
  cacao 20 000, chocolaterie 26 000, manguier 32 000, café 38 000, torréfacteur 45 000, moulin à café 55 000,
  rizière 65 000, vanille 80 000, théiers 95 000, séchoir 105 000, cannelle 120 000.
- Glaces au sucre : il faut le moulin à sucre (exotique:moulin_sucre) ; au chocolat : la chocolaterie.
- Supermarché : les fruits exotiques vont dans l'espace Fruits (4 colonnes) ; nouvel espace « Épicerie exotique »
  (sucres, chocolat, cafés, riz, vanille, thé, cannelle).
- Nouveaux produits : ananas, bananes, noix de coco, mangues, riz, vanille, feuilles de thé, thé, cannelle.
- Joueur automatique : crèmerie ≈ 2 h 45, ferme exotique ≈ 3 h 50, boulangerie ≈ 4 h 50, glacier ≈ 6 h 10.
- Supermarché : plus personne n'attend devant un rayon qui a des articles (il se sert, même s'ils étaient réservés
  par d'autres ; les réservations ne dépassent jamais le stock). Le compteur des caddies affiche ceux qui sont pris
  (« Caddies pris : 108 / 150 »), comme le parking (avant : ceux qui restaient, ce qui prêtait à confusion).
- Serre : la caisse de sachets est à gauche, sa zone juste à droite, dans l'axe étal ↔ colonne (trajet tout droit).

## ⚠️ Règle : ne jamais remettre la sauvegarde à zéro
La joueuse le fait elle-même si elle le veut. Ne plus changer SaveVersion : les nouveautés doivent se charger
avec une sauvegarde existante (lieux, stations et niveaux absents : valeurs par défaut).

## 4.1 — école, supermarché
- École : quand il y a des tables libres en primaire (plus que d'élèves dans la file), un élève arrive toutes les
  2 à 4 s (sinon 8 à 12 s) : l'école tourne à pleine capacité même quand la formation est très rapide.
- Supermarché : l'épicerie exotique a sa barrière « FERMÉ » (toutes les barrières « barriere_* » se dessinent).
- Supermarché en miroir : les premiers espaces (ferme, légumes, fruits…) sont à droite, près de l'entrée, et dans
  chaque espace les premiers rayons et palettes aussi. Mur gauche à x = 100.
- Zone du joueur du supermarché : sous la route du bas, à l'horizontale (setupYardH : hélico, 5 × 2 téléporteurs,
  Rayons +, Entrepôt, vestiaire). Un passage piéton mène de la zone aux caddies et au sas d'entrée.
- Supermarché : les caisses 6 à 12 se dessinent (la liste des caisses dessinées s'arrêtait à 5).
- Parking du supermarché en sens unique (plus aucun croisement) : arrivée par la route du bas (voie proche), route
  d'accès à gauche du parking (on monte), allées vers la droite, route de sortie à droite des places au sol
  (SuperExitX, on monte) jusqu'à la route du haut, départ vers la droite. Le parking à étages est de l'autre côté de
  la sortie (porte sur la sortie, à l'allée n° 3 ; on y va par l'allée du bas). Flèches peintes, « ENTRÉE » / « SORTIE ».
- Crèmerie : les quais juste après la dernière vitrine (fromage de brebis), le premier vers le supermarché ; le
  compost dans la boutique, à droite des vitrines. Ferme : le compost à droite de l'étal du miel.
- École : local des fournitures encore plus haut (rangées d'étagères plus espacées).
- Ferme : la serre et le verger ont chacun une seule équipe (« Ouvrier de la serre », « Ouvrier du verger »),
  6 ouvriers au plus, un par colonne ouverte ; un ouvrier va où il y a du travail. Les anciennes équipes par
  colonne (sauvegardes) rejoignent leur équipe. Un ouvrier ne monte plus en haut d'une colonne pour moins de
  3 légumes ; les sachets de graines en trop sont rangés (plus de promenade avec un sachet inutile).
- Points d'école « Mains expertes » (Entraînement, tous les lieux avec employés) : les employés prennent et posent
  la marchandise plus vite (+1 % par point ; pts_mains).
- Serre, verger, fruits exotiques : quand il reste de la place dans ses bras après une colonne, l'ouvrier va
  cueillir dans une autre colonne de son équipe (chainJob : au moins 2 mûrs ou presque, la plus proche), puis livre
  tout. Mesure (3 colonnes, 1 ouvrier, 10 min) : il couvre les 3 colonnes et porte jusqu'à 2 sortes à la fois.
- Ferme exotique aérée (W = 5800 ; espaces écartés, passages entre les champs). Sucre et café : 3 ouvriers ;
  riz : 1 seul (un niveau d'embauche plus haut que le maximum est ramené au maximum). Des canards dans la rizière.
- Fleurs et miel : champ de fleurs doublé (6 colonnes, 18 cases ; la ferme s'élargit de 240), six ruches au lieu
  de trois (une ruche toutes les 3 s au lieu de 6 : deux fois plus de miel), 2 ouvriers des fleurs (le 2e à
  75 000 €, dernier de l'ordre d'embauche).
- Touches : Tab ouvre la carte de la ville (et, sur la carte, ramène au dernier lieu) ; Échap ne sert plus qu'à fermer une fenêtre.
- Supermarché : le message « Parking plein » / « Plus de caddies » passe avant les autres conseils (il était caché
  par « Tu peux acheter… ») ; en arrivant au supermarché, on voit aussi les refus survenus pendant qu'on était ailleurs.
  Mesure sur la vraie sauvegarde (marketing 111 : 263 clients voulus pour 250 places) : la réputation tombe de 2,1
  à 1,2 en 10 min ; avec assez de places et de caddies, elle remonte à 5. Mains plus rapides (×2, ×4, ×10) ou
  2 magasiniers de plus par espace : aucun effet mesurable (ce ne sont pas eux qui manquent) → non changés.
- Ferme exotique : la case d'embauche dit tout ce dont s'occupe l'ouvrier (« Épices » : vanille, théiers,
  cannelliers et séchoir à thé).
- Simulateur : `go run ./cmd/balance -humain [jours]` (options -seance 30 -pause 120) : séances de 30 min entre
  8 h et 23 h, jeu fermé entre deux (vraie sauvegarde écrite puis relue : gains hors ligne compris), partie
  commencée un dimanche à 15 h. Le bot « humain » finit ses lieux avant d'en acheter un autre (rien de bon marché
  restant, champs complets « pour le plaisir »), embauche volontiers, prend le veilleur de nuit, répartit ses
  points sur tout ce qui existe, gère pousses/niveaux/stocks des autres lieux ; pas de formation express.
  BALANCE_SAVE=fichier.json écrit la sauvegarde finale (pour comparer avec une vraie).
  Comparaison avec la vraie partie (dim 15 h → mar 17 h, 6 h de jeu) : la joueuse a 805 €/s, 8,9 M gagnés,
  super + crèmerie + exotique ; le bot humain 71 €/s, 1 M, super seulement (récoltes ×3, livraisons ×7 chez elle).
  Le bot reste donc très en dessous d'une vraie joueuse : ses temps sont un minimum pessimiste.
- Commandes : plus d'argent, un cadeau (sim/gifts.go), annoncé dès l'arrivée de la commande (message, bandeau,
  caisse) ; tirage au hasard parmi les habitants, donc les cadeaux spéciaux sont rares :
  particuliers → ventes ×2 dans le lieu de la commande (5 min) ; la maîtresse d'école → 1 point université
  (pts.formation) ; la cantine de l'école → l'école forme 2 fois plus vite ; le maire → lieux à moitié prix
  (PriceOf, ancien prix barré, conseil « va sur la carte ») ; le garagiste → camions et drones ×2 ; la fleuriste →
  « embellissement » : des fleurs poussent partout dans le lieu (en pot à l'intérieur) et des pétales volent,
  pour faire joli (flowers_draw.go) ; le vétérinaire → animaux et ruches ×2 dans toute la ville ; le marin →
  bateaux ×2 (seulement quand le port existera : il ne commande pas avant).
  Livraison express : cadeau doublé (voir plus bas : l'effet, pas la durée). Commande ratée : cadeau perdu et −1 étoile.
  Un seul cadeau à la fois : la commande suivante attend la fin du cadeau en cours (bandeau « Cadeau » avec le
  temps restant, à la place de la commande). Le temps du cadeau ne s'écoule que jeu ouvert (sauvegarde : Cadeau,
  relu après le calcul des gains d'absence). Les ventes ×2 ne comptent pas dans le revenu des absences.
  Une ancienne commande d'une sauvegarde (sans cadeau) est encore payée en argent.
- Ferme exotique, décor du nord (exoticWild, exotique_draw.go) : une cascade en haut à gauche (falaise, au-dessus des
  arbres fruitiers, x ≈ 420) tombe dans une rivière rapide qui file vers la droite (y −540 à −425, courant animé,
  rochers tous les 520), fougères sur la berge ; la forêt amazonienne au nord (deux rangs d'arbres) et sur les
  côtés ouest et est. Le joueur monte jusqu'à la rivière (Top −400) ; Place.DecoTop (−950) laisse la caméra
  montrer le décor.
- Supermarché : jamais plus de 290 clients à la fois (SuperMaxClients), quel que soit le nombre de rayons et de
  points de marketing. Parking (300) ou caddies (300) au maximum et plus de place : le client ne vient pas, sans
  mécontentement (les refus ne comptent que tant qu'on peut encore acheter places ou caddies ; des voitures qui
  repartent gardent leur place un moment, d'où 63 refus en 10 min avec 290 clients avant ce changement).
- Caisses : une caisse sans caissière garde ses clients (ils attendent puis repartent fâchés) : c'est voulu, le
  joueur doit encaisser lui-même ou embaucher (ne pas « corriger »).
  Mesure (sa sauvegarde, tous les rayons, 10 min) : 9 caisses avec caissière suffisent pour 290 clients (files de
  10 au plus, 2 fâchés, réputation 3 → 5) ; même résultat de 150 à 290 clients, et avec 12 caisses.
- LE PORT (sim/port.go, port_draw.go), 85 000 €, après le glacier (touche 0 ; Tab : la carte). Plan : la mer à
  gauche (et au nord), la criée en bois sans murs au centre (x 2700 à 4150), la zone du joueur à droite en haut
  (setupYard(4300)), le parking à droite en bas ; la route passe sur un pont au-dessus de la mer. Place.Water :
  rectangles d'eau où personne ne marche (solidAt).
  Une jetée part de la criée vers l'ouest (y 640 à 960). Quatre sections de 520 (de la criée vers le large) :
  au nord les bateaux (sardine, cabillaud, thon, saumon), au sud les parcs (huîtres, moules, saint-jacques,
  bassin à gambas). Un panier (Pallet, 40) par produit au bord de la jetée.
  Bateaux (Boat, pas sauvegardés) : 1 offert avec la section (la sardine dès l'achat), jusqu'à 4 (« Bateau +1 »
  sur la jetée) ; chacun a son couloir (il part tout droit vers le nord, disparaît, revient : jamais de croisement) ;
  20 s le voyage ; il vide sa soute (1 + pts/3) dans le panier. Aucun humain à bord.
  Parcs : 4 × 4 cases sous l'eau (Field Dredged : seule la drague récolte), 1 case au départ, « Case +1 » depuis
  la jetée ; une drague par parc (bras jaune, filet pour les gambas) va chercher la case mûre la plus proche et
  ramène 1 + pts/4 coquillages dans le panier.
  Ouvriers du port (8 au plus, embauche à l'entrée de la jetée) : paniers → étals. Le joueur porte et encaisse.
  Criée : deux rangées de 4 étals (poissons en haut, coquillages en bas ; un drone par étal vers le quai du camion),
  la caisse au milieu, clients seulement pour ce qui est en stock. Supermarché : espace « Poissonnerie ».
  Points (mécaniciens) : bateaux vitesse / soute, dragues vitesse / quantité ; plus les points habituels.
  Prix de vente : sardine 30, cabillaud 50, thon 70, saumon 90, huîtres 110, moules 125, saint-jacques 150,
  gambas 170. Déblocages : 15 000 → 250 000 €. Mesure (sa sauvegarde, sans points) : 6 €/s au début, 26 €/s avec
  3 poissons, 2 bateaux chacun et 4 ouvriers, 71 €/s tout acheté (sa ferme : 30, crèmerie 71, exotique 46).
  Le marin commande au port ; son cadeau (bateaux ×2) marche.
- Zone du joueur : six rangées de téléporteurs (11 destinations avec le port) ; améliorations, commandes et
  vestiaire un peu plus bas. Zone horizontale (supermarché) : trois rangées de cinq.
- Port : « Paniers + » (clé port.entrepot, comme l'entrepôt ailleurs), dans la criée à droite de « Étals + » :
  40 → 100 produits par panier en 12 niveaux (6 000 € puis ×1,6).
- Port : au départ, 1 poisson par voyage (+1 tous les 3 points « Bateaux : soute ») et 1 coquillage par passage
  de drague (+1 tous les 4 points « Dragues : quantité ») : les points de l'école sont indispensables.
- Favoris (Game.Fav, sauvegardé : Favoris) : les touches 1 à 9 et 0 vont aux lieux choisis. Sur la carte, la main
  (ou la souris) sur un lieu + un chiffre : il prend cette touche (il quitte l'ancienne ; l'ancien occupant n'a plus
  de touche). Au départ : les touches d'avant (DefaultFav). La touche s'affiche sur la vignette de la carte et sur
  les téléporteurs (rien si le lieu n'est pas un favori). Tab : la carte.
- Zone du joueur plus large (YardW 830, elle grandit vers la gauche : les lieux donnent toujours l'ancien bord
  gauche, sauf le port qui la décale) : 32 téléporteurs (4 colonnes × 8 rangées) ; améliorations à y 912 ;
  commandes en bas à gauche, vestiaire en bas à droite. Zone horizontale (supermarché) : 2 150 de large, vers la
  gauche, 30 téléporteurs (10 × 3).
- Port : jetée deux fois plus large (y 520 à 1040, 520 de large) : paniers et cases au bord, une allée libre au
  milieu ; bateaux et parcs suivent (PortBerthY, PortParcY dépendent de la jetée). Ouvriers du port en
  poissonniers (marinière, grand tablier ciré bleu, bottes jaunes, bonnet ; look.mariniere, look.boots) ;
  caissière en marinière.
- Cadeau du maire : il marchait (PriceOf) mais se voyait mal ; sur la carte, bandeau rouge « Cadeau du maire …
  Encore m:ss » et « (−50 %) » après le prix des lieux à vendre.
- Règle des employés : autant d'employés que de lieux exploitables de la zone (champs, machines, bâtiments
  d'animaux, parcs). Ferme : maïs 3 (champ, moulin, poulailler), blé 5 (champ, moulin, étable, chèvrerie,
  bergerie), serre 6, verger 6, fleurs 2 (champ, ruches). Exotique : fruits 4, sucre 3, cacao 2, café 3, riz 1,
  épices 4. Élevage : volaille 4, bovins 3, ovins 3, porcs 2. Boucherie : volaille 6, bœuf 4, agneau 2,
  charcuterie 4. Boulangerie 6, crèmerie 9, glacier 6 (autant que de machines). Port 8.
  Pas appliquée : supermarché (4 magasiniers par espace, pas de ressources) et école (2 surveillants).
  Une sauvegarde avec plus d'employés que le maximum en garde le maximum (blé : 6 → 5).
- Supermarché : un magasinier par rayon (Max = nombre de rayons de l'espace ; coûts ×3 jusqu'au 4e, puis ×1,15).
- Points des collégiens (type « marketing ») : « Clients » (pts_marketing) et « Patience » (pts_patience) ne
  s'achètent plus (les points déjà placés comptent toujours). À la place :
  Panier moyen (pts_panier, lieux qui vendent) : +1 % par point ; chaque article d'un client : +1 produit tous
  les 100 points, le reste en chance ; 10 objets au plus, +1 tous les 10 points.
  Clients fidèles (pts_fidele) : gains d'absence × (1 + moyenne des points des lieux qui vendent / 100).
  Réclame (pts_reclame, lieux à commandes) : durée du cadeau × (1 + pts/100), fixée à l'annonce (Order.Dur) ;
  si la personne tirée est un particulier, chance min(50 %, pts/200) de tirer à la place une personne spéciale.
  Badges « Bien en vue » et « Star de la pub » : points de collégiens (anciens et nouveaux) au lieu du marketing.
- Points « Soigneurs : Sélection génétique » (pts_soin_rent, ferme et ferme d'élevage) : les animaux qui mangent
  (poulailler, étable, chèvrerie, bergerie, ruches ; mare, prés, porcherie) consomment moins, comme la rentabilité
  des machines (RentSave : −50 % à 100 points, 75 % au plus). À ajouter dans tout futur lieu où des animaux mangent.
- Mairie, cartes des points : leur hauteur s'adapte au nombre de cartes (tout tient au-dessus de Valider/Annuler ;
  textes et boutons +1/+10 réduits d'autant, police à 80 % au moins).
- Serre : plus de graines. Les légumes repoussent tout seuls après la récolte (comme le maïs, le blé, les fruits) ;
  plus de caisses de sachets. Station.Serre marque les colonnes de la serre (dessin, achat de case) ; Replant et
  le code des graines restent mais ne servent plus (anciens sachets : les ouvriers les jettent). Badge « Jardinier » :
  ouvrir les six colonnes de la serre (ou 100 sachets déjà plantés).
- Livraison express : c'est l'effet du cadeau qui double, pas sa durée (Boost.Mult 4 au lieu de 2) : ventes ×4,
  lieux au quart du prix, école/camions et drones/animaux/bateaux ×4, grand embellissement (fleurs plus grandes,
  deux fois plus de pétales), 2 points université. La commande l'annonce : « Livrée avant m:ss : ventes ×4 … ! ».
- Correction : les rayons « Poissonnerie » du supermarché restaient fermés (« D'abord à la ferme ») : produced()
  ne connaissait pas le port ; et sans rayon au supermarché, les drones du port ne s'affichaient pas. Le port
  compte maintenant (étal construit au port → rayon achetable), et un rayon fermé dit « D'abord au port ».
- Boulangerie, recettes : brioche = farine, œufs, beurre, sucre en poudre ; croissant = farine, œufs, lait, beurre,
  sucre en poudre ; pain au chocolat = farine, œufs, beurre, chocolat. Brioches et croissants demandent donc le
  beurre (crèmerie) et le moulin à sucre (ferme exotique). La réserve reçoit aussi le sucre en poudre (camion de
  la ferme exotique).
- Commande de l'école : tous les adultes (profs, surveillants, secrétaire) veulent du pain blé-maïs (plus de
  baguette de maïs). ⚠️ Règle : une commande peut demander un produit pas encore fabriqué (c'est voulu : ça
  pousse à ouvrir la nouvelle marchandise) ; ne jamais remplacer par un produit disponible.
- Ateliers : une machine à plusieurs ingrédients est plus large (260 + 70 par ingrédient au-delà de 2), sa zone
  de dépôt aussi (jeu aéré). Réserve : palettes resserrées s'il y en a beaucoup.
- Supermarché : caisses plus bas (SuperRegY 1390), rangées de rayons un peu plus serrées (220), espaces et
  barrières jusqu'à 1170 (SuperBarY1), bas du magasin à 1604 (SuperFloorY1) : un couloir derrière les caissières
  permet de passer d'un espace ouvert à l'autre même quand celui du milieu est fermé. L'entrepôt est interdit aux
  clients (Place.NoCust ; grille de chemins à part pour les clients : gridFor(true), routeC, clearC).
- Camions : le camion du port porte « LE PORT » (avant : « MA FERME » par défaut, d'où l'impression qu'il
  s'arrêtait au mauvais endroit ; il s'arrête bien au-dessus de la poissonnerie).
- Glacier (7 recettes, 3 300 de large, 11 palettes) : sorbet fraise, sorbet mangue (15 000 €), sorbet ananas
  (25 000), glace cerise (35 000), glace noix de coco (50 000), glace chocolat (65 000), glace poire-chocolat
  (90 000). Glaces : lait, crème, sucre en poudre, œufs + le parfum. Nouveaux produits SorbetMangue, SorbetAnanas,
  GlaceCoco (sorbet cerise et glace fraise ne se font plus).
- ⚠️ Règle : jamais de machine derrière la caissière (au-dessus de la caisse, à gauche des vitrines) : la 2e rangée
  de machines est décalée d'un cran vers la droite (shift2) ; glacier corrigé.
- Supermarché : chaque espace a assez de colonnes pour tous ses rayons (3 par colonne, calculé) ; colonnes de
  320 (un rayon de 180 : grand passage entre voisins, marges égales) ; poissonnerie avant la boucherie.
- Glacier : machines écartées (colW 720), rangées plus hautes (row0 440, écart 300 : passage avant les vitrines) ;
  quai juste après la dernière vitrine, compost à droite de la 4e machine (glace cerise) ; toutes les fabriques de
  la même couleur (la glace chocolat était bleue : reste d'une ancienne couleur).
- Ferme : grenier à foin au milieu du champ de blé (au-dessus), grange décalée à droite (la case « Stock + »
  ne la touche plus).
- Ferme d'élevage : « Frigos + » dans la zone du joueur après les éleveurs ; « Entrepôt » à gauche de la réserve ;
  compost à droite de la porcherie ; plus de conseil « Embauche une caissière » là où il n'y a pas de caisse ;
  route du haut (camions de ma ferme) ; quai du camion frigorifique entre les frigos bovins et ovins (x 1420).
- Ateliers (boulangerie, crèmerie, glacier, boucherie) : « Entrepôt » à gauche de la réserve, compost bien dans
  le bâtiment. Boucherie : le carrelage partout (plus de sol coloré par espace, seulement leur nom).
- Points du lycée : « Négociation » ne s'achète plus (points déjà placés : comptent toujours). Mairie : la liste
  des lieux défile (molette ou boutons Haut / Bas) quand elle est trop longue.
- Camions : « Camions : charge » renforcé : +2 produits par voyage par point (avant +1 tous les 2) et quais
  +1 place tous les 2 points (avant tous les 10). Mesure sur sa sauvegarde (8 min) : le quai de la ferme était
  plein (32 par produit) — les camions limitaient bien. Mais l'essentiel des rayons vides venait des magasiniers
  (rayons vides avec palette pleine : 9 en moyenne ; avec un magasinier par rayon : 1) et, pour le reste, de la
  production (œufs, choux, pommes de terre, beurre, fromage blanc, yaourts, fruits exotiques : quai vide).
- Rayons vides du supermarché (sim/rayons.go) : sa sauvegarde du 11/11 (7 000 €/s) avait ~20 rayons presque
  vides, entrepôt vide, magasiniers inoccupés. Les maillons qui bloquaient, lieu par lieu : boulangerie et glacier
  = drones (charge 3 et 1) ; viennoiseries = beurre de la crèmerie ; glaces = crème ; sorbets mangue/ananas = fruits
  de l'exotique ; exotique = camion (100 par voyage pour 13 produits) puis drones (thé, cannelle : 60 s de trajet) ;
  crèmerie = production. Elle refuse un camion « par produit » : un lieu riche demande plus de mécaniciens, voulu.
  Ajouté : sous le quart de sa contenance, un rayon montre une pastille (magasinier bleu, camion orange, drone
  violet, usine rouge + l'ingrédient qui manque, cadenas gris), selon la cause la plus fréquente des ~30 dernières
  secondes ; devant le rayon, une bulle explique et dit quoi faire (taille constante à l'écran) ; le message du bas
  résume les deux causes principales (après les conseils urgents, avant les achats possibles).

## 4.1 — prix selon la carte, ateliers aérés (12 novembre)
- Constat sur sa sauvegarde (8 500 €/s, 10 lieux) : le supermarché fait 96 % des revenus (8 111 €/s ; port 80,
  crèmerie 61, exotique 48, glacier 48, boulangerie 40, ferme 31, boucherie 10). Les lieux tardifs étaient presque
  gratuits : ferme exotique 35 000 € et tout son contenu ≈ 1,9 M (≈ 3 min de revenus) ; le bot humain achetait
  exotique, boulangerie et glacier en 20 min.
- ⚠️ Règle : les prix dépendent de la POSITION DU LIEU SUR LA CARTE (sim.MapOrder, = ordre des tuiles de drawMap),
  pas de l'ordre d'achat : il faut inciter à ouvrir les lieux dans l'ordre de la carte. Un nouveau lieu se met dans
  MapOrder, à sa place. g.Places est trié dans cet ordre (points d'école, listes…).
- sim/prices.go : les 4 premiers lieux (ferme, école, supermarché, crèmerie) gardent leurs prix (début rapide). Dès
  le 5e (ferme exotique), chaque lieu coûte 2 × plus que le précédent : achat 500 000 € (exotique), 1 M, 2 M, 4 M,
  8 M, 16 M (boucherie) ; contenu complet = contenu de la crèmerie (≈ 16 M) × 2^(rang − 4). Place.Scale = cible /
  contenu de base : tous les prix du lieu (déblocages, améliorations au sol, stocks, drones, employés, pousses) sont
  multipliés par ce même coefficient (exotique ×17, boulangerie ×15, glacier ×13, port ×45, élevage ×144, boucherie
  ×43), arrondis à 3 chiffres. Les rayons du supermarché gardent leurs prix (calculés avant).
  Choix de la conceptrice : ×2 à chaque lieu, POUR TOUJOURS ; pour que ça tienne jusqu'au 20e lieu, les futurs lieux
  devront vendre plus cher (chaque nouveau lieu doit à peu près doubler les revenus). Elle prévoit aussi des lieux
  qui n'envoient rien au supermarché (pour qu'il ne soit plus le centre de tout).
- Sauvegarde : les montants payés (déblocages, cases au sol, stocks, pousses) sont écrits SANS le coefficient du lieu
  et remultipliés au chargement : une ancienne sauvegarde garde tout ce qui était payé (pas de SaveVersion changée).
- Supermarché : 1 objet livré = 3 en réserve (avant 2).
- Ateliers (atelier.go) : chaque machine est placée selon sa vraie largeur + sa case « Stock + » (stockBoxW 118) +
  une allée (machineGap 240 ; glacier 320) ; le lieu s'élargit s'il le faut (boulangerie 3 140, crèmerie 3 620,
  glacier 4 020). Rangées à 450 et 800, vitrines plus bas (shelfY + 120) : un couloir de ~100 sous la 2e rangée
  (on passe sans marcher dans une zone). Bases des drones plus près des vitrines (Station.PadOff 85 : le mur du
  bas est à 1 514). Compost contre le mur de droite, à hauteur de la 1re rangée (crèmerie : il n'est plus dans
  l'allée ; glacier : il ne chevauche plus le « Stock + » de la glace cerise). Quais juste après la dernière vitrine
  (boulangerie : à droite des pains au chocolat). Boucherie : mêmes rangées et vitrines.
- Ferme d'élevage : cases « Stock + » pour les bâtiments d'animaux (poulailler, mare, prés, porcherie ; pas les
  abattoirs) ; le lieu s'élargit (3 500) et tout est espacé (espaces 150–1110, 1160–1810, 1860–2510, 2560–3010 ;
  quai à 1680 ; compost à 3130).
- Boucherie : « Il faut les bovins » alors qu'ils étaient achetés : needsMet cherchait une station « bovin » (le pré
  s'appelle pre_bovin) ; il accepte maintenant aussi l'identifiant d'un déblocage payé (builtOrUnlocked).
- Ferme : un 4e ouvrier du maïs (95 000 €) et un 7e ouvrier de la serre (120 000 €, quand les 6 colonnes sont
  ouvertes), en fin d'ordre d'embauche (demandés par la conceptrice, au-delà de la règle « un par ressource »).
- Captures d'une vraie sauvegarde (shots_save.go) : `CAPSAVE=sauvegarde.json ./jeu -captures dossier`
  (CAPPLACES=boulangerie,glacier pour quelques lieux ; CAPSEC, CAPZOOM). Sous Linux : paquets libasound2-dev,
  libgl1-mesa-dev, libx*-dev, xvfb-run, et un ~/.asoundrc « pcm.!default { type null } ».
- Glacier (8 recettes) : la glace à la cerise devient la glace à la vanille (même place, 35 000 € de base ;
  produit GlaceVanille, renommé « Glace à la vanille », 160 €) ; nouvelle glace au café juste après (GlaceCafe,
  ajoutée en fin de liste des produits, 165 € ; 42 000 € de base). Les glaces n'ont plus ni lait ni sucre :
  vanille = crème, œufs, vanille ; café = crème, œufs, café moulu ; coco = crème, œufs, noix de coco ;
  chocolat = crème, œufs, chocolat ; poire-chocolat = crème, œufs, poires, chocolat. Réserve du glacier : fraises,
  mangues, ananas, vanille, café moulu, noix de coco, crème, œufs, chocolat, poires (plus de lait, sucre, cerises).
  8 serveurs (un par fabrique : 1,5 M et 3,2 M de base en plus). Supermarché : rayons Glace à la vanille et
  Glace au café. Sauvegarde : ce qui était payé pour la glace à la cerise (fabrique, stock, drone, rayon du
  supermarché) passe à la vanille (migrateGlaces, au chargement).
- Commandes : le garagiste (cadeau camions et drones ×2) ne commande pas tant que le supermarché n'est pas acheté
  (avant, ni camion ni drone : son cadeau ne servait à rien).
- Caisses : le joueur à moins de 170 d'une caissière (HelpDist, Place.PlayerHelps) l'aide : leurs vitesses
  s'additionnent (petite boutique : 1,2 s → 0,35 s par article ; supermarché : l'aide du joueur s'ajoute aux
  points « Caisses »). Bulle « Tu l'aides : ça va plus vite ! » au-dessus de la caissière. Pas à l'accueil de l'école.

## 4.1 — le domaine (11e lieu), drones en plus, silo à blé (13 novembre)
- LE DOMAINE (sim/domaine.go, domaine_draw.go), à la place de « Caviste » sur la carte (après la boucherie, rang 11 :
  32 M à l'achat, contenu ≈ 2 milliards par la courbe des prix ; coefficient ≈ ×330). Lieu « hybride » d'après son
  plan : au nord le houblon (gauche) et la vigne (droite) ; sous la vigne le stockage du raisin (réserve sans clients)
  et la machine à jus de raisin ; un chemin ; les fabriques : brasserie, presse à vin blanc, presse à vin rouge (une
  bassine, la presse et un foudre ; le rouge fermente plus longtemps : 10 s contre 6) ; à droite la distillerie en
  colonne (whisky, vodka, alcool de poire, cidrerie, pressoir à pommes), chacune en face de sa palette ; la réserve
  (palettes blé, pommes de terre, poires, pommes) contre une ROUTE VERTICALE des camions au bord droit (choix de la
  conceptrice : chaque lieu doit avoir sa particularité) ; les quais (supermarché, crêperie, zoo) au nord de la
  réserve ; le camion de ma ferme monte la route jusqu'à la réserve (backTruckV, drawDockV).
  Boutique : caisse + étals vin blanc, vin rouge, bière, cidre, alcool de poire, vodka, whisky ; réserves sans clients
  pour le jus de raisin et le jus de pomme (drones → quais). Le raisin ne se vend qu'au supermarché (rayon Fruits).
  Quais crêperie (cidre) et zoo (jus) : prêts, invisibles tant que ces lieux n'existent pas.
  Rien n'est offert : vigne, jus de raisin, vin blanc, vin rouge, houblon, brasserie, palette de pommes (il faut le
  pommier), pressoir, cidrerie, palette de poires (poirier), alambic, palette de pommes de terre (serre), vodka,
  palette de blé (silo à blé de ma ferme), whisky. Chaque machine de la distillerie attend sa palette.
  Équipes : vigne 4, houblon 2, distillerie 5 (embauches dans la zone du joueur). Points : agronomes, mécaniciens
  (machines, drones, camions), entraînement, caisses, panier, fidèles, réclame, comptables, logistique.
  Commandes : des particuliers (comme à la ferme). Zone du joueur : « L'alcool est dangereux pour la santé.
  À consommer avec modération. » (au-dessus du vestiaire).
  Prix de vente (le domaine doit doubler les revenus) : raisin 60, jus de raisin 120, jus de pomme 130, vin blanc 350,
  vin rouge 450, bière 500 (après les vins, à sa demande), cidre 550, alcool de poire 650, vodka 750, whisky 1 000.
  Mesure (tout débloqué, champs pleins, palettes pleines, sans points, 15 min) : tout est fabriqué, ≈ 300 €/s sur
  place, le reste part au supermarché.
- Supermarché : l'espace « Alcools » devient « Boissons » (jus, vins, bière, cidre, alcools ; ouvert avec le domaine) ;
  le raisin au rayon Fruits. Les rayons gardent les prix du supermarché (9 000 € + la moitié du déblocage de base).
- Ferme : silo à blé (30 000 €, à gauche du grenier à foin, caché tant que le domaine n'est pas acheté) : les ouvriers
  du blé y rangent du blé, un drone de fret l'emporte au quai « → domaine » (nouveau quai de la ferme).
- Palettes : une palette pas encore construite n'existe pas pour les camions (pallet() : Built).
- Ferme : un 2e puis un 3e drone par étal (« Drone +1 », 3 puis 9 fois le prix du 1er) quand le produit part vers au
  moins 2 lieux (DroneMax, destCount). Station.Drones (liste) ; les drones d'un étal se partagent le stock (rien de
  pris deux fois) et vont chacun au quai qui en manque le plus. Mesure sur sa partie à 10 lieux : la plupart des
  drones volent 20 à 50 % du temps ; fraises (61 %, quai en manque 27 % du temps), pommes de terre (53 %, 20 %) et
  œufs (29 %, 12 %) saturaient. Elle craint que la production de pommes de terre ne suive pas : à surveiller.
- Ferme : la case d'embauche des ouvriers de la serre (et du verger) est une grande case (300 × 92) au milieu,
  sous les six colonnes, dans l'herbe entre les deux chemins (y 684 ; pas sur le chemin, à sa demande ; avant :
  petite case sous la 1re colonne) : le texte se lit en entier et on
  comprend que l'ouvrier est pour toute la serre, tout le verger.

## 4.1 — retouches (salade, quais, parking, aide…)
- La salade remplace le chou (nom « Salades », dessin de feuilles frisées) ; l'identifiant reste Chou / « chou »
  (sauvegardes, rayon du supermarché, drones…).
- Supermarché : le parking au sol est décalé de 90 vers la droite (SuperParkX = SuperCartX + 510) : la route
  d'accès (jusqu'à SuperCartX + 480) ne passe plus dessus ; ses allées partent de la route d'accès.
- Caissière sans client depuis 5 s (CashierBoredAfter ; Agent.Bored compté dans Place.step) : elle sifflote (notes
  de musique, comme les ouvriers qui s'ennuient : drawNotes).
- Ferme, quais (x définitifs, non décalés) : supermarché 1 700, boulangerie 2 050, crèmerie 2 892 (sous le lait de
  chèvre, au milieu des laits), ferme d'élevage 3 250, domaine 3 600, glacier 4 380 (juste à gauche des fruits).
- Supermarché, rayon d'un lieu pas encore ouvert : « Depuis la ferme exotique » (et « Depuis la crèmerie »,
  « Depuis le glacier », « Depuis le port », « Depuis le domaine »…) au lieu de « D'abord à… ».
- L'aide s'ouvre avec la touche « , ? » (avec ou sans majuscule), dans un lieu ou sur la carte. Les touches de
  caractères se lisent maintenant par le caractère tapé (AppendInputChars) : sur son clavier AZERTY, la touche
  « , ? » est à la place du M anglais (avant, elle coupait le son) ; M coupe le son (le caractère m).
- Ruches : stock plus cher (BaseCost 12 000 : 6 000 € puis ×2,2, après la bergerie ; avant 60 €).
- Ferme exotique : les noms des espaces (sucre, cacao, café…) au milieu de leurs champs ; la case « Ouvrier
  Fruits +1 » est une grande case dans l'herbe, au milieu sous les quatre arbres (comme la serre et le verger).
- Ressenti de la conceptrice (nouvelle partie, 4 h 55 de jeu, 1 375 €/s, ferme exotique en cours) : le délai pour
  atteindre le lieu suivant lui plaît ; elle doit bien choisir son prochain achat (c'est le but). Seule déception :
  le stock des ruches trop bon marché (corrigé). Garder ce rythme comme référence.
- Drones, dans TOUS les lieux (plus seulement la ferme) : un drone par lieu acheté qui demande le produit de l'étal
  (DroneMax = destCount : deux lieux → deux drones, trois → trois…), au plus un par quai du lieu ; chaque drone
  de plus coûte 3 fois le précédent. Le drone suivant ne s'achète que quand la demande augmente (nouveau lieu
  acheté, rayon ouvert au supermarché…). Sa partie : seuls les trois laits (supermarché + crèmerie) ont 2 lieux.
- Raffinerie « lente » (sa partie : 36 points de vitesse, ×2,08 → 1,44 s par sucre) : elle travaille 100 % du temps,
  la canne s'entasse devant : c'est le goulot de l'espace sucre. Décision de la conceptrice : ne rien changer, il
  faut mettre plus de points « Machines : vitesse » (pas de 2 sucres par canne, pas de temps de base plus court).
- Supermarché, choix de la caisse (superRegister) : le client estime le TEMPS d'attente (articles de tous les
  caddies devant lui × vitesse de la caisse) + un quart du temps de marche ; caisse sans caissière : +60 s (évitée,
  mais celui qui y va y reste : voulu). Avant : nombre de personnes + distance/700, d'où des files de 5 au milieu et
  de 1 aux bouts. Mesure (sa partie, 9 caisses, 10 min) : files moyennes 0,6 à 5,4 → 1,7 à 2,9.
- Boulangerie : pain de mie (PainMie, ajouté en fin de liste des produits, 32 €), juste après la baguette : farine de
  blé, 3,5 s, 1 500 € de base. 7 machines → 7 boulangers (800 000 € de base en plus). Supermarché : rayon Pain de mie
  (espace Boulangerie). Le contenu de la boulangerie grandit : son coefficient de prix baisse un peu (×15 → ×12).
- Ferme exotique : la case « Ouvrier Épices +1 » est une grande case dans l'herbe, au milieu de l'espace : l'ouvrier
  est pour toutes les épices, pas seulement la vanille. Le séchoir doit rester sous le thé (sa demande) : les champs
  sont dans l'ordre vanille, cannelliers, théiers (le thé en dernier, son séchoir dessous à droite, x 5 310) ; la
  case est sous les cannelliers, au milieu. Étals dans le même ordre. L'ordre d'achat ne change pas.
- Boulangerie : « Four à baguettes » s'appelle « Tradition », « Four à maïs » s'appelle « Baguette maïs »
  (identifiants inchangés).
- Ferme exotique, ordre d'achat des épices = ordre des champs : vanille 80 000, cannelliers 95 000 (avant 120 000),
  théiers 105 000 (avant 95 000), séchoir 120 000 (avant 105 000), prix de base. Sauvegarde : thé et séchoir
  payés en entier à l'ancien prix restent payés (migrateEpices).
- Bases des drones centrées sous leur étal, quelle que soit sa largeur (DronePad) : les étals de la ferme
  exotique (126, 220, 360 de large) avaient leur « H » décalé à gauche. Ferme comprise : son portail est déplacé
  entre la caisse et l'étal du maïs (FarmGateX 1 070 ; avant 1 250, où tombait la base du drone du maïs) ; le
  passage piéton et le parking le suivent (ParkX d'après Gate.X), clôture et panneau aussi.

## Page publique (GitHub)
- README.md bilingue (français puis anglais) : présentation, téléchargement, commandes, captures (docs/images),
  avis (formulaires d'issues), Tipeee (https://fr.tipeee.com/nathalie314/), licence, notes pour développeurs.
  Autrice affichée : Nathalie314.
- Licence : CC BY-NC-SA 4.0 (son souhait : qu'on puisse modifier et utiliser le jeu, mais pas se faire de l'argent
  dessus) ; les bibliothèques (Ebitengine, polices Go) gardent la leur. Fichier LICENSE.
- .github/workflows/release.yml : à chaque push sur main (ou à la main, onglet Actions), tests de sim, compilation
  Windows, ToutUneVille.zip publié dans une Release « version-N » marquée « latest ». Lien permanent :
  https://github.com/Nathalie314/Toute-une-ville/releases/latest/download/ToutUneVille.zip
- .github/ISSUE_TEMPLATE : « Signaler un bug » (bug.yml), « Proposer une idée » (idee.yml), lien Tipeee (config.yml).

## La pâtisserie (12e lieu)
- Atelier (patisserieSpec, newWorkshop), après le domaine sur la carte : 64 M à l'achat, contenu ≈ 4 milliards (×2).
  Particularité : chaque recette demande au moins 3 ingrédients. Rien n'est offert (même la 1re machine s'achète).
  Tartes aux pommes (pommes, beurre, farine ; 5 000 de base), riz au lait (riz, lait, vanille ; 9 000), clafoutis
  aux cerises (cerises, farine, œufs, lait ; 15 000), gâteaux au chocolat (chocolat, beurre, farine, œufs ; 25 000 ;
  produit Gateau qui existait sans être fabriqué), éclairs au café (farine, œufs, lait, beurre, café moulu ; 40 000),
  fraisiers (farine, œufs, lait, beurre, fraises ; 60 000). Prix de vente : 1 100, 1 250, 1 400, 1 600, 1 800, 2 000 €.
  Réserve de 11 palettes ; quais « → pâtisserie » à ma ferme (x 4 750, sous les fruits), à la crèmerie (beurre) et à la
  ferme exotique (riz, vanille, chocolat, café moulu). 6 pâtissiers, points de chefs, commandes de particuliers.
  Supermarché : l'espace « Pâtisserie » s'ouvre. Décor rose, tenues blanches à tablier rose et toque.
- Correction de NOTES : la cantine de l'école ne demande PLUS de pain d'épices ni de gâteau aux fruits (ces produits
  ont disparu du jeu) : primaire baguette, collège brioche, lycée croissant, université pain au chocolat, adultes
  pain blé-maïs (schoolOrder).
- Ateliers : le nom du lieu (panneau du bas) descend sur le mur (les bases des drones sont juste au-dessus).
- Drones : la description du premier drone ne dit plus « un drone de plus » (Desc reçoit le niveau à acheter).
- Ateliers : les derniers carreaux du sol sont coupés au bord (ils débordaient sur l'herbe, à droite et en bas).
  Pâtisserie : son quai « → supermarché » est juste à droite de la vitrine des fraisiers (x 1 610), sur la route.
- Badges pour les lieux qui n'en avaient pas : ferme exotique (9), ferme d'élevage (5), boucherie (8), port (8),
  domaine (8), pâtisserie (8), soit 126 badges. Base commune (placeBadges) : acheter le lieu, tout débloquer,
  toute l'équipe, premier client, N clients servis, 5 étoiles ; plus 2 ou 3 badges de fabrication par lieu.
  Sections Port, Domaine, Pâtisserie dans la page des badges. Glacier : « Cent boules » compte les glaces à la
  vanille (la glace nature n'existe plus), « Tous les parfums » = toutes les fabriques (il y en a 8).
  ⚠️ Règle (CLAUDE.md) : chaque nouveau lieu reçoit ses badges. Elle n'y tient pas, mais des joueurs adorent.

## Version navigateur (pour son iPad)
- Même code, compilé en WebAssembly (GOOS=js GOARCH=wasm) ; l'exe Windows reste identique. Les deux sont fabriqués
  à chaque push sur main (release.yml) : le .zip dans la Release, puis une 2e tâche « web » met la version web sur GitHub Pages
  (https://nathalie314.github.io/Toute-une-ville/ ; réglage à faire une fois : Settings → Pages → Source « GitHub Actions »).
- Sauvegarde : fichier JSON sous Windows (save_file.go), localStorage du navigateur sur le web (save_web.go,
  clé « ToutUneVille_sauvegarde »), avec une sauvegarde de plus quand on quitte ou cache la page (pagehide).
  Les deux parties sont séparées.
- Tactile (pointer.go) : un doigt posé = clic maintenu (on avance vers lui, on appuie sur les boutons) ; deux
  doigts = pincer pour zoomer. Pas de bouton « Quitter » sur le web ; la ligne d'aide du bas parle du doigt.
- web/index.html : écran de chargement vert avec barre de progression (≈ 19 Mo), plein écran sur iPad avec
  « Sur l'écran d'accueil » (icône web/icone.png).
