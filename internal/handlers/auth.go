package handlers

import (
	"net/http"

	"github.com/stpotter16/gather/internal/password"
)

type loginProps struct {
	baseProps
	Error string
}

func (s *Server) loginGet(w http.ResponseWriter, r *http.Request) {
	renderAuthPage(w, r, http.StatusOK, "login.html", loginProps{baseProps: newBaseProps(r)})
}

func (s *Server) loginPost(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	pw := r.FormValue("password")

	fail := func() {
		renderAuthPage(w, r, http.StatusUnauthorized, "login.html", loginProps{
			baseProps: newBaseProps(r),
			Error:     "Invalid email or password.",
		})
	}

	user, err := s.store.GetUserByEmail(r.Context(), email)
	if err != nil {
		fail()
		return
	}

	ok, err := password.Verify(pw, user.PasswordHash)
	if err != nil || !ok {
		fail()
		return
	}

	s.sessions.Set(w, user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) logoutPost(w http.ResponseWriter, r *http.Request) {
	s.sessions.Clear(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
