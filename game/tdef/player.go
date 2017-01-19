package tdef

import (
	"fmt"
	"math"
)

type Player struct {
	owner  int // who owns this, player 1 or 2?
	income int // X coins per second
	coins  int // total number of coins

	MainTower *Unit // if this dies you die

	Units  []*Unit         // list of all units
	Towers [NUMPLOTS]*Unit // list of all towers (CORE AND OBJECTIVES ARE NOT TOWERS), this is organized by plot
}

func (p *Player) Owner() int {
	return p.owner
}
func (p *Player) Income() int {
	return p.income
}
func (p *Player) SetIncome(income int) {
	p.income = income
}
func (p *Player) Coins() int {
	return p.coins
}
func (p *Player) SetCoins(coins int) {
	p.coins = coins
}

// TODO change terminology, unit should be troop
// will eventually be a list of some sort
// returns true if player can afford unit, false otherwise
func (p *Player) BuyUnit(x, lane, enum int) bool {
	if p.coins >= 300 {
		p.AddUnit(NewUnit(x, lane, enum))
		p.coins -= 300
		return true
	}
	return false
}

// checks to see if a tower plot is within the player's territory
func (p *Player) isPlotInTerritory(plot int) bool {
	var x int
	x = GAMEWIDTH*(plot%4)/4 + GAMEWIDTH/8
	if (p.owner == 1 && x > GAMEWIDTH/2) || (p.owner == 2 && x < GAMEWIDTH/2) { // out of territory
		return false
	}
	return true
}

// will eventually be a list of some sort
// returns true if player can afford tower, false otherwise
// enum: 10 = basic shooting tower, 11 = income tower
func (p *Player) BuyTower(plot, enum int) bool {
	if p.isPlotInTerritory(plot) == true && p.Towers[plot] == nil {
		if enum == 10 && p.coins >= 500 && p.income >= 100 {
			newTower := NewTower(plot, enum)
			p.AddUnit(newTower)
			p.Towers[plot] = newTower
			p.coins -= 500
			p.income -= 100
			return true
		} else if enum == 11 && p.coins >= 2000 {
			newTower := NewTower(plot, enum)
			p.AddUnit(newTower)
			p.Towers[plot] = newTower
			p.coins -= 2000
			p.income += 100
			return true
		}
	}
	return false
}

func (p *Player) AddUnit(unit *Unit) {
	if unit == nil {
		return
	}
	p.Units = append(p.Units, unit)
}

func NewPlayer(owner int) *Player {
	var corex, objx int
	switch owner {
	case 1:
		corex = 0
		objx = XOFFSET
	case 2:
		corex = GAMEWIDTH - 1
		objx = GAMEWIDTH - 1 - XOFFSET
	}
	mainTower := NewCoreTower(corex, MIDY, -2) // need to figure out where maintowers belong, temporarily on midlane
	return &Player{
		owner:     owner,
		income:    500,
		coins:     0,
		MainTower: mainTower,
		Units:     []*Unit{NewCoreTower(objx, TOPY, -1), NewCoreTower(objx, MIDY, -1), NewCoreTower(objx, BOTY, -1)}, // inits lane objectives
		Towers:    [NUMPLOTS]*Unit{},
	}
}

// searches each lane for the closest object to current unit within range
// TODO n^2, we can probably find optimizations
func (p *Player) FindClosestUnit(unit *Unit) (*Unit, float64) {
	var minUnit *Unit
	var minDist float64

	for _, element := range p.Units {
		diffX := intAbsDiff(unit.X(), element.X())
		diffY := intAbsDiff(unit.Y(), element.Y())
		dist := math.Pow(float64(unit.X()-element.X()), 2) + math.Pow(float64(unit.Y()-element.Y()), 2)
		if (minUnit == nil || dist < minDist) && diffX <= unit.Reach() && diffY <= unit.Reach() {
			minDist = dist
			minUnit = element
		}
	}

	if minUnit == nil { // if no other towers, attack main tower
		if p.owner == 1 && unit.X() <= unit.Reach() {
			return p.MainTower, float64(unit.X())
		} else if p.owner == 2 && unit.X() >= GAMEWIDTH-unit.Reach() {
			return p.MainTower, float64(GAMEWIDTH - unit.X())
		}
	}

	return minUnit, minDist
}

// generates a JSON object in string form that is used for display purposes
func (p *Player) ExportJSON() string { // used for exporting to screen
	unitString := `"units": [`
	for index, element := range p.Units {
		unitString += element.ExportJSON()
		if index != len(p.Units)-1 {
			unitString += ","
		}
	}
	unitString += `], "mainTower": ` + p.MainTower.ExportJSON() + "}"
	return fmt.Sprintf(`{"owner": %d, "income": %d, "coins": %d, `, p.owner, p.income, p.coins) + unitString
}

// iterates over each of a player's units to see whether they should shoot or move.
// if they have a target and it's valid, they'll shoot at it
// if they have an invalid target, but find a new valid one, they'll shoot at it
// else, they'll move.
// this function call does not actually trigger shooting or moving, this just sets the "target" ptr of each unit.
func (p *Player) SetUnitTargets(other *Player, frame int64) {
	for _, element := range append(p.Units, p.MainTower) {
		if element.speed == 0 || frame%int64(element.speed) == 0 {
			if !element.VerifyTarget() {
				unit, _ := other.FindClosestUnit(element)
				element.SetTarget(unit)
			}
		}
	}
}

// iterates over each of a player's units and shoots at the unit's set target or move accordingly
func (p *Player) IterateUnits(frame int64) {
	for _, element := range append(p.Units, p.MainTower) {
		if element.speed == 0 || frame%int64(element.speed) == 0 {
			if element.Target() == nil && p.owner == 1 {
				element.SetX(element.X() + element.Stride())
			} else if element.Target() == nil && p.owner == 2 {
				element.SetX(element.X() - element.Stride())
			} else { // found a target, fire
				element.Target().SetHP(element.Target().HP() - element.Damage())
			}
		}
	}
}

// units with <=0 hp don't die until this step, they are cleaned up here.
func (p *Player) UnitCleanup() {
	for index, element := range p.Units {
		if element.HP() < 0 {
			if index == len(p.Units)-1 {
				p.Units = p.Units[:index]
			} else {
				p.Units = append(p.Units[:index], p.Units[index+1:]...)
			}
		}
	}
	for index, element := range p.Towers {
		if element == nil { // note that Towers is an array that will always be of size NUMPLOTS, not a slice
			continue
		}
		if element.HP() < 0 {
			p.Towers[index] = nil
		}
	}
}

func (p *Player) IsAlive() bool { // checks life of main tower
	return p.MainTower.HP() > 0
}
