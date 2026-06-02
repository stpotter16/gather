package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
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

	// Fetch all going members per event; handlers slice to the display limit.
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

func (s Store) GetEventDetail(ctx context.Context, eventID int) (store.EventDetail, error) {
	var d store.EventDetail
	err := s.pool.QueryRow(ctx, `
		SELECT e.id, e.name, e.start_date, e.end_date, e.location,
		       COALESCE(e.description, ''), e.created_by, u.name
		FROM events e
		JOIN users u ON u.id = e.created_by
		WHERE e.id = $1
	`, eventID).Scan(
		&d.ID, &d.Name, &d.StartDate, &d.EndDate, &d.Location,
		&d.Description, &d.CreatedBy, &d.CreatedByName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.EventDetail{}, store.ErrNotFound
		}
		return store.EventDetail{}, fmt.Errorf("getting event: %w", err)
	}

	memberRows, err := s.pool.Query(ctx, `
		SELECT
			u.id, u.name, u.avatar_color, em.status, em.invited_at, ib.name,
			i.event_id IS NOT NULL,
			COALESCE(i.arrival_mode, ''),
			i.arrival_date,
			COALESCE(i.arrival_time::TEXT, ''),
			COALESCE(i.arrival_flight_number, ''),
			COALESCE(i.arrival_airline, ''),
			COALESCE(i.arrival_origin, ''),
			COALESCE(i.arrival_destination, ''),
			COALESCE(i.arrival_details, ''),
			COALESCE(i.departure_mode, ''),
			i.departure_date,
			COALESCE(i.departure_time::TEXT, ''),
			COALESCE(i.departure_flight_number, ''),
			COALESCE(i.departure_airline, ''),
			COALESCE(i.departure_origin, ''),
			COALESCE(i.departure_destination, ''),
			COALESCE(i.departure_details, '')
		FROM event_members em
		JOIN users u  ON u.id  = em.user_id
		JOIN users ib ON ib.id = em.invited_by
		LEFT JOIN itineraries i ON i.event_id = em.event_id AND i.user_id = em.user_id
		WHERE em.event_id = $1
		ORDER BY CASE em.status WHEN 'going' THEN 1 WHEN 'pending' THEN 2 ELSE 3 END, em.invited_at
	`, eventID)
	if err != nil {
		return store.EventDetail{}, fmt.Errorf("querying members: %w", err)
	}
	defer memberRows.Close()

	for memberRows.Next() {
		var m store.EventDetailMember
		if err := memberRows.Scan(
			&m.UserID, &m.Name, &m.AvatarColor, &m.Status, &m.InvitedAt, &m.InvitedByName,
			&m.HasItinerary,
			&m.ArrivalMode, &m.ArrivalDate, &m.ArrivalTime,
			&m.ArrivalFlightNumber, &m.ArrivalAirline, &m.ArrivalOrigin, &m.ArrivalDestination, &m.ArrivalDetails,
			&m.DepartureMode, &m.DepartureDate, &m.DepartureTime,
			&m.DepartureFlightNumber, &m.DepartureAirline, &m.DepartureOrigin, &m.DepartureDestination, &m.DepartureDetails,
		); err != nil {
			return store.EventDetail{}, fmt.Errorf("scanning member: %w", err)
		}
		d.Members = append(d.Members, m)
	}
	if err := memberRows.Err(); err != nil {
		return store.EventDetail{}, fmt.Errorf("iterating members: %w", err)
	}

	accomRows, err := s.pool.Query(ctx, `
		SELECT a.id, a.label, a.url, u.name
		FROM accommodations a
		JOIN users u ON u.id = a.added_by
		WHERE a.event_id = $1
		ORDER BY a.created_at
	`, eventID)
	if err != nil {
		return store.EventDetail{}, fmt.Errorf("querying accommodations: %w", err)
	}
	defer accomRows.Close()

	for accomRows.Next() {
		var a store.Accommodation
		if err := accomRows.Scan(&a.ID, &a.Label, &a.URL, &a.AddedBy); err != nil {
			return store.EventDetail{}, fmt.Errorf("scanning accommodation: %w", err)
		}
		d.Accommodations = append(d.Accommodations, a)
	}
	if err := accomRows.Err(); err != nil {
		return store.EventDetail{}, fmt.Errorf("iterating accommodations: %w", err)
	}

	return d, nil
}

func (s Store) IsEventMember(ctx context.Context, eventID, userID int) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM event_members WHERE event_id = $1 AND user_id = $2)`,
		eventID, userID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking event membership: %w", err)
	}
	return exists, nil
}

func (s Store) UpdateMemberStatus(ctx context.Context, eventID, userID int, status string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE event_members SET status = $1, responded_at = NOW() WHERE event_id = $2 AND user_id = $3`,
		status, eventID, userID,
	)
	if err != nil {
		return fmt.Errorf("updating member status: %w", err)
	}
	return nil
}

func (s Store) UpsertItinerary(ctx context.Context, eventID, userID int, in store.ItineraryInput) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO itineraries (
			event_id, user_id,
			arrival_mode,   arrival_date,            arrival_time,
			arrival_flight_number,  arrival_airline,
			arrival_origin,         arrival_destination,  arrival_details,
			departure_mode, departure_date,          departure_time,
			departure_flight_number, departure_airline,
			departure_origin,        departure_destination, departure_details
		) VALUES (
			$1, $2,
			NULLIF($3,''),  NULLIF($4,'')::DATE,     NULLIF($5,'')::TIME,
			NULLIF($6,''),  NULLIF($7,''),
			NULLIF($8,''),  NULLIF($9,''),            NULLIF($10,''),
			NULLIF($11,''), NULLIF($12,'')::DATE,     NULLIF($13,'')::TIME,
			NULLIF($14,''), NULLIF($15,''),
			NULLIF($16,''), NULLIF($17,''),           NULLIF($18,'')
		)
		ON CONFLICT (event_id, user_id) DO UPDATE SET
			arrival_mode            = EXCLUDED.arrival_mode,
			arrival_date            = EXCLUDED.arrival_date,
			arrival_time            = EXCLUDED.arrival_time,
			arrival_flight_number   = EXCLUDED.arrival_flight_number,
			arrival_airline         = EXCLUDED.arrival_airline,
			arrival_origin          = EXCLUDED.arrival_origin,
			arrival_destination     = EXCLUDED.arrival_destination,
			arrival_details         = EXCLUDED.arrival_details,
			departure_mode          = EXCLUDED.departure_mode,
			departure_date          = EXCLUDED.departure_date,
			departure_time          = EXCLUDED.departure_time,
			departure_flight_number = EXCLUDED.departure_flight_number,
			departure_airline       = EXCLUDED.departure_airline,
			departure_origin        = EXCLUDED.departure_origin,
			departure_destination   = EXCLUDED.departure_destination,
			departure_details       = EXCLUDED.departure_details
	`,
		eventID, userID,
		in.ArrivalMode, in.ArrivalDate, in.ArrivalTime,
		in.ArrivalFlightNumber, in.ArrivalAirline,
		in.ArrivalOrigin, in.ArrivalDestination, in.ArrivalDetails,
		in.DepartureMode, in.DepartureDate, in.DepartureTime,
		in.DepartureFlightNumber, in.DepartureAirline,
		in.DepartureOrigin, in.DepartureDestination, in.DepartureDetails,
	)
	if err != nil {
		return fmt.Errorf("upserting itinerary: %w", err)
	}
	return nil
}
