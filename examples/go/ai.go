package main

import (
	"fmt"
	"os"

	"github.com/gorilla/websocket"
)

func main() {
	serverURL := "ws://localhost:8080/wsplay"
	devkey := "LaurettasEagerlyBoorishSponge"

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

	for {
		_, msg, err = conn.ReadMessage()
		checkErr(err)
		fmt.Printf("%s\n", msg)
		conn.WriteMessage(1, []byte("b00 01"))
	}

}

func checkErr(err error) {
	if err != nil {
		fmt.Println("FATAL:", err)
		os.Exit(1)
	}
}
