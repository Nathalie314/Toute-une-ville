package sim

// Déplacement des clients (en contournant les rayons) et caddies du supermarché.

// cWalk : comme walkTo, mais en contournant ce qui est construit (rayons, caisses…).
func (p *Place) cWalk(c *Customer, to V, speed, dt float64) bool {
	c.routeT -= dt
	if c.routeTo != to || c.routeT <= 0 {
		c.routeTo, c.routeT = to, 4
		c.route = nil
		if !p.clearC(c.Pos, to) {
			c.route = p.routeC(c.Pos, to)
			if n := len(c.route); n > 0 && c.route[n-1] == to {
				c.route = c.route[:n-1]
			}
		}
	}
	for len(c.route) > 0 {
		if c.walkTo(c.route[0], speed, dt) {
			c.route = c.route[1:]
			continue
		}
		return false
	}
	return c.walkTo(to, speed, dt)
}

// CartSpot : là où l'on prend et rend les caddies (en bas de leur colonne).
func (p *Place) CartSpot() (V, bool) { return p.cartSpot, p.cartSpot != V{} }

// CartsFree : caddies rangés (achetés, moins ceux réservés par les clients en route ou dans le magasin).
func (p *Place) CartsFree() int {
	if p.ID != "super" {
		return 0
	}
	return max(0, p.CartsTotal()-p.cartsOut)
}

// cartIn, cartOut : le chemin du client au supermarché passe déjà par les caddies.
func (p *Place) cartIn(c *Customer)  {}
func (p *Place) cartOut(c *Customer) {}

// atCartSpot : le client vient d'arriver aux caddies : il en prend un, ou le rend.
func (p *Place) atCartSpot(c *Customer) {
	if s, ok := p.CartSpot(); !ok || c.Pos.Dist(s) > 6 {
		return
	}
	if c.Cart && c.State == cLeave {
		c.Cart = false
		p.releaseCart(c)
	} else if !c.Cart && c.State == cEnter && c.cartRes {
		c.Cart = true
	}
}

// releaseCart : le caddie réservé par ce client revient dans la colonne.
func (p *Place) releaseCart(c *Customer) {
	if c.cartRes {
		c.cartRes = false
		p.cartsOut = max(0, p.cartsOut-1)
	}
}
