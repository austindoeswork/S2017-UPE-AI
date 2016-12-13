package game

import (
	"time"

	"testing"
	// "github.com/austindoeswork/tower_game/game"
)

func TestPong(t *testing.T) {
	p := New(100, 100, 10)
	_, err := p.Start()
	if err == nil {
		t.Fatal("Expected err")
	}
	p.AddPlayer()
	p.AddPlayer()
	p.Start()

	// time.Sleep(time.Minute)
}
