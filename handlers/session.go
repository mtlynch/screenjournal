package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	simple_sessions "github.com/mtlynch/simpleauth/v2/sessions"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type (
	Session struct {
		Username screenjournal.Username
		IsAdmin  bool
	}

	SessionManager interface {
		CreateSession(http.ResponseWriter, context.Context, screenjournal.Username, bool) error
		SessionFromContext(context.Context) (Session, error)
		EndSession(context.Context, http.ResponseWriter)
		// WrapRequest wraps the given handler, adding the Session object (if
		// there's an active session) to the request context before passing control
		// to the next handler.
		WrapRequest(http.Handler) http.Handler
	}

	sessionManager struct {
		inner simple_sessions.Manager
	}

	serializableSession struct {
		Username string `json:"username"`
		IsAdmin  bool   `json:"isAdmin"`
	}
)

var ErrNoSessionFound = errors.New("no session in request context")

func NewSessionManager(dbPath string) (sessionManager, error) {
	inner, err := simple_sessions.New(dbPath)
	if err != nil {
		log.Fatalf("failed to create session manager: %v", err)
	}
	return sessionManager{
		inner: inner,
	}, nil
}

func (sm sessionManager) CreateSession(w http.ResponseWriter, ctx context.Context, username screenjournal.Username, isAdmin bool) error {
	key := simple_sessions.KeyFromBytes([]byte(username.String()))
	if err := sm.inner.CreateSession(w, ctx, key, serializeSession(Session{
		Username: username,
		IsAdmin:  isAdmin,
	})); err != nil {
		return err
	}
	return nil
}

func (sm sessionManager) SessionFromContext(ctx context.Context) (Session, error) {
	b, err := sm.inner.SessionFromContext(ctx)
	if err != nil {
		// Wrap the third-party error with a local one.
		if err == simple_sessions.ErrNoSessionFound {
			return Session{}, ErrNoSessionFound
		}
		return Session{}, err
	}

	session, err := deserializeSession(b)
	if err != nil {
		return Session{}, err
	}

	return session, nil
}

func (sm sessionManager) EndSession(ctx context.Context, w http.ResponseWriter) {
	sm.inner.EndSession(ctx, w)
}

func (sm sessionManager) WrapRequest(h http.Handler) http.Handler {
	return sm.inner.WrapRequest(h)
}

func serializeSession(sess Session) []byte {
	ss := serializableSession{
		Username: sess.Username.String(),
		IsAdmin:  sess.IsAdmin,
	}
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(ss); err != nil {
		log.Fatalf("failed to serialize session to JSON: %v", err)
	}
	return b.Bytes()
}

func deserializeSession(b []byte) (Session, error) {
	var ss serializableSession
	if err := json.NewDecoder(bytes.NewReader(b)).Decode(&ss); err != nil {
		return Session{}, err
	}

	return Session{
		Username: screenjournal.Username(ss.Username),
		IsAdmin:  ss.IsAdmin,
	}, nil
}
