package jeff

import (
	"net/http"

	"github.com/abraithwaite/jeff"
	"github.com/abraithwaite/jeff/memory"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
)

type (
	manager struct {
		j *jeff.Jeff
	}
)

func New() (sessions.Manager, error) {
	store := memory.New()
	options := []func(*jeff.Jeff){jeff.CookieName("token")}
	options = append(options, extraOptions()...)
	j := jeff.New(store, options...)
	return manager{
		j: j,
	}, nil
}

func (m manager) CreateSession(w http.ResponseWriter, r *http.Request, username screenjournal.Username) error {
	return m.j.Set(r.Context(), w, []byte(username.String()))
}

func (m manager) SessionFromRequest(r *http.Request) (sessions.Session, error) {
	sess := jeff.ActiveSession(r.Context())
	if len(sess.Key) == 0 {
		return sessions.Session{}, sessions.ErrNotAuthenticated
	}

	return sessions.Session{
		UserAuth: screenjournal.UserAuth{
			Username: screenjournal.Username(string(sess.Key)),
			IsAdmin:  true, // TODO: Determine this dynamicaly
		},
	}, nil
}

func (m manager) EndSession(_ *http.Request, w http.ResponseWriter) error {
	return nil
}

func (m manager) WrapRequest(next http.Handler) http.Handler {
	return m.j.Public(next)
}
