package tmdb_test

import (
	"fmt"
	"testing"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/metadata/tmdb"
)

func TestParseImagePath(t *testing.T) {
	for _, tt := range []struct {
		description string
		in          string
		id          screenjournal.ImagePath
		err         error
	}{
		{
			"valid image path",
			"/6FfCtAuVAW8XJjZ7eWeLibRLWTw.jpg",
			screenjournal.ImagePath("/6FfCtAuVAW8XJjZ7eWeLibRLWTw.jpg"),
			nil,
		},
		{
			"empty string is invalid",
			"",
			screenjournal.ImagePath(""),
			tmdb.ErrInvalidImagePath,
		},
		{
			"path without leading slash is invalid",
			"6FfCtAuVAW8XJjZ7eWeLibRLWTw.jpg",
			screenjournal.ImagePath(""),
			tmdb.ErrInvalidImagePath,
		},
		{
			"path with directory traversal is invalid",
			"/../6FfCtAuVAW8XJjZ7eWeLibRLWTw.jpg",
			screenjournal.ImagePath(""),
			tmdb.ErrInvalidImagePath,
		},
	} {
		t.Run(fmt.Sprintf("%s [%s]", tt.description, tt.in), func(t *testing.T) {
			id, err := tmdb.ParseImagePath(tt.in)
			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := id.String(), tt.id.String(); got != want {
				t.Errorf("imagePath=%s, want=%s", got, want)
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
			"ID of The Jerk is valid",
			"tt0079367",
			screenjournal.ImdbID("tt0079367"),
			nil,
		},
		{
			"ID of Carl Reiner valid",
			"nm0005348",
			screenjournal.ImdbID("nm0005348"),
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
			"tt00793678",
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
