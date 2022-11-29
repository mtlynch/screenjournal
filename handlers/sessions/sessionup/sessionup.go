package sessionup

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/swithek/sessionup"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
	"github.com/swithek/sessionup/memstore"
)

type (
	wrapper struct {
		manager *sessionup.Manager
	}
)

func New() (sessions.Manager, error) {
	store := memstore.New(time.Minute * 5)
	wrapper := wrapper{
		manager: sessionup.NewManager(store, sessionup.Secure(false)),
	}
	return wrapper, nil
}

func (wrapper wrapper) CreateSession(w http.ResponseWriter, r *http.Request, username screenjournal.Username) error {
	return wrapper.manager.Init(w, r, username.String())
}

func (wrapper wrapper) SessionFromRequest(r *http.Request) (sessions.Session, error) {
	log.Printf("getting session from request") // DEBUG
	ss, ok := sessionup.FromContext(r.Context())
	log.Printf("ss=%+v, ok=%v", ss, ok) // DEBUG
	if !ok {
		return sessions.Session{}, sessions.ErrNotAuthenticated
	}

	return sessions.Session{
		UserAuth: screenjournal.UserAuth{
			Username: screenjournal.Username(ss.ID),
		},
	}, nil
}

func (wrapper wrapper) EndSession(ctx context.Context, w http.ResponseWriter) error {
	return wrapper.manager.Revoke(ctx, w)
}
