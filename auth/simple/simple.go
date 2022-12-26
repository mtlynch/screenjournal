package simple

import (
	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/auth"
)

type (
	UserStore interface {
		ReadUser(screenjournal.Username) (screenjournal.User, error)
	}

	authenticator struct {
		store UserStore
	}
)

func New(store UserStore) auth.Authenticator {
	return authenticator{
		store: store,
	}
}

func (a authenticator) Authenticate(username screenjournal.Username, password screenjournal.Password) (screenjournal.User, error) {
	u, err := a.store.ReadUser(username)
	if err != nil {
		if err == screenjournal.ErrUserNotFound {
			return screenjournal.User{}, screenjournal.ErrInvalidCredentials
		}
		return screenjournal.User{}, err
	}

	if err := u.PasswordHash.MatchesPlaintext(password); err != nil {
		if err == screenjournal.ErrIncorrectPassword {
			return screenjournal.User{}, screenjournal.ErrInvalidCredentials
		}
		return screenjournal.User{}, err
	}

	return u, nil
}
