package sim

// La zone du joueur : dans chaque lieu, une dalle de béton réservée au joueur
// (personne d'autre n'y va). On y trouve l'hélicoptère et ses téléporteurs vers
// les autres lieux, les améliorations des ouvriers du lieu, la
// caisse des commandes (s'il y en a) et le vestiaire (page « Personnage »).

// géométrie de la zone (décalages depuis Yard.X)
const (
	YardW      = 830.0 // (assez large pour 32 téléporteurs : quatre colonnes de huit)
	YardTop    = 140.0
	YardBottom = 1505.0
	yardRow1   = 912.0 // améliorations des ouvriers (sous 8 rangées de téléporteurs)
	yardOldW   = 560.0 // (largeur d'avant : les lieux donnent toujours le bord gauche d'avant, la zone grandit vers la gauche)
)

// YardPad : centre de l'hélisurface.
func (p *Place) YardPad() V {
	if p.yardH {
		return V{p.Yard.X + 150, p.Yard.Y + 190}
	}
	return V{p.Yard.X + 150, 300}
}

// YardTele : case n° i des téléporteurs (quatre colonnes, huit rangées, à droite de l'hélico ;
// zone à l'horizontale : dix colonnes, trois rangées).
func (p *Place) YardTele(i int) Rect {
	if p.yardH {
		return Rect{p.Yard.X + 300 + float64(i%10)*130, p.Yard.Y + 20 + float64(i/10)*112, 124, 88} // (dix colonnes, trois rangées)
	}
	return Rect{p.Yard.X + 300 + float64(i%4)*130, 160 + float64(i/4)*90, 124, 84} // (quatre colonnes, huit rangées)
}

// setupYardH : la zone du joueur à l'horizontale (supermarché : sous la route,
// x : bord gauche, y : bord du haut). Hélico, téléporteurs, améliorations, vestiaire.
func (p *Place) setupYardH(x, y float64) {
	p.yardH = true
	p.Yard = Rect{x - 650, y, 2150, 380} // (agrandie vers la gauche : trente téléporteurs)
	col := 0
	for _, id := range []string{"etals", "entrepot"} {
		if sp := p.spot(id); sp != nil {
			sp.Zone = Rect{x + 990 + float64(col)*135, y + 30, 130, 124}
			col++
		}
	}
	v := &Station{ID: "vestiaire", Name: "Vestiaire", Kind: Trash, Built: true, yard: true}
	v.Body = Rect{x + 1000, y + 236, 230, 110}
	v.Zone = Rect{-1000, -1000, 0, 0}
	v.place = p.ID
	p.Stations = append(p.Stations, v)
	p.Vestiaire = Rect{x + 1250, y + 250, 180, 90}
	p.Start = V{p.Yard.X + 150, y + 330}
}

// YardRowY : hauteur de la rangée d'améliorations des ouvriers (pour le titre peint au sol).
func YardRowY() float64 { return yardRow1 }

// MinX, MaxX : bornes du terrain (la zone du joueur peut déborder du lieu).
// (marge de 360 : la caméra peut centrer la zone, qui n'est plus cachée par les encadrés en haut à gauche)
func (p *Place) MinX() float64 { return min(0, p.Yard.X-360) }
func (p *Place) MaxX() float64 { return max(p.W, p.Yard.X+p.Yard.W+360) }

// setupYard : place la zone du joueur (x : bord gauche de la dalle), y range les
// améliorations des ouvriers, les commandes et le vestiaire, et fait
// arriver le joueur à l'hélicoptère.
func (p *Place) setupYard(x float64) {
	x += yardOldW - YardW // (le bord droit ne bouge pas : la zone grandit vers la gauche)
	p.Yard = Rect{x, YardTop, YardW, YardBottom - YardTop}
	ux := func(u float64) float64 { return x + u }
	// améliorations des ouvriers (les drones ne s'améliorent qu'avec les points de l'école)
	col := 0
	for _, id := range []string{"ouvriers", "vitesse", "capacite"} {
		if sp := p.spot(id); sp != nil {
			sp.Zone = Rect{ux(10 + float64(col)*135), yardRow1, 130, 124}
			col++
		}
	}
	// commandes, juste sous les améliorations
	if c := p.Station("commande"); c != nil {
		c.Body = Rect{ux(190), 1215, 180, 70} // (son panneau dépasse au-dessus)
		c.Zone = Rect{ux(170), 1300, 220, 100}
		c.yard = true
	}
	// vestiaire (décor) et sa porte
	v := &Station{ID: "vestiaire", Name: "Vestiaire", Kind: Trash, Built: true, yard: true}
	v.Body = Rect{ux(480), 1215, 260, 120}
	v.Zone = Rect{-1000, -1000, 0, 0}
	v.place = p.ID
	p.Stations = append(p.Stations, v)
	p.Vestiaire = Rect{ux(520), 1350, 180, 90}
	// on arrive en hélicoptère
	p.Start = V{ux(150), 450}
}

func (p *Place) spot(id string) *Spot {
	for _, sp := range p.Spots {
		if sp.ID == id {
			return sp
		}
	}
	return nil
}
