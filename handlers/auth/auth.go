package auth

import (
	"errors"
	"net/http"

	"github.com/mtlynch/screenjournal/v2"
)

type Authenticator interface {
	StartSession(w http.ResponseWriter, r *http.Request)
	ClearSession(w http.ResponseWriter)
	Authenticate(r *http.Request) (screenjournal.UserAuth, error)
}

var ErrNotAuthenticated = errors.New("no auth cookie")
