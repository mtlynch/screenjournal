package auth

import (
	simple_auth "github.com/mtlynch/simpleauth/v2/auth"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type (
	GenericAuthenticator interface {
		Authenticate(username, password string) error
	}

	Authenticator struct {
		inner GenericAuthenticator
	}

	UserStore interface {
		ReadUser(screenjournal.Username) (screenjournal.User, error)
	}
)

func New(userStore UserStore) Authenticator {
	return Authenticator{
		inner: simple_auth.New(authStore{
			userStore: userStore,
		}),
	}
}

func (a Authenticator) Authenticate(username screenjournal.Username, password screenjournal.Password) error {
	return a.inner.Authenticate(username.String(), password.String())
}
