package parse

import (
	"net/mail"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

// Email parses a raw email string into an Email object, validating that it's
// well-formed.
func Email(email string) (screenjournal.Email, error) {
	a, err := mail.ParseAddress(email)
	if err != nil {
		return screenjournal.Email(""), err
	}
	return screenjournal.Email(a.Address), nil
}
