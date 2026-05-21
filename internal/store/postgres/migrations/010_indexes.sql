-- event_members is queried by both user_id (home page) and event_id (event detail)
CREATE INDEX ON event_members (user_id);
CREATE INDEX ON event_members (event_id);

-- all child tables are queried by event_id on the event detail page
CREATE INDEX ON itineraries (event_id);
CREATE INDEX ON accommodations (event_id);
CREATE INDEX ON meals (event_id);
CREATE INDEX ON dishes (meal_id);
CREATE INDEX ON food_restrictions (event_id);
CREATE INDEX ON groceries (event_id);
CREATE INDEX ON activities (event_id);
CREATE INDEX ON activity_votes (activity_id);
