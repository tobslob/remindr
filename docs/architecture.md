# Architecture

## Overview

Remindr is organized as a layered HTTP application:

1. `cmd/api` receives and validates HTTP requests
2. `cmd/tokens` handles JWT signing and verification
3. `internal/store` performs database reads and writes
4. `internal/db` manages connection setup and schema migrations
5. `internal/reminder` holds reminder domain types and runtime components

## Design Methodology

- API-first design with resource-oriented routes under `/v1`
- ownership-first data modeling so every task, tag, task-tag relationship, and reminder belongs to a user
- soft-delete policy for long-lived resources where recovery may matter
- explicit reminder management rather than coupling reminder creation to task creation
- database constraints used to reinforce ownership and uniqueness rules

## Request Flow

### Public endpoints

- request enters `chi` router in [`cmd/api/api.go`](../cmd/api/api.go)
- handler validates JSON and calls the appropriate store method
- handler writes JSON response via [`cmd/utils/json.go`](../cmd/utils/json.go)

### Authenticated endpoints

- `AuthMiddleware` in [`cmd/api/middlewares.go`](../cmd/api/middlewares.go)
- parses `Authorization` header
- verifies JWT through [`cmd/tokens/jwt.go`](../cmd/tokens/jwt.go)
- loads the user from the store
- injects the user into context via [`internal/requestctx/user.go`](../internal/requestctx/user.go)

## Store Layer

`Storage` in [`internal/store/storage.go`](../internal/store/storage.go) aggregates interfaces for:

- `Users`
- `Tasks`
- `Tags`
- `TaskTags`
- `Reminders`

Each concrete store uses `database/sql` directly and normalizes common DB errors with [`internal/store/errors.go`](../internal/store/errors.go).

## Users

[`internal/store/users.go`](../internal/store/users.go)

- users are hard-deleted
- email lookup powers login
- uniqueness conflicts map to `ErrConflict`

## Tasks

[`internal/store/tasks.go`](../internal/store/tasks.go)

Key behavior:

- tasks belong to a user
- tasks are soft-deleted with `deleted_at`
- task listing supports search, status, priority, created/due/completed time windows, and cursor pagination
- deleting tasks also deletes associated reminders in the same transaction

### Pagination

Task listing uses cursor pagination:

- ordered by `created_at DESC, id DESC`
- accepts `last_id`
- returns `next_last_id` when more rows are available

The parsing logic lives in [`cmd/utils/pagination.go`](../cmd/utils/pagination.go).

## Tags

[`internal/store/tags.go`](../internal/store/tags.go)

Key behavior:

- tags belong to a user
- tags are soft-deleted
- tag names are unique per user
- soft-deleted tags still reserve the name, which supports future undo/restore semantics

## Task/Tag Relationships

[`internal/store/taskTags.go`](../internal/store/taskTags.go)

Key behavior:

- many-to-many relationship between tasks and tags
- join rows are hard-deleted when detached
- attach, detach, and replace operations serialize per task with row locking
- ownership is enforced in both store logic and the database

The current API style is:

- attach one tag to one task
- replace the complete tag set for a task
- fetch tags for one or many tasks
- fetch tasks for one tag

## Reminders

[`internal/store/reminder.go`](../internal/store/reminder.go)

What exists today:

- authenticated reminder CRUD handlers in [`cmd/api/reminder.go`](../cmd/api/reminder.go)
- reminder persistence
- reminder cancellation
- claim-due batch transition from `pending` to `processing`
- mark reminder `sent`
- mark reminder back to `pending` or `failed`
- fetch one `processing` reminder for sending

Runtime behavior:

- the service starts in [`cmd/api/main.go`](../cmd/api/main.go) alongside the HTTP server
- the scheduler claims due reminders immediately on startup, then every 30 seconds by default
- each claim moves eligible reminders to `processing`, increments `attempts`, and enqueues jobs
- worker orchestration consumes claimed reminder jobs from an in-memory queue
- workers call `reminder.Sender`, then mark reminders `sent` or failed
- failed reminders return to `pending` until the third attempt, when the store marks them `failed`
- stale `processing` reminders older than 10 minutes can be claimed again
- default transport sender logs due reminders
- custom transports can implement `reminder.Sender`

What does not exist yet:

- automatic reminder creation from task flows

The runtime components live in:

- [`internal/reminder/model.go`](../internal/reminder/model.go)
- [`internal/reminder/scheduler.go`](../internal/reminder/scheduler.go)
- [`internal/reminder/service.go`](../internal/reminder/service.go)
- [`internal/reminder/worker.go`](../internal/reminder/worker.go)
- [`internal/reminder/sender.go`](../internal/reminder/sender.go)

## Integrity Strategy

The project uses both application-level checks and database constraints.

Examples:

- task/tag ownership is enforced by authenticated request flow and schema rules
- reminder/task ownership is enforced by authenticated request flow and schema rules
- duplicate logical reminders are rejected as part of the data model

## Runtime and Deployment

[`cmd/api/main.go`](../cmd/api/main.go) wires:

- env loading
- DB connection pool setup
- JWT maker creation
- store construction
- reminder service construction and startup
- HTTP server startup

Reminder service defaults:

- interval: `30s`
- batch size: `25`
- workers: `2`
- queue size: same as batch size

The service currently uses `reminder.NewLogSender(log.Default())`, so delivery is observable in logs but not sent to an external provider.

Development container support:

- [`Dockerfile.dev`](../Dockerfile.dev)
- [`docker-compose.yml`](../docker-compose.yml)

## Error Handling Conventions

- `ErrNotFound` for missing resources
- `ErrConflict` for uniqueness violations
- `ErrInvalidCursor` for invalid pagination cursors

Handlers map those store errors to HTTP responses through:

- [`cmd/utils/errors.go`](../cmd/utils/errors.go)

## Testing Status

Current automated coverage is light.

- there is at least one utility test file: [`cmd/utils/pagination_test.go`](../cmd/utils/pagination_test.go)
- store and reminder runtime packages do not currently have dedicated behavior tests

That means a lot of implementation confidence currently comes from `go test ./...` compile safety plus manual store review rather than broad end-to-end tests.
