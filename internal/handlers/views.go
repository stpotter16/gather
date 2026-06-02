package handlers

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/stpotter16/gather/internal/handlers/middleware"
	"github.com/stpotter16/gather/internal/store"
)

//go:embed templates
var templateFS embed.FS

var templateFuncs = template.FuncMap{
	"initial": func(s string) string {
		r := []rune(s)
		if len(r) == 0 {
			return "?"
		}
		return string(r[0])
	},
	"hostname": func(rawURL string) string {
		u, err := url.Parse(rawURL)
		if err != nil || u.Host == "" {
			return rawURL
		}
		return u.Hostname()
	},
	"lower": strings.ToLower,
	"cookNames": func(cooks []store.MealCook) string {
		names := make([]string, len(cooks))
		for i, c := range cooks {
			parts := strings.Fields(c.Name)
			if len(parts) > 0 {
				names[i] = parts[0]
			}
		}
		return strings.Join(names, ", ")
	},
	"formatTime": func(s string) string {
		if s == "" {
			return ""
		}
		t, err := time.Parse("15:04:05", s)
		if err != nil {
			return s
		}
		return t.Format("3:04 pm")
	},
}

type baseProps struct {
	CspNonce string
	User     store.User
}

func newBaseProps(r *http.Request) baseProps {
	nonce, _ := middleware.NonceFromContext(r.Context())
	user, _ := middleware.UserFromContext(r.Context())
	return baseProps{CspNonce: nonce, User: user}
}

// renderAuthPage renders a page using only base.html (no app nav).
func renderAuthPage(w http.ResponseWriter, r *http.Request, status int, page string, data any) {
	tmpl, err := template.New("base.html").Funcs(templateFuncs).ParseFS(
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
	tmpl, err := template.New("base.html").Funcs(templateFuncs).ParseFS(
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
