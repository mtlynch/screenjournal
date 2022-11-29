package gorilla

import (
	"crypto/rand"
	"errors"
	"log"
	"net/http"

	gorilla "github.com/gorilla/sessions"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
)

const cookieName = "user-token"
const usernameKey = "username"

type (
	wrapper struct {
		store gorilla.Store
	}
)

func New() (sessions.Manager, error) {
	return wrapper{
		//store: gorilla.NewFilesystemStore("/tmp/foo1", randBytes(32)),
		store: gorilla.NewCookieStore([]byte("dummywubby")),
	}, nil
}

func (wrapper wrapper) CreateSession(w http.ResponseWriter, r *http.Request, username screenjournal.Username) error {
	session, err := wrapper.store.Get(r, cookieName)
	if err != nil {
		log.Printf("couldn't get session store: %v", err)
		return err
	}

	session.Values[usernameKey] = username.String()

	if err := session.Save(r, w); err != nil {
		log.Printf("couldn't save session store: %v", err)
		return err
	}

	return nil
}

func (wrapper wrapper) SessionFromRequest(r *http.Request) (sessions.Session, error) {
	session, err := wrapper.store.Get(r, cookieName)
	if err != nil {
		return sessions.Session{}, err
	}

	v, ok := session.Values[usernameKey]
	if !ok {
		return sessions.Session{}, sessions.ErrNotAuthenticated
	}

	usernameRaw, ok := v.(string)
	if !ok {
		return sessions.Session{}, errors.New("unexpected type") // TODO
	}

	log.Printf("raw username = %v", usernameRaw)

	return sessions.Session{
		UserAuth: screenjournal.UserAuth{
			Username: screenjournal.Username(usernameRaw),
		},
	}, nil
}

func (wrapper wrapper) EndSession(r *http.Request, w http.ResponseWriter) error {
	session, err := wrapper.store.Get(r, cookieName)
	if err != nil {
		log.Printf("couldn't get session while terminating session") // DEBUG
		return err
	}
	session.Options.MaxAge = 0
	//return session.Save(r, w)
	return nil
}

func randBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return b
}
