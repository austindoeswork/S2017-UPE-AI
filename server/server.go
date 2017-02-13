// TODO: split into mini handlers
// SERVER.GO IS IN DESPERATE NEED OF REFACTORING, TOO MUCH STUFF IS HERE
package server

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"os" // when calling ExecuteTemplate you can use os.Stdout instead to output to screen

	"math/rand" // used for the main game TV move generation (MOVE FROM HERE)
	"regexp"
	"strconv"

	"github.com/austindoeswork/S2017-UPE-AI/dbinterface"
	"github.com/austindoeswork/S2017-UPE-AI/gamemanager"
	"github.com/gorilla/sessions"
)

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

// TODO MOVE FROM THIS FILE
// generates a unit move bXX YY, where XX is 00 thru 11 and YY is 01 thru 03
// these are fed into the sample game TV
func generateSampleGameMove() []byte {
	var move []byte
	unitChoice := rand.Intn(12)
	if unitChoice >= 10 {
		move = append(move, []byte(strconv.Itoa(unitChoice))...)
	} else {
		move = append(move, append([]byte("0"), []byte(strconv.Itoa(unitChoice))...)...)
	}
	move = append([]byte("b"), move...)
	laneChoice := rand.Intn(3) + 1
	move = append(move, append([]byte(" 0"), []byte(strconv.Itoa(laneChoice))...)...)
	// fmt.Println(string(move[:]))
	return move
}

// TODO MOVE FROM THIS FILE
// Creates a game TV that will be played on the starting screen of the landing page
func (s *Server) CreateSampleGameTV() {
	gameName := "mainpagegame"
	if !s.gm.HasGame(gameName) {
		err := s.gm.NewGame(gameName, true)
		if err != nil {
			log.Println("ERR: creating game", err)
		}
	}
	quitIn1 := make(chan bool)
	gameCtrl1, err := s.gm.ControlGame(gameName, "mainpageAI1", quitIn1)
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
	gameCtrl2, err := s.gm.ControlGame(gameName, "mainpageAI2", quitIn2)
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

	for {
		if !s.gm.HasGame(gameName) { // with the invincibleCore, this should never end, but just in case??
			return
		}
		time.Sleep(250)
		gameCtrl1.Input() <- generateSampleGameMove()
		gameCtrl2.Input() <- generateSampleGameMove()
	}
}

func New(port, staticDir string, db *dbinterface.DB) *Server {
	m, err := NewMailer("team@aicomp.io")
	if err != nil {
		log.Println(err)
	}
	os.Mkdir("./identicons", 0777)
	s := Server{
		port:      port,
		staticDir: staticDir,
		db:        db,
		gm:        gamemanager.New(),
		store:     sessions.NewCookieStore([]byte("secret")),
		mailer:    m,
	}
	go s.CreateSampleGameTV()
	return &s
}

func (s *Server) Start() {
	s.templates = template.Must(template.ParseGlob("./templates/*.html")) // dynamically load all templates with .html ending

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(s.staticDir))))
	http.HandleFunc("/game", s.handleGame)
	http.HandleFunc("/gamelist", s.handleGameList)
	// http.HandleFunc("/userlist", s.handleUserList)
	http.HandleFunc("/signout", s.handleLogout) // ?? for some reason on my machine if this is logout it doesn't detect it...
	http.HandleFunc("/login", s.handleLogin)
	http.HandleFunc("/signup", s.handleSignup)
	http.HandleFunc("/profile", s.handleProfile)
	http.HandleFunc("/docs", s.handleDocs)
	http.HandleFunc("/wsjoin", s.handleJoinWS)
	http.HandleFunc("/wsplay", s.handlePlayWS)
	http.HandleFunc("/wswatch", s.handleWatchWS)
	http.HandleFunc("/", s.handleHome)

	http.ListenAndServe(s.port, nil)
}

//
// TEMPLATE
//

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
