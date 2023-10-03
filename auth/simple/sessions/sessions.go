package sessions

import (
	"context"
	"errors"
	"net/http"
)

type (
	Manager interface {
		CreateSession(http.ResponseWriter, context.Context, Key, Session) error
		SessionFromRequest(context.Context) (Session, error)
		EndSession(context.Context, http.ResponseWriter)
		// WrapRequest wraps the given handler, adding the Session object (if
		// there's an active session) to the request context before passing control
		// to the next handler.
		WrapRequest(http.Handler) http.Handler
	}

	Key struct {
		key []byte
	}

	Session []byte
)

var ErrNotAuthenticated = errors.New("user has no active session")

func KeyFromBytes(b []byte) Key {
	return Key{
		key: b,
	}
}

func (k Key) Bytes() []byte {
	return k.key
}
