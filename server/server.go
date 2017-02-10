// TODO: split into mini handlers
package server

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	// "os" // when calling ExecuteTemplate you can use os.Stdout instead to output to screen
	"time"

	"fmt"

	"os"

	"regexp"

	"github.com/austindoeswork/S2017-UPE-AI/dbinterface"
	"github.com/austindoeswork/S2017-UPE-AI/gamemanager"
	"github.com/gorilla/sessions"
)

// Upgrades a regular ResponseWriter to WebSocketResponseWriter
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// isAlpha checks if the given string contains only alphabetic characters
var isAlpha = regexp.MustCompile(`^[A-Za-z\s]+$`).MatchString

// isAlphaNumeric checks if the given string contains only alphanumeric characters
var isAlphaNumeric = regexp.MustCompile(`^[A-Za-z\d]+$`).MatchString

// Server handles websockets and creation of games
// TODO create a router/handler
// TODO game manager?? think about this
type Server struct {
	port      string
	staticDir string
	db        *dbinterface.DB
	gm        *gamemanager.GameManager
	store     *sessions.CookieStore
	templates *template.Template
	mailer    *Mailer
}

// This data is passed into templates so that we can have dynamic information
type Page struct {
	Title    string
	Flash    []string
	Username string
	Data     interface{}
}

// TODO does it break the server if there is no login cookie? this should be tested
// Loads username if possible from cookie, and loads the template
func (s *Server) ExecuteUserTemplate(res http.ResponseWriter, req *http.Request, template string, data Page) {
	if cookie, err := req.Cookie("login"); err == nil {
		if username, err := s.db.VerifyCookie(cookie); err == nil {
			data.Username = username
		}
	}
	session, err := s.store.Get(req, "flash")
	if err != nil {
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	rawFlashes := session.Flashes()
	flashes := make([]string, len(rawFlashes))
	for i, d := range rawFlashes {
		flashes[i] = d.(string)
	}
	session.Save(req, res)
	data.Flash = flashes
	err = s.templates.ExecuteTemplate(res, template, data)
	if err != nil {
		log.Fatal("Cannot Get View ", err)
	}
}

func New(port, staticDir string, db *dbinterface.DB) *Server {
	m, err := NewMailer("team@aicomp.io")
	if err != nil {
		log.Println(err)
	}
	os.Mkdir("./identicons", 0777)
	return &Server{
		port:      port,
		staticDir: staticDir,
		db:        db,
		gm:        gamemanager.New(),
		store:     sessions.NewCookieStore([]byte("secret")),
		mailer:    m,
	}
}

// called by /login
func (s *Server) handleLogin(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		s.ExecuteUserTemplate(res, req, "login", Page{Title: "Login"})
		return
	}
	session, err := s.store.Get(req, "flash")
	if err != nil {
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	username := req.FormValue("username")
	password := req.FormValue("password")
	if username == "" || password == "" {
		session.AddFlash("Username and password required.")
		session.Save(req, res)
		s.ExecuteUserTemplate(res, req, "login", Page{Title: "Login"})
		return
	}
	// dbinterface processes login request and returns cookie if valid request
	cookie, err := s.db.VerifyLogin(username, password)
	if err != nil { // login failed, send them back to the page
		session.AddFlash("Invalid username or password")
		session.Save(req, res)
		http.Redirect(res, req, "/login", http.StatusMovedPermanently)
		return
	}
	http.SetCookie(res, cookie)
	http.Redirect(res, req, "/profile", http.StatusFound)
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
	session, err := s.store.Get(req, "flash")
	if err != nil {
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	session.AddFlash("You have been logged out.")
	session.Save(req, res)
	http.Redirect(res, req, "/", http.StatusFound)
}

// TODO should this be replaced with a try catch block?
// called by /profile
func (s *Server) handleProfile(res http.ResponseWriter, req *http.Request) {
	if cookie, err := req.Cookie("login"); err == nil {
		if username, err := s.db.VerifyCookie(cookie); err == nil {
			if profile, err := s.db.GetUser(username); err == nil {
				profile.ProfilePicture, _ = LoadIdenticon(profile.ProfilePicture)
				s.ExecuteUserTemplate(res, req, "profile", Page{Title: "Profile", Username: username,
					Data: profile})
				return
			}
		}
	}
	session, err := s.store.Get(req, "flash")
	if err != nil {
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	session.AddFlash("You must login first.")
	session.Save(req, res)
	http.Redirect(res, req, "/login", http.StatusFound)
}

// handleSignup serves the signup.html template on not POST requests. On POST requests, we handle
// 	the request to create a new user.
func (s *Server) handleSignup(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		s.ExecuteUserTemplate(res, req, "signup", Page{Title: "Signup"})
		return
	}
	session, err := s.store.Get(req, "flash")
	if err != nil {
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	fullName := req.FormValue("name")
	email := req.FormValue("email")
	username := req.FormValue("username")
	password := req.FormValue("password")
	if fullName == "" || email == "" || username == "" || password == "" {
		session.AddFlash("All fields are required.")
		session.Save(req, res)
		s.ExecuteUserTemplate(res, req, "signup", Page{Title: "Signup"})
		return
	}
	if !isAlpha(fullName) {
		session.AddFlash("Name field can only contain alphabetic characters.")
		session.Save(req, res)
		s.ExecuteUserTemplate(res, req, "signup", Page{Title: "Signup"})
		return
	}
	if !isAlphaNumeric(username) {
		session.AddFlash("Username field can only contain alphanumeric characters.")
		session.Save(req, res)
		s.ExecuteUserTemplate(res, req, "signup", Page{Title: "Signup"})
		return
	}
	profilePicLoc := fmt.Sprintf("./identicons/%s.png", username)
	user := &dbinterface.User{
		Name:           fullName,
		Email:          email,
		ProfilePicture: profilePicLoc,
		Username:       username,
	}
	cookie, err := s.db.SignupUser(user, password)
	if err != nil {
		if err.Error() == "username exists" {
			session.AddFlash(fmt.Sprintf("Username '%s' already exists.", username))
			session.Save(req, res)
			s.ExecuteUserTemplate(res, req, "signup", Page{Title: "Signup"})
			return
		}
		if err.Error() == "email exists" {
			session.AddFlash(fmt.Sprintf("Email '%s' already exists.", email))
			session.Save(req, res)
			s.ExecuteUserTemplate(res, req, "signup", Page{Title: "Signup"})
			return
		}
		http.Error(res, "Server error, unable to create your account.", http.StatusInternalServerError)
		return
	}
	if s.mailer != nil {
		_, _, err = s.mailer.MailSignup(email, fullName)
		if err != nil {
			session.AddFlash(fmt.Sprintf("Failed to send email to '%s'.", email))
			session.Save(req, res)
		}
	}
	hash := GenerateHash(username)
	NewIdenticon(hash, nil).Save(profilePicLoc)
	http.SetCookie(res, cookie)
	http.Redirect(res, req, "/profile", http.StatusFound)
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
	profile, err := s.db.GetUserFromApiKey(string(idmessage))
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

func (s *Server) handlePlayWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ERR: upgrading websocket", err)
		return
	}
	defer conn.Close()

	log.Printf("SERVER: playws\n")

	// AUTHENTICATE USER
	//TODO - timeout on this
	idmt, idmessage, err := conn.ReadMessage()
	if err != nil || idmt == CloseMessage {
		log.Println("ERR: reading websocket", err)
		return
	}
	profile, err := s.db.GetUserFromApiKey(string(idmessage))
	if err != nil {
		log.Println("ERR: getting profile", err)
		return
	}
	userName := profile.Username

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

func (s *Server) Start() {
	s.templates = template.Must(template.ParseGlob("./templates/*.html")) // dynamically load all templates with .html ending

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(s.staticDir))))
	http.HandleFunc("/game", s.handleGame)
	http.HandleFunc("/gamelist", s.handleGameList)
	http.HandleFunc("/signout", s.handleLogout) // ?? for some reason on my machine if this is logout it doesn't detect it...
	http.HandleFunc("/login", s.handleLogin)
	http.HandleFunc("/signup", s.handleSignup)
	http.HandleFunc("/profile", s.handleProfile)
	http.HandleFunc("/docs", s.handleDocs)
	http.HandleFunc("/wsjoin", s.handleJoinWS)
	http.HandleFunc("/wsplay", s.handlePlayWS)
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
