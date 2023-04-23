package parse

import (
	"errors"
	"strings"

	"github.com/mtlynch/screenjournal/v2"
)

const commentMaxLength = 9000

var ErrInvalidComment = errors.New("invalid comment")

func Comment(comment string) (screenjournal.Comment, error) {
	if strings.TrimSpace(comment) != comment {
		return screenjournal.Comment(""), ErrInvalidComment
	}

	if len(comment) > commentMaxLength {
		return screenjournal.Comment(""), ErrInvalidComment
	}

	if isReservedWord(comment) {
		return screenjournal.Comment(""), ErrInvalidComment
	}
	if len(comment) < 1 {
		return screenjournal.Comment(""), ErrInvalidComment
	}

	if scriptTagPattern.FindString(comment) != "" {
		return screenjournal.Comment(""), ErrInvalidComment
	}

	return screenjournal.Comment(comment), nil
}
