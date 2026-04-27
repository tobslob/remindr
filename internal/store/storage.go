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
	AttachTagToTask(context.Context, *TaskTag) error
	GetTagsByTaskIDs(context.Context, []uuid.UUID, uuid.UUID) (map[uuid.UUID][]*Tag, error)
	GetTasksByTagID(context.Context, uuid.UUID, uuid.UUID) ([]*Task, error)
	DetachTagFromTask(context.Context, uuid.UUID, uuid.UUID) error
}

type Storage struct {
	Users
	Tasks
	Tags
	TaskTags
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		Users: &UserStore{db: db},
		Tasks: &TaskStore{db: db},
		Tags: &TagStore{db: db},
		TaskTags: &TaskTagStore{db: db},
	}
}
