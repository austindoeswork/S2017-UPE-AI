package server

import (
	"log"
	"net/http"

	"time"

	"github.com/gorilla/websocket" // when calling ExecuteTemplate you can use os.Stdout instead to output to screen

	"fmt"

	// below imports used for replaying replays from file (move to another file?)
	"bufio"
	"os"

	"github.com/austindoeswork/S2017-UPE-AI/dbinterface"
)

// Upgrades a regular ResponseWriter to WebSocketResponseWriter
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
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

func (s *Server) handleLeaderboard(res http.ResponseWriter, req *http.Request) {
	data, err := s.db.GetLeaderboard()
	if err != nil {
		s.ExecuteUserTemplate(res, req, "leaderboard", Page{Title: "Leaderboard", Data: nil})
	} else {
		for index := range data {
			data[index].ProfilePicture, _ = LoadIdenticon(data[index].ProfilePicture)
		}
		s.ExecuteUserTemplate(res, req, "leaderboard", Page{Title: "Leaderboard", Data: data})
	}
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
	profilePicLoc := fmt.Sprintf("./identicons/%s.png",
		validFilenameCharacters.ReplaceAllString(username, "_"))
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

// this function does not belong in handler.go, move
func (s *Server) playReplayGame(gameName, fullReplayName, name1, name2 string) {
	quitIn1 := make(chan bool)
	gameCtrl1, err := s.gm.ControlGame(gameName, name1, quitIn1)
	if err != nil {
		log.Println("ERR: could not add controller", err)
		return
	}
	defer func() {
		select {
		case quitIn1 <- true:
		default:
		}
	}()
	quitOut1 := make(chan bool)
	_, err = s.gm.WatchGame(gameName, quitOut1)
	if err != nil {
		log.Println("ERR: could not add watcher", err)
		return
	}
	defer func() {
		select {
		case quitOut1 <- true:
		default:
		}
	}()

	quitIn2 := make(chan bool)
	gameCtrl2, err := s.gm.ControlGame(gameName, name2, quitIn2)
	if err != nil {
		log.Println("ERR: could not add controller", err)
		return
	}
	defer func() {
		select {
		case quitIn2 <- true:
		default:
		}
	}()

	quitOut2 := make(chan bool)
	_, err = s.gm.WatchGame(gameName, quitOut2)
	if err != nil {
		log.Println("ERR: could not add watcher", err)
		return
	}
	defer func() {
		select {
		case quitOut2 <- true:
		default:
		}
	}()

	file, err := os.Open("dbinterface/replays/" + fullReplayName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// ALSO MIGHT BE VERY SLIGHTLY OFF BECAUSE OF 29 INSTEAD OF 30 CLOCK SPEED (NOTE: AT 30 IT'S INCONSISTENT)
	frameNS := time.Duration(int(1e9) / 29) // hard coded, TODO: make not hardcoded
	clk := time.NewTicker(frameNS)
	for {
		select {
		case <-clk.C:
			if scanner.Scan() {
				command := scanner.Text()
				// fmt.Println("1:" + command)
				gameCtrl1.Input() <- []byte(command)
			}
			if scanner.Scan() {
				command := scanner.Text()
				// fmt.Println("2:" + command)
				gameCtrl2.Input() <- []byte(command)
			}
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}
}

func (s *Server) handleReplayWS(w http.ResponseWriter, r *http.Request) {
	replayName := r.FormValue("game")
	if len(replayName) <= 0 {
		w.Write([]byte("ERR: no replayName provided"))
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ERR: upgrading websocket", err)
		return
	}
	defer conn.Close()

	replay, err := s.db.GetReplay(replayName) // at this point, we've verified that the game exists
	if err != nil {
		log.Println("ERR: no such replay " + replayName + ", err: " + err.Error())
		return
	}

	gameName, err := s.gm.NewOpenGame(true) // the true here indicates that the game is a replay and should not save its own replay
	if err != nil {
		log.Println("ERR: creating game", err)
		return
	}
	go s.playReplayGame(gameName, replay.FullReplayName, replay.Username1, replay.Username2)

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

	userName, err := s.authConn(conn)
	if err != nil {
		log.Println("ERR: could not authenticate connection", err)
		return
	}

	// MAKE GAME IF DNE
	if !s.gm.HasGame(gameName) {
		err = s.gm.NewGame(gameName, false, false)
		if err != nil {
			log.Println("ERR: creating game", err)
		}
	}

	s.playConn(conn, gameName, userName)
}

func (s *Server) handlePlayWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ERR: upgrading websocket", err)
		return
	}
	defer conn.Close()

	userName, err := s.authConn(conn)
	if err != nil {
		log.Println("ERR: could not authenticate connection", err)
		return
	}

	gameName, err := s.gm.PopOpenGame()
	if err != nil {
		gameName, err = s.gm.NewOpenGame(false)
		if err != nil {
			log.Println("ERR: creating game", err)
			return
		}
	}
	if !s.gm.HasGame(gameName) {
		log.Println("ERR: creating game", err)
	}

	log.Printf("SERVER: playws %s\n", gameName)

	s.playConn(conn, gameName, userName)
}

func (s *Server) authConn(conn *websocket.Conn) (string, error) {
	//TODO - timeout on this
	idmt, idmessage, err := conn.ReadMessage()
	if err != nil || idmt == CloseMessage {
		return "", nil
	}
	profile, err := s.db.GetUserFromApiKey(string(idmessage))
	if err != nil {
		return "", nil
	}
	userName := profile.Username
	return userName, nil
}

func (s *Server) playConn(conn *websocket.Conn, gameName, userName string) {
	quit := make(chan bool)
	gameCtrl, err := s.gm.ControlGame(gameName, userName, quit)
	if err != nil {
		log.Println("ERR: could not add controller", err)
		return
	}
	defer func() {
		select {
		case quit <- true:
		default:
		}
	}()

	connectedMsg := struct {
		Player   int
		Username string
		Gamename string
	}{
		gameCtrl.Player(),
		userName,
		gameName,
	}

	conn.WriteJSON(&connectedMsg)

	gameInput := gameCtrl.Input()
	gameOutput := gameCtrl.Output()

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
		t := time.NewTimer(time.Second * 1000)
		select {
		case msg, more := <-gameOutput:
			if !t.Stop() {
				<-t.C
			}
			if more {
				t.Reset(time.Second * 1000)

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

func (s *Server) handleWatch(res http.ResponseWriter, req *http.Request) {
	s.ExecuteUserTemplate(res, req, "watch", Page{Title: "Watch", Data: s.gm.ListGames()})
}

func (s *Server) handleReplays(res http.ResponseWriter, req *http.Request) {
	data, err := s.db.GetReplays()
	if err != nil {
		log.Fatal(err)
	} else {
		s.ExecuteUserTemplate(res, req, "replays", Page{Title: "Replays", Data: data})
	}
}

func (s *Server) handleGame(res http.ResponseWriter, req *http.Request) {
	if cookie, err := req.Cookie("login"); err == nil {
		if username, err := s.db.VerifyCookie(cookie); err == nil {
			if profile, err := s.db.GetUser(username); err == nil {
				s.ExecuteUserTemplate(res, req, "game", Page{Title: "Game", Username: username,
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

func (s *Server) handleHome(res http.ResponseWriter, req *http.Request) {
	s.ExecuteUserTemplate(res, req, "home", Page{Title: "Home"})
}

func (s *Server) handleDocs(res http.ResponseWriter, req *http.Request) {
	s.ExecuteUserTemplate(res, req, "docs", Page{Title: "Documentation"})
}

func (s *Server) handleChangelog(res http.ResponseWriter, req *http.Request) {
	s.ExecuteUserTemplate(res, req, "changelog", Page{Title: "Changelog"})
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
