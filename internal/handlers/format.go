package handlers

import (
	"fmt"
	"time"

	"github.com/stpotter16/gather/internal/store"
)

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
