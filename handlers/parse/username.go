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

func Usernames(usernames []string) ([]screenjournal.Username, error) {
	parsed := []screenjournal.Username{}
	for _, id := range usernames {
		p, err := Username(id)
		if err != nil {
			return []screenjournal.Username{}, err
		}
		parsed = append(parsed, p)
	}
	return parsed, nil
}
