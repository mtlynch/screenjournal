package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

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
			log.Printf("invalid auth request: %v", err)
			http.Error(w, fmt.Sprintf("Invalid credentials: %v", err), http.StatusBadRequest)
			return
		}

		user, err := s.authenticator.Authenticate(username, password)
		if err != nil {
			log.Printf("auth failed for user %s: %v", username, err)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		if err := s.sessionManager.CreateSession(w, r, user); err != nil {
			log.Printf("failed to create session for user %s: %v", user.Username.String(), err)
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) authDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.sessionManager.EndSession(r, w)
	}
}

func (s Server) populateAuthenticationContext(next http.Handler) http.Handler {
	return s.sessionManager.WrapRequest(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := s.sessionManager.SessionFromRequest(r)
		if err != nil {
			if err != sessions.ErrNotAuthenticated {
				log.Printf("invalid session token: %v", err)
			}
			s.sessionManager.EndSession(r, w)
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), contextKeyUser, session.User)
		next.ServeHTTP(w, r.WithContext(ctx))
	}))
}

func (s Server) requireAuthenticationForAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := userFromContext(r.Context()); !ok {
			s.sessionManager.EndSession(r, w)
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s Server) requireAuthenticationForView(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := userFromContext(r.Context()); !ok {
			s.sessionManager.EndSession(r, w)

			newURL := "/login?next=" + url.QueryEscape(r.URL.String())
			http.Redirect(w, r, newURL, http.StatusTemporaryRedirect)
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

func userFromContext(ctx context.Context) (screenjournal.User, bool) {
	user, ok := ctx.Value(contextKeyUser).(screenjournal.User)
	if !ok {
		return screenjournal.User{}, false
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
