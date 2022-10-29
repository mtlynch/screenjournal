package parse

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/mtlynch/screenjournal/v2"
)

const watchDateFormat = time.RFC3339

var ErrWatchDateUnrecognizedFormat = fmt.Errorf("unrecognized format for watch date, must be in %s format", watchDateFormat)
var ErrWatchDateTooLate = errors.New("expire time must be at least one hour in the future")

func WatchDate(raw string) (screenjournal.WatchDate, error) {
	log.Printf("DEBUG: %s", raw)
	t, err := time.Parse(watchDateFormat, raw)
	if err != nil {
		return screenjournal.WatchDate{}, ErrWatchDateUnrecognizedFormat
	}

	now := time.Now().UTC()
	cutoff := now.AddDate( /*years=*/ 1 /*months=*/, 0 /*days=*/, 1)
	if t.After(cutoff) {
		return screenjournal.WatchDate{}, ErrWatchDateTooLate
	}

	return screenjournal.WatchDate(t), nil
}
