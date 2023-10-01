package jeff

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/mtlynch/jeff"
	"github.com/mtlynch/jeff/sqlite"

	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
)

type (
	manager struct {
		j *jeff.Jeff
	}

	serializableUser struct {
		Username string `json:"username"`
		IsAdmin  bool   `json:"isAdmin"`
	}
)

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

func (m manager) CreateSession(w http.ResponseWriter, r *http.Request, key sessions.Key, session sessions.Session) error {
	return m.j.Set(r.Context(), w, key.Bytes(), session)
}

func (m manager) SessionFromRequest(r *http.Request) (sessions.Session, error) {
	sess := jeff.ActiveSession(r.Context())
	if len(sess.Key) == 0 {
		return sessions.Session{}, sessions.ErrNotAuthenticated
	}

	return sess.Meta, nil
}

func (m manager) EndSession(r *http.Request, w http.ResponseWriter) {
	sess := jeff.ActiveSession(r.Context())
	if len(sess.Key) > 0 {
		if err := m.j.Delete(r.Context(), sess.Key); err != nil {
			log.Printf("failed to delete session: %v", err)
		}
	}

	if err := m.j.Clear(r.Context(), w); err != nil {
		log.Printf("failed to clear session: %v", err)
	}
}

func (m manager) WrapRequest(next http.Handler) http.Handler {
	return m.j.Public(next)
}
