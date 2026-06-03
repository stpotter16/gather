# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

**Gather** — a family gathering coordination website. Users create events, invite family members, and collaborate on logistics: arrival/departure itineraries, meal planning, and activity brainstorming.

## Stack

- **Language:** Go (net/http, html/template — no framework)
- **Database:** PostgreSQL via pgx/v5, raw SQL (no ORM, no codegen)
- **Auth:** HMAC-signed cookies (no DB sessions — avoids per-request Neon cold-start penalty)
- **CSS:** Tailwind v4, compiled to `internal/handlers/static/css/style.css` via `@tailwindcss/cli`
- **Hosting:** Fly.io + Neon (serverless Postgres)
- **Dev environment:** Nix flake (`nix develop` or `make shell`)
- **Live reload:** `modd` (`make server/live`)

## Build commands

```
make shell          # enter nix dev environment
make server/build   # compile to ./tmp/server
make server/run     # build + run
make server/live    # build + run with modd live reload
make css/build      # compile Tailwind CSS (run once before first serve)
make db/migrate     # apply pending migrations (run before deploying schema changes)
make test/go        # run tests
make lint/go        # run Go linter
```

Run `make css/build` once before first serving — the compiled CSS is committed to the repo but must be generated initially.

## Project structure

```
cmd/server/main.go              entry point, run() pattern
cmd/migrate/main.go             migration CLI — connect, apply pending migrations, exit
internal/handlers/              HTTP layer
  middleware/                   logging, CSP, auth
  static/css/style.css          compiled Tailwind output (committed)
  templates/layouts/            base.html, app.html
  templates/pages/              one file per page
  routes.go                     all route registration
  server.go                     NewServer(), wires middleware
  views.go                      template rendering helpers + embed
  static.go                     embedded static file serving
internal/sessions/              HMAC signed-cookie session management
internal/store/                 store interface
  postgres/                     pgx/v5 implementation
    migrations/                 numbered SQL files (001_..., 002_..., ...)
style/input.css                 Tailwind source (just @import "tailwindcss")
dev-scripts/                    shell scripts called by Makefile
```

## Implementation status

The core app is built and functional.

### Nice-to-have (app works without these)

**Location typeahead on create/edit event** — suggestions as the user types in the location field on the create/edit event forms. Low priority given the field is filled once per event. Options:
- **Mapbox Geocoding** (recommended) — generous free tier (100k req/month), simple REST API, public token can be scoped to the app's domain so no backend proxy is needed. A debounced `input` listener fetches suggestions and renders a small dropdown; no schema changes or new routes required.
- **Google Places Autocomplete** — most complete results but billing starts from day one.

**Flight number lookup** — the itinerary form already has fields for flight number, airline, origin, destination, and times. A "Look up" button would auto-fill them from a flight data API. Implementation: a new `GET /api/flights?number=AA123&date=2026-07-04` route that proxies to the API (keeps the key server-side), returning route + times for the frontend to populate. Options:
- **AviationStack** (already referenced in codebase) — free tier is 100 req/month but covers real-time flights only, not future/scheduled. Accessing scheduled future flights requires the $9.99/month plan.
- **AeroDataBox** (via RapidAPI) — 100 free calls/day, covers scheduled flights.

**Email notifications** — would unlock invite notifications (send when someone is added to an event) and the nudge button (reminder to pending invitees). Push notifications are out of scope — they require a service worker/PWA setup for marginal gain. Email only. Options:
- **Resend** (recommended) — 3k emails/month free, clean REST API, good Go support. Add `RESEND_API_KEY` secret to both Fly apps, add a thin `internal/email/` package. Plain text emails are fine for a family app.
- **Postmark** — excellent deliverability, 100 free/month (likely sufficient).
- **SendGrid** — 100/day free tier.

## Code quality backlog

**`event_detail.html` is ~1400 lines** — Go templates support `{{ template "name" . }}` composition. Splitting into partials (`_meal_plan.html`, `_activities.html`, `_grocery_panel.html`, `_modals.html`, etc.) would make the file navigable. `views.go`'s `renderPage` would need to accept additional template files.

**`style-src 'unsafe-inline'` in CSP** — Tailwind compiles to an external stylesheet, so inline styles aren't needed for that. The `style="background-color: ..."` element attributes are the only thing requiring it. Could be removed by moving dynamic colors to CSS custom properties or a data-attribute approach, but this is low priority.

## Deployment

Two Fly.io apps — `app-cometogather` (production) and `app-cometogather-staging` (staging), both in `iad`. Both are deployed manually.

**Config files:**
- `fly.toml` — production
- `fly.staging.toml` — staging

Both scale to zero (`min_machines_running = 0`).

**Deploy commands:**
```bash
fly deploy --config fly.toml           # production
fly deploy --config fly.staging.toml   # staging
```

The release command in both configs runs `/app/migrate` before traffic switches to the new version. Failed migrations abort the deploy.

**One-time setup per environment:**
```bash
fly apps create app-cometogather
fly apps create app-cometogather-staging

fly secrets set \
  DATABASE_URL=<neon-prod-pooler-url> \
  DATABASE_DIRECT_URL=<neon-prod-direct-url> \
  GATHER_HMAC_SECRET=<secret> \
  --app app-cometogather

fly secrets set \
  DATABASE_URL=<neon-staging-pooler-url> \
  DATABASE_DIRECT_URL=<neon-staging-direct-url> \
  GATHER_HMAC_SECRET=<secret> \
  --app app-cometogather-staging
```

**Neon connection strings:**
- `DATABASE_URL` — the **pooler** URL (hostname contains `-pooler`). Used by the server; routes through PgBouncer in transaction mode for efficient connection handling.
- `DATABASE_DIRECT_URL` — the **direct** URL (no `-pooler`). Used only by `cmd/migrate`; DDL requires a direct Postgres connection, not PgBouncer.

Use a `staging` branch off `main` for the staging database. Each environment has its own secrets so sessions don't bleed between environments.

## Auth and user management

No self-serve signup. Users are added to the database manually by the admin via SQL. Inviting is restricted to existing users.

Sessions use HMAC-signed cookies (not DB-backed). Cookie payload: `userID|expires`, signed with `GATHER_HMAC_SECRET`. On logout the cookie is cleared client-side. To force-logout all users, rotate the secret.

## Feature notes

**Meal Plan tab** (formerly "Food"):
- Global food restrictions card at the top — one entry per person, free text
- Meals organised by day; each meal has a cook assignment (one or more attendees) and a dish list
- Responsibility is at the meal level, not the dish level
- Grocery panel slides in from the right; split into "To buy" and "To bring" (with per-item attendee assignment)

**Accommodations** (Overview tab):
- Flexible — can be one shared rental or many per-person links
- Stored as labelled URLs; anyone in the event can add one
- No API integration (Airbnb and VRBO have no public API)

**Itineraries:**
- Flight number lookup via a flight data API (e.g. AviationStack) — route and times auto-fill
- Also supports Driving and Other modes

## Screens (HTML mockups)

Live in `screens/` — reference for design and UI decisions, not production code.

- `screens/gatherings.html` — home screen
- `screens/create-gathering.html` — create event form
- `screens/event.html` — full event detail page (Overview, Itineraries, Meal Plan, Activities tabs)
- `screens/invite.html` — invite modal
- `screens/add-itinerary.html` — add itinerary modal

## Database schema

Migrations in `internal/store/postgres/migrations/`, applied in order at startup.

### users (001)
| column | type | notes |
|---|---|---|
| id | SERIAL PK | |
| name | TEXT | display name |
| email | TEXT UNIQUE | login identifier |
| avatar_color | TEXT | hex color for initial avatar |
| password_hash | TEXT | bcrypt |
| created_at | TIMESTAMPTZ | |

No self-serve signup — rows inserted manually by admin.

### events (002)
| column | type | notes |
|---|---|---|
| id | SERIAL PK | |
| name | TEXT | |
| start_date | DATE | date only, not TIMESTAMPTZ |
| end_date | DATE | |
| location | TEXT | |
| description | TEXT | nullable |
| created_by | INTEGER FK → users | |
| created_at | TIMESTAMPTZ | |

### event_members (003)
| column | type | notes |
|---|---|---|
| id | SERIAL PK | |
| event_id | INTEGER FK → events | CASCADE delete |
| user_id | INTEGER FK → users | |
| status | TEXT | CHECK: 'pending', 'going', 'declined'; default 'pending' |
| invited_by | INTEGER FK → users | |
| invited_at | TIMESTAMPTZ | |
| responded_at | TIMESTAMPTZ | nullable |

UNIQUE(event_id, user_id).

### itineraries (004)
One row per person per event. Flat columns for both arrival and departure.

| prefix | columns |
|---|---|
| arrival_ | mode CHECK('flying','driving','other'), date DATE, time TIME, flight_number, airline, origin, destination, details |
| departure_ | same set of columns |

UNIQUE(event_id, user_id). All columns nullable — a person may have only one direction filled in.

### accommodations (005)
| column | type | notes |
|---|---|---|
| id | SERIAL PK | |
| event_id | INTEGER FK → events | CASCADE delete |
| label | TEXT | human-readable name |
| url | TEXT | link to Airbnb/VRBO/etc. |
| added_by | INTEGER FK → users | |
| created_at | TIMESTAMPTZ | |

### meals (006)
Three tables: `meals`, `meal_assignments`, `dishes`.

**meals**: id, event_id FK, name, date DATE, created_at  
**meal_assignments**: (meal_id, user_id) composite PK — who is cooking  
**dishes**: id, meal_id FK, name, notes — no per-dish owner; responsibility is at the meal level

### food_restrictions (007)
| column | type | notes |
|---|---|---|
| id | SERIAL PK | |
| event_id | INTEGER FK → events | CASCADE delete |
| user_id | INTEGER FK → users | |
| restriction | TEXT | free text |

UNIQUE(event_id, user_id) — one restriction entry per person per event. Scoped per-event rather than globally per-user for flexibility.

### groceries (008)
| column | type | notes |
|---|---|---|
| id | SERIAL PK | |
| event_id | INTEGER FK → events | CASCADE delete |
| name | TEXT | |
| category | TEXT | CHECK: 'buy' or 'bring' |
| assigned_to | INTEGER FK → users | nullable — only relevant for 'bring' items |
| is_checked | BOOLEAN | default FALSE |
| created_at | TIMESTAMPTZ | |

### activities (009)
Two tables: `activities` and `activity_votes`.

**activities**: id, event_id FK, name, description (nullable), suggested_by FK → users, status CHECK('idea','confirmed') default 'idea', created_at  
**activity_votes**: (activity_id, user_id) composite PK. Vote count via `SELECT COUNT(*) FROM activity_votes WHERE activity_id = $1`.

---

## Security conventions

**Event membership:** All mutation handlers for event-scoped resources (meals, dishes, groceries, activities, itineraries, invites) must verify the current user is a member of the event before touching the database. Use `s.store.IsEventMember(r.Context(), eventID, user.ID)` at the top of the handler and return 403 if false. The `RequireAuth` middleware only checks login status — it does not verify event membership.

**Embedding server data for client JS:** Use `html/template.JS` to safely embed Go values into `<script>` blocks without HTML-escaping. Marshal the value to JSON, cast to `template.JS`, and reference it in the template: `var x = {{ .MyJSField }};`. Used for modal pre-population (e.g., itinerary data) where the server-rendered value needs to be readable by JS on the same page.

## Client–server communication

All state-mutating requests (POST/PUT/DELETE) send **JSON bodies** — not HTML form encoding.

**Frontend pattern:**
- Forms have no `method`/`action` attributes; a `<script nonce="{{ .CspNonce }}">` block intercepts submit
- `fetch` options: `method`, `credentials: "include"`, `headers: {"Content-Type": "application/json"}`, `body: JSON.stringify({...})`
- Errors (`!resp.ok`): `(await resp.text()).trim()` shown in a hidden error element (remove `hidden` class to display)
- Success: parse `await resp.json()` for the created ID if needed, then `window.location.href = "..."` to navigate

**Backend pattern:**
- Decode with `json.NewDecoder(r.Body).Decode(&body)` — never `r.ParseForm()` / `r.FormValue()`
- Validation failure: `http.Error(w, "Human-readable message.", http.StatusUnprocessableEntity)`
- Created resource: `w.Header().Set("Content-Type", "application/json"); w.WriteHeader(http.StatusCreated); fmt.Fprintf(w, '{"id":%d}', id)`
- No-content success: `w.WriteHeader(http.StatusNoContent)`
- No server-side redirects from mutation handlers — the client navigates

**CSRF:** `SameSite=Lax` on the session cookie is sufficient. JSON-only endpoints reject cross-origin form POSTs; no explicit CSRF tokens needed.

## Design conventions

- **Accent:** amber (`bg-amber-500`, `text-amber-600`)
- **Neutral base:** stone scale (`bg-stone-100` page bg, `bg-white` cards)
- **Cards:** `bg-white rounded-xl border border-stone-200`
- **Section headers inside cards:** `text-xs font-semibold text-stone-400 uppercase tracking-wide`
- **Avatars:** colored `rounded-full` divs with a single capital letter initial
