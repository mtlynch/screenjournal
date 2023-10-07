package auth

import "errors"

type (
	Authenticator interface {
		Authenticate(username, password string) error
	}

	AuthStore interface {
		ReadPasswordHash(username string) (PasswordHash, error)
	}

	authenticator struct {
		store AuthStore
	}
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrIncorrectPassword = errors.New("password does not match stored hash")
)

func New(store AuthStore) Authenticator {
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
		return ErrIncorrectPassword
	}

	return nil
}
