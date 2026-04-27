package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type TaskTagStore struct {
	db *sql.DB
}

type TaskTag struct {
	ID        uuid.UUID `json:"id"`
	TaskID    uuid.UUID `json:"task_id"`
	TagID     uuid.UUID `json:"tag_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *TaskTagStore) AttachTagToTask(ctx context.Context, taskTag *TaskTag) error {
	query := `INSERT INTO task_tags (task_id, tag_id) VALUES ($1, $2) RETURNING id, created_at`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	if err := s.db.QueryRowContext(
		ctx,
		query,
		taskTag.TaskID,
		taskTag.TagID,
	).Scan(
		&taskTag.ID,
		&taskTag.CreatedAt,
	); err != nil {
		return normalizeStoreError(err)
	}
	return nil
}

func (s *TaskTagStore) GetTagsByTaskIDs(ctx context.Context, taskIDs []uuid.UUID, userID uuid.UUID) (map[uuid.UUID][]*Tag, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	if len(taskIDs) == 0 {
		return map[uuid.UUID][]*Tag{}, nil
	}

	query := `
	SELECT
		tt.task_id,
		t.id,
		t.name,
		t.color,
		t.created_at
	FROM task_tags tt
	JOIN tasks tk ON tk.id = tt.task_id
	JOIN tags t ON t.id = tt.tag_id
	WHERE tt.task_id = ANY($1)
		AND tk.user_id = $2
		AND t.user_id = $2
		AND tk.deleted_at IS NULL
		AND t.deleted_at IS NULL
	ORDER BY tt.task_id, t.created_at ASC, t.id ASC;
	`

	rows, err := s.db.QueryContext(ctx, query, pq.Array(taskIDs), userID)
	if err != nil {
		return nil, normalizeStoreError(err)
	}
	defer rows.Close()

	tagsByTaskID := make(map[uuid.UUID][]*Tag)

	for rows.Next() {
		var taskID uuid.UUID
		tag := &Tag{}
		if err := rows.Scan(
			&taskID,
			&tag.ID,
			&tag.Name,
			&tag.Color,
			&tag.CreatedAt,
		); err != nil {
			return nil, normalizeStoreError(err)
		}
		tagsByTaskID[taskID] = append(tagsByTaskID[taskID], tag)
	}

	if err := rows.Err(); err != nil {
		return nil, normalizeStoreError(err)
	}
	return tagsByTaskID, nil
}

func (s *TaskTagStore) DetachTagFromTask(ctx context.Context, taskID uuid.UUID, tagID uuid.UUID) error {
	query := `DELETE FROM task_tags WHERE task_id = $1 AND tag_id = $2`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	if _, err := s.db.ExecContext(ctx, query, taskID, tagID); err != nil {
		return normalizeStoreError(err)
	}
	return nil
}

func (s *TaskTagStore) GetTasksByTagID(ctx context.Context, tagID uuid.UUID, userID uuid.UUID) ([]*Task, error) {
	query := `
	SELECT
		tk.id,
		tk.user_id,
		tk.title,
		tk.description,
		tk.status,
		tk.priority,
		tk.due_at,
		tk.created_at,
		tk.updated_at,
		tk.completed_at
	FROM task_tags tt
	JOIN tasks tk ON tk.id = tt.task_id
	JOIN tags t ON t.id = tt.tag_id
	WHERE tt.tag_id = $1
		AND tk.user_id = $2
		AND t.user_id = $2
		AND tk.deleted_at IS NULL
		AND t.deleted_at IS NULL;
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, tagID, userID)
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
	return tasks, nil
}
