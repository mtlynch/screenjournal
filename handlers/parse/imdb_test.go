package parse_test

import (
	"fmt"
	"testing"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
)

func TestImdbID(t *testing.T) {
	for _, tt := range []struct {
		description string
		in          string
		id          screenjournal.ImdbID
		err         error
	}{
		{
			"ID of The Jerk is valid",
			"tt0079367",
			screenjournal.ImdbID("tt0079367"),
			nil,
		},
		{
			"ID with alternate prefix is valid",
			"nm0079367",
			screenjournal.ImdbID("nm0079367"),
			nil,
		},
		{
			"empty string is invalid",
			"",
			screenjournal.ImdbID(""),
			parse.ErrInvalidImdbID,
		},
		{
			"ID with missing prefix is invalid",
			"0079367",
			screenjournal.ImdbID(""),
			parse.ErrInvalidImdbID,
		},
		{
			"ID with too few characters is invalid",
			"tt007936",
			screenjournal.ImdbID(""),
			parse.ErrInvalidImdbID,
		},
		{
			"ID with too many characters is invalid",
			"tt00793678",
			screenjournal.ImdbID(""),
			parse.ErrInvalidImdbID,
		},
	} {
		t.Run(fmt.Sprintf("%s [%s]", tt.description, tt.in), func(t *testing.T) {
			id, err := parse.ImdbID(tt.in)
			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := id.String(), tt.id.String(); got != want {
				t.Errorf("id=%s, want=%s", got, want)
			}
		})
	}
}
