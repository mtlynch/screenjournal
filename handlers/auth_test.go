package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
	"github.com/mtlynch/screenjournal/v2/metadata"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

var (
	userA = screenjournal.User{
		Username:     screenjournal.Username("userA"),
		PasswordHash: mustCreatePasswordHash("dummyp@ss"),
		Email:        screenjournal.Email("userA@example.com"),
		IsAdmin:      true,
	}
	userB = screenjournal.User{
		Username:     screenjournal.Username("userB"),
		PasswordHash: mustCreatePasswordHash("p@ssw0rd123"),
		Email:        screenjournal.Email("userB@example.com"),
		IsAdmin:      false,
	}
)

func TestAuthPost(t *testing.T) {
	for _, tt := range []struct {
		description  string
		payload      string
		users        []screenjournal.User
		status       int
		matchingUser screenjournal.User
	}{
		{
			description: "valid credentials succeed for admin user",
			payload: `{
					"username": "userA",
					"password": "dummyp@ss"
				}`,
			users: []screenjournal.User{
				userA,
				userB,
			},
			status:       http.StatusOK,
			matchingUser: userA,
		},
		{
			description: "valid credentials succeed for non-admin user",
			payload: `{
					"username": "userB",
					"password": "p@ssw0rd123"
				}`,
			users: []screenjournal.User{
				userA,
				userB,
			},
			status:       http.StatusOK,
			matchingUser: userB,
		},
		{
			description: "invalid password fails",
			payload: `{
					"username": "userA",
					"password": "wrongpass"
				}`,
			users: []screenjournal.User{
				userA,
				userB,
			},
			status:       http.StatusUnauthorized,
			matchingUser: screenjournal.User{},
		},
		{
			description: "invalid username fails",
			payload: `{
					"username": "nouserlikeme",
					"password": "wrongpass"
				}`,
			users: []screenjournal.User{
				userA,
				userB,
			},
			status:       http.StatusUnauthorized,
			matchingUser: screenjournal.User{},
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()

			for _, user := range tt.users {
				if err := dataStore.InsertUser(user); err != nil {
					panic(err)
				}
			}

			authenticator := auth.New(dataStore)
			var nilMetadataFinder metadata.Finder
			sessionManager := newMockSessionManager([]mockSession{})

			s := handlers.New(authenticator, nilAnnouncer, &sessionManager, dataStore, nilMetadataFinder)

			req, err := http.NewRequest("POST", "/api/auth", strings.NewReader(tt.payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Add("Content-Type", "text/json")

			w := httptest.NewRecorder()
			s.Router().ServeHTTP(w, req)

			if got, want := w.Code, tt.status; got != want {
				t.Fatalf("httpStatus=%v, want=%v", got, want)
			}

			if tt.status != http.StatusOK {
				return
			}

			sessionsCreated := []sessions.Session{}
			for _, session := range sessionManager.sessions {
				sessionsCreated = append(sessionsCreated, session)
			}

			if got, want := len(sessionsCreated), 1; got != want {
				t.Fatalf("count(sessions)=%d, want=%d", got, want)
			}
			if got, want := sessionsCreated[0].User, tt.matchingUser; !reflect.DeepEqual(got, want) {
				t.Errorf("user=%+v, want=%+v", got, want)
			}
		})
	}
}

func mustCreatePasswordHash(plaintext string) screenjournal.PasswordHash {
	h, err := auth.NewPasswordHash(screenjournal.Password(plaintext))
	if err != nil {
		panic(err)
	}
	return screenjournal.PasswordHash(h.Bytes())
}
