package sessions

import (
	"context"
	"database/sql"
	"time"

	simple_sessions "codeberg.org/mtlynch/simpleauth/v3/sessions"

	"github.com/mtlynch/screenjournal/v2/store/sqlite"
)

const sessionLifetime = 30 * 24 * time.Hour

// NewManager creates a session manager backed by SQLite.
//
// The `dbFunc` parameter allows callers to choose the database at request time
// from the context rather than binding the manager to a single `*sql.DB` at
// startup. We need this for e2e tests because the e2e tests rely on the session
// manager having several SQLite databases at once (one for each parallel test)
// that it needs to resolve on a per-request basis.
func NewManager(dbFunc func(context.Context) *sql.DB, requireTls bool) simple_sessions.Manager {
	return simple_sessions.NewManager(simple_sessions.Config{
		Store:      sqlite.NewSessionStore(dbFunc),
		RequireTLS: requireTls,
		Now:        time.Now,
		Lifetime:   sessionLifetime,
	})
}
