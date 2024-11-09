package parse

import (
	"errors"
	"log"
	"strconv"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

var (
	ErrInvalidTvShowID     = errors.New("invalid TV show ID")
	ErrInvalidTvShowSeason = errors.New("invalid TV show season")
)

func TvShowIDFromString(raw string) (screenjournal.TvShowID, error) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		log.Printf("failed to parse TV show ID: %v", err)
		return screenjournal.TvShowID(0), ErrInvalidTvShowID
	}

	if id == 0 {
		return screenjournal.TvShowID(0), ErrInvalidTvShowID
	}

	return screenjournal.TvShowID(id), nil
}

func TvShowSeason(raw string) (screenjournal.TvShowSeason, error) {
	id, err := strconv.ParseUint(raw, 10, 8)
	if err != nil {
		log.Printf("failed to parse TV show season: %v", err)
		return screenjournal.TvShowSeason(0), ErrInvalidTvShowSeason
	}

	if id == 0 {
		return screenjournal.TvShowSeason(0), ErrInvalidTvShowSeason
	}

	return screenjournal.TvShowSeason(id), nil
}
