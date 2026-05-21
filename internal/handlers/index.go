package handlers

import "net/http"

func (s *Server) indexGet(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, http.StatusOK, "index.html", newBaseProps(r))
}
