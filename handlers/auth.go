package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	simple_sessions "codeberg.org/mtlynch/simpleauth/v3/sessions"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
)

type contextKey struct {
	name string
}

type authenticatedSession struct {
	Username screenjournal.Username
	IsAdmin  bool
}

var contextKeySession = &contextKey{"user"}

func (s Server) authPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, err := credentialsFromRequest(r)
		if err != nil {
			log.Printf("invalid auth request: %v", err)
			http.Error(w, fmt.Sprintf("Invalid credentials: %v", err), http.StatusBadRequest)
			return
		}

		if err := s.getAuthenticator(r).Authenticate(username, password); err != nil {
			log.Printf("auth failed for user %s: %v", username, err)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		user, err := s.getDB(r).ReadUser(username)
		if err != nil {
			log.Printf("failed to read user data from database %s: %v", username, err)
			http.Error(w, "Failed to read user information", http.StatusInternalServerError)
			return
		}

		userID, err := userIDFromUsername(user.Username)
		if err != nil {
			log.Printf("failed to create user ID for user %s: %v", user.Username.String(), err)
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}

		if err := s.sessionManager.LogIn(r.Context(), w, userID); err != nil {
			log.Printf("failed to create session for user %s: %v", user.Username.String(), err)
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) authDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.sessionManager.LogOut(r.Context(), w); err != nil {
			log.Printf("failed to end session: %v", err)
			http.Error(w, "Failed to end session", http.StatusInternalServerError)
		}
	}
}

func (s Server) populateAuthenticationContext(next http.Handler) http.Handler {
	return s.sessionManager.LoadUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := s.sessionManager.UserIDFromContext(r.Context())
		if err != nil {
			if !errors.Is(err, simple_sessions.ErrNoSessionFound) {
				log.Printf("invalid session token: %v", err)
			}
			if err := s.sessionManager.LogOut(r.Context(), w); err != nil {
				log.Printf("failed to end invalid session: %v", err)
				http.Error(w, "Failed to end session", http.StatusInternalServerError)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		username := screenjournal.Username(userID.String())
		user, err := s.getDB(r).ReadUser(username)
		if err != nil {
			if errors.Is(err, store.ErrUserNotFound) {
				if err := s.sessionManager.LogOut(r.Context(), w); err != nil {
					log.Printf("failed to end session for missing user: %v", err)
					http.Error(w, "Failed to end session", http.StatusInternalServerError)
					return
				}
				next.ServeHTTP(w, r)
				return
			}
			log.Printf("failed to read session user: %v", err)
			http.Error(w, "Failed to read user information", http.StatusInternalServerError)
			return
		}

		session := authenticatedSession{
			Username: user.Username,
			IsAdmin:  user.IsAdmin,
		}
		ctx := context.WithValue(r.Context(), contextKeySession, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	}))
}

func (s Server) requireAuthenticationForAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := sessionFromContext(r.Context()); !ok {
			if err := s.sessionManager.LogOut(r.Context(), w); err != nil {
				log.Printf("failed to end unauthenticated session: %v", err)
				http.Error(w, "Failed to end session", http.StatusInternalServerError)
				return
			}
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s Server) requireAuthenticationForView(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := sessionFromContext(r.Context()); !ok {
			if err := s.sessionManager.LogOut(r.Context(), w); err != nil {
				log.Printf("failed to end unauthenticated session: %v", err)
				http.Error(w, "Failed to end session", http.StatusInternalServerError)
				return
			}

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
			log.Printf("attempt to perform admin action by non-admin user: %v", mustGetUsernameFromContext(r.Context()))
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

func userIDFromUsername(username screenjournal.Username) (simple_sessions.UserID, error) {
	return simple_sessions.NewUserID(username.String())
}

func isAdmin(ctx context.Context) bool {
	sess, ok := sessionFromContext(ctx)
	if !ok {
		return false
	}

	return sess.IsAdmin
}

func isAuthenticated(ctx context.Context) bool {
	_, ok := sessionFromContext(ctx)
	return ok
}

func sessionFromContext(ctx context.Context) (authenticatedSession, bool) {
	session, ok := ctx.Value(contextKeySession).(authenticatedSession)
	if !ok {
		return authenticatedSession{}, false
	}
	return session, true
}

func usernameFromContext(ctx context.Context) (screenjournal.Username, bool) {
	sess, ok := sessionFromContext(ctx)
	if !ok {
		return screenjournal.Username(""), false
	}
	return sess.Username, true
}

func mustGetUsernameFromContext(ctx context.Context) screenjournal.Username {
	username, ok := usernameFromContext(ctx)
	if !ok {
		panic("No session in context in an authenticated handler")
	}
	return username
}
