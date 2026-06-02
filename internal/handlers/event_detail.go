package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/stpotter16/gather/internal/handlers/middleware"
	"github.com/stpotter16/gather/internal/store"
)

type timelineRow struct {
	Name        string
	AvatarColor string
	LeftPct     int
	WidthPct    int
}

type memberView struct {
	store.EventDetailMember
	IsHost      bool
	DateSummary string
	InvitedAgo  string
}

type mealDayView struct {
	Label string // e.g. "Saturday · Jul 4"
	Meals []store.Meal
}

type eventDetailProps struct {
	baseProps
	Detail               store.EventDetail
	DateRange            string
	DayCount             int
	TimelineCols         []string
	TimelineRows         []timelineRow
	HeaderAvatars        []memberView
	HeaderExtra          int
	Going                []memberView
	Pending              []memberView
	Declined             []memberView
	GoingCount           int
	PendingCount         int
	DeclinedCount        int
	CurrentStatus         string
	InvitedBy             string
	HasCurrentItinerary   bool
	CurrentItineraryJSON  template.JS
	GoingMembersJSON      template.JS
	// Invite
	InviteCandidates []store.InviteCandidate
	// Activities
	Confirmed  []store.Activity
	Ideas      []store.Activity
	// Meal plan
	Restrictions         []store.FoodRestriction
	MealDays             []mealDayView
	ToBuy                []store.GroceryItem
	ToBring              []store.GroceryItem
	CurrentRestriction   string
	HasRestriction       bool
}

func (s *Server) eventDetailGet(w http.ResponseWriter, r *http.Request) {
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

	var currentStatus, invitedBy string
	var currentMember store.EventDetailMember
	isMember := false
	for _, m := range detail.Members {
		if m.UserID == user.ID {
			isMember = true
			currentStatus = m.Status
			invitedBy = m.InvitedByName
			currentMember = m
			break
		}
	}
	if !isMember {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	duration := int(detail.EndDate.Sub(detail.StartDate).Hours()/24) + 1

	var timelineCols []string
	for i := range duration {
		timelineCols = append(timelineCols, detail.StartDate.AddDate(0, 0, i).Format("Jan 2"))
	}

	var timelineRows []timelineRow
	for _, m := range detail.Members {
		if m.Status != "going" {
			continue
		}
		arrival := detail.StartDate
		departure := detail.EndDate
		if m.ArrivalDate != nil {
			arrival = *m.ArrivalDate
		}
		if m.DepartureDate != nil {
			departure = *m.DepartureDate
		}
		leftPct := int(arrival.Sub(detail.StartDate).Hours() / 24 / float64(duration) * 100)
		widthPct := int((departure.Sub(arrival).Hours()/24+1) / float64(duration) * 100)
		timelineRows = append(timelineRows, timelineRow{
			Name:        m.Name,
			AvatarColor: m.AvatarColor,
			LeftPct:     leftPct,
			WidthPct:    widthPct,
		})
	}

	var going, pending, declined []memberView
	for _, m := range detail.Members {
		mv := memberView{
			EventDetailMember: m,
			IsHost:            m.UserID == detail.CreatedBy,
		}
		switch m.Status {
		case "going":
			mv.DateSummary = memberDateSummary(m, detail.Event)
			going = append(going, mv)
		case "pending":
			mv.InvitedAgo = invitedAgo(m.InvitedAt)
			pending = append(pending, mv)
		case "declined":
			declined = append(declined, mv)
		}
	}

	const maxHeaderAvatars = 4
	headerAvatars := going
	headerExtra := 0
	if len(going) > maxHeaderAvatars {
		headerAvatars = going[:maxHeaderAvatars]
		headerExtra = len(going) - maxHeaderAvatars
	}

	currentItineraryJSON := buildItineraryJSON(currentMember)

	candidates, err := s.store.GetInviteCandidates(r.Context(), id)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	activities, err := s.store.GetActivities(r.Context(), id, user.ID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	var confirmed, ideas []store.Activity
	for _, a := range activities {
		if a.Status == "confirmed" {
			confirmed = append(confirmed, a)
		} else {
			ideas = append(ideas, a)
		}
	}

	mealPlan, err := s.store.GetMealPlanData(r.Context(), id)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Group meals by day
	var mealDays []mealDayView
	dayIndex := make(map[string]int)
	for _, m := range mealPlan.Meals {
		key := m.Date.Format("2006-01-02")
		idx, ok := dayIndex[key]
		if !ok {
			idx = len(mealDays)
			dayIndex[key] = idx
			mealDays = append(mealDays, mealDayView{
				Label: m.Date.Format("Monday · Jan 2"),
			})
		}
		mealDays[idx].Meals = append(mealDays[idx].Meals, m)
	}

	// Split groceries by category
	var toBuy, toBring []store.GroceryItem
	for _, g := range mealPlan.Groceries {
		if g.Category == "buy" {
			toBuy = append(toBuy, g)
		} else {
			toBring = append(toBring, g)
		}
	}

	// Find current user's food restriction
	var currentRestriction string
	hasRestriction := false
	for _, rr := range mealPlan.Restrictions {
		if rr.UserID == user.ID {
			currentRestriction = rr.Restriction
			hasRestriction = true
			break
		}
	}

	props := eventDetailProps{
		baseProps:            newBaseProps(r),
		Detail:               detail,
		DateRange:            formatDateRange(detail.StartDate, detail.EndDate),
		DayCount:             duration,
		TimelineCols:         timelineCols,
		TimelineRows:         timelineRows,
		HeaderAvatars:        headerAvatars,
		HeaderExtra:          headerExtra,
		Going:                going,
		Pending:              pending,
		Declined:             declined,
		GoingCount:           len(going),
		PendingCount:         len(pending),
		DeclinedCount:        len(declined),
		CurrentStatus:        currentStatus,
		InvitedBy:            invitedBy,
		InviteCandidates:     candidates,
		Confirmed:            confirmed,
		Ideas:                ideas,
		HasCurrentItinerary:  currentMember.HasItinerary,
		CurrentItineraryJSON: currentItineraryJSON,
		GoingMembersJSON:     buildGoingMembersJSON(going),
		Restrictions:         mealPlan.Restrictions,
		MealDays:             mealDays,
		ToBuy:                toBuy,
		ToBring:              toBring,
		CurrentRestriction:   currentRestriction,
		HasRestriction:       hasRestriction,
	}

	renderPage(w, r, http.StatusOK, "event_detail.html", props)
}

func (s *Server) eventRSVPPost(w http.ResponseWriter, r *http.Request) {
	id, ok := parseEventID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request.", http.StatusBadRequest)
		return
	}
	if body.Status != "going" && body.Status != "declined" {
		http.Error(w, "Invalid status.", http.StatusUnprocessableEntity)
		return
	}

	user, _ := middleware.UserFromContext(r.Context())
	if err := s.store.UpdateMemberStatus(r.Context(), id, user.ID, body.Status); err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) itineraryUpsertPut(w http.ResponseWriter, r *http.Request) {
	id, ok := parseEventID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}

	var input store.ItineraryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request.", http.StatusBadRequest)
		return
	}

	user, _ := middleware.UserFromContext(r.Context())
	if err := s.store.UpsertItinerary(r.Context(), id, user.ID, input); err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseEventID(r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func memberDateSummary(m store.EventDetailMember, e store.Event) string {
	if m.ArrivalDate == nil && m.DepartureDate == nil {
		return ""
	}
	arrival := e.StartDate
	departure := e.EndDate
	if m.ArrivalDate != nil {
		arrival = *m.ArrivalDate
	}
	if m.DepartureDate != nil {
		departure = *m.DepartureDate
	}
	return formatDateRange(arrival, departure)
}

func invitedAgo(t time.Time) string {
	days := int(time.Since(t).Hours() / 24)
	switch {
	case days == 0:
		return "Invited today"
	case days == 1:
		return "Invited yesterday"
	default:
		return fmt.Sprintf("Invited %d days ago", days)
	}
}

func buildGoingMembersJSON(going []memberView) template.JS {
	type member struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		AvatarColor string `json:"avatarColor"`
	}
	members := make([]member, len(going))
	for i, m := range going {
		members[i] = member{ID: m.UserID, Name: m.Name, AvatarColor: m.AvatarColor}
	}
	b, _ := json.Marshal(members)
	return template.JS(b)
}

// buildItineraryJSON serialises the current user's itinerary for modal pre-population.
func buildItineraryJSON(m store.EventDetailMember) template.JS {
	type itinData struct {
		HasItinerary          bool   `json:"hasItinerary"`
		ArrivalMode           string `json:"arrival_mode"`
		ArrivalDate           string `json:"arrival_date"`
		ArrivalTime           string `json:"arrival_time"`
		ArrivalFlightNumber   string `json:"arrival_flight_number"`
		ArrivalAirline        string `json:"arrival_airline"`
		ArrivalOrigin         string `json:"arrival_origin"`
		ArrivalDestination    string `json:"arrival_destination"`
		ArrivalDetails        string `json:"arrival_details"`
		DepartureMode         string `json:"departure_mode"`
		DepartureDate         string `json:"departure_date"`
		DepartureTime         string `json:"departure_time"`
		DepartureFlightNumber string `json:"departure_flight_number"`
		DepartureAirline      string `json:"departure_airline"`
		DepartureOrigin       string `json:"departure_origin"`
		DepartureDestination  string `json:"departure_destination"`
		DepartureDetails      string `json:"departure_details"`
	}

	timeToInput := func(t *time.Time) string {
		if t == nil {
			return ""
		}
		return t.Format("2006-01-02")
	}
	// Strip seconds from "HH:MM:SS" → "HH:MM" for <input type="time">
	trimSecs := func(s string) string {
		if len(s) >= 5 {
			return s[:5]
		}
		return s
	}

	d := itinData{
		HasItinerary:          m.HasItinerary,
		ArrivalMode:           m.ArrivalMode,
		ArrivalDate:           timeToInput(m.ArrivalDate),
		ArrivalTime:           trimSecs(m.ArrivalTime),
		ArrivalFlightNumber:   m.ArrivalFlightNumber,
		ArrivalAirline:        m.ArrivalAirline,
		ArrivalOrigin:         m.ArrivalOrigin,
		ArrivalDestination:    m.ArrivalDestination,
		ArrivalDetails:        m.ArrivalDetails,
		DepartureMode:         m.DepartureMode,
		DepartureDate:         timeToInput(m.DepartureDate),
		DepartureTime:         trimSecs(m.DepartureTime),
		DepartureFlightNumber: m.DepartureFlightNumber,
		DepartureAirline:      m.DepartureAirline,
		DepartureOrigin:       m.DepartureOrigin,
		DepartureDestination:  m.DepartureDestination,
		DepartureDetails:      m.DepartureDetails,
	}

	b, _ := json.Marshal(d)
	return template.JS(b)
}
