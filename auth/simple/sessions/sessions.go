package sessions

import (
	"errors"
	"net/http"
)

type (
	Manager interface {
		CreateSession(http.ResponseWriter, *http.Request, Metadata) error
		SessionFromRequest(*http.Request) (Session, error)
		EndSession(*http.Request, http.ResponseWriter)
		// WrapRequest wraps the given handler, adding the Session object (if
		// there's an active session) to the request context before passing control
		// to the next handler.
		WrapRequest(http.Handler) http.Handler
	}

	Session interface {
		Metadata() Metadata
	}

	Metadata struct {
		Username string `json:"username"`
		IsAdmin  bool   `json:"isAdmin"`
	}
)

var ErrNotAuthenticated = errors.New("user has no active screenjournal session")
