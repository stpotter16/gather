package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/stpotter16/gather/internal/handlers/middleware"
)

func (s *Server) invitePost(w http.ResponseWriter, r *http.Request) {
	id, ok := parseEventID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	user, _ := middleware.UserFromContext(r.Context())
	if ok, _ := s.store.IsEventMember(r.Context(), id, user.ID); !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var body struct {
		UserIDs []int `json:"user_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request.", http.StatusBadRequest)
		return
	}
	if len(body.UserIDs) == 0 {
		http.Error(w, "Select at least one person to invite.", http.StatusUnprocessableEntity)
		return
	}

	if err := s.store.InviteUsers(r.Context(), id, user.ID, body.UserIDs); err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
