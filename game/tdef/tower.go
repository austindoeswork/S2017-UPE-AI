/*
TOWER DEFINITIONS, all of these inherit from UnitBase (just like troops do)
*/
package tdef

// TODO: CREATE SEPARATE INSTANCE FOR OBJECTIVES, as on death objectives increase territory size
type Core struct {
	UnitBase
}

func (u *Core) CheckBuyable(income, bits int) bool {
	return false
}
func (u *Core) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *Core) Prep(owner *Player, opponent *Player) {
	if !u.VerifyTarget() {
		unit, _ := opponent.FindClosestUnit(u)
		u.target = unit
	}
}
func (u *Core) Iterate() {
	if u.target != nil {
		u.target.ReceiveDamage(u.damage)
	}
}

func (u *Core) Birth(owner *Player, opponent *Player) {}
func (u *Core) Die(owner *Player, opponent *Player)   {} // player will be able to detect game loss itself

func NewCore(x, y, owner int) Unit {
	return &Core{UnitBase{
		enum:   -2,
		x:      x,
		y:      y,
		owner:  owner,
		damage: 100,
		maxhp:  10000,
		hp:     10000,
		speed:  20,
		stride: 0,
		reach:  300,
	}}
}

func NewObjective(x, y, owner int) Unit {
	return &Core{UnitBase{
		enum:   -1,
		x:      x,
		y:      y,
		owner:  owner,
		damage: 30,
		maxhp:  1000,
		hp:     1000,
		speed:  15,
		stride: 0,
		reach:  150,
	}}
}

type Peashooter struct {
	UnitBase
}

func (u *Peashooter) CheckBuyable(income, bits int) bool {
	return income >= 100 && bits >= 500
}
func (u *Peashooter) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *Peashooter) Prep(owner *Player, opponent *Player) {
	if !u.VerifyTarget() {
		unit, _ := opponent.FindClosestUnit(u)
		u.target = unit
	}
}
func (u *Peashooter) Iterate() {
	if u.target != nil {
		u.target.ReceiveDamage(u.damage)
	}
}

func (u *Peashooter) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 500)
	owner.SetIncome(owner.Income() - 100)
}
func (u *Peashooter) Die(owner *Player, opponent *Player) {
	owner.SetIncome(owner.Income() + 100)
}

func NewPeashooter(x, y, owner int) Unit {
	return &Peashooter{UnitBase{
		enum:   50,
		x:      x,
		y:      y,
		owner:  owner,
		damage: 10,
		maxhp:  200,
		hp:     200,
		speed:  3,
		stride: 0,
		reach:  300,
	}}
}

type Bank struct {
	UnitBase
}

func (u *Bank) CheckBuyable(income, bits int) bool {
	return bits >= 1000
}
func (u *Bank) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *Bank) Prep(owner *Player, opponent *Player) {}
func (u *Bank) Iterate()                             {}
func (u *Bank) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 1000)
	owner.SetIncome(owner.Income() + 100)
}
func (u *Bank) Die(owner *Player, opponent *Player) {
	owner.SetIncome(owner.Income() - 100)
}

func NewBank(x, y, owner int) Unit {
	return &Bank{UnitBase{
		enum:   51,
		x:      x,
		y:      y,
		owner:  owner,
		damage: 10,
		maxhp:  200,
		hp:     200,
		speed:  3,
		stride: 0,
		reach:  300,
	}}
}
