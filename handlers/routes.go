package handlers

import "net/http"

func (s *Server) routes() {
	s.router.HandleFunc("/api/auth", s.authPost()).Methods(http.MethodPost)
	s.router.HandleFunc("/api/auth", s.authDelete()).Methods(http.MethodDelete)
	s.router.HandleFunc("/api/users/{username}", s.usersPut()).Methods(http.MethodPut)
	s.router.Use(s.populateAuthenticationContext)

	adminApis := s.router.PathPrefix("/api/admin").Subrouter()
	adminApis.Use(s.requireAuthenticationForAPI)
	adminApis.Use(s.requireAdmin)
	adminApis.HandleFunc("/repopulate-movies", s.repopulateMoviesGet()).Methods(http.MethodGet)
	adminApis.HandleFunc("/invites", s.invitesPost()).Methods(http.MethodPost)

	authenticatedApis := s.router.PathPrefix("/api").Subrouter()
	authenticatedApis.Use(s.requireAuthenticationForAPI)
	authenticatedApis.HandleFunc("/search", s.searchGet()).Methods(http.MethodGet)
	authenticatedApis.HandleFunc("/reviews", s.reviewsPost()).Methods(http.MethodPost)
	authenticatedApis.HandleFunc("/reviews/{reviewID}", s.reviewsPut()).Methods(http.MethodPut)

	static := s.router.PathPrefix("/").Subrouter()
	static.PathPrefix("/css/").HandlerFunc(serveStaticResource()).Methods(http.MethodGet)
	static.PathPrefix("/js/").HandlerFunc(serveStaticResource()).Methods(http.MethodGet)
	static.PathPrefix("/third-party/").HandlerFunc(serveStaticResource()).Methods(http.MethodGet)

	adminViews := s.router.PathPrefix("/admin").Subrouter()
	adminViews.Use(s.requireAuthenticationForView)
	adminViews.Use(s.requireAdmin)
	adminViews.Use(enforceContentSecurityPolicy)
	adminViews.HandleFunc("/invites", s.invitesGet()).Methods(http.MethodGet)
	adminViews.HandleFunc("/invites/new", s.invitesNewGet()).Methods(http.MethodGet)

	views := s.router.PathPrefix("/").Subrouter()
	views.Use(upgradeToHttps)
	views.Use(enforceContentSecurityPolicy)
	views.HandleFunc("/about", s.aboutGet()).Methods(http.MethodGet)
	views.HandleFunc("/login", s.logInGet()).Methods(http.MethodGet)
	views.HandleFunc("/reviews/new", s.reviewsNewGet()).Methods(http.MethodGet)
	views.HandleFunc("/sign-up", s.signUpGet()).Methods(http.MethodGet)
	views.HandleFunc("/", s.indexGet()).Methods(http.MethodGet)

	authenticatedViews := s.router.PathPrefix("/").Subrouter()
	authenticatedViews.Use(s.requireAuthenticationForView)
	authenticatedViews.Use(enforceContentSecurityPolicy)
	authenticatedViews.HandleFunc("/reviews", s.reviewsGet()).Methods(http.MethodGet)
	authenticatedViews.HandleFunc("/reviews/by/{username}", s.reviewsGet()).Methods(http.MethodGet)
	authenticatedViews.HandleFunc("/reviews/{reviewID}", s.reviewsReadGet()).Methods(http.MethodGet)
	authenticatedViews.HandleFunc("/reviews/{reviewID}/edit", s.reviewsEditGet()).Methods(http.MethodGet)

	s.addDevRoutes()
}
