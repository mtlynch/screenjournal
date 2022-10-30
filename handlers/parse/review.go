package parse

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/mtlynch/screenjournal/v2"
)

const watchDateFormat = time.RFC3339

var (
	ErrInvalidMediaTitle           = errors.New("invalid media title")
	ErrWatchDateUnrecognizedFormat = fmt.Errorf("unrecognized format for watch date, must be in %s format", watchDateFormat)
	ErrWatchDateTooLate            = fmt.Errorf("watch time must be no later than %s", time.Now().Format("2006-01-02"))
	ErrInvalidBlurb                = errors.New("invalid blurb")

	scriptTagPattern = regexp.MustCompile(`(?i)<\s*/?script\s*>`)
)

func MediaTitle(raw string) (screenjournal.MediaTitle, error) {
	if isReservedWord(raw) {
		return screenjournal.MediaTitle(""), ErrInvalidMediaTitle
	}

	if len(raw) < 5 || len(raw) > 160 {
		return screenjournal.MediaTitle(""), ErrInvalidMediaTitle
	}

	if scriptTagPattern.FindString(raw) != "" {
		return screenjournal.MediaTitle(""), ErrInvalidMediaTitle
	}

	return screenjournal.MediaTitle(raw), nil
}

func Rating(raw int) (screenjournal.Rating, error) {
	return screenjournal.Rating(raw), nil
}

func WatchDate(raw string) (screenjournal.WatchDate, error) {
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

func Blurb(raw string) (screenjournal.Blurb, error) {
	if isReservedWord(raw) {
		return screenjournal.Blurb(""), ErrInvalidBlurb
	}

	if len(raw) < 5 || len(raw) > 160 {
		return screenjournal.Blurb(""), ErrInvalidBlurb
	}

	if scriptTagPattern.FindString(raw) != "" {
		return screenjournal.Blurb(""), ErrInvalidBlurb
	}

	return screenjournal.Blurb(raw), nil
}
