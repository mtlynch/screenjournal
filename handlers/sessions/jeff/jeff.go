package jeff

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/abraithwaite/jeff"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
	"github.com/mtlynch/screenjournal/v2/handlers/sessions/jeff/sqlite_store"
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
	store, err := sqlite_store.New(db)
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

func (m manager) CreateSession(w http.ResponseWriter, r *http.Request, user screenjournal.User) error {
	meta, err := serializeUser(user)
	if err != nil {
		return err
	}
	return m.j.Set(r.Context(), w, []byte(user.Username.String()), meta)
}

func (m manager) SessionFromRequest(r *http.Request) (sessions.Session, error) {
	sess := jeff.ActiveSession(r.Context())
	if len(sess.Key) == 0 {
		return sessions.Session{}, sessions.ErrNotAuthenticated
	}

	user, err := deserializeUser(sess.Meta)
	if err != nil {
		return sessions.Session{}, err
	}

	return sessions.Session{
		User: user,
	}, nil
}

func (m manager) EndSession(r *http.Request, w http.ResponseWriter) error {
	sess := jeff.ActiveSession(r.Context())
	if len(sess.Key) == 0 {
		return nil
	}

	return m.j.Delete(r.Context(), sess.Key)
}

func (m manager) WrapRequest(next http.Handler) http.Handler {
	return m.j.Public(next)
}

func serializeUser(user screenjournal.User) ([]byte, error) {
	su := serializableUser{
		Username: user.Username.String(),
		IsAdmin:  user.IsAdmin,
	}
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(su); err != nil {
		log.Fatalf("failed to serialize user to JSON: %v", err)
	}
	return b.Bytes(), nil
}

func deserializeUser(b []byte) (screenjournal.User, error) {
	var su serializableUser
	if err := json.NewDecoder(bytes.NewReader(b)).Decode(&su); err != nil {
		return screenjournal.User{}, err
	}

	return screenjournal.User{
		Username: screenjournal.Username(su.Username),
		IsAdmin:  su.IsAdmin,
	}, nil
}
