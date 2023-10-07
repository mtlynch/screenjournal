package auth

import "errors"

type (
	AuthStore interface {
		ReadPasswordHash(username string) (PasswordHash, error)
	}

	Authenticator struct {
		store AuthStore
	}
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrIncorrectPassword = errors.New("password does not match stored hash")
)

func New(store AuthStore) Authenticator {
	return Authenticator{
		store: store,
	}
}

func (a Authenticator) Authenticate(username, password string) error {
	h, err := a.store.ReadPasswordHash(username)
	if err != nil {
		return err
	}

	if ok := h.MatchesPlaintext(password); !ok {
		return ErrIncorrectPassword
	}

	return nil
}
