package game

import (
	"fmt"
	"time"
)

type Game struct {
	//chan input
	//chan output //pushes gamestate at framerate?
	//players
	framerate int
	board     *Board
}

func New(framerate int) *Game {
	return &Game{
		framerate: framerate,
		board: &Board{
			name:  "jerry",
			sizex: 50,
			sizey: 50,
		},
	}
}

func (g *Game) Start() {
	microPerFrame := 1000 * 1000 / (g.framerate)
	frameDuration := time.Microsecond * time.Duration(microPerFrame)
	for {
		starttime := time.Now()
		//get input
		//update game
		//push output
		elapsed := time.Now().Sub(starttime)
		delta := frameDuration.Nanoseconds() - elapsed.Nanoseconds()
		if delta > 0 {
			time.Sleep(time.Duration(delta))
		}
	}
}

type Board struct {
	name   string
	sizex  int
	sizey  int
	actors []*Actor
	towers []*Tower
}

type Tower struct {
	name string
	x    int
	y    int
}

type Actor struct {
	name string
	x    int
	y    int
	// speed float64 //squares/second
}

func (a *Actor) Position() (int, int) {
	return a.x, a.y
}
func (a *Actor) SetPosition(x, y int) {
	a.x = x
	a.y = y
}
func (a *Actor) Name() string {
	return a.name
}

func (b *Board) PPrint() {
	fmt.Printf("Board: %s\n", b.name)
	for _, actor := range b.actors {
		actorx, actory := actor.Position()
		fmt.Printf("Actor: %s %d,%d\n", actor.Name(), actorx, actory)
	}
}

func (b *Board) AddActor(a *Actor) {
	b.actors = append(b.actors, a)
}
