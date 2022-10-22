package server

import (
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	port int
}

func New(port int) *Server {
	return &Server{port}
}

func (s *Server) Listen() {
	s.registerStaticHandler()

	port := fmt.Sprintf(":%d", s.port)
	log.Printf("Server up on %s\n", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) registerStaticHandler() {
	fs := http.FileServer(http.Dir("./pages"))
	// http.Handle("/pages/", http.StripPrefix("/pages/", fs))
	http.Handle("/", fs)
}
