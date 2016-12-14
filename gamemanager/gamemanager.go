package gamemanager

import (
	"fmt"
	"log"
	"time"

	"github.com/austindoeswork/S2017-UPE-AI/game"
)

// GameManager holds games and maintains connections between the game and the players
type GameManager struct {
	games map[string]*gameWrapper
}

func New() *GameManager {
	gm := &GameManager{
		games: make(map[string]*gameWrapper),
	}

	go func() {
		clk := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-clk.C:
				log.Println("game janitor starting")
				for name, g := range gm.games {
					if g.Status() == game.DONE {
						log.Printf("cleaned: %s\n", name)
						delete(gm.games, name)
					}
				}
			}
		}
	}()

	return gm
}

// CONNECT creates a game if it doesn't exist
// returns an id, input and output channel
func (gm *GameManager) Connect(name string) (int, chan []byte, chan []byte, error) {
	g, exists := gm.games[name]
	if !exists {
		g = NewGameWrapper()
		gm.games[name] = g
	}

	id, inputChan, err := g.AddPlayer()
	if err != nil {
		return -1, nil, nil, err
	}

	outputChan := make(chan []byte)
	g.AddListener(id, outputChan)

	log.Printf("%s %d: added", name, id)
	if g.Status() == game.READY {
		err = g.Start()
		if err != nil {
			log.Fatal("GAME SAID IT WAS READY WHEN IT WASN'T")
		}
	}
	return id, inputChan, outputChan, nil
}

func (gm *GameManager) Disconnect(name string, id int) error {
	if g, exists := gm.games[name]; exists {
		return g.DeleteListener(id)
	}
	return fmt.Errorf("game %s DNE", name)

}

// gameWrapper handles output comms with the listeners
type gameWrapper struct {
	game.Game
	listeners map[int]chan []byte
}

// TODO think of a way to handle multiple types of games
func NewGameWrapper() *gameWrapper {
	g, gchan := game.NewPong()
	gw := &gameWrapper{
		g,
		make(map[int]chan []byte),
	}

	// mux output chan
	go func() {
		clk := time.NewTicker(10 * time.Second)
		for {
			if gw.Status() == game.RUNNING && len(gw.listeners) == 0 {
				gw.Quit()
				return
			}
			if gw.Status() == game.DONE {
				return
			}
			select {
			case out := <-gchan:
				gw.sendListeners(out)
			case <-clk.C: //reset loop
			}
		}
	}()

	return gw
}

func (gw *gameWrapper) AddListener(id int, listener chan []byte) {
	gw.listeners[id] = listener
}
func (gw *gameWrapper) DeleteListener(id int) error {
	if listener, exists := gw.listeners[id]; exists {
		close(listener)
		delete(gw.listeners, id)
		return nil
	}
	return fmt.Errorf("listener %d DNE", id)
}

func (gw *gameWrapper) sendListeners(msg []byte) {
	for _, listener := range gw.listeners {
		select {
		case listener <- msg:
		default:
		}
	}

}
