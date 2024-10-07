package parse_test

import (
	"testing"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func TestMovieIDFromString(t *testing.T) {
	for _, tt := range []struct {
		description string
		in          string
		idExpected  screenjournal.MovieID
		errExpected error
	}{
		{
			"parses valid movie ID",
			"5",
			screenjournal.MovieID(5),
			nil,
		},
		{
			"rejects 0 as an invalid movie ID",
			"0",
			screenjournal.MovieID(0),
			parse.ErrInvalidMovieID,
		},
		{
			"rejects decimal as an invalid movie ID",
			"2.4",
			screenjournal.MovieID(0),
			parse.ErrInvalidMovieID,
		},
		{
			"rejects non-number as an invalid movie ID",
			"banana",
			screenjournal.MovieID(0),
			parse.ErrInvalidMovieID,
		},
		{
			"rejects negative number as an invalid movie ID",
			"-5",
			screenjournal.MovieID(0),
			parse.ErrInvalidMovieID,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			idActual, err := parse.MovieIDFromString(tt.in)

			if got, want := err, tt.errExpected; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := idActual, tt.idExpected; !got.Equal(want) {
				t.Errorf("movieID=%v, want=%v", got, want)
			}
		})
	}
}
