package gamemanager

import (
	"fmt"
	"log"
	"time"

	"github.com/austindoeswork/S2017-UPE-AI/game"
)

// GameManager holds games and maintains connections between the game and the players
type GameManager struct {
	games     map[string]*gameWrapper
	watcherID int // TODO figure out how to handle watcher identification vs player id
}

func New() *GameManager {
	gm := &GameManager{
		games:     make(map[string]*gameWrapper),
		watcherID: -1,
	}

	go func() {
		clk := time.NewTicker(1 * time.Minute)
		for {
			select {
			case <-clk.C:
				total := len(gm.games)
				count := 0
				for name, g := range gm.games {
					if g.Status() == game.DONE {
						log.Printf("cleaned: %s\n", name)
						delete(gm.games, name)
						count++
					}
				}
				log.Printf("game janitor: cleaned %d/%d games", count, total)
			}
		}
	}()

	return gm
}

// CONNECT creates a game if it doesn't exist
// returns an id, input and output channel
func (gm *GameManager) Connect(name string) (int, chan<- []byte, <-chan []byte, error) {
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

	log.Printf("%s %d: connected", name, id)
	if g.Status() == game.READY {
		err = g.Start()
		if err != nil {
			log.Fatal("GAME SAID IT WAS READY WHEN IT WASN'T")
		}
	}
	return id, inputChan, outputChan, nil
}

func (gm *GameManager) Watch(name string) (int, <-chan []byte, error) {
	g, exists := gm.games[name]
	if !exists {
		return -1, nil, fmt.Errorf("game %s DNE", name)
	}

	outputChan := make(chan []byte)
	id := gm.watcherID
	g.AddListener(id, outputChan)
	gm.watcherID--

	log.Printf("%s %d(watcher): connected", name, id)
	return id, outputChan, nil
}

func (gm *GameManager) Disconnect(name string, id int) error {
	if g, exists := gm.games[name]; exists {
		log.Printf("%s %d: disconnected", name, id)
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
		defer gw.cleanup()
		begin := time.Now()
		clk := time.NewTicker(10 * time.Second)
		for {
			if len(gw.listeners) == 0 {
				if gw.Status() == game.NOTREADY { // TODO think about where to put this
					if time.Now().Sub(begin).Seconds() > 10 {
						log.Println("NOTREADY timeout")
						gw.Quit()
						return
					}
				}
				if gw.Status() == game.RUNNING {
					gw.Quit()
					return
				}
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

func (gw *gameWrapper) cleanup() {
	for id, _ := range gw.listeners {
		gw.DeleteListener(id)
	}
}
