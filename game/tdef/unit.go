package tdef

import (
	"fmt"
)

/*
Unit enum:
-2: Main core
-1: Objective towers
--
00: Nut
*01: Bolt
*02: Grease Monkey
03:

50: Peashooter
51: Bank
*/

// Troops and towers are units. All of their internal variables are private to promote good coding practice
type Unit interface {
	ExportJSON() string // ExportJSON is used for sending information to the front-end
	Enum() int          // type of unit (i.e. 0 is nut)
	Owner() int
	X() int
	Y() int
	HP() int
	MaxHP() int
	Speed() int
	Stride() int
	Reach() int
	// note that VerifyTarget() is in the UnitBase implementation, but probably shouldn't be a part of the required interface

	// below here is not implemented by UnitBase
	CheckBuyable(income, bits int) bool    // returns true if having income and # of bits will afford the unit (change to Player*?)
	Prep(owner *Player, opponent *Player)  // called by each unit each turn (will figure out if unit is attacking or moving normally
	Iterate()                              // called by each unit each turn (this will attack or move as necessary)
	ReceiveDamage(damage int)              // called when this unit is under attack. this should NOT kill the unit.
	Die(owner *Player, opponent *Player)   // called by unit cleanup, used for interesting death effects like Scrapheap
	Birth(owner *Player, opponent *Player) // called by unit creation, used for interesting spawn effects like Gandhi
}

// UnitBase is a very basic implementation of a Unit that is overridden for all purposes
// Even simple units like Nuts should pull from here. UnitBases are not meant to be actual units.
type UnitBase struct {
	owner  int
	enum   int
	x      int // bottom left = 0, bottom right = max x
	y      int // bottom left = 0, top left = max y
	damage int // damage it deals in an attack
	maxhp  int // total hp, doesnt change
	hp     int // current hp
	speed  int // every <speed> frames, this unit acts (higher speed is slower)
	stride int // when moving, this unit moves <stride> pixels
	reach  int // range of a unit (rename?)

	target Unit // nil = move, non-nil = shoot
}

// things not implemented by UnitBase: Attack, ReceiveDamage, Iterate

func (ub *UnitBase) Owner() int {
	return ub.owner
}
func (ub *UnitBase) ExportJSON() string { // rest of information is not really important to front-end
	return fmt.Sprintf(`{"x": %d, "y": %d, "maxhp": %d, "hp": %d, "enum": %d}`, ub.x, ub.y, ub.maxhp, ub.hp, ub.enum)
}
func (ub *UnitBase) Enum() int {
	return ub.enum
}
func (ub *UnitBase) X() int {
	return ub.x
}
func (ub *UnitBase) Y() int {
	return ub.y
}

/* func (ub *UnitBase) SetX(x int) {
	ub.x = x
}
func (ub *UnitBase) SetY(y int) {
	ub.y = y
} */

func (ub *UnitBase) Speed() int {
	return ub.speed
}
func (ub *UnitBase) MaxHP() int {
	return ub.maxhp
}
func (ub *UnitBase) HP() int {
	return ub.hp
}
func (ub *UnitBase) Stride() int {
	return ub.stride
}
func (ub *UnitBase) Reach() int {
	return ub.reach
}

// if target is valid, shoot at it until it's dead
// otherwise, find another one or walk instead
func (ub *UnitBase) VerifyTarget() bool {
	if ub.target == nil || ub.target.HP() <= 0 ||
		intAbsDiff(ub.x, ub.target.X()) > ub.reach ||
		intAbsDiff(ub.y, ub.target.Y()) > ub.reach {
		// note that we don't actually have to check against the true euclidean distance for valid target
		// because in this game units only move along the x-axis, so when they leave rectangular ranges
		// they will actually only become unreachable if their target leaves through the x-direction
		ub.target = nil
		return false
	}
	return true
}

// this is just temporary, everything so far is a "unit" that can move and everything, although towers are units with 0 stride
// NOTE that instead of specifying a y coord for a new unit, you specify a LANE #, and lane is auto set.
func NewTroop(x, lane, owner, enum int) Unit {
	var y int
	switch lane {
	case 1:
		y = TOPY
	case 2:
		y = MIDY
	case 3:
		y = BOTY
	}
	return NewNut(x, y, owner)
}

// For lane towers, specify PLOT not x, y
// Note that territory checking should NOT be handled here, this assumes that it is a valid Tower
func NewTower(plot, owner, enum int) Unit {
	if plot >= NUMPLOTS { // error checking should really not have to be done here
		return nil
	}

	var x, y int
	x = GAMEWIDTH*(plot%4)/4 + GAMEWIDTH/8
	y = GAMEHEIGHT*int(plot/4)/4 + GAMEHEIGHT/8
	switch enum {
	case 50:
		return NewPeashooter(x, y, owner)
	case 51:
		return NewBank(x, y, owner)
	default:
		return nil
	}
}
