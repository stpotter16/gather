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

## Managing users

No self-serve signup. Add users manually:

```bash
make user/add EMAIL=sam@example.com NAME="Sam Potter"
make user/add EMAIL=sam@example.com NAME="Sam Potter" COLOR="#38bdf8"
```

Password is prompted interactively. `COLOR` is an optional hex value — defaults to a random color from a built-in palette.

## Schema changes

Migrations live in `internal/store/postgres/migrations/` as numbered SQL files (`001_...`, `002_...`). The server does **not** auto-migrate — run migrations explicitly before deploying new code:

```bash
make db/migrate
```

## Deployment

Two environments on Fly.io, both deployed manually:

```bash
fly deploy --config fly.toml           # production  (app-cometogather)
fly deploy --config fly.staging.toml   # staging     (app-cometogather-staging)
```

Migrations run automatically as a release command before traffic switches — a failed migration aborts the deploy.

**First-time setup:**

```bash
fly apps create app-cometogather
fly apps create app-cometogather-staging

fly secrets set \
  DATABASE_URL=<neon-pooler-url> \
  DATABASE_DIRECT_URL=<neon-direct-url> \
  GATHER_HMAC_SECRET=<secret> \
  --app app-cometogather

fly secrets set \
  DATABASE_URL=<neon-pooler-url> \
  DATABASE_DIRECT_URL=<neon-direct-url> \
  GATHER_HMAC_SECRET=<secret> \
  --app app-cometogather-staging
```

Neon provides both URLs in the connection details panel. `DATABASE_URL` is the pooler (hostname contains `-pooler`) — used by the server. `DATABASE_DIRECT_URL` is the direct connection — used by migrations. Use a separate Neon branch for staging. Generate an HMAC secret with `make secrets/hmac`.
