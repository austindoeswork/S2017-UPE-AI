package server

import (
	"database/sql"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"

	"time"

	"github.com/austindoeswork/S2017-UPE-AI/gamemanager"
)

// Upgrades a regular ResponseWriter to WebSocketResponseWriter
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
	db        *sql.DB
	gm        *gamemanager.GameManager
}

func New(port, staticDir string, db *sql.DB) *Server {
	return &Server{
		port:      port,
		staticDir: staticDir,
		db:        db,
		gm:        gamemanager.New(),
	}
}

// TODO: move to database interface file
func (s *Server) handleLogin(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.ServeFile(res, req, "./static/login.html")
		return
	}
	username := req.FormValue("username")
	password := req.FormValue("password")
	var databaseUsername string
	var databasePassword string
	var apikey string

	err := s.db.QueryRow("SELECT username, password, apikey FROM users WHERE username=?", username).Scan(&databaseUsername, &databasePassword, &apikey)
	if err != nil {
		http.Redirect(res, req, "/login", 301)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(databasePassword), []byte(password))
	if err != nil {
		http.Redirect(res, req, "/login", 301)
		return
	}
	res.Write([]byte("Hello " + databaseUsername + ", your apikey is " + apikey))
}

// TODO: move to database interface file
func (s *Server) handleSignup(res http.ResponseWriter, req *http.Request) {
	// Serve signup.html to get requests to /signup
	if req.Method != "POST" {
		http.ServeFile(res, req, "./static/signup.html")
		return
	}

	username := req.FormValue("username")
	password := req.FormValue("password")

	var user string

	err := s.db.QueryRow("SELECT username FROM users WHERE username=?", username).Scan(&user)

	switch { // Username is available
	case err == sql.ErrNoRows:
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		apikey := string(GenerateUniqueKey(true, true, true, true))
		if err != nil {
			http.Error(res, "Server error, unable to create your account.", 500)
			return
		}

		_, err = s.db.Exec("INSERT INTO users(username, password, apikey) VALUES(?, ?, ?)", username, hashedPassword, apikey)
		if err != nil {
			http.Error(res, "Server error, unable to create your account.", 500)
			return
		}

		res.Write([]byte("User created! Your apikey is " + apikey))
		return
	case err != nil:
		http.Error(res, "Server error, unable to create your account.", 500)
		return
	default:
		http.Redirect(res, req, "/", 301)
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

	quit := make(chan bool)
	gameOutput, err := s.gm.WatchGame(gameName, quit)
	if err != nil {
		log.Println("ERR: could not add watcher", err)
		return
	}
	defer func() {
		select {
		case quit <- true:
		default:
		}
	}()

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

	// TODO sanitization on gameName
	if !s.gm.HasGame(gameName) {
		err = s.gm.NewGame(gameName)
		if err != nil {
			log.Println("ERR: creating game", err)
		}
	}

	quitIn := make(chan bool)
	gameInput, err := s.gm.ControlGame(gameName, quitIn)
	if err != nil {
		log.Println("ERR: could not add controller", err)
		return
	}
	defer func() {
		select {
		case quitIn <- true:
		default:
		}
	}()

	quitOut := make(chan bool)
	gameOutput, err := s.gm.WatchGame(gameName, quitOut)
	if err != nil {
		log.Println("ERR: could not add watcher", err)
		return
	}
	defer func() {
		select {
		case quitOut <- true:
		default:
		}
	}()

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
	defer conn.Close()
	for {
		t := time.NewTimer(time.Second * 10)
		select {
		case msg, more := <-gameOutput:
			if !t.Stop() {
				<-t.C
			}
			if more {
				t.Reset(time.Second * 10)

				err := conn.WriteMessage(TextMessage, msg)
				if err != nil {
					log.Println("output socket closed")
					return
				}
			} else { // chan has been closed
				log.Println("output channel closed")
				return
			}
			// TODO add efficient timeout
		case <-t.C:
			log.Println("chanToWS timeout")
			return
		}
	}
}

func (s *Server) Start() {
	// http.Handle("/asdf/", http.StripPrefix("/asdf/", http.FileServer(http.Dir(staticdir))))
	http.Handle("/", http.FileServer(http.Dir(s.staticDir)))
	http.HandleFunc("/login", s.handleLogin)
	http.HandleFunc("/signup", s.handleSignup)
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
