package tmdb_test

import (
	"fmt"
	"math"
	"testing"
	"time"

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

func TestParseImdbID(t *testing.T) {
	for _, tt := range []struct {
		description string
		in          string
		id          screenjournal.ImdbID
		err         error
	}{
		{
			"ID with 7 digits is valid",
			"tt0079367",
			screenjournal.ImdbID("tt0079367"),
			nil,
		},
		{
			"ID with 8 digits is valid",
			"tt14596320",
			screenjournal.ImdbID("tt14596320"),
			nil,
		},
		{
			"empty string is invalid",
			"",
			screenjournal.ImdbID(""),
			tmdb.ErrInvalidImdbID,
		},
		{
			"ID with missing prefix is invalid",
			"0079367",
			screenjournal.ImdbID(""),
			tmdb.ErrInvalidImdbID,
		},
		{
			"ID with too few characters is invalid",
			"tt007936",
			screenjournal.ImdbID(""),
			tmdb.ErrInvalidImdbID,
		},
		{
			"ID with too many characters is invalid",
			"tt012345678",
			screenjournal.ImdbID(""),
			tmdb.ErrInvalidImdbID,
		},
	} {
		t.Run(fmt.Sprintf("%s [%s]", tt.description, tt.in), func(t *testing.T) {
			id, err := tmdb.ParseImdbID(tt.in)
			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := id.String(), tt.id.String(); got != want {
				t.Errorf("id=%s, want=%s", got, want)
			}
		})
	}
}

func TestParseReleaseDate(t *testing.T) {
	for _, tt := range []struct {
		description string
		in          string
		releaseDate screenjournal.ReleaseDate
		err         error
	}{
		{
			"standard release date is valid",
			"2022-07-15",
			screenjournal.ReleaseDate(time.Date(2022, time.July, 15, 0, 0, 0, 0, time.UTC)),
			nil,
		},
		{
			"empty string is invalid",
			"",
			screenjournal.ReleaseDate{},
			tmdb.ErrInvalidReleaseDate,
		},
	} {
		t.Run(fmt.Sprintf("%s [%s]", tt.description, tt.in), func(t *testing.T) {
			rd, err := tmdb.ParseReleaseDate(tt.in)
			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := rd.Time(), tt.releaseDate.Time(); got != want {
				t.Errorf("releaseDate=%v, want=%v", got, want)
			}
		})
	}
}
