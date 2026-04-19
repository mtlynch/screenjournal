package sessions

import (
	"database/sql"
	"time"

	simple_sessions "codeberg.org/mtlynch/simpleauth/v3/sessions"
)

func NewManager(db *sql.DB, requireTls bool) simple_sessions.Manager {
	return simple_sessions.NewManager(simple_sessions.Config{
		Store:      store{db: db},
		RequireTLS: requireTls,
		Now:        time.Now,
	})
}
