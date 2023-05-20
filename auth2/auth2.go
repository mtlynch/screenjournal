package auth2

import (
	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/auth/simple"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/store"
)

type Authenticator interface {
	Authenticate(username, password string) error
}

type authStore struct {
	inner store.Store
}

func New(store store.Store) Authenticator {
	return simple.New(NewAuthStore(store))
}

// TODO: Refactor
func NewAuthStore(s store.Store) simple.AuthStore {
	return authStore{
		inner: s,
	}
}

func (s authStore) ReadPasswordHash(usernameRaw string) (auth.PasswordHash, error) {
	username, err := parse.Username(usernameRaw)
	if err != nil {
		return nil, err
	}

	user, err := s.inner.ReadUser(username)
	if err != nil {
		return nil, err
	}

	return user.PasswordHash, nil
}
