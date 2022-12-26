package auth

import (
	"github.com/mtlynch/screenjournal/v2"
)

type Authenticator interface {
	Authenticate(screenjournal.Username, screenjournal.Password) (screenjournal.User, error)
}
