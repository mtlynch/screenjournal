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
	adminApis.HandleFunc("/repopulate/movies", s.repopulateMoviesGet()).Methods(http.MethodGet)
	adminApis.HandleFunc("/repopulate/tv", s.repopulateTvShowsGet()).Methods(http.MethodGet)
	adminApis.HandleFunc("/invites", s.invitesPost()).Methods(http.MethodPost)

	authenticatedApis := s.router.PathPrefix("/api").Subrouter()
	authenticatedApis.Use(s.requireAuthenticationForAPI)
	authenticatedApis.HandleFunc("/comments", s.commentsPost()).Methods(http.MethodPost)
	authenticatedApis.HandleFunc("/comments/add", s.commentsAddGet()).Methods(http.MethodGet)
	authenticatedApis.HandleFunc("/comments/edit", s.commentsEditGet()).Methods(http.MethodGet)
	authenticatedApis.HandleFunc("/comments/{commentID}", s.commentsGet()).Methods(http.MethodGet)
	authenticatedApis.HandleFunc("/comments/{commentID}", s.commentsPut()).Methods(http.MethodPut)
	authenticatedApis.HandleFunc("/comments/{commentID}", s.commentsDelete()).Methods(http.MethodDelete)
	authenticatedApis.HandleFunc("/search", s.searchGet()).Methods(http.MethodGet)

	static := s.router.PathPrefix("/").Subrouter()
	static.PathPrefix("/css/").HandlerFunc(serveStaticResource()).Methods(http.MethodGet)
	static.PathPrefix("/js/").HandlerFunc(serveStaticResource()).Methods(http.MethodGet)
	static.PathPrefix("/third-party/").HandlerFunc(serveStaticResource()).Methods(http.MethodGet)

	adminViews := s.router.PathPrefix("/admin").Subrouter()
	adminViews.Use(s.requireAuthenticationForView)
	adminViews.Use(s.requireAdmin)
	adminViews.Use(enforceContentSecurityPolicy)
	adminViews.HandleFunc("/invites", s.invitesGet()).Methods(http.MethodGet)

	views := s.router.PathPrefix("/").Subrouter()
	views.Use(upgradeToHttps)
	views.Use(enforceContentSecurityPolicy)
	views.HandleFunc("/about", s.aboutGet()).Methods(http.MethodGet)
	views.HandleFunc("/login", s.logInGet()).Methods(http.MethodGet)
	views.HandleFunc("/sign-up", s.signUpGet()).Methods(http.MethodGet)
	views.HandleFunc("/", s.indexGet()).Methods(http.MethodGet)

	// Transitional subrouter as we get rid of the idea of separate API routes vs.
	// view routes.
	authenticatedRoutes := s.router.PathPrefix("/").Subrouter()
	authenticatedRoutes.Use(s.requireAuthenticationForAPI)
	authenticatedRoutes.Use(enforceContentSecurityPolicy)
	authenticatedRoutes.HandleFunc("/account/notifications", s.accountNotificationsPut()).Methods(http.MethodPut)
	authenticatedRoutes.HandleFunc("/account/password", s.accountChangePasswordPut()).Methods(http.MethodPut)
	authenticatedRoutes.HandleFunc("/reviews", s.reviewsPost()).Methods(http.MethodPost)
	authenticatedRoutes.HandleFunc("/reviews/{reviewID}", s.reviewsPut()).Methods(http.MethodPut)
	authenticatedRoutes.HandleFunc("/reviews/{reviewID}", s.reviewsDelete()).Methods(http.MethodDelete)

	// Transitional subrouter as we get rid of the idea of separate API routes vs.
	// view routes.
	adminRoutes := s.router.PathPrefix("/admin").Subrouter()
	adminRoutes.Use(s.requireAuthenticationForAPI)
	adminRoutes.Use(s.requireAdmin)
	adminRoutes.Use(enforceContentSecurityPolicy)
	adminRoutes.HandleFunc("/invites", s.invitesPost()).Methods(http.MethodPost)

	authenticatedViews := s.router.PathPrefix("/").Subrouter()
	authenticatedViews.Use(s.requireAuthenticationForView)
	authenticatedViews.Use(enforceContentSecurityPolicy)
	authenticatedViews.HandleFunc("/account/change-password", s.accountChangePasswordGet()).Methods(http.MethodGet)
	authenticatedViews.HandleFunc("/account/notifications", s.accountNotificationsGet()).Methods(http.MethodGet)
	authenticatedViews.HandleFunc("/account/security", s.accountSecurityGet()).Methods(http.MethodGet)
	authenticatedViews.HandleFunc("/movies/{movieID}", s.moviesReadGet()).Methods(http.MethodGet)
	authenticatedViews.HandleFunc("/tv-shows/{tvShowID}", s.tvShowsReadGet()).Methods(http.MethodGet)
	authenticatedViews.HandleFunc("/reviews", s.reviewsGet()).Methods(http.MethodGet)
	authenticatedViews.HandleFunc("/reviews/by/{username}", s.reviewsGet()).Methods(http.MethodGet)
	authenticatedViews.HandleFunc("/reviews/new", s.reviewsNewTitleSearchGet()).Methods(http.MethodGet)
	authenticatedViews.HandleFunc("/reviews/new/tv/pick-season", s.reviewsNewPickSeasonGet()).Methods(http.MethodGet)
	authenticatedViews.HandleFunc("/reviews/new/write", s.reviewsNewWriteReviewGet()).Methods(http.MethodGet)
	authenticatedViews.HandleFunc("/reviews/{reviewID}/edit", s.reviewsEditGet()).Methods(http.MethodGet)
	authenticatedViews.HandleFunc("/users", s.usersGet()).Methods(http.MethodGet)

	s.addDevRoutes()
}
