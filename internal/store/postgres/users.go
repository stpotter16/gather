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

func (s Store) GetUserByEmail(ctx context.Context, email string) (store.User, string, error) {
	var u store.User
	var hash string
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, email, avatar_color, password_hash
		FROM users
		WHERE email = $1
	`, email).Scan(&u.ID, &u.Name, &u.Email, &u.AvatarColor, &hash)
	if err != nil {
		return store.User{}, "", fmt.Errorf("getting user by email: %w", err)
	}
	return u, hash, nil
}

func (s Store) GetInviteCandidates(ctx context.Context, eventID int) ([]store.InviteCandidate, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, avatar_color
		FROM users
		WHERE id NOT IN (SELECT user_id FROM event_members WHERE event_id = $1)
		ORDER BY name
	`, eventID)
	if err != nil {
		return nil, fmt.Errorf("querying invite candidates: %w", err)
	}
	defer rows.Close()

	var candidates []store.InviteCandidate
	for rows.Next() {
		var c store.InviteCandidate
		if err := rows.Scan(&c.ID, &c.Name, &c.AvatarColor); err != nil {
			return nil, fmt.Errorf("scanning candidate: %w", err)
		}
		candidates = append(candidates, c)
	}
	return candidates, rows.Err()
}

func (s Store) InviteUsers(ctx context.Context, eventID, inviterID int, userIDs []int) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, userID := range userIDs {
		_, err := tx.Exec(ctx, `
			INSERT INTO event_members (event_id, user_id, status, invited_by, invited_at)
			VALUES ($1, $2, 'pending', $3, NOW())
			ON CONFLICT (event_id, user_id) DO NOTHING
		`, eventID, userID, inviterID)
		if err != nil {
			return fmt.Errorf("inviting user %d: %w", userID, err)
		}
	}

	return tx.Commit(ctx)
}

func (s Store) GetUserByID(ctx context.Context, id int) (store.User, error) {
	var u store.User
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, email, avatar_color
		FROM users
		WHERE id = $1
	`, id).Scan(&u.ID, &u.Name, &u.Email, &u.AvatarColor)
	if err != nil {
		return store.User{}, fmt.Errorf("getting user by ID: %w", err)
	}
	return u, nil
}
