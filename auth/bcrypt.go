package auth

import "golang.org/x/crypto/bcrypt"

type bcryptPasswordHash []byte

func (h bcryptPasswordHash) MatchesPlaintext(plaintext string) bool {
	return bcrypt.CompareHashAndPassword(h.Bytes(), []byte(plaintext)) == nil
}

func (h bcryptPasswordHash) Bytes() []byte {
	return []byte(h)
}
