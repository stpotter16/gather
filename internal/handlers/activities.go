package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/stpotter16/gather/internal/handlers/middleware"
)

func (s *Server) activityCreatePost(w http.ResponseWriter, r *http.Request) {
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
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request.", http.StatusBadRequest)
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" {
		http.Error(w, "Name is required.", http.StatusUnprocessableEntity)
		return
	}

	activityID, err := s.store.CreateActivity(r.Context(), id, user.ID, body.Name, strings.TrimSpace(body.Description))
	if err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"id":` + strconv.Itoa(activityID) + `}`))
}

func (s *Server) activityVotePost(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	activityID, err := strconv.Atoi(r.PathValue("activityID"))
	if err != nil || activityID <= 0 {
		http.NotFound(w, r)
		return
	}

	user, _ := middleware.UserFromContext(r.Context())
	if ok, _ := s.store.IsEventMember(r.Context(), eventID, user.ID); !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := s.store.ToggleActivityVote(r.Context(), activityID, user.ID); err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
