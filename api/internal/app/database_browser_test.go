package app

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSensitiveDatabaseColumns(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"password_hash", "token_hash", "api_secret", "private_key"} {
		if !isSensitiveDatabaseColumn(name) {
			t.Fatalf("expected %q to be sensitive", name)
		}
	}
	for _, name := range []string{"username", "content", "created_at", "must_change_password"} {
		if isSensitiveDatabaseColumn(name) {
			t.Fatalf("did not expect %q to be sensitive", name)
		}
	}
}

func TestParseDatabasePagination(t *testing.T) {
	t.Parallel()
	request := httptest.NewRequest("GET", "/?limit=50&offset=100", nil)
	limit, offset, err := parseDatabasePagination(request)
	if err != nil || limit != 50 || offset != 100 {
		t.Fatalf("unexpected pagination: limit=%d offset=%d err=%v", limit, offset, err)
	}

	request = httptest.NewRequest("GET", "/?limit=101", nil)
	if _, _, err := parseDatabasePagination(request); err == nil {
		t.Fatal("expected oversized limit to fail")
	}
}

func TestDatabaseJSONValueTruncatesLongText(t *testing.T) {
	t.Parallel()
	value := strings.Repeat("数", databaseBrowserTextLimit+10)
	result, ok := databaseJSONValue(value).(string)
	if !ok || !strings.HasSuffix(result, "…") {
		t.Fatalf("expected long text to be truncated, got %T", result)
	}
}
