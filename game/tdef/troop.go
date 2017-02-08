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
func (u *Nut) Iterate(owner *Player, opponent *Player) {
	if u.target != nil {
		u.target.ReceiveDamage(u.damage)
	} else {
		if u.owner == 1 {
			u.x += u.stride
		} else {
			u.x -= u.stride
		}
	}
}
func (u *Nut) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 200)
}
func (u *Nut) Die(owner *Player, opponent *Player) {
	if u.infected == true {
		opponent.AddUnit(NewMalware(u.x, u.y, opponent.Owner()))
	}
	for _, element := range opponent.Towers {
		if element == nil {
			continue
		}
		if element.Enum() == 54 && intAbsDiff(element.X(), u.x) <= element.Reach() &&
			intAbsDiff(element.Y(), u.y) <= element.Reach() { // junkyard change for killing unit
			opponent.SetBits(opponent.Bits() + 500)
		}
	}
}

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
		enabled: true,
		infected: false,
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
func (u *Bolt) Iterate(owner *Player, opponent *Player) {
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
func (u *Bolt) Die(owner *Player, opponent *Player) {
	if u.infected == true {
		opponent.AddUnit(NewMalware(u.x, u.y, opponent.Owner()))
	}
	for _, element := range opponent.Towers {
		if element == nil {
			continue
		}
		if element.Enum() == 54 && intAbsDiff(element.X(), u.x) <= element.Reach() &&
			intAbsDiff(element.Y(), u.y) <= element.Reach() { // junkyard change for killing unit
			opponent.SetBits(opponent.Bits() + 500)
		}
	}
}

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
		enabled: true,
		infected: false,
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

func (u *GreaseMonkey) Iterate(owner *Player, opponent *Player) {
	if u.move == true {
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
	// do we ever need to check to see if GreaseMonkeys are going to be leaving the field?
}
func (u *GreaseMonkey) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 300)
}
func (u *GreaseMonkey) Die(owner *Player, opponent *Player) {
	if u.infected == true {
		opponent.AddUnit(NewMalware(u.x, u.y, opponent.Owner()))
	}
	for _, element := range opponent.Towers {
		if element == nil {
			continue
		}
		if element.Enum() == 54 && intAbsDiff(element.X(), u.x) <= element.Reach() &&
			intAbsDiff(element.Y(), u.y) <= element.Reach() { // junkyard change for killing unit
			opponent.SetBits(opponent.Bits() + 500)
		}
	}
}

func NewGreaseMonkey(x, y, owner int) Unit {
	return &GreaseMonkey{ // note the slightly different initializer when you need to init values outside of UB (like move)
		UnitBase: UnitBase{
			owner:  owner,
			enum:   2,
			x:      x,
			y:      y,
			speed:  5,
			damage: 0,
			hp:     75,
			maxhp:  75,
			stride: 10,
			reach:  200,
			enabled: true,
			infected: false,
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
func (u *Walker) Iterate(owner *Player, opponent *Player) {
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
func (u *Walker) Die(owner *Player, opponent *Player) {
	if u.infected == true {
		opponent.AddUnit(NewMalware(u.x, u.y, opponent.Owner()))
	}
	for _, element := range opponent.Towers {
		if element == nil {
			continue
		}
		if element.Enum() == 54 && intAbsDiff(element.X(), u.x) <= element.Reach() &&
			intAbsDiff(element.Y(), u.y) <= element.Reach() { // junkyard change for killing unit
			opponent.SetBits(opponent.Bits() + 500)
		}
	}
}

func NewWalker(x, y, owner int) Unit {
	return &Walker{UnitBase{
		owner:  owner,
		enum:   3,
		x:      x,
		y:      y,
		speed:  2,
		damage: 5,
		hp:     800,
		maxhp:  800,
		stride: 10,
		reach:  200,
		enabled: true,
		infected: false,
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
func (u *Aimbot) Iterate(owner *Player, opponent *Player) {
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
func (u *Aimbot) Die(owner *Player, opponent *Player) {
	if u.infected == true {
		opponent.AddUnit(NewMalware(u.x, u.y, opponent.Owner()))
	}
	for _, element := range opponent.Towers {
		if element == nil {
			continue
		}
		if element.Enum() == 54 && intAbsDiff(element.X(), u.x) <= element.Reach() &&
			intAbsDiff(element.Y(), u.y) <= element.Reach() { // junkyard change for killing unit
			opponent.SetBits(opponent.Bits() + 500)
		}
	}
}

func NewAimbot(x, y, owner int) Unit {
	return &Aimbot{UnitBase{
		owner:  owner,
		enum:   4,
		x:      x,
		y:      y,
		speed:  60,
		damage: 100,
		hp:     100,
		maxhp:  100,
		stride: 5,
		reach:  1000,
		enabled: true,
		infected: false,
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
func (u *HardDrive) Iterate(owner *Player, opponent *Player) {
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
func (u *HardDrive) Die(owner *Player, opponent *Player) {
	if u.infected == true {
		opponent.AddUnit(NewMalware(u.x, u.y, opponent.Owner()))
	}
	for _, element := range opponent.Towers {
		if element == nil {
			continue
		}
		if element.Enum() == 54 && intAbsDiff(element.X(), u.x) <= element.Reach() &&
			intAbsDiff(element.Y(), u.y) <= element.Reach() { // junkyard change for killing unit
			opponent.SetBits(opponent.Bits() + 500)
		}
	}
}

func NewHardDrive(x, y, owner int) Unit {
	return &HardDrive{UnitBase{
		owner:  owner,
		enum:   5,
		x:      x,
		y:      y,
		speed:  5,
		damage: 50,
		hp:     500,
		maxhp:  500,
		stride: 5,
		reach:  50,
		enabled: true,
		infected: false,
	}}
}

/*
(Scrapheap) [Control]
Bulky and pricey melee fighter with ridiculous HP that damages itself over time.
It has low damage and is a little slow in terms of speed, but otherwise is similar in stats to a Nut.
When it dies it creates two Nuts and a Bolt in its stead. This unit is heavily a defensive one.
*/
type Scrapheap struct {
	UnitBase
}

func (u *Scrapheap) CheckBuyable(income, bits int) bool {
	return bits >= 9000
}
func (u *Scrapheap) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *Scrapheap) Prep(owner *Player, opponent *Player) {
	if !u.VerifyTarget() {
		unit, _ := opponent.FindClosestUnit(u)
		u.target = unit
	}
}
func (u *Scrapheap) Iterate(owner *Player, opponent *Player) {
	if u.target != nil {
		u.target.ReceiveDamage(u.damage)
	} else {
		if u.owner == 1 {
			u.x += u.stride
		} else {
			u.x -= u.stride
		}
	}
	u.hp -= 30
}
func (u *Scrapheap) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 9000)
}
func (u *Scrapheap) Die(owner *Player, opponent *Player) { // stagger the units slightly to make it visually make more sense
	owner.AddUnit(NewNut(u.x-8, u.y, owner.Owner()))
	owner.AddUnit(NewNut(u.x-4, u.y, owner.Owner()))
	owner.AddUnit(NewBolt(u.x+4, u.y, owner.Owner()))
	owner.AddUnit(NewBolt(u.x+8, u.y, owner.Owner()))
	if u.infected == true {
		opponent.AddUnit(NewMalware(u.x, u.y, opponent.Owner()))
	}
	for _, element := range opponent.Towers {
		if element == nil {
			continue
		}
		if element.Enum() == 54 && intAbsDiff(element.X(), u.x) <= element.Reach() &&
			intAbsDiff(element.Y(), u.y) <= element.Reach() { // junkyard change for killing unit
			opponent.SetBits(opponent.Bits() + 500)
		}
	}
}

func NewScrapheap(x, y, owner int) Unit {
	return &Scrapheap{UnitBase{
		owner:  owner,
		enum:   6,
		x:      x,
		y:      y,
		speed:  5,
		damage: 8,
		hp:     9000,
		maxhp:  9000,
		stride: 5,
		reach:  120,
		enabled: true,
		infected: false,
	}}
}

/*
(Gas Guzzler) [Control]
Large unit that starts with a massive amount of HP, does damage equal to its HP and walks incredibly slowly with a very low range. Decently expensive and also costs a little bit of income, but is also a must-deal-with threat.
*/
type GasGuzzler struct {
	UnitBase
}

func (u *GasGuzzler) CheckBuyable(income, bits int) bool {
	return bits >= 10000 && income >= 50
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
func (u *GasGuzzler) Iterate(owner *Player, opponent *Player) {
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
	owner.SetIncome(owner.Income() - 50)
}
func (u *GasGuzzler) Die(owner *Player, opponent *Player) {
	owner.SetIncome(owner.Income() + 50)
	if u.infected == true {
		opponent.AddUnit(NewMalware(u.x, u.y, opponent.Owner()))
	}
	for _, element := range opponent.Towers {
		if element == nil {
			continue
		}
		if element.Enum() == 54 && intAbsDiff(element.X(), u.x) <= element.Reach() &&
			intAbsDiff(element.Y(), u.y) <= element.Reach() { // junkyard change for killing unit
			opponent.SetBits(opponent.Bits() + 500)
		}
	}
}

func NewGasGuzzler(x, y, owner int) Unit {
	return &GasGuzzler{UnitBase{
		owner:  owner,
		enum:   7,
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
	return bits >= 8000
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
func (u *Terminator) Iterate(owner *Player, opponent *Player) {
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
	owner.SetBits(owner.Bits() - 8000)
}
func (u *Terminator) Die(owner *Player, opponent *Player) {
	if u.infected == true {
		opponent.AddUnit(NewMalware(u.x, u.y, opponent.Owner()))
	}
	for _, element := range opponent.Towers {
		if element == nil {
			continue
		}
		if element.Enum() == 54 && intAbsDiff(element.X(), u.x) <= element.Reach() &&
			intAbsDiff(element.Y(), u.y) <= element.Reach() { // junkyard change for killing unit
			opponent.SetBits(opponent.Bits() + 500)
		}
	}
}

func NewTerminator(x, y, owner int) Unit {
	return &Terminator{UnitBase{
		owner:  owner,
		enum:   8,
		x:      x,
		y:      y,
		speed:  5,
		damage: 100,
		hp:     80,
		maxhp:  80,
		stride: 6,
		reach:  120,
		enabled: true,
		infected: false,
	}}
}

/*
(Blackhat) [Aggro Specialty]
Assassin that instantly kills troops, very high speed, very low range and low HP and is pretty expensive.
Blackhats disable enemy towers within a sizeable range.
*/
type Blackhat struct {
	UnitBase
}

func (u *Blackhat) CheckBuyable(income, bits int) bool {
	return bits >= 2500
}
func (u *Blackhat) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *Blackhat) Prep(owner *Player, opponent *Player) {
	if !u.VerifyTarget() {
		unit, _ := opponent.FindClosestUnit(u)
		u.target = unit
	}
	for _, element := range opponent.Towers {
		if element == nil {
			continue
		}
		if intAbsDiff(element.Y(), u.y) <= 150 {
			element.SetEnabled(opponent, false)
		}
	}
}
func (u *Blackhat) Iterate(owner *Player, opponent *Player) {
	if u.target != nil {
		u.target.ReceiveDamage(u.target.HP()) // instant kill
	} else {
		if u.owner == 1 {
			u.x += u.stride
		} else {
			u.x -= u.stride
		}
		// HOPEFULLY NO NEED TO CHECK IF u.x < 0 or > GAMEWIDTH, BECAUSE WE SHOULD BE ATTACKING CORE
	}
}
func (u *Blackhat) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 2500)
}
func (u *Blackhat) Die(owner *Player, opponent *Player) {
	if u.infected == true {
		opponent.AddUnit(NewMalware(u.x, u.y, opponent.Owner()))
	}
	for _, element := range opponent.Towers {
		if element == nil {
			continue
		}
		if element.Enum() == 54 && intAbsDiff(element.X(), u.x) <= element.Reach() &&
			intAbsDiff(element.Y(), u.y) <= element.Reach() { // junkyard change for killing unit
			opponent.SetBits(opponent.Bits() + 500)
		}
	}
}

func NewBlackhat(x, y, owner int) Unit {
	return &Blackhat{UnitBase{
		owner:  owner,
		enum:   9,
		x:      x,
		y:      y,
		speed:  1, // right now, the blackhat has to act each turn in order to make sure it's constantly disabling the enemy
		damage: 0,
		hp:     50,
		maxhp:  50,
		stride: 6,
		reach:  50,
		enabled: true,
		infected: false,
	}}
}

/*
(Malware) [Midrange Specialty]
Virus that has pretty low HP and average reach/speed/stride, but above average damage. 
When it attacks a troop, it infects the troop. Upon that troop’s death another Malware will spawn.
*/

type Malware struct {
	UnitBase
}

func (u *Malware) CheckBuyable(income, bits int) bool {
	return bits >= 6000
}
func (u *Malware) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *Malware) Prep(owner *Player, opponent *Player) {
	if !u.VerifyTarget() {
		unit, _ := opponent.FindClosestUnit(u)
		u.target = unit
	}
}
func (u *Malware) Iterate(owner *Player, opponent *Player) {
	if u.target != nil {
		u.target.ReceiveDamage(u.damage)
		u.target.SetInfected()
	} else {
		if u.owner == 1 {
			u.x += u.stride
		} else {
			u.x -= u.stride
		}
		// HOPEFULLY NO NEED TO CHECK IF u.x < 0 or > GAMEWIDTH, BECAUSE WE SHOULD BE ATTACKING CORE
	}
}
func (u *Malware) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 2500)
}
func (u *Malware) Die(owner *Player, opponent *Player) {
	if u.infected == true {
		opponent.AddUnit(NewMalware(u.x, u.y, opponent.Owner()))
	}
	for _, element := range opponent.Towers {
		if element == nil {
			continue
		}
		if element.Enum() == 54 && intAbsDiff(element.X(), u.x) <= element.Reach() &&
			intAbsDiff(element.Y(), u.y) <= element.Reach() { // junkyard change for killing unit
			opponent.SetBits(opponent.Bits() + 500)
		}
	}
}

func NewMalware(x, y, owner int) Unit {
	return &Malware{UnitBase{
		owner:  owner,
		enum:   10,
		x:      x,
		y:      y,
		speed:  5, // right now, the blackhat has to act each turn in order to make sure it's constantly disabling the enemy
		damage: 30,
		hp:     80,
		maxhp:  80,
		stride: 4,
		reach:  200,
		enabled: true,
		infected: false,
	}}
}

/*
(Gandhi) [Control Specialty]
Exorbitantly expensive troop that vaporizes all troops in its lane when it’s created. 
It does absolutely nothing when it’s out and has very low HP. 
If Gandhi reaches the other side of the stage you win the game. A player can only buy a single Gandhi during a game.
*/
type Gandhi struct {
	UnitBase
}

func (u *Gandhi) CheckBuyable(income, bits int) bool {
	return bits >= 500000
}
func (u *Gandhi) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *Gandhi) Prep(owner *Player, opponent *Player) {
	if !u.VerifyTarget() {
		unit, _ := opponent.FindClosestUnit(u)
		u.target = unit
	}
	for _, element := range opponent.Towers {
		if element == nil {
			continue
		}
		if intAbsDiff(element.Y(), u.y) <= 150 {
			element.SetEnabled(opponent, false)
		}
	}
}
func (u *Gandhi) Iterate(owner *Player, opponent *Player) {
	if u.owner == 1 {
		u.x += u.stride
		if u.x >= GAMEWIDTH {
			opponent.MainTower.ReceiveDamage(10000)
		}
	} else {
		u.x -= u.stride
		if u.x <= 0 {
			opponent.MainTower.ReceiveDamage(10000)
		}
	}
}
func (u *Gandhi) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 500000)
	for _, element := range owner.Units {
		if element != u && intAbsDiff(element.Y(), u.y) <= 1 { // ghetto way of checking lane
			element.ReceiveDamage(1000000)
		}
	}
	for _, element := range opponent.Units {
		if intAbsDiff(element.Y(), u.y) <= 1 { // ghetto way of checking lane
			element.ReceiveDamage(1000000)
		}
	}
}
func (u *Gandhi) Die(owner *Player, opponent *Player) {
	if u.infected == true {
		opponent.AddUnit(NewMalware(u.x, u.y, opponent.Owner()))
	}
	for _, element := range opponent.Towers {
		if element == nil {
			continue
		}
		if element.Enum() == 54 && intAbsDiff(element.X(), u.x) <= element.Reach() &&
			intAbsDiff(element.Y(), u.y) <= element.Reach() { // junkyard change for killing unit
			opponent.SetBits(opponent.Bits() + 500)
		}
	}
}

func NewGandhi(x, y, owner int) Unit {
	return &Gandhi{UnitBase{
		owner:  owner,
		enum:   8,
		x:      x,
		y:      y,
		speed:  1, // right now, the blackhat has to act each turn in order to make sure it's constantly disabling the enemy
		damage: 0,
		hp:     50,
		maxhp:  50,
		stride: 6,
		reach:  50,
		enabled: true,
		infected: false,
	}}
}
