package jeff

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/mtlynch/jeff"
	"github.com/mtlynch/jeff/sqlite"

	"github.com/mtlynch/screenjournal/v2/auth/simple"
	"github.com/mtlynch/screenjournal/v2/auth/simple/sessions"
)

type (
	manager struct {
		j *jeff.Jeff
	}

	serializableUser struct {
		Username_ string `json:"username"`
		IsAdmin_  bool   `json:"isAdmin"`
	}

	session struct {
		user simple.User
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

func (m manager) CreateSession(w http.ResponseWriter, r *http.Request, user simple.User) error {
	meta, err := serializeUser(user)
	if err != nil {
		return err
	}
	return m.j.Set(r.Context(), w, []byte(user.Username()), meta)
}

func (m manager) SessionFromRequest(r *http.Request) (sessions.Session, error) {
	sess := jeff.ActiveSession(r.Context())
	if len(sess.Key) == 0 {
		return nil, sessions.ErrNotAuthenticated
	}

	user, err := deserializeUser(sess.Meta)
	if err != nil {
		return nil, err
	}

	return session{
		user: user,
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

func (s session) User() simple.User {
	return s.user
}

func (u serializableUser) Username() string {
	return u.Username_
}

func (u serializableUser) IsAdmin() bool {
	return u.IsAdmin_
}

func serializeUser(user simple.User) ([]byte, error) {
	su := serializableUser{
		Username_: user.Username(),
		IsAdmin_:  user.IsAdmin(),
	}
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(su); err != nil {
		log.Fatalf("failed to serialize user to JSON: %v", err)
	}
	return b.Bytes(), nil
}

func deserializeUser(b []byte) (simple.User, error) {
	var su serializableUser
	if err := json.NewDecoder(bytes.NewReader(b)).Decode(&su); err != nil {
		return nil, err
	}

	return &su, nil
}
