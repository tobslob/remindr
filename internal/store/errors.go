package store

import (
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

	return err
}
