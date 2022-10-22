package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Server struct {
	port int
}

func New(port int) *Server {
	return &Server{port}
}

func (s *Server) Listen() {
	s.registerStaticHandler()
	s.registerTemplateHandler()

	port := fmt.Sprintf(":%d", s.port)
	log.Printf("Server up on %s\n", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) registerStaticHandler() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
}

func (s *Server) registerTemplateHandler() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		lp := filepath.Join("templates", "layout.html")
		fp := filepath.Join("templates", filepath.Clean(r.URL.Path))

		// If requested resource is "/", return index.html
		if r.URL.Path == "/" {
			fp = filepath.Join("templates", "index.html")
		}

		// Check that requested resource exists
		info, err := os.Stat(fp)
		if err != nil {
			if os.IsNotExist(err) {
				http.NotFound(w, r)
				return
			}
		}

		// Check if requested resource is a directory
		if info.IsDir() {
			http.NotFound(w, r)
			return
		}

		tmpl, err := template.ParseFiles(lp, fp)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = tmpl.ExecuteTemplate(w, "layout", nil)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}
