package parse

import (
	"errors"

	"github.com/mtlynch/screenjournal/v2"
)

var ErrInvalidTmdbID = errors.New("invalid TMDB ID")

func TmdbID(raw int) (screenjournal.TmdbID, error) {
	if raw <= 0 {
		return screenjournal.TmdbID(0), ErrInvalidTmdbID
	}

	return screenjournal.TmdbID(raw), nil
}
