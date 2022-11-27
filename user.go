package screenjournal

import (
	"golang.org/x/crypto/bcrypt"
)

type (
	Username string

	PasswordHash struct {
		bytes []byte
	}

	UserAuth struct {
		IsAdmin  bool
		Username Username
	}
)

func (u Username) String() string {
	return string(u)
}

func (u Username) IsEmpty() bool {
	return u.String() == ""
}

func (u Username) Equal(o Username) bool {
	return u.String() == o.String()
}

func (ua UserAuth) IsEmpty() bool {
	return ua.Username == ""
}

func NewPasswordHash(plaintext []byte) PasswordHash {
	bytes, err := bcrypt.GenerateFromPassword(plaintext, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return PasswordHash{
		bytes: bytes,
	}
}

func (h PasswordHash) MatchesPlaintext(plaintext string) error {
	return bcrypt.CompareHashAndPassword(h.bytes, []byte(plaintext))
}

func (h PasswordHash) String() string {
	return string(h.bytes)
}
