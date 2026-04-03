package utils

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

type PaginationQuery struct {
	LastID *uuid.UUID
	Limit  int
}

func GetPaginationFromQuery(r *http.Request) (PaginationQuery, error) {
	limit := DefaultLimit
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		value, err := strconv.Atoi(limitParam)
		if err != nil || value < 1 {
			return PaginationQuery{}, errors.New("limit must be a positive integer")
		}
		if value > MaxLimit {
			return PaginationQuery{}, fmt.Errorf("limit must be less than or equal to %d", MaxLimit)
		}
		limit = value
	}

	var lastID *uuid.UUID
	lastIDParam := r.URL.Query().Get("last_id")
	if lastIDParam == "" {
		lastIDParam = r.URL.Query().Get("lastId")
	}
	if lastIDParam != "" {
		value, err := uuid.Parse(lastIDParam)
		if err != nil {
			return PaginationQuery{}, errors.New("last_id must be a valid uuid")
		}
		lastID = &value
	}

	return PaginationQuery{
		LastID: lastID,
		Limit:  limit,
	}, nil
}
