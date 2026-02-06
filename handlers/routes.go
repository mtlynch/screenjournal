package handlers

import "net/http"

func withMiddleware(h http.Handler, mw ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}

// withMiddlewareFunc is a convenience wrapper for http.HandlerFunc.
func withMiddlewareFunc(h http.HandlerFunc, mw ...func(http.Handler) http.Handler) http.Handler {
	return withMiddleware(h, mw...)
}

func (s *Server) routes() {
	// Unauthenticated APIs
	s.router.Handle("POST /api/auth", s.authPost())
	s.router.Handle("DELETE /api/auth", s.authDelete())
	s.router.Handle("PUT /api/users/{username}", s.usersPut())

	// Admin APIs
	adminAPI := func(h http.HandlerFunc) http.Handler {
		return withMiddlewareFunc(h, s.requireAuthenticationForAPI, s.requireAdmin)
	}
	s.router.Handle("GET /api/admin/repopulate/movies", adminAPI(s.repopulateMoviesGet()))
	s.router.Handle("GET /api/admin/repopulate/tv", adminAPI(s.repopulateTvShowsGet()))
	s.router.Handle("POST /api/admin/invites", adminAPI(s.invitesPost()))

	// Authenticated APIs
	authAPI := func(h http.HandlerFunc) http.Handler {
		return withMiddlewareFunc(h, s.requireAuthenticationForAPI)
	}
	s.router.Handle("POST /api/comments", authAPI(s.commentsPost()))
	s.router.Handle("GET /api/comments/add", authAPI(s.commentsAddGet()))
	s.router.Handle("GET /api/comments/edit", authAPI(s.commentsEditGet()))
	s.router.Handle("GET /api/comments/{commentID}", authAPI(s.commentsGet()))
	s.router.Handle("PUT /api/comments/{commentID}", authAPI(s.commentsPut()))
	s.router.Handle("DELETE /api/comments/{commentID}", authAPI(s.commentsDelete()))
	s.router.Handle("GET /api/search", authAPI(s.searchGet()))

	// Static files
	staticHandler := getStaticFilesHandler()
	s.router.Handle("GET /css/", staticHandler)
	s.router.Handle("GET /js/", staticHandler)
	s.router.Handle("GET /third-party/", staticHandler)

	// Admin views
	adminViewMW := func(h http.HandlerFunc) http.Handler {
		return withMiddlewareFunc(h, s.requireAuthenticationForView, s.requireAdmin, enforceContentSecurityPolicy)
	}
	s.router.Handle("GET /admin/invites", adminViewMW(s.invitesGet()))
	s.router.Handle("GET /admin/reset-password", adminViewMW(s.passwordResetAdminGet()))

	// Public views
	viewMW := func(h http.HandlerFunc) http.Handler {
		return withMiddlewareFunc(h, upgradeToHttps, enforceContentSecurityPolicy)
	}
	s.router.Handle("GET /about", viewMW(s.aboutGet()))
	s.router.Handle("GET /login", viewMW(s.logInGet()))
	s.router.Handle("GET /sign-up", viewMW(s.signUpGet()))
	s.router.Handle("GET /account/password-reset", viewMW(s.accountPasswordResetGet()))
	s.router.Handle("PUT /account/password-reset", viewMW(s.accountPasswordResetPut()))
	s.router.Handle("GET /{$}", viewMW(s.indexGet()))

	// Authenticated routes (transitional)
	authRouteMW := func(h http.HandlerFunc) http.Handler {
		return withMiddlewareFunc(h, s.requireAuthenticationForAPI, enforceContentSecurityPolicy)
	}
	s.router.Handle("PUT /account/notifications", authRouteMW(s.accountNotificationsPut()))
	s.router.Handle("PUT /account/password", authRouteMW(s.accountChangePasswordPut()))
	s.router.Handle("POST /reviews", authRouteMW(s.reviewsPost()))
	s.router.Handle("PUT /reviews/{reviewID}", authRouteMW(s.reviewsPut()))
	s.router.Handle("DELETE /reviews/{reviewID}", authRouteMW(s.reviewsDelete()))
	s.router.Handle("POST /reactions", authRouteMW(s.reactionsPost()))
	s.router.Handle("DELETE /reactions/{reactionID}", authRouteMW(s.reactionsDelete()))

	// Admin routes (transitional)
	adminRouteMW := func(h http.HandlerFunc) http.Handler {
		return withMiddlewareFunc(h, s.requireAuthenticationForAPI, s.requireAdmin, enforceContentSecurityPolicy)
	}
	s.router.Handle("POST /admin/invites", adminRouteMW(s.invitesPost()))
	s.router.Handle("POST /admin/reset-password", adminRouteMW(s.passwordResetAdminPost()))
	s.router.Handle("DELETE /admin/reset-password/{token}", adminRouteMW(s.passwordResetAdminDelete()))

	// Authenticated views
	authViewMW := func(h http.HandlerFunc) http.Handler {
		return withMiddlewareFunc(h, s.requireAuthenticationForView, enforceContentSecurityPolicy)
	}
	s.router.Handle("GET /account/change-password", authViewMW(s.accountChangePasswordGet()))
	s.router.Handle("GET /account/notifications", authViewMW(s.accountNotificationsGet()))
	s.router.Handle("GET /account/security", authViewMW(s.accountSecurityGet()))
	s.router.Handle("GET /activity", authViewMW(s.activityGet()))
	s.router.Handle("GET /movies/{movieID}", authViewMW(s.moviesReadGet()))
	s.router.Handle("GET /tv-shows/{tvShowID}", authViewMW(s.tvShowsReadGet()))
	s.router.Handle("GET /reviews", authViewMW(s.reviewsGet()))
	s.router.Handle("GET /reviews/new", authViewMW(s.reviewsNewTitleSearchGet()))
	s.router.Handle("GET /reviews/new/tv/pick-season", authViewMW(s.reviewsNewPickSeasonGet()))
	s.router.Handle("GET /reviews/new/write", authViewMW(s.reviewsNewWriteReviewGet()))
	// "/reviews/by/{username}" and "/reviews/{reviewID}/edit" conflict in
	// Go's ServeMux because neither pattern is more specific than the other
	// (they both match "/reviews/by/edit"). We use a catch-all wildcard and
	// dispatch manually.
	s.router.Handle("GET /reviews/{segment}/{rest...}", authViewMW(s.reviewsSubpathDispatch()))
	s.router.Handle("GET /users", authViewMW(s.usersGet()))

	s.addDevRoutes()
}

// reviewsSubpathDispatch handles the ambiguous /reviews/{segment}/{rest...}
// pattern, manually dispatching to either /reviews/by/{username} or
// /reviews/{reviewID}/edit.
func (s *Server) reviewsSubpathDispatch() http.HandlerFunc {
	byUsername := s.reviewsGet()
	editReview := s.reviewsEditGet()
	return func(w http.ResponseWriter, r *http.Request) {
		segment := r.PathValue("segment")
		rest := r.PathValue("rest")
		switch {
		case segment == "by" && rest != "":
			r.SetPathValue("username", rest)
			byUsername.ServeHTTP(w, r)
		case rest == "edit":
			r.SetPathValue("reviewID", segment)
			editReview.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	}
}
