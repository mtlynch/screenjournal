package handlers_test

import (
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

type passwordResetTestData struct {
	users struct {
		userA screenjournal.User
		userB screenjournal.User
	}
	passwordResetEntries struct {
		validToken   screenjournal.PasswordResetEntry
		expiredToken screenjournal.PasswordResetEntry
	}
}

func makePasswordResetTestData() passwordResetTestData {
	td := passwordResetTestData{}

	td.users.userA = screenjournal.User{
		Username:     screenjournal.Username("userA"),
		PasswordHash: mustCreatePasswordHash("oldpass123"),
		Email:        screenjournal.Email("userA@example.com"),
		IsAdmin:      false,
	}

	td.users.userB = screenjournal.User{
		Username:     screenjournal.Username("userB"),
		PasswordHash: mustCreatePasswordHash("userBpass456"),
		Email:        screenjournal.Email("userB@example.com"),
		IsAdmin:      true,
	}

	td.passwordResetEntries.validToken = screenjournal.PasswordResetEntry{
		Username:  td.users.userA.Username,
		Token:     screenjournal.PasswordResetToken("valid-token-123"),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days from now
	}

	td.passwordResetEntries.expiredToken = screenjournal.PasswordResetEntry{
		Username:  td.users.userB.Username,
		Token:     screenjournal.PasswordResetToken("expired-token-456"),
		ExpiresAt: time.Now().Add(-1 * time.Hour), // 1 hour ago
	}

	return td
}

func TestAccountPasswordResetPut(t *testing.T) {
	for _, tt := range []struct {
		description          string
		payload              string
		token                screenjournal.PasswordResetToken
		existingUsers        []screenjournal.User
		existingTokens       []screenjournal.PasswordResetEntry
		expectedStatus       int
		expectSessionFor     screenjournal.Username
		expectSessionIsAdmin bool
	}{
		{
			description: "valid password reset creates session and logs user in",
			payload:     "password=newpass123",
			token:       makePasswordResetTestData().passwordResetEntries.validToken.Token,
			existingUsers: []screenjournal.User{
				makePasswordResetTestData().users.userA,
			},
			existingTokens: []screenjournal.PasswordResetEntry{
				makePasswordResetTestData().passwordResetEntries.validToken,
			},
			expectedStatus:       http.StatusOK,
			expectSessionFor:     makePasswordResetTestData().users.userA.Username,
			expectSessionIsAdmin: makePasswordResetTestData().users.userA.IsAdmin,
		},
		{
			description: "admin user password reset creates admin session",
			payload:     "password=newadminpass789",
			token:       screenjournal.PasswordResetToken("admin-token-999"),
			existingUsers: []screenjournal.User{
				makePasswordResetTestData().users.userB,
			},
			existingTokens: []screenjournal.PasswordResetEntry{
				{
					Username:  makePasswordResetTestData().users.userB.Username,
					Token:     screenjournal.PasswordResetToken("admin-token-999"),
					ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
				},
			},
			expectedStatus:       http.StatusOK,
			expectSessionFor:     makePasswordResetTestData().users.userB.Username,
			expectSessionIsAdmin: makePasswordResetTestData().users.userB.IsAdmin,
		},
		{
			description: "expired token is rejected and no session created",
			payload:     "password=newpass123",
			token:       makePasswordResetTestData().passwordResetEntries.expiredToken.Token,
			existingUsers: []screenjournal.User{
				makePasswordResetTestData().users.userB,
			},
			existingTokens: []screenjournal.PasswordResetEntry{
				makePasswordResetTestData().passwordResetEntries.expiredToken,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			description: "nonexistent token is rejected and no session created",
			payload:     "password=newpass123",
			token:       screenjournal.PasswordResetToken("nonexistent-token"),
			existingUsers: []screenjournal.User{
				makePasswordResetTestData().users.userA,
			},
			existingTokens: []screenjournal.PasswordResetEntry{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			description: "invalid password is rejected and no session created",
			payload:     "password=short",
			token:       makePasswordResetTestData().passwordResetEntries.validToken.Token,
			existingUsers: []screenjournal.User{
				makePasswordResetTestData().users.userA,
			},
			existingTokens: []screenjournal.PasswordResetEntry{
				makePasswordResetTestData().passwordResetEntries.validToken,
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
			sessionManager := newMockSessionManager([]mockSessionEntry{})

			s := handlers.New(authenticator, nilAnnouncer, &sessionManager, dataStore, nilMetadataFinder)

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
				if got, want := len(sessionManager.sessions), 0; got != want {
					t.Errorf("sessionCount=%d, want=%d", got, want)
				}
				return
			}

			// Verify that a session was created.
			if got, want := len(sessionManager.sessions), 1; got != want {
				t.Fatalf("sessionCount=%d, want=%d", got, want)
			}

			// Find the created session.
			var createdSession sessions.Session
			for _, session := range sessionManager.sessions {
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

			// Verify that the password was actually changed.
			// Extract new password from payload.
			var newPassword screenjournal.Password
			if strings.Contains(tt.payload, "password=newpass123") {
				newPassword = screenjournal.Password("newpass123")
			} else if strings.Contains(tt.payload, "password=newadminpass789") {
				newPassword = screenjournal.Password("newadminpass789")
			}

			if err := authenticator.Authenticate(tt.expectSessionFor, newPassword); err != nil {
				t.Errorf("new password (%s) is not valid after reset", newPassword.String())
			}

			// Verify that the password reset token was deleted.
			_, err = dataStore.ReadPasswordResetEntry(tt.token)
			if err == nil {
				t.Error("password reset token should be deleted after use")
			}
		})
	}
}
