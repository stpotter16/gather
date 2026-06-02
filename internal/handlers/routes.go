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
	protected.HandleFunc("GET /events/new", s.newEventGet)
	protected.HandleFunc("POST /events/new", s.newEventPost)
	protected.HandleFunc("GET /events/{id}", s.eventDetailGet)
	protected.HandleFunc("POST /events/{id}/rsvp", s.eventRSVPPost)
	mux.Handle("/", middleware.RequireAuth(s.sessions, s.store, protected))
}
