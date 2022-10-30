package parse_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
)

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
			et, err := parse.WatchDate(tt.input)
			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := et, tt.output; got != want {
				t.Errorf("watchDate=%v, want=%v", got, want)
			}
		})
	}
}

func mustParseWatchDate(s string) screenjournal.WatchDate {
	et, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return screenjournal.WatchDate(et)
}
