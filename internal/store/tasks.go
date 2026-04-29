package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type TaskStore struct {
	db *sql.DB
}

type Status string

const (
	Todo       Status = "todo"
	InProgress Status = "in_progress"
	Done       Status = "done"
)

type Task struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      Status     `json:"status"`
	Priority    *string    `json:"priority"`
	DueAt       time.Time  `json:"due_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type TaskFilter struct {
	LastID *uuid.UUID
	Limit  int
	Search string
	Status *Status
	// Priority remains a string because the column is still free-form in the
	// schema, even though the API validates the supported values.
	Priority *string
	// Upper bounds are exclusive so date-only query params can map cleanly
	// to "before next day" filters.
	CreatedFrom     *time.Time
	CreatedBefore   *time.Time
	DueFrom         *time.Time
	DueBefore       *time.Time
	CompletedFrom   *time.Time
	CompletedBefore *time.Time
}

type PaginationMetadata struct {
	Limit      int        `json:"limit"`
	LastID     *uuid.UUID `json:"last_id,omitempty"`
	NextLastID *uuid.UUID `json:"next_last_id,omitempty"`
	HasMore    bool       `json:"has_more"`
}

type TasksPage struct {
	Tasks      []*Task            `json:"tasks"`
	Pagination PaginationMetadata `json:"pagination"`
}

func (s *TaskStore) Create(ctx context.Context, task *Task) error {
	query := `INSERT INTO tasks (user_id, title, description, status, priority, due_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	if err := s.db.QueryRowContext(
		ctx,
		query,
		task.UserID,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.DueAt,
	).Scan(
		&task.ID,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return normalizeStoreError(err)
	}
	return nil
}

func (s *TaskStore) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Task, error) {
	query := `SELECT id, user_id, title, description, status, priority, due_at, created_at, updated_at, completed_at FROM tasks WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	task := &Task{}
	if err := s.db.QueryRowContext(ctx, query, id, userID).Scan(
		&task.ID,
		&task.UserID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.DueAt,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.CompletedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, normalizeStoreError(err)
	}

	return task, nil
}

func (s *TaskStore) GetTasks(ctx context.Context, userID uuid.UUID, filter TaskFilter) (*TasksPage, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `SELECT id, user_id, title, description, status, priority, due_at, created_at, updated_at, completed_at FROM tasks WHERE user_id = $1 AND deleted_at IS NULL`
	args := []any{userID}

	if filter.Search != "" {
		args = append(args, "%"+filter.Search+"%")
		query += fmt.Sprintf(` AND ((title || ' ' || description) ILIKE $%d)`, len(args))
	}

	if filter.Status != nil {
		args = append(args, *filter.Status)
		query += fmt.Sprintf(` AND status = $%d`, len(args))
	}

	if filter.Priority != nil {
		args = append(args, *filter.Priority)
		query += fmt.Sprintf(` AND priority = $%d`, len(args))
	}

	query, args = appendTimeRangeFilter(query, args, "created_at", filter.CreatedFrom, filter.CreatedBefore)
	query, args = appendTimeRangeFilter(query, args, "due_at", filter.DueFrom, filter.DueBefore)
	query, args = appendTimeRangeFilter(query, args, "completed_at", filter.CompletedFrom, filter.CompletedBefore)

	if filter.LastID != nil {
		cursorCreatedAt, err := s.getTaskCursorCreatedAt(ctx, userID, *filter.LastID)
		if err != nil {
			return nil, err
		}

		cursorCreatedAtParamPosition := len(args) + 1
		cursorIDParamPosition := len(args) + 2
		query += fmt.Sprintf(
			` AND (created_at < $%d OR (created_at = $%d AND id < $%d))`,
			cursorCreatedAtParamPosition,
			cursorCreatedAtParamPosition,
			cursorIDParamPosition,
		)
		args = append(args, cursorCreatedAt, *filter.LastID)
	}

	limitParamPosition := len(args) + 1
	query += fmt.Sprintf(` ORDER BY created_at DESC, id DESC LIMIT $%d`, limitParamPosition)

	rows, err := s.db.QueryContext(ctx, query, append(args, filter.Limit+1)...)
	if err != nil {
		return nil, normalizeStoreError(err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		if err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Priority,
			&task.DueAt,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.CompletedAt,
		); err != nil {
			return nil, normalizeStoreError(err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, normalizeStoreError(err)
	}

	hasMore := len(tasks) > filter.Limit

	var nextLastID *uuid.UUID
	if hasMore {
		nextID := tasks[filter.Limit-1].ID
		nextLastID = &nextID
		tasks = tasks[:filter.Limit]
	}

	return &TasksPage{
		Tasks: tasks,
		Pagination: PaginationMetadata{
			Limit:      filter.Limit,
			LastID:     filter.LastID,
			NextLastID: nextLastID,
			HasMore:    hasMore,
		},
	}, nil
}

func appendTimeRangeFilter(query string, args []any, column string, from *time.Time, before *time.Time) (string, []any) {
	if from != nil {
		args = append(args, *from)
		query += fmt.Sprintf(" AND %s >= $%d", column, len(args))
	}

	if before != nil {
		args = append(args, *before)
		query += fmt.Sprintf(" AND %s < $%d", column, len(args))
	}

	return query, args
}

func (s *TaskStore) getTaskCursorCreatedAt(ctx context.Context, userID uuid.UUID, lastID uuid.UUID) (time.Time, error) {
	query := `SELECT created_at FROM tasks WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`

	var createdAt time.Time
	if err := s.db.QueryRowContext(ctx, query, lastID, userID).Scan(&createdAt); err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, ErrInvalidCursor
		}
		return time.Time{}, normalizeStoreError(err)
	}

	return createdAt, nil
}

func softDeleteTasks(ctx context.Context, tx *sql.Tx, query string, args ...any) ([]uuid.UUID, error) {
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, normalizeStoreError(err)
	}
	defer rows.Close()

	var taskIDs []uuid.UUID
	for rows.Next() {
		var taskID uuid.UUID
		if err := rows.Scan(&taskID); err != nil {
			return nil, normalizeStoreError(err)
		}
		taskIDs = append(taskIDs, taskID)
	}

	if err := rows.Err(); err != nil {
		return nil, normalizeStoreError(err)
	}

	if len(taskIDs) == 0 {
		return nil, ErrNotFound
	}

	return taskIDs, nil
}

func deleteRemindersForTaskIDs(ctx context.Context, tx *sql.Tx, taskIDs []uuid.UUID) error {
	if len(taskIDs) == 0 {
		return nil
	}

	_, err := tx.ExecContext(ctx, `DELETE FROM reminders WHERE task_id = ANY($1)`, pq.Array(taskIDs))
	return normalizeStoreError(err)
}

func (s *TaskStore) DeleteByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return normalizeStoreError(err)
	}
	defer tx.Rollback()

	taskIDs, err := softDeleteTasks(
		ctx,
		tx,
		`UPDATE tasks SET deleted_at = NOW() WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL RETURNING id`,
		id,
		userID,
	)
	if err != nil {
		return normalizeStoreError(err)
	}

	if err := deleteRemindersForTaskIDs(ctx, tx, taskIDs); err != nil {
		return normalizeStoreError(err)
	}

	return normalizeStoreError(tx.Commit())
}

func (s *TaskStore) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return normalizeStoreError(err)
	}
	defer tx.Rollback()

	taskIDs, err := softDeleteTasks(
		ctx,
		tx,
		`UPDATE tasks SET deleted_at = NOW() WHERE user_id = $1 AND deleted_at IS NULL RETURNING id`,
		userID,
	)
	if err != nil {
		return normalizeStoreError(err)
	}

	if err := deleteRemindersForTaskIDs(ctx, tx, taskIDs); err != nil {
		return normalizeStoreError(err)
	}

	return normalizeStoreError(tx.Commit())
}

func (s *TaskStore) DeleteByIDs(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return normalizeStoreError(err)
	}
	defer tx.Rollback()

	taskIDs, err := softDeleteTasks(
		ctx,
		tx,
		`UPDATE tasks SET deleted_at = NOW() WHERE id = ANY($1) AND user_id = $2 AND deleted_at IS NULL RETURNING id`,
		pq.Array(ids),
		userID,
	)
	if err != nil {
		return normalizeStoreError(err)
	}

	if err := deleteRemindersForTaskIDs(ctx, tx, taskIDs); err != nil {
		return normalizeStoreError(err)
	}

	return normalizeStoreError(tx.Commit())
}

func (s *TaskStore) UpdateByID(ctx context.Context, userID uuid.UUID, task *Task) error {
	query := `UPDATE tasks
	SET title = $1,
	    description = $2,
	    priority = $3,
	    due_at = $4,
	    updated_at = NOW()
	WHERE id = $5 AND user_id = $6 AND deleted_at IS NULL
	RETURNING id, user_id, title, description, status, priority, due_at, created_at, updated_at, completed_at, deleted_at`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	if err := s.db.QueryRowContext(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Priority,
		task.DueAt,
		task.ID,
		userID,
	).Scan(
		&task.ID,
		&task.UserID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.DueAt,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.CompletedAt,
		&task.DeletedAt,
	); err != nil {
		return normalizeStoreError(err)
	}

	return nil
}
