package tdef

import (
	"fmt"
)

type Unit struct {
	x      int // bottom left = 0, bottom right = max x
	y      int // bottom left = 0, top left = max y
	damage int // damage it deals in an attack
	maxhp  int // total hp, doesnt change
	hp     int // current hp
	speed  int // every <speed> frames, this unit acts (higher speed is slower)
	stride int // when moving, this unit moves <stride> pixels
	reach  int // range of a unit (rename?)

	target *Unit // nil = move, non-nil = shoot
}

// ExportJSON is used for sending information to the front-end
func (u *Unit) ExportJSON() string { // rest of information is not really important to front-end
	return fmt.Sprintf(`{"x": %d, "y": %d, "maxhp": %d, "hp": %d}`, u.x, u.y, u.maxhp, u.hp)
}

func (u *Unit) X() int {
	return u.x
}
func (u *Unit) Y() int {
	return u.y
}
func (u *Unit) SetX(x int) {
	u.x = x
}
func (u *Unit) SetY(y int) {
	u.y = y
}
func (u *Unit) Damage() int {
	return u.damage
}
func (u *Unit) MaxHP() int {
	return u.maxhp
}
func (u *Unit) HP() int {
	return u.hp
}
func (u *Unit) SetHP(hp int) {
	u.hp = hp
}
func (u *Unit) Stride() int {
	return u.stride
}
func (u *Unit) Reach() int {
	return u.reach
}
func (u *Unit) Target() *Unit {
	return u.target
}

// if target is valid, shoot at it until it's dead
// otherwise, find another one or walk instead
func (u *Unit) VerifyTarget() bool {
	if u.target == nil || u.target.HP() <= 0 ||
		intAbsDiff(u.x, u.target.X()) > u.reach ||
		intAbsDiff(u.y, u.target.Y()) > u.reach {
		// note that we don't actually have to check against the true euclidean distance for valid target
		// because in this game units only move along the x-axis, so when they leave rectangular ranges
		// they will actually only become unreachable if their target leaves through the x-direction
		u.target = nil
		return false
	}
	return true
}
func (u *Unit) SetTarget(unit *Unit) {
	u.target = unit
}

// this is just temporary, everything so far is a "unit" that can move and everything, although towers are units with 0 stride
// NOTE that instead of specifying a y coord for a new unit, you specify a LANE #, and lane is auto set.
func NewUnit(x int, lane int) *Unit {
	var y int
	switch lane {
	case 1:
		y = TOPY
	case 2:
		y = MIDY
	case 3:
		y = BOTY
	}
	return &Unit{
		x:      x,
		y:      y,
		speed:  3,
		damage: 10,
		hp:     100,
		maxhp:  100,
		stride: 5, // updateGrid automatically handles owners' units moving in opposite dirs
		reach:  150,
	}
}

// NOTE: you don't specify y-coordinate of towers, only lane
func NewTower(x int, lane int) *Unit {
	var y int
	switch lane {
	case 1:
		y = TOPY
	case 2:
		y = MIDY
	case 3:
		y = BOTY
	}
	return &Unit{
		x:      x,
		y:      y,
		damage: 100,
		maxhp:  1000,
		hp:     1000,
		speed:  20,
		stride: 0,
		reach:  300,
	}
}
