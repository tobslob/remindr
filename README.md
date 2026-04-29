# Remindr

Remindr is a Go + PostgreSQL backend for user-owned tasks, tags, and reminder records. It exposes a versioned JSON API, stores all user data in PostgreSQL, and runs a lightweight reminder scheduler inside the API process.

The API supports:

- user registration and login with JWT authentication
- CRUD-style task management
- tag management with per-user unique tag names
- many-to-many task/tag relationships
- reminder CRUD, persistence, lifecycle state transitions, and background delivery processing

Current reminder note:

- authenticated reminder CRUD endpoints are available
- the reminder scheduler/worker runtime starts with the API process
- due reminders are claimed in batches, sent by workers, and marked `sent` or retried
- the default sender logs due reminders; plug in another `reminder.Sender` for email, SMS, push, webhooks, or any other transport
- reminders are created explicitly through the reminder API; task creation does not automatically create reminders

## Tech Stack

- Go `1.26`
- PostgreSQL
- `chi` router
- JWT auth
- SQL migrations via `golang-migrate`

## Development Tools

- Docker Compose for local service orchestration
- Air for local live reload in the development container
- `make` targets for migration workflows
- Postman or any HTTP client for API testing

## Project Layout

- `cmd/api` HTTP server, handlers, middleware, route mounting
- `cmd/utils` request parsing, validation helpers, error responses, pagination
- `cmd/tokens` JWT creation and verification
- `internal/store` persistence layer
- `internal/db` database connection setup and migrations
- `internal/reminder` reminder domain package, scheduler, workers, and sender interface

More detail:

- [docs/architecture.md](docs/architecture.md)
- [docs/api.md](docs/api.md)
- [docs/data-model.md](docs/data-model.md)

## Environment Variables

The server expects these environment variables:

- `ADDR`
- `DB_ADDR`
- `DB_MAX_OPEN_CONNS`
- `DB_MAX_IDLE_CONNS`
- `DB_MAX_IDLE_TIME`
- `TOKEN_SECRET_KEY`

Notes:

- `.env` is loaded automatically by `internal/env/env.go`
- `TOKEN_SECRET_KEY` must be at least 32 characters for JWT signing

## Local Development

Before starting the app, create a `.env` file with the environment variables above and point `DB_ADDR` at a reachable PostgreSQL database.

### Run with Air

The development container uses Air and exposes the API on `localhost:8081`.

```bash
docker compose up --build
```

The container maps port `8081` on the host to `8080` in the app.

### Run directly

```bash
go run ./cmd/api
```

Run tests:

```bash
go test ./...
```

## Migrations

Create a migration:

```bash
make migration add_some_change
```

Apply migrations:

```bash
make migrate-up
```

Rollback migrations:

```bash
make migrate-down 1
```

The migration files live in `internal/db/migrations`.

## Auth Model

- `POST /v1/users/register` creates a user
- `POST /v1/users/login` returns a JWT
- authenticated routes require `Authorization: Bearer <token>`

The auth middleware:

- verifies the token
- loads the user from the database
- stores the user in request context

## API Summary

### Public

- `GET /v1/healthz`
- `POST /v1/users/register`
- `POST /v1/users/login`

### Authenticated tasks

- `POST /v1/tasks/`
- `GET /v1/tasks/`
- `GET /v1/tasks/{id}`
- `PATCH /v1/tasks/{id}`
- `DELETE /v1/tasks/{id}`
- `DELETE /v1/tasks/`
- `DELETE /v1/tasks/bulk`
- `POST /v1/tasks/{id}/tags/{tag_id}`
- `PUT /v1/tasks/{id}/tags`
- `GET /v1/tasks/tags`

### Authenticated tags

- `POST /v1/tags/`
- `GET /v1/tags/`
- `GET /v1/tags/{id}`
- `GET /v1/tags/{id}/tasks`
- `PATCH /v1/tags/{id}`
- `DELETE /v1/tags/{id}`
- `DELETE /v1/tags/{task_id}/{id}`

### Authenticated reminders

- `POST /v1/reminders/`
- `GET /v1/reminders/task/{task_id}`
- `GET /v1/reminders/{id}`
- `PATCH /v1/reminders/{id}`
- `DELETE /v1/reminders/{id}`

See [docs/api.md](docs/api.md) for request and query details.

## Reminder Runtime

When the API starts, `cmd/api/main.go` creates a reminder service with the default config:

- scheduler interval: `30s`
- claim batch size: `25`
- worker count: `2`
- queue size: same as the batch size

The scheduler immediately claims due reminders, then repeats on the configured interval. Claimed reminders move from `pending` to `processing`; workers load each `processing` reminder, call the configured sender, then mark it `sent` on success. Failed sends are returned to `pending` until the third attempt, when the reminder is marked `failed`.

## Methodology

- API-first resource design with JWT-protected user-owned data
- user-scoped ownership rules for tasks, tags, task-tag relationships, and reminders
- soft-delete strategy for tasks and tags
- explicit reminder management instead of implicit task-side reminder creation
- database-backed integrity rules for uniqueness and ownership boundaries

## Current Limitations

- default reminder delivery logs due reminders instead of sending through a real external transport
- task status is currently independent from reminder status
- reminder service tuning is currently hard-coded in `cmd/api/main.go` rather than exposed through environment variables
