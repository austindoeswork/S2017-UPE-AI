package game

import (
	"testing"
	// "github.com/austindoeswork/tower_game/game"
)

func TestBoard(t *testing.T) {
	b := Board{
		name:  "testboard",
		sizex: 15,
		sizey: 15,
	}
	b.AddActor(&Actor{"testactor", 5, 5})
	b.PPrint()

	g := New(30)
	g.Start()
}
