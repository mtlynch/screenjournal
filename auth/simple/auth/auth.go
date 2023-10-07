package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type (
	Authenticator interface {
		Authenticate(username, password string) error
	}

	PasswordHash interface {
		MatchesPlaintext(string) bool
		Bytes() []byte
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

func NewPasswordHash(plaintext string) (PasswordHash, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return bcryptPasswordHash(bytes), nil
}

func NewPasswordHashFromBytes(bytes []byte) PasswordHash {
	return bcryptPasswordHash(bytes)
}
