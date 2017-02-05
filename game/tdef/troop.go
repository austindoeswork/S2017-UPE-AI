/*
TROOP DEFINITIONS, these all inherit from UnitBase (parent struct) and Unit (interface)
A troop definition consists of a definition of Die/Birth/Attack/ReceiveDamage/Iterate
*/
package tdef

/*
(Nut) [Aggro]
Standard foot-soldier with below average HP, average stride/reach and high speed/damage.
This unit is great for high DPS but needs to be supplemented with meatier troops. The Nut is the cheapest unit in the game, and the most cost efficient.
*/
type Nut struct {
	UnitBase
}

func (u *Nut) CheckBuyable(income, bits int) bool {
	return bits >= 200
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
	owner.SetBits(owner.Bits() - 200)
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
		reach:  120,
	}}
}

/*
(Bolt) [Aggro]
Bolts have worse speed/reach than Nuts, but above average HP.
Bolts deal damage as a percent of current enemy HP as opposed to a standard fixed damage amount.
On average, Bolts do better versus large opponents than Nuts, but are worse versus smaller opponents than Nuts. Bolts cost a little more than Nuts.
*/
type Bolt struct {
	UnitBase
}

func (u *Bolt) CheckBuyable(income, bits int) bool {
	return bits >= 400
}
func (u *Bolt) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *Bolt) Prep(owner *Player, opponent *Player) {
	if !u.VerifyTarget() {
		unit, _ := opponent.FindClosestUnit(u)
		u.target = unit
	}
}
func (u *Bolt) Iterate() {
	if u.target != nil {
		u.target.ReceiveDamage(u.target.MaxHP()/10)
	} else {
		if u.owner == 1 {
			u.x += u.stride
		} else {
			u.x -= u.stride
		}
		// HOPEFULLY NO NEED TO CHECK IF u.x < 0 or > GAMEWIDTH, BECAUSE WE SHOULD BE ATTACKING CORE
	}
}
func (u *Bolt) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 400)
}
func (u *Bolt) Die(owner *Player, opponent *Player) {}

func NewBolt(x, y, owner int) Unit {
	return &Bolt{UnitBase{
		owner:  owner,
		enum:   1,
		x:      x,
		y:      y,
		speed:  8,
		damage: 0,
		hp:     300,
		maxhp:  300,
		stride: 15,
		reach:  100,
	}}
}

/*
(Grease Monkey) [Aggro]
Healer troop that heals all friendly troops by a fixed amount in a small AOE around it. 
Grease Monkeys are intended to not be very cost efficient unless there is a decently sized group of units.
*/
type GreaseMonkey struct {
	UnitBase
	move bool
}

func (u *GreaseMonkey) CheckBuyable(income, bits int) bool {
	return bits >= 300
}
func (u *GreaseMonkey) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *GreaseMonkey) Prep(owner *Player, opponent *Player) {
	u.move = true // grease monkeys only move if they don't heal ANYTHING within range
	for _, element := range owner.Units {
		if element == u { // it doesn't heal itself
			continue
		}
		if element.HP() < element.MaxHP() {
			diffX := intAbsDiff(u.X(), element.X())
			diffY := intAbsDiff(u.Y(), element.Y())
			if diffX <= u.Reach() && diffY <= u.Reach() {
				element.SetHP(element.HP() + 10)
				if element.HP() > element.MaxHP() { // grease monkeys do not heal above max HP
					element.SetHP(element.MaxHP())
				}
				u.move = false
			}
		}
	}
}

func (u *GreaseMonkey) Iterate() {
	if u.move == true {
		if u.owner == 1 {
			u.x += u.stride
		} else {
			u.x -= u.stride
		}
	}
	// do we ever need to check to see if GreaseMonkeys are going to be leaving the field?
}
func (u *GreaseMonkey) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 300)
}
func (u *GreaseMonkey) Die(owner *Player, opponent *Player) {}

func NewGreaseMonkey(x, y, owner int) Unit {
	return &GreaseMonkey{ // note the slightly different initializer when you need to init values outside of UB (like move)
		UnitBase: UnitBase{
			owner:  owner,
			enum:   1,
			x:      x,
			y:      y,
			speed:  5,
			damage: 0,
			hp:     75,
			maxhp:  75,
			stride: 10,
			reach:  200,
		},
		move: true, // additional field for GreaseMonkeys (because they don't have a strict targetting system)
	}
}
