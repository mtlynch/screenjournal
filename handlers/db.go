package handlers

import (
  "net/http"

  "github.com/mtlynch/screenjournal/v2/store/sqlite"
)

func (s Server) getDB(*http.Request) sqlite.Store {
  return s.store
}

func (s Server) getAuthenticator(*http.Request) Authenticator {
  return s.authenticator
}
