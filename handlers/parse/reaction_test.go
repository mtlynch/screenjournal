package parse_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func TestReactionID(t *testing.T) {
	for _, tt := range []struct {
		description string
		in          string
		id          screenjournal.ReactionID
		err         error
	}{
		{
			"ID of 1 is valid",
			"1",
			screenjournal.ReactionID(1),
			nil,
		},
		{
			"ID of MaxUint64 is valid",
			fmt.Sprintf("%d", uint64(math.MaxUint64)),
			screenjournal.ReactionID(math.MaxUint64),
			nil,
		},
		{
			"ID of -1 is invalid",
			"-1",
			screenjournal.ReactionID(0),
			parse.ErrInvalidReactionID,
		},
		{
			"ID of 0 is invalid",
			"0",
			screenjournal.ReactionID(0),
			parse.ErrInvalidReactionID,
		},
		{
			"non-numeric ID is invalid",
			"banana",
			screenjournal.ReactionID(0),
			parse.ErrInvalidReactionID,
		},
	} {
		t.Run(fmt.Sprintf("%s [%s]", tt.description, tt.in), func(t *testing.T) {
			id, err := parse.ReactionID(tt.in)
			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := id.UInt64(), tt.id.UInt64(); got != want {
				t.Errorf("id=%d, want=%d", got, want)
			}
		})
	}
}

func TestReactionEmoji(t *testing.T) {
	for _, tt := range []struct {
		description string
		in          string
		emoji       string
		err         error
	}{
		{
			"thumbs up emoji is valid",
			"👍",
			"👍",
			nil,
		},
		{
			"eyes emoji is valid",
			"👀",
			"👀",
			nil,
		},
		{
			"surprised emoji is valid",
			"😯",
			"😯",
			nil,
		},
		{
			"thinking emoji is valid",
			"🤔",
			"🤔",
			nil,
		},
		{
			"pancakes emoji is valid",
			"🥞",
			"🥞",
			nil,
		},
		{
			"heart emoji is not in allowed list",
			"❤️",
			"",
			parse.ErrInvalidReactionEmoji,
		},
		{
			"empty string is invalid",
			"",
			"",
			parse.ErrInvalidReactionEmoji,
		},
		{
			"text is invalid",
			"hello",
			"",
			parse.ErrInvalidReactionEmoji,
		},
	} {
		t.Run(fmt.Sprintf("%s [%s]", tt.description, tt.in), func(t *testing.T) {
			emoji, err := parse.ReactionEmoji(tt.in)
			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := emoji.String(), tt.emoji; got != want {
				t.Errorf("emoji=%v, want=%v", got, want)
			}
		})
	}
}
