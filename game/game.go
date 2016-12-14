package game

import (
	"github.com/austindoeswork/S2017-UPE-AI/game/pong"
)

const (
	NOTREADY = 0 // not enough players / not setup
	READY    = 1 // ready to start
	RUNNING  = 2 // playing now
	DONE     = 3 // done, clean me up
)

type Game interface {
	AddPlayer() (int, chan []byte, error)
	Start() error
	Quit()

	Status() int
}

func NewPong() (*pong.Pong, chan []byte) {
	p, out := pong.New(30, 20, 60)
	return p, out
}
