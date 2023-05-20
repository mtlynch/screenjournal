package auth2

import (
	"github.com/mtlynch/screenjournal/v2"
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
	return simple.New(authStore{
		inner: store,
	})
}

func NewPasswordHash(plaintext string) (screenjournal.PasswordHash, error) {
	h, err := auth.NewPasswordHash(plaintext)
	if err != nil {
		return screenjournal.PasswordHash{}, err
	}
	return screenjournal.PasswordHash(h.Bytes()), nil
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

	return auth.NewPasswordHashFromBytes(user.PasswordHash.Bytes()), nil
}
