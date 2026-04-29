package reminder

import (
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
