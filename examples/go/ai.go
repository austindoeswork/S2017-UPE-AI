package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gorilla/websocket"
)

func main() {
	serverURL := "ws://localhost:8080/wsplay"
	devkey := "YOURDEVKEY"

	//open websocket
	var dialer *websocket.Dialer
	conn, _, err := dialer.Dial(serverURL, nil)
	checkErr(err)
	defer conn.Close()

	//write our devkey
	err = conn.WriteMessage(1, []byte(devkey))
	checkErr(err)

	//receive acknowledgement
	_, msg, err := conn.ReadMessage()
	checkErr(err)
	fmt.Printf("%s\n", msg)

	//send game inputs
	frame := &Frame{}
	for {
		_, msg, err = conn.ReadMessage()
		_ = msg
		checkErr(err)
		json.Unmarshal(msg, frame)
		if frame.P1.MainCore.Hp < 0 {
			fmt.Println("PLAYER 1 hp critical:", frame.P1.MainCore.Hp)
			return
		}
		if frame.P2.MainCore.Hp < 0 {
			fmt.Println("PLAYER 2 hp critical:", frame.P2.MainCore.Hp)
			return
		}

		// fmt.Printf("%s\n", msg) // uncomment this to output all frames
		conn.WriteMessage(1, []byte("b01 01"))
	}
}

func checkErr(err error) {
	if err != nil {
		fmt.Println("FATAL:", err)
		os.Exit(1)
	}
}

type GameInfo struct {
	Player   int    `json:"Player"`
	UserName string `json:"UserName"`
	GameName string `json:"GameName"`
}

type Frame struct {
	W  int `json:"w"`
	H  int `json:"h"`
	P1 struct {
		Owner  int      `json:"owner"`
		Income int      `json:"income"`
		Bits   int      `json:"bits"`
		Towers []string `json:"towers"`
		Troops []struct {
			Owner int `json:"owner"`
			X     int `json:"x"`
			Y     int `json:"y"`
			Maxhp int `json:"maxhp"`
			Hp    int `json:"hp"`
			Enum  int `json:"enum"`
		} `json:"troops"`
		MainCore struct {
			Owner int `json:"owner"`
			X     int `json:"x"`
			Y     int `json:"y"`
			Maxhp int `json:"maxhp"`
			Hp    int `json:"hp"`
			Enum  int `json:"enum"`
		} `json:"mainCore"`
	} `json:"p1"`
	P2 struct {
		Owner  int      `json:"owner"`
		Income int      `json:"income"`
		Bits   int      `json:"bits"`
		Towers []string `json:"towers"`
		Troops []struct {
			Owner int `json:"owner"`
			X     int `json:"x"`
			Y     int `json:"y"`
			Maxhp int `json:"maxhp"`
			Hp    int `json:"hp"`
			Enum  int `json:"enum"`
		} `json:"troops"`
		MainCore struct {
			Owner int `json:"owner"`
			X     int `json:"x"`
			Y     int `json:"y"`
			Maxhp int `json:"maxhp"`
			Hp    int `json:"hp"`
			Enum  int `json:"enum"`
		} `json:"mainCore"`
	} `json:"p2"`
}
