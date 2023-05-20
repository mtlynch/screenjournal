package screenjournal

import (
	"errors"

	"github.com/mtlynch/screenjournal/v2/auth"
)

type (
	Email    string
	Username string
	Password string

	// EmailSubscriber represents a user or entity that subscribes to events via
	// email notifications.
	EmailSubscriber struct {
		Username Username
		Email    Email
	}

	PasswordHash struct {
		hash auth.PasswordHash
	}

	User struct {
		IsAdmin      bool
		Username     Username
		Email        Email
		PasswordHash PasswordHash
	}
)

var (
	ErrUserNotFound = errors.New("user not found")
	//ErrIncorrectPassword = errors.New("password does not match stored hash")
)

func (e Email) String() string {
	return string(e)
}

func (u Username) String() string {
	return string(u)
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

func (ph PasswordHash) MatchesPlaintext(plaintext string) bool {
	if ph.hash == nil {
		return false
	}
	return ph.hash.MatchesPlaintext(plaintext)
}

func NewPasswordHash(hash []byte) PasswordHash {
	return PasswordHash{
		hash: auth.NewPasswordHashFromBytes(hash),
	}
}

func (ph PasswordHash) String() string {
	if ph.hash == nil {
		return ""
	}
	return ph.hash.String()
}

func (ph PasswordHash) Bytes() []byte {
	if ph.hash == nil {
		return []byte{}
	}
	return ph.hash.Bytes()
}

func (u User) IsEmpty() bool {
	return u.Username == ""
}
