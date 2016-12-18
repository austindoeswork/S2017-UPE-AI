package main

import (
	"flag"
	"fmt"
	_ "net/http/pprof"

	"github.com/austindoeswork/S2017-UPE-AI/server"
)

var (
	// TODO use a config file
	staticDir   = flag.String("static", "./static/", "directory of static files")
	versionFlag = flag.Bool("v", false, "git commit hash")

	commithash string
)

func main() {
	flag.Parse()
	if *versionFlag {
		fmt.Println(commithash)
		return
	}
	fmt.Println("version: " + commithash)

	s := server.New(":8080", *staticDir)
	fmt.Println("server starting 8080")
	s.Start()
}
