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
		// grease monkeys don't attack, so we check to make sure they don't go out of bounds
		if u.owner == 1 {
			u.x += u.stride
			if u.x >= GAMEWIDTH {
				u.x = GAMEWIDTH - 1
			}
		} else {
			u.x -= u.stride
			if u.x < 0 {
				u.x = 0
			}
		}
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
		u.target.ReceiveDamage(u.target.MaxHP() / 10)
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

/*
(Walker) [Midrange]
Speedy (high speed, stride) and durable (high HP) troop that can be quickly deployed to contribute to ongoing fights. Has pretty low DPS, but its mobility, above-average reach and very reasonable cost helps compensate for this.
*/

type Walker struct {
	UnitBase
}

func (u *Walker) CheckBuyable(income, bits int) bool {
	return bits >= 800
}
func (u *Walker) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *Walker) Prep(owner *Player, opponent *Player) {
	if !u.VerifyTarget() {
		unit, _ := opponent.FindClosestUnit(u)
		u.target = unit
	}
}
func (u *Walker) Iterate() {
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
func (u *Walker) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 800)
}
func (u *Walker) Die(owner *Player, opponent *Player) {}

func NewWalker(x, y, owner int) Unit {
	return &Walker{UnitBase{
		owner:  owner,
		enum:   0,
		x:      x,
		y:      y,
		speed:  2,
		damage: 5,
		hp:     800,
		maxhp:  800,
		stride: 10,
		reach:  200,
	}}
}

/*
(Aimbot) [Midrange]
Sniper that does very high damage with a ridiculous reach and incredibly low speed/stride/HP. It also has a decently high cost.  Aimbots can shoot into neighboring lanes. They can kill Nuts in one shot, but have such a slow firing rate that they will easily be overwhelmed. This unit is great for long range presence, but definitely needs to be coddled.
*/

type Aimbot struct {
	UnitBase
}

func (u *Aimbot) CheckBuyable(income, bits int) bool {
	return bits >= 3000
}
func (u *Aimbot) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *Aimbot) Prep(owner *Player, opponent *Player) {
	if !u.VerifyTarget() {
		unit, _ := opponent.FindClosestUnit(u)
		u.target = unit
	}
}
func (u *Aimbot) Iterate() {
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
func (u *Aimbot) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 3000)
}
func (u *Aimbot) Die(owner *Player, opponent *Player) {}

func NewAimbot(x, y, owner int) Unit {
	return &Aimbot{UnitBase{
		owner:  owner,
		enum:   0,
		x:      x,
		y:      y,
		speed:  60,
		damage: 100,
		hp:     100,
		maxhp:  100,
		stride: 5,
		reach:  1000,
	}}
}

/*
(Hard Drive) [Midrange]
Bulky melee fighter with above-average (but not totally cost efficient) HP, low speed, high damage and high cost. Additionally has a caveat that a single source of damage can only deal up to 30 damage to a Hard Drive. Perfect for breaking through high damage areas, but is countered by high DPS troops.
*/

type HardDrive struct {
	UnitBase
}

func (u *HardDrive) CheckBuyable(income, bits int) bool {
	return bits >= 2500
}
func (u *HardDrive) ReceiveDamage(damage int) {
	if damage < 30 {
		u.hp -= damage
	} else {
		u.hp -= 30
	}
}
func (u *HardDrive) Prep(owner *Player, opponent *Player) {
	if !u.VerifyTarget() {
		unit, _ := opponent.FindClosestUnit(u)
		u.target = unit
	}
}
func (u *HardDrive) Iterate() {
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
func (u *HardDrive) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 2500)
}
func (u *HardDrive) Die(owner *Player, opponent *Player) {}

func NewHardDrive(x, y, owner int) Unit {
	return &HardDrive{UnitBase{
		owner:  owner,
		enum:   0,
		x:      x,
		y:      y,
		speed:  5,
		damage: 50,
		hp:     500,
		maxhp:  500,
		stride: 5,
		reach:  50,
	}}
}

// SCRAPHEAP COMING SOON

/*
(Gas Guzzler) [Control]
Large unit that starts with a massive amount of HP, does damage equal to its HP and walks incredibly slowly with a very low range. Decently expensive and also costs a little bit of income, but is also a must-deal-with threat.
*/
type GasGuzzler struct {
	UnitBase
}

func (u *GasGuzzler) CheckBuyable(income, bits int) bool {
	return bits >= 10000
}
func (u *GasGuzzler) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *GasGuzzler) Prep(owner *Player, opponent *Player) {
	if !u.VerifyTarget() {
		unit, _ := opponent.FindClosestUnit(u)
		u.target = unit
	}
}
func (u *GasGuzzler) Iterate() {
	if u.target != nil {
		u.target.ReceiveDamage(u.HP())
	} else {
		if u.owner == 1 {
			u.x += u.stride
		} else {
			u.x -= u.stride
		}
		// HOPEFULLY NO NEED TO CHECK IF u.x < 0 or > GAMEWIDTH, BECAUSE WE SHOULD BE ATTACKING CORE
	}
}
func (u *GasGuzzler) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 10000)
}
func (u *GasGuzzler) Die(owner *Player, opponent *Player) {}

func NewGasGuzzler(x, y, owner int) Unit {
	return &GasGuzzler{UnitBase{
		owner:  owner,
		enum:   0,
		x:      x,
		y:      y,
		speed:  5,
		damage: 0,
		hp:     10000,
		maxhp:  10000,
		stride: 5,
		reach:  50,
	}}
}

/*
(Terminator) [Control]
Super-soldier troop that has incredibly high DPS and below-average health, but otherwise has comparable stats to a Nut. Terminators are extremely expensive.
*/
type Terminator struct {
	UnitBase
}

func (u *Terminator) CheckBuyable(income, bits int) bool {
	return bits >= 9000
}
func (u *Terminator) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *Terminator) Prep(owner *Player, opponent *Player) {
	if !u.VerifyTarget() {
		unit, _ := opponent.FindClosestUnit(u)
		u.target = unit
	}
}
func (u *Terminator) Iterate() {
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
func (u *Terminator) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 9000)
}
func (u *Terminator) Die(owner *Player, opponent *Player) {}

func NewTerminator(x, y, owner int) Unit {
	return &Terminator{UnitBase{
		owner:  owner,
		enum:   0,
		x:      x,
		y:      y,
		speed:  5,
		damage: 100,
		hp:     80,
		maxhp:  80,
		stride: 6,
		reach:  120,
	}}
}
