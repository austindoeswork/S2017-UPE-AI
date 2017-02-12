package tdef

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

// TODO find a better place for this too, is this necessary?
// std lib doesn't seem to have an int abs, converting to/from float64 seems unnecessary
// this is used in player.go, unit.go
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

	GAMEWIDTH  = 1600 // temporarily hardcoded until i figure out a work around (DD)
	GAMEHEIGHT = 600

	PLOTBUFFER = 300 // how far away from spawn do plots start
	LANEWIDTH  = 50  // towers spawn LANEWIDTH above and below each lane
	PLOTWIDTH  = 100 // width of each plot
	NUMPLOTS   = ((GAMEWIDTH-2*PLOTBUFFER)/PLOTWIDTH + 1) * 6

	TOPY    = 475 // y coordinate of top lane
	MIDY    = 280 // ditto above but for mid
	BOTY    = 100 // ditto
	XOFFSET = 200 // used for x-positioning of lane objectives
)

// calculates where plot X will be, -1 -1 if not a valid plot
func getPlotPosition(plot int) (int, int) {
	if plot < 0 || plot >= NUMPLOTS {
		return -1, -1
	}
	plotLane := plot / (NUMPLOTS / 6)
	var x, y int
	if plotLane < 2 {
		y = TOPY
	} else if plotLane < 4 {
		y = MIDY
	} else if plotLane < 6 {
		y = BOTY
	}
	if plotLane%2 == 0 {
		y += 50
	} else {
		y -= 50
	}

	x = PLOTBUFFER + plot%(NUMPLOTS/6)*PLOTWIDTH
	return x, y
}

type TowerDefense struct {
	p1input chan []byte
	p2input chan []byte
	p1cmd   []byte
	p2cmd   []byte
	output  chan []byte //pushes gamestate at framerate
	quit    chan bool
	status  int

	players [2]*Player // in the future perhaps make this another const: NUMPLAYERS

	width  int
	height int
	fps    int
	frame  int64

	winner   int
	demoGame bool
}

func New(width, height, fps int, demoGame bool) (*TowerDefense, []chan<- []byte, <-chan []byte) {
	outputChan := make(chan []byte)
	p1 := NewPlayer(1, demoGame)
	p2 := NewPlayer(2, demoGame)
	p1input := make(chan []byte, 5)
	p2input := make(chan []byte, 5)
	return &TowerDefense{
		p1input:  p1input,
		p2input:  p2input,
		output:   outputChan,
		quit:     make(chan bool),
		status:   READY,
		players:  [2]*Player{p1, p2},
		width:    width,
		height:   height,
		fps:      fps,
		frame:    0,
		winner:   -1,
		demoGame: demoGame,
	}, []chan<- []byte{p1input, p2input}, outputChan
}

func (t *TowerDefense) DetermineWinner() {
	if !t.players[0].IsAlive() {
		t.winner = 2
	} else if !t.players[1].IsAlive() {
		t.winner = 1
	} else {
		tiebreak1 := t.players[0].GetTiebreak()
		tiebreak2 := t.players[1].GetTiebreak()
		if tiebreak1 > tiebreak2 {
			t.winner = 1
		} else if tiebreak1 < tiebreak2 {
			t.winner = 2
		} else {
			t.winner = 0
		}
	}
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
				if t.demoGame == false &&
					(!t.players[0].IsAlive() || !t.players[1].IsAlive() || t.frame == int64(t.fps*300)) {
					t.DetermineWinner()
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
		for _, player := range t.players {
			player.SetBits(player.Bits() + player.Income())
		}
	}
	// First, get player's commands and interpret them
	p1string := string(t.p1cmd)
	p2string := string(t.p2cmd)
	controlPlayer(t, p1string, 1)
	controlPlayer(t, p2string, 2)
	// Then, prepare each player for the coming frame
	for _, player := range t.players {
		player.PrepPlayer()
	}
	// Then, each unit decides whether it's going to shoot or move
	// It doesn't actually do anything yet, this avoids race conditions
	for index, player := range t.players {
		player.PrepUnits(t.players[(index+1)%2], t.frame)
	}
	// Then, everything acts. Units with sub-zero HP do not die yet, and STILL act.
	for index, player := range t.players {
		player.IterateUnits(t.players[(index+1)%2], t.frame)
	}
	// Finally, units with sub-zero HP all are cleared out at once.
	for index, player := range t.players {
		player.UnitCleanup(t.players[(index+1)%2])
	}
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
*/
func controlPlayer(tdef *TowerDefense, input string, playernum int) {
	unitEnum, lane := interpretCommand(input) // only one unit type exists currently
	if unitEnum == 0 && lane == 0 {           // no move
		return
	}

	player := tdef.players[playernum-1]
	if unitEnum < 50 {
		player.BuyTroop(player.Spawns[lane-1], lane, unitEnum, tdef.players[playernum%2])
	} else {
		player.BuyTower(lane, unitEnum, tdef.players[playernum%2]) // note that lane for towers means plot
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
	outString += `"p1":` + t.players[0].ExportJSON() + `, "p2":` + t.players[1].ExportJSON() + "}"
	return []byte(outString)
}
