// TODO: split into mini handlers
// SERVER.GO IS IN DESPERATE NEED OF REFACTORING, TOO MUCH STUFF IS HERE
package server

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"os" // when calling ExecuteTemplate you can use os.Stdout instead to output to screen

	"bytes"     // used for main game TV move generation (MOVE FROM HERE)
	"math/rand" // used for the main game TV move generation (MOVE FROM HERE)

	"regexp"
	"strconv"

	"github.com/austindoeswork/S2017-UPE-AI/dbinterface"
	"github.com/austindoeswork/S2017-UPE-AI/gamemanager"
	"github.com/gorilla/sessions"

	// following imports used for dynamic template reloading
	"github.com/fsnotify/fsnotify" // adds watchers to watch for template change
	"sync"                         // mutex to prevent people from accessing templates during change
)

// isAlpha checks if the given string contains only alphabetic characters
var isAlpha = regexp.MustCompile(`^[A-Za-z\s]+$`).MatchString

// isAlphaNumeric checks if the given string contains only alphanumeric characters
var isAlphaNumeric = regexp.MustCompile(`^[A-Za-z\d]+$`).MatchString

// validFilenameCharacters is a regex that matches characters that are not allowed in filenames
var validFilenameCharacters = regexp.MustCompile("[\\\\/:\"*?<>|]+")

// Server handles websockets and creation of games
// TODO: simplify server class? or just document better
type Server struct {
	port      string
	staticDir string
	db        *dbinterface.DB
	gm        *gamemanager.GameManager
	store     *sessions.CookieStore
	mailer    *Mailer
	tw        *TemplateWaiter // dynamic template reloading and serving
}

// TODO MOVE FROM THIS FILE
// generates either a troop move bXX YY, where XX is 00 thru 11 and YY is 01 thru 03
// or generates a tower move bXX YY, where XX is 50 thru 59 and YY is 01 thru 65
// these are fed into the sample game TV
func generateSampleGameMove() []byte {
	var buffer bytes.Buffer
	buffer.Write([]byte("b"))
	troopOrTower := rand.Intn(4)
	if troopOrTower >= 1 { // make troop
		troopChoice := rand.Intn(12)
		for troopChoice == 4 { // aimbots are not fun to watch T B H
			troopChoice = rand.Intn(12)
		}
		if troopChoice >= 10 {
			buffer.WriteString(strconv.Itoa(troopChoice))
		} else {
			buffer.Write([]byte("0"))
			buffer.WriteString(strconv.Itoa(troopChoice))
		}
		buffer.Write([]byte(" 0"))
		buffer.WriteString(strconv.Itoa(rand.Intn(3) + 1)) // lane choice
	} else { // make tower
		towerChoice := rand.Intn(10) + 50
		plotChoice := rand.Intn(66)
		buffer.WriteString(strconv.Itoa(towerChoice))
		buffer.Write([]byte(" "))
		if plotChoice >= 10 {
			buffer.WriteString(strconv.Itoa(plotChoice))
		} else {
			buffer.Write([]byte("0"))
			buffer.WriteString(strconv.Itoa(plotChoice))
		}
	}
	return buffer.Bytes()
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
	gameCtrl1, err := s.gm.ControlGame(gameName, "HAL 9000", quitIn1)
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
	gameCtrl2, err := s.gm.ControlGame(gameName, "Deep Blue", quitIn2)
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
		time.Sleep(1000) // how expensive is this? maybe replace with something else?
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
		gm:        gamemanager.New(db),
		store:     sessions.NewCookieStore([]byte("secret")),
		mailer:    m,
		tw:        NewTemplateWaiter(),
	}
	go s.CreateSampleGameTV()
	return &s
}

func (s *Server) Start() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(s.staticDir))))
	http.HandleFunc("/game", s.handleGame)
	http.HandleFunc("/watch", s.handleWatch)
	http.HandleFunc("/leaderboard", s.handleLeaderboard)
	http.HandleFunc("/replays", s.handleReplayList)
	http.HandleFunc("/changelog", s.handleChangelog)
	http.HandleFunc("/signout", s.handleLogout) // ?? for some reason on my machine if this is logout it doesn't detect it...
	http.HandleFunc("/login", s.handleLogin)
	http.HandleFunc("/signup", s.handleSignup)
	http.HandleFunc("/profile", s.handleProfile)
	http.HandleFunc("/docs", s.handleDocs)
	http.HandleFunc("/wsjoin", s.handleJoinWS)
	http.HandleFunc("/wsplay", s.handlePlayWS)
	http.HandleFunc("/wswatch", s.handleWatchWS)
	http.HandleFunc("/wsreplay", s.handleReplayWS)
	http.HandleFunc("/", s.handleHome)

	err := http.ListenAndServe(s.port, nil)
	log.Panic(err)
}

//
// TEMPLATE
//

// This data is passed into templates so that we can have dynamic information
type Page struct {
	Title    string      // title of page
	Flash    []string    // flash message
	Username string      // username of user
	Data     interface{} // any additional data
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
	err = s.tw.ExecuteTemplate(res, template, data)
	if err != nil {
		log.Fatal("Cannot Get View ", err)
	}
}

// TemplateWaiter holds all of the templates, ready to be served.
type TemplateWaiter struct {
	templates *template.Template
	mutex     sync.RWMutex
	watcher   *fsnotify.Watcher
}

func NewTemplateWaiter() *TemplateWaiter {
	tw := TemplateWaiter{
		templates: template.Must(template.ParseGlob("./templates/*.html")),
		watcher:   nil,
	}
	tw.Watch()
	return &tw
}

func (tw *TemplateWaiter) Watch() {
	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()
		err = watcher.Add("./templates/")
		if err != nil {
			log.Fatal(err)
		}
		tw.watcher = watcher

		for {
			select {
			case event := <-tw.watcher.Events:
				// log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					tw.mutex.Lock()
					tw.templates = template.Must(template.ParseGlob("./templates/*.html"))
					tw.mutex.Unlock()
				}
			case err := <-tw.watcher.Errors:
				if err != nil {
					log.Println("error:", err)
				}
			}
		}
	}()
}

func (tw *TemplateWaiter) ExecuteTemplate(res http.ResponseWriter, template string, data Page) error {
	tw.mutex.Lock()
	defer tw.mutex.Unlock()
	return tw.templates.ExecuteTemplate(res, template, data)
}
