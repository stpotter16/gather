package postgres

import (
	"context"
	"fmt"
)

func (s Store) AddAccommodation(ctx context.Context, eventID, userID int, label, url string) (int, error) {
	var id int
	err := s.pool.QueryRow(ctx,
		`INSERT INTO accommodations (event_id, added_by, label, url) VALUES ($1, $2, $3, $4) RETURNING id`,
		eventID, userID, label, url,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("adding accommodation: %w", err)
	}
	return id, nil
}

func (s Store) DeleteAccommodation(ctx context.Context, accommodationID, eventID int) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM accommodations WHERE id = $1 AND event_id = $2`,
		accommodationID, eventID,
	)
	if err != nil {
		return fmt.Errorf("deleting accommodation: %w", err)
	}
	return nil
}
