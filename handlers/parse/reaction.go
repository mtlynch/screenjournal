package parse

import (
	"errors"
	"log"
	"strconv"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

var (
	ErrInvalidReactionID    = errors.New("invalid reaction ID")
	ErrInvalidReactionEmoji = errors.New("invalid reaction emoji")
)

var allowedEmojis = map[string]bool{
	"👍": true,
	"👀": true,
	"😯": true,
	"🤔": true,
	"🥞": true,
}

// AllowedReactionEmojis returns the list of allowed emoji strings for
// reactions.
func AllowedReactionEmojis() []string {
	return []string{"👍", "👀", "😯", "🤔", "🥞"}
}

func ReactionID(raw string) (screenjournal.ReactionID, error) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		log.Printf("failed to parse reaction ID: %v", err)
		return screenjournal.ReactionID(0), ErrInvalidReactionID
	}

	if id == 0 {
		return screenjournal.ReactionID(0), ErrInvalidReactionID
	}

	return screenjournal.ReactionID(id), nil
}

func ReactionEmoji(raw string) (screenjournal.ReactionEmoji, error) {
	if raw == "" {
		return screenjournal.ReactionEmoji{}, ErrInvalidReactionEmoji
	}

	if !allowedEmojis[raw] {
		return screenjournal.ReactionEmoji{}, ErrInvalidReactionEmoji
	}

	return screenjournal.NewReactionEmoji(raw), nil
}
