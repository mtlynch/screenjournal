package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/auth"
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

func (s Server) populateAuthenticationContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := s.authenticator.Authenticate(r)
		if err == auth.ErrNotAuthenticated {
			next.ServeHTTP(w, r)
			return
		} else if err != nil {
			s.authenticator.ClearSession(w)
			http.Error(w, "Invalid username", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), contextKeyUser, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s Server) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := userFromContext(r.Context()); !ok {
			s.authenticator.ClearSession(w)
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s Server) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't require admin on HTTP OPTIONS requests
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		if !isAdmin(r.Context()) {
			log.Printf("attempt to perform admin action by non-admin user: %v", usernameFromContext(r.Context()))
			http.Error(w, "You must be an administrative user to perform this action", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isAdmin(ctx context.Context) bool {
	user, ok := userFromContext(ctx)
	if !ok {
		return false
	}

	return user.IsAdmin
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

func usernameFromContext(ctx context.Context) screenjournal.Username {
	user, ok := userFromContext(ctx)
	if !ok {
		return screenjournal.Username("")
	}
	return user.Username
}
