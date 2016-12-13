package server

import (
	"log"
	"net/http"
	// "time"

	"github.com/gorilla/websocket"

	"github.com/austindoeswork/S2017-UPE-AI/game"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ============================================================================
// GAME
// ============================================================================

// Game is a wrapper for the game struct
type Game struct {
	*game.Pong // TODO convert to interface
	listeners  map[*websocket.Conn]bool
}

func NewGame() *Game {
	return &Game{
		game.New(30, 20, 60),
		make(map[*websocket.Conn]bool),
	}
}

func (g *Game) Start() error {
	outputChan, err := g.Pong.Start()
	if err != nil {
		return err
	}
	go func() {
		for {
			if len(g.listeners) == 0 {
				g.Quit()
				return
			}
			select {
			case out := <-outputChan:
				g.sendListeners(out)
			}
		}
	}()
	return nil
}

func (g *Game) AddListener(listener *websocket.Conn) {
	g.listeners[listener] = true
}

func (g *Game) sendListeners(msg []byte) {
	for conn, _ := range g.listeners {
		err := conn.WriteMessage(TextMessage, msg)
		if err != nil {
			delete(g.listeners, conn)
		}
	}

}

// ============================================================================
// SERVER
// ============================================================================

// Server handles websockets and creation of games
// TODO create a router/handler
// TODO game manager?? think about this
type Server struct {
	port      string
	staticDir string

	games map[string]*Game
}

func New(port, staticDir string) *Server {
	return &Server{
		port:      port,
		staticDir: staticDir,
		games:     make(map[string]*Game),
	}
}

func (s *Server) handlePlayerWS(w http.ResponseWriter, r *http.Request, gameInput chan []byte, gameName string) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ERR upgrading websocket:", err)
		return
	}
	defer func() {
		// if the game was never started, remove it
		// if val, ok := s.games[gameName]; ok {
		// if !val.IsStarted() {
		// delete(s.games, gameName)
		// }
		// }
		log.Println("closing ws")
		c.Close()
	}()

	s.games[gameName].AddListener(c)

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("ERR reading websocket:", err)
			return
		}
		log.Printf("recv: %s", message)
		gameInput <- message

		// err = c.WriteMessage(mt, message)
		// if err != nil {
		// log.Println("ERR writing websocket:", err)
		// return
		// }
	}
}

func (s *Server) Start() {
	// http.Handle("/asdf/", http.StripPrefix("/asdf/", http.FileServer(http.Dir(staticdir))))

	http.Handle("/", http.FileServer(http.Dir(s.staticDir)))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		gameName := r.FormValue("game")
		if len(gameName) <= 0 {
			w.Write([]byte("ERROR: no game name provided"))
			return
		}

		if _, exists := s.games[gameName]; !exists {
			log.Println("created new game")
			s.games[gameName] = NewGame()
		}

		playerID, inputChan, err := s.games[gameName].AddPlayer()

		// TODO decide a better spot to do this
		if playerID == 2 {
			s.games[gameName].Start()
		}

		if err != nil {
			w.Write([]byte("ERROR: cannot add player"))
			log.Println(err)
			return
		}

		log.Println("opening ws", gameName)
		s.handlePlayerWS(w, r, inputChan, gameName)
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
