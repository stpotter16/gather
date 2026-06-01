package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stpotter16/gather/internal/password"
)

func (s *Server) loginGet(w http.ResponseWriter, r *http.Request) {
	renderAuthPage(w, r, http.StatusOK, "login.html", newBaseProps(r))
}

func (s *Server) loginPost(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request.", http.StatusBadRequest)
		return
	}

	fail := func() {
		http.Error(w, "Invalid email or password.", http.StatusUnauthorized)
	}

	user, hash, err := s.store.GetUserByEmail(r.Context(), body.Email)
	if err != nil {
		fail()
		return
	}

	ok, err := password.Verify(body.Password, hash)
	if err != nil || !ok {
		fail()
		return
	}

	s.sessions.Set(w, user.ID)
	fmt.Fprint(w, "OK")
}

func (s *Server) logoutPost(w http.ResponseWriter, r *http.Request) {
	s.sessions.Clear(w)
	w.WriteHeader(http.StatusNoContent)
}
