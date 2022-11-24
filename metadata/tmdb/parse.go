package tmdb

import (
	"errors"
	"regexp"

	"github.com/mtlynch/screenjournal/v2"
)

var (
	ErrInvalidImagePath = errors.New("invalid path to HTTP image")

	imagePathPattern = regexp.MustCompile(`^/[0-9A-Za-z]{27}\.[a-z]{3}`)
)

func ParseImagePath(raw string) (screenjournal.ImagePath, error) {
	if !imagePathPattern.MatchString(raw) {
		return screenjournal.ImagePath(""), ErrInvalidImagePath
	}
	return screenjournal.ImagePath(raw), nil
}
