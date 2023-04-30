package parse

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/mtlynch/screenjournal/v2"
)

const commentMaxLength = 9000

var (
	ErrInvalidCommentID = errors.New("invalid comment ID")
	ErrInvalidComment   = errors.New("invalid comment")
)

func CommentID(raw string) (screenjournal.CommentID, error) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		log.Printf("failed to parse comment ID: %v", err)
		return screenjournal.CommentID(0), ErrInvalidCommentID
	}

	if id == 0 {
		return screenjournal.CommentID(0), ErrInvalidCommentID
	}

	return screenjournal.CommentID(id), nil
}

func CommentText(comment string) (screenjournal.CommentText, error) {
	if strings.TrimSpace(comment) != comment {
		return screenjournal.CommentText(""), ErrInvalidComment
	}

	if len(comment) > commentMaxLength {
		return screenjournal.CommentText(""), ErrInvalidComment
	}

	if isReservedWord(comment) {
		return screenjournal.CommentText(""), ErrInvalidComment
	}
	if len(comment) < 1 {
		return screenjournal.CommentText(""), ErrInvalidComment
	}

	if scriptTagPattern.FindString(comment) != "" {
		return screenjournal.CommentText(""), ErrInvalidComment
	}

	return screenjournal.CommentText(comment), nil
}
