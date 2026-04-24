//go:build !dev

package handlers

import (
	"net/http"
)

func (s *Server) initDev() {
	// no-op
}

func (s *Server) addDevRoutes() {
	// no-op
}

func (s Server) getDB(*http.Request) Store {
	return s.store
}

func (s Server) getAuthenticator(_ *http.Request) Authenticator {
	return s.authenticator
}

func setDevAuthCookie(_ http.ResponseWriter, _ bool) {
	// no-op in production
}

func getDevAuthCookie(_ *http.Request) (isAdmin bool, ok bool) {
	// no-op in production
	return false, false
}

func clearDevAuthCookie(_ http.ResponseWriter) {
	// no-op in production
}
