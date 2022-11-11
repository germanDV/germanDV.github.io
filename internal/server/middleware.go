package server

import (
	"compress/gzip"
	"log"
	"net/http"
	"strings"
)

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.Method, r.RequestURI, r.Proto)
		next.ServeHTTP(w, r)
	})
}

func dontPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Paaaanic in %s %s: %s", r.Method, r.RequestURI, err)
				w.Header().Set("Connection", "close")
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
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

type WrappedResponseWriter struct {
	rw http.ResponseWriter
	gw *gzip.Writer
}

func (wr *WrappedResponseWriter) Header() http.Header {
	return wr.rw.Header()
}
func (wr *WrappedResponseWriter) Write(bytes []byte) (int, error) {
	return wr.gw.Write(bytes) // Use gzip writer.
}
func (wr *WrappedResponseWriter) WriteHeader(statusCode int) {
	wr.rw.WriteHeader(statusCode)
}
func (wr *WrappedResponseWriter) Flush() {
	wr.gw.Flush()
	wr.gw.Close()
}

func NewWrappedResponseWriter(w http.ResponseWriter) *WrappedResponseWriter {
	return &WrappedResponseWriter{
		rw: w,
		gw: gzip.NewWriter(w),
	}
}

func gzipper(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		wrapped := NewWrappedResponseWriter(w)
		wrapped.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(wrapped, r)
		defer wrapped.Flush()
	})
}
