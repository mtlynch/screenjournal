package test_sqlite_test

import (
	"strings"
	"testing"

	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

func TestNewDBEnforcesForeignKeys(t *testing.T) {
	db := test_sqlite.NewDB(t)

	_, err := db.Exec(`
		INSERT INTO auth_sessions (session_id, user_id, created_at, expires_at)
		VALUES ('session-id', 'missing-user', '2026-01-02T03:04:05Z', '2026-01-03T03:04:05Z')
	`)

	if err == nil {
		t.Fatalf("err=nil, want foreign key constraint error")
	}
	if got, want := err.Error(), "FOREIGN KEY constraint failed"; !strings.Contains(got, want) {
		t.Errorf("err=%q, want substring %q", got, want)
	}
}
