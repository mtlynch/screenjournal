package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

type invitesTestData struct {
	sessions struct {
		adminUser   mockSessionEntry
		regularUser mockSessionEntry
	}
}

func makeInvitesTestData() invitesTestData {
	td := invitesTestData{}
	td.sessions.adminUser = newMockSessionEntry("admintok555", screenjournal.Username("admin"))
	td.sessions.regularUser = newMockSessionEntry("abc123", screenjournal.Username("regularUser"))
	return td
}

func TestInvitesPost(t *testing.T) {
	for _, tt := range []struct {
		description     string
		payload         string
		sessionToken    string
		sessions        []mockSessionEntry
		status          int
		expectedInvitee screenjournal.Invitee
	}{
		{
			description:  "creates a new invite successfully",
			payload:      "invitee=Frank",
			sessionToken: makeInvitesTestData().sessions.adminUser.token,
			sessions: []mockSessionEntry{
				makeInvitesTestData().sessions.adminUser,
				makeInvitesTestData().sessions.regularUser,
			},
			status:          http.StatusOK,
			expectedInvitee: screenjournal.Invitee("Frank"),
		},
		{
			description:  "rejects request with missing invitee field",
			payload:      "banana=true",
			sessionToken: makeInvitesTestData().sessions.adminUser.token,
			sessions: []mockSessionEntry{
				makeInvitesTestData().sessions.adminUser,
				makeInvitesTestData().sessions.regularUser,
			},
			status: http.StatusBadRequest,
		},
		{
			description:  "rejects invite creation if requesting user is not admin",
			payload:      "invitee=Frank",
			sessionToken: "dummy-invalid-token",
			sessions: []mockSessionEntry{
				makeInvitesTestData().sessions.adminUser,
				makeInvitesTestData().sessions.regularUser,
			},
			status: http.StatusUnauthorized,
		},
		{
			description:  "rejects invite creation if user is not authenticated",
			payload:      "invitee=Frank",
			sessionToken: "dummy-invalid-token",
			sessions: []mockSessionEntry{
				makeInvitesTestData().sessions.adminUser,
				makeInvitesTestData().sessions.regularUser,
			},
			status: http.StatusUnauthorized,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()

			insertMockUsersForSessions(
				t,
				dataStore,
				tt.sessions,
				screenjournal.Username("admin"),
			)

			authenticator := auth.New(dataStore)
			sessionManager := newMockSessionManager(tt.sessions)
			s := handlers.New(handlers.ServerParams{
				Authenticator:  authenticator,
				SessionManager: &sessionManager,
				Store:          dataStore,
			})

			req, err := http.NewRequest("POST", "/admin/invites", strings.NewReader(tt.payload))
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

			invites, err := dataStore.ReadSignupInvitations()
			if err != nil {
				t.Fatalf("failed to read invites from datastore: %v", err)
			}
			if len(invites) != 1 {
				t.Fatalf("expected 1 invite, got %d", len(invites))
			}

			if got, want := invites[0].Invitee, tt.expectedInvitee; got != want {
				t.Errorf("invitee=%+v, want=%+v", got, want)
			}
		})
	}
}
