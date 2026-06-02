package handlers

import (
	"net/http"
	"strconv"

	"github.com/stpotter16/gather/internal/handlers/middleware"
	"github.com/stpotter16/gather/internal/store"
)

func parsePathInt(r *http.Request, key string) (int, bool) {
	v, err := strconv.Atoi(r.PathValue(key))
	if err != nil || v <= 0 {
		return 0, false
	}
	return v, true
}

func parseEventID(r *http.Request) (int, bool) {
	return parsePathInt(r, "id")
}

// requireMember verifies the current user is a member of eventID. On success it
// returns the user and true. On failure it writes the appropriate error response
// and returns false — the caller must return immediately.
func (s *Server) requireMember(w http.ResponseWriter, r *http.Request, eventID int) (store.User, bool) {
	user, _ := middleware.UserFromContext(r.Context())
	ok, err := s.store.IsEventMember(r.Context(), eventID, user.ID)
	if err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return store.User{}, false
	}
	if !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return store.User{}, false
	}
	return user, true
}
