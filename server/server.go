package server

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func serveWS(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ERR upgrading websocket:", err)
		return
	}
	defer func() {
		log.Println("closing ws")
		c.Close()
	}()
	//running at 30 FPS
	frameNS := time.Duration(int(1e9) / 30)
	clk := time.NewTicker(frameNS)
	//main loop
	for {
		select {
		case <-clk.C:
			err = c.WriteMessage(1, []byte("frame"))
			if err != nil {
				log.Println("ERR writing websocket:", err)
				return
			}
		}
	}
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("ERR reading websocket:", err)
			return
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("ERR writing websocket:", err)
			return
		}
	}
}
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

type Server struct {
	port      string
	staticDir string
}

func New(port, staticDir string) *Server {
	return &Server{
		port:      port,
		staticDir: staticDir,
	}
}

func (s *Server) Start() {
	// http.Handle("/asdf/", http.StripPrefix("/asdf/", http.FileServer(http.Dir(staticdir))))
	http.Handle("/", http.FileServer(http.Dir(s.staticDir)))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.FormValue("game"))
		serveWS(w, r)
	})
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
