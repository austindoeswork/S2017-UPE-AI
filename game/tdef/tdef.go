package tdef

import (
	// "errors"
	"fmt"
	"log"
	"math"
	"time"
)

// TODO find a better place for this
const (
	NOTREADY = 0 // not enough players / not setup
	READY    = 1 // ready to start
	RUNNING  = 2 // playing now
	DONE     = 3 // done, clean me up
)

type Unit struct {
	owner  int // which player owns this? 1 or 2
	x      int // no y currently, one lane
	damage int // damage it deals in an attack
	maxhp  int // total hp, doesnt change
	hp     int // current hp
	speed  int // every <speed> frames, this unit acts (higher speed is slower)
	stride int // when moving, this unit moves <stride> pixels
	reach  int // range of a unit (rename?)
}

func (p *Unit) ExportJSON() string {
	return fmt.Sprintf(`{"owner": %d, "x": %d, "maxhp": %d, "hp": %d}`, p.owner, p.x, p.maxhp, p.hp)
}

func (p *Unit) Owner() int {
	return p.owner
}
func (p *Unit) Position() int {
	return p.x
}
func (p *Unit) SetPosition(x int) {
	p.x = x
}
func (p *Unit) Damage() int {
	return p.damage
}
func (p *Unit) MaxHP() int {
	return p.maxhp
}
func (p *Unit) HP() int {
	return p.hp
}
func (p *Unit) SetHP(hp int) {
	p.hp = hp
}
func (p *Unit) Stride() int {
	return p.stride
}

/* func (p *Unit) PercentHP() float32 {
	return 1.0 * p.hp / p.maxhp
} */
func (p *Unit) Reach() int {
	return p.reach
}

// this is just temporary, everything so far is a "unit" that can move and everything, although towers are
// units with 0 stride
func NewUnit(owner int, x int) *Unit {
	return &Unit{
		owner:  owner,
		x:      x,
		speed:  1,
		damage: 10,
		hp:     100,
		maxhp:  100,
		stride: 1, // updateGrid automatically handles owners' units moving in opposite dirs
		reach:  3,
	}
}

func NewTower(owner int, x int) *Unit {
	return &Unit{
		owner:  owner,
		x:      x,
		damage: 40,
		maxhp:  1000,
		hp:     1000,
		speed:  5,
		stride: 0,
		reach:  5,
	}
}

type TowerDefense struct {
	p1input chan []byte
	p2input chan []byte
	p1cmd   []byte
	p2cmd   []byte
	output  chan []byte //pushes gamestate at framerate
	quit    chan bool
	status  int

	p1    *Unit // TODO rename, these represent towers
	p2    *Unit
	units []*Unit

	width  int
	height int
	fps    int
	frame  int64

	winner int
}

func New(width, height, fps int) (*TowerDefense, []chan<- []byte, <-chan []byte) {
	outputChan := make(chan []byte)
	p1 := NewTower(1, 0)
	p2 := NewTower(2, width-1)
	units := []*Unit{p1, p2}
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
		units:   units,
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
				p.updateInputs()

				// if p.p1cmd != nil {
				// log.Println("1", p.p1cmd)
				// }
				// if p.p2cmd != nil {
				// log.Println("2", p.p2cmd)
				// }

				p.updateGame()

				p.p1cmd = nil
				p.p2cmd = nil

				select {
				case p.output <- p.stateJSON(): //send output
				default:
				}
				if p.p1.HP() <= 0 || p.p2.HP() <= 0 {
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

func (p *TowerDefense) Quit() {
	log.Println("ABORTING GAME...")
	p.status = DONE
	p.quit <- true
}
func (p *TowerDefense) Status() int {
	return p.status
}
func (p *TowerDefense) Winner() int {
	return p.winner
}
func (p *TowerDefense) MinPlayers() int {
	return 2
}

// temporary N^2
// TODO, make it like a heap
func (p *TowerDefense) updateGame() {
	p1string := string(p.p1cmd)
	p2string := string(p.p2cmd)
	controlPlayer(p, p1string, 1)
	controlPlayer(p, p2string, 2)

	for index, element := range p.units {
		if element.speed == 0 || p.frame%int64(element.speed) == 0 {
			// first check to see if there is an enemy unit in range
			move := true
			closest_enemy_dist := -1
			closest_enemy_index := -1
			for index2, element2 := range p.units { // O(n^2) check, will need an interesting algorithm because simple sort may not suffice
				if element.Owner() != element2.Owner() {
					dist := int(math.Abs(float64(element2.Position() - element.Position())))
					if dist <= element.Reach() {
						move = false
						if closest_enemy_dist == -1 || dist <= closest_enemy_dist {
							closest_enemy_dist = dist
							closest_enemy_index = index2
						}
					}
				}
			}
			// TODO THERE IS A MASSIVE RACE CONDITION HERE:
			// assume two units have identical range, if they are range+1 pixels apart, the first to reach this code will not
			// fire, not seeing anything in range. it will move, and the other will shoot. this is bad behavior because it
			// depends on the create time of the actual unit, which is dumb af
			if move { // didn't find a target, let's move
				if element.Owner() == 1 {
					element.SetPosition(element.Position() + element.Stride())
				} else {
					element.SetPosition(element.Position() - element.Stride())
				}
			} else { // found a target, fire
				p.units[closest_enemy_index].SetHP(p.units[closest_enemy_index].HP() - element.Damage())
				if p.units[closest_enemy_index].HP() < 0 { // THERE SHOULD BE A STEP WHERE DEATHS ARE RESOLVED
					p.units = append(p.units[:closest_enemy_index], p.units[closest_enemy_index+1:]...)
				}
			}
			if 0 > element.Position() || element.Position() > p.width { // temporary
				p.units = append(p.units[:index], p.units[index+1:]...)
			}
		}
	}
}

func controlPlayer(tdef *TowerDefense, input string, playernum int) {
	switch input {
	case "up":
		if playernum == 1 {
			tdef.units = append(tdef.units, NewUnit(1, 0))
		} else {
			tdef.units = append(tdef.units, NewUnit(2, tdef.width-1))
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
	outString += `"p1":` + p.p1.ExportJSON() + `, "p2":` + p.p2.ExportJSON() + `,`
	outString += `"units": [`
	for index, element := range p.units {
		outString += element.ExportJSON()
		if index != len(p.units)-1 {
			outString += ","
		}
	}
	outString += "]}"
	return []byte(outString)
}
