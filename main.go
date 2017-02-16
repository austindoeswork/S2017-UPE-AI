package main

import (
	"flag"
	"fmt"
	_ "net/http/pprof"

	_ "github.com/go-sql-driver/mysql"

	"math/rand" // testing seeding once on startup
	"time"      // used for seeding

	"github.com/austindoeswork/S2017-UPE-AI/dbinterface"
	"github.com/austindoeswork/S2017-UPE-AI/server"
)

var (
	// TODO use a config file
	db          *dbinterface.DB
	staticDir   = flag.String("static", "./static/", "directory of static files")
	versionFlag = flag.Bool("v", false, "git commit hash")

	commithash string
)

func main() {
	// TODO move this seed to init?
	rand.Seed(time.Now().UTC().UnixNano()) // seed on startup to current time

	db := dbinterface.NewDB()
	defer db.Close()

	// when running with the bash script, this will save the commit hash for binary debugging
	flag.Parse()
	if *versionFlag {
		fmt.Println(commithash)
		return
	}
	fmt.Println("version: " + commithash)
	// end bash script help

	s := server.New(":80", *staticDir, db)
	fmt.Println("server starting 80")
	s.Start()
}
