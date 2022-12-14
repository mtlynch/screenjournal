package screenjournal

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type (
	Email    string
	Username string
	Password string

	PasswordHash struct {
		bytes []byte
	}

	User struct {
		IsAdmin      bool
		Username     Username
		Email        Email
		PasswordHash PasswordHash
	}
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrIncorrectPassword = errors.New("password does not match stored hash")
)

func (e Email) String() string {
	return string(e)
}

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

func NewPasswordHashFromBytes(bytes []byte) PasswordHash {
	return PasswordHash{
		bytes: bytes,
	}
}

func (h PasswordHash) MatchesPlaintext(plaintext Password) error {
	err := bcrypt.CompareHashAndPassword(h.bytes, []byte(plaintext.String()))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return ErrIncorrectPassword
	}
	return err
}

func (h PasswordHash) String() string {
	return string(h.bytes)
}

func (h PasswordHash) Bytes() []byte {
	return h.bytes
}
