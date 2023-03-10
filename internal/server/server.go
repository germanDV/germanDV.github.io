package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"germandv.xyz/internal/editor"
	"germandv.xyz/internal/filer"
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
		Handler:      dontPanic(logger(mux)),
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

	if os.Getenv("ENV") == "development" {
		s.registerPreviewHandler()
	}

	log.Printf("Server up on :%d\n", s.port)
	err := s.server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) registerStaticHandler() {
	fs := http.FileServer(http.Dir("./docs"))
	fsWithTimeout := http.TimeoutHandler(fs, 5*time.Second, "Timeout\n")
	s.mux.Handle("/", fsWithTimeout)
}

func (s *Server) registerHealthCheckHandler() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	s.mux.Handle("/healthcheck", http.HandlerFunc(handler))
}

func (s *Server) registerAnalyticsHandler() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("WIP\n"))
	}
	s.mux.Handle("/analytics", basicAuth(http.HandlerFunc(handler)))
}

func (s *Server) registerPreviewHandler() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		file := strings.TrimPrefix(r.URL.Path, "/preview/")
		if file == "" {
			drafts, err := filer.ListDrafts()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			html := "<ul>"
			for _, d := range drafts {
				html += fmt.Sprintf(`<li><a href="%s">%s</a></li>`, d, d)
			}
			html += "</ul>"

			w.Header().Add("Content-Type", "text/html")
			w.Write([]byte(html))
			return
		}

		tmpl, entry, err := editor.Preview(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.ExecuteTemplate(w, "layout", entry)
	}

	s.mux.Handle("/preview/", http.HandlerFunc(handler))
}
