package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
)

type contextKey struct {
	name string
}

var contextKeyUser = &contextKey{"user"}

func (s Server) authPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, err := credentialsFromRequest(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid credentials: %v", err), http.StatusBadRequest)
			return
		}

		if err := s.authenticator.Authenticate(username, password); err != nil {
			http.Error(w, fmt.Sprintf("Invalid credentials: %v", err), http.StatusUnauthorized)
			return
		}

		if err := s.sessionManager.CreateSession(w, r, username); err != nil {
			http.Error(w, fmt.Sprintf("Failed to create session: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) authDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.sessionManager.EndSession(r, w); err != nil {
			log.Printf("error terminating user session: %v", err)
		}
	}
}

func (s Server) populateAuthenticationContext(next http.Handler) http.Handler {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, err := s.sessionManager.SessionFromRequest(r)
		if err != nil {
			if err == sessions.ErrNotAuthenticated {
				next.ServeHTTP(w, r)
				return
			} else {
				http.Error(w, fmt.Sprintf("Invalid session token: %v", err), http.StatusBadRequest)
				return
			}
		}

		ctx := context.WithValue(r.Context(), contextKeyUser, sess.UserAuth)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
	return s.sessionManager.WrapRequest(fn)
}

func (s Server) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := userFromContext(r.Context()); !ok {
			if err := s.sessionManager.EndSession(r, w); err != nil {
				log.Printf("error terminating user session: %v", err)
			}
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

func credentialsFromRequest(r *http.Request) (screenjournal.Username, screenjournal.Password, error) {
	body := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&body); err != nil {
		return screenjournal.Username(""), screenjournal.Password(""), err
	}

	username, err := parse.Username(body.Username)
	if err != nil {
		return screenjournal.Username(""), screenjournal.Password(""), err
	}

	password, err := parse.Password(body.Password)
	if err != nil {
		return screenjournal.Username(""), screenjournal.Password(""), err
	}

	return username, password, nil
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
