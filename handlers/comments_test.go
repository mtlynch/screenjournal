package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/auth/simple"
	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
	"github.com/mtlynch/screenjournal/v2/metadata"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

func TestCommentsPost(t *testing.T) {
	for _, tt := range []struct {
		description      string
		route            string
		payload          string
		sessionToken     string
		sessions         []mockSession
		expectedComments screenjournal.ReviewComment
		status           int
	}{
		{
			description: "allows user to comment on an existing review",
			route:       "/api/reviews/1/comment",
			payload: `{
					"comment": "Good insights!"
				}`,
			sessionToken: "abc123",
			sessions: []mockSession{
				{
					token: "abc123",
					session: sessions.Session{
						User: screenjournal.User{
							Username: screenjournal.Username("userA"),
						},
					},
				},
			},
			expectedComments: screenjournal.ReviewComment{
				Comment: screenjournal.Comment("Good insights!"),
			},
			status: http.StatusOK,
		},
		// TODO: Test comment on non-existent review
		{
			description: "rejects comment update if user is not authenticated",
			route:       "/api/reviews/1/comment",
			payload: `{
					"comment": "I haven't logged in, but I'm commenting anyway!"
				}`,
			sessionToken: "dummy-invalid-token",
			sessions:     []mockSession{},
			status:       http.StatusUnauthorized,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()

			// Populate datastore with dummy users.
			for _, s := range tt.sessions {
				dataStore.InsertUser(s.session.User)
			}

			authenticator := simple.New(dataStore)
			var nilMetadataFinder metadata.Finder
			sessionManager := newMockSessionManager(tt.sessions)

			s := handlers.New(authenticator, nilAnnouncer, &sessionManager, dataStore, nilMetadataFinder)

			req, err := http.NewRequest("POST", tt.route, strings.NewReader(tt.payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Add("Content-Type", "text/json")
			req.AddCookie(&http.Cookie{
				Name:  "token",
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

			/*prefs, err := dataStore.ReadComments(tt.sessions[0].session.User.Username)
			if err != nil {
				t.Fatalf("failed to read notification preferences from datastore for %s: %v", tt.sessions[0].session.User.Username, err)
			}
			if got, want := prefs, tt.expectedPrefs; got != want {
				t.Errorf("notificationPreferences=%+v, got=%+v", got, want)
			}*/
		})
	}
}
