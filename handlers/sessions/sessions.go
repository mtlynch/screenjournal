package sessions

import (
	"errors"
	"net/http"

	"github.com/mtlynch/screenjournal/v2"
)

type (
	Manager interface {
		CreateSession(http.ResponseWriter, *http.Request, screenjournal.Username) error
		SessionFromRequest(*http.Request) (Session, error)
		EndSession(*http.Request, http.ResponseWriter) error
		WrapRequest(http.Handler) http.Handler
	}

	Session struct {
		UserAuth screenjournal.UserAuth
	}
)

var ErrNotAuthenticated = errors.New("no auth cookie")
