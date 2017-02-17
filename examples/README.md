# How to make your AI

```
+-------+
|NP     | <-- websocket --> your bot
|Compete| <-- websocket --> your enemy
|Server | --- websocket --> loyal fan
+-------+
```

list of language examples:

- [golang](go/ai.go)


## 0 Overview
- Get your devkey
- Open up a websocket to either of: 
	- `ws://localhost/wsjoin?game=NameOfGame`
	- a specific room (that someone else can join)
	- `ws://localhost/wsplay`
	- play next available person
- Send devkey through the websocket
- Begin receiving game states
- Send game commands
- Win hella ca$h

## 1 Credentials

- Go to <a href="npcompete.io">npcompete.io</a>
- Signup with a username and password 	
- Get your devkey under your <a href="npcompete.io/profile">profile</a> 
- Do not share this key, as it will identify your bot
- nice

## 2 Websocket - Tutorial

This tutorial is in golang, but you can find other language examples within this directory

- We will be using the gorilla websocket package

```
$> go get github.com/gorilla/websocket
```
- Now lets make a file, and call it `ai.go`

```
package main

import (
	"fmt" // for io
	"github.com/gorilla/websocket" // for websockets
)

func main() {
	// some variables we need to connect
	serverURL := "ws://npcompete.io/wsplay"
	devkey := "YOURKEYHERE"
	
	//...
}

```

- Quickly open up a websocket

```
	//open websocket
	var dialer *websocket.Dialer
	conn, _, _ := dialer.Dial(serverURL, nil)
	defer conn.Close()
	
	//...
```

- Then, let's send our key thru the socket

```
	//write our devkey
	conn.WriteMessage(1, []byte(devkey))
	
	//...
```

- The server will respond with which player you are, your username, and the game name

```
	//receive acknowledgement
	_, msg, _ := conn.ReadMessage()
	fmt.Printf("%s\n", msg)
	
	//...
```

- Once the game starts, the server will being sending game states as json objects
- Now let's respond to each game frame with our super smart strategy

```
	//send game inputs
	for {
		_, msg, _ = conn.ReadMessage()
		_ = msg
		// fmt.Printf("%s\n", msg) 
		// uncomment this to output all frames
		
		conn.WriteMessage(1, []byte("b00 02"))
		//                           ^   ^  
		//                           |   |
		//                           |   lane #
		//                code for a "nut" troop
	}
```
- TODO: watch your game at <a href=TODO>TODO</a>
- We're done! [Full Code HERE](go/ai.go)
