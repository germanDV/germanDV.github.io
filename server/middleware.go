package server

import (
	"log"
	"net/http"
)

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.Method, r.RequestURI, r.Proto)
		next.ServeHTTP(w, r)
	})
}

func basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()

		if !ok {
			w.Header().Add("WWW-Authenticate", "Basic")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("This endpoint requires authentication.\n"))
			return
		}

		if !grantPermission(username, password) {
			http.Error(w, "Bad credentials.", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
