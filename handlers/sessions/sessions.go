package sessions

import (
	"time"

	simple_sessions "codeberg.org/mtlynch/simpleauth/v3/sessions"
)

const sessionLifetime = 30 * 24 * time.Hour

func NewManager(store simple_sessions.Store, requireTls bool) simple_sessions.Manager {
	return simple_sessions.NewManager(simple_sessions.Config{
		Store:      store,
		RequireTLS: requireTls,
		Now:        time.Now,
		Lifetime:   sessionLifetime,
	})
}
