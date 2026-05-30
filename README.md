# Workout Tracker

Personal workout tracking web app with trainer–client roles. Mobile-first, Russian UI.

**Live:** [dymko.ru](https://dymko.ru)

## Stack

- **Backend:** Go 1.22, [chi](https://github.com/go-chi/chi) router
- **Database:** PostgreSQL 16 (pgx/pgxpool, no ORM)
- **Frontend:** Go `html/template`, [HTMX](https://htmx.org), PWA
- **Deployment:** Docker Compose, Caddy reverse proxy, GitHub Actions CI/CD

## Features

- **Workout logging** — exercises, sets (weight / reps / RPE), notes, media upload
- **Active session** — real-time set tracking during a workout
- **Trainer–client** — trainers manage clients, assign workouts, view progress
- **Exercise catalog** — shared reference list with muscle groups
- **Exercise progress** — per-client history: max weight, sets, volume per session
- **Workout templates** — trainer-owned blueprints applied to any client
- **Analytics** — tonnage trends, personal records, exercise frequency
- **"Start from last"** — pre-fill a new workout from a previous session
- **Gyms** — track which gym each workout was done at

## Local Development

```bash
# Prerequisites: Docker, Go 1.22+

cp .env.example .env        # configure DATABASE_URL if needed

make db                     # start Postgres in Docker
make dev                    # run server on :8080

make seed                   # create initial admin user
make test                   # run tests
```

## Project Structure

```
cmd/server/         entry point — wires dependencies, router
internal/
  config/           env vars
  db/               pgxpool connection
  model/            plain Go structs
  repository/       raw SQL queries (one file per domain)
  handler/          HTTP handlers
  middleware/       RequireAuth, RequireAdmin
  session/          cookie + Postgres session store
migrations/         numbered SQL files (001–008)
web/
  templates/        html/template pages and partials
  static/           PWA manifest, service worker, icons
scripts/            import tools (Obsidian / iCloud journal → DB)
```

## Roles

| Role | Access |
|---|---|
| `client` | own workouts, active session |
| `trainer` | client management, templates, exercise catalog, client progress |
| `admin` (flag) | user management, trainer–client assignments |

## Deployment

Push to `master` → GitHub Actions → SSH into VPS → `git pull && docker compose up --build -d`.

Migrations are applied manually via psql pipe on first deploy of a new migration file.
