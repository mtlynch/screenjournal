package simple

import (
	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/store"
)

type (
	authenticator struct {
		store store.Store
	}
)

func New(store store.Store) (auth.Authenticator, error) {
	return authenticator{
		store: store,
	}, nil
}

func (a authenticator) Authenticate(username screenjournal.Username, password screenjournal.Password) (screenjournal.User, error) {
	u, err := a.store.ReadUser(username)
	if err != nil {
		return screenjournal.User{}, err
	}

	if err := u.PasswordHash.MatchesPlaintext(password); err != nil {
		return screenjournal.User{}, err
	}

	return u, nil
}
