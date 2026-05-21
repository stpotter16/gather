package handlers

import "net/http"

func indexGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderPage(w, r, http.StatusOK, "index.html", newBaseProps(r))
	}
}
