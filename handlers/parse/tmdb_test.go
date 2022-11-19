package parse_test

import (
	"fmt"
	"testing"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
)

func TestTmdbID(t *testing.T) {
	for _, tt := range []struct {
		description string
		in          int
		id          screenjournal.TmdbID
		err         error
	}{
		{
			"ID of 36955 is valid",
			36955,
			screenjournal.TmdbID(36955),
			nil,
		},
		{
			"negative ID is invalid",
			-1,
			screenjournal.TmdbID(0),
			parse.ErrInvalidTmdbID,
		},
	} {
		t.Run(fmt.Sprintf("%s [%d]", tt.description, tt.in), func(t *testing.T) {
			id, err := parse.TmdbID(tt.in)
			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := id.Int(), tt.id.Int(); got != want {
				t.Errorf("id=%d, want=%d", got, want)
			}
		})
	}
}
