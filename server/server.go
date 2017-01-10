// TODO: split into mini handlers
package server

import (
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
	// "os" // when calling ExecuteTemplate you can use os.Stdout instead to output to screen
	"time"

	"github.com/austindoeswork/S2017-UPE-AI/dbinterface"
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
	db        *dbinterface.DB
	gm        *gamemanager.GameManager
	templates *template.Template
}

// This data is passed into templates so that we can have dynamic information
type Page struct {
	Title    string
	Username string
	Data     string
}

// TODO does it break the server if there is no login cookie? this should be tested
// Loads username if possible from cookie, and loads the template
func (s *Server) ExecuteUserTemplate(res http.ResponseWriter, req *http.Request, template string, data Page) {
	if cookie, err := req.Cookie("login"); err == nil {
		if username, err := s.db.VerifyCookie(cookie); err == nil {
			data.Username = username
		}
	}
	err := s.templates.ExecuteTemplate(res, template, data)
	if err != nil {
		log.Fatal("Cannot Get View ", err)
	}
}

func New(port, staticDir string, db *dbinterface.DB) *Server {
	return &Server{
		port:      port,
		staticDir: staticDir,
		db:        db,
		gm:        gamemanager.New(),
	}
}

// called by /login
func (s *Server) handleLogin(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		s.ExecuteUserTemplate(res, req, "login", Page{Title: "Login"})
		return
	}
	username := req.FormValue("username")
	password := req.FormValue("password")

	// dbinterface processes login request and returns cookie if valid request
	cookie, err := s.db.VerifyLogin(username, password)
	if err != nil { // login failed, send them back to the page
		http.Redirect(res, req, "/login", 301)
		return
	}
	http.SetCookie(res, cookie)
	http.Redirect(res, req, "/profile", 302)
}

// called by /logout
func (s *Server) handleLogout(res http.ResponseWriter, req *http.Request) {
	cookie := &http.Cookie{ // seems a little trivial to put in dbinterface, but can be moved
		Name:    "login",
		Value:   "",
		Path:    "/",
		Expires: time.Now(),
	}
	http.SetCookie(res, cookie)
	http.Redirect(res, req, "/", 302)
}

// TODO should this be replaced with a try catch block?
// called by /profile
func (s *Server) handleProfile(res http.ResponseWriter, req *http.Request) {
	if cookie, err := req.Cookie("login"); err == nil {
		if username, err := s.db.VerifyCookie(cookie); err == nil {
			if profile, err := s.db.GetProfile(username); err == nil {
				s.ExecuteUserTemplate(res, req, "profile", Page{Title: "Profile", Username: username, Data: profile.Apikey})
				return
			}
		}
	}
	http.Redirect(res, req, "/signup", 302) // TODO add an "error, incorrect logged in page"
}

func (s *Server) handleSignup(res http.ResponseWriter, req *http.Request) {
	// Serve signup.html to get requests to /signup
	if req.Method != "POST" {
		s.ExecuteUserTemplate(res, req, "signup", Page{Title: "Signup"})
		return
	}

	username := req.FormValue("username")
	password := req.FormValue("password")

	cookie, err := s.db.SignupUser(username, password)
	if err != nil { // TODO make errors more verbose
		http.Error(res, "Server error, unable to create your account.", 500)
		return
	} else {
		http.SetCookie(res, cookie)
		http.Redirect(res, req, "/profile", 302)
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

func (s *Server) handleGame(res http.ResponseWriter, req *http.Request) {
	s.ExecuteUserTemplate(res, req, "game", Page{Title: "Game"})
}

func (s *Server) handleHome(res http.ResponseWriter, req *http.Request) {
	s.ExecuteUserTemplate(res, req, "home", Page{Title: "Home"})
}

func (s *Server) Start() {
	s.templates = template.Must(template.ParseGlob("./templates/*.html")) // dynamically load all templates with .html ending

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(s.staticDir))))
	http.HandleFunc("/game", s.handleGame)
	http.HandleFunc("/signout", s.handleLogout) // ?? for some reason on my machine if this is logout it doesn't detect it...
	http.HandleFunc("/login", s.handleLogin)
	http.HandleFunc("/signup", s.handleSignup)
	http.HandleFunc("/profile", s.handleProfile)
	http.HandleFunc("/wsjoin", s.handleJoinWS)
	http.HandleFunc("/wswatch", s.handleWatchWS)
	http.HandleFunc("/", s.handleHome)

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
