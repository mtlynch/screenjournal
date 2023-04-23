package parse_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
)

func TestComment(t *testing.T) {
	for _, tt := range []struct {
		description string
		in          string
		comment     screenjournal.Comment
		err         error
	}{
		{
			"regular comment is valid",
			"I agree completely!",
			screenjournal.Comment("I agree completely!"),
			nil,
		},
		{
			"'undefined' as a comment is invalid",
			"undefined",
			screenjournal.Comment(""),
			parse.ErrInvalidComment,
		},
		{
			"'null' as a comment is invalid",
			"null",
			screenjournal.Comment(""),
			parse.ErrInvalidComment,
		},
		{
			"empty string is invalid",
			"",
			screenjournal.Comment(""),
			parse.ErrInvalidComment,
		},
		{
			"single character comment is valid",
			"a",
			screenjournal.Comment("a"),
			nil,
		},
		{
			"comment with leading spaces is invalid",
			" I thought it was bad.",
			screenjournal.Comment(""),
			parse.ErrInvalidComment,
		},
		{
			"comment with trailing spaces is invalid",
			"I thought it was bad.   ",
			screenjournal.Comment(""),
			parse.ErrInvalidComment,
		},
		{
			"comment with more than 9000 characters is invalid",
			strings.Repeat("A", 9001),
			screenjournal.Comment(""),
			parse.ErrInvalidComment,
		},
		{
			"comment with <script> tag is invalid",
			"Needed more <script>",
			screenjournal.Comment(""),
			parse.ErrInvalidComment,
		},
		{
			"comment with closing </script> tag is invalid",
			"Needed more </script>",
			screenjournal.Comment(""),
			parse.ErrInvalidComment,
		},
		{
			"comment with <script> tag with whitespace is invalid",
			"Needed more <\t\r\n script >",
			screenjournal.Comment(""),
			parse.ErrInvalidComment,
		},
		{
			"comment with <script> tag with strange case is invalid",
			"Needed more <ScRIpT>",
			screenjournal.Comment(""),
			parse.ErrInvalidComment,
		},
	} {
		t.Run(fmt.Sprintf("%s [%s]", tt.description, tt.in), func(t *testing.T) {
			comment, err := parse.Comment(tt.in)

			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := comment, tt.comment; got != want {
				t.Errorf("comment=%v, want=%v", got, want)
			}
		})
	}
}
