package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ItemStore struct {
	db *sql.DB
}

type Status string

const (
	Todo       Status   = "todo"
	InProgress Status   = "in_progress"
	Done       Status   = "done"
	Archived   Status   = "archived"
)

type Item struct {
	ID          uuid.UUID   `json:"id"`
	UserID      uuid.UUID   `json:"user_id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Status      Status      `json:"status"`
	Priority		*string     `json:"priority"`
	DueAt       *time.Time  `json:"due_at,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	CompletedAt	 *time.Time `json:"completed_at,omitempty"`
	DeletedAt	 *time.Time   `json:"deleted_at,omitempty"`
}

func (s *ItemStore) Create(ctx context.Context, item *Item) error {
	query := `INSERT INTO items (user_id, title, description, status, priority, due_at, completed_at, deleted_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	if err := s.db.QueryRowContext(
		ctx,
		query,
		item.UserID,
		item.Title,
		item.Description,
		item.Status,
		item.Priority,
		item.DueAt,
		item.CompletedAt,
		item.DeletedAt,
	).Scan(
		&item.ID,
		&item.CreatedAt,
	); err != nil {
		return err
	}
	return nil
}

func (s *ItemStore) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Item, error) {
	query := `SELECT id, user_id, title, description, status, priority, due_at, created_at, updated_at, completed_at FROM items WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	item := &Item{}
	if err := s.db.QueryRowContext(ctx, query, id, userID).Scan(
		&item.ID,
		&item.UserID,
		&item.Title,
		&item.Description,
		&item.Status,
		&item.Priority,
		&item.DueAt,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.CompletedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return item, nil
}

func (s *ItemStore) GetItems(ctx context.Context, userID uuid.UUID) ([]*Item, error) {
	query := `SELECT id, user_id, title, description, status, priority, due_at, created_at, updated_at, completed_at FROM items WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		item := &Item{}
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Title,
			&item.Description,
			&item.Status,
			&item.Priority,
			&item.DueAt,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.CompletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *ItemStore) DeleteByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	query := `UPDATE items SET deleted_at = NOW() WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`

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

func (s *ItemStore) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE items SET deleted_at = NOW() WHERE user_id = $1 AND deleted_at IS NULL`

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

func (s *ItemStore) DeleteByIDs(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) error {
	query := `UPDATE items SET deleted_at = NOW() WHERE id = ANY($1) AND user_id = $2 AND deleted_at IS NULL`

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

func (s *ItemStore) UpdateByID(ctx context.Context, userID uuid.UUID, item *Item) error {
	query := `UPDATE itemsSET title = $1, description = $2, status = $3, priority = $4, due_at = $5, completed_at = $6, updated_at = NOW() WHERE id = $7 AND user_id = $8 AND delete_at IS NULL`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(
		ctx,
		query,
		item.Title,
		item.Description,
		item.Status,
		item.Priority,
		item.DueAt,
		item.CompletedAt,
		item.ID,
		item.UserID,
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
