package simple

import (
	"github.com/mtlynch/screenjournal/v2/auth"
)

type (
	AuthStore interface {
		ReadPasswordHash(username string) (auth.PasswordHash, error)
	}

	authenticator struct {
		store AuthStore
	}
)

func New(store AuthStore) auth.Authenticator {
	return authenticator{
		store: store,
	}
}

func (a authenticator) Authenticate(username, password string) error {
	h, err := a.store.ReadPasswordHash(username)
	if err != nil {
		return err
	}

	if ok := h.MatchesPlaintext(password); !ok {
		return auth.ErrIncorrectPassword
	}

	return nil
}
