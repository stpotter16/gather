package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/stpotter16/gather/internal/store"
)

func (s Store) GetMealPlanData(ctx context.Context, eventID int) (store.MealPlanData, error) {
	var data store.MealPlanData

	// Food restrictions
	rrRows, err := s.pool.Query(ctx, `
		SELECT u.id, u.name, u.avatar_color, fr.restriction
		FROM food_restrictions fr
		JOIN users u ON u.id = fr.user_id
		WHERE fr.event_id = $1
		ORDER BY u.name
	`, eventID)
	if err != nil {
		return data, fmt.Errorf("querying food restrictions: %w", err)
	}
	defer rrRows.Close()
	for rrRows.Next() {
		var r store.FoodRestriction
		if err := rrRows.Scan(&r.UserID, &r.Name, &r.AvatarColor, &r.Restriction); err != nil {
			return data, fmt.Errorf("scanning restriction: %w", err)
		}
		data.Restrictions = append(data.Restrictions, r)
	}
	if err := rrRows.Err(); err != nil {
		return data, fmt.Errorf("iterating restrictions: %w", err)
	}

	// Meals
	mealRows, err := s.pool.Query(ctx,
		`SELECT id, name, date FROM meals WHERE event_id = $1 ORDER BY date, created_at`,
		eventID,
	)
	if err != nil {
		return data, fmt.Errorf("querying meals: %w", err)
	}
	defer mealRows.Close()
	mealIndex := make(map[int]int)
	for mealRows.Next() {
		var m store.Meal
		if err := mealRows.Scan(&m.ID, &m.Name, &m.Date); err != nil {
			return data, fmt.Errorf("scanning meal: %w", err)
		}
		mealIndex[m.ID] = len(data.Meals)
		data.Meals = append(data.Meals, m)
	}
	if err := mealRows.Err(); err != nil {
		return data, fmt.Errorf("iterating meals: %w", err)
	}

	if len(data.Meals) > 0 {
		// Meal cooks
		cookRows, err := s.pool.Query(ctx, `
			SELECT ma.meal_id, u.id, u.name, u.avatar_color
			FROM meal_assignments ma
			JOIN users u ON u.id = ma.user_id
			WHERE ma.meal_id IN (SELECT id FROM meals WHERE event_id = $1)
			ORDER BY ma.meal_id
		`, eventID)
		if err != nil {
			return data, fmt.Errorf("querying meal cooks: %w", err)
		}
		defer cookRows.Close()
		for cookRows.Next() {
			var mealID int
			var c store.MealCook
			if err := cookRows.Scan(&mealID, &c.UserID, &c.Name, &c.AvatarColor); err != nil {
				return data, fmt.Errorf("scanning cook: %w", err)
			}
			if idx, ok := mealIndex[mealID]; ok {
				data.Meals[idx].Cooks = append(data.Meals[idx].Cooks, c)
			}
		}
		if err := cookRows.Err(); err != nil {
			return data, fmt.Errorf("iterating cooks: %w", err)
		}

		// Dishes
		dishRows, err := s.pool.Query(ctx, `
			SELECT id, meal_id, name
			FROM dishes
			WHERE meal_id IN (SELECT id FROM meals WHERE event_id = $1)
			ORDER BY meal_id, id
		`, eventID)
		if err != nil {
			return data, fmt.Errorf("querying dishes: %w", err)
		}
		defer dishRows.Close()
		for dishRows.Next() {
			var mealID int
			var d store.Dish
			if err := dishRows.Scan(&d.ID, &mealID, &d.Name); err != nil {
				return data, fmt.Errorf("scanning dish: %w", err)
			}
			if idx, ok := mealIndex[mealID]; ok {
				data.Meals[idx].Dishes = append(data.Meals[idx].Dishes, d)
			}
		}
		if err := dishRows.Err(); err != nil {
			return data, fmt.Errorf("iterating dishes: %w", err)
		}
	}

	// Groceries
	gRows, err := s.pool.Query(ctx, `
		SELECT g.id, g.name, g.category,
		       COALESCE(u.name, ''), COALESCE(u.avatar_color, ''),
		       g.is_checked
		FROM groceries g
		LEFT JOIN users u ON u.id = g.assigned_to
		WHERE g.event_id = $1
		ORDER BY g.category, g.created_at
	`, eventID)
	if err != nil {
		return data, fmt.Errorf("querying groceries: %w", err)
	}
	defer gRows.Close()
	for gRows.Next() {
		var g store.GroceryItem
		if err := gRows.Scan(&g.ID, &g.Name, &g.Category, &g.AssignedName, &g.AssignedColor, &g.IsChecked); err != nil {
			return data, fmt.Errorf("scanning grocery: %w", err)
		}
		data.Groceries = append(data.Groceries, g)
	}
	if err := gRows.Err(); err != nil {
		return data, fmt.Errorf("iterating groceries: %w", err)
	}

	return data, nil
}

func (s Store) UpsertFoodRestriction(ctx context.Context, eventID, userID int, restriction string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO food_restrictions (event_id, user_id, restriction)
		VALUES ($1, $2, $3)
		ON CONFLICT (event_id, user_id) DO UPDATE SET restriction = EXCLUDED.restriction
	`, eventID, userID, restriction)
	if err != nil {
		return fmt.Errorf("upserting food restriction: %w", err)
	}
	return nil
}

func (s Store) CreateMeal(ctx context.Context, eventID int, name string, date time.Time) (int, error) {
	var id int
	err := s.pool.QueryRow(ctx,
		`INSERT INTO meals (event_id, name, date) VALUES ($1, $2, $3) RETURNING id`,
		eventID, name, date,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("creating meal: %w", err)
	}
	return id, nil
}

func (s Store) AddDish(ctx context.Context, mealID int, name string) (int, error) {
	var id int
	err := s.pool.QueryRow(ctx,
		`INSERT INTO dishes (meal_id, name) VALUES ($1, $2) RETURNING id`,
		mealID, name,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("adding dish: %w", err)
	}
	return id, nil
}

func (s Store) AddGrocery(ctx context.Context, eventID int, name, category string) (int, error) {
	var id int
	err := s.pool.QueryRow(ctx,
		`INSERT INTO groceries (event_id, name, category) VALUES ($1, $2, $3) RETURNING id`,
		eventID, name, category,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("adding grocery: %w", err)
	}
	return id, nil
}

func (s Store) ToggleGrocery(ctx context.Context, groceryID, eventID int) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE groceries SET is_checked = NOT is_checked WHERE id = $1 AND event_id = $2`,
		groceryID, eventID,
	)
	if err != nil {
		return fmt.Errorf("toggling grocery: %w", err)
	}
	return nil
}

func (s Store) DeleteMeal(ctx context.Context, mealID, eventID int) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM meals WHERE id = $1 AND event_id = $2`,
		mealID, eventID,
	)
	if err != nil {
		return fmt.Errorf("deleting meal: %w", err)
	}
	return nil
}

func (s Store) DeleteDish(ctx context.Context, dishID, mealID int) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM dishes WHERE id = $1 AND meal_id = $2`,
		dishID, mealID,
	)
	if err != nil {
		return fmt.Errorf("deleting dish: %w", err)
	}
	return nil
}

func (s Store) DeleteGrocery(ctx context.Context, groceryID, eventID int) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM groceries WHERE id = $1 AND event_id = $2`,
		groceryID, eventID,
	)
	if err != nil {
		return fmt.Errorf("deleting grocery: %w", err)
	}
	return nil
}

func (s Store) AssignCook(ctx context.Context, mealID, userID int) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO meal_assignments (meal_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		mealID, userID,
	)
	if err != nil {
		return fmt.Errorf("assigning cook: %w", err)
	}
	return nil
}

func (s Store) RemoveCook(ctx context.Context, mealID, userID int) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM meal_assignments WHERE meal_id = $1 AND user_id = $2`,
		mealID, userID,
	)
	if err != nil {
		return fmt.Errorf("removing cook: %w", err)
	}
	return nil
}
