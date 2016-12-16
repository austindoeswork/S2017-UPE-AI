package server

import (
	"log"
	"net/http"
	// "time"
	"github.com/gorilla/websocket"

	"github.com/austindoeswork/S2017-UPE-AI/gamemanager"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Server handles websockets and creation of games
// TODO create a router/handler
// TODO game manager?? think about this
type Server struct {
	port      string
	staticDir string

	gm *gamemanager.GameManager
}

func New(port, staticDir string) *Server {
	return &Server{
		port:      port,
		staticDir: staticDir,
		gm:        gamemanager.New(),
	}
}

func (s *Server) handleWatchWS(w http.ResponseWriter, r *http.Request) {
	gameName := r.FormValue("game")
	if len(gameName) <= 0 {
		w.Write([]byte("ERR: no gameName provided"))
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ERR: upgrading websocket", err)
		return
	}
	defer conn.Close()

	id, gameOutput, err := s.gm.Watch(gameName)
	if err != nil {
		log.Println("ERR: could not add watcher", err)
		return
	}
	defer s.gm.Disconnect(gameName, id)

	// handle output
	chanToWS(gameOutput, conn)
}
func (s *Server) handleJoinWS(w http.ResponseWriter, r *http.Request) {
	gameName := r.FormValue("game")
	if len(gameName) <= 0 {
		w.Write([]byte("ERR: no gameName provided"))
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ERR: upgrading websocket", err)
		return
	}
	defer conn.Close()

	id, gameInput, gameOutput, err := s.gm.Connect(gameName)
	if err != nil {
		log.Println("ERR: could not add player", err)
		return
	}
	defer s.gm.Disconnect(gameName, id)

	// handle output
	go chanToWS(gameOutput, conn)

	// handle input
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil || mt == CloseMessage {
			log.Println("ERR: reading websocket", err)
			return
		}
		gameInput <- message
	}
}

func chanToWS(gameOutput <-chan []byte, conn *websocket.Conn) {
	for {
		select {
		case msg, more := <-gameOutput:
			if more {
				err := conn.WriteMessage(TextMessage, msg)
				if err != nil {
					log.Println("output socket closed")
					return
				}
			} else { // chan has been closed
				log.Println("output channel closed")
				conn.Close()
				return
			}
			// TODO add efficient timeout
		}
	}
}

func (s *Server) Start() {
	// http.Handle("/asdf/", http.StripPrefix("/asdf/", http.FileServer(http.Dir(staticdir))))

	http.Handle("/", http.FileServer(http.Dir(s.staticDir)))
	http.HandleFunc("/wsjoin", s.handleJoinWS)
	http.HandleFunc("/wswatch", s.handleWatchWS)

	// echo ws for testing
	http.HandleFunc("/wstest", func(w http.ResponseWriter, r *http.Request) {
		echo(w, r)
	})

	http.ListenAndServe(s.port, nil)
}

// from https://godoc.org/github.com/gorilla/websocket
const (
	// TextMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	TextMessage = 1

	// BinaryMessage denotes a binary data message.
	BinaryMessage = 2

	// CloseMessage denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the FormatCloseMessage
	// function to format a close message payload.
	CloseMessage = 8

	// PingMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PingMessage = 9

	// PongMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PongMessage = 10
)

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}
