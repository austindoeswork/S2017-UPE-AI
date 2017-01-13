/*
TODO:
Troops should have their own targetting and behaviors and such
Add deploy time
Add towers and plots
Add actual troop ideas
Change unit list to sorted tower/troop lanes
*/

package tdef

import (
	"fmt"
	"log"
	"strconv"
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
	GAMEHEIGHT = 600
	TOPY       = GAMEHEIGHT * 3 / 4 // y coordinate of top lane
	MIDY       = GAMEHEIGHT / 2     // ditto above but for mid
	BOTY       = GAMEHEIGHT / 4     // ditto
	XOFFSET    = GAMEWIDTH / 4      // used for x-positioning of lane objectives
)

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

func (t *TowerDefense) Start() error {
	if t.status == RUNNING {
		return fmt.Errorf("ERR game already running")
	}
	t.status = RUNNING

	frameNS := time.Duration(int(1e9) / t.fps)
	clk := time.NewTicker(frameNS)
	go func() {
		log.Println("GAME STARTED YOO")
		for {
			select {
			case <-clk.C: //nxt frame
				if t.status == DONE {
					log.Println("GAME DIED OF UNNATURAL CAUSES")
					return
				}
				t.frame++
				t.updateInputs()

				/* if t.p1cmd != nil {
					log.Println("1", t.p1cmd)
				}
				if t.p2cmd != nil {
					log.Println("2", t.p2cmd)
				} */

				t.updateGame()

				t.p1cmd = nil
				t.p2cmd = nil

				select {
				case t.output <- t.stateJSON(): //send output
				default:
				}
				if !t.p1.IsAlive() || !t.p2.IsAlive() {
					t.status = DONE
					log.Println("GAME DIED OF NATURAL CAUSES")
					close(t.output)
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

// updateGame() is called every frame
func (t *TowerDefense) updateGame() {
	// Note that the first frame that occurs is frame 1 (hence mod = 1 rather than 0)
	// Every second, award player's income to player's coins
	if t.frame%int64(t.fps) == 1 {
		t.p1.SetCoins(t.p1.Coins() + t.p1.Income())
		t.p2.SetCoins(t.p2.Coins() + t.p2.Income())
	}
	// First, get player's commands and interpret them
	p1string := string(t.p1cmd)
	p2string := string(t.p2cmd)
	controlPlayer(t, p1string, 1)
	controlPlayer(t, p2string, 2)
	// Then, each unit decides whether it's going to shoot or move
	// It doesn't actually do anything yet, this avoids race conditions
	t.p1.SetUnitTargets(t.p2, t.frame)
	t.p2.SetUnitTargets(t.p1, t.frame)
	// Then, everything acts. Units with sub-zero HP do not die yet, and STILL act.
	t.p1.IterateUnits(t.frame)
	t.p2.IterateUnits(t.frame)
	// Finally, units with sub-zero HP all are cleared out at once.
	t.p1.UnitCleanup()
	t.p2.UnitCleanup()
}

/*
Takes in a command, e.g. b01 01
And outputs <unit type>, <lane>

Remember: top/mid/bot = 1/2/3
No action (lane) = 0
*/
func interpretCommand(input string) (int, int) {
	if len(input) < 6 || input[0] != 'b' {
		return 0, 0
	}
	unitEnum, err := strconv.Atoi(input[1:3])
	if err != nil {
		return 0, 0
	}

	lane, err := strconv.Atoi(input[4:6])
	if err != nil {
		return 0, 0
	}
	return unitEnum, lane
}

/*
Called each turn by updateGame().

Currently, pressing up spawns a unit in top lane, right spawns in mid lane and down spawns
in bot lane. Of course, this is conditional in whether player has enough coins (handled in
BuyUnit()).

If you're going to extend this functionality by adding more keycontrols, note that you need to
first add the key to /static/js/tdef.js (tdef.js only sends certain keypresses to the server)
*/

func controlPlayer(tdef *TowerDefense, input string, playernum int) {
	unitEnum, lane := interpretCommand(input) // only one unit type exists currently
	if unitEnum == 0 && lane == 0 {           // no move
		return
	}
	if playernum == 1 {
		if unitEnum != 10 {
			tdef.p1.BuyUnit(0, lane, unitEnum)
		} else {
			tdef.p1.BuyTower(lane, unitEnum) // note that lane for towers means plot
		}
	} else {
		if unitEnum != 10 {
			tdef.p2.BuyUnit(tdef.width-1, lane, unitEnum)
		} else {
			tdef.p2.BuyTower(lane, unitEnum) // note that lane for towers means plot
		}
	}
}

func (t *TowerDefense) updateInputs() {
	p1done := false
	p2done := false
	for !p1done {
		select {
		case cmd := <-t.p1input:
			t.p1cmd = cmd
		default:
			p1done = true
		}
	}
	for !p2done {
		select {
		case cmd := <-t.p2input:
			t.p2cmd = cmd
		default:
			p2done = true
		}
	}
}

func (t *TowerDefense) stateJSON() []byte {
	outString := fmt.Sprintf(`{ "w": %d, "h": %d, `, t.width, t.height)
	outString += `"p1":` + t.p1.ExportJSON() + `, "p2":` + t.p2.ExportJSON() + "}"
	return []byte(outString)
}
