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
	GetTasks(context.Context, uuid.UUID, PaginationFilter) (*TasksPage, error)
	DeleteByID(context.Context, uuid.UUID, uuid.UUID) error
	DeleteAllByUserID(context.Context, uuid.UUID) error
	DeleteByIDs(context.Context, []uuid.UUID, uuid.UUID) error
	UpdateByID(context.Context, uuid.UUID, *Task) error
}

type Storage struct {
	Users
	Tasks
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		Users: &UserStore{db: db},
		Tasks: &TaskStore{db: db},
	}
}
