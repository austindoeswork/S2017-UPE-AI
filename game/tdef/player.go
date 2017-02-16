package tdef

import (
	"bytes"
	"fmt"
	"math"
	"sort" // used for sorting the lanes
)

// IDENTIFICATION OF PLAYERS (usernames etc) IS HANDLED BY GAMEMANGER/WRAPPER, NOT HERE
// GAMES ONLY DEAL WITH INTERNAL GAME LOGIC
type Player struct {
	owner  int // who owns this, player 1 or 2?
	income int // X coins per second
	bits   int // total number of coins

	MainTower Unit // if this dies you die
 
	Top []Unit // list of top lane units
	Mid []Unit // ditto for mid lane
	Bot []Unit // ditto for bot lane
	Towers [NUMPLOTS]Unit // list of all towers (CORE AND OBJECTIVES ARE NOT TOWERS), this is organized by plot

	// special unit things
	madeGandhi bool // true if Gandhi has been made (player cannot make another gandhi), false otherwise
}

// Determine's a player's tiebreak score in the event of time running out
func (p *Player) GetTiebreak() int {
	pts := 0
	for _, elem := range p.Top {
		pts += elem.HP()
	}
	for _, elem := range p.Mid {
		pts += elem.HP()
	}
	for _, elem := range p.Bot {
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

// returns true if player can afford unit, false otherwise
func (p *Player) BuyTroop(lane, enum int, opponent *Player) bool {
	var x int
	if p.owner == 1 {
		x = 0
	} else {
		x = GAMEWIDTH - 1
	}
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

// TROOPS CANNOT BE ADDED OUTSIDE OF THE LANES
func (p *Player) AddUnit(unit Unit) {
	if unit == nil {
		return
	} else if unit.Y() == TOPY {
		p.Top = append(p.Top, unit)
	} else if unit.Y() == MIDY {
		p.Mid = append(p.Mid, unit)
	} else if unit.Y() == BOTY {
		p.Bot = append(p.Bot, unit)
	}
}

func NewPlayer(owner int, demoGame bool) *Player {
	var corex, objx int
	switch owner {
	case 1:
		corex = 0
		objx = XOFFSET
	case 2:
		corex = GAMEWIDTH - 1
		objx = GAMEWIDTH - 1 - XOFFSET
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
		MainTower: mainTower,
		Top:     []Unit{NewObjective(objx, TOPY, owner)}, // inits lane objectives
		Mid: []Unit{NewObjective(objx, MIDY, owner)},
		Bot: []Unit{NewObjective(objx, BOTY, owner)},
		Towers:    [NUMPLOTS]Unit{},
	}
}

// Given a slice of *Unit called list sorted by X value and a *Unit u,
// binary search to find the *Unit closest to u in list based on X value
func (p *Player) BinarySearchUnits(list []Unit, u Unit) Unit {
	// empty list
	if len(list) == 0 { return nil }

	start := 0
	end := len(list) - 1

	for start <= end {
		mid := (start + end) / 2

		// lower bound
		if mid == 0 {
			if len(list) == 1 {
				return list[0]
			} else {
				if intAbsDiff(list[0].X(), u.X()) < intAbsDiff(list[1].X(), u.X()) {
					return list[0]
				} else {
					return list[1]
				}
			}
		}

		// upper bound
		if mid == len(list) - 1 {
			return list[len(list) - 1]
		}

		// X value match exactly
		if list[mid].X() == u.X() || list[mid+1].X() == u.X() {
			return u
		}

		// range is good enough
		if list[mid].X() < u.X() && list[mid+1].X() > u.X() {
			if intAbsDiff(list[mid].X(), u.X()) < intAbsDiff(list[mid+1].X(), u.X()) {
				return list[mid]
			} else {
				return list[mid+1]
			}
		}

		// find the right range
		if list[mid].X() < u.X() {
			start = mid + 1
		} else {
			end = mid
		}
	}
	return nil
}

func getEuclidDist(unit1 Unit, unit2 Unit) float64 {
	return math.Pow(float64(unit1.X()-unit2.X()), 2) + math.Pow(float64(unit1.Y()-unit2.Y()), 2)
}

// assume that Top, Mid and Bot are sorted lists of Units
func (p *Player) FindClosestUnit(unit Unit) (Unit, float64) {
	var minUnit Unit
	var minDist float64

	if intAbsDiff(unit.Y(), TOPY) <= unit.Reach() { // GET RID OF BOILERPLATE EVENTUALLY
		found := p.BinarySearchUnits(p.Top, unit)
		if found != nil {
			dist := getEuclidDist(unit, found)
			if intAbsDiff(unit.X(), found.X()) <= unit.Reach() &&
				(minUnit == nil || minDist < dist) {
				minUnit = found
				minDist = dist
			}
		}
	}
	if intAbsDiff(unit.Y(), MIDY) <= unit.Reach() {
		found := p.BinarySearchUnits(p.Mid, unit)
		if found != nil {
			dist := getEuclidDist(unit, found)
			if intAbsDiff(unit.X(), found.X()) <= unit.Reach() &&
				(minUnit == nil || minDist < dist) {
				minUnit = found
				minDist = dist
			}
		}
	}
	if intAbsDiff(unit.Y(), BOTY) <= unit.Reach() {
		found := p.BinarySearchUnits(p.Bot, unit)
		if found != nil {
			dist := getEuclidDist(unit, found)
			if intAbsDiff(unit.X(), found.X()) <= unit.Reach() &&
				(minUnit == nil || minDist < dist) {
				minUnit = found
				minDist = dist
			}
		}
	}
	
	if minUnit != nil {
		minDist = math.Pow(float64(unit.X()-minUnit.X()), 2) + math.Pow(float64(unit.Y()-minUnit.Y()), 2)
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
func (p *Player) ExportJSON(buffer *bytes.Buffer) { // used for exporting to screen
	buffer.WriteString(fmt.Sprintf(`{"owner": %d, "income": %d, "bits": %d, `, p.owner, p.income, p.bits))
	buffer.WriteString(`"towers": [`)
	for index, element := range p.Towers {
		if element == nil {
			buffer.WriteString(`"nil"`)
		} else {
			element.ExportJSON(buffer)
		}
		if index != len(p.Towers)-1 {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString(`], "troops": [`)
	totalSize := len(p.Top) + len(p.Mid) + len(p.Bot)
	for i := 0; i < totalSize; i++ {
		if i < len(p.Top) {
			p.Top[i].ExportJSON(buffer)
		} else if i - len(p.Top) < len(p.Mid) {
			p.Mid[i - len(p.Top)].ExportJSON(buffer)
		} else {
			p.Bot[i - len(p.Top) - len(p.Mid)].ExportJSON(buffer)
		}
		if i != totalSize - 1 {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString(`], "mainTower": `)
	p.MainTower.ExportJSON(buffer)
	buffer.WriteString("}")
}

// iterates before anything happens, just a frame initialization stage
func (p *Player) PrepPlayer() {
	sort.Sort(SortByX(p.Top))
	sort.Sort(SortByX(p.Mid))
	sort.Sort(SortByX(p.Bot))
	for _, element := range p.Towers { // pre-prep phase: reenable all towers
		if element == nil {
			continue
		}
		element.SetEnabled(p, true)
	}
}

// helper function that preps all units in a lane
func (p *Player) prepLane(other *Player, lane []Unit, frame int64) {
	for _, element := range lane {
		if element.Speed() == 0 || frame%int64(element.Speed()) == 0 {
			element.Prep(p, other)
		}
	}
}

// iterates over each of a player's units to see whether they should shoot or move.
// if they have a target and it's valid, they'll shoot at it
// if they have an invalid target, but find a new valid one, they'll shoot at it
// else, they'll move.
// this function call does not actually trigger shooting or moving, this just sets the "target" ptr of each unit.
func (p *Player) PrepUnits(other *Player, frame int64) {	
	p.prepLane(other, p.Top, frame)
	p.prepLane(other, p.Mid, frame)
	p.prepLane(other, p.Bot, frame)

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

// helper function that iterates all the troops in a lane
func (p *Player) iterateLane(other *Player, lane []Unit, frame int64) {
	for _, element := range lane {
		if element.Speed() == 0 || frame%int64(element.Speed()) == 0 {
			element.Iterate(p, other)
		}
	}
}

// iterates over each of a player's units and shoots at the unit's set target or move accordingly
func (p *Player) IterateUnits(other *Player, frame int64) {
	p.iterateLane(other, p.Top, frame)
	p.iterateLane(other, p.Mid, frame)
	p.iterateLane(other, p.Bot, frame)

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

/*
BELOW:
Functions that pertain to the cleanup of units in the game, this step happens at the end of each frame.
Troops/towers that die (go below 0 HP) do not die until this phase, which is when everything is cleaned at once.
*/

func (p *Player) triggerTroopDeath(other *Player, lane []Unit) []Unit {
	for _, element := range lane { // first pass, let the dead units have their death
		if element.HP() <= 0 {
			element.Die(p, other) // we iterate twice because sometimes the length of p.Units changes in Die()
		}
	}
	return lane
}

func (p *Player) removeDeadTroops(lane []Unit) []Unit {
	alive := 0                        // number of alive units
	for _, element := range lane { // second pass, remove dead units
		if element.HP() > 0 {
			lane[alive] = element
			alive++
		}
	}
	lane = lane[:alive] // delete dead units, but (TODO) i suspect these are still in the memory!!
	return lane
}

// units with <=0 hp don't die until this step, they are cleaned up here.
func (p *Player) UnitCleanup(other *Player) {
	p.Top = p.triggerTroopDeath(other, p.Top)
	p.Mid = p.triggerTroopDeath(other, p.Mid)
	p.Bot = p.triggerTroopDeath(other, p.Bot)
	p.Top = p.removeDeadTroops(p.Top)
	p.Mid = p.removeDeadTroops(p.Mid)
	p.Bot = p.removeDeadTroops(p.Bot)
	
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
