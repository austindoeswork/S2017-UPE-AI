package gamemanager

import (
	"fmt"
	"log"
	"time"

	"github.com/austindoeswork/S2017-UPE-AI/game"
)

type GM interface {
	// NewGame creates a new game object and starts it when it has
	// enough controllers
	NewGame(gameName string) error

	// ControlGame returns the input channel to a game if it exists
	ControlGame(gameName string, quit chan bool) (chan<- []byte, error)

	// WatchGame returns the output channel to a game if it exists
	WatchGame(gameName string, quit chan bool) (<-chan []byte, error)
}

type GameManager struct {
	games map[string]*GameWrapper
}

func New() *GameManager {
	gm := &GameManager{
		games: make(map[string]*GameWrapper),
	}
	go func() {
		gm.Janitor()
	}()
	return gm
}

func (gm *GameManager) Janitor() {
	clk := time.NewTicker(time.Second * 120)
	for {
		select {
		case <-clk.C:
			total := len(gm.games)
			count := 0
			for gameName, gw := range gm.games {
				if gw.Status() == game.DONE {
					count++
					delete(gm.games, gameName)
				}
			}
			log.Printf("janitor: cleaned %d/%d games yo.", count, total)
		}
	}
}

func (gm *GameManager) HasGame(gameName string) bool {
	_, exists := gm.games[gameName]
	return exists
}

func (gm *GameManager) NewGame(gameName string) error {
	if _, exists := gm.games[gameName]; exists {
		return fmt.Errorf("ERR game already exists")
	}
	gw := NewGameWrapper()
	gm.games[gameName] = gw

	go func() {
		time.AfterFunc(time.Second*10, func() {
			if gw.Status() != game.RUNNING && gw.Status() != game.DONE {
				log.Printf("game: %s timed out.", gameName)
				gw.Quit()
			}
		})
	}()
	return nil
}

func (gm *GameManager) ControlGame(gameName string, quit chan bool) (chan<- []byte, error) {
	gw, exists := gm.games[gameName]
	if !exists {
		return nil, fmt.Errorf("ERR no such game")
	}
	if gw.Status() == game.DONE {
		return nil, fmt.Errorf("ERR game over")
	}
	input, err := gw.getOpenInput()
	if err != nil {
		return nil, err
	}

	go func() {
		select {
		case <-quit:
			log.Printf("game: %s controller QUIT", gameName)
			err := gw.closeInput(input)
			if err != nil {
				log.Panic(err) // TODO idiomatic way to log unexpected errors?
			}
			return
		}
	}()

	log.Printf("game: %s controller GIVEN", gameName)
	return input, nil
}
func (gm *GameManager) WatchGame(gameName string, quit chan bool) (<-chan []byte, error) {
	gw, exists := gm.games[gameName]
	if !exists {
		return nil, fmt.Errorf("ERR no such game")
	}
	if gw.Status() == game.DONE {
		return nil, fmt.Errorf("ERR game over")
	}
	output := gw.getOutput()

	go func() {
		select {
		case <-quit:
			log.Printf("game: %s watcher QUIT", gameName)
			err := gw.closeOutput(output)
			if err != nil {
				log.Panic(err) // TODO idiomatic way to log unexpected errors?
			}
			return
		}
	}()

	log.Printf("game: %s watcher GIVEN", gameName)
	return output, nil
}

type GameWrapper struct {
	game.Game
	// TODO think about resetting connections
	// gameInput maps an input to whether they are connected
	gameInputMap map[chan<- []byte]bool
	activeInputs int
	gameOutput   <-chan []byte
	listenerMap  map[chan []byte]bool
}

// TODO allow creation of different games (pong, scrabble, whatever)
func NewGameWrapper() *GameWrapper {
	g, inputs, output := game.NewTowerDef()
	gameInputMap := make(map[chan<- []byte]bool)
	listenerMap := make(map[chan []byte]bool)

	for _, in := range inputs {
		gameInputMap[in] = false
	}

	gw := &GameWrapper{
		g,
		gameInputMap,
		0,
		output,
		listenerMap,
	}

	go gw.multiplex() // start sending output to listeners
	return gw
}

func (gw *GameWrapper) multiplex() {
	for {
		select {
		case msg, more := <-gw.gameOutput:
			if more {
				for listener, _ := range gw.listenerMap {
					select {
					case listener <- msg: // dont block on a listener
					default:
					}
				}
			} else {
				log.Println("stopping the multiplex")
				for listener, _ := range gw.listenerMap {
					close(listener)
				}
				return
			}
		}
	}
}

func (gw *GameWrapper) Ready() bool {
	return gw.activeInputs == gw.MinPlayers()
}

// getOpenInput returns the first open input chan it encounters
func (gw *GameWrapper) getOpenInput() (chan<- []byte, error) {
	for input, assigned := range gw.gameInputMap {
		if !assigned {
			gw.gameInputMap[input] = true
			gw.activeInputs++
			if gw.activeInputs == gw.MinPlayers() {
				gw.Start()
			}
			return input, nil
		}
	}
	return nil, fmt.Errorf("ERR no open input chan")
}
func (gw *GameWrapper) closeInput(input chan<- []byte) error {
	if _, exists := gw.gameInputMap[input]; !exists {
		return fmt.Errorf("ERR no such input chan")
	}
	gw.activeInputs--
	if gw.activeInputs == 0 && gw.Status() != game.DONE {
		gw.Quit()
	}
	gw.gameInputMap[input] = false
	return nil
}
func (gw *GameWrapper) getOutput() chan []byte {
	output := make(chan []byte)
	gw.listenerMap[output] = true

	return output
}
func (gw *GameWrapper) closeOutput(output chan []byte) error {
	if _, exists := gw.listenerMap[output]; !exists {
		return fmt.Errorf("ERR no such output chan")
	}
	delete(gw.listenerMap, output)
	return nil
}
