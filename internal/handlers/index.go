package handlers

import (
	"net/http"
	"time"

	"github.com/stpotter16/gather/internal/store"
)

type eventView struct {
	store.EventSummary
	DateRange    string
	DaysLabel    string // empty for past events
	ExtraAvatars int    // going members beyond the 4 shown
}

type indexProps struct {
	baseProps
	Upcoming []eventView
	Past     []eventView
}

func (s *Server) indexGet(w http.ResponseWriter, r *http.Request) {
	base := newBaseProps(r)
	events, err := s.store.GetEventsForUser(r.Context(), base.User.ID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	props := indexProps{baseProps: base}
	today := time.Now().Truncate(24 * time.Hour)

	for _, e := range events {
		v := eventView{
			EventSummary: e,
			DateRange:    formatDateRange(e.StartDate, e.EndDate),
		}
		const maxAvatars = 4
		if len(v.Members) > maxAvatars {
			v.ExtraAvatars = len(v.Members) - maxAvatars
			v.Members = v.Members[:maxAvatars]
		}
		if !e.StartDate.Before(today) {
			v.DaysLabel = daysLabel(e.StartDate)
			props.Upcoming = append(props.Upcoming, v)
		} else {
			props.Past = append(props.Past, v)
		}
	}

	renderPage(w, r, http.StatusOK, "index.html", props)
}

