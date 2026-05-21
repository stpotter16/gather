package handlers

import (
	"fmt"
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

func formatDateRange(start, end time.Time) string {
	if start.Year() != end.Year() {
		return fmt.Sprintf("%s – %s", start.Format("Jan 2, 2006"), end.Format("Jan 2, 2006"))
	}
	if start.Month() != end.Month() {
		return fmt.Sprintf("%s – %s", start.Format("Jan 2"), end.Format("Jan 2, 2006"))
	}
	return fmt.Sprintf("%s – %d, %d", start.Format("Jan 2"), end.Day(), end.Year())
}

func daysLabel(t time.Time) string {
	days := int(time.Until(t).Hours() / 24)
	switch {
	case days <= 0:
		return "Today"
	case days == 1:
		return "Tomorrow"
	case days < 30:
		return fmt.Sprintf("In %d days", days)
	case days < 365:
		months := (days + 15) / 30
		if months == 1 {
			return "In 1 month"
		}
		return fmt.Sprintf("In %d months", months)
	default:
		years := (days + 182) / 365
		if years == 1 {
			return "In 1 year"
		}
		return fmt.Sprintf("In %d years", years)
	}
}
