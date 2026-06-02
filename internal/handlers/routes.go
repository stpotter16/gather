package handlers

import (
	"net/http"

	"github.com/stpotter16/gather/internal/handlers/middleware"
)

func (s *Server) addRoutes(mux *http.ServeMux) {
	mux.Handle("GET /static/", http.StripPrefix("/static/", serveStaticFiles()))

	mux.HandleFunc("GET /login", s.loginGet)
	mux.HandleFunc("POST /login", s.loginPost)
	mux.HandleFunc("POST /logout", s.logoutPost)

	protected := http.NewServeMux()
	protected.HandleFunc("GET /{$}", s.indexGet)
	protected.HandleFunc("GET /account", s.accountGet)
	protected.HandleFunc("POST /account/password", s.changePasswordPost)
	protected.HandleFunc("GET /events/new", s.newEventGet)
	protected.HandleFunc("POST /events/new", s.newEventPost)
	protected.HandleFunc("GET /events/{id}", s.eventDetailGet)
	protected.HandleFunc("GET /events/{id}/edit", s.editEventGet)
	protected.HandleFunc("PUT /events/{id}", s.editEventPut)
	protected.HandleFunc("POST /events/{id}/rsvp", s.eventRSVPPost)
	protected.HandleFunc("PUT /events/{id}/itinerary", s.itineraryUpsertPut)
	protected.HandleFunc("POST /events/{id}/food-restrictions", s.foodRestrictionUpsertPost)
	protected.HandleFunc("POST /events/{id}/meals", s.mealCreatePost)
	protected.HandleFunc("DELETE /events/{id}/meals/{mealID}", s.mealDeleteDelete)
	protected.HandleFunc("POST /events/{id}/meals/{mealID}/dishes", s.dishCreatePost)
	protected.HandleFunc("DELETE /events/{id}/meals/{mealID}/dishes/{dishID}", s.dishDeleteDelete)
	protected.HandleFunc("POST /events/{id}/meals/{mealID}/cooks", s.cookAssignPost)
	protected.HandleFunc("DELETE /events/{id}/meals/{mealID}/cooks/{userID}", s.cookRemoveDelete)
	protected.HandleFunc("POST /events/{id}/groceries", s.groceryCreatePost)
	protected.HandleFunc("POST /events/{id}/groceries/{groceryID}/toggle", s.groceryTogglePost)
	protected.HandleFunc("DELETE /events/{id}/groceries/{groceryID}", s.groceryDeleteDelete)
	protected.HandleFunc("POST /events/{id}/activities", s.activityCreatePost)
	protected.HandleFunc("POST /events/{id}/activities/{activityID}/confirm", s.activityConfirmPost)
	protected.HandleFunc("POST /events/{id}/activities/{activityID}/vote", s.activityVotePost)
	protected.HandleFunc("DELETE /events/{id}/membership", s.eventLeaveDelete)
	protected.HandleFunc("POST /events/{id}/invites", s.invitePost)
	mux.Handle("/", middleware.RequireAuth(s.sessions, s.store, protected))
}
