package handlers

import (
	"net/http"

	"github.com/stpotter16/gather/internal/handlers/middleware"
)

func (s *Server) addRoutes(mux *http.ServeMux) {
	mux.Handle("GET /static/", http.StripPrefix("/static/", serveStaticFiles()))

	mux.HandleFunc("GET /login", s.loginGet)
	mux.HandleFunc("POST /login", s.loginPost)
	mux.HandleFunc("POST /logout", s.logoutPost)

	protected := http.NewServeMux()
	protected.HandleFunc("GET /{$}", s.indexGet)
	mux.Handle("/", middleware.RequireAuth(s.sessions, s.store, protected))
}
