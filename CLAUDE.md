# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Stack

Go 1.22 · chi router · PostgreSQL 16 (pgx/pgxpool) · Go `html/template` · HTMX · PWA

## Common Commands

```bash
make db        # start Postgres in Docker (required before make dev)
make dev       # run server locally (go run ./cmd/server)
make run       # full stack via docker compose (app + db)
make build     # compile to ./bin/server
make seed      # create initial admin user (go run ./cmd/seed)
make test      # go test ./...
make tidy      # go mod tidy
```

Local server binds to `:8080`. Requires `DATABASE_URL` env var (or defaults to `postgres://workout:workout_secret@localhost:5432/workout_tracker?sslmode=disable`). Copy `.env.example` to `.env` and source it.

## Architecture

### Entry point

`cmd/server/main.go` wires all dependencies manually: config → DB pool → repositories → handlers → middleware → chi router. No DI framework.

### Layer structure

| Layer | Package | Role |
|---|---|---|
| Config | `internal/config` | Reads env vars with defaults |
| DB | `internal/db` | Opens `pgxpool.Pool`; no ORM |
| Models | `internal/model` | Plain Go structs — no DB tags |
| Repositories | `internal/repository` | Raw SQL via pgx; one file per domain |
| Handlers | `internal/handler` | HTTP handlers; call repos directly |
| Middleware | `internal/middleware` | `RequireAuth` and `RequireAdmin` chi middleware |
| Session | `internal/session` | Cookie + Postgres `sessions` table, 7-day TTL |
| Templates | `web/templates` | Go `html/template`; always parsed fresh per request |

### Auth & roles

- Two roles: `trainer` and `client` (checked via `user.Role`).
- Admin flag (`user.IsAdmin`) is independent of role; grants access to `/admin/*` routes.
- Session ID is stored in an `HttpOnly` cookie. On each request `RequireAuth` resolves session → user from DB.

### Templates

`renderTemplate` in `internal/handler/render.go` always loads `base.html` + the target page + all four workout partials (`exercise_block`, `exercise_row`, `set_row`, `active_set_row`). Custom template funcs are all defined in `tmplFuncs` in the same file (e.g. `formatTonnage`, `workoutDuration`, `wellbeingEmoji`, `dict`, `iterate`).

HTMX partials use `renderPartial`, which parses only the target partial file and executes the named template matching `filepath.Base(name)`.

### Migrations

SQL files in `migrations/` are numbered sequentially (`001_init.sql` … `005_active_session.sql`). Docker Compose mounts them into `/docker-entrypoint-initdb.d` so they run automatically on a fresh DB container. There is **no runtime migration runner** — apply new migrations manually or recreate the container.

### Deployment

Push to `master` → GitHub Actions (`.github/workflows/deploy.yml`) → SSH into server → `git pull && docker compose up --build -d`.

## Key Domain Concepts

- **Workout** — belongs to a user, optionally supervised by a trainer; contains ordered `WorkoutExercise` entries each with `Set` rows (weight, reps, RPE, rest).
- **Template** (`WorkoutTemplate`) — trainer-owned workout blueprint with typed exercises; can be applied to create a new workout for a client.
- **Active session** — a workout can be started (`started_at`) and finished (`ended_at`); the `/workouts/{id}/active` screen lets clients tick off sets in real time.
- **Trainer–client relationship** — many-to-many via `trainer_clients` table; trainers see only their assigned clients.
