package handlers

import (
	"github.com/gorilla/mux"

	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/metadata"
	"github.com/mtlynch/screenjournal/v2/sessions"
	"github.com/mtlynch/screenjournal/v2/store"
)

type Server struct {
	router         *mux.Router
	authenticator  auth.Authenticator
	sessionManager sessions.Manager
	store          store.Store
	metadataFinder metadata.Finder
}

// Router returns the underlying router interface for the server.
func (s Server) Router() *mux.Router {
	return s.router
}

// New creates a new server with all the state it needs to satisfy HTTP
// requests.
func New(authenticator auth.Authenticator, sessionManager sessions.Manager, store store.Store, metadataFinder metadata.Finder) Server {
	s := Server{
		router:         mux.NewRouter(),
		authenticator:  authenticator,
		sessionManager: sessionManager,
		store:          store,
		metadataFinder: metadataFinder,
	}

	s.routes()
	return s
}
