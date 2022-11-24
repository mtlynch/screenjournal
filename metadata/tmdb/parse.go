package tmdb

import (
	"errors"
	"math"

	"github.com/mtlynch/screenjournal/v2"
)

var ErrInvalidTmdbID = errors.New("invalid TMDB ID")

func ParseTmdbID(raw int) (screenjournal.TmdbID, error) {
	if raw <= 0 || raw > math.MaxInt32 {
		return screenjournal.TmdbID(0), ErrInvalidTmdbID
	}

	return screenjournal.TmdbID(raw), nil
}
