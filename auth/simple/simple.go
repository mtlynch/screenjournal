package simple

import (
	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/auth"
)

type (
	authenticator struct {
		username     screenjournal.Username
		passwordHash screenjournal.PasswordHash
	}
)

func New(username screenjournal.Username, password screenjournal.Password) (auth.Authenticator, error) {
	return authenticator{
		username:     username,
		passwordHash: screenjournal.NewPasswordHash(password),
	}, nil
}

func (a authenticator) Authenticate(username screenjournal.Username, password screenjournal.Password) error {
	if !a.username.Equal(username) {
		return auth.ErrInvalidCredentials
	}
	return a.passwordHash.MatchesPlaintext(password)
}
