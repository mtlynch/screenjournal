package handlers

import (
	"context"
	"net/http"

	"github.com/mtlynch/screenjournal/v2"
)

type contextKey struct {
	name string
}

var contextKeyUser = &contextKey{"user"}

func (s Server) authPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.authenticator.StartSession(w, r)
	}
}

func (s Server) authDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.authenticator.ClearSession(w)
	}
}

func (s Server) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := s.authenticator.Authenticate(r)
		if err != nil {
			s.authenticator.ClearSession(w)
			http.Error(w, "Invalid username", http.StatusBadRequest)
			return
		}

		if user.IsEmpty() {
			s.authenticator.ClearSession(w)
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), contextKeyUser, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isAuthenticated(ctx context.Context) bool {
	user, ok := userFromContext(ctx)
	if !ok {
		return false
	}
	return !user.IsEmpty()
}

func userFromContext(ctx context.Context) (screenjournal.UserAuth, bool) {
	user, ok := ctx.Value(contextKeyUser).(screenjournal.UserAuth)
	if !ok {
		return screenjournal.UserAuth{}, false
	}
	return user, true
}
