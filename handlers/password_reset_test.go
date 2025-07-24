package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

// Simple mock session manager implementation for this test.
type passwordResetMockSessionManager struct {
	sessions map[string]sessions.Session
}

func (sm *passwordResetMockSessionManager) CreateSession(w http.ResponseWriter, ctx context.Context, username screenjournal.Username, isAdmin bool) error {
	token := "mock-session-token-12345"
	sm.sessions[token] = sessions.Session{
		Username: username,
		IsAdmin:  isAdmin,
	}
	http.SetCookie(w, &http.Cookie{
		Name:  "mock-session-token",
		Value: token,
	})
	return nil
}

func (sm *passwordResetMockSessionManager) SessionFromContext(ctx context.Context) (sessions.Session, error) {
	// Not used in this test.
	return sessions.Session{}, nil
}

func (sm *passwordResetMockSessionManager) SessionFromToken(token string) (sessions.Session, error) {
	// Not used in this test.
	return sessions.Session{}, nil
}

func (sm *passwordResetMockSessionManager) EndSession(context.Context, http.ResponseWriter) {}

func (sm *passwordResetMockSessionManager) WrapRequest(next http.Handler) http.Handler {
	return next // Simple passthrough for this test.
}

func TestAccountPasswordResetPut(t *testing.T) {
	// Helper to create password hash inline.
	createPasswordHash := func(plaintext string) screenjournal.PasswordHash {
		h, err := auth.HashPassword(screenjournal.Password(plaintext))
		if err != nil {
			t.Fatalf("failed to create password hash: %v", err)
		}
		return screenjournal.PasswordHash(h.Bytes())
	}

	for _, tt := range []struct {
		description          string
		payload              string
		token                screenjournal.PasswordResetToken
		existingUsers        []screenjournal.User
		existingTokens       []screenjournal.PasswordResetEntry
		expectedStatus       int
		expectSessionFor     screenjournal.Username
		expectSessionIsAdmin bool
		newPasswordInPayload screenjournal.Password
	}{
		{
			description: "valid password reset creates session and logs user in",
			payload:     "password=newpass123",
			token:       screenjournal.PasswordResetToken("valid-token-123"),
			existingUsers: []screenjournal.User{
				{
					Username:     screenjournal.Username("userA"),
					PasswordHash: createPasswordHash("oldpass123"),
					Email:        screenjournal.Email("userA@example.com"),
					IsAdmin:      false,
				},
			},
			existingTokens: []screenjournal.PasswordResetEntry{
				{
					Username:  screenjournal.Username("userA"),
					Token:     screenjournal.PasswordResetToken("valid-token-123"),
					ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days from now.
				},
			},
			expectedStatus:       http.StatusOK,
			expectSessionFor:     screenjournal.Username("userA"),
			expectSessionIsAdmin: false,
			newPasswordInPayload: screenjournal.Password("newpass123"),
		},
		{
			description: "admin user password reset creates admin session",
			payload:     "password=newadminpass789",
			token:       screenjournal.PasswordResetToken("admin-token-999"),
			existingUsers: []screenjournal.User{
				{
					Username:     screenjournal.Username("userB"),
					PasswordHash: createPasswordHash("userBpass456"),
					Email:        screenjournal.Email("userB@example.com"),
					IsAdmin:      true,
				},
			},
			existingTokens: []screenjournal.PasswordResetEntry{
				{
					Username:  screenjournal.Username("userB"),
					Token:     screenjournal.PasswordResetToken("admin-token-999"),
					ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days from now.
				},
			},
			expectedStatus:       http.StatusOK,
			expectSessionFor:     screenjournal.Username("userB"),
			expectSessionIsAdmin: true,
			newPasswordInPayload: screenjournal.Password("newadminpass789"),
		},
		{
			description: "expired token is rejected and no session created",
			payload:     "password=newpass123",
			token:       screenjournal.PasswordResetToken("expired-token-456"),
			existingUsers: []screenjournal.User{
				{
					Username:     screenjournal.Username("userB"),
					PasswordHash: createPasswordHash("userBpass456"),
					Email:        screenjournal.Email("userB@example.com"),
					IsAdmin:      true,
				},
			},
			existingTokens: []screenjournal.PasswordResetEntry{
				{
					Username:  screenjournal.Username("userB"),
					Token:     screenjournal.PasswordResetToken("expired-token-456"),
					ExpiresAt: time.Now().Add(-1 * time.Hour), // 1 hour ago.
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			description: "nonexistent token is rejected and no session created",
			payload:     "password=newpass123",
			token:       screenjournal.PasswordResetToken("nonexistent-token"),
			existingUsers: []screenjournal.User{
				{
					Username:     screenjournal.Username("userA"),
					PasswordHash: createPasswordHash("oldpass123"),
					Email:        screenjournal.Email("userA@example.com"),
					IsAdmin:      false,
				},
			},
			existingTokens: []screenjournal.PasswordResetEntry{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			description: "invalid password is rejected and no session created",
			payload:     "password=short",
			token:       screenjournal.PasswordResetToken("valid-token-123"),
			existingUsers: []screenjournal.User{
				{
					Username:     screenjournal.Username("userA"),
					PasswordHash: createPasswordHash("oldpass123"),
					Email:        screenjournal.Email("userA@example.com"),
					IsAdmin:      false,
				},
			},
			existingTokens: []screenjournal.PasswordResetEntry{
				{
					Username:  screenjournal.Username("userA"),
					Token:     screenjournal.PasswordResetToken("valid-token-123"),
					ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days from now.
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()

			// Insert existing users into datastore.
			for _, user := range tt.existingUsers {
				if err := dataStore.InsertUser(user); err != nil {
					t.Fatalf("failed to insert test user %+v: %v", user, err)
				}
			}

			// Insert existing password reset tokens into datastore.
			for _, token := range tt.existingTokens {
				if err := dataStore.InsertPasswordResetEntry(token); err != nil {
					t.Fatalf("failed to insert password reset entry %+v: %v", token, err)
				}
			}

			authenticator := auth.New(dataStore)

			// Create mock session manager for this test.
			sessionMgr := &passwordResetMockSessionManager{
				sessions: make(map[string]sessions.Session),
			}

			// Create handler with nil announcer and metadata finder.
			var nilAnnouncer handlers.Announcer
			var nilMetadataFinder handlers.MetadataFinder

			s := handlers.New(authenticator, nilAnnouncer, sessionMgr, dataStore, nilMetadataFinder)

			url := "/account/password-reset?token=" + tt.token.String()
			req, err := http.NewRequest("PUT", url, strings.NewReader(tt.payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rec := httptest.NewRecorder()
			s.Router().ServeHTTP(rec, req)
			res := rec.Result()

			if got, want := res.StatusCode, tt.expectedStatus; got != want {
				t.Fatalf("httpStatus=%v, want=%v", got, want)
			}

			if tt.expectedStatus != http.StatusOK {
				// Verify no session was created on failure.
				if got, want := len(sessionMgr.sessions), 0; got != want {
					t.Errorf("sessionCount=%d, want=%d", got, want)
				}
				return
			}

			// Verify that a session was created.
			if got, want := len(sessionMgr.sessions), 1; got != want {
				t.Fatalf("sessionCount=%d, want=%d", got, want)
			}

			// Find the created session.
			var createdSession sessions.Session
			for _, session := range sessionMgr.sessions {
				createdSession = session
				break
			}

			// Verify the session is for the correct user.
			if got, want := createdSession.Username, tt.expectSessionFor; !got.Equal(want) {
				t.Errorf("sessionUsername=%+v, want=%+v", got, want)
			}

			// Verify the session has correct admin status.
			if got, want := createdSession.IsAdmin, tt.expectSessionIsAdmin; got != want {
				t.Errorf("sessionIsAdmin=%t, want=%t", got, want)
			}

			// Verify that the password was actually changed by attempting to
			// authenticate with the new password.
			if err := authenticator.Authenticate(tt.expectSessionFor, tt.newPasswordInPayload); err != nil {
				t.Errorf("new password (%s) is not valid after reset", tt.newPasswordInPayload.String())
			}

			// Verify that the password reset token was deleted.
			_, err = dataStore.ReadPasswordResetEntry(tt.token)
			if err == nil {
				t.Error("password reset token should be deleted after use")
			}
		})
	}
}
