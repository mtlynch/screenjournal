package auth

import (
	"errors"

	"github.com/mtlynch/screenjournal/v2"
)

type Authenticator interface {
	Authenticate(screenjournal.Username, screenjournal.Password) error
}

var ErrInvalidCredentials = errors.New("invalid credentials")
