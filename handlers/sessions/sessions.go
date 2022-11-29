package sessions

import (
	"context"
	"errors"
	"net/http"

	"github.com/mtlynch/screenjournal/v2"
)

type (
	Manager interface {
		CreateSession(http.ResponseWriter, *http.Request, screenjournal.Username) error
		SessionFromRequest(*http.Request) (Session, error)
		EndSession(context.Context, http.ResponseWriter) error
	}

	Session struct {
		UserAuth screenjournal.UserAuth
	}
)

var ErrNotAuthenticated = errors.New("no auth cookie")
