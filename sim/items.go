package sim

// Item est une marchandise que l'on peut porter, transformer ou vendre.
// (L'ordre ne doit pas changer : les sauvegardes s'en servent. On ajoute à la fin.)
type Item int

const (
	Ble Item = iota
	Mais
	Farine
	Pain
	Oeuf
	Legume
	Lait
	Fleur
	Miel
	Gateau
	Canne
	Fromage
	Saucisse
	FarineMais
	Fruit
	Creme
	Sucre
	Cacao
	Chocolat
	BaguetteMais
	PainMele
	Brioche
	PainEpices
	Tarte
	GateauFruits
	GlaceVanille
	GlaceFruits
	GlaceChoco
	Crayons // fournitures de l'école (ne se vendent pas)
	Cahiers
	Geometrie
	Atlas
	Calculatrice
	Microscope
	Ordinateur
	Blouse
	GlaceFruitLait // (ajoutée à la fin : les sauvegardes rangent les produits par numéro)
	// version 4.0
	LaitChevre
	LaitBrebis
	CuisseCanard
	CafeVert
	CafeGrains
	CafeMoulu
	Beurre
	FromageBlanc
	FromageChevre
	FromageBrebis
	Yaourt
	YaourtChevre
	YaourtBrebis
	Croissant
	PainChoco
	// version 4.1 : la serre, le verger, le sucre en deux étapes, les glaces aux fruits
	Carotte
	Tomate
	Courgette
	Poireau
	Chou
	PommeDeTerre
	Fraise
	Pomme
	Poire
	Cerise
	Abricot
	Figue
	GrainesCarotte // sachets de graines (gratuits, ne se vendent pas)
	GrainesTomate
	GrainesCourgette
	GrainesPoireau
	GrainesChou
	PlantsPdt
	SucreRoux
	SucreMorceaux
	SucrePoudre
	SorbetFraise
	SorbetCerise
	GlaceFraise
	GlaceCerise
	GlacePoireChoco
	// étape 6 : la ferme d'élevage et la boucherie-charcuterie
	Foin
	Poulet // vivants (ferme d'élevage)
	Canard
	Vache
	Veau
	Mouton
	Agneau
	Cochon
	PouletAbattu // carcasses (frigos de la ferme d'élevage)
	CanardAbattu
	Boeuf
	VeauCarcasse
	MoutonCarcasse
	AgneauCarcasse
	Porc
	PouletCru // boucherie-charcuterie
	PouletRoti
	CuissePoulet
	Aiguillettes
	Magret
	FoieGras
	Steak
	SteakHache
	RotiBoeuf
	MorceauxVeau
	Gigot
	CoteAgneau
	Jambon
	Cotelette
	Pate
	// la ferme exotique
	Ananas
	Banane
	NoixCoco
	Mangue
	Riz
	Vanille
	FeuillesThe
	The
	Cannelle
	// le port
	Sardine
	Cabillaud
	Thon
	Saumon
	Huitre
	Moule
	SaintJacques
	Gambas
	// le glacier (nouvelles recettes)
	SorbetMangue
	SorbetAnanas
	GlaceCoco
	GlaceCafe // (ajoutée à la fin : les sauvegardes rangent les produits par numéro)
	// le domaine (vigne, houblon, distillerie)
	Houblon
	Raisin
	JusRaisin
	VinBlanc
	VinRouge
	Biere
	JusPomme
	Cidre
	AlcoolPoire
	Vodka
	Whisky
	PainMie // (boulangerie, ajouté à la fin : les sauvegardes rangent les produits par numéro)
	// la pâtisserie (le gâteau au chocolat existait déjà : Gateau)
	TartePommes
	RizLait
	Clafoutis
	Eclair
	Fraisier
	// la ferme exotique : deux arbres de plus (après le manguier)
	Citron
	Orange
	Limonade  // (pressoir à citrons, ferme exotique)
	JusOrange // (pressoir à oranges, ferme exotique)
	// produits d'un futur lieu (le supermarché leur garde déjà un rayon)
	Noisette
	Noix
	Olive
	// la crêperie : la crêpe nature (faite sans arrêt), puis les crêpes garnies (sur commande)
	CrepeNature
	CrepeSucre
	CrepeBananeChoco
	CrepePommeCannelle
	CrepePoireVanille
	CrepeSaucisseOeuf
	CrepeThonPoireau
	CrepeCanard
	CrepeSteakBrebis
	PilonPoulet // (boucherie, après les cuisses)
	Avocat      // (ferme exotique, après les mangues)
	TarteCitron // (pâtisserie, entre les éclairs et les fraisiers)
	Concombre   // (serre de ma ferme, 7e colonne)
	// l'huilerie : trois graines (non vendables) et six huiles
	Tournesol
	Colza
	Lin
	HuileTournesol
	HuileColza
	HuileOlive
	HuileLin
	HuileNoisette
	HuileNoix
	Poivre      // (ferme exotique, espace épices)
	Sel         // (port : le marais salant)
	BaiesPoivre // (ferme exotique : récoltées, elles partent au séchoir à poivre)
	// le port, second ponton : coquillages et crustacés
	Bigorneau
	Palourde
	Crabe
	Tourteau
	Araignee
	Langoustine
	Homard
	Langouste
	NItems
)

type ItemInfo struct {
	Name  string
	Price float64
}

// (Des produits ne servent plus — pain d'épices, tarte et gâteau aux fruits, sorbet et glace aux fruits,
// glaces et sorbets cerise et fraise, cuisses de canard, sucre roux, sachets de graines — mais restent dans
// la liste : leur numéro est dans les sauvegardes. Ils ne sont pas traduits dans langues/en.txt.)
var Items = [NItems]ItemInfo{
	Ble:              {"Blé", 3},
	Mais:             {"Maïs", 5},
	Farine:           {"Farine de blé", 10},
	Pain:             {"Baguette", 30},
	Oeuf:             {"Œufs", 12},
	Legume:           {"Légumes", 15},
	Lait:             {"Lait", 30},
	Fleur:            {"Fleurs", 24},
	Miel:             {"Miel", 40},
	Gateau:           {"Gâteau au chocolat", 1600}, // (pâtisserie)
	Canne:            {"Canne à sucre", 20},
	Fromage:          {"Fromage de vache", 55},
	Saucisse:         {"Saucisses", 90},
	FarineMais:       {"Farine de maïs", 12},
	Fruit:            {"Fruits", 22},
	Creme:            {"Crème", 40},
	Sucre:            {"Sucre", 32},
	Cacao:            {"Cacao", 35},
	Chocolat:         {"Chocolat (tablette)", 60},
	BaguetteMais:     {"Baguette de maïs", 34},
	PainMele:         {"Pain blé-maïs", 45},
	Brioche:          {"Brioche", 70},
	PainEpices:       {"Pain d'épices", 80},
	Tarte:            {"Tarte aux fruits", 90},
	GateauFruits:     {"Gâteau aux fruits", 85},
	GlaceVanille:     {"Glace à la vanille", 160},
	GlaceFruits:      {"Sorbet aux fruits", 55},
	GlaceFruitLait:   {"Glace aux fruits", 90},
	GlaceChoco:       {"Glace au chocolat", 180},
	Crayons:          {"Crayons de couleur", 0},
	Cahiers:          {"Cahiers", 0},
	Geometrie:        {"Trousse de géométrie", 0},
	Atlas:            {"Atlas", 0},
	Calculatrice:     {"Calculatrice", 0},
	Microscope:       {"Microscope", 0},
	Ordinateur:       {"Ordinateur portable", 0},
	Blouse:           {"Blouse de labo", 0},
	LaitChevre:       {"Lait de chèvre", 30},
	LaitBrebis:       {"Lait de brebis", 30},
	CuisseCanard:     {"Cuisses de canard", 22},
	CafeVert:         {"Café vert", 45},
	CafeGrains:       {"Café en grains", 130},
	CafeMoulu:        {"Café moulu", 190},
	Beurre:           {"Beurre", 75},
	FromageBlanc:     {"Fromage blanc", 35},
	FromageChevre:    {"Fromage de chèvre", 75},
	FromageBrebis:    {"Fromage de brebis", 85},
	Yaourt:           {"Yaourt de vache", 30},
	YaourtChevre:     {"Yaourt de chèvre", 50},
	YaourtBrebis:     {"Yaourt de brebis", 58},
	Croissant:        {"Croissant", 75},
	PainChoco:        {"Pain au chocolat / chocolatine", 95},
	Carotte:          {"Carottes", 16},
	Tomate:           {"Tomates", 20},
	Courgette:        {"Courgettes", 24},
	Poireau:          {"Poireaux", 28},
	Chou:             {"Salades", 32}, // (Chou : la salade, qui a remplacé le chou)
	PommeDeTerre:     {"Pommes de terre", 36},
	Fraise:           {"Fraises", 26},
	Pomme:            {"Pommes", 30},
	Poire:            {"Poires", 34},
	Cerise:           {"Cerises", 40},
	Abricot:          {"Abricots", 46},
	Figue:            {"Figues", 54},
	GrainesCarotte:   {"Graines de carotte", 0},
	GrainesTomate:    {"Graines de tomate", 0},
	GrainesCourgette: {"Graines de courgette", 0},
	GrainesPoireau:   {"Graines de poireau", 0},
	GrainesChou:      {"Graines de salade", 0},
	PlantsPdt:        {"Plants de pomme de terre", 0},
	SucreRoux:        {"Sucre roux", 30},
	SucreMorceaux:    {"Sucre en morceaux", 45},
	SucrePoudre:      {"Sucre en poudre", 50},
	SorbetFraise:     {"Sorbet à la fraise", 60},
	SorbetCerise:     {"Sorbet à la cerise", 80},
	GlaceFraise:      {"Glace à la fraise", 140},
	GlaceCerise:      {"Glace à la cerise", 160},
	GlacePoireChoco:  {"Glace poire-chocolat", 210},
	Foin:             {"Foin", 0},
	Poulet:           {"Poulet", 0},
	Canard:           {"Canard", 0},
	Vache:            {"Vache", 0},
	Veau:             {"Veau", 0},
	Mouton:           {"Mouton", 0},
	Agneau:           {"Agneau", 0},
	Cochon:           {"Cochon", 0},
	PouletAbattu:     {"Poulet abattu", 0},
	CanardAbattu:     {"Canard abattu", 0},
	Boeuf:            {"Carcasse de bœuf", 0},
	VeauCarcasse:     {"Carcasse de veau", 0},
	MoutonCarcasse:   {"Carcasse de mouton", 0},
	AgneauCarcasse:   {"Carcasse d'agneau", 0},
	Porc:             {"Carcasse de porc", 0},
	PouletCru:        {"Poulet cru entier", 70},
	PouletRoti:       {"Poulet rôti", 110},
	CuissePoulet:     {"Cuisses de poulet", 85},
	PilonPoulet:      {"Pilons de poulet", 80},
	Avocat:           {"Avocats", 64},
	TarteCitron:      {"Tarte au citron meringuée", 1900},
	Concombre:        {"Concombres", 40},
	Tournesol:        {"Graines de tournesol", 12},
	Colza:            {"Graines de colza", 12},
	Lin:              {"Graines de lin", 14},
	HuileTournesol:   {"Huile de tournesol", 400},
	HuileColza:       {"Huile de colza", 450},
	HuileOlive:       {"Huile d'olive", 700},
	HuileLin:         {"Huile de lin", 550},
	HuileNoisette:    {"Huile de noisette", 850},
	HuileNoix:        {"Huile de noix", 900},
	Poivre:           {"Poivre", 110},
	Sel:              {"Sel", 180},
	BaiesPoivre:      {"Baies de poivre", 30},
	Bigorneau:        {"Bigorneaux", 190},
	Palourde:         {"Palourdes", 210},
	Crabe:            {"Crabes", 230},
	Tourteau:         {"Tourteaux", 260},
	Araignee:         {"Araignées de mer", 290},
	Langoustine:      {"Langoustines", 330},
	Homard:           {"Homards", 380},
	Langouste:        {"Langoustes", 430},
	Aiguillettes:     {"Aiguillettes de canard", 120},
	Magret:           {"Magret de canard", 150},
	FoieGras:         {"Foie gras", 280},
	Steak:            {"Steak", 130},
	SteakHache:       {"Steak haché", 150},
	RotiBoeuf:        {"Rôti de bœuf", 170},
	MorceauxVeau:     {"Morceaux de veau", 160},
	Gigot:            {"Gigot de mouton", 180},
	CoteAgneau:       {"Côtes d'agneau", 200},
	Jambon:           {"Jambon", 120},
	Cotelette:        {"Côtelettes de porc", 110},
	Pate:             {"Pâté", 140},
	Ananas:           {"Ananas", 45},
	Banane:           {"Bananes", 32},
	NoixCoco:         {"Noix de coco", 50},
	Mangue:           {"Mangues", 58},
	Citron:           {"Citrons", 70},
	Orange:           {"Oranges", 76},
	Limonade:         {"Limonade", 120},
	JusOrange:        {"Jus d'orange", 130},
	Noisette:         {"Noisettes", 80},
	Noix:             {"Noix", 85},
	Olive:            {"Olives", 75},
	// (la crêperie : une table paie ses crêpes et son cidre en partant ; des plats de restaurant, chers,
	// pour que la courbe des prix tienne : chaque lieu double à peu près les revenus)
	CrepeNature:        {"Crêpe nature", 150},
	CrepeSucre:         {"Crêpe au sucre", 5000},
	CrepeBananeChoco:   {"Crêpe banane-chocolat", 7500},
	CrepePommeCannelle: {"Crêpe pomme-cannelle", 6500},
	CrepePoireVanille:  {"Crêpe poire-glace vanille", 9500},
	CrepeSaucisseOeuf:  {"Crêpe saucisse-œufs", 7000},
	CrepeThonPoireau:   {"Crêpe thon-poireaux", 8000},
	CrepeCanard:        {"Crêpe au canard flambé au whisky", 13000},
	CrepeSteakBrebis:   {"Crêpe steak haché-brebis", 10000},
	Riz:                {"Riz (sachet)", 28},
	Vanille:            {"Gousses de vanille", 120},
	FeuillesThe:        {"Feuilles de thé", 20},
	The:                {"Thé (boîte)", 95},
	Cannelle:           {"Bâtons de cannelle", 75},
	Sardine:            {"Sardines", 30},
	Cabillaud:          {"Cabillaud", 50},
	Thon:               {"Thon", 70},
	Saumon:             {"Saumon", 90},
	Huitre:             {"Huîtres", 110},
	Moule:              {"Moules", 125},
	SaintJacques:       {"Saint-Jacques", 150},
	Gambas:             {"Gambas", 170},
	SorbetMangue:       {"Sorbet à la mangue", 75},
	SorbetAnanas:       {"Sorbet à l'ananas", 85},
	GlaceCoco:          {"Glace à la noix de coco", 170},
	GlaceCafe:          {"Glace au café", 165},
	// (le domaine : des produits chers, pour que les revenus suivent la courbe des prix)
	Houblon:     {"Houblon", 8},
	Raisin:      {"Raisin", 60},
	JusRaisin:   {"Jus de raisin", 120},
	VinBlanc:    {"Vin blanc", 350},
	VinRouge:    {"Vin rouge", 450},
	Biere:       {"Bière", 500},
	JusPomme:    {"Jus de pomme", 130},
	Cidre:       {"Cidre", 550},
	AlcoolPoire: {"Alcool de poire", 650},
	Vodka:       {"Vodka", 750},
	Whisky:      {"Whisky", 1000},
	PainMie:     {"Pain de mie", 32},
	// (la pâtisserie, 12e lieu : plus cher que le domaine, pour que la courbe des prix tienne)
	TartePommes: {"Tarte aux pommes", 1100},
	RizLait:     {"Riz au lait", 1250},
	Clafoutis:   {"Clafoutis aux cerises", 1400},
	Eclair:      {"Éclair au café", 1800},
	Fraisier:    {"Fraisier", 2000},
}

// DomaineSold : ce que vend le domaine sur place (le raisin et les jus ne partent
// qu'en camion : supermarché, zoo) ; DrinksSold : l'espace « Boissons » du supermarché.
var DomaineSold = []Item{VinBlanc, VinRouge, Biere, Cidre, AlcoolPoire, Vodka, Whisky}
var DrinksSold = []Item{JusRaisin, JusPomme, VinBlanc, VinRouge, Biere, Cidre, AlcoolPoire, Vodka, Whisky}

// Legumes, Fruits : la serre et le verger (une colonne chacun), et leurs graines.
var Legumes = []Item{Carotte, Tomate, Courgette, Poireau, Chou, PommeDeTerre, Concombre}
var Fruits = []Item{Fraise, Pomme, Poire, Cerise, Abricot, Figue}
var SeedOf = map[Item]Item{Carotte: GrainesCarotte, Tomate: GrainesTomate, Courgette: GrainesCourgette, Poireau: GrainesPoireau, Chou: GrainesChou, PommeDeTerre: PlantsPdt}

// IsSeed : un sachet de graines (ou des plants de pommes de terre).
func IsSeed(it Item) bool { return it >= GrainesCarotte && it <= PlantsPdt }

// FarmSold : ce que vendent les étals de la ferme, dans l'ordre où on les débloque
// (le blé, les fleurs, la canne, le cacao et le café vert ne se vendent pas tels quels).
// (4.1 : plus de viande ; la canne, le cacao et le café partiront à la ferme exotique)
var FarmSold = append(append(append([]Item{Mais, FarineMais, Oeuf, Farine, Lait, LaitChevre, LaitBrebis}, Legumes...), Fruits...), Miel)

// BakerySold, CremerieSold et IceSold : ce que fabriquent la boulangerie
// (pains et viennoiseries), la crèmerie et le glacier. Les gâteaux, tartes et
// pains d'épices iront à la future pâtisserie.
// PatisserieSold : ce que fabrique la pâtisserie (chaque recette : au moins 3 ingrédients).
var PatisserieSold = []Item{TartePommes, RizLait, Clafoutis, Gateau, Eclair, TarteCitron, Fraisier}

var BakerySold = []Item{Pain, PainMie, BaguetteMais, PainMele, Brioche, Croissant, PainChoco}
var CremerieSold = []Item{Creme, Beurre, FromageBlanc, Yaourt, Fromage, YaourtChevre, FromageChevre, YaourtBrebis, FromageBrebis}
var IceSold = []Item{SorbetFraise, SorbetMangue, SorbetAnanas, GlaceVanille, GlaceCafe, GlaceCoco, GlaceChoco, GlacePoireChoco}

// ExoFruits : les arbres de la ferme exotique (une colonne chacun, comme le verger).
var ExoFruits = []Item{Ananas, Banane, NoixCoco, Mangue, Avocat, Citron, Orange}

// ExoticSold : ce que vendent les étals de la ferme exotique (la canne, le cacao,
// le café vert et les feuilles de thé ne se vendent pas tels quels).
var ExoticSold = []Item{Ananas, Banane, NoixCoco, Mangue, Avocat, Citron, Orange, SucreMorceaux, SucrePoudre, Chocolat, CafeGrains, CafeMoulu, Riz, Vanille, The, Cannelle, Poivre, Limonade, JusOrange}

// EpicerieSold : le rayon « Épicerie » du supermarché (sucre, chocolat, café, riz, épices de
// la ferme exotique ; noisettes et noix de l'huilerie).
var EpicerieSold = []Item{SucreMorceaux, SucrePoudre, Chocolat, CafeGrains, CafeMoulu, Riz, Vanille, The, Cannelle, Poivre, Sel, Noisette, Noix} // (le poivre de la ferme exotique, le sel du port, puis les noisettes et les noix de l'huilerie)

// FutureSold : produits d'un lieu qui n'existe pas encore (leur rayon attend).
var FutureSold = []Item{}

// HuilesSold : l'espace « Huiles » du supermarché (dans l'ordre d'achat à l'huilerie) ;
// HuilerieRaw : ce que l'huilerie envoie tel quel (olives aux fruits, noisettes et noix à l'épicerie).
var HuilesSold = []Item{HuileTournesol, HuileColza, HuileOlive, HuileLin, HuileNoisette, HuileNoix}
var HuilerieRaw = []Item{Olive, Noisette, Noix}

// BoucherieSold : ce que vend la boucherie-charcuterie (par espace).
var BoucherieSold = []Item{PouletRoti, CuissePoulet, PilonPoulet, Aiguillettes, Magret, FoieGras,
	Steak, SteakHache, RotiBoeuf, MorceauxVeau, Gigot, CoteAgneau, Saucisse, Jambon, Cotelette, Pate}

// CrepesSucrees, CrepesSalees : les crêpes garnies de la crêperie (chacun en commande une de chaque).
var CrepesSucrees = []Item{CrepeSucre, CrepeBananeChoco, CrepePommeCannelle, CrepePoireVanille}
var CrepesSalees = []Item{CrepeSaucisseOeuf, CrepeThonPoireau, CrepeCanard, CrepeSteakBrebis}
