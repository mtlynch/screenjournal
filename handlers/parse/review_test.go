package parse_test

import (
	"fmt"
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
			"9½ Weeks 2: The Kegend of Curly's Gold",
			screenjournal.MediaTitle("9½ Weeks 2: The Kegend of Curly's Gold"),
			nil,
		},
		{
			"single-character title is invalid",
			"j",
			screenjournal.MediaTitle(""),
			parse.ErrInvalidMediaTitle,
		},
		{
			"empty string is invalid",
			"",
			screenjournal.MediaTitle(""),
			parse.ErrInvalidMediaTitle,
		},
		{
			"title with exactly 160 characters is valid",
			strings.Repeat("A", 160),
			screenjournal.MediaTitle(strings.Repeat("A", 160)),
			nil,
		},
		{
			"title with more than 160 characters is invalid",
			strings.Repeat("A", 161),
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

func TestWatchDate(t *testing.T) {
	for _, tt := range []struct {
		description string
		input       string
		output      screenjournal.WatchDate
		err         error
	}{
		{
			description: "valid expiration",
			input:       "2022-10-28T00:00:00.000-04:00",
			output:      mustParseWatchDate("2022-10-28T00:00:00.000-04:00"),
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

func mustParseWatchDate(s string) screenjournal.WatchDate {
	wd, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return screenjournal.WatchDate(wd)
}
