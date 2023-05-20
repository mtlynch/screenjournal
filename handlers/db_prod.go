//go:build !dev

package handlers

import (
	"net/http"

	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/store"
)

func (s *Server) addDevRoutes() {
	// no-op
}

func (s Server) getDB(*http.Request) store.Store {
	return s.store
}

func (s Server) getAuthenticator(r *http.Request) auth.Authenticator {
	return s.authenticator
}
