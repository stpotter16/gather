CREATE TABLE food_restrictions (
    id SERIAL PRIMARY KEY,
    event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id),
    restriction TEXT NOT NULL,
    UNIQUE (event_id, user_id)
);
