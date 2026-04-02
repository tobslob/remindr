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
	QueryTimeoutDuration = time.Second * 5
)

type Users interface {
	Create(context.Context, *User) error
	GetByID(context.Context, uuid.UUID) (*User, error)
	GetByEmail(context.Context, string) (*User, error)
	DeleteByID(context.Context, uuid.UUID) error
	UpdateByID(context.Context, *User) error
}

type Items interface {
	Create(context.Context, *Item) error
	GetByID(context.Context, uuid.UUID, uuid.UUID) (*Item, error)
	GetItems(context.Context, uuid.UUID) ([]*Item, error)
	DeleteByID(context.Context, uuid.UUID, uuid.UUID) error
	DeleteAllByUserID(context.Context, uuid.UUID) error
	DeleteByIDs(context.Context, []uuid.UUID, uuid.UUID) error
	UpdateByID(context.Context, uuid.UUID, *Item) error
}

type Storage struct {
	Users
	Items
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		Users: &UserStore{db: db},
		Items: &ItemStore{db: db},
	}
}
