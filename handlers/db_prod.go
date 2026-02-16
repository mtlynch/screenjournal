//go:build !dev

package handlers

import (
	"net/http"
)

func (s *Server) addDevRoutes() {
	// no-op
}

type staticDBProvider struct {
	store Store
}

func newDBProvider(store Store) dbProvider {
	return staticDBProvider{
		store: store,
	}
}

func (p staticDBProvider) dbForRequest(r *http.Request) dbService {
	return dbService{
		Store:   p.store,
		request: r,
	}
}

func (p staticDBProvider) authenticatorForRequest(_ *http.Request, fallback Authenticator) Authenticator {
	return fallback
}
