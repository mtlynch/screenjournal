package simple

import (
	"crypto/subtle"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/auth"
)

type (
	authenticator struct {
		username screenjournal.Username
		password screenjournal.Password
	}
)

func New(username screenjournal.Username, password screenjournal.Password) (authenticator, error) {
	return authenticator{
		username: username,
		password: password,
	}, nil
}

func (a authenticator) Authenticate(username screenjournal.Username, password screenjournal.Password) error {
	// This is an insecure, placeholder implementation for authentication until we
	// switch to bcrypt.
	valid := []byte(a.username.String() + a.password.String())
	attempt := []byte(username.String() + password.String())
	if subtle.ConstantTimeCompare(valid, attempt) != 1 {
		return auth.ErrInvalidCredentials
	}
	return nil
}
