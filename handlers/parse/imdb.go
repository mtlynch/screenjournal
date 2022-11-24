package parse

import (
	"errors"
	"regexp"

	"github.com/mtlynch/screenjournal/v2"
)

var (
	ErrInvalidImdbID = errors.New("invalid IMDB ID")

	imdbIDPattern = regexp.MustCompile(`^((tt)|(nm)|(co)|(ev)|(ch)|(ni))[0-9]{7}$`)
)

func ImdbID(raw string) (screenjournal.ImdbID, error) {
	if !imdbIDPattern.MatchString(raw) {
		return screenjournal.ImdbID(""), ErrInvalidImdbID
	}
	return screenjournal.ImdbID(raw), nil
}
