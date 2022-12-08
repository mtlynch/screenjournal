package screenjournal

import (
	"golang.org/x/crypto/bcrypt"
)

type (
	Username string
	Password string

	PasswordHash struct {
		bytes []byte
	}

	User struct {
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

func (pw Password) String() string {
	return string(pw)
}

// Equal returns true if two passwords match. Only use this in testing, never as
// a way of authenticating actual user passwords.
func (pw Password) Equal(o Password) bool {
	return pw.String() == o.String()
}

func (u User) IsEmpty() bool {
	return u.Username == ""
}

func NewPasswordHash(plaintext Password) PasswordHash {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plaintext.String()), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return PasswordHash{
		bytes: bytes,
	}
}

func (h PasswordHash) MatchesPlaintext(plaintext Password) error {
	return bcrypt.CompareHashAndPassword(h.bytes, []byte(plaintext.String()))
}

func (h PasswordHash) String() string {
	return string(h.bytes)
}
