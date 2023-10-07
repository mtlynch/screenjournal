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

// Hash a plaintext password into a secure password hash.
func HashPassword(plaintext string) (PasswordHash, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return bcryptPasswordHash(bytes), nil
}

// Converts raw bytes into a password hash. Note that this doesn't perform a
// hash on the bytes. The bytes represent an already-hashed password.
func PasswordHashFromBytes(bytes []byte) PasswordHash {
	return bcryptPasswordHash(bytes)
}
