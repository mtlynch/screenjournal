package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/passwordreset"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

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
		username             screenjournal.Username
		token                screenjournal.PasswordResetToken
		existingUsers        []screenjournal.User
		existingTokens       []screenjournal.PasswordResetEntry
		expectedStatus       int
		expectSessionFor     screenjournal.Username
		newPasswordInPayload screenjournal.Password
	}{
		{
			description: "valid password reset creates session and logs user in",
			payload:     "password=newpass123",
			username:    screenjournal.Username("userA"),
			token:       screenjournal.NewPasswordResetTokenFromString("ABCDEFGHJKLMNPQRSTUVWXYZabcdef23"),
			existingUsers: []screenjournal.User{
				{
					Username:     screenjournal.Username("userA"),
					PasswordHash: createPasswordHash("oldpass123"),
					Email:        screenjournal.Email("userA@example.com"),
				},
			},
			existingTokens: []screenjournal.PasswordResetEntry{
				{
					Username:  screenjournal.Username("userA"),
					Token:     screenjournal.NewPasswordResetTokenFromString("ABCDEFGHJKLMNPQRSTUVWXYZabcdef23"),
					ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days from now.
				},
			},
			expectedStatus:       http.StatusOK,
			expectSessionFor:     screenjournal.Username("userA"),
			newPasswordInPayload: screenjournal.Password("newpass123"),
		},
		{
			description: "expired token is rejected and no session created",
			payload:     "password=newpass123",
			username:    screenjournal.Username("userB"),
			token:       screenjournal.NewPasswordResetTokenFromString("ABCDEFGHJKLMNPQRSTUVWXYZabcdef56"),
			existingUsers: []screenjournal.User{
				{
					Username:     screenjournal.Username("userB"),
					PasswordHash: createPasswordHash("userBpass456"),
					Email:        screenjournal.Email("userB@example.com"),
				},
			},
			existingTokens: []screenjournal.PasswordResetEntry{
				{
					Username:  screenjournal.Username("userB"),
					Token:     screenjournal.NewPasswordResetTokenFromString("ABCDEFGHJKLMNPQRSTUVWXYZabcdef56"),
					ExpiresAt: time.Now().Add(-1 * time.Hour), // 1 hour ago.
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			description: "nonexistent token is rejected and no session created",
			payload:     "password=newpass123",
			username:    screenjournal.Username("userA"),
			token:       screenjournal.NewPasswordResetTokenFromString("ABCDEFGHJKLMNPQRSTUVWXYZabcdefNE"),
			existingUsers: []screenjournal.User{
				{
					Username:     screenjournal.Username("userA"),
					PasswordHash: createPasswordHash("oldpass123"),
					Email:        screenjournal.Email("userA@example.com"),
				},
			},
			existingTokens: []screenjournal.PasswordResetEntry{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			description: "invalid password is rejected and no session created",
			payload:     "password=short",
			username:    screenjournal.Username("userA"),
			token:       screenjournal.NewPasswordResetTokenFromString("ABCDEFGHJKLMNPQRSTUVWXYZabcdef12"),
			existingUsers: []screenjournal.User{
				{
					Username:     screenjournal.Username("userA"),
					PasswordHash: createPasswordHash("oldpass123"),
					Email:        screenjournal.Email("userA@example.com"),
				},
			},
			existingTokens: []screenjournal.PasswordResetEntry{
				{
					Username:  screenjournal.Username("userA"),
					Token:     screenjournal.NewPasswordResetTokenFromString("ABCDEFGHJKLMNPQRSTUVWXYZabcdef12"),
					ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days from now.
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			description: "valid token for different user is rejected",
			payload:     "password=newpass123",
			username:    screenjournal.Username("userB"),
			token:       screenjournal.NewPasswordResetTokenFromString("ABCDEFGHJKLMNPQRSTUVWXYZabcdef23"),
			existingUsers: []screenjournal.User{
				{
					Username:     screenjournal.Username("userA"),
					PasswordHash: createPasswordHash("oldpass123"),
					Email:        screenjournal.Email("userA@example.com"),
				},
				{
					Username:     screenjournal.Username("userB"),
					PasswordHash: createPasswordHash("userBpass456"),
					Email:        screenjournal.Email("userB@example.com"),
				},
			},
			existingTokens: []screenjournal.PasswordResetEntry{
				{
					Username:  screenjournal.Username("userA"),
					Token:     screenjournal.NewPasswordResetTokenFromString("ABCDEFGHJKLMNPQRSTUVWXYZabcdef23"),
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

			sessionManager := newMockSessionManager([]mockSessionEntry{})

			passwordResetter := passwordreset.NewNoEmail(dataStore, time.Now)

			s := handlers.New(handlers.ServerParams{
				Authenticator:    authenticator,
				SessionManager:   &sessionManager,
				Store:            dataStore,
				PasswordResetter: passwordResetter,
			})

			url := "/account/password-reset?username=" + tt.username.String() + "&token=" + tt.token.String()
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
			var createdSession mockSession
			for _, session := range sessionManager.sessions {
				createdSession = session
				break
			}

			// Verify the session is for the correct user.
			if got, want := createdSession.Username, tt.expectSessionFor; !got.Equal(want) {
				t.Errorf("sessionUsername=%+v, want=%+v", got, want)
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
