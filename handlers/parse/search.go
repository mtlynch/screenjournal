package parse

import (
	"errors"
	"strings"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

var (
	ErrSearchQueryTooShort                  = errors.New("search query is too short")
	ErrSearchQueryTooLong                   = errors.New("search query is too long")
	ErrSearchQueryContainsInvalidCharacters = errors.New("search query contains invalid characters")

	SearchQueryMinLength = 2
	SearchQueryMaxLength = 100
)

func SearchQuery(raw string) (screenjournal.SearchQuery, error) {
	if len(raw) < SearchQueryMinLength {
		return screenjournal.SearchQuery(""), ErrSearchQueryTooShort
	}

	if len(raw) > SearchQueryMaxLength {
		return screenjournal.SearchQuery(""), ErrSearchQueryTooLong
	}

	query := strings.TrimSpace(raw)

	return screenjournal.SearchQuery(query), nil
}
