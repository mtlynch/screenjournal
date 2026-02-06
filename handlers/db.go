package handlers

import "net/http"

func (s Server) getDB(r *http.Request) dbService {
	return s.db.dbForRequest(r)
}

func (s Server) getAuthenticator(r *http.Request) Authenticator {
	return s.db.authenticatorForRequest(r, s.authenticator)
}
