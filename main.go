package main

import (
	"flag"
	"fmt"

	// "github.com/austindoeswork/tower_game/game"
	"github.com/austindoeswork/tower_game/server"
)

var (
	staticDir = flag.String("static", "/etc/tgww/", "directory of static files")
)

func main() {
	flag.Parse()
	s := server.New(":8080", *staticDir)
	fmt.Println("server starting 8080")
	s.Start()
}

//Game
//Core loop
//1 poll input
//2 calc gamestate
//3 output

//
