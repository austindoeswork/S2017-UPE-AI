/*
TROOP DEFINITIONS, these all inherit from UnitBase (parent struct) and Unit (interface)
A troop definition consists of a definition of Die/Birth/Attack/ReceiveDamage/Iterate
*/
package tdef

// Nuts have pretty straightforward implementations
type Nut struct {
	UnitBase
}

func (u *Nut) CheckBuyable(income, bits int) bool {
	return bits >= 100
}
func (u *Nut) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *Nut) Prep(owner *Player, opponent *Player) {
	if !u.VerifyTarget() {
		unit, _ := opponent.FindClosestUnit(u)
		u.target = unit
	}
}
func (u *Nut) Iterate() {
	if u.target != nil {
		u.target.ReceiveDamage(u.damage)
	} else {
		if u.owner == 1 {
			u.x += u.stride
		} else {
			u.x -= u.stride
		}
		// HOPEFULLY NO NEED TO CHECK IF u.x < 0 or > GAMEWIDTH, BECAUSE WE SHOULD BE ATTACKING CORE
	}
}
func (u *Nut) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 100)
}
func (u *Nut) Die(owner *Player, opponent *Player) {}

func NewNut(x, y, owner int) Unit {
	return &Nut{UnitBase{
		owner:  owner,
		enum:   0,
		x:      x,
		y:      y,
		speed:  5,
		damage: 10,
		hp:     100,
		maxhp:  100,
		stride: 10,
		reach:  100,
	}}
}
