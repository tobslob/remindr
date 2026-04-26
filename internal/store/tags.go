package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type TagStore struct {
	db *sql.DB
}

type Tag struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Name      string     `json:"name"`
	Color     string     `json:"color"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

func (s *TagStore) Create(ctx context.Context, tag *Tag) error {
	query := `INSERT INTO tags (user_id, name, color) VALUES ($1, $2, $3) RETURNING id, created_at`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	if err := s.db.QueryRowContext(
		ctx,
		query,
		tag.UserID,
		tag.Name,
		tag.Color,
	).Scan(
		&tag.ID,
		&tag.CreatedAt,
	); err != nil {
		return normalizeStoreError(err)
	}
	return nil
}

func (s *TagStore) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Tag, error) {
	query := `SELECT id, user_id, name, color, created_at, updated_at, deleted_at FROM tags WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	tag := &Tag{}
	if err := s.db.QueryRowContext(ctx, query, id, userID).Scan(
		&tag.ID,
		&tag.UserID,
		&tag.Name,
		&tag.Color,
		&tag.CreatedAt,
		&tag.UpdatedAt,
		&tag.DeletedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return tag, nil
}

func (s *TagStore) GetTags(ctx context.Context, userID uuid.UUID) ([]*Tag, error) {
	query := `SELECT id, user_id, name, color, created_at, updated_at, deleted_at FROM tags WHERE user_id = $1 AND deleted_at IS NULL`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []*Tag
	for rows.Next() {
		tag := &Tag{}
		if err := rows.Scan(
			&tag.ID,
			&tag.UserID,
			&tag.Name,
			&tag.Color,
			&tag.CreatedAt,
			&tag.UpdatedAt,
			&tag.DeletedAt,
		); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (s *TagStore) DeleteByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	query := `UPDATE tags SET deleted_at = NOW() WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	result, err := s.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *TagStore) UpdateByID(ctx context.Context, id uuid.UUID, tag *Tag) error {
	query := `UPDATE tags SET name = $1, color = $2, updated_at = NOW() WHERE id = $3 AND user_id = $4 AND deleted_at IS NULL`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	result, err := s.db.ExecContext(ctx, query, tag.Name, tag.Color, id, tag.UserID)
	if err != nil {
		return normalizeStoreError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
