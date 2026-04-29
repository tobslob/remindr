package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrConflict          = errors.New("resource already exist")
	ErrInvalidCursor     = errors.New("invalid pagination cursor")
	QueryTimeoutDuration = time.Second * 5
)

type Users interface {
	Create(context.Context, *User) error
	GetByID(context.Context, uuid.UUID) (*User, error)
	GetByEmail(context.Context, string) (*User, error)
	DeleteByID(context.Context, uuid.UUID) error
	UpdateByID(context.Context, *User) error
}

type Tasks interface {
	Create(context.Context, *Task) error
	GetByID(context.Context, uuid.UUID, uuid.UUID) (*Task, error)
	GetTasks(context.Context, uuid.UUID, TaskFilter) (*TasksPage, error)
	DeleteByID(context.Context, uuid.UUID, uuid.UUID) error
	DeleteAllByUserID(context.Context, uuid.UUID) error
	DeleteByIDs(context.Context, []uuid.UUID, uuid.UUID) error
	UpdateByID(context.Context, uuid.UUID, *Task) error
}

type Tags interface {
	Create(context.Context, *Tag) error
	GetByID(context.Context, uuid.UUID, uuid.UUID) (*Tag, error)
	GetTags(context.Context, uuid.UUID) ([]*Tag, error)
	DeleteByID(context.Context, uuid.UUID, uuid.UUID) error
	UpdateByID(context.Context, uuid.UUID, *Tag) error
}

type TaskTags interface {
	AttachTagToTask(context.Context, uuid.UUID, *TaskTag) error
	GetTagsByTaskIDs(context.Context, []uuid.UUID, uuid.UUID) (map[uuid.UUID][]*Tag, error)
	GetTasksByTagID(context.Context, uuid.UUID, uuid.UUID) ([]*Task, error)
	DetachTagFromTask(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error
	ReplaceTaskTags(context.Context, uuid.UUID, uuid.UUID, []uuid.UUID) error
}

type Reminders interface {
	Create(context.Context, *Reminder) error
	GetByID(context.Context, uuid.UUID, uuid.UUID) (*Reminder, error)
	GetByTaskID(context.Context, uuid.UUID, uuid.UUID) ([]*Reminder, error)
	UpdateByID(context.Context, uuid.UUID, uuid.UUID, *Reminder) error
	DeleteByID(context.Context, uuid.UUID, uuid.UUID) error
	CancelByID(context.Context, uuid.UUID, uuid.UUID) error
	ClaimDue(context.Context, int64) ([]*Reminder, error)
	MarkSentByID(context.Context, uuid.UUID) error
	MarkFailedByID(context.Context, uuid.UUID, string) error
	GetForSendingByID(context.Context, uuid.UUID) (*Reminder, error)
}

type Storage struct {
	db *sql.DB
	Users
	Tasks
	Tags
	TaskTags
	Reminders
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		db:        db,
		Users:     &UserStore{db: db},
		Tasks:     &TaskStore{db: db},
		Tags:      &TagStore{db: db},
		TaskTags:  &TaskTagStore{db: db},
		Reminders: &ReminderStore{db: db},
	}
}
