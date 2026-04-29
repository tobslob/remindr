# Remindr

Remindr is a Go + PostgreSQL backend for user-owned tasks, tags, and reminder records.

The API supports:

- user registration and login with JWT authentication
- CRUD-style task management
- tag management with per-user unique tag names
- many-to-many task/tag relationships
- reminder persistence and lifecycle state transitions

Current reminder note:

- reminder rows and constraints are implemented in the database and store layer
- the runtime reminder scheduler/worker/sender packages are still stubs, so reminders are not automatically created or dispatched yet

## Tech Stack

- Go `1.26`
- PostgreSQL
- `chi` router
- JWT auth
- SQL migrations via `golang-migrate`

## Project Layout

- `cmd/api` HTTP server, handlers, middleware, route mounting
- `cmd/utils` request parsing, validation helpers, error responses, pagination
- `cmd/tokens` JWT creation and verification
- `internal/store` persistence layer
- `internal/db` database connection setup and migrations
- `internal/reminder` reminder domain package and future runtime hooks

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

See [docs/api.md](docs/api.md) for request and query details.

## Implementation Highlights

- tasks are soft-deleted with `deleted_at`
- deleting tasks also deletes their reminder rows in the same transaction
- tags are soft-deleted and keep unique name ownership per user
- task/tag joins are tenant-scoped both in store logic and the database
- reminders are tenant-scoped both in store logic and the database
- duplicate logical reminders are blocked by a unique constraint on `(task_id, user_id, type, remind_at)`

## Current Limitations

- reminder runtime processing is not wired yet; the scheduler/worker/sender files are placeholders
- task status is currently independent from reminder status
- reminder creation is available in the store layer, but no API endpoint or automatic task-side creation currently invokes it
