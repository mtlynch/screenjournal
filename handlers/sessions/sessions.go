package sessions

import (
	"errors"
	"net/http"

	"github.com/mtlynch/screenjournal/v2"
)

type (
	Manager interface {
		CreateSession(http.ResponseWriter, *http.Request, screenjournal.User) error
		SessionFromRequest(*http.Request) (Session, error)
		EndSession(*http.Request, http.ResponseWriter) error
		WrapRequest(http.Handler) http.Handler
	}

	Session struct {
		User screenjournal.User
	}
)

var ErrNotAuthenticated = errors.New("user has no active screenjournal session")
