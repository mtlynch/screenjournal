package parse_test

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func TestReviewIDFromString(t *testing.T) {
	for _, tt := range []struct {
		description string
		in          string
		id          screenjournal.ReviewID
		err         error
	}{
		{
			"ID of 1 is valid",
			"1",
			screenjournal.ReviewID(1),
			nil,
		},
		{
			"ID of MaxUint64 is valid",
			fmt.Sprintf("%d", uint64(math.MaxUint64)),
			screenjournal.ReviewID(math.MaxUint64),
			nil,
		},
		{
			"ID of -1 is invalid",
			"-1",
			screenjournal.ReviewID(0),
			parse.ErrInvalidReviewID,
		},
		{
			"ID of 0 is invalid",
			"0",
			screenjournal.ReviewID(0),
			parse.ErrInvalidReviewID,
		},
		{
			"non-numeric ID is invalid",
			"banana",
			screenjournal.ReviewID(0),
			parse.ErrInvalidReviewID,
		},
	} {
		t.Run(fmt.Sprintf("%s [%s]", tt.description, tt.in), func(t *testing.T) {
			id, err := parse.ReviewIDFromString(tt.in)
			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := id.UInt64(), tt.id.UInt64(); got != want {
				t.Errorf("id=%d, want=%d", got, want)
			}
		})
	}
}

func TestMediaType(t *testing.T) {
	for _, tt := range []struct {
		description string
		in          string
		mediaType   screenjournal.MediaType
		err         error
	}{
		{
			"movie type is a valid media type",
			"movie",
			screenjournal.MediaTypeMovie,
			nil,
		},
		{
			"tv-show type is a valid media type",
			"tv-show",
			screenjournal.MediaTypeTvShow,
			nil,
		},
		{
			"banana is an invalid media type",
			"banana",
			screenjournal.MediaType(""),
			parse.ErrInvalidMediaType,
		},
		{
			"empty string is an invalid media type",
			"",
			screenjournal.MediaType(""),
			parse.ErrInvalidMediaType,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			mt, err := parse.MediaType(tt.in)
			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := mt.String(), tt.mediaType.String(); got != want {
				t.Errorf("mediaType=%s, want=%s", got, want)
			}
		})
	}
}

func TestMediaTitle(t *testing.T) {
	for _, tt := range []struct {
		description string
		in          string
		title       screenjournal.MediaTitle
		err         error
	}{
		{
			"regular title is valid",
			"Meet Joe Black",
			screenjournal.MediaTitle("Meet Joe Black"),
			nil,
		},
		{
			"title with non-alphanumeric characters is valid",
			"9½ Weeks 2: The Legend of Curly's Gold",
			screenjournal.MediaTitle("9½ Weeks 2: The Legend of Curly's Gold"),
			nil,
		},
		{
			"single-character title is invalid",
			"j",
			screenjournal.MediaTitle(""),
			parse.ErrInvalidMediaTitle,
		},
		{
			"two-character title is valid",
			"It",
			screenjournal.MediaTitle("It"),
			nil,
		},
		{
			"empty string is invalid",
			"",
			screenjournal.MediaTitle(""),
			parse.ErrInvalidMediaTitle,
		},
		{
			"title with exactly MediaTitleMaxLength characters is valid",
			strings.Repeat("A", parse.MediaTitleMaxLength),
			screenjournal.MediaTitle(strings.Repeat("A", parse.MediaTitleMaxLength)),
			nil,
		},
		{
			"title with more than MediaTitleMaxLength characters is invalid",
			strings.Repeat("A", parse.MediaTitleMaxLength+1),
			screenjournal.MediaTitle(""),
			parse.ErrInvalidMediaTitle,
		},
		{
			"title with <script> tag is invalid",
			"GiggleFest 2k22 <script>",
			screenjournal.MediaTitle(""),
			parse.ErrInvalidMediaTitle,
		},
		{
			"title with closing </script> tag is invalid",
			"GiggleFest 2k22 </script>",
			screenjournal.MediaTitle(""),
			parse.ErrInvalidMediaTitle,
		},
		{
			"title with <script> tag with whitespace is invalid",
			"GiggleFest 2k22 <\t\r\n script >",
			screenjournal.MediaTitle(""),
			parse.ErrInvalidMediaTitle,
		},
		{
			"title with <script> tag with strange case is invalid",
			"GiggleFest 2k22 <ScRIpT>",
			screenjournal.MediaTitle(""),
			parse.ErrInvalidMediaTitle,
		},
		{
			"'undefined' as a title is invalid",
			"undefined",
			screenjournal.MediaTitle(""),
			parse.ErrInvalidMediaTitle,
		},
	} {
		t.Run(fmt.Sprintf("%s [%s]", tt.description, tt.in), func(t *testing.T) {
			title, err := parse.MediaTitle(tt.in)
			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := title, tt.title; got != want {
				t.Errorf("title=%s, want=%s", got, want)
			}
		})
	}
}

func TestRating(t *testing.T) {
	for _, tt := range []struct {
		description string
		in          int
		rating      screenjournal.Rating
		err         error
	}{
		{
			"rating of 1 is valid",
			1,
			screenjournal.NewRating(1),
			nil,
		},
		{
			"rating of 5 is valid",
			5,
			screenjournal.NewRating(5),
			nil,
		},
		{
			"rating of 10 is valid",
			10,
			screenjournal.NewRating(10),
			nil,
		},
		{
			"rating of -1 is invalid",
			-1,
			screenjournal.Rating{},
			parse.ErrInvalidRating,
		},
		{
			"rating of 11 is invalid",
			11,
			screenjournal.Rating{},
			parse.ErrInvalidRating,
		},
		{
			"rating of MaxInt is invalid",
			math.MaxInt,
			screenjournal.Rating{},
			parse.ErrInvalidRating,
		},
		{
			"rating of MinInt is invalid",
			math.MinInt,
			screenjournal.Rating{},
			parse.ErrInvalidRating,
		},
		{
			"rating of 0 is invalid",
			0,
			screenjournal.Rating{},
			parse.ErrInvalidRating,
		},
	} {
		t.Run(fmt.Sprintf("%s [%d]", tt.description, tt.in), func(t *testing.T) {
			rating, err := parse.Rating(tt.in)
			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}

			if tt.rating.IsNil() {
				if !rating.IsNil() {
					t.Errorf("expected nil rating, got %v", rating.UInt8())
				}
			} else {
				if rating.IsNil() {
					t.Errorf("expected non-nil rating")
				} else if got, want := rating.UInt8(), tt.rating.UInt8(); got != want {
					t.Errorf("rating=%d, want=%d", got, want)
				}
			}
		})
	}
}

func TestRatingFromString(t *testing.T) {
	for _, tt := range []struct {
		description string
		in          string
		rating      screenjournal.Rating
		err         error
	}{
		{
			"valid rating string",
			"5",
			screenjournal.NewRating(5),
			nil,
		},
		{
			"empty string returns nil rating",
			"",
			screenjournal.Rating{},
			nil,
		},
		{
			"non-numeric string is invalid",
			"banana",
			screenjournal.Rating{},
			&strconv.NumError{},
		},
	} {
		t.Run(fmt.Sprintf("%s [%s]", tt.description, tt.in), func(t *testing.T) {
			rating, err := parse.RatingFromString(tt.in)

			if tt.err == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			} else if tt.err != nil && err == nil {
				t.Fatalf("expected error of type %T, got nil", tt.err)
			} else if tt.err != nil && err != nil {
				// Just check error type, not exact message
				if got, want := fmt.Sprintf("%T", err), fmt.Sprintf("%T", tt.err); got != want {
					t.Fatalf("err=%v (%T), want error of type %T", err, err, tt.err)
				}
			}

			if tt.rating.IsNil() {
				if !rating.IsNil() {
					t.Errorf("expected nil rating, got %v", rating.UInt8())
				}
			} else {
				if rating.IsNil() {
					t.Errorf("expected non-nil rating")
				} else if got, want := rating.UInt8(), tt.rating.UInt8(); got != want {
					t.Errorf("rating=%d, want=%d", got, want)
				}
			}
		})
	}
}

func TestWatchDate(t *testing.T) {
	for _, tt := range []struct {
		description string
		input       string
		output      screenjournal.WatchDate
		err         error
	}{
		{
			description: "valid watch date",
			input:       "2022-10-28",
			output:      mustParseWatchDate("2022-10-28"),
			err:         nil,
		},
		{
			description: "reject watch date in the future",
			input:       "3000-01-01",
			output:      screenjournal.WatchDate{},
			err:         parse.ErrWatchDateTooLate,
		},
		{
			description: "empty string is invalid",
			input:       "",
			output:      screenjournal.WatchDate{},
			err:         parse.ErrWatchDateUnrecognizedFormat,
		},
		{
			description: "string with letters causes error",
			input:       "banana",
			output:      screenjournal.WatchDate{},
			err:         parse.ErrWatchDateUnrecognizedFormat,
		},
	} {
		t.Run(fmt.Sprintf("%s [%s]", tt.description, tt.input), func(t *testing.T) {
			wd, err := parse.WatchDate(tt.input)
			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := wd.Time(), tt.output.Time(); !got.Equal(want) {
				t.Errorf("watchDate=%v, want=%v", got, want)
			}
		})
	}
}

func TestBlurb(t *testing.T) {
	for _, tt := range []struct {
		explanation string
		in          string
		blurb       screenjournal.Blurb
		err         error
	}{
		{
			"short blurb is valid",
			"I loved it!",
			screenjournal.Blurb("I loved it!"),
			nil,
		},
		{
			"blurb with exactly 3000 characters is valid",
			strings.Repeat("A", 3000),
			screenjournal.Blurb(strings.Repeat("A", 3000)),
			nil,
		},
		{
			"empty blurb is valid",
			"",
			screenjournal.Blurb(""),
			nil,
		},
		{
			"blurb with leading spaces is valid",
			" I thought it was bad.",
			screenjournal.Blurb("I thought it was bad."),
			nil,
		},
		{
			"blurb with trailing spaces is valid",
			"I thought it was bad.   ",
			screenjournal.Blurb("I thought it was bad."),
			nil,
		},
		{
			"blurb with more than 9000 characters is invalid",
			strings.Repeat("A", 9001),
			screenjournal.Blurb(""),
			parse.ErrInvalidBlurb,
		},
		{
			"blurb with <script> tag is invalid",
			"Needed more <script>",
			screenjournal.Blurb(""),
			parse.ErrInvalidBlurb,
		},
		{
			"blurb with closing </script> tag is invalid",
			"Needed more </script>",
			screenjournal.Blurb(""),
			parse.ErrInvalidBlurb,
		},
		{
			"blurb with <script> tag with whitespace is invalid",
			"Needed more <\t\r\n script >",
			screenjournal.Blurb(""),
			parse.ErrInvalidBlurb,
		},
		{
			"blurb with <script> tag with strange case is invalid",
			"Needed more <ScRIpT>",
			screenjournal.Blurb(""),
			parse.ErrInvalidBlurb,
		},
		{
			"blurb that's a reserved word is invalid",
			"undefined",
			screenjournal.Blurb(""),
			parse.ErrInvalidBlurb,
		},
	} {
		t.Run(tt.explanation, func(t *testing.T) {
			blurb, err := parse.Blurb(tt.in)
			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := blurb, tt.blurb; got != want {
				t.Errorf("blurb=%s, want=%s", got, want)
			}
		})
	}
}

func mustParseWatchDate(s string) screenjournal.WatchDate {
	wd, err := parse.WatchDate(s)
	if err != nil {
		panic(err)
	}
	return wd
}
