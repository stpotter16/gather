package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stpotter16/gather/internal/handlers/middleware"
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

func (s *Server) accountGet(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, http.StatusOK, "account.html", newBaseProps(r))
}

func (s *Server) changePasswordPost(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
		ConfirmPassword string `json:"confirm_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request.", http.StatusBadRequest)
		return
	}

	if body.CurrentPassword == "" {
		http.Error(w, "Current password is required.", http.StatusUnprocessableEntity)
		return
	}
	if len(body.NewPassword) < 8 {
		http.Error(w, "New password must be at least 8 characters.", http.StatusUnprocessableEntity)
		return
	}
	if body.NewPassword != body.ConfirmPassword {
		http.Error(w, "Passwords do not match.", http.StatusUnprocessableEntity)
		return
	}

	user, _ := middleware.UserFromContext(r.Context())

	_, currentHash, err := s.store.GetUserByEmail(r.Context(), user.Email)
	if err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	ok, err := password.Verify(body.CurrentPassword, currentHash)
	if err != nil || !ok {
		http.Error(w, "Current password is incorrect.", http.StatusUnprocessableEntity)
		return
	}

	newHash, err := password.Hash(body.NewPassword)
	if err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	if err := s.store.UpdatePassword(r.Context(), user.ID, newHash); err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
