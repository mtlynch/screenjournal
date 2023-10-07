package auth

import (
	simple_auth "github.com/mtlynch/screenjournal/v2/auth/simple/auth"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
)

type (
	Authenticator interface {
		Authenticate(username screenjournal.Username, password screenjournal.Password) error
	}

	authenticator struct {
		inner simple_auth.Authenticator
	}

	authStore struct {
		inner store.Store
	}
)

func New(store store.Store) Authenticator {
	return authenticator{
		inner: simple_auth.New(authStore{
			inner: store,
		}),
	}
}

func (a authenticator) Authenticate(username screenjournal.Username, password screenjournal.Password) error {
	return a.inner.Authenticate(username.String(), password.String())
}

func HashPassword(password screenjournal.Password) (screenjournal.PasswordHash, error) {
	h, err := simple_auth.HashPassword(password.String())
	if err != nil {
		return screenjournal.PasswordHash{}, err
	}
	return screenjournal.PasswordHash(h.Bytes()), nil
}

func (s authStore) ReadPasswordHash(usernameRaw string) (simple_auth.PasswordHash, error) {
	username, err := parse.Username(usernameRaw)
	if err != nil {
		return nil, err
	}

	user, err := s.inner.ReadUser(username)
	if err != nil {
		return nil, err
	}

	return simple_auth.PasswordHashFromBytes(user.PasswordHash.Bytes()), nil
}
