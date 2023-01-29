package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/auth/simple"
	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/metadata"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

func TestUsersPut(t *testing.T) {
	for _, tt := range []struct {
		description string
		route       string
		payload     string
		users       []screenjournal.User
		invites     []screenjournal.SignupInvitation
		status      int
	}{
		{
			description: "valid signup succeeds",
			route:       "/api/users/userA",
			payload: `{
					"email": "userA@example.com",
					"password": "dummyp@ss"
				}`,
			users:   []screenjournal.User{},
			invites: []screenjournal.SignupInvitation{},
			status:  http.StatusOK,
		},
		{
			description: "rejects signup of existing username",
			route:       "/api/users/userA",
			payload: `{
					"email": "someuser@example.com",
					"password": "dummyp@ss",
					"inviteCode": "abc456"
				}`,
			users: []screenjournal.User{
				userA,
				userB,
			},
			invites: []screenjournal.SignupInvitation{
				{
					Invitee:    screenjournal.Invitee("Sammy"),
					InviteCode: screenjournal.InviteCode("abc456"),
				},
			},
			status: http.StatusConflict,
		},
		{
			description: "rejects signup with email associated with another user",
			route:       "/api/users/someguy",
			payload: `{
					"email": "userA@example.com",
					"password": "dummyp@ss",
					"inviteCode": "abc456"
				}`,
			users: []screenjournal.User{
				userA,
				userB,
			},
			invites: []screenjournal.SignupInvitation{
				{
					Invitee:    screenjournal.Invitee("Sammy"),
					InviteCode: screenjournal.InviteCode("abc456"),
				},
			},
			status: http.StatusConflict,
		},
		{
			description: "rejects signup with missing invite code",
			route:       "/api/users/sammy123",
			payload: `{
					"email": "sammy123@example.com",
					"password": "dummyp@ss"
				}`,
			users: []screenjournal.User{
				userA,
				userB,
			},
			invites: []screenjournal.SignupInvitation{
				{
					Invitee:    screenjournal.Invitee("Sammy"),
					InviteCode: screenjournal.InviteCode("abc456"),
				},
			},
			status: http.StatusForbidden,
		},
		{
			description: "rejects signup with incorrect invite code",
			route:       "/api/users/sammy123",
			payload: `{
					"email": "sammy123@example.com",
					"password": "dummyp@ss",
					"inviteCode": "abc456"
				}`,
			users: []screenjournal.User{
				userA,
				userB,
			},
			invites: []screenjournal.SignupInvitation{
				{
					Invitee:    screenjournal.Invitee("Sammy"),
					InviteCode: screenjournal.InviteCode("232323"),
				},
			},
			status: http.StatusForbidden,
		},
		{
			description: "rejects invalid username",
			route:       "/api/users/q",
			payload: `{
					"email": "userA@example.com",
					"password": "dummyp@ss"
				}`,
			users:   []screenjournal.User{},
			invites: []screenjournal.SignupInvitation{},
			status:  http.StatusBadRequest,
		},
		{
			description: "rejects invalid email",
			route:       "/api/users/userA",
			payload: `{
					"email": "userA@example@com",
					"password": "dummyp@ss"
				}`,
			users:   []screenjournal.User{},
			invites: []screenjournal.SignupInvitation{},
			status:  http.StatusBadRequest,
		},
		{
			description: "rejects invalid password",
			route:       "/api/users/userA",
			payload: `{
					"email": "userA@example.com",
					"password": "a"
				}`,
			users:   []screenjournal.User{},
			invites: []screenjournal.SignupInvitation{},
			status:  http.StatusBadRequest,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()

			for _, user := range tt.users {
				if err := dataStore.InsertUser(user); err != nil {
					panic(err)
				}
			}
			for _, invite := range tt.invites {
				if err := dataStore.InsertSignupInvitation(invite); err != nil {
					panic(err)
				}
			}

			authenticator := simple.New(dataStore)
			var nilMetadataFinder metadata.Finder
			sessionManager := newMockSessionManager([]mockSession{})

			s := handlers.New(authenticator, nilAnnouncer, &sessionManager, dataStore, nilMetadataFinder)

			req, err := http.NewRequest("PUT", tt.route, strings.NewReader(tt.payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Add("Content-Type", "text/json")

			w := httptest.NewRecorder()
			s.Router().ServeHTTP(w, req)

			if got, want := w.Code, tt.status; got != want {
				t.Fatalf("httpStatus=%v, want=%v", got, want)
			}
		})
	}
}
