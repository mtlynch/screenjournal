package parse

import (
	"errors"
	"log"
	"strconv"

	"github.com/mtlynch/screenjournal/v2"
)

var ErrInvalidMovieID = errors.New("invalid movie ID")

func MovieIDFromString(raw string) (screenjournal.MovieID, error) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		log.Printf("failed to parse movie ID: %v", err)
		return screenjournal.MovieID(0), ErrInvalidMovieID
	}

	if id == 0 {
		return screenjournal.MovieID(0), ErrInvalidMovieID
	}

	return screenjournal.MovieID(id), nil
}
