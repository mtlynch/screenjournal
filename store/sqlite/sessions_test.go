package sqlite_test

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	simple_sessions "codeberg.org/mtlynch/simpleauth/v3/sessions"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store/sqlite"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

func TestDeleteExpiredSessions(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, time.May, 2, 12, 0, 0, 0, time.UTC)

	for _, tt := range []struct {
		description   string
		expiresAt     time.Time
		sessionExists bool
	}{
		{
			description:   "cleanup deletes a session that expired before now",
			expiresAt:     now.Add(-24 * time.Hour),
			sessionExists: false,
		},
		{
			description:   "cleanup deletes a session that expires at now",
			expiresAt:     now,
			sessionExists: false,
		},
		{
			description:   "cleanup preserves a session that expires after now",
			expiresAt:     now.Add(24 * time.Hour),
			sessionExists: true,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			store := test_sqlite.New()
			username := "session-user"
			session := simple_sessions.Session{
				ID:        newSessionID(t, username),
				UserID:    newUserID(t, username),
				CreatedAt: now.Add(-48 * time.Hour),
				ExpiresAt: tt.expiresAt,
			}
			insertUser(t, store, session.UserID.String())
			if err := store.CreateSession(ctx, session); err != nil {
				t.Fatalf("CreateSession err=%v, want=%v", err, nil)
			}

			if err := store.DeleteExpiredSessions(ctx, now); err != nil {
				t.Fatalf("DeleteExpiredSessions err=%v, want=%v", err, nil)
			}

			readSession, exists := readSessionIfExists(t, store, ctx, session)
			if got, want := exists, tt.sessionExists; got != want {
				t.Fatalf("session exists=%v, want=%v", got, want)
			}
			if !tt.sessionExists {
				if err := store.CreateSession(ctx, session); err != nil {
					t.Fatalf("CreateSession after cleanup err=%v, want=%v", err, nil)
				}
				return
			}
			if got, want := readSession, session; got != want {
				t.Fatalf("session=%v, want=%v", got, want)
			}
		})
	}
}

func TestReadSession(t *testing.T) {
	ctx := context.Background()
	createdAt := time.Date(2026, time.May, 2, 12, 0, 0, 0, time.UTC)

	for _, tt := range []struct {
		description     string
		expiresAt       time.Time
		sessionExpected bool
	}{
		{
			description:     "reads a session that expires in the future",
			expiresAt:       time.Date(2100, time.January, 1, 0, 0, 0, 0, time.UTC),
			sessionExpected: true,
		},
		{
			description:     "does not read a session that expired in the past",
			expiresAt:       time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
			sessionExpected: false,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			store := test_sqlite.New()
			username := "session-user"
			session := simple_sessions.Session{
				ID:        newSessionID(t, username),
				UserID:    newUserID(t, username),
				CreatedAt: createdAt,
				ExpiresAt: tt.expiresAt,
			}
			insertUser(t, store, session.UserID.String())
			if err := store.CreateSession(ctx, session); err != nil {
				t.Fatalf("CreateSession err=%v, want=%v", err, nil)
			}

			readSession, err := store.ReadSession(ctx, session.ID)
			if !tt.sessionExpected {
				if got, want := err, simple_sessions.ErrNoSessionFound; !errors.Is(got, want) {
					t.Fatalf("err=%v, want=%v", got, want)
				}
				return
			}
			if got, want := err, error(nil); got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := readSession, session; got != want {
				t.Fatalf("session=%v, want=%v", got, want)
			}
		})
	}
}

func newSessionID(t *testing.T, username string) simple_sessions.ID {
	t.Helper()
	id, err := simple_sessions.NewID(
		base64.RawURLEncoding.EncodeToString(
			[]byte(username + "-sessionID-sessionID-sessionID"),
		),
	)
	if err != nil {
		t.Fatalf("NewID err=%v, want=%v", err, nil)
	}
	return id
}

func newUserID(t *testing.T, username string) simple_sessions.UserID {
	t.Helper()
	userID, err := simple_sessions.NewUserID(username)
	if err != nil {
		t.Fatalf("NewUserID err=%v, want=%v", err, nil)
	}
	return userID
}

func insertUser(t *testing.T, store sqlite.Store, username string) {
	t.Helper()
	user := screenjournal.User{
		Username:     screenjournal.Username(username),
		Email:        screenjournal.Email(username + "@example.com"),
		PasswordHash: screenjournal.PasswordHash("dummy-password-hash"),
	}
	if err := store.InsertUser(user); err != nil {
		t.Fatalf("InsertUser err=%v, want=%v", err, nil)
	}
}

func readSessionIfExists(
	t *testing.T,
	store sqlite.Store,
	ctx context.Context,
	expectedSession simple_sessions.Session,
) (simple_sessions.Session, bool) {
	t.Helper()
	session, err := store.ReadSession(ctx, expectedSession.ID)
	if errors.Is(err, simple_sessions.ErrNoSessionFound) {
		return simple_sessions.Session{}, false
	}
	if err != nil {
		t.Fatalf("ReadSession err=%v, want=%v", err, nil)
	}
	return session, true
}
