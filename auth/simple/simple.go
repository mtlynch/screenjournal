package simple

import (
	"crypto/subtle"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/auth"
)

type (
	authenticator struct {
		adminUsername screenjournal.Username
		adminPassword screenjournal.Password
	}
)

func New(adminUsername screenjournal.Username, adminPassword screenjournal.Password) (authenticator, error) {
	return authenticator{
		adminUsername: adminUsername,
		adminPassword: adminPassword,
	}, nil
}

func (a authenticator) Authenticate(username screenjournal.Username, password screenjournal.Password) (screenjournal.User, error) {
	// This is an insecure, placeholder implementation for authentication until we
	// switch to bcrypt.
	valid := []byte(a.adminUsername.String() + a.adminPassword.String())
	attempt := []byte(username.String() + password.String())
	if subtle.ConstantTimeCompare(valid, attempt) != 1 {
		return screenjournal.User{}, auth.ErrInvalidCredentials
	}
	return screenjournal.User{
		Username: username,
		IsAdmin:  username.Equal(a.adminUsername),
	}, nil
}
