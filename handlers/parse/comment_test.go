package parse_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
)

func TestCommentText(t *testing.T) {
	for _, tt := range []struct {
		description string
		in          string
		comment     screenjournal.CommentText
		err         error
	}{
		{
			"regular comment is valid",
			"I agree completely!",
			screenjournal.CommentText("I agree completely!"),
			nil,
		},
		{
			"'undefined' as a comment is invalid",
			"undefined",
			screenjournal.CommentText(""),
			parse.ErrInvalidComment,
		},
		{
			"'null' as a comment is invalid",
			"null",
			screenjournal.CommentText(""),
			parse.ErrInvalidComment,
		},
		{
			"empty string is invalid",
			"",
			screenjournal.CommentText(""),
			parse.ErrInvalidComment,
		},
		{
			"single character comment is valid",
			"a",
			screenjournal.CommentText("a"),
			nil,
		},
		{
			"comment with leading spaces is invalid",
			" I thought it was bad.",
			screenjournal.CommentText(""),
			parse.ErrInvalidComment,
		},
		{
			"comment with trailing spaces is invalid",
			"I thought it was bad.   ",
			screenjournal.CommentText(""),
			parse.ErrInvalidComment,
		},
		{
			"comment with more than 9000 characters is invalid",
			strings.Repeat("A", 9001),
			screenjournal.CommentText(""),
			parse.ErrInvalidComment,
		},
		{
			"comment with <script> tag is invalid",
			"Needed more <script>",
			screenjournal.CommentText(""),
			parse.ErrInvalidComment,
		},
		{
			"comment with closing </script> tag is invalid",
			"Needed more </script>",
			screenjournal.CommentText(""),
			parse.ErrInvalidComment,
		},
		{
			"comment with <script> tag with whitespace is invalid",
			"Needed more <\t\r\n script >",
			screenjournal.CommentText(""),
			parse.ErrInvalidComment,
		},
		{
			"comment with <script> tag with strange case is invalid",
			"Needed more <ScRIpT>",
			screenjournal.CommentText(""),
			parse.ErrInvalidComment,
		},
	} {
		t.Run(fmt.Sprintf("%s [%s]", tt.description, tt.in), func(t *testing.T) {
			comment, err := parse.CommentText(tt.in)

			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := comment, tt.comment; got != want {
				t.Errorf("comment=%v, want=%v", got, want)
			}
		})
	}
}
