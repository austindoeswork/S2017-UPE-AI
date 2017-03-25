package tdef

import (
	"bytes"
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

func NewOutputQueue(size, minpop int) *OutputQueue {
	return &OutputQueue{
		nodes:  make([][]byte, size),
		size:   size,
		minpop: minpop,
	}
}

// Queue is a basic FIFO queue based on a circular list that resizes as needed.
type OutputQueue struct {
	nodes [][]byte
	size  int
	head  int
	tail  int
	count int

	minpop int
}

// Push adds a node to the queue.
func (q *OutputQueue) Push(n []byte) {
	if q.head == q.tail && q.count > 0 {
		nodes := make([][]byte, len(q.nodes)+q.size)
		copy(nodes, q.nodes[q.head:])
		copy(nodes[len(q.nodes)-q.head:], q.nodes[:q.head])
		q.head = 0
		q.tail = len(q.nodes)
		q.nodes = nodes
	}
	q.nodes[q.tail] = n
	q.tail = (q.tail + 1) % len(q.nodes)
	q.count++
}

// Pop removes and returns a node from the queue in first to last order.
func (q *OutputQueue) Pop() []byte {
	if q.count < q.minpop {
		return nil
	}
	node := q.nodes[q.head]
	q.head = (q.head + 1) % len(q.nodes)
	q.count--
	return node
}

type TowerDefense struct {
	p1input  chan []byte
	p2input  chan []byte
	p1output chan []byte
	p2output chan []byte

	p1cmd  []byte
	p2cmd  []byte
	output chan []byte //pushes gamestate at framerate
	oq     *OutputQueue
	quit   chan bool
	status int

	players [2]*Player // in the future perhaps make this another const: NUMPLAYERS

	width  int
	height int
	fps    int
	frame  int64

	winner   int
	demoGame bool
}

func New(width, height, fps int, demoGame bool) (*TowerDefense, []*Controller, <-chan []byte) {
	outputChan := make(chan []byte)
	p1 := NewPlayer(1, "", demoGame)
	p2 := NewPlayer(2, "", demoGame)
	p1input := make(chan []byte, 5)
	p2input := make(chan []byte, 5)
	p1output := make(chan []byte, 5)
	p2output := make(chan []byte, 5)
	p1controller := &Controller{
		1,
		p1input,
		p1output,
	}
	p2controller := &Controller{
		2,
		p2input,
		p2output,
	}
	return &TowerDefense{
		p1input:  p1input,
		p2input:  p2input,
		p1output: p1output,
		p2output: p2output,
		output:   outputChan,
		oq:       NewOutputQueue(600, 500),
		quit:     make(chan bool),
		status:   READY,
		players:  [2]*Player{p1, p2},
		width:    width,
		height:   height,
		fps:      fps,
		frame:    0,
		winner:   -1,
		demoGame: demoGame,
	}, []*Controller{p1controller, p2controller}, outputChan
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

func (t *TowerDefense) SetPlayerName(num int, name string) error {
	if num <= 0 || num >= 3 {
		return fmt.Errorf("invalid player num")
	}
	t.players[num-1].SetName(name)
	return nil
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
				t.updateGame()

				t.p1cmd = nil
				t.p2cmd = nil

				// sends fogged output
				p0minX, p0maxX := t.players[0].Horizon()
				p1minX, p1maxX := t.players[1].Horizon()
				select {
				case t.p1output <- t.stateJSON(p0minX, p0maxX):
				default:
				}
				select {
				case t.p2output <- t.stateJSON(p1minX, p1maxX):
				default:
				}

				//send delayed output
				t.sendWatcher(t.stateJSON(0, GAMEWIDTH))

				if t.demoGame == false &&
					(!t.players[0].IsAlive() || !t.players[1].IsAlive() || t.frame == int64(t.fps*300)) {
					t.DetermineWinner()
				}

				if !t.players[0].IsAlive() || !t.players[1].IsAlive() || t.Winner() != -1 {
					t.sendWatcher(t.stateJSON(0, GAMEWIDTH))
					t.sendWatcher(t.stateJSON(0, GAMEWIDTH))
					t.sendWatcher(t.stateJSON(0, GAMEWIDTH))
					for i := 0; i < 500; i++ {
						t.oq.Push(nil)
					}
					for fr := t.oq.Pop(); fr != nil; fr = t.oq.Pop() {
						select {
						case <-clk.C:
							t.sendWatcher(fr)
						}
					}
					time.Sleep(time.Second)
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
		player.BuyTroop(lane, unitEnum, tdef.players[playernum%2])
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

func (t *TowerDefense) sendWatcher(m []byte) {
	t.oq.Push(m)
	out := t.oq.Pop()
	if out != nil {
		select {
		case t.output <- out: //send output
		default:
		}
	}

}

// generates stateJSON string with limits
func (t *TowerDefense) stateJSON(minX, maxX int) []byte {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(`{"w":%d,"h":%d,"f":%d,`, t.width, t.height, t.frame))
	buffer.WriteString(`"p1":`)
	t.players[0].ExportJSON(&buffer, minX, maxX)
	buffer.WriteString(`,"p2":`)
	t.players[1].ExportJSON(&buffer, minX, maxX)
	buffer.WriteString("}")
	return buffer.Bytes()
}
