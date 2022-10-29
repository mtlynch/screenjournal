package parse

import (
	"fmt"
	"log"
	"time"

	"github.com/mtlynch/screenjournal/v2"
)

const watchDateFormat = time.RFC3339

var ErrWatchDateUnrecognizedFormat = fmt.Errorf("unrecognized format for watch date, must be in %s format", watchDateFormat)
var ErrWatchDateTooLate = fmt.Errorf("watch time must be no later than %s", time.Now().Format("2006-01-02"))

func WatchDate(raw string) (screenjournal.WatchDate, error) {
	log.Printf("DEBUG: %s", raw)
	t, err := time.Parse(watchDateFormat, raw)
	if err != nil {
		return screenjournal.WatchDate{}, ErrWatchDateUnrecognizedFormat
	}

	now := time.Now()
	tomorrow := now.AddDate( /*years=*/ 0 /*months=*/, 0 /*days=*/, 1)
	if t.After(tomorrow) {
		return screenjournal.WatchDate{}, ErrWatchDateTooLate
	}

	return screenjournal.WatchDate(t), nil
}
