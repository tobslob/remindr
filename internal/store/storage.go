package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrConflict          = errors.New("resource already exist")
	QueryTimeoutDuration = time.Second * 5
)

type Users interface {
	Create(context.Context, *User) error
	GetByID(context.Context, string) (*User, error)
	GetByEmail(context.Context, string) (*User, error)
	DeleteByID(context.Context, string) error
	UpdateByID(context.Context, *User) error
}

type Items interface {
	Create(context.Context, *Item) error
	GetByID(context.Context, string) (*Item, error)
	GetAll(context.Context, string) ([]*Item, error)
	DeleteByID(context.Context, string) error
	DeleteAllByUserID(context.Context, string) error
	DeleteByIDs(context.Context, []string) error
	UpdateByID(context.Context, *Item) error
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
