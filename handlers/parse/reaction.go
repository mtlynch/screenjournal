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

	for _, allowed := range screenjournal.AllowedReactionEmojis() {
		if allowed.String() == raw {
			return allowed, nil
		}
	}

	return screenjournal.ReactionEmoji{}, ErrInvalidReactionEmoji
}
