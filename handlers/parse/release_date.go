package parse

import (
	"errors"
	"time"

	"github.com/mtlynch/screenjournal/v2"
)

var (
	ErrInvalidReleaseDate = errors.New("invalid release date")
)

func ReleaseDate(raw string) (screenjournal.ReleaseDate, error) {
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return screenjournal.ReleaseDate{}, err
	}
	return screenjournal.ReleaseDate(t), nil
}
