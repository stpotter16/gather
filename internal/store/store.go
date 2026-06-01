package store

import (
	"context"
	"time"
)

type User struct {
	ID           int
	Name         string
	Email        string
	AvatarColor  string
	PasswordHash string
}

type EventMember struct {
	Name        string
	AvatarColor string
}

type EventSummary struct {
	ID           int
	Name         string
	StartDate    time.Time
	EndDate      time.Time
	Location     string
	MemberCount  int
	GoingCount   int
	PendingCount int
	Members      []EventMember // up to 4 going members, for avatar stack
}

type Store interface {
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserByID(ctx context.Context, id int) (User, error)
	GetEventsForUser(ctx context.Context, userID int) ([]EventSummary, error)
	CreateEvent(ctx context.Context, name, location, description string, startDate, endDate time.Time, createdBy int) (int, error)
}
