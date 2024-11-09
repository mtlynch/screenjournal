package parse_test

import (
	"testing"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func TestTvShowIDFromString(t *testing.T) {
	for _, tt := range []struct {
		description string
		in          string
		idExpected  screenjournal.TvShowID
		errExpected error
	}{
		{
			"parses valid TV show ID",
			"5",
			screenjournal.TvShowID(5),
			nil,
		},
		{
			"rejects 0 as an invalid TV show ID",
			"0",
			screenjournal.TvShowID(0),
			parse.ErrInvalidTvShowID,
		},
		{
			"rejects decimal as an invalid TV show ID",
			"2.4",
			screenjournal.TvShowID(0),
			parse.ErrInvalidTvShowID,
		},
		{
			"rejects non-number as an invalid TV show ID",
			"banana",
			screenjournal.TvShowID(0),
			parse.ErrInvalidTvShowID,
		},
		{
			"rejects negative number as an invalid TV show ID",
			"-5",
			screenjournal.TvShowID(0),
			parse.ErrInvalidTvShowID,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			idActual, err := parse.TvShowIDFromString(tt.in)

			if got, want := err, tt.errExpected; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := idActual, tt.idExpected; !got.Equal(want) {
				t.Errorf("tvShowID=%v, want=%v", got, want)
			}
		})
	}
}

func TestTvShowSeason(t *testing.T) {
	for _, tt := range []struct {
		description    string
		in             string
		seasonExpected screenjournal.TvShowSeason
		errExpected    error
	}{
		{
			"parses valid TV show season",
			"5",
			screenjournal.TvShowSeason(5),
			nil,
		},
		{
			"rejects 0 as an invalid TV show season",
			"0",
			screenjournal.TvShowSeason(0),
			parse.ErrInvalidTvShowSeason,
		},
		{
			"rejects decimal as an invalid TV show season",
			"2.4",
			screenjournal.TvShowSeason(0),
			parse.ErrInvalidTvShowSeason,
		},
		{
			"rejects non-number as an invalid TV show season",
			"banana",
			screenjournal.TvShowSeason(0),
			parse.ErrInvalidTvShowSeason,
		},
		{
			"rejects negative number as an invalid TV show season",
			"-5",
			screenjournal.TvShowSeason(0),
			parse.ErrInvalidTvShowSeason,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			idActual, err := parse.TvShowSeason(tt.in)

			if got, want := err, tt.errExpected; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := idActual, tt.seasonExpected; !got.Equal(want) {
				t.Errorf("tvShowID=%v, want=%v", got, want)
			}
		})
	}
}
