package tmdb

import (
	"errors"
	"math"
	"regexp"
	"time"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

var (
	ErrInvalidTmdbID      = errors.New("invalid TMDB ID")
	ErrInvalidImdbID      = errors.New("invalid IMDB ID")
	ErrInvalidReleaseDate = errors.New("invalid release date")

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

func ParseReleaseDate(raw string) (screenjournal.ReleaseDate, error) {
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return screenjournal.ReleaseDate{}, ErrInvalidReleaseDate
	}
	return screenjournal.ReleaseDate(t), nil
}
