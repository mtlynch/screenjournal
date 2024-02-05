package sessions

import (
	"context"
	"database/sql"
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

	Manager struct {
		inner simple_sessions.Manager
	}
)

var ErrNoSessionFound = errors.New("no session in request context")

func NewManager(dbPath string) (Manager, error) {
	// We could theoretically share the DB pointer with the main DB store, if we
	// do the sql.Open call in main.go and share it.
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return Manager{}, err
	}

	simpleManager, err := simple_sessions.NewManager(db)
	if err != nil {
		log.Fatalf("failed to create session manager: %v", err)
	}
	return Manager{
		inner: simpleManager,
	}, nil
}

func (sm Manager) CreateSession(w http.ResponseWriter, ctx context.Context, username screenjournal.Username, isAdmin bool) error {
	key := simple_sessions.KeyFromBytes([]byte(username.String()))
	if err := sm.inner.CreateSession(w, ctx, key, serializeSession(Session{
		Username: username,
		IsAdmin:  isAdmin,
	})); err != nil {
		return err
	}
	return nil
}

func (sm Manager) SessionFromContext(ctx context.Context) (Session, error) {
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

func (sm Manager) EndSession(ctx context.Context, w http.ResponseWriter) {
	sm.inner.EndSession(ctx, w)
}

func (sm Manager) WrapRequest(h http.Handler) http.Handler {
	return sm.inner.WrapRequest(h)
}
