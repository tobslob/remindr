package store

import (
	"context"
	"database/sql"
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
	Archived   Status = "archived"
)

type Task struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      Status     `json:"status"`
	Priority    *string    `json:"priority"`
	DueAt       *time.Time `json:"due_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

func (s *TaskStore) Create(ctx context.Context, task *Task) error {
	query := `INSERT INTO tasks (user_id, title, description, status, priority, due_at, completed_at, deleted_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at`

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
		task.CompletedAt,
		task.DeletedAt,
	).Scan(
		&task.ID,
		&task.CreatedAt,
	); err != nil {
		return err
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
		return nil, err
	}

	return task, nil
}

func (s *TaskStore) GetTasks(ctx context.Context, userID uuid.UUID) ([]*Task, error) {
	query := `SELECT id, user_id, title, description, status, priority, due_at, created_at, updated_at, completed_at FROM tasks WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (s *TaskStore) DeleteByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	query := `UPDATE tasks SET deleted_at = NOW() WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *TaskStore) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE tasks SET deleted_at = NOW() WHERE user_id = $1 AND deleted_at IS NULL`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *TaskStore) DeleteByIDs(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) error {
	query := `UPDATE tasks SET deleted_at = NOW() WHERE id = ANY($1) AND user_id = $2 AND deleted_at IS NULL`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, pq.Array(ids), userID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *TaskStore) UpdateByID(ctx context.Context, userID uuid.UUID, task *Task) error {
	query := `UPDATE tasks SET title = $1, description = $2, priority = $3, due_at = $4, updated_at = NOW() WHERE id = $5 AND user_id = $6 AND deleted_at IS NULL`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Priority,
		task.DueAt,
		task.ID,
		task.UserID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
