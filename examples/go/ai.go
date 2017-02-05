package main

import (
	"fmt"
	"os"

	"github.com/gorilla/websocket"
)

func main() {
	serverURL := "ws://localhost:9090/wsplay"
	devkey := "ShanikasEasilyGoofyRamen"

	var dialer *websocket.Dialer

	conn, _, err := dialer.Dial(serverURL, nil)
	checkErr(err)

	err = conn.WriteMessage(1, []byte(devkey))
	checkErr(err)

	_, msg, err := conn.ReadMessage()
	checkErr(err)

	fmt.Printf("welcome %s\n", msg)
	// _ = conn
	// _ = devkey

}

func checkErr(err error) {
	if err != nil {
		fmt.Println("FATAL:", err)
		os.Exit(1)
	}
}
