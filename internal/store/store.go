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
	// Itinerary — zero/nil values mean not entered
	HasItinerary          bool
	ArrivalMode           string
	ArrivalDate           *time.Time
	ArrivalTime           string // "HH:MM:SS" from Postgres, formatted at render time
	ArrivalFlightNumber   string
	ArrivalAirline        string
	ArrivalOrigin         string
	ArrivalDestination    string
	ArrivalDetails        string
	DepartureMode         string
	DepartureDate         *time.Time
	DepartureTime         string
	DepartureFlightNumber string
	DepartureAirline      string
	DepartureOrigin       string
	DepartureDestination  string
	DepartureDetails      string
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

// ItineraryInput carries raw string values for a user's arrival + departure.
// Empty string means "not set"; the store layer converts to NULL.
// Date format: YYYY-MM-DD. Time format: HH:MM.
type ItineraryInput struct {
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

type FoodRestriction struct {
	UserID      int
	Name        string
	AvatarColor string
	Restriction string
}

type MealCook struct {
	Name        string
	AvatarColor string
}

type Dish struct {
	ID   int
	Name string
}

type Meal struct {
	ID     int
	Name   string
	Date   time.Time
	Cooks  []MealCook
	Dishes []Dish
}

type GroceryItem struct {
	ID            int
	Name          string
	Category      string
	AssignedName  string
	AssignedColor string
	IsChecked     bool
}

type MealPlanData struct {
	Restrictions []FoodRestriction
	Meals        []Meal
	Groceries    []GroceryItem
}

type InviteCandidate struct {
	ID          int
	Name        string
	AvatarColor string
}

type Activity struct {
	ID               int
	Name             string
	Description      string
	SuggestedBy      string
	SuggestedByColor string
	Status           string // 'idea' or 'confirmed'
	VoteCount        int
	UserVoted        bool
}

type Store interface {
	GetUserByEmail(ctx context.Context, email string) (User, string, error)
	GetUserByID(ctx context.Context, id int) (User, error)
	GetEventsForUser(ctx context.Context, userID int) ([]EventSummary, error)
	CreateEvent(ctx context.Context, name, location, description string, startDate, endDate time.Time, createdBy int) (int, error)
	GetEventDetail(ctx context.Context, eventID int) (EventDetail, error)
	IsEventMember(ctx context.Context, eventID, userID int) (bool, error)
	UpdateMemberStatus(ctx context.Context, eventID, userID int, status string) error
	UpsertItinerary(ctx context.Context, eventID, userID int, input ItineraryInput) error
	GetMealPlanData(ctx context.Context, eventID int) (MealPlanData, error)
	UpsertFoodRestriction(ctx context.Context, eventID, userID int, restriction string) error
	CreateMeal(ctx context.Context, eventID int, name string, date time.Time) (int, error)
	AddDish(ctx context.Context, mealID int, name string) (int, error)
	AddGrocery(ctx context.Context, eventID int, name, category string) (int, error)
	ToggleGrocery(ctx context.Context, groceryID, eventID int) error
	GetActivities(ctx context.Context, eventID, userID int) ([]Activity, error)
	CreateActivity(ctx context.Context, eventID, userID int, name, description string) (int, error)
	ToggleActivityVote(ctx context.Context, activityID, userID int) error
	GetInviteCandidates(ctx context.Context, eventID int) ([]InviteCandidate, error)
	InviteUsers(ctx context.Context, eventID, inviterID int, userIDs []int) error
}
