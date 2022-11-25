package handlers

import "net/http"

func (s *Server) routes() {
	s.router.HandleFunc("/api/auth", s.authPost()).Methods(http.MethodPost)
	s.router.HandleFunc("/api/auth", s.authDelete()).Methods(http.MethodDelete)
	// Populate auth information in request context.
	s.router.Use(s.checkAuthentication)

	authenticatedApis := s.router.PathPrefix("/api").Subrouter()
	authenticatedApis.Use(s.requireAuthentication)
	authenticatedApis.HandleFunc("/search", s.searchGet()).Methods(http.MethodGet)
	authenticatedApis.HandleFunc("/reviews", s.reviewsPost()).Methods(http.MethodPost)
	authenticatedApis.HandleFunc("/reviews/{reviewID}", s.reviewsPut()).Methods(http.MethodPut)

	static := s.router.PathPrefix("/").Subrouter()
	static.PathPrefix("/js/").HandlerFunc(serveStaticResource()).Methods(http.MethodGet)
	static.PathPrefix("/third-party/").HandlerFunc(serveStaticResource()).Methods(http.MethodGet)

	authenticatedViews := s.router.PathPrefix("/").Subrouter()
	authenticatedViews.Use(s.requireAuthentication)
	authenticatedViews.HandleFunc("/reviews", s.reviewsGet()).Methods(http.MethodGet)
	authenticatedViews.HandleFunc("/reviews/{reviewID}/edit", s.reviewsEditGet()).Methods(http.MethodGet)

	views := s.router.PathPrefix("/").Subrouter()
	views.Use(upgradeToHttps)
	views.HandleFunc("/about", s.aboutGet()).Methods(http.MethodGet)
	views.HandleFunc("/login", s.logInGet()).Methods(http.MethodGet)
	views.HandleFunc("/logout", s.logOutGet()).Methods(http.MethodGet)
	views.HandleFunc("/reviews/new", s.reviewsNewGet()).Methods(http.MethodGet)
	views.HandleFunc("/sign-up", s.signUpGet()).Methods(http.MethodGet)
	views.HandleFunc("/", s.indexGet()).Methods(http.MethodGet)

	s.addDevRoutes()
}
