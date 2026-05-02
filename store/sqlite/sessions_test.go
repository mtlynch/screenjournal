package sqlite_test

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	simple_sessions "codeberg.org/mtlynch/simpleauth/v3/sessions"

	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

func TestCreateSessionRejectsDuplicateSessionID(t *testing.T) {
	store := test_sqlite.New()
	ctx := context.Background()
	sessionID := mustSessionID(t)
	originalUserID := mustUserID(t, "original-user")
	replacementUserID := mustUserID(t, "replacement-user")
	createdAt := time.Date(2026, time.April, 19, 12, 30, 0, 0, time.UTC)

	originalSession := simple_sessions.Session{
		ID:        sessionID,
		UserID:    originalUserID,
		CreatedAt: createdAt,
		ExpiresAt: createdAt.Add(24 * time.Hour),
	}
	if err := store.CreateSession(ctx, originalSession); err != nil {
		t.Fatalf("CreateSession err=%v, want=%v", err, nil)
	}

	replacementSession := simple_sessions.Session{
		ID:        sessionID,
		UserID:    replacementUserID,
		CreatedAt: createdAt.Add(time.Hour),
		ExpiresAt: createdAt.Add(48 * time.Hour),
	}
	if err := store.CreateSession(ctx, replacementSession); err == nil {
		t.Fatal("CreateSession err=nil, want duplicate session ID error")
	}

	session, err := store.ReadSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("ReadSession err=%v, want=%v", err, nil)
	}
	if got, want := session.UserID, originalUserID; got != want {
		t.Errorf("session.UserID=%v, want=%v", got, want)
	}
	if got, want := session.CreatedAt, originalSession.CreatedAt; !got.Equal(want) {
		t.Errorf("session.CreatedAt=%v, want=%v", got, want)
	}
	if got, want := session.ExpiresAt, originalSession.ExpiresAt; !got.Equal(want) {
		t.Errorf("session.ExpiresAt=%v, want=%v", got, want)
	}
}

func mustSessionID(t *testing.T) simple_sessions.ID {
	t.Helper()
	sessionID, err := simple_sessions.NewID(
		base64.RawURLEncoding.EncodeToString(make([]byte, 32)),
	)
	if err != nil {
		t.Fatalf("NewID err=%v, want=%v", err, nil)
	}
	return sessionID
}

func mustUserID(t *testing.T, raw string) simple_sessions.UserID {
	t.Helper()
	userID, err := simple_sessions.NewUserID(raw)
	if err != nil {
		t.Fatalf("NewUserID err=%v, want=%v", err, nil)
	}
	return userID
}
