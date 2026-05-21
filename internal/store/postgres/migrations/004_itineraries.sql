CREATE TABLE itineraries (
    id SERIAL PRIMARY KEY,
    event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id),

    arrival_mode TEXT NOT NULL CHECK (arrival_mode IN ('flying', 'driving', 'other')),
    arrival_date DATE,
    arrival_time TIME,
    arrival_flight_number TEXT,
    arrival_airline TEXT,
    arrival_origin TEXT,
    arrival_destination TEXT,
    arrival_details TEXT,

    departure_mode TEXT NOT NULL CHECK (departure_mode IN ('flying', 'driving', 'other')),
    departure_date DATE,
    departure_time TIME,
    departure_flight_number TEXT,
    departure_airline TEXT,
    departure_origin TEXT,
    departure_destination TEXT,
    departure_details TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (event_id, user_id)
);
