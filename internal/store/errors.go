package store

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

func normalizeStoreError(err error) error {
	if err == nil {
		return nil
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return ErrConflict
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}

	return err
}
