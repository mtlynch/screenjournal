package parse

import (
	"errors"
	"log"
	"math"
	"strconv"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

var ErrInvalidTmdbID = errors.New("invalid TMDB ID")

func TmdbIDFromString(raw string) (screenjournal.TmdbID, error) {
	id, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		log.Printf("failed to parse TMDB ID: %v", err)
		return screenjournal.TmdbID(0), ErrInvalidTmdbID
	}

	return TmdbID(int(id))
}

// TODO: Merge?
func TmdbID(raw int) (screenjournal.TmdbID, error) {
	if raw <= 0 || raw > math.MaxInt32 {
		return screenjournal.TmdbID(0), ErrInvalidTmdbID
	}

	return screenjournal.TmdbID(raw), nil
}
