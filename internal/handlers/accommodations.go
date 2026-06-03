package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) accommodationCreatePost(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	user, ok := s.requireMember(w, r, eventID)
	if !ok {
		return
	}

	var body struct {
		Label string `json:"label"`
		URL   string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request.", http.StatusBadRequest)
		return
	}
	body.Label = strings.TrimSpace(body.Label)
	body.URL = strings.TrimSpace(body.URL)
	if body.Label == "" {
		http.Error(w, "Label is required.", http.StatusUnprocessableEntity)
		return
	}
	if body.URL == "" {
		http.Error(w, "URL is required.", http.StatusUnprocessableEntity)
		return
	}

	id, err := s.store.AddAccommodation(r.Context(), eventID, user.ID, body.Label, body.URL)
	if err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"id":` + strconv.Itoa(id) + `}`))
}

func (s *Server) accommodationDeleteDelete(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if _, ok := s.requireMember(w, r, eventID); !ok {
		return
	}
	accommodationID, ok := parsePathInt(r, "accommodationID")
	if !ok {
		http.NotFound(w, r)
		return
	}

	if err := s.store.DeleteAccommodation(r.Context(), accommodationID, eventID); err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
