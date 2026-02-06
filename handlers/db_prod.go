//go:build !dev

package handlers

import (
	"net/http"
)

func (s *Server) addDevRoutes() {
	// no-op
}

func (s Server) getDB(*http.Request) Store {
	return s.store
}

func (s Server) getAuthenticator(_ *http.Request) Authenticator {
	return s.authenticator
}

// devMiddleware returns the handler unchanged in production builds.
func (s Server) devMiddleware(h http.Handler) http.Handler {
	return h
}
