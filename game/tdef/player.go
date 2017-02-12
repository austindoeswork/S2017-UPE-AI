package tdef

import (
	"fmt"
	"math"
)

// IDENTIFICATION OF PLAYERS (usernames etc) IS HANDLED BY GAMEMANGER/WRAPPER, NOT HERE
// GAMES ONLY DEAL WITH INTERNAL GAME LOGIC
type Player struct {
	owner  int // who owns this, player 1 or 2?
	income int // X coins per second
	bits   int // total number of coins

	Spawns [3]int // X coordinates of the three lanes (0, 1, 2 = x positions for lanes 1, 2, 3)

	MainTower Unit // if this dies you die

	Units  []Unit         // list of all units
	Towers [NUMPLOTS]Unit // list of all towers (CORE AND OBJECTIVES ARE NOT TOWERS), this is organized by plot

	// special unit things
	madeGandhi bool // true if Gandhi has been made (player cannot make another gandhi), false otherwise
}

// Determine's a player's tiebreak score in the event of time running out
func (p *Player) GetTiebreak() int {
	pts := 0
	for _, elem := range p.Units {
		pts += elem.HP()
	}
	for _, elem := range p.Towers {
		if elem != nil {
			pts += elem.HP()
		}
	}
	pts += p.MainTower.HP() * 3
	pts += p.bits / 1000
	pts += p.income
	return pts
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
func (p *Player) Bits() int {
	return p.bits
}
func (p *Player) SetBits(bits int) {
	p.bits = bits
}

// TODO: remove lane from here, players can hold their own spawns
// returns true if player can afford unit, false otherwise
func (p *Player) BuyTroop(x, lane, enum int, opponent *Player) bool {
	troop := NewTroop(x, lane, p.owner, enum)
	for _, element := range p.Towers {
		if element == nil {
			continue
		}
		if element.Enum() == 57 &&
			(intAbsDiff(element.Y(), troop.Y()) <= 100) &&
			((p.owner == 1 && element.X() > x) || (p.owner == 2 && element.X() < x)) {
			x = element.X()
		}
	}
	troop.SetX(x)
	if troop.CheckBuyable(p.income, p.bits) {
		if enum == 11 { // we're trying to create gandhi
			if p.madeGandhi == true {
				return false
			} else {
				p.madeGandhi = true
			}
		}
		troop.Birth(p, opponent) // we call birth before we add the unit officially to make gandhi work
		p.AddUnit(troop)
		return true
	}
	return false
}

// checks to see if a tower plot is within the player's territory
func (p *Player) isPlotInTerritory(x, y int) bool {
	if (p.owner == 1 && x >= GAMEWIDTH/2) || (p.owner == 2 && x <= GAMEWIDTH/2) { // out of territory
		return false
	}
	return true
}

// will eventually be a list of some sort
// returns true if player can afford tower, false otherwise
func (p *Player) BuyTower(plot, enum int, opponent *Player) bool {
	x, y := getPlotPosition(plot)
	if x == -1 || y == -1 {
		return false
	}
	if p.isPlotInTerritory(x, y) == true && p.Towers[plot] == nil {
		newTower := NewTower(x, y, p.owner, enum)
		if newTower.CheckBuyable(p.income, p.bits) {
			p.Towers[plot] = newTower
			newTower.Birth(p, opponent)
			return true
		}
	}
	return false
}

func (p *Player) AddUnit(unit Unit) {
	if unit == nil {
		return
	}
	p.Units = append(p.Units, unit)
}

func NewPlayer(owner int, demoGame bool) *Player {
	var corex, objx int
	var spawns [3]int
	switch owner {
	case 1:
		corex = 0
		objx = XOFFSET
		spawns = [3]int{0, 0, 0}
	case 2:
		corex = GAMEWIDTH - 1
		objx = GAMEWIDTH - 1 - XOFFSET
		spawns = [3]int{GAMEWIDTH - 1, GAMEWIDTH - 1, GAMEWIDTH - 1}
	}
	var mainTower Unit
	var bits, income int
	if demoGame == true {
		mainTower = NewInvincibleCore(corex, MIDY, owner)
		bits = 10000
		income = 10000
	} else {
		mainTower = NewCore(corex, MIDY, owner) // need to figure out where maintowers belong, temporarily on midlane
		bits = 1000000                          // TODO: change to 0 (made it ridic high for testing)
		income = 500
	}
	return &Player{
		owner:     owner,
		income:    income,
		bits:      bits,
		Spawns:    spawns,
		MainTower: mainTower,
		Units:     []Unit{NewObjective(objx, TOPY, owner), NewObjective(objx, MIDY, owner), NewObjective(objx, BOTY, owner)}, // inits lane objectives
		Towers:    [NUMPLOTS]Unit{},
	}
}

// searches each lane for the closest object to current unit within range
// TODO n^2, we can probably find optimizations
func (p *Player) FindClosestUnit(unit Unit) (Unit, float64) {
	var minUnit Unit
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

	for _, element := range p.Towers {
		if element == nil { // p.Towers will always be the total number of plots
			continue
		}
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
	unitString := `"towers": [`
	for index, element := range p.Towers {
		if element == nil {
			unitString += `"nil"`
		} else {
			unitString += element.ExportJSON()
		}
		if index != len(p.Towers)-1 {
			unitString += ","
		}
	}
	unitString += `], "troops": [`
	for index, element := range p.Units {
		unitString += element.ExportJSON()
		if index != len(p.Units)-1 {
			unitString += ","
		}
	}
	unitString += `], "mainTower": ` + p.MainTower.ExportJSON() + "}"
	return fmt.Sprintf(`{"owner": %d, "income": %d, "bits": %d, `, p.owner, p.income, p.bits) + unitString
}

// iterates before anything happens, just a frame initialization stage
func (p *Player) PrepPlayer() {
	for _, element := range p.Towers { // pre-prep phase: reenable all towers
		if element == nil {
			continue
		}
		element.SetEnabled(p, true)
	}
}

// iterates over each of a player's units to see whether they should shoot or move.
// if they have a target and it's valid, they'll shoot at it
// if they have an invalid target, but find a new valid one, they'll shoot at it
// else, they'll move.
// this function call does not actually trigger shooting or moving, this just sets the "target" ptr of each unit.
func (p *Player) PrepUnits(other *Player, frame int64) {
	// I WOULD RATHER HAVE THIS BOILERPLATE CODE THAN USE MEMORY TRYING TO MERGE ALL THE LISTS INTO A SINGLE FOR LOOP
	for _, element := range p.Units {
		if element.Speed() == 0 || frame%int64(element.Speed()) == 0 {
			element.Prep(p, other)
		}
	}

	if p.MainTower.Speed() == 0 || frame%int64(p.MainTower.Speed()) == 0 {
		p.MainTower.Prep(p, other)
	}

	for _, element := range p.Towers {
		if element == nil {
			continue
		}
		if element.Speed() == 0 || frame%int64(element.Speed()) == 0 {
			element.Prep(p, other)
		}
	}
}

// iterates over each of a player's units and shoots at the unit's set target or move accordingly
func (p *Player) IterateUnits(other *Player, frame int64) {
	for _, element := range p.Units {
		if element.Speed() == 0 || frame%int64(element.Speed()) == 0 {
			element.Iterate(p, other)
		}
	}

	if p.MainTower.Speed() == 0 || frame%int64(p.MainTower.Speed()) == 0 {
		p.MainTower.Iterate(p, other)
	}

	for _, element := range p.Towers {
		if element == nil {
			continue
		}
		if element.Speed() == 0 || frame%int64(element.Speed()) == 0 {
			element.Iterate(p, other)
		}
	}
}

// units with <=0 hp don't die until this step, they are cleaned up here.
func (p *Player) UnitCleanup(other *Player) {
	for _, element := range p.Units { // first pass, let the dead units have their death
		if element.HP() <= 0 {
			element.Die(p, other) // we iterate twice because sometimes the length of p.Units changes in Die()
		}
	}
	alive := 0                        // number of alive units
	for _, element := range p.Units { // second pass, remove dead units
		if element.HP() > 0 {
			p.Units[alive] = element
			alive++
		}
	}
	p.Units = p.Units[:alive] // delete dead units, but (TODO) i suspect these are still in the memory!!
	for index, element := range p.Towers {
		if element == nil { // note that Towers is an array that will always be of size NUMPLOTS, not a slice
			continue
		}
		if element.HP() < 0 {
			element.Die(p, other)
			p.Towers[index] = nil
		}
	}
}

func (p *Player) IsAlive() bool { // checks life of main tower
	return p.MainTower.HP() > 0
}
