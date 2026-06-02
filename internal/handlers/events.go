package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/stpotter16/gather/internal/handlers/middleware"
	"github.com/stpotter16/gather/internal/store"
)

type editEventProps struct {
	baseProps
	Event store.EventDetail
}

func (s *Server) newEventGet(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, http.StatusOK, "new_event.html", newBaseProps(r))
}

func (s *Server) editEventGet(w http.ResponseWriter, r *http.Request) {
	id, ok := parseEventID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}

	detail, err := s.store.GetEventDetail(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user, _ := middleware.UserFromContext(r.Context())
	isMember := false
	for _, m := range detail.Members {
		if m.UserID == user.ID {
			isMember = true
			break
		}
	}
	if !isMember {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	renderPage(w, r, http.StatusOK, "edit_event.html", editEventProps{
		baseProps: newBaseProps(r),
		Event:     detail,
	})
}

func (s *Server) editEventPut(w http.ResponseWriter, r *http.Request) {
	id, ok := parseEventID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}

	if _, ok := s.requireMember(w, r, id); !ok {
		return
	}

	var body struct {
		Name        string `json:"name"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		Location    string `json:"location"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request.", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(body.Name)
	location := strings.TrimSpace(body.Location)
	description := strings.TrimSpace(body.Description)

	if name == "" {
		http.Error(w, "Name is required.", http.StatusUnprocessableEntity)
		return
	}
	if body.StartDate == "" {
		http.Error(w, "Start date is required.", http.StatusUnprocessableEntity)
		return
	}
	if body.EndDate == "" {
		http.Error(w, "End date is required.", http.StatusUnprocessableEntity)
		return
	}
	if location == "" {
		http.Error(w, "Location is required.", http.StatusUnprocessableEntity)
		return
	}

	startDate, err := time.Parse("2006-01-02", body.StartDate)
	if err != nil {
		http.Error(w, "Invalid start date.", http.StatusUnprocessableEntity)
		return
	}
	endDate, err := time.Parse("2006-01-02", body.EndDate)
	if err != nil {
		http.Error(w, "Invalid end date.", http.StatusUnprocessableEntity)
		return
	}
	if endDate.Before(startDate) {
		http.Error(w, "End date must be on or after start date.", http.StatusUnprocessableEntity)
		return
	}

	if err := s.store.UpdateEvent(r.Context(), id, name, location, description, startDate, endDate); err != nil {
		http.Error(w, "Something went wrong. Please try again.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) newEventPost(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		Location    string `json:"location"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request.", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(body.Name)
	location := strings.TrimSpace(body.Location)
	description := strings.TrimSpace(body.Description)

	if name == "" {
		http.Error(w, "Name is required.", http.StatusUnprocessableEntity)
		return
	}
	if body.StartDate == "" {
		http.Error(w, "Start date is required.", http.StatusUnprocessableEntity)
		return
	}
	if body.EndDate == "" {
		http.Error(w, "End date is required.", http.StatusUnprocessableEntity)
		return
	}
	if location == "" {
		http.Error(w, "Location is required.", http.StatusUnprocessableEntity)
		return
	}

	startDate, err := time.Parse("2006-01-02", body.StartDate)
	if err != nil {
		http.Error(w, "Invalid start date.", http.StatusUnprocessableEntity)
		return
	}
	endDate, err := time.Parse("2006-01-02", body.EndDate)
	if err != nil {
		http.Error(w, "Invalid end date.", http.StatusUnprocessableEntity)
		return
	}
	if endDate.Before(startDate) {
		http.Error(w, "End date must be on or after start date.", http.StatusUnprocessableEntity)
		return
	}

	user, _ := middleware.UserFromContext(r.Context())
	id, err := s.store.CreateEvent(r.Context(), name, location, description, startDate, endDate, user.ID)
	if err != nil {
		http.Error(w, "Something went wrong. Please try again.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"id":%d}`, id)
}
