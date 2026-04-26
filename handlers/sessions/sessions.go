package sessions

import (
	"time"

	simple_sessions "codeberg.org/mtlynch/simpleauth/v3/sessions"

	"github.com/mtlynch/screenjournal/v2/store/sqlite"
)

const sessionLifetime = 30 * 24 * time.Hour

// NewManager creates a session manager backed by SQLite.
func NewManager(store sqlite.Store, requireTls bool) simple_sessions.Manager {
	return simple_sessions.NewManager(simple_sessions.Config{
		Store:      store,
		RequireTLS: requireTls,
		Now:        time.Now,
		Lifetime:   sessionLifetime,
	})
}
