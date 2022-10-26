package server

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

type Server struct {
	mux    *http.ServeMux
	server *http.Server
	port   int
}

func New(port int) *Server {
	mux := &http.ServeMux{}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      mux,
	}

	return &Server{
		server: server,
		mux:    mux,
		port:   port,
	}
}

func (s *Server) Listen() {
	s.registerHealthCheckHandler()
	s.registerAnalyticsHandler()
	s.registerStaticHandler()
	s.registerIndexHandler()

	log.Printf("Server up on :%d\n", s.port)
	err := s.server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) registerStaticHandler() {
	fs := http.FileServer(http.Dir("./pages"))
	fsWithTimeout := http.TimeoutHandler(fs, 5*time.Second, "Timeout\n")
	s.mux.Handle("/blog/", http.StripPrefix("/blog/", logger(fsWithTimeout)))
}

func (s *Server) registerIndexHandler() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("pages", "index.html"))
	}
	s.mux.Handle("/", logger(http.HandlerFunc(handler)))
}

func (s *Server) registerHealthCheckHandler() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	s.mux.Handle("/health_check", logger(http.HandlerFunc(handler)))
}

func (s *Server) registerAnalyticsHandler() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Top secret data\n"))
	}
	s.mux.Handle("/analytics", logger(basicAuth(http.HandlerFunc(handler))))
}
