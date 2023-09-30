package sessions

import (
	"errors"
	"net/http"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type (
	Manager interface {
		CreateSession(http.ResponseWriter, *http.Request, screenjournal.User) error
		SessionFromRequest(*http.Request) (Session, error)
		EndSession(*http.Request, http.ResponseWriter)
		// WrapRequest wraps the given handler, adding the Session object (if
		// there's an active session) to the request context before passing control
		// to the next handler.
		WrapRequest(http.Handler) http.Handler
	}

	Session struct {
		User screenjournal.User
	}
)

var ErrNotAuthenticated = errors.New("user has no active screenjournal session")
