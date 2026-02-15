package quiet

import (
	"log"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type Resetter struct{}

func New() Resetter {
	return Resetter{}
}

func (r Resetter) RequestReset(user screenjournal.User) error {
	log.Printf("skipping password reset email for %s because no email sender is configured", user.Username)
	return nil
}
