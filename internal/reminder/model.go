package reminder

import (
	"time"

	"github.com/google/uuid"
)

type ReminderType string

const (
	ReminderTypeBeforeDue ReminderType = "before_due"
	ReminderTypeDueNow    ReminderType = "due_now"
)

type ReminderStatus string

const (
	ReminderStatusPending    ReminderStatus = "pending"
	ReminderStatusProcessing ReminderStatus = "processing"
	ReminderStatusSent       ReminderStatus = "sent"
	ReminderStatusFailed     ReminderStatus = "failed"
	ReminderStatusCancelled  ReminderStatus = "cancelled"
)

type Job struct {
	ReminderID uuid.UUID
}

type Record struct {
	ID     uuid.UUID `json:"id"`
	TaskID uuid.UUID `json:"task_id"`
	UserID uuid.UUID `json:"user_id"`

	Type   ReminderType   `json:"type"`
	Status ReminderStatus `json:"status"`

	RemindAt time.Time `json:"remind_at"`

	Attempts         int     `json:"attempts"`
	LastAttemptError *string `json:"last_attempt_error"`

	SentAt *time.Time `json:"sent_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
