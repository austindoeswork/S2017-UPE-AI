package tdef

import (
	"bytes"
	"fmt"
)

/*
Unit enum:
-2: Main core
-1: Objective towers
--
00: Nut
01: Bolt
02: Grease Monkey
03: Walker
04: Aimbot
05: Hard Drive
06: Scrapheap
07: Gas Guzzler
08: Terminator
09: Blackhat
10: Malware
11: Gandhi

50: Peashooter
51: Firewall
52: Guardian
53: Bank
54: Junkyard
55: Start Up
56: Corporation
57: Warp Drive
58: Jamming Station
59: Hotspot
*/

// Troops and towers are units. All of their internal variables are private to promote good coding practice
type Unit interface {
	ExportJSON(buffer *bytes.Buffer) // ExportJSON is used for sending information to the front-end
	Enum() int                       // type of unit (i.e. 0 is nut)
	Owner() int
	X() int
	Y() int
	SetX(x int)
	SetY(y int)
	HP() int
	Damage() int
	SetDamage(dmg int)
	MaxHP() int
	SetHP(hp int)
	Speed() int
	SetSpeed(speed int)
	Stride() int
	SetStride(stride int)
	Reach() int

	// special functions
	SetEnabled(owner *Player, enable bool) // player is needed sometimes when towers being disabled affects their player's income (i.e. banks)
	Enabled() bool
	SetInfected() // units are only infected, never uninfected
	Infected() bool

	// note that VerifyTarget() is in the UnitBase implementation, but probably shouldn't be a part of the required interface

	// below here is not implemented by UnitBase
	CheckBuyable(income, bits int) bool      // returns true if having income and # of bits will afford the unit (change to Player*?)
	Prep(owner *Player, opponent *Player)    // called by each unit each turn (will figure out if unit is attacking or moving normally
	Iterate(owner *Player, opponent *Player) // called by each unit each turn (this will attack or move as necessary)
	ReceiveDamage(damage int)                // called when this unit is under attack. this should NOT kill the unit.
	Die(owner *Player, opponent *Player)     // called by unit cleanup, used for interesting death effects like Scrapheap
	Birth(owner *Player, opponent *Player)   // called by unit creation, used for interesting spawn effects like Gandhi
}

// THE FOLLOWING INTERFACE MAKES UNIT SLICES SORTABLE
type SortByX []Unit

func (u SortByX) Len() int {
	return len(u)
}

func (u SortByX) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}

func (u SortByX) Less(i, j int) bool {
	return u[i].X() <= u[j].X()
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

	// special characteristics
	enabled  bool // towers can be disabled by blackhats
	infected bool // troops can be infected by malware
}

// things not implemented by UnitBase: Attack, ReceiveDamage, Iterate

func (ub *UnitBase) Owner() int {
	return ub.owner
}
func (ub *UnitBase) ExportJSON(buffer *bytes.Buffer) { // rest of information is not really important to front-end
	buffer.WriteString(fmt.Sprintf(`{"owner":%d,"x":%d,"y":%d,"maxhp":%d,"hp":%d,"enum":%d}`, ub.owner, ub.x, ub.y, ub.maxhp, ub.hp, ub.enum))
}
func (ub *UnitBase) Enum() int {
	return ub.enum
}
func (ub *UnitBase) X() int {
	return ub.x
}
func (ub *UnitBase) SetX(x int) {
	ub.x = x
}
func (ub *UnitBase) Y() int {
	return ub.y
}
func (ub *UnitBase) SetY(y int) {
	ub.y = y
}
func (ub *UnitBase) Damage() int {
	return ub.damage
}
func (ub *UnitBase) SetDamage(damage int) {
	ub.damage = damage
}
func (ub *UnitBase) Speed() int {
	return ub.speed
}
func (ub *UnitBase) SetSpeed(speed int) {
	ub.speed = speed
}

func (ub *UnitBase) MaxHP() int {
	return ub.maxhp
}
func (ub *UnitBase) HP() int {
	return ub.hp
}
func (ub *UnitBase) SetHP(hp int) {
	ub.hp = hp
}
func (ub *UnitBase) Stride() int {
	return ub.stride
}
func (ub *UnitBase) SetStride(stride int) {
	ub.stride = stride
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

// SPECIAL CHARACTERISTICS

// specific to towers and blackhats
// we require the owner here for overriden methods (banks on disable need to be able to reduce player income)
func (ub *UnitBase) SetEnabled(owner *Player, enable bool) {
	ub.enabled = enable
}
func (ub *UnitBase) Enabled() bool {
	return ub.enabled
}

// specific to troops and malware
func (ub *UnitBase) SetInfected() {
	ub.infected = true
}
func (ub *UnitBase) Infected() bool {
	return ub.infected
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
	switch enum {
	case 0:
		return NewNut(x, y, owner)
	case 1:
		return NewBolt(x, y, owner)
	case 2:
		return NewGreaseMonkey(x, y, owner)
	case 3:
		return NewWalker(x, y, owner)
	case 4:
		return NewAimbot(x, y, owner)
	case 5:
		return NewHardDrive(x, y, owner)
	case 6:
		return NewScrapheap(x, y, owner)
	case 7:
		return NewGasGuzzler(x, y, owner)
	case 8:
		return NewTerminator(x, y, owner)
	case 9:
		return NewBlackhat(x, y, owner)
	case 10:
		return NewMalware(x, y, owner)
	case 11:
		return NewGandhi(x, y, owner)
	default:
		return nil
	}
}

// For lane towers, specify PLOT not x, y
// Note that territory checking should NOT be handled here, this assumes that it is a valid Tower
func NewTower(x, y, owner, enum int) Unit {
	switch enum {
	case 50:
		return NewPeashooter(x, y, owner)
	case 51:
		return NewFirewall(x, y, owner)
	case 52:
		return NewGuardian(x, y, owner)
	case 53:
		return NewBank(x, y, owner)
	case 54:
		return NewJunkyard(x, y, owner)
	case 55:
		return NewStartUp(x, y, owner)
	case 56:
		return NewCorporation(x, y, owner)
	case 57:
		return NewWarpDrive(x, y, owner)
	case 58:
		return NewJammingStation(x, y, owner)
	case 59:
		return NewHotspot(x, y, owner)
	default:
		return nil
	}
}
