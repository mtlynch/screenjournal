package parse_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mtlynch/screenjournal/v2/parse"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func TestSearchQuery(t *testing.T) {
	for _, tt := range []struct {
		description string
		in          string
		query       screenjournal.SearchQuery
		err         error
	}{
		{
			"valid query",
			"Meet Joe Black",
			screenjournal.SearchQuery("Meet Joe Black"),
			nil,
		},
		{
			"query with leading and trailing spaces",
			"  Inception  ",
			screenjournal.SearchQuery("Inception"),
			nil,
		},
		{
			"query with minimum length",
			strings.Repeat("A", parse.SearchQueryMinLength),
			screenjournal.SearchQuery(strings.Repeat("A", parse.SearchQueryMinLength)),
			nil,
		},
		{
			"query with maximum length",
			strings.Repeat("A", parse.SearchQueryMaxLength),
			screenjournal.SearchQuery(strings.Repeat("A", parse.SearchQueryMaxLength)),
			nil,
		},
		{
			"query too short",
			"A",
			screenjournal.SearchQuery(""),
			parse.ErrSearchQueryTooShort,
		},
		{
			"query too long",
			strings.Repeat("A", parse.SearchQueryMaxLength+1),
			screenjournal.SearchQuery(""),
			parse.ErrSearchQueryTooLong,
		},
		{
			"empty query",
			"",
			screenjournal.SearchQuery(""),
			parse.ErrSearchQueryTooShort,
		},
		{
			"query with special characters",
			"Star Wars: Episode IV - A New Hope",
			screenjournal.SearchQuery("Star Wars: Episode IV - A New Hope"),
			nil,
		},
		{
			"query with numbers",
			"2001: A Space Odyssey",
			screenjournal.SearchQuery("2001: A Space Odyssey"),
			nil,
		},
	} {
		t.Run(fmt.Sprintf("%s [%s]", tt.description, tt.in), func(t *testing.T) {
			query, err := parse.SearchQuery(tt.in)
			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := query, tt.query; got != want {
				t.Errorf("query=%s, want=%s", got, want)
			}
		})
	}
}
