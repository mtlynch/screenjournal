package simple

import (
	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/auth"
)

type (
	authenticator struct {
		adminUsername     screenjournal.Username
		adminPasswordHash screenjournal.PasswordHash
	}
)

func New(adminUsername screenjournal.Username, adminPassword screenjournal.Password) (auth.Authenticator, error) {
	return authenticator{
		adminUsername:     adminUsername,
		adminPasswordHash: screenjournal.NewPasswordHash(adminPassword),
	}, nil
}

func (a authenticator) Authenticate(username screenjournal.Username, password screenjournal.Password) (screenjournal.User, error) {
	if !a.adminUsername.Equal(username) {
		return screenjournal.User{}, auth.ErrInvalidCredentials
	}
	if err := a.adminPasswordHash.MatchesPlaintext(password); err != nil {
		return screenjournal.User{}, err
	}
	return screenjournal.User{
		Username: username,
		IsAdmin:  username.Equal(a.adminUsername),
	}, nil
}
