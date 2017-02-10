package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	// "os" // when calling ExecuteTemplate you can use os.Stdout instead to output to screen
)

// Upgrades a regular ResponseWriter to WebSocketResponseWriter
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

//
// HANDLERS
//

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
		Expires: time.Now().Add(time.Hour * 2400),
	}
	http.SetCookie(res, cookie)
	http.Redirect(res, req, "/", 302)
}

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
		http.Error(res, "Server error, unable to create your account:"+err.Error(), 500)
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

	log.Printf("SERVER: joinws %s\n", gameName)

	// AUTHENTICATE USER
	//TODO - timeout on this
	idmt, idmessage, err := conn.ReadMessage()
	if err != nil || idmt == CloseMessage {
		log.Println("ERR: reading websocket", err)
		return
	}
	profile, err := s.db.GetProfileFromApiKey(string(idmessage))
	if err != nil {
		log.Println("ERR: getting profile", err)
		return
	}
	userName := profile.Username

	// MAKE GAME IF DNE
	if !s.gm.HasGame(gameName) {
		err = s.gm.NewGame(gameName)
		if err != nil {
			log.Println("ERR: creating game", err)
		}
	}

	s.playGame(conn, userName, gameName)
}

func (s *Server) handlePlayWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ERR: upgrading websocket", err)
		return
	}
	defer conn.Close()

	// AUTHENTICATE USER
	//TODO - timeout on this
	idmt, idmessage, err := conn.ReadMessage()
	if err != nil || idmt == CloseMessage {
		log.Println("ERR: reading websocket", err)
		return
	}
	profile, err := s.db.GetProfileFromApiKey(string(idmessage))
	if err != nil {
		conn.WriteMessage(1, []byte("ERROR invalid devkey"))
		log.Println("ERR: getting profile", err)
		return
	}
	userName := profile.Username
	conn.WriteMessage(1, []byte(userName))

	// MAKE GAME IF DNE
	gameName, err := s.gm.PopOpenGame()
	if err != nil {
		gameName, err = s.gm.NewOpenGame()
		if err != nil {
			log.Println("ERR: creating game", err)
			return
		}
	}
	if !s.gm.HasGame(gameName) {
		log.Println("ERR: creating game", err)
	}

	log.Printf("SERVER: playws %s\n", gameName)

	s.playGame(conn, userName, gameName)
}

func (s *Server) playGame(conn *websocket.Conn, userName, gameName string) {
	// GET CONTROLLER
	quitIn := make(chan bool)
	gameInput, err := s.gm.ControlGame(gameName, userName, quitIn)
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

	// GET WATCHER
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

func (s *Server) handleGameList(w http.ResponseWriter, r *http.Request) {
	gamelist := (s.gm.ListGames())
	b, err := json.Marshal(&gamelist)
	if err != nil {
		log.Fatal(err)
		return
	}
	w.Write(b)
	return
}

func (s *Server) handleGame(res http.ResponseWriter, req *http.Request) {
	s.ExecuteUserTemplate(res, req, "game", Page{Title: "Game"})
}

func (s *Server) handleHome(res http.ResponseWriter, req *http.Request) {
	s.ExecuteUserTemplate(res, req, "home", Page{Title: "Home"})
}

func (s *Server) handleDocs(res http.ResponseWriter, req *http.Request) {
	s.ExecuteUserTemplate(res, req, "docs", Page{Title: "Documentation"})
}

//
// TEMPLATE FUNCTIONS
//

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