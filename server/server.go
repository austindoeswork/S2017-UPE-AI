// TODO: split into mini handlers
package server

import (
	"html/template"
	"net/http"

	// "os" // when calling ExecuteTemplate you can use os.Stdout instead to output to screen

	"github.com/austindoeswork/S2017-UPE-AI/dbinterface"
	"github.com/austindoeswork/S2017-UPE-AI/gamemanager"
)

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

func New(port, staticDir string, db *dbinterface.DB) *Server {
	return &Server{
		port:      port,
		staticDir: staticDir,
		db:        db,
		gm:        gamemanager.New(),
	}
}

func (s *Server) Start() {
	s.templates = template.Must(template.ParseGlob("./templates/*.html")) // dynamically load all templates with .html ending

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(s.staticDir))))
	// handler.RegisterStaticHandler("/static/", "/static/", s.staticDir)
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

	http.ListenAndServe(s.port, nil)
}
