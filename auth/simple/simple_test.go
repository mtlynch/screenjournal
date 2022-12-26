package simple_test

import (
	"testing"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/auth/simple"
)

type mockUserStore struct {
	users []screenjournal.User
}

func (us mockUserStore) ReadUser(username screenjournal.Username) (screenjournal.User, error) {
	for _, u := range us.users {
		if u.Username.Equal(username) {
			return u, nil
		}
	}
	return screenjournal.User{}, screenjournal.ErrUserNotFound
}

func TestAuthenticate(t *testing.T) {
	for _, tt := range []struct {
		description string
		store       simple.UserStore
		username    screenjournal.Username
		password    screenjournal.Password
		user        screenjournal.User
		err         error
	}{
		{
			"authenticates when password is valid",
			mockUserStore{
				users: []screenjournal.User{
					{
						Username:     screenjournal.Username("dummyuser"),
						PasswordHash: screenjournal.NewPasswordHash("dummy-p@ssword"),
					},
				},
			},
			screenjournal.Username("dummyuser"),
			screenjournal.Password("dummy-p@ssword"),
			screenjournal.User{
				Username:     screenjournal.Username("dummyuser"),
				PasswordHash: screenjournal.NewPasswordHash("dummy-p@ssword"),
			},
			nil,
		},
		{
			"returns ErrInvalidCredentials when password is invalid",
			mockUserStore{
				users: []screenjournal.User{
					{
						Username:     screenjournal.Username("dummyuser"),
						PasswordHash: screenjournal.NewPasswordHash("dummy-p@ssword"),
					},
				},
			},
			screenjournal.Username("dummyuser"),
			screenjournal.Password("wrongpass"),
			screenjournal.User{},
			screenjournal.ErrInvalidCredentials,
		},
		{
			"returns ErrInvalidCredentials when user is not found",
			mockUserStore{
				users: []screenjournal.User{
					{
						Username:     screenjournal.Username("dummyuser"),
						PasswordHash: screenjournal.NewPasswordHash("dummy-p@ssword"),
					},
				},
			},
			screenjournal.Username("madeupuser"),
			screenjournal.Password("dummy-p@ssword"),
			screenjournal.User{},
			screenjournal.ErrInvalidCredentials,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			authenticator := simple.New(tt.store)
			user, err := authenticator.Authenticate(tt.username, tt.password)
			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}

			if got, want := user.Username.String(), user.Username.String(); got != want {
				t.Errorf("username=%s, want=%s", got, want)
			}
		})
	}
}
