package postgres

import (
	"context"
	"fmt"

	"github.com/stpotter16/gather/internal/store"
)

func (s Store) GetActivities(ctx context.Context, eventID, userID int) ([]store.Activity, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			a.id, a.name, COALESCE(a.description, ''),
			u.name, u.avatar_color,
			a.status,
			(SELECT COUNT(*) FROM activity_votes WHERE activity_id = a.id),
			EXISTS (SELECT 1 FROM activity_votes WHERE activity_id = a.id AND user_id = $2)
		FROM activities a
		JOIN users u ON u.id = a.suggested_by
		WHERE a.event_id = $1
		ORDER BY CASE a.status WHEN 'confirmed' THEN 0 ELSE 1 END, a.created_at
	`, eventID, userID)
	if err != nil {
		return nil, fmt.Errorf("querying activities: %w", err)
	}
	defer rows.Close()

	var activities []store.Activity
	for rows.Next() {
		var a store.Activity
		if err := rows.Scan(
			&a.ID, &a.Name, &a.Description,
			&a.SuggestedBy, &a.SuggestedByColor,
			&a.Status, &a.VoteCount, &a.UserVoted,
		); err != nil {
			return nil, fmt.Errorf("scanning activity: %w", err)
		}
		activities = append(activities, a)
	}
	return activities, rows.Err()
}

func (s Store) CreateActivity(ctx context.Context, eventID, userID int, name, description string) (int, error) {
	var id int
	err := s.pool.QueryRow(ctx,
		`INSERT INTO activities (event_id, suggested_by, name, description) VALUES ($1, $2, $3, NULLIF($4, '')) RETURNING id`,
		eventID, userID, name, description,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("creating activity: %w", err)
	}
	return id, nil
}

func (s Store) ToggleActivityVote(ctx context.Context, activityID, userID int) error {
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM activity_votes WHERE activity_id = $1 AND user_id = $2`,
		activityID, userID,
	)
	if err != nil {
		return fmt.Errorf("removing vote: %w", err)
	}
	if tag.RowsAffected() == 0 {
		_, err = s.pool.Exec(ctx,
			`INSERT INTO activity_votes (activity_id, user_id) VALUES ($1, $2)`,
			activityID, userID,
		)
		if err != nil {
			return fmt.Errorf("adding vote: %w", err)
		}
	}
	return nil
}
