package sim

import "testing"

// Le premier arrivé à la caisse passe en premier, même s'il est entré plus tard à la ferme.
func TestQueueOrder(t *testing.T) {
	g := NewGame()
	p := g.farm()
	reg := p.Register()
	a := &Customer{State: cShop, idx: 1, Wants: []Want{{It: Ble, N: 1, Got: 0}}, Pos: V{500, 1300}} // n'a rien : part
	b := &Customer{State: cShop, idx: 1, Wants: []Want{{It: Ble, N: 1, Got: 1}}, Pos: reg.QueueHead}
	p.Customers = []*Customer{a, b}
	p.customerStep(b, 0.01) // b arrive à la caisse
	c := &Customer{State: cShop, idx: 1, Wants: []Want{{It: Ble, N: 1, Got: 1}}, Pos: V{1500, 1300}}
	p.Customers = append([]*Customer{c}, p.Customers...) // c est entré avant tout le monde à la ferme
	p.customerStep(c, 0.01)                              // mais arrive à la caisse après b
	p.customerStep(b, 0.01)
	p.customerStep(c, 0.01)
	if b.slot != 0 || c.slot != 1 {
		t.Fatalf("ordre de passage : b=%d c=%d", b.slot, c.slot)
	}
}
