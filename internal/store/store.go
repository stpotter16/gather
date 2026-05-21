package store

import "context"

type User struct {
	ID          int
	Name        string
	Email       string
	AvatarColor string
	PasswordHash string
}

type Store interface {
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserByID(ctx context.Context, id int) (User, error)
}
