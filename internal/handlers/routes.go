package handlers

import "net/http"

func addRoutes(mux *http.ServeMux) {
	mux.Handle("GET /static/", http.StripPrefix("/static/", serveStaticFiles()))
	mux.HandleFunc("GET /{$}", indexGet())
}
