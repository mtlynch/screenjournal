package tmdb

import (
	"errors"
	"math"
	"regexp"

	"github.com/mtlynch/screenjournal/v2"
)

var (
	ErrInvalidTmdbID = errors.New("invalid TMDB ID")
	ErrInvalidImdbID = errors.New("invalid IMDB ID")

	imdbIDPattern = regexp.MustCompile(`^tt[0-9]{7,8}$`)
)

func ParseTmdbID(raw int) (screenjournal.TmdbID, error) {
	if raw <= 0 || raw > math.MaxInt32 {
		return screenjournal.TmdbID(0), ErrInvalidTmdbID
	}

	return screenjournal.TmdbID(raw), nil
}

func ParseImdbID(raw string) (screenjournal.ImdbID, error) {
	if !imdbIDPattern.MatchString(raw) {
		return screenjournal.ImdbID(""), ErrInvalidImdbID
	}
	return screenjournal.ImdbID(raw), nil
}
