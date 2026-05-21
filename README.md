# Gather

Family gathering coordination — itineraries, meal planning, activities.

## Local setup

```bash
make shell       # enter nix dev environment
make css/build   # compile Tailwind (first time only)
cp .env.example .env
# edit .env — set DATABASE_URL and GATHER_HMAC_SECRET
make db/migrate  # apply schema migrations
make server/live # start server with live reload
```

Generate an HMAC secret:

```bash
make secrets/hmac
```

## Schema changes

Migrations live in `internal/store/postgres/migrations/` as numbered SQL files (`001_...`, `002_...`). The server does **not** auto-migrate — run migrations explicitly before deploying new code:

```bash
make db/migrate
```

## Deployment

```bash
make server/deploy  # deploys to Fly.io
```

Set `DATABASE_URL` in Fly secrets to Neon's **pooler** URL (hostname contains `-pooler`).
