package tmdb_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/metadata/tmdb"
)

func TestParseTmdbID(t *testing.T) {
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
			"ID of math.MaxInt32 is valid",
			math.MaxInt32,
			screenjournal.TmdbID(math.MaxInt32),
			nil,
		},
		{
			"ID of math.MaxInt64 is invalid",
			math.MaxInt64,
			screenjournal.TmdbID(0),
			tmdb.ErrInvalidTmdbID,
		},
		{
			"zero ID is invalid",
			0,
			screenjournal.TmdbID(0),
			tmdb.ErrInvalidTmdbID,
		},
		{
			"negative ID is invalid",
			-1,
			screenjournal.TmdbID(0),
			tmdb.ErrInvalidTmdbID,
		},
	} {
		t.Run(fmt.Sprintf("%s [%d]", tt.description, tt.in), func(t *testing.T) {
			id, err := tmdb.ParseTmdbID(tt.in)
			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := id.Int32(), tt.id.Int32(); got != want {
				t.Errorf("id=%d, want=%d", got, want)
			}
		})
	}
}
