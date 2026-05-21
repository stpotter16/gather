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
make test/go        # run tests
make lint/go        # run Go linter
```

Run `make css/build` once before first serving — the compiled CSS is committed to the repo but must be generated initially.

## Project structure

```
cmd/server/main.go              entry point, run() pattern
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

## Design conventions

- **Accent:** amber (`bg-amber-500`, `text-amber-600`)
- **Neutral base:** stone scale (`bg-stone-100` page bg, `bg-white` cards)
- **Cards:** `bg-white rounded-xl border border-stone-200`
- **Section headers inside cards:** `text-xs font-semibold text-stone-400 uppercase tracking-wide`
- **Avatars:** colored `rounded-full` divs with a single capital letter initial
