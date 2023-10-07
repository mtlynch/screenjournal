package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2/auth/simple/sessions"
	jeff_sessions "github.com/mtlynch/screenjournal/v2/auth/simple/sessions/jeff"

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
		inner sessions.Manager
	}

	serializableSession struct {
		Username string `json:"username"`
		IsAdmin  bool   `json:"isAdmin"`
	}
)

func NewSessionManager(dbPath string) (sessionManager, error) {
	inner, err := jeff_sessions.New(dbPath)
	if err != nil {
		log.Fatalf("failed to create session manager: %v", err)
	}
	return sessionManager{
		inner: inner,
	}, nil
}

func (sm sessionManager) CreateSession(w http.ResponseWriter, ctx context.Context, username screenjournal.Username, isAdmin bool) error {
	if err := sm.inner.CreateSession(w, ctx, sessionKeyFromUsername(username), serializeSession(Session{
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
