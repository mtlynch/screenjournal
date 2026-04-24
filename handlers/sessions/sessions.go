package sessions

import (
	"context"
	"database/sql"
	"time"

	simple_sessions "codeberg.org/mtlynch/simpleauth/v3/sessions"
)

const sessionLifetime = 30 * 24 * time.Hour

func NewManager(dbFunc func(context.Context) *sql.DB, requireTls bool) simple_sessions.Manager {
	return simple_sessions.NewManager(simple_sessions.Config{
		Store:      store{dbFunc: dbFunc},
		RequireTLS: requireTls,
		Now:        time.Now,
		Lifetime:   sessionLifetime,
	})
}
