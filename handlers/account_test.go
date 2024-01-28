package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/metadata"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

type mockUserEntry struct {
	sessionToken string
	username     screenjournal.Username
	password     screenjournal.Password
}

func TestAccountChangePasswordPost(t *testing.T) {
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
			description: "valid request changes password",
			payload: `{
					"oldPassword":"oldpass123",
					"newPassword":"newpass456"
				}`,
			sessionToken:     "abc123",
			expectedStatus:   http.StatusOK,
			expectedPassword: screenjournal.Password("newpass456"),
		},
		{
			description: "reject password change if old password is incorrect",
			payload: `{
					"oldPassword":"wrongpass",
					"newPassword":"newpass456"
				}`,
			sessionToken:     "abc123",
			expectedStatus:   http.StatusUnauthorized,
			expectedPassword: screenjournal.Password("oldpass123"),
		},
		{
			description: "reject password change if new password doesn't meet requirements",
			payload: `{
					"oldPassword":"oldpass123",
					"newPassword":"pass"
				}`,
			sessionToken:     "abc123",
			expectedStatus:   http.StatusBadRequest,
			expectedPassword: screenjournal.Password("oldpass123"),
		},
		{
			description: "invalid JSON does not change password",
			payload: `{
					"oldPassword":`,
			sessionToken:     "abc123",
			expectedStatus:   http.StatusBadRequest,
			expectedPassword: screenjournal.Password("newpass456"),
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()

			mockSessionEntries := []mockSessionEntry{}

			// Populate datastore and session manager with dummy users.
			for _, entry := range userEntries {
				dataStore.InsertUser(
					screenjournal.User{
						Username:     entry.username,
						PasswordHash: mustCreatePasswordHash(entry.password.String()),
					})
				mockSessionEntries = append(mockSessionEntries, mockSessionEntry{
					token: entry.sessionToken,
					session: handlers.Session{
						Username: entry.username,
					},
				})
			}
			authenticator := auth.New(dataStore)
			sessionManager := newMockSessionManager(mockSessionEntries)

			var nilMetadataFinder metadata.Finder
			s := handlers.New(authenticator, nilAnnouncer, &sessionManager, dataStore, nilMetadataFinder)

			req, err := http.NewRequest("POST", "/api/account/change-password", strings.NewReader(tt.payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Add("Content-Type", "text/json")
			req.AddCookie(&http.Cookie{
				Name:  mockSessionTokenName,
				Value: tt.sessionToken,
			})

			w := httptest.NewRecorder()
			s.Router().ServeHTTP(w, req)

			if got, want := w.Code, tt.expectedStatus; got != want {
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

func TestAccountNotificationsPost(t *testing.T) {
	for _, tt := range []struct {
		description   string
		payload       string
		sessionToken  string
		sessions      []mockSessionEntry
		expectedPrefs screenjournal.NotificationPreferences
		status        int
	}{
		{
			description: "allows user to subscribe to new reviews and comments",
			payload: `{
					"isSubscribedToNewReviews":true,
					"isSubscribedToAllComments":true
				}`,
			sessionToken: "abc123",
			sessions: []mockSessionEntry{
				{
					token: "abc123",
					session: handlers.Session{
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
			description: "allows user to unsubscribe to new reviews but subscribe to comments",
			payload: `{
							"isSubscribedToNewReviews":false,
							"isSubscribedToAllComments":true
						}`,
			sessionToken: "abc123",
			sessions: []mockSessionEntry{
				{
					token: "abc123",
					session: handlers.Session{
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
			description: "allows user to subscribe to new reviews but unsubscribe from comments",
			payload: `{
							"isSubscribedToNewReviews":true,
							"isSubscribedToAllComments":false
						}`,
			sessionToken: "abc123",
			sessions: []mockSessionEntry{
				{
					token: "abc123",
					session: handlers.Session{
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
			description: "rejects non-bool value for review subscription status",
			payload: `{
							"isSubscribedToNewReviews":"banana",
							"isSubscribedToAllComments":true
						}`,
			sessionToken: "abc123",
			sessions: []mockSessionEntry{
				{
					token: "abc123",
					session: handlers.Session{
						Username: screenjournal.Username("userA"),
					},
				},
			},
			status: http.StatusBadRequest,
		},
		{
			description: "rejects subscription update if user is not authenticated",
			payload: `{
							"isSubscribedToNewReviews":true,
							"isSubscribedToAllComments":true
						}`,
			sessionToken: "dummy-invalid-token",
			sessions:     []mockSessionEntry{},
			status:       http.StatusUnauthorized,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()

			// Populate datastore with dummy users.
			for _, s := range tt.sessions {
				dataStore.InsertUser(
					screenjournal.User{
						Username: s.session.Username,
						IsAdmin:  s.session.IsAdmin,
					})
			}

			authenticator := auth.New(dataStore)
			var nilMetadataFinder metadata.Finder
			sessionManager := newMockSessionManager(tt.sessions)

			s := handlers.New(authenticator, nilAnnouncer, &sessionManager, dataStore, nilMetadataFinder)

			req, err := http.NewRequest("POST", "/api/account/notifications", strings.NewReader(tt.payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Add("Content-Type", "text/json")
			req.AddCookie(&http.Cookie{
				Name:  mockSessionTokenName,
				Value: tt.sessionToken,
			})

			w := httptest.NewRecorder()
			s.Router().ServeHTTP(w, req)

			if got, want := w.Code, tt.status; got != want {
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
