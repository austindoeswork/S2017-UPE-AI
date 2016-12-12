package main

import (
	"flag"
	"fmt"

	"github.com/austindoeswork/S2017-UPE-AI/server"
)

var (
	// TODO use a config file
	staticDir = flag.String("static", "./static/", "directory of static files")
)

func main() {
	flag.Parse()
	s := server.New(":8080", *staticDir)
	fmt.Println("server starting 8080")
	s.Start()
}
