package handlers

import "net/http"

func (s *Server) routes() {
	static := s.router.PathPrefix("/").Subrouter()
	static.PathPrefix("/third-party/").HandlerFunc(serveStaticResource()).Methods(http.MethodGet)

	views := s.router.PathPrefix("/").Subrouter()
	views.Use(upgradeToHttps)
	views.HandleFunc("/about", s.aboutGet()).Methods(http.MethodGet)
	views.HandleFunc("/login", s.logInGet()).Methods(http.MethodGet)
	views.HandleFunc("/sign-up", s.signUpGet()).Methods(http.MethodGet)
	views.HandleFunc("/", s.indexGet()).Methods(http.MethodGet)
}
