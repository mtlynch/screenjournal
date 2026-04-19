package sessions

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	simple_sessions "codeberg.org/mtlynch/simpleauth/v3/sessions"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type Manager struct {
	inner simple_sessions.Manager
}

var ErrNoSessionFound = simple_sessions.ErrNoSessionFound

func NewManager(db *sql.DB, requireTls bool) Manager {
	return Manager{
		inner: simple_sessions.NewManager(simple_sessions.Config{
			Store:      store{db: db},
			RequireTLS: requireTls,
			Now:        time.Now,
		}),
	}
}

func (sm Manager) CreateSession(w http.ResponseWriter, ctx context.Context, username screenjournal.Username) error {
	userID, err := simple_sessions.NewUserID(username.String())
	if err != nil {
		return err
	}
	if err := sm.inner.CreateSession(ctx, w, userID); err != nil {
		return err
	}
	return nil
}

func (sm Manager) UsernameFromContext(ctx context.Context) (screenjournal.Username, error) {
	sess, err := sm.inner.SessionFromContext(ctx)
	if err != nil {
		return screenjournal.Username(""), err
	}

	return screenjournal.Username(sess.UserID.String()), nil
}

func (sm Manager) EndSession(ctx context.Context, w http.ResponseWriter) error {
	return sm.inner.EndSession(ctx, w)
}

func (sm Manager) WrapRequest(h http.Handler) http.Handler {
	return sm.inner.WrapRequest(h)
}
