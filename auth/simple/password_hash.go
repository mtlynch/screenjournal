package simple

import (
	"golang.org/x/crypto/bcrypt"
)

type PasswordHash interface {
	MatchesPlaintext(string) bool
	Bytes() []byte
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
