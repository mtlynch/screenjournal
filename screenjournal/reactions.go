package screenjournal

import (
	"strconv"
	"time"
)

type ReactionID uint64

// ReactionEmoji represents a validated emoji reaction. It should only be
// created via the parse.ReactionEmoji function.
type ReactionEmoji struct {
	value string
}

type ReviewReaction struct {
	ID      ReactionID
	Owner   Username
	Emoji   ReactionEmoji
	Created time.Time
	Review  Review
}

func (id ReactionID) UInt64() uint64 {
	return uint64(id)
}

func (id ReactionID) String() string {
	return strconv.FormatUint(id.UInt64(), 10)
}

func (e ReactionEmoji) String() string {
	return e.value
}

// NewReactionEmoji creates a ReactionEmoji from a validated string. This
// function should only be called from the parse package after validation.
func NewReactionEmoji(s string) ReactionEmoji {
	return ReactionEmoji{value: s}
}
