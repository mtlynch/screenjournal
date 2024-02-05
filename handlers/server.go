package handlers

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mtlynch/screenjournal/v2/announce"
	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
	"github.com/mtlynch/screenjournal/v2/metadata"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
)

type (
	SessionManager interface {
		CreateSession(http.ResponseWriter, context.Context, screenjournal.Username, bool) error
		SessionFromContext(context.Context) (sessions.Session, error)
		EndSession(context.Context, http.ResponseWriter)
		// WrapRequest wraps the given handler, adding the Session object (if
		// there's an active session) to the request context before passing control
		// to the next handler.
		WrapRequest(http.Handler) http.Handler
	}

	Authenticator interface {
		Authenticate(username screenjournal.Username, password screenjournal.Password) error
	}

	Server struct {
		router         *mux.Router
		authenticator  Authenticator
		announcer      announce.Announcer
		sessionManager SessionManager
		store          store.Store
		metadataFinder metadata.Finder
	}
)

// Router returns the underlying router interface for the server.
func (s Server) Router() *mux.Router {
	return s.router
}

// New creates a new server with all the state it needs to satisfy HTTP
// requests.
func New(authenticator Authenticator, announcer announce.Announcer, sessionManager SessionManager, store store.Store, metadataFinder metadata.Finder) Server {
	s := Server{
		router:         mux.NewRouter(),
		authenticator:  authenticator,
		announcer:      announcer,
		sessionManager: sessionManager,
		store:          store,
		metadataFinder: metadataFinder,
	}

	s.routes()
	return s
}
