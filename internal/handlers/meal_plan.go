package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stpotter16/gather/internal/handlers/middleware"
)

func (s *Server) foodRestrictionUpsertPost(w http.ResponseWriter, r *http.Request) {
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
		Restriction string `json:"restriction"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request.", http.StatusBadRequest)
		return
	}
	body.Restriction = strings.TrimSpace(body.Restriction)
	if body.Restriction == "" {
		http.Error(w, "Restriction cannot be empty.", http.StatusUnprocessableEntity)
		return
	}

	if err := s.store.UpsertFoodRestriction(r.Context(), id, user.ID, body.Restriction); err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) mealCreatePost(w http.ResponseWriter, r *http.Request) {
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
		Name string `json:"name"`
		Date string `json:"date"` // YYYY-MM-DD
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
	date, err := time.Parse("2006-01-02", body.Date)
	if err != nil {
		http.Error(w, "Invalid date.", http.StatusUnprocessableEntity)
		return
	}

	mealID, err := s.store.CreateMeal(r.Context(), id, body.Name, date)
	if err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"id":` + strconv.Itoa(mealID) + `}`))
}

func (s *Server) dishCreatePost(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	user, _ := middleware.UserFromContext(r.Context())
	if ok, _ := s.store.IsEventMember(r.Context(), eventID, user.ID); !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	mealID, err := strconv.Atoi(r.PathValue("mealID"))
	if err != nil || mealID <= 0 {
		http.NotFound(w, r)
		return
	}

	var body struct {
		Name string `json:"name"`
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

	dishID, err := s.store.AddDish(r.Context(), mealID, body.Name)
	if err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"id":` + strconv.Itoa(dishID) + `}`))
}

func (s *Server) groceryCreatePost(w http.ResponseWriter, r *http.Request) {
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
		Name     string `json:"name"`
		Category string `json:"category"`
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
	if body.Category != "buy" && body.Category != "bring" {
		http.Error(w, "Invalid category.", http.StatusUnprocessableEntity)
		return
	}

	groceryID, err := s.store.AddGrocery(r.Context(), id, body.Name, body.Category)
	if err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"id":` + strconv.Itoa(groceryID) + `}`))
}

func (s *Server) mealDeleteDelete(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	user, _ := middleware.UserFromContext(r.Context())
	if ok, _ := s.store.IsEventMember(r.Context(), eventID, user.ID); !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	mealID, err := strconv.Atoi(r.PathValue("mealID"))
	if err != nil || mealID <= 0 {
		http.NotFound(w, r)
		return
	}
	if err := s.store.DeleteMeal(r.Context(), mealID, eventID); err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) dishDeleteDelete(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	user, _ := middleware.UserFromContext(r.Context())
	if ok, _ := s.store.IsEventMember(r.Context(), eventID, user.ID); !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	mealID, err := strconv.Atoi(r.PathValue("mealID"))
	if err != nil || mealID <= 0 {
		http.NotFound(w, r)
		return
	}
	dishID, err := strconv.Atoi(r.PathValue("dishID"))
	if err != nil || dishID <= 0 {
		http.NotFound(w, r)
		return
	}
	if err := s.store.DeleteDish(r.Context(), dishID, mealID); err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) groceryDeleteDelete(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	user, _ := middleware.UserFromContext(r.Context())
	if ok, _ := s.store.IsEventMember(r.Context(), eventID, user.ID); !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	groceryID, err := strconv.Atoi(r.PathValue("groceryID"))
	if err != nil || groceryID <= 0 {
		http.NotFound(w, r)
		return
	}
	if err := s.store.DeleteGrocery(r.Context(), groceryID, eventID); err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) cookAssignPost(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	user, _ := middleware.UserFromContext(r.Context())
	if ok, _ := s.store.IsEventMember(r.Context(), eventID, user.ID); !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	mealID, err := strconv.Atoi(r.PathValue("mealID"))
	if err != nil || mealID <= 0 {
		http.NotFound(w, r)
		return
	}

	var body struct {
		UserID int `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request.", http.StatusBadRequest)
		return
	}
	if body.UserID <= 0 {
		http.Error(w, "user_id is required.", http.StatusUnprocessableEntity)
		return
	}

	if err := s.store.AssignCook(r.Context(), mealID, body.UserID); err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) cookRemoveDelete(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	user, _ := middleware.UserFromContext(r.Context())
	if ok, _ := s.store.IsEventMember(r.Context(), eventID, user.ID); !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	mealID, err := strconv.Atoi(r.PathValue("mealID"))
	if err != nil || mealID <= 0 {
		http.NotFound(w, r)
		return
	}
	cookUserID, err := strconv.Atoi(r.PathValue("userID"))
	if err != nil || cookUserID <= 0 {
		http.NotFound(w, r)
		return
	}

	if err := s.store.RemoveCook(r.Context(), mealID, cookUserID); err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) groceryTogglePost(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}

	groceryID, err := strconv.Atoi(r.PathValue("groceryID"))
	if err != nil || groceryID <= 0 {
		http.NotFound(w, r)
		return
	}

	if err := s.store.ToggleGrocery(r.Context(), groceryID, eventID); err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
