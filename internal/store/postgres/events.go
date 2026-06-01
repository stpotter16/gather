package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/stpotter16/gather/internal/store"
)

func (s Store) CreateEvent(ctx context.Context, name, location, description string, startDate, endDate time.Time, createdBy int) (int, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var desc *string
	if description != "" {
		desc = &description
	}

	var id int
	err = tx.QueryRow(ctx,
		`INSERT INTO events (name, start_date, end_date, location, description, created_by) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		name, startDate, endDate, location, desc, createdBy,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("inserting event: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO event_members (event_id, user_id, status, invited_by, invited_at) VALUES ($1, $2, 'going', $2, NOW())`,
		id, createdBy,
	)
	if err != nil {
		return 0, fmt.Errorf("inserting event member: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("committing transaction: %w", err)
	}
	return id, nil
}

func (s Store) GetEventsForUser(ctx context.Context, userID int) ([]store.EventSummary, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			e.id,
			e.name,
			e.start_date,
			e.end_date,
			e.location,
			(SELECT COUNT(*) FROM event_members WHERE event_id = e.id) AS member_count,
			(SELECT COUNT(*) FROM event_members WHERE event_id = e.id AND status = 'going') AS going_count,
			(SELECT COUNT(*) FROM event_members WHERE event_id = e.id AND status = 'pending') AS pending_count
		FROM events e
		JOIN event_members em ON em.event_id = e.id AND em.user_id = $1
		ORDER BY e.start_date
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("querying events: %w", err)
	}
	defer rows.Close()

	var events []store.EventSummary
	for rows.Next() {
		var e store.EventSummary
		if err := rows.Scan(
			&e.ID, &e.Name, &e.StartDate, &e.EndDate, &e.Location,
			&e.MemberCount, &e.GoingCount, &e.PendingCount,
		); err != nil {
			return nil, fmt.Errorf("scanning event: %w", err)
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating events: %w", err)
	}

	if len(events) == 0 {
		return events, nil
	}

	// Fetch up to 4 going members per event for the avatar stack.
	memberRows, err := s.pool.Query(ctx, `
		SELECT em.event_id, u.name, u.avatar_color
		FROM event_members em
		JOIN users u ON u.id = em.user_id
		WHERE em.status = 'going'
		  AND em.event_id IN (
		      SELECT event_id FROM event_members WHERE user_id = $1
		  )
		ORDER BY em.event_id, em.invited_at
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("querying members: %w", err)
	}
	defer memberRows.Close()

	membersByEvent := make(map[int][]store.EventMember)
	for memberRows.Next() {
		var eventID int
		var m store.EventMember
		if err := memberRows.Scan(&eventID, &m.Name, &m.AvatarColor); err != nil {
			return nil, fmt.Errorf("scanning member: %w", err)
		}
		membersByEvent[eventID] = append(membersByEvent[eventID], m)
	}
	if err := memberRows.Err(); err != nil {
		return nil, fmt.Errorf("iterating members: %w", err)
	}

	for i := range events {
		events[i].Members = membersByEvent[events[i].ID]
	}

	return events, nil
}
