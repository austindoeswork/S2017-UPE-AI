package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "net/http/pprof"

	"math/rand" // testing seeding once on startup
	"time"      // used for seeding

	"github.com/austindoeswork/S2017-UPE-AI/server"
)

var (
	// TODO use a config file
	db          *sql.DB
	err         error // TODO taken from tutorial, is global err object really necessary?
	staticDir   = flag.String("static", "./static/", "directory of static files")
	versionFlag = flag.Bool("v", false, "git commit hash")

	commithash string
)

/*
	DB Note: DB is initialized here, but details are in database interface (separate object) and some login/signup stuff is handled
	by server/server.go
*/

func main() {
	// TODO move this seed to init
	rand.Seed(time.Now().UTC().UnixNano()) // seed on startup to current time

	db, err = sql.Open("mysql", "root@/aicomp") // assumes there is a local MySQL database with user root and no password
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	// when running with the bash script, this will save the commit hash for binary debugging
	flag.Parse()
	if *versionFlag {
		fmt.Println(commithash)
		return
	}
	fmt.Println("version: " + commithash)
	// end bash script help

	s := server.New(":8080", *staticDir, db)
	fmt.Println("server starting 8080")
	s.Start()
}
