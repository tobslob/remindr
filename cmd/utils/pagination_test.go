package utils

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tobslob/todoApp/internal/store"
)

func TestGetTaskListQueryParsesSearchAndRanges(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/tasks?limit=10&search=quarterly%20report&status=in-progress&priority=HIGH&created_from=2026-04-01&created_to=2026-04-01&dueAtFrom=2026-04-02T15:04:05Z&completed_at_to=2026-04-03", nil)

	query, err := GetTaskListQuery(req)
	if err != nil {
		t.Fatalf("expected query to parse, got error: %v", err)
	}

	if query.Limit != 10 {
		t.Fatalf("expected limit 10, got %d", query.Limit)
	}

	if query.Search != "quarterly report" {
		t.Fatalf("expected trimmed search term, got %q", query.Search)
	}

	assertStatusEquals(t, query.Status, store.InProgress)
	assertStringEquals(t, query.Priority, "high")
	assertTimeEquals(t, query.CreatedFrom, time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC))
	assertTimeEquals(t, query.CreatedTo, time.Date(2026, time.April, 2, 0, 0, 0, 0, time.UTC))
	assertTimeEquals(t, query.DueFrom, time.Date(2026, time.April, 2, 15, 4, 5, 0, time.UTC))
	assertTimeEquals(t, query.CompletedTo, time.Date(2026, time.April, 4, 0, 0, 0, 0, time.UTC))
}

func TestGetTaskListQueryRejectsInvalidRange(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/tasks?created_from=2026-04-06&created_to=2026-04-04", nil)

	_, err := GetTaskListQuery(req)
	if err == nil {
		t.Fatal("expected invalid range error, got nil")
	}
}

func TestGetTaskListQueryRejectsInvalidTimeFormat(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/tasks?due_to=04-03-2026", nil)

	_, err := GetTaskListQuery(req)
	if err == nil {
		t.Fatal("expected invalid time format error, got nil")
	}
}

func TestGetTaskListQueryRejectsInvalidStatus(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/tasks?status=started", nil)

	_, err := GetTaskListQuery(req)
	if err == nil {
		t.Fatal("expected invalid status error, got nil")
	}
}

func TestGetTaskListQueryRejectsInvalidPriority(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/tasks?priority=urgent", nil)

	_, err := GetTaskListQuery(req)
	if err == nil {
		t.Fatal("expected invalid priority error, got nil")
	}
}

func assertTimeEquals(t *testing.T, actual *time.Time, expected time.Time) {
	t.Helper()

	if actual == nil {
		t.Fatalf("expected time %s, got nil", expected.Format(time.RFC3339))
	}

	if !actual.Equal(expected) {
		t.Fatalf("expected time %s, got %s", expected.Format(time.RFC3339), actual.Format(time.RFC3339))
	}
}

func assertStatusEquals(t *testing.T, actual *store.Status, expected store.Status) {
	t.Helper()

	if actual == nil {
		t.Fatalf("expected status %s, got nil", expected)
	}

	if *actual != expected {
		t.Fatalf("expected status %s, got %s", expected, *actual)
	}
}

func assertStringEquals(t *testing.T, actual *string, expected string) {
	t.Helper()

	if actual == nil {
		t.Fatalf("expected string %q, got nil", expected)
	}

	if *actual != expected {
		t.Fatalf("expected string %q, got %q", expected, *actual)
	}
}
