package postgres

import (
	"context"
	"fmt"

	"github.com/stpotter16/gather/internal/store"
)

func (s Store) CreateUser(ctx context.Context, name, email, avatarColor, passwordHash string) (int, error) {
	var id int
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (name, email, avatar_color, password_hash) VALUES ($1, $2, $3, $4) RETURNING id`,
		name, email, avatarColor, passwordHash,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("inserting user: %w", err)
	}
	return id, nil
}

func (s Store) GetUserByEmail(ctx context.Context, email string) (store.User, error) {
	var u store.User
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, email, avatar_color, password_hash
		FROM users
		WHERE email = $1
	`, email).Scan(&u.ID, &u.Name, &u.Email, &u.AvatarColor, &u.PasswordHash)
	if err != nil {
		return store.User{}, fmt.Errorf("getting user by email: %w", err)
	}
	return u, nil
}

func (s Store) GetUserByID(ctx context.Context, id int) (store.User, error) {
	var u store.User
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, email, avatar_color, password_hash
		FROM users
		WHERE id = $1
	`, id).Scan(&u.ID, &u.Name, &u.Email, &u.AvatarColor, &u.PasswordHash)
	if err != nil {
		return store.User{}, fmt.Errorf("getting user by ID: %w", err)
	}
	return u, nil
}
