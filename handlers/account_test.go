package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

type mockUserEntry struct {
	sessionToken string
	username     screenjournal.Username
	password     screenjournal.Password
}

var nilMetadataFinder handlers.MetadataFinder

func TestAccountChangePasswordPut(t *testing.T) {
	userEntries := []mockUserEntry{
		{
			sessionToken: "abc123",
			username:     screenjournal.Username("userA"),
			password:     screenjournal.Password("oldpass123"),
		},
	}

	for _, tt := range []struct {
		description      string
		payload          string
		sessionToken     string
		expectedStatus   int
		expectedPassword screenjournal.Password
	}{
		{
			description:      "valid request changes password",
			payload:          "old-password=oldpass123&password=newpass456",
			sessionToken:     "abc123",
			expectedStatus:   http.StatusOK,
			expectedPassword: screenjournal.Password("newpass456"),
		},
		{
			description:      "reject password change if old password is incorrect",
			payload:          "old-password=wrongpass&password=newpass456",
			sessionToken:     "abc123",
			expectedStatus:   http.StatusUnauthorized,
			expectedPassword: screenjournal.Password("oldpass123"),
		},
		{
			description:      "reject password change if old password matches new password",
			payload:          "old-password=oldpass123&password=oldpass123",
			sessionToken:     "abc123",
			expectedStatus:   http.StatusBadRequest,
			expectedPassword: screenjournal.Password("oldpass123"),
		},
		{
			description:      "reject password change if new password doesn't meet requirements",
			payload:          "old-password=oldpass123&password=pass",
			sessionToken:     "abc123",
			expectedStatus:   http.StatusBadRequest,
			expectedPassword: screenjournal.Password("oldpass123"),
		},
		{
			description:      "missing parameters do not change password",
			payload:          "password=newpass456",
			sessionToken:     "abc123",
			expectedStatus:   http.StatusBadRequest,
			expectedPassword: screenjournal.Password("oldpass123"),
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()

			mockSessionEntries := []mockSessionEntry{}

			// Populate datastore and session manager with dummy users.
			for _, entry := range userEntries {
				mockUser := screenjournal.User{
					Username:     entry.username,
					PasswordHash: mustCreatePasswordHash(entry.password.String()),
					Email:        screenjournal.Email(entry.username + "@example.com"),
				}
				if err := dataStore.InsertUser(mockUser); err != nil {
					t.Fatalf("failed to insert mock user %+v: %v", mockUser, err)
				}
				mockSessionEntries = append(mockSessionEntries, mockSessionEntry{
					token: entry.sessionToken,
					session: sessions.Session{
						Username: entry.username,
					},
				})
			}
			authenticator := auth.New(dataStore)
			sessionManager := newMockSessionManager(mockSessionEntries)
			s := handlers.New(authenticator, nilAnnouncer, &sessionManager, dataStore, nilMetadataFinder)

			req, err := http.NewRequest("PUT", "/account/password", strings.NewReader(tt.payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.AddCookie(&http.Cookie{
				Name:  mockSessionTokenName,
				Value: tt.sessionToken,
			})

			rec := httptest.NewRecorder()
			s.Router().ServeHTTP(rec, req)
			res := rec.Result()

			if got, want := res.StatusCode, tt.expectedStatus; got != want {
				t.Fatalf("httpStatus=%v, want=%v", got, want)
			}

			session, err := sessionManager.SessionFromToken(tt.sessionToken)
			if err != nil {
				t.Fatalf("couldn't map session token (%s) to session: %v", tt.sessionToken, err)
			}

			if err := authenticator.Authenticate(session.Username, tt.expectedPassword); err != nil {
				t.Errorf("expected password (%s) is not valid after request", tt.expectedPassword.String())
			}
		})
	}
}

func TestAccountNotificationsPut(t *testing.T) {
	for _, tt := range []struct {
		description   string
		payload       string
		sessionToken  string
		sessions      []mockSessionEntry
		expectedPrefs screenjournal.NotificationPreferences
		status        int
	}{
		{
			description:  "allows user to subscribe to new reviews and comments",
			payload:      "new-reviews=on&all-comments=on",
			sessionToken: "abc123",
			sessions: []mockSessionEntry{
				{
					token: "abc123",
					session: sessions.Session{
						Username: screenjournal.Username("userA"),
					},
				},
			},
			expectedPrefs: screenjournal.NotificationPreferences{
				NewReviews:     true,
				AllNewComments: true,
			},
			status: http.StatusOK,
		},
		{
			description:  "allows user to unsubscribe to new reviews but subscribe to comments",
			payload:      "all-comments=on",
			sessionToken: "abc123",
			sessions: []mockSessionEntry{
				{
					token: "abc123",
					session: sessions.Session{
						Username: screenjournal.Username("userA"),
					},
				},
			},
			expectedPrefs: screenjournal.NotificationPreferences{
				NewReviews:     false,
				AllNewComments: true,
			},
			status: http.StatusOK,
		},
		{
			description:  "allows user to subscribe to new reviews but unsubscribe from comments",
			payload:      "new-reviews=on",
			sessionToken: "abc123",
			sessions: []mockSessionEntry{
				{
					token: "abc123",
					session: sessions.Session{
						Username: screenjournal.Username("userA"),
					},
				},
			},
			expectedPrefs: screenjournal.NotificationPreferences{
				NewReviews:     true,
				AllNewComments: false,
			},
			status: http.StatusOK,
		},
		{
			description:  "rejects subscription update if user is not authenticated",
			payload:      "new-reviews=on&all-comments=on",
			sessionToken: "dummy-invalid-token",
			sessions:     []mockSessionEntry{},
			status:       http.StatusUnauthorized,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()

			// Populate datastore with dummy users.
			for _, s := range tt.sessions {
				mockUser := screenjournal.User{
					Username:     s.session.Username,
					IsAdmin:      s.session.IsAdmin,
					Email:        screenjournal.Email(s.session.Username + "@example.com"),
					PasswordHash: screenjournal.PasswordHash("dummy-pw-hash"),
				}
				if err := dataStore.InsertUser(mockUser); err != nil {
					t.Fatalf("failed to insert mock user %+v: %v", mockUser, err)
				}
			}

			authenticator := auth.New(dataStore)
			sessionManager := newMockSessionManager(tt.sessions)

			s := handlers.New(authenticator, nilAnnouncer, &sessionManager, dataStore, nilMetadataFinder)

			req, err := http.NewRequest("PUT", "/account/notifications", strings.NewReader(tt.payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.AddCookie(&http.Cookie{
				Name:  mockSessionTokenName,
				Value: tt.sessionToken,
			})

			rec := httptest.NewRecorder()
			s.Router().ServeHTTP(rec, req)
			res := rec.Result()

			if got, want := res.StatusCode, tt.status; got != want {
				t.Fatalf("httpStatus=%v, want=%v", got, want)
			}

			if tt.status != http.StatusOK {
				return
			}

			prefs, err := dataStore.ReadNotificationPreferences(tt.sessions[0].session.Username)
			if err != nil {
				t.Fatalf("failed to read notification preferences from datastore for %s: %v", tt.sessions[0].session.Username, err)
			}
			if got, want := prefs, tt.expectedPrefs; got != want {
				t.Errorf("notificationPreferences=%+v, got=%+v", got, want)
			}
		})
	}
}
