# Data Model

This document summarizes the current logical schema represented by the migrations in `internal/db/migrations`.

## Users

Core table: `users`

Important fields:

- `id`
- `username`
- `email`
- `password`
- `created_at`
- `updated_at`

Important rules:

- `username` unique
- `email` unique

## Tasks

Core table: `tasks`

Important fields:

- `id`
- `user_id`
- `title`
- `description`
- `status`
- `priority`
- `due_at`
- `completed_at`
- `deleted_at`
- `created_at`
- `updated_at`

Important rules:

- task belongs to one user
- `due_at` is enforced `NOT NULL`
- tasks are soft-deleted via `deleted_at`
- active-listing and search indexes exist for performance

Status values in code:

- `todo`
- `in_progress`
- `done`

## Tags

Core table: `tags`

Important fields:

- `id`
- `user_id`
- `name`
- `color`
- `deleted_at`
- `created_at`
- `updated_at`

Important rules:

- tag belongs to one user
- tag is soft-deleted
- unique name per user
- composite uniqueness also exists on `(id, user_id)` to support ownership-safe foreign keys

## Task Tags

Core table: `task_tags`

Important fields:

- `task_id`
- `tag_id`
- `user_id`
- `created_at`

Important rules:

- primary key on `(task_id, tag_id)`
- row belongs to one user through `user_id`
- composite foreign key `(task_id, user_id) -> tasks(id, user_id)`
- composite foreign key `(tag_id, user_id) -> tags(id, user_id)`
- hard-deleted on detach

This prevents linking a task owned by one user to a tag owned by another user.

## Reminders

Core table: `reminders`

Important fields:

- `id`
- `task_id`
- `user_id`
- `type`
- `status`
- `remind_at`
- `attempts`
- `last_attempt_error`
- `sent_at`
- `created_at`
- `updated_at`

Reminder type values in code:

- `before_due`
- `due_now`

Reminder status values in code:

- `pending`
- `processing`
- `sent`
- `failed`
- `cancelled`

Important rules:

- reminder belongs to a task and a user
- composite foreign key `(task_id, user_id) -> tasks(id, user_id)`
- duplicate logical reminders are blocked by unique constraint on `(task_id, user_id, type, remind_at)`

## Ownership Strategy

The project uses both:

- application/store-level checks
- database constraints

This is especially important in relationship tables like `task_tags` and `reminders`, where tenant mismatches can be subtle and dangerous.

## Soft Delete Strategy

Soft-deleted:

- `tasks`
- `tags`

Hard-deleted:

- `task_tags`
- `reminders` when their parent task is soft-deleted by the task store cleanup path

## Reminder Runtime Status

The schema and store support reminders, but the runtime processing loop is not implemented yet.

That means:

- reminder rows can be created, updated, deleted, and constrained correctly
- claim/send lifecycle methods exist in the store
- scheduler/worker/sender runtime components are still placeholders
