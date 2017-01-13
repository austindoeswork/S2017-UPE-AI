package tdef

import (
	"fmt"
)

/*
Unit enum:
Main core: -2
Objective towers: -1
Footsoldier: 0
*/
type Unit struct {
	enum   int
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
	return fmt.Sprintf(`{"x": %d, "y": %d, "maxhp": %d, "hp": %d, "enum": %d}`, u.x, u.y, u.maxhp, u.hp, u.enum)
}

func (u *Unit) Enum() int {
	return u.enum
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
func NewUnit(x, lane, enum int) *Unit {
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
		enum:   0,
		x:      x,
		y:      y,
		speed:  5,
		damage: 10,
		hp:     100,
		maxhp:  100,
		stride: 5, // updateGrid automatically handles owners' units moving in opposite dirs
		reach:  150,
	}
}

// NOTE: you specify y and not LANE # for core towers.
func NewCoreTower(x, y, enum int) *Unit {
	var hp, speed, damage, reach int
	switch enum {
	case -2:
		damage = 100
		hp = 1000
		speed = 20
		reach = 300
	case -1:
		damage = 40
		hp = 500
		speed = 5
		reach = 100
	}
	return &Unit{
		enum:   enum,
		x:      x,
		y:      y,
		damage: damage,
		maxhp:  hp,
		hp:     hp,
		speed:  speed,
		stride: 0,
		reach:  reach,
	}
}

// For lane towers, specify PLOT not x, y
func NewTower(plot, enum int) *Unit {
	var x, y int
	x = GAMEWIDTH*(plot%4)/4 + GAMEWIDTH/8
	y = GAMEHEIGHT*int(plot/4)/4 + GAMEHEIGHT/8
	var hp, speed, damage, reach int
	switch enum {
	case 10:
		damage = 50
		hp = 200
		speed = 5
		reach = 200
	}
	return &Unit{
		enum:   enum,
		x:      x,
		y:      y,
		damage: damage,
		maxhp:  hp,
		hp:     hp,
		speed:  speed,
		stride: 0,
		reach:  reach,
	}
}
