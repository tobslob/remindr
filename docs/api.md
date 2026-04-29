# API Reference

Base path: `/v1`

## Auth

Authenticated routes require:

```http
Authorization: Bearer <token>
```

## Health

### `GET /healthz`

Returns plain text:

```text
OK
```

## Users

### `POST /users/register`

Creates a user.

Request body:

```json
{
  "username": "alice",
  "email": "alice@example.com",
  "password": "secret123"
}
```

Validation:

- `username` required
- `email` required and must be valid
- `password` required, 6-100 chars, alphanumeric

### `POST /users/login`

Returns a JWT.

Request body:

```json
{
  "email": "alice@example.com",
  "password": "secret123"
}
```

Response shape:

```json
{
  "token": "<jwt>"
}
```

## Tasks

### `POST /tasks/`

Creates a task.

Request body:

```json
{
  "title": "Pay rent",
  "description": "Transfer funds",
  "priority": "high",
  "due_at": "2026-05-01T09:00:00Z"
}
```

Notes:

- `priority` defaults to `medium` if omitted
- `due_at` must be in the future
- task `status` is created as `todo`

### `GET /tasks/`

Lists tasks for the authenticated user.

Supported query params:

- `limit`
- `last_id` or `lastId`
- `search` or `q`
- `status`
- `priority`
- `created_from`
- `created_to`
- `due_from`
- `due_to`
- `completed_from`
- `completed_to`

Examples:

```http
GET /v1/tasks/?limit=20&status=todo&priority=high
GET /v1/tasks/?q=rent&due_from=2026-05-01&due_to=2026-05-31
```

### `GET /tasks/{id}`

Returns one task owned by the authenticated user.

### `PATCH /tasks/{id}`

Updates a task.

Request body fields are optional:

```json
{
  "title": "New title",
  "description": "New description",
  "priority": "low",
  "due_at": "2026-06-01T10:00:00Z"
}
```

Notes:

- omitted fields keep their current values
- if provided, `due_at` must still be in the future

### `DELETE /tasks/{id}`

Soft-deletes one task.

Implementation note:

- associated reminder rows are deleted in the same transaction

### `DELETE /tasks/`

Soft-deletes all tasks for the authenticated user.

### `DELETE /tasks/bulk`

Soft-deletes many tasks at once.

Expected query style:

```http
DELETE /v1/tasks/bulk?ids=<uuid1>,<uuid2>
```

Use the exact format accepted by [`cmd/utils/getIDs.go`](../cmd/utils/getIDs.go).

## Tags

### `POST /tags/`

Creates a tag.

Request body:

```json
{
  "name": "work",
  "color": "#ffffff"
}
```

Notes:

- tag names are normalized to lowercase and trimmed
- tag names are unique per user

### `GET /tags/`

Lists all active tags for the authenticated user.

### `GET /tags/{id}`

Returns one tag owned by the authenticated user.

### `PATCH /tags/{id}`

Updates a tag.

Request body:

```json
{
  "name": "personal",
  "color": "#00ff00"
}
```

### `DELETE /tags/{id}`

Soft-deletes a tag.

## Task/Tag Relationships

### `POST /tasks/{id}/tags/{tag_id}`

Attaches one tag to one task.

### `PUT /tasks/{id}/tags`

Replaces the entire tag set for a task.

Request body:

```json
{
  "tag_ids": [
    "11111111-1111-1111-1111-111111111111",
    "22222222-2222-2222-2222-222222222222"
  ]
}
```

Notes:

- `tag_ids` is required
- duplicate IDs are normalized away before update
- an empty array clears all tag links for the task

### `GET /tasks/tags`

Fetches tags for many tasks.

Query style:

```http
GET /v1/tasks/tags?ids=<uuid1>,<uuid2>
```

Response shape is a map keyed by task ID.

### `GET /tags/{id}/tasks`

Returns tasks associated with one tag.

### `DELETE /tags/{task_id}/{id}`

Detaches tag `{id}` from task `{task_id}`.

## Reminders

### `POST /reminders/`

Creates a reminder for a task owned by the authenticated user.

Request body:

```json
{
  "task_id": "11111111-1111-1111-1111-111111111111",
  "type": "due_now",
  "remind_at": "2026-05-01T09:00:00Z"
}
```

Notes:

- `task_id`, `type`, and `remind_at` are required
- `type` must be one of `before_due` or `due_now`
- `remind_at` must be in the future
- the authenticated user is used as the reminder owner; the client does not supply `user_id`
- new reminders begin in the `pending` lifecycle state
- duplicate logical reminders are rejected if `(task_id, user_id, type, remind_at)` already exists

### `GET /reminders/task/{task_id}`

Returns all reminders for a task owned by the authenticated user.

### `GET /reminders/{id}`

Returns one reminder by ID for the authenticated user.

### `PATCH /reminders/{id}`

Updates a reminder.

Request body fields are optional:

```json
{
  "type": "before_due",
  "remind_at": "2026-06-01T10:00:00Z"
}
```

Notes:

- if provided, `type` must be one of `before_due` or `due_now`
- if provided, `remind_at` must be in the future

### `DELETE /reminders/{id}`

Deletes a reminder owned by the authenticated user.

## Error Behavior

Typical mappings:

- `400` invalid JSON or validation failure
- `401` missing or invalid auth
- `404` resource not found
- `409` uniqueness conflict
- `500` internal error

The store layer normalizes common SQL cases such as unique violations and missing rows before handlers convert them into HTTP responses.
