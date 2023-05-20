package simple

import (
	"log"

	"github.com/mtlynch/screenjournal/v2/auth"
)

type (
	AuthStore interface {
		//InsertUser(username, password string) error
		ReadPasswordHash(username string) (auth.PasswordHash, error)
	}

	authenticator struct {
		store AuthStore
	}
)

func New(store AuthStore) auth.Authenticator {
	return authenticator{
		store: store,
	}
}

func (a authenticator) Authenticate(username, password string) error {
	h, err := a.store.ReadPasswordHash(username)
	if err != nil {
		return err
	}

	log.Printf("attempted password = %s", password) // DEBUG
	log.Printf("existing hash = %s", h.String())    // DEBUG

	if err := h.MatchesPlaintext(password); err != nil {
		log.Printf("plaintext doesn't match, returning %s", err) // DEBUG
		return err
	}

	log.Printf("plaintext matches!") // DEBUG

	return nil
}
