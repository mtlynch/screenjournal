package handlers

import "net/http"

func (s *Server) routes() {
	views := s.router.PathPrefix("/").Subrouter()
	views.Use(upgradeToHttps)
	views.HandleFunc("/", s.indexGet()).Methods(http.MethodGet)
}
