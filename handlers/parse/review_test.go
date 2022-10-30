package parse_test

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
)

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
			screenjournal.Rating(1),
			nil,
		},
		{
			"rating of 10 is valid",
			10,
			screenjournal.Rating(10),
			nil,
		},
		{
			"rating of -1 is invalid",
			-1,
			screenjournal.Rating(0),
			parse.ErrInvalidRating,
		},
		{
			"rating of MaxInt is invalid",
			math.MaxInt,
			screenjournal.Rating(0),
			parse.ErrInvalidRating,
		},
		{
			"rating of MinInt is invalid",
			math.MinInt,
			screenjournal.Rating(0),
			parse.ErrInvalidRating,
		},
		{
			"rating of 0 is invalid",
			0,
			screenjournal.Rating(0),
			parse.ErrInvalidRating,
		},
	} {
		t.Run(fmt.Sprintf("%s [%d]", tt.description, tt.in), func(t *testing.T) {
			rating, err := parse.Rating(tt.in)
			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := rating.UInt8(), tt.rating.UInt8(); got != want {
				t.Errorf("rating=%d, want=%d", got, want)
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
			input:       "2022-10-28T00:00:00-04:00",
			output:      mustParseWatchDate("2022-10-28T00:00:00-04:00"),
			err:         nil,
		},
		{
			description: "reject watch date in the future",
			input:       "3000-01-01T00:00:00Z",
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
			"blurb with leading spaces is invalid",
			" I thought it was bad.",
			screenjournal.Blurb(""),
			parse.ErrInvalidBlurb,
		},
		{
			"blurb with trailing spaces is invalid",
			"I thought it was bad.   ",
			screenjournal.Blurb(""),
			parse.ErrInvalidBlurb,
		},
		{
			"blurb with more than 3000 characters is invalid",
			strings.Repeat("A", 3001),
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
	wd, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return screenjournal.WatchDate(wd)
}
