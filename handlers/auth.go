package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/mtlynch/screenjournal/v2/auth/simple/sessions"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type contextKey struct {
	name string
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

		if err := s.sessionManager.CreateSession(w, r, sessionKeyFromUsername(user.Username), SerializeSession(Session{
			Username: user.Username,
			IsAdmin:  user.IsAdmin,
		})); err != nil {
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
		b, err := s.sessionManager.SessionFromRequest(r)
		if err != nil {
			if err != sessions.ErrNotAuthenticated {
				log.Printf("invalid session token: %v", err)
			}
			s.sessionManager.EndSession(r, w)
			next.ServeHTTP(w, r)
			return
		}

		session, err := DeserializeSession(b)
		if err != nil {
			log.Printf("failed to deserialize session: %v", err)
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), contextKeySession, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	}))
}

func (s Server) requireAuthenticationForAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := sessionFromContext(r.Context()); !ok {
			s.sessionManager.EndSession(r, w)
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s Server) requireAuthenticationForView(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := sessionFromContext(r.Context()); !ok {
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

func sessionFromContext(ctx context.Context) (Session, bool) {
	session, ok := ctx.Value(contextKeySession).(Session)
	if !ok {
		return Session{}, false
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

func sessionKeyFromUsername(username screenjournal.Username) sessions.Key {
	return sessions.KeyFromBytes([]byte(username.String()))
}
