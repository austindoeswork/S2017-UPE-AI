package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"

	"github.com/gorilla/websocket"
)

var Player int

func NewTroopInput(enum, lane int) string { //0 top 1 mid 2 bot
	if enum < 10 {
		return fmt.Sprintf("b0%d 0%d", enum, lane)
	} else {
		return fmt.Sprintf("b%d 0%d", enum, lane)
	}
}

func plotToSpot(plot int) (int, int) { //returns 0-6 for lane, 0-10 for num
	lane := int(math.Ceil(float64((plot+1)/11))) - 1
	num := plot % 11
	return lane, num
}
func spotToPlot(lane, num int) string {
	p := lane*11 + num
	// fmt.Println("p = ", lane, "*11", "+", num)
	if p < 10 {
		return fmt.Sprintf("0%d", p)
	} else {
		return fmt.Sprintf("%d", p)
	}
}

func NewTowerInput(enum, lane int, num int) string { //0,1 top 2,3 mid 4,5 bot
	plot := spotToPlot(lane, num)
	if enum < 10 {
		return fmt.Sprintf("b0%d %s", enum, plot)
	} else {
		return fmt.Sprintf("b%d %s", enum, plot)
	}
}

func main() {
	serverURL := "ws://npcompete.io/wsplay"
	devkey := "YOURKEYHERE"

	//open websocket
	var dialer *websocket.Dialer
	conn, _, err := dialer.Dial(serverURL, nil)
	checkErr(err)
	defer conn.Close()

	//write our devkey
	err = conn.WriteMessage(1, []byte(devkey))
	checkErr(err)

	//receive acknowledgement
	gi := &GameInfo{}
	_, msg, err := conn.ReadMessage()
	checkErr(err)
	fmt.Printf("%s\n", msg)
	json.Unmarshal(msg, gi)
	Player = gi.Player

	var buildDirection int
	if Player == 1 {
		buildDirection = 1
	} else {
		buildDirection = -1
	}
	//send game inputs
	frame := &Frame{}
	for {
		_, msg, err = conn.ReadMessage()
		_ = msg
		checkErr(err)
		json.Unmarshal(msg, frame)
		if frame.P1.MainCore.Hp < 0 {
			fmt.Println("Player 2 Wins!")
			return
		}
		if frame.P2.MainCore.Hp < 0 {
			fmt.Println("Player 1 Wins!")
			return
		}

		var input string
		if Player == 1 {
			if frame.P1.Income < 550 {
				input = NewTowerInput(53, 3, 5-(5*buildDirection))
			} else {
				input = NewTroopInput(0, 1)
			}
		} else {
			if frame.P2.Income < 550 {
				input = NewTowerInput(53, 3, 5-(5*buildDirection))
			} else {
				input = NewTroopInput(0, 1)
			}
		}
		fmt.Println(input)
		conn.WriteMessage(1, []byte(input))
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
