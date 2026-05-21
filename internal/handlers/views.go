package handlers

import (
	"embed"
	"html/template"
	"log"
	"net/http"

	"github.com/stpotter16/gather/internal/handlers/middleware"
)

//go:embed templates
var templateFS embed.FS

type baseProps struct {
	CspNonce string
}

func newBaseProps(r *http.Request) baseProps {
	nonce, _ := middleware.NonceFromContext(r.Context())
	return baseProps{CspNonce: nonce}
}

// renderAuthPage renders a page using only base.html (no app nav).
func renderAuthPage(w http.ResponseWriter, r *http.Request, status int, page string, data any) {
	tmpl, err := template.New("base.html").ParseFS(
		templateFS,
		"templates/layouts/base.html",
		"templates/pages/"+page,
	)
	if err != nil {
		log.Printf("template parse error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("template execute error: %v", err)
	}
}

func renderPage(w http.ResponseWriter, r *http.Request, status int, page string, data any) {
	tmpl, err := template.New("base.html").ParseFS(
		templateFS,
		"templates/layouts/base.html",
		"templates/layouts/app.html",
		"templates/pages/"+page,
	)
	if err != nil {
		log.Printf("template parse error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("template execute error: %v", err)
	}
}
