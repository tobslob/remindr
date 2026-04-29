package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/tobslob/remindr/internal/reminder"
)

type ReminderStore struct {
	db *sql.DB
}

type Reminder struct {
	ID     uuid.UUID
	TaskID uuid.UUID
	UserID uuid.UUID

	Type   reminder.ReminderType
	Status reminder.ReminderStatus

	RemindAt time.Time

	Attempts         int
	LastAttemptError *string

	SentAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s *ReminderStore) CreateReminder(ctx context.Context, r *Reminder) error {
	query :=
		`INSERT INTO reminders (task_id, user_id, type, status, remind_at, attempts, last_attempt_error, sent_at, created_at, updated_at)
		SELECT t.id, t.user_id, $3, $4, $5, $6, $7, $8, $9, $10
		FROM tasks t
		WHERE t.id = $1 AND t.user_id = $2 AND t.deleted_at IS NULL
		RETURNING id, created_at, updated_at`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	if err := s.db.QueryRowContext(
		ctx,
		query,
		r.TaskID,
		r.UserID,
		r.Type,
		r.Status,
		r.RemindAt,
		r.Attempts,
		r.LastAttemptError,
		r.SentAt,
		time.Now(),
		time.Now(),
	).Scan(
		&r.ID,
		&r.CreatedAt,
		&r.UpdatedAt,
	); err != nil {
		return normalizeStoreError(err)
	}

	return nil
}

func (s *ReminderStore) CancelReminder(ctx context.Context, id uuid.UUID, userId uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var reminderID uuid.UUID
	if err := s.db.QueryRowContext(
		ctx,
		`UPDATE reminders SET status = 'cancelled', updated_at = now() WHERE id = $1 AND user_id = $2 AND status = 'pending' RETURNING id`,
		id,
		userId,
	).Scan(&reminderID); err != nil {
		return normalizeStoreError(err)
	}

	return nil
}

func (s *ReminderStore) ClaimDueReminders(ctx context.Context, limit int64) ([]*Reminder, error) {
	query := `
	UPDATE reminders
	SET status = 'processing',
	    attempts = attempts + 1,
	    updated_at = now()
	WHERE id IN (
	    SELECT r.id
	    FROM reminders r
	    JOIN tasks t ON t.id = r.task_id AND t.user_id = r.user_id
	    WHERE r.remind_at <= now()
	      AND r.status = 'pending'
	      AND t.deleted_at IS NULL
	    ORDER BY r.remind_at ASC
	    LIMIT $1
	    FOR UPDATE OF r SKIP LOCKED
	)
	RETURNING id, task_id, user_id, type, status, remind_at, attempts, last_attempt_error, sent_at, created_at, updated_at;`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, normalizeStoreError(err)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, normalizeStoreError(err)
	}
	defer rows.Close()

	var reminders []*Reminder
	for rows.Next() {
		r := &Reminder{}
		if err := rows.Scan(
			&r.ID,
			&r.TaskID,
			&r.UserID,
			&r.Type,
			&r.Status,
			&r.RemindAt,
			&r.Attempts,
			&r.LastAttemptError,
			&r.SentAt,
			&r.CreatedAt,
			&r.UpdatedAt,
		); err != nil {
			return nil, normalizeStoreError(err)
		}
		reminders = append(reminders, r)
	}

	if err := rows.Err(); err != nil {
		return nil, normalizeStoreError(err)
	}

	if err := rows.Close(); err != nil {
		return nil, normalizeStoreError(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, normalizeStoreError(err)
	}

	return reminders, nil
}

func (s *ReminderStore) MarkReminderSent(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var reminderID uuid.UUID
	if err := s.db.QueryRowContext(
		ctx,
		`UPDATE reminders SET status = 'sent', sent_at = now(), updated_at = now() WHERE id = $1 AND status = 'processing' RETURNING id`,
		id,
	).Scan(&reminderID); err != nil {
		return normalizeStoreError(err)
	}

	return nil
}

func (s *ReminderStore) MarkReminderFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var reminderID uuid.UUID
	if err := s.db.QueryRowContext(
		ctx,
		`UPDATE reminders
		SET status = CASE
				WHEN attempts >= 3 THEN 'failed'
				ELSE 'pending'
			END,
		last_attempt_error = $2,
		updated_at = now()
		WHERE id = $1 AND status = 'processing'
		RETURNING id`,
		id,
		errMsg,
	).Scan(&reminderID); err != nil {
		return normalizeStoreError(err)
	}

	return nil
}

func (s *ReminderStore) GetReminderForSending(ctx context.Context, id uuid.UUID) (*Reminder, error) {
	query := `
	SELECT r.id, r.task_id, r.user_id, r.type, r.status, r.remind_at, r.attempts, r.last_attempt_error, r.sent_at, r.created_at, r.updated_at
	FROM reminders r
	JOIN tasks t ON t.id = r.task_id AND t.user_id = r.user_id
	WHERE r.id = $1
	  AND r.status = 'processing'
	  AND t.deleted_at IS NULL`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	r := &Reminder{}
	if err := s.db.QueryRowContext(ctx, query, id).Scan(
		&r.ID,
		&r.TaskID,
		&r.UserID,
		&r.Type,
		&r.Status,
		&r.RemindAt,
		&r.Attempts,
		&r.LastAttemptError,
		&r.SentAt,
		&r.CreatedAt,
		&r.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, normalizeStoreError(err)
		}
		return nil, normalizeStoreError(err)
	}

	return r, nil
}
