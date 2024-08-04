package parse

import (
	"errors"
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

const watchDateFormat = time.DateOnly

var (
	ErrInvalidReviewID             = errors.New("invalid review ID")
	ErrInvalidMediaTitle           = errors.New("invalid media title")
	ErrInvalidRating               = fmt.Errorf("rating must be between %d and %d", MinRating, MaxRating)
	ErrWatchDateUnrecognizedFormat = fmt.Errorf("unrecognized format for watch date, must be in %s format", watchDateFormat)
	ErrWatchDateTooLate            = fmt.Errorf("watch time must be no later than %s", time.Now().Format(time.DateOnly))
	ErrInvalidBlurb                = errors.New("invalid blurb")

	MediaTitleMinLength = 2
	MediaTitleMaxLength = 160

	scriptTagPattern = regexp.MustCompile(`(?i)<\s*/?script\s*>`)
	MinRating        = uint8(1)
	MaxRating        = uint8(5)
	blurbMaxLength   = 9000
)

func ReviewID(id uint64) (screenjournal.ReviewID, error) {
	if id == 0 {
		return screenjournal.ReviewID(0), ErrInvalidReviewID
	}

	return screenjournal.ReviewID(id), nil
}

func ReviewIDFromString(raw string) (screenjournal.ReviewID, error) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		log.Printf("failed to parse review ID: %v", err)
		return screenjournal.ReviewID(0), ErrInvalidReviewID
	}

	return ReviewID(id)
}

func MediaTitle(raw string) (screenjournal.MediaTitle, error) {
	if isReservedWord(raw) {
		return screenjournal.MediaTitle(""), ErrInvalidMediaTitle
	}

	if len(raw) < MediaTitleMinLength || len(raw) > MediaTitleMaxLength {
		return screenjournal.MediaTitle(""), ErrInvalidMediaTitle
	}

	if scriptTagPattern.FindString(raw) != "" {
		return screenjournal.MediaTitle(""), ErrInvalidMediaTitle
	}

	return screenjournal.MediaTitle(raw), nil
}

func RatingFromString(raw string) (screenjournal.Rating, error) {
	i, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return screenjournal.Rating(0), err
	}
	return Rating(int(i))
}

func Rating(raw int) (screenjournal.Rating, error) {
	if raw > math.MaxUint8 {
		return screenjournal.Rating(0), ErrInvalidRating
	}
	ratingUint8 := uint8(raw)
	if ratingUint8 < MinRating || ratingUint8 > MaxRating {
		return screenjournal.Rating(0), ErrInvalidRating
	}

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
	if len(raw) > blurbMaxLength {
		return screenjournal.Blurb(""), ErrInvalidBlurb
	}

	blurb := strings.TrimSpace(raw)

	if isReservedWord(blurb) {
		return screenjournal.Blurb(""), ErrInvalidBlurb
	}

	if scriptTagPattern.FindString(blurb) != "" {
		return screenjournal.Blurb(""), ErrInvalidBlurb
	}

	return screenjournal.Blurb(blurb), nil
}
