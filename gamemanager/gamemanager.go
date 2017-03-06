package gamemanager

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/austindoeswork/S2017-UPE-AI/game"
)

type GM interface {
	// NewGame creates a new game object and starts it when it has
	// enough controllers
	NewGame(gameName string, demoGame bool) error

	// ControlGame returns the input channel to a game if it exists
	ControlGame(gameName string, quit chan bool) (chan<- []byte, error)

	// WatchGame returns the output channel to a game if it exists
	WatchGame(gameName string, quit chan bool) (<-chan []byte, error)
}

type GameManager struct {
	mux       *sync.Mutex
	games     map[string]*GameWrapper
	opengames []string
}

func New() *GameManager {
	gm := &GameManager{
		mux:       &sync.Mutex{},
		games:     make(map[string]*GameWrapper),
		opengames: []string{},
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
			gm.mux.Lock()
			for gameName, gw := range gm.games {
				if gw.Status() == game.DONE {
					count++
					delete(gm.games, gameName)
				}
			}
			gm.mux.Unlock()
			log.Printf("janitor: cleaned %d/%d games yo.", count, total)
		}
	}
}

func (gm *GameManager) HasGame(gameName string) bool {
	gm.mux.Lock()
	_, exists := gm.games[gameName]
	gm.mux.Unlock()
	return exists
}

func (gm *GameManager) NewGame(gameName string, demoGame bool) error {
	gm.mux.Lock()
	defer gm.mux.Unlock()
	if _, exists := gm.games[gameName]; exists {
		return fmt.Errorf("ERR game already exists")
	}
	gw := NewGameWrapper(demoGame)
	gm.games[gameName] = gw

	go func() {
		time.AfterFunc(time.Second*60, func() {
			if gw.Status() != game.RUNNING && gw.Status() != game.DONE {
				log.Printf("game: %s timed out.", gameName)
				gw.Quit()
			}
		})
	}()
	return nil
}

func (gm *GameManager) PopOpenGame() (string, error) {
	if len(gm.opengames) == 0 {
		return "", fmt.Errorf("no open games")
	}
	name := gm.opengames[0]

	gm.mux.Lock()
	defer gm.mux.Unlock()

	gm.opengames = append(gm.opengames[:0], gm.opengames[1:]...)
	if _, ok := gm.games[name]; !ok {
		return "", fmt.Errorf("error opening game")
	}
	if gm.games[name].Status() > game.READY {
		return gm.PopOpenGame()
	}
	return name, nil
}

func (gm *GameManager) NewOpenGame() (string, error) {
	rint := rand.Int()
	rstr := strconv.Itoa(rint)

	err := gm.NewGame(rstr, false)
	if err != nil {
		return "", err
	}
	gm.opengames = append(gm.opengames, rstr)
	return rstr, nil
}

type GameInfo struct {
	Name    string
	Players []string
}

func (gm *GameManager) ListGames() []*GameInfo {
	list := []*GameInfo{}
	gm.mux.Lock()
	defer gm.mux.Unlock()
	for name, g := range gm.games {
		if g.Status() == game.RUNNING {
			ginfo := &GameInfo{name, g.PlayerNames()}
			list = append(list, ginfo)
		}
	}
	return list
}

func (gm *GameManager) ControlGame(gameName string, userName string, quit chan bool) (game.Controller, error) {
	gm.mux.Lock()
	defer gm.mux.Unlock()

	gw, exists := gm.games[gameName]
	if !exists {
		return nil, fmt.Errorf("ERR no such game")
	}
	if gw.Status() == game.DONE {
		return nil, fmt.Errorf("ERR game over")
	}

	controller, err := gw.getOpenController(userName)
	if err != nil {
		return nil, err
	}

	fmt.Println("setting player name", controller.Player(), userName)
	gw.SetPlayerName(controller.Player(), userName)

	go func() {
		select {
		case <-quit:
			log.Printf("game: %s %s controller QUIT", gameName, userName)
			err := gw.closeController(controller)
			if err != nil {
				log.Println(err)
			}
			return
		}
	}()

	log.Printf("game: %s %s controller GIVEN", gameName, userName)
	return controller, nil
}

func (gm *GameManager) WatchGame(gameName string, quit chan bool) (<-chan []byte, error) {
	gm.mux.Lock()
	defer gm.mux.Unlock()
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

//////////////////////////////////

type GameWrapper struct {
	game.Game
	// TODO think about resetting connections
	// gameInput maps an input to whether they are connected
	gameControllerMap map[game.Controller]string
	activeControllers int
	gameOutput        <-chan []byte
	listenerMap       map[chan []byte]bool
}

// TODO allow creation of different games (pong, scrabble, whatever)
func NewGameWrapper(isdemo bool) *GameWrapper {
	g, controllers, output := game.NewTowerDef(isdemo)
	gameControllerMap := make(map[game.Controller]string)
	listenerMap := make(map[chan []byte]bool)

	for _, c := range controllers {
		gameControllerMap[c] = ""
	}

	gw := &GameWrapper{
		g,
		gameControllerMap,
		0,
		output,
		listenerMap,
	}

	go gw.multiplex() // start sending output to listeners
	return gw
}

func (gw *GameWrapper) PlayerNames() []string {
	res := []string{}
	for _, name := range gw.gameControllerMap {
		res = append(res, name)
	}
	return res
}

func (gw *GameWrapper) multiplex() {
	for {
		select {
		case msg, more := <-gw.gameOutput:
			if more {
				for listener := range gw.listenerMap {
					select {
					case listener <- msg: // dont block on a listener
					default:
					}
				}
			} else {
				log.Println("stopping the multiplex")
				for listener := range gw.listenerMap {
					close(listener)
				}
				return
			}
		}
	}
}

func (gw *GameWrapper) Ready() bool {
	return gw.activeControllers == gw.MinPlayers()
}

// getOpenController returns the first open controller it encounters
func (gw *GameWrapper) getOpenController(userName string) (game.Controller, error) {
	if userName == "" {
		return nil, fmt.Errorf("ERR invalid username")
	}
	for c, currentUser := range gw.gameControllerMap {
		if currentUser == "" {
			gw.gameControllerMap[c] = userName
			gw.activeControllers++
			if gw.activeControllers == gw.MinPlayers() {
				gw.Start()
			}
			return c, nil
		}
	}
	return nil, fmt.Errorf("ERR no open controller")
}

func (gw *GameWrapper) closeController(c game.Controller) error {
	if _, exists := gw.gameControllerMap[c]; !exists {
		return fmt.Errorf("ERR no such input chan")
	}
	gw.activeControllers--
	if gw.activeControllers == 0 && gw.Status() != game.DONE {
		gw.Quit()
	}
	gw.gameControllerMap[c] = ""
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
	// TODO austin add a mutex
	delete(gw.listenerMap, output)
	return nil
}
