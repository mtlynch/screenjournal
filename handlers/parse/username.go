package parse

import (
	"errors"
	"regexp"

	"github.com/mtlynch/screenjournal/v2"
)

var (
	ErrInvalidUsername = errors.New("invalid username")

	usernamePattern = regexp.MustCompile(`^[a-zA-Z\.0-9]{2,80}$`)
)

func Username(username string) (screenjournal.Username, error) {
	if isReservedWord(username) {
		return screenjournal.Username(""), ErrInvalidUsername
	}

	if !usernamePattern.MatchString(username) {
		return screenjournal.Username(""), ErrInvalidUsername
	}

	return screenjournal.Username(username), nil
}
