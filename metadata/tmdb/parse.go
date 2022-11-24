package tmdb

import (
	"errors"
	"regexp"
	"time"

	"github.com/mtlynch/screenjournal/v2"
)

var (
	ErrInvalidImagePath   = errors.New("invalid path to HTTP image")
	ErrInvalidImdbID      = errors.New("invalid IMDB ID")
	ErrInvalidReleaseDate = errors.New("invalid release date")

	imdbIDPattern    = regexp.MustCompile(`^((tt)|(nm)|(co)|(ev)|(ch)|(ni))[0-9]{7}$`)
	imagePathPattern = regexp.MustCompile(`^/[0-9A-Za-z]{27}\.[a-z]{3}`)
)

func ParseImdbID(raw string) (screenjournal.ImdbID, error) {
	if !imdbIDPattern.MatchString(raw) {
		return screenjournal.ImdbID(""), ErrInvalidImdbID
	}
	return screenjournal.ImdbID(raw), nil
}

func ParseImagePath(raw string) (screenjournal.ImagePath, error) {
	if !imagePathPattern.MatchString(raw) {
		return screenjournal.ImagePath(""), ErrInvalidImagePath
	}
	return screenjournal.ImagePath(raw), nil
}

func ParseReleaseDate(raw string) (screenjournal.ReleaseDate, error) {
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return screenjournal.ReleaseDate{}, err
	}
	return screenjournal.ReleaseDate(t), nil
}
