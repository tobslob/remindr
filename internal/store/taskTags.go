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
	TaskID    uuid.UUID `json:"task_id"`
	TagID     uuid.UUID `json:"tag_id"`
	CreatedAt time.Time `json:"created_at"`
}

func deduplicateUUIDs(ids []uuid.UUID) []uuid.UUID {
	if len(ids) == 0 {
		return ids
	}

	deduplicated := make([]uuid.UUID, 0, len(ids))
	seen := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		if _, exists := seen[id]; exists {
			continue
		}

		seen[id] = struct{}{}
		deduplicated = append(deduplicated, id)
	}

	return deduplicated
}

func (s *TaskTagStore) lockTaskForTagMutation(ctx context.Context, tx *sql.Tx, taskID uuid.UUID, userID uuid.UUID) error {
	var lockedTaskID uuid.UUID
	if err := tx.QueryRowContext(
		ctx,
		`SELECT id FROM tasks WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL FOR UPDATE`,
		taskID,
		userID,
	).Scan(&lockedTaskID); err != nil {
		return normalizeStoreError(err)
	}

	return nil
}

func (s *TaskTagStore) ensureActiveTag(ctx context.Context, tx *sql.Tx, tagID uuid.UUID, userID uuid.UUID) error {
	var lockedTagID uuid.UUID
	if err := tx.QueryRowContext(
		ctx,
		`SELECT id FROM tags WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL FOR UPDATE`,
		tagID,
		userID,
	).Scan(&lockedTagID); err != nil {
		return normalizeStoreError(err)
	}

	return nil
}

func (s *TaskTagStore) AttachTagToTask(ctx context.Context, userID uuid.UUID, taskTag *TaskTag) error {
	query := `INSERT INTO task_tags (task_id, tag_id, user_id) VALUES ($1, $2, $3) RETURNING task_id, tag_id, created_at`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return normalizeStoreError(err)
	}
	defer tx.Rollback()

	if err := s.lockTaskForTagMutation(ctx, tx, taskTag.TaskID, userID); err != nil {
		return err
	}

	if err := s.ensureActiveTag(ctx, tx, taskTag.TagID, userID); err != nil {
		return err
	}

	if err := tx.QueryRowContext(
		ctx,
		query,
		taskTag.TaskID,
		taskTag.TagID,
		userID,
	).Scan(
		&taskTag.TaskID,
		&taskTag.TagID,
		&taskTag.CreatedAt,
	); err != nil {
		return normalizeStoreError(err)
	}

	if err := tx.Commit(); err != nil {
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
		t.user_id,
		t.id,
		t.name,
		t.color,
		t.created_at
	FROM task_tags tt
	JOIN tasks tk ON tk.id = tt.task_id AND tk.user_id = tt.user_id
	JOIN tags t ON t.id = tt.tag_id AND t.user_id = tt.user_id
	WHERE tt.task_id = ANY($1)
		AND tt.user_id = $2
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
			&tag.UserID,
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

func (s *TaskTagStore) DetachTagFromTask(ctx context.Context, taskID uuid.UUID, tagID uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM task_tags WHERE task_id = $1 AND tag_id = $2 AND user_id = $3`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return normalizeStoreError(err)
	}
	defer tx.Rollback()

	if err := s.lockTaskForTagMutation(ctx, tx, taskID, userID); err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, query, taskID, tagID, userID)
	if err != nil {
		return normalizeStoreError(err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return normalizeStoreError(err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return normalizeStoreError(err)
	}

	return nil
}

func (s *TaskTagStore) ReplaceTaskTags(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, tagIDs []uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	tagIDs = deduplicateUUIDs(tagIDs)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return normalizeStoreError(err)
	}
	defer tx.Rollback()

	if err := s.lockTaskForTagMutation(ctx, tx, taskID, userID); err != nil {
		return err
	}

	if len(tagIDs) > 0 {
		rows, err := tx.QueryContext(
			ctx,
			`SELECT id FROM tags WHERE id = ANY($1) AND user_id = $2 AND deleted_at IS NULL FOR UPDATE`,
			pq.Array(tagIDs),
			userID,
		)
		if err != nil {
			return normalizeStoreError(err)
		}

		validTagCount := 0
		for rows.Next() {
			validTagCount++
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return normalizeStoreError(err)
		}
		if err := rows.Close(); err != nil {
			return normalizeStoreError(err)
		}

		if validTagCount != len(tagIDs) {
			return ErrNotFound
		}

		if _, err := tx.ExecContext(
			ctx,
			`DELETE FROM task_tags WHERE task_id = $1 AND user_id = $2 AND NOT (tag_id = ANY($3))`,
			taskID,
			userID,
			pq.Array(tagIDs),
		); err != nil {
			return normalizeStoreError(err)
		}

		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO task_tags (task_id, tag_id, user_id)
			 SELECT $1, u.tag_id, $2
			 FROM unnest($3::uuid[]) AS u(tag_id)
			 ON CONFLICT (task_id, tag_id) DO NOTHING`,
			taskID,
			userID,
			pq.Array(tagIDs),
		); err != nil {
			return normalizeStoreError(err)
		}
	} else {
		if _, err := tx.ExecContext(ctx, `DELETE FROM task_tags WHERE task_id = $1 AND user_id = $2`, taskID, userID); err != nil {
			return normalizeStoreError(err)
		}
	}

	if err := tx.Commit(); err != nil {
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
	JOIN tasks tk ON tk.id = tt.task_id AND tk.user_id = tt.user_id
	JOIN tags t ON t.id = tt.tag_id AND t.user_id = tt.user_id
	WHERE tt.tag_id = $1
		AND tt.user_id = $2
		AND tk.deleted_at IS NULL
		AND t.deleted_at IS NULL
	ORDER BY tk.created_at DESC, tk.id DESC;
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
