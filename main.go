package main

import (
	"flag"
	"fmt"
	_ "net/http/pprof"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"

	"math/rand" // testing seeding once on startup
	"time" // used for seeding

	"github.com/austindoeswork/S2017-UPE-AI/server"
)

var (
	// TODO use a config file
	staticDir = flag.String("static", "./static/", "directory of static files")
	db *sql.DB
	err error // TODO taken from tutorial, is global err object really necessary?
)

/*
	DB Note: DB is initialized here, but details are in database interface (separate object) and some login/signup stuff is handled 
	by server/server.go
*/

func main() {
	rand.Seed(time.Now().UTC().UnixNano()) // seed on startup to current time

     	db, err = sql.Open("mysql", "root@/aicomp")
	if err != nil {
	        panic(err.Error())    
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
	       panic(err.Error())
	}

	flag.Parse()
	s := server.New(":8080", *staticDir, db)
	fmt.Println("server starting 8080")
	s.Start()
}
