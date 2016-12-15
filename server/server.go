package server

import (
	"log"
	"net/http"
	// "time"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
	"database/sql"
		
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
	db	  *sql.DB
	gm *gamemanager.GameManager
}

func New(port, staticDir string, db *sql.DB) *Server {
	return &Server{
		port:      port,
		staticDir: staticDir,
		db:	   db,
		gm:        gamemanager.New(),
	}
}

func (s *Server) handleLogin(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.ServeFile(res, req, "./static/login.html")
		return
	}
	username := req.FormValue("username")
	password := req.FormValue("password")
	var databaseUsername  string
	var databasePassword  string

	err := s.db.QueryRow("SELECT username, password FROM users WHERE username=?", username).Scan(&databaseUsername, &databasePassword)
	if err != nil {
		http.Redirect(res, req, "/login", 301)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(databasePassword), []byte(password))
	if err != nil {
		http.Redirect(res, req, "/login", 301)
		return		   
	}
	res.Write([]byte("Hello " + databaseUsername))   
}

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

	switch {
		// Username is available
	case err == sql.ErrNoRows:
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(res, "Server error, unable to create your account.", 500)    
			return
		} 

		_, err = s.db.Exec("INSERT INTO users(username, password) VALUES(?, ?)", username, hashedPassword)
		if err != nil {
			http.Error(res, "Server error, unable to create your account.", 500)    
			return
		}

		res.Write([]byte("User created!"))
		return
	case err != nil: 
		http.Error(res, "Server error, unable to create your account.", 500)    
		return
	default: 
		http.Redirect(res, req, "/", 301)
	}
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	gameName := r.FormValue("game")
	if len(gameName) <= 0 {
		w.Write([]byte("ERR no gameName provided"))
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ERR upgrading websocket:", err)
		return
	}
	defer func() {
		conn.Close()
	}()

	id, gameInput, gameOutput, err := s.gm.Connect(gameName)
	if err != nil {
		log.Println("ERROR: could not add player", err)
		return
	}
	defer s.gm.Disconnect(gameName, id)

	// handle output
	go func() {
		for {
			select {
			case msg, more := <-gameOutput:
				if more {
					err = conn.WriteMessage(TextMessage, msg)
					if err != nil {
						log.Printf("%s %d: %s", gameName, id, "output socket closed")
						return
					}
				} else {
					log.Printf("%s %d: %s", gameName, id, "output channel closed")
					return
				}
				// TODO add efficient timeout
			}
		}
	}()

	// handle input
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil || mt == CloseMessage {
			log.Println("ERR reading websocket:", err)
			return
		}
		gameInput <- message
	}
}

func (s *Server) Start() {
	// http.Handle("/asdf/", http.StripPrefix("/asdf/", http.FileServer(http.Dir(staticdir))))
	http.Handle("/", http.FileServer(http.Dir(s.staticDir)))
	http.HandleFunc("/login", s.handleLogin)
	http.HandleFunc("/signup", s.handleSignup)
	http.HandleFunc("/ws", s.handleWS)

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
