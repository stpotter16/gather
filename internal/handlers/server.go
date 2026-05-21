package handlers

import (
	"net/http"

	"github.com/stpotter16/gather/internal/handlers/middleware"
)

func NewServer() http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux)
	var handler http.Handler = mux
	handler = middleware.CspMiddleware(handler)
	handler = middleware.LoggingWrapper(handler)
	return handler
}
