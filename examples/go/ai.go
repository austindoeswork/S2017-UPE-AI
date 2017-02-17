package main

import (
	"fmt"
	"os"

	"github.com/gorilla/websocket"
)

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
	_, msg, err := conn.ReadMessage()
	checkErr(err)
	fmt.Printf("%s\n", msg)

	//send game inputs
	for {
		_, msg, err = conn.ReadMessage()
		_ = msg
		checkErr(err)
		// fmt.Printf("%s\n", msg) // uncomment this to output all frames
		conn.WriteMessage(1, []byte("b00 02"))
	}

}

func checkErr(err error) {
	if err != nil {
		fmt.Println("FATAL:", err)
		os.Exit(1)
	}
}
