package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand" // testing seeding once on startup
	"os/user"
	"path"
	"time" // used for seeding

	_ "net/http/pprof"

	_ "github.com/go-sql-driver/mysql"

	"github.com/austindoeswork/S2017-UPE-AI/config"
	"github.com/austindoeswork/S2017-UPE-AI/dbinterface"
	"github.com/austindoeswork/S2017-UPE-AI/server"
)

var (
	db          *dbinterface.DB
	staticDir   = flag.String("static", "./static/", "directory of static files")
	versionFlag = flag.Bool("v", false, "git commit hash")

	//places to look for config
	configPaths = []string{
		path.Join(homeDir(), "npc.conf"),
		"./npc.conf",
		"/etc/npc.conf",
	}

	//holds git commit hash
	commithash string
)

func main() {
	flag.Parse()
	if *versionFlag {
		fmt.Println(commithash)
		return
	}

	c, err := getConfig()
	if err != nil {
		fmt.Println("config error: " + err.Error())
		fmt.Println("Example: \n" + config.Example())
		fmt.Println("Must Exist At One Of: ")
		for _, p := range configPaths {
			fmt.Println("  ", p)
		}
		return
	}

	rand.Seed(time.Now().UTC().UnixNano()) // seed on startup to current time

	db := dbinterface.NewDB(c.DatabaseAddress(), c.DatabaseName(), c.DatabaseUsername(), c.DatabasePassword())
	defer db.Close()

	// when building with the bash script, this will save the commit hash for binary debugging
	fmt.Println("version: " + commithash)

	s := server.New(c.ServerAddress(), *staticDir, db)
	log.Println("BLASTOFF:", c.ServerAddress())
	s.Start()
}

// getConfig returns a config file if found, otherwise an error
func getConfig() (*config.Config, error) {
	for _, configPath := range configPaths {
		b, err := ioutil.ReadFile(configPath)
		if err != nil {
			continue
		}
		var c config.Config
		if err := json.Unmarshal(b, &c); err != nil {
			continue
		}
		if err := c.Validate(); err != nil {
			continue
		}
		cb, _ := json.MarshalIndent(&c, "", "  ")
		fmt.Printf("config: %s\n%s\n", configPath, string(cb))
		return &c, nil
	}
	return nil, errors.New("failed to read config")
}

// homeDir gets the user's home directory
func homeDir() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return u.HomeDir
}
