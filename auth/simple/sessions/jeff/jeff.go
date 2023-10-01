package jeff

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/mtlynch/jeff"
	"github.com/mtlynch/jeff/sqlite"

	"github.com/mtlynch/screenjournal/v2/auth/simple/sessions"
)

type (
	manager struct {
		j *jeff.Jeff
	}

	session struct {
		metadata sessions.Metadata
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

func (m manager) CreateSession(w http.ResponseWriter, r *http.Request, meta sessions.Metadata) error {
	b, err := serializeMetadata(meta)
	if err != nil {
		return err
	}
	return m.j.Set(r.Context(), w, []byte(meta.Username), b)
}

func (m manager) SessionFromRequest(r *http.Request) (sessions.Session, error) {
	sess := jeff.ActiveSession(r.Context())
	if len(sess.Key) == 0 {
		return nil, sessions.ErrNotAuthenticated
	}

	meta, err := deserializeMetadata(sess.Meta)
	if err != nil {
		return nil, err
	}

	return session{
		metadata: meta,
	}, nil
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

func (s session) Metadata() sessions.Metadata {
	return s.Metadata()
}

func serializeMetadata(meta sessions.Metadata) ([]byte, error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(meta); err != nil {
		log.Fatalf("failed to serialize session metadata to JSON: %v", err)
	}
	return b.Bytes(), nil
}

func deserializeMetadata(b []byte) (sessions.Metadata, error) {
	var meta sessions.Metadata
	if err := json.NewDecoder(bytes.NewReader(b)).Decode(&meta); err != nil {
		return sessions.Metadata{}, err
	}

	return meta, nil
}
