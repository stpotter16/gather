package handlers

import (
	"net/http"

	"github.com/stpotter16/gather/internal/handlers/middleware"
	"github.com/stpotter16/gather/internal/sessions"
	"github.com/stpotter16/gather/internal/store"
)

type Server struct {
	store    store.Store
	sessions *sessions.Manager
}

func NewServer(st store.Store, sm *sessions.Manager) http.Handler {
	s := &Server{store: st, sessions: sm}
	mux := http.NewServeMux()
	s.addRoutes(mux)
	var handler http.Handler = mux
	handler = middleware.CspMiddleware(handler)
	handler = middleware.LoggingWrapper(handler)
	return handler
}
