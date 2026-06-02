package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) activityCreatePost(w http.ResponseWriter, r *http.Request) {
	id, ok := parseEventID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	user, ok := s.requireMember(w, r, id)
	if !ok {
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

func (s *Server) activityConfirmPost(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	activityID, ok := parsePathInt(r, "activityID")
	if !ok {
		http.NotFound(w, r)
		return
	}
	if _, ok := s.requireMember(w, r, eventID); !ok {
		return
	}
	if err := s.store.ConfirmActivity(r.Context(), activityID, eventID); err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) activityVotePost(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	activityID, ok := parsePathInt(r, "activityID")
	if !ok {
		http.NotFound(w, r)
		return
	}
	user, ok := s.requireMember(w, r, eventID)
	if !ok {
		return
	}

	if err := s.store.ToggleActivityVote(r.Context(), activityID, user.ID); err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
