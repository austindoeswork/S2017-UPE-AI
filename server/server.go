package server

import (
	"log"
	"net/http"
	// "github.com/gorilla/websocket"
)

type Server struct {
	port      string
	staticDir string
}

func New(port, staticDir string) *Server {
	return &Server{
		port:      port,
		staticDir: staticDir,
	}
}

func (s *Server) Start() {
	// http.Handle("/asdf/", http.StripPrefix("/asdf/", http.FileServer(http.Dir(staticdir))))
	http.Handle("/", http.FileServer(http.Dir(s.staticDir)))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.FormValue("game"))
	})
	http.HandleFunc("/darwin", func(w http.ResponseWriter, r *http.Request) {
		log.Println("sup bro.")
	})
	http.ListenAndServe(s.port, nil)
}
