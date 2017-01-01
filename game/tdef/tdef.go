// TODO: split tdef.go into unit.go, player.go, etc

package tdef

import (
	// "errors"
	"fmt"
	"log"
	"math"
	"time"
)

// TODO find a better place for this too, is this necessary?
// std lib doesn't seem to have an int abs, converting to/from float64 seems unnecessary
func intAbsDiff(x, y int) int {
	if x >= y {
		return x - y
	} else {
		return y - x
	}
}

// TODO find a better place for this
const (
	NOTREADY = 0 // not enough players / not setup
	READY    = 1 // ready to start
	RUNNING  = 2 // playing now
	DONE     = 3 // done, clean me up

	GAMEWIDTH  = 800 // temporarily hardcoded until i figure out a work around (DD)
	GAMEHEIGHT = 500
	TOPY       = GAMEHEIGHT * 3 / 4 // y coordinate of top lane
	MIDY       = GAMEHEIGHT / 2     // ditto above but for mid
	BOTY       = GAMEHEIGHT / 4     // ditto
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
		stride: 1, // updateGrid automatically handles owners' units moving in opposite dirs
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

type Player struct {
	owner  int // who owns this, player 1 or 2?
	income int // X coins per second
	coins  int // total number of coins

	MainTower *Unit // if this dies you die

	Units []*Unit // list of all units
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

// unitEnum will eventually be a list of some sort
func (p *Player) BuyUnit(x, lane int) bool {
	if p.coins >= 100 {
		p.AddUnit(NewUnit(x, lane))
		p.coins -= 100
		return true
	}
	return false
}

func (p *Player) AddUnit(unit *Unit) {
	p.Units = append(p.Units, unit)
}

func NewPlayer(owner int) *Player {
	var x int
	switch owner {
	case 1:
		x = 0
	case 2:
		x = GAMEWIDTH - 1
	}
	mainTower := NewTower(x, 2) // need to figure out where maintowers belong, temporarily on midlane

	return &Player{
		owner:     owner,
		income:    500,
		coins:     0,
		MainTower: mainTower,
		Units:     make([]*Unit, 0),
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
}

func (p *Player) IsAlive() bool { // checks life of main tower
	return p.MainTower.HP() > 0
}

type TowerDefense struct {
	p1input chan []byte
	p2input chan []byte
	p1cmd   []byte
	p2cmd   []byte
	output  chan []byte //pushes gamestate at framerate
	quit    chan bool
	status  int

	p1 *Player
	p2 *Player

	width  int
	height int
	fps    int
	frame  int64

	winner int
}

func New(width, height, fps int) (*TowerDefense, []chan<- []byte, <-chan []byte) {
	outputChan := make(chan []byte)
	p1 := NewPlayer(1)
	p2 := NewPlayer(2)
	p1input := make(chan []byte, 5)
	p2input := make(chan []byte, 5)
	return &TowerDefense{
		p1input: p1input,
		p2input: p2input,
		output:  outputChan,
		quit:    make(chan bool),
		status:  READY,
		p1:      p1,
		p2:      p2,
		width:   width,
		height:  height,
		fps:     fps,
		frame:   0,
		winner:  -1,
	}, []chan<- []byte{p1input, p2input}, outputChan
}

func (p *TowerDefense) Start() error {
	if p.status == RUNNING {
		return fmt.Errorf("ERR game already running")
	}
	p.status = RUNNING

	frameNS := time.Duration(int(1e9) / p.fps)
	clk := time.NewTicker(frameNS)
	go func() {
		log.Println("GAME STARTED YOO")
		for {
			select {
			case <-clk.C: //nxt frame
				if p.status == DONE {
					log.Println("GAME DIED OF UNNATURAL CAUSES")
					return
				}
				p.frame++
				// fmt.Printf("\ncalculated: %d", p.frame)
				p.updateInputs()

				/* if p.p1cmd != nil {
					log.Println("1", p.p1cmd)
				}
				if p.p2cmd != nil {
					log.Println("2", p.p2cmd)
				} */

				p.updateGame()

				p.p1cmd = nil
				p.p2cmd = nil

				select {
				case p.output <- p.stateJSON(): //send output
				default:
				}
				if !p.p1.IsAlive() || !p.p2.IsAlive() {
					p.status = DONE
					log.Println("GAME DIED OF NATURAL CAUSES")
					close(p.output)
					return
				}
			}
		}
	}()
	return nil
}

func (t *TowerDefense) Quit() {
	log.Println("ABORTING GAME...")
	t.status = DONE
	t.quit <- true
}
func (t *TowerDefense) Status() int {
	return t.status
}
func (t *TowerDefense) Winner() int {
	return t.winner
}
func (t *TowerDefense) MinPlayers() int {
	return 2
}

func (t *TowerDefense) updateGame() {
	// award coins to people (note that frame 1 is the first frame of the game)
	if t.frame%int64(t.fps) == 1 {
		t.p1.SetCoins(t.p1.Coins() + t.p1.Income())
		t.p2.SetCoins(t.p2.Coins() + t.p2.Income())
	}

	p1string := string(t.p1cmd)
	p2string := string(t.p2cmd)
	controlPlayer(t, p1string, 1)
	controlPlayer(t, p2string, 2)
	t.p1.SetUnitTargets(t.p2, t.frame)
	t.p2.SetUnitTargets(t.p1, t.frame)
	t.p1.IterateUnits(t.frame)
	t.p2.IterateUnits(t.frame)
	t.p1.UnitCleanup()
	t.p2.UnitCleanup()
}

func controlPlayer(tdef *TowerDefense, input string, playernum int) {
	switch input {
	case "up":
		if playernum == 1 {
			tdef.p1.BuyUnit(0, 1)
		} else {
			tdef.p2.BuyUnit(tdef.width-1, 1)
		}
	case "right":
		if playernum == 1 {
			tdef.p1.BuyUnit(0, 2)
		} else {
			tdef.p2.BuyUnit(tdef.width-1, 2)
		}
	case "down":
		if playernum == 1 {
			tdef.p1.BuyUnit(0, 3)
		} else {
			tdef.p2.BuyUnit(tdef.width-1, 3)
		}
	}
}

func (p *TowerDefense) updateInputs() {
	p1done := false
	p2done := false
	for !p1done {
		select {
		case cmd := <-p.p1input:
			p.p1cmd = cmd
		default:
			p1done = true
		}
	}
	for !p2done {
		select {
		case cmd := <-p.p2input:
			p.p2cmd = cmd
		default:
			p2done = true
		}
	}
}

func (p *TowerDefense) stateJSON() []byte {
	outString := fmt.Sprintf(`{ "w": %d, "h": %d, `, p.width, p.height)
	outString += `"p1":` + p.p1.ExportJSON() + `, "p2":` + p.p2.ExportJSON() + "}"
	return []byte(outString)
}
