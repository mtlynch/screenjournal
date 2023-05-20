package auth

import (
	"errors"
	"log"

	"golang.org/x/crypto/bcrypt"
)

type (
	Username string
	Password string

	bcryptPasswordHash []byte

	PasswordHash interface {
		MatchesPlaintext(string) error
		String() string
		Bytes() []byte
	}

	Authenticator interface {
		Authenticate(username, password string) error
	}
)

var ErrIncorrectPassword = errors.New("password does not match stored hash")

func (pw Password) String() string {
	return string(pw)
}

func NewPasswordHash(plaintext string) (PasswordHash, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	log.Printf("password %s -> %s", plaintext, string(bytes))
	return bcryptPasswordHash(bytes), nil
}

func NewPasswordHashFromBytes(bytes []byte) PasswordHash {
	return bcryptPasswordHash(bytes)
}

// TODO: return bool
func (h bcryptPasswordHash) MatchesPlaintext(plaintext string) error {
	err := bcrypt.CompareHashAndPassword(h.Bytes(), []byte(plaintext))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return ErrIncorrectPassword
	}
	return err
}

func (h bcryptPasswordHash) String() string {
	return string(h)
}

func (h bcryptPasswordHash) Bytes() []byte {
	return []byte(h)
}
