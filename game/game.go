package game

import (
	// "github.com/austindoeswork/S2017-UPE-AI/game/pong"
	"github.com/austindoeswork/S2017-UPE-AI/game/tdef"
)

// TODO cleanup
const (
	NOTREADY = 0 // not enough players / not setup
	READY    = 1 // ready to start
	RUNNING  = 2 // playing now
	DONE     = 3 // done, clean me up
)

// A game speaks to its players and watchers thru channels
// There can be multiple input channels
// For now, there is only one output channel
// These are returned when the game is created

// A game will close it's output socket when quit is called
// it will also close when the game ends

type Controller interface {
	Player() int
	Input() chan<- []byte
	Output() <-chan []byte
}

type Game interface {
	Start() error
	Quit()

	Status() int
	MinPlayers() int
}

func NewTowerDef(isdemo bool) (*tdef.TowerDefense, []*tdef.Controller, <-chan []byte) {
	p, inArr, out := tdef.New(1600, 600, 30, isdemo)
	return p, inArr, out
}

// uncomment and change in gamemanager.go to test with pong.go
/* func NewPong() (*pong.Pong, []chan<- []byte, <-chan []byte) {
	p, inArr, out := pong.New(30, 20, 100)
	return p, inArr, out
} */
