package tmdb

import (
	"errors"
	"regexp"
	"time"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

var (
	ErrInvalidImdbID      = errors.New("invalid IMDB ID")
	ErrInvalidReleaseDate = errors.New("invalid release date")

	imdbIDPattern = regexp.MustCompile(`^tt[0-9]{7,8}$`)
)

func ParseImdbID(raw string) (screenjournal.ImdbID, error) {
	if !imdbIDPattern.MatchString(raw) {
		return screenjournal.ImdbID(""), ErrInvalidImdbID
	}
	return screenjournal.ImdbID(raw), nil
}

func ParseReleaseDate(raw string) (screenjournal.ReleaseDate, error) {
	t, err := time.Parse(time.DateOnly, raw)
	if err != nil {
		return screenjournal.ReleaseDate{}, ErrInvalidReleaseDate
	}
	return screenjournal.ReleaseDate(t), nil
}
