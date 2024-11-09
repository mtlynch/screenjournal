package parse

import (
	"errors"
	"log"
	"math"
	"net/url"
	"strconv"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

var (
	ErrInvalidTmdbID     = errors.New("invalid TMDB ID")
	ErrInvalidPosterPath = errors.New("invalid poster path")
)

func TmdbIDFromString(raw string) (screenjournal.TmdbID, error) {
	id, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		log.Printf("failed to parse TMDB ID: %v", err)
		return screenjournal.TmdbID(0), ErrInvalidTmdbID
	}

	return TmdbID(int(id))
}

func TmdbID(raw int) (screenjournal.TmdbID, error) {
	if raw <= 0 || raw > math.MaxInt32 {
		return screenjournal.TmdbID(0), ErrInvalidTmdbID
	}

	return screenjournal.TmdbID(raw), nil
}

func PosterPath(raw string) (url.URL, error) {
	pp, err := url.Parse(raw)
	if err != nil {
		return url.URL{}, ErrInvalidPosterPath
	}

	return *pp, nil
}
