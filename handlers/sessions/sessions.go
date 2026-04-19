package sessions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"codeberg.org/mtlynch/simpleauth/v3"
	simple_sessions "codeberg.org/mtlynch/simpleauth/v3/sessions"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type (
	Session struct {
		Username screenjournal.Username
		IsAdmin  bool
	}

	Manager struct {
		inner simple_sessions.Manager
	}
)

var ErrNoSessionFound = errors.New("no session in request context")

func NewManager(db *sql.DB, requireTls bool) (Manager, error) {
	store, err := newStore(db)
	if err != nil {
		return Manager{}, err
	}
	return Manager{
		inner: simple_sessions.NewManager(simple_sessions.Config{
			Store:      store,
			RequireTLS: requireTls,
			Now:        time.Now,
		}),
	}, nil
}

func (sm Manager) CreateSession(w http.ResponseWriter, ctx context.Context, username screenjournal.Username, isAdmin bool) error {
	userID, err := simpleauth.NewUserID(username.String())
	if err != nil {
		return err
	}
	if err := sm.inner.CreateSession(ctx, w, simpleauth.User{
		ID: userID,
		SessionData: json.RawMessage(serializeSession(Session{
			Username: username,
			IsAdmin:  isAdmin,
		})),
	}); err != nil {
		return err
	}
	return nil
}

func (sm Manager) SessionFromContext(ctx context.Context) (Session, error) {
	sess, err := sm.inner.SessionFromContext(ctx)
	if err != nil {
		// Wrap the third-party error with a local one.
		if errors.Is(err, simple_sessions.ErrNoSessionFound) {
			return Session{}, ErrNoSessionFound
		}
		return Session{}, err
	}

	session, err := deserializeSession(sess.SessionData)
	if err != nil {
		return Session{}, err
	}

	return session, nil
}

func (sm Manager) EndSession(ctx context.Context, w http.ResponseWriter) error {
	return sm.inner.EndSession(ctx, w)
}

func (sm Manager) WrapRequest(h http.Handler) http.Handler {
	return sm.inner.WrapRequest(h)
}
