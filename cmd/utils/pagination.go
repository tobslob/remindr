package utils

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tobslob/todoApp/internal/store"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

type PaginationQuery struct {
	LastID *uuid.UUID
	Limit  int
}

type TaskListQuery struct {
	PaginationQuery
	Search        string
	Status        *store.Status
	Priority      *string
	CreatedFrom   *time.Time
	CreatedTo     *time.Time
	DueFrom       *time.Time
	DueTo         *time.Time
	CompletedFrom *time.Time
	CompletedTo   *time.Time
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

func GetTaskListQuery(r *http.Request) (TaskListQuery, error) {
	pagination, err := GetPaginationFromQuery(r)
	if err != nil {
		return TaskListQuery{}, err
	}

	queryValues := r.URL.Query()
	search := strings.TrimSpace(getFirstQueryValue(queryValues, "search", "q"))
	status, err := parseTaskStatusQuery(queryValues, "status")
	if err != nil {
		return TaskListQuery{}, err
	}

	priority, err := parseTaskPriorityQuery(queryValues, "priority")
	if err != nil {
		return TaskListQuery{}, err
	}

	createdFrom, err := parseLowerBoundQuery(queryValues, "created_from", "created_at_from", "createdFrom", "createdAtFrom")
	if err != nil {
		return TaskListQuery{}, err
	}

	createdTo, err := parseUpperBoundQuery(queryValues, "created_to", "created_at_to", "createdTo", "createdAtTo")
	if err != nil {
		return TaskListQuery{}, err
	}

	dueFrom, err := parseLowerBoundQuery(queryValues, "due_from", "due_at_from", "dueFrom", "dueAtFrom")
	if err != nil {
		return TaskListQuery{}, err
	}

	dueTo, err := parseUpperBoundQuery(queryValues, "due_to", "due_at_to", "dueTo", "dueAtTo")
	if err != nil {
		return TaskListQuery{}, err
	}

	completedFrom, err := parseLowerBoundQuery(queryValues, "completed_from", "completed_at_from", "completedFrom", "completedAtFrom")
	if err != nil {
		return TaskListQuery{}, err
	}

	completedTo, err := parseUpperBoundQuery(queryValues, "completed_to", "completed_at_to", "completedTo", "completedAtTo")
	if err != nil {
		return TaskListQuery{}, err
	}

	if err := validateRange("created_from", "created_to", createdFrom, createdTo); err != nil {
		return TaskListQuery{}, err
	}

	if err := validateRange("due_from", "due_to", dueFrom, dueTo); err != nil {
		return TaskListQuery{}, err
	}

	if err := validateRange("completed_from", "completed_to", completedFrom, completedTo); err != nil {
		return TaskListQuery{}, err
	}

	return TaskListQuery{
		PaginationQuery: pagination,
		Search:          search,
		Status:          status,
		Priority:        priority,
		CreatedFrom:     createdFrom,
		CreatedTo:       createdTo,
		DueFrom:         dueFrom,
		DueTo:           dueTo,
		CompletedFrom:   completedFrom,
		CompletedTo:     completedTo,
	}, nil
}

func getFirstQueryValue(values url.Values, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(values.Get(key)); value != "" {
			return value
		}
	}

	return ""
}

func parseTaskStatusQuery(values url.Values, keys ...string) (*store.Status, error) {
	rawValue := getFirstQueryValue(values, keys...)
	if rawValue == "" {
		return nil, nil
	}

	normalizedStatus := normalizeEnumValue(rawValue, true)
	switch store.Status(normalizedStatus) {
	case store.Todo, store.InProgress, store.Done, store.Archived:
		status := store.Status(normalizedStatus)
		return &status, nil
	default:
		return nil, errors.New("status must be one of todo, in_progress, done, archived")
	}
}

func parseTaskPriorityQuery(values url.Values, keys ...string) (*string, error) {
	rawValue := getFirstQueryValue(values, keys...)
	if rawValue == "" {
		return nil, nil
	}

	normalizedPriority := normalizeEnumValue(rawValue, false)
	switch normalizedPriority {
	case "low", "medium", "high":
		return &normalizedPriority, nil
	default:
		return nil, errors.New("priority must be one of low, medium, high")
	}
}

func normalizeEnumValue(value string, mapHyphenToUnderscore bool) string {
	normalizedValue := strings.ToLower(strings.TrimSpace(value))
	if mapHyphenToUnderscore {
		normalizedValue = strings.ReplaceAll(normalizedValue, "-", "_")
	}

	return normalizedValue
}

func parseLowerBoundQuery(values url.Values, canonical string, aliases ...string) (*time.Time, error) {
	return parseTimeQuery(values, canonical, false, aliases...)
}

func parseUpperBoundQuery(values url.Values, canonical string, aliases ...string) (*time.Time, error) {
	return parseTimeQuery(values, canonical, true, aliases...)
}

func parseTimeQuery(values url.Values, canonical string, upperBound bool, aliases ...string) (*time.Time, error) {
	keys := append([]string{canonical}, aliases...)
	rawValue := getFirstQueryValue(values, keys...)
	if rawValue == "" {
		return nil, nil
	}

	parsedTime, isDateOnly, err := parseFlexibleTime(rawValue)
	if err != nil {
		return nil, fmt.Errorf("%s must be RFC3339 or YYYY-MM-DD", canonical)
	}

	if upperBound && isDateOnly {
		parsedTime = parsedTime.Add(24 * time.Hour)
	}

	return &parsedTime, nil
}

func parseFlexibleTime(value string) (time.Time, bool, error) {
	if parsedTime, err := time.Parse(time.RFC3339, value); err == nil {
		return parsedTime, false, nil
	}

	if parsedDate, err := time.Parse(time.DateOnly, value); err == nil {
		return parsedDate, true, nil
	}

	return time.Time{}, false, errors.New("invalid time format")
}

func validateRange(fromName string, toName string, from *time.Time, to *time.Time) error {
	if from != nil && to != nil && from.After(*to) {
		return fmt.Errorf("%s must be before %s", fromName, toName)
	}

	return nil
}
