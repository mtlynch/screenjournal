//go:build !dev

package handlers

import (
	"net/http"

	"github.com/mtlynch/screenjournal/v2/store/sqlite"
)

func (s *Server) initDev() {
	// no-op
}

func (s *Server) addDevRoutes() {
	// no-op
}

func (s Server) getDB(r *http.Request) sqlite.Store {
	return s.store.WithContext(r.Context())
}

func (s Server) getAuthenticator(_ *http.Request) Authenticator {
	return s.authenticator
}
