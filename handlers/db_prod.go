//go:build !dev

package handlers

import (
	"net/http"
)

func (s *Server) addDevRoutes() {
	// no-op
}

func (s Server) getDB(r *http.Request) dbService {
	return dbService{
		Store:   s.store,
		request: r,
	}
}

func (s Server) getAuthenticator(_ *http.Request) Authenticator {
	return s.authenticator
}
