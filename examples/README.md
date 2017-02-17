# How to make your AI
```
+-------+
|NP     | <-- websocket --> your bot
|Compete| <-- websocket --> your enemy
|Server | --- websocket --> loyal fan
+-------+
```

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

## 2 Opening a Websocket

This tutorial is in golang, but you can find other language examples within this directory

- We will be using the gorilla websocket package

```
$> go get github.com/gorilla/websocket
```
- Now lets make a file, and call it `ai.go`

```
package main

import (
)

func main() {
	// some variables we need to connect
	serverURL := "ws://npcompete.io/wsplay"
	devkey := "ShanikasEasilyGoofyRamen"
}

```

- I'm going to write a quick 
