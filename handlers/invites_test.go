package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

type invitesTestData struct {
	users struct {
		adminUser   screenjournal.User
		regularUser screenjournal.User
	}
	sessions struct {
		adminUser   mockSessionEntry
		regularUser mockSessionEntry
	}
}

func makeInvitesTestData() invitesTestData {
	td := invitesTestData{}
	td.users.adminUser = screenjournal.User{
		Username: screenjournal.Username("admin"),
	}
	td.sessions.adminUser = mockSessionEntry{
		token: "admintok555",
		session: sessions.Session{
			Username: td.users.adminUser.Username,
			IsAdmin:  true,
		},
	}
	td.users.regularUser = screenjournal.User{
		Username: screenjournal.Username("regularUser"),
	}
	td.sessions.regularUser = mockSessionEntry{
		token: "abc123",
		session: sessions.Session{
			Username: td.users.regularUser.Username,
		},
	}
	return td
}

func TestInvitesPost(t *testing.T) {
	for _, tt := range []struct {
		description    string
		payload        string
		sessionToken   string
		sessions       []mockSessionEntry
		status         int
		expectedInvite screenjournal.SignupInvitation
	}{
		{
			description:  "creates a new invite successfully",
			payload:      "invitee=Frank",
			sessionToken: makeInvitesTestData().sessions.adminUser.token,
			sessions: []mockSessionEntry{
				makeInvitesTestData().sessions.adminUser,
				makeInvitesTestData().sessions.regularUser,
			},
			status: http.StatusOK,
			expectedInvite: screenjournal.SignupInvitation{
				Invitee: screenjournal.Invitee("Frank"),
			},
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
			store := test_sqlite.New()

			for _, s := range tt.sessions {
				store.InsertUser(screenjournal.User{
					Username: s.session.Username,
				})
			}

			authenticator := auth.New(store)
			sessionManager := newMockSessionManager(tt.sessions)
			s := handlers.New(authenticator, nilAnnouncer, &sessionManager, store, nilMetadataFinder)

			req, err := http.NewRequest("POST", "/admin/invites", strings.NewReader(tt.payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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

			invites, err := store.ReadSignupInvitations()
			if err != nil {
				t.Fatalf("failed to read invites from datastore: %v", err)
			}
			if len(invites) != 1 {
				t.Fatalf("expected 1 invite, got %d", len(invites))
			}

			tt.expectedInvite.InviteCode = invites[0].InviteCode // Capture the generated invite code
			if got, want := invites[0], tt.expectedInvite; !reflect.DeepEqual(got, want) {
				t.Errorf("invite=%+v, want=%+v", got, want)
			}
		})
	}
}
