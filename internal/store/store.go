package store

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type User struct {
	ID          int
	Name        string
	Email       string
	AvatarColor string
}

type EventMember struct {
	Name        string
	AvatarColor string
}

type Event struct {
	ID          int
	Name        string
	StartDate   time.Time
	EndDate     time.Time
	Location    string
	Description string
	CreatedBy   int
}

type EventSummary struct {
	Event
	MemberCount  int
	GoingCount   int
	PendingCount int
	Members      []EventMember // all going members; handlers slice to display limit
}

type EventDetailMember struct {
	UserID        int
	Name          string
	AvatarColor   string
	Status        string
	InvitedAt     time.Time
	InvitedByName string
	ArrivalDate   *time.Time
	DepartureDate *time.Time
}

type Accommodation struct {
	ID      int
	Label   string
	URL     string
	AddedBy string
}

type EventDetail struct {
	Event
	CreatedByName  string
	Members        []EventDetailMember
	Accommodations []Accommodation
}

type Store interface {
	GetUserByEmail(ctx context.Context, email string) (User, string, error)
	GetUserByID(ctx context.Context, id int) (User, error)
	GetEventsForUser(ctx context.Context, userID int) ([]EventSummary, error)
	CreateEvent(ctx context.Context, name, location, description string, startDate, endDate time.Time, createdBy int) (int, error)
	GetEventDetail(ctx context.Context, eventID int) (EventDetail, error)
	UpdateMemberStatus(ctx context.Context, eventID, userID int, status string) error
}
