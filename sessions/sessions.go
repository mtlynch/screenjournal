package sessions

import (
	"context"
	"errors"
	"net/http"

	"github.com/mtlynch/screenjournal/v2"
)

type Manager interface {
	Create(http.ResponseWriter, *http.Request, screenjournal.Username) error
	End(context.Context, http.ResponseWriter)
}

var ErrNotAuthenticated = errors.New("no auth cookie")
