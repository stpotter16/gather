package store

import (
	"context"
	"time"
)

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

type Store interface {
	GetUserByEmail(ctx context.Context, email string) (User, string, error)
	GetUserByID(ctx context.Context, id int) (User, error)
	GetEventsForUser(ctx context.Context, userID int) ([]EventSummary, error)
	CreateEvent(ctx context.Context, name, location, description string, startDate, endDate time.Time, createdBy int) (int, error)
}
