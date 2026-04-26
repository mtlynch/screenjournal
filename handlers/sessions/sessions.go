package sessions

import (
	"context"
	"time"

	simple_sessions "codeberg.org/mtlynch/simpleauth/v3/sessions"

	"github.com/mtlynch/screenjournal/v2/store/sqlite"
)

const sessionLifetime = 30 * 24 * time.Hour

// NewManager creates a session manager backed by SQLite.
//
// The `storeFunc` parameter allows callers to choose the database at request
// time from the context rather than binding the manager to a single store at
// startup. Much of the app already resolves the database per request at its own
// callsites, but the session manager is constructed once up front and performs
// its own store operations later through simpleauth. This callback preserves
// request-time database selection for e2e tests that isolate parallel sessions
// into separate SQLite databases.
func NewManager(storeFunc func(context.Context) sqlite.Store, requireTls bool) simple_sessions.Manager {
	return simple_sessions.NewManager(simple_sessions.Config{
		Store:      sqlite.NewSessionStore(storeFunc),
		RequireTLS: requireTls,
		Now:        time.Now,
		Lifetime:   sessionLifetime,
	})
}
