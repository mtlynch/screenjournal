package jeff

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/mtlynch/jeff"
	"github.com/mtlynch/jeff/sqlite"

	"github.com/mtlynch/screenjournal/v2/auth/simple/sessions"
)

type manager struct {
	j *jeff.Jeff
}

func New(dbPath string) (sessions.Manager, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return manager{}, err
	}
	store, err := sqlite.New(db)
	if err != nil {
		return manager{}, err
	}
	options := []func(*jeff.Jeff){jeff.CookieName("token")}
	options = append(options, extraOptions()...)
	j := jeff.New(store, options...)
	return manager{
		j: j,
	}, nil
}

func (m manager) CreateSession(w http.ResponseWriter, ctx context.Context, key sessions.Key, session sessions.Session) error {
	return m.j.Set(ctx, w, key.Bytes(), session)
}

func (m manager) SessionFromContext(ctx context.Context) (sessions.Session, error) {
	sess := jeff.ActiveSession(ctx)
	if len(sess.Key) == 0 {
		return sessions.Session{}, sessions.ErrNoSessionFound
	}

	return sess.Meta, nil
}

func (m manager) EndSession(ctx context.Context, w http.ResponseWriter) {
	sess := jeff.ActiveSession(ctx)
	if len(sess.Key) > 0 {
		if err := m.j.Delete(ctx, sess.Key); err != nil {
			log.Printf("failed to delete session: %v", err)
		}
	}

	if err := m.j.Clear(ctx, w); err != nil {
		log.Printf("failed to clear session: %v", err)
	}
}

func (m manager) WrapRequest(next http.Handler) http.Handler {
	return m.j.Public(next)
}
