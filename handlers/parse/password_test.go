package parse_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func TestPassword(t *testing.T) {
	for _, tt := range []struct {
		description string
		plaintext   string
		password    screenjournal.Password
		err         error
	}{
		{
			"parses valid password",
			"dummy-p@ssword",
			screenjournal.Password("dummy-p@ssword"),
			nil,
		},
		{
			"empty string is invalid",
			"",
			screenjournal.Password(""),
			parse.ErrPasswordTooShort,
		},
		{
			"single character password is invalid",
			"q",
			screenjournal.Password(""),
			parse.ErrPasswordTooShort,
		},
		{
			"password with exactly 8 characters is valid",
			strings.Repeat("A", 8),
			screenjournal.Password(strings.Repeat("A", 8)),
			nil,
		},
		{
			"password with exactly 7 characters is invalid",
			strings.Repeat("A", 7),
			screenjournal.Password(""),
			parse.ErrPasswordTooShort,
		},
		{
			"password with exactly 40 characters is valid",
			strings.Repeat("A", 40),
			screenjournal.Password(strings.Repeat("A", 40)),
			nil,
		},
		{
			"password with exactly 41 characters is invalid",
			strings.Repeat("A", 41),
			screenjournal.Password(""),
			parse.ErrPasswordTooLong,
		},
		{
			"password with emoji characters is invalid",
			"passw😊rd123",
			screenjournal.Password(""),
			parse.ErrPasswordHasInvalidCharacters,
		},
	} {
		t.Run(fmt.Sprintf("%s [%s]", tt.description, tt.plaintext), func(t *testing.T) {
			pw, err := parse.Password(tt.plaintext)

			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := pw, tt.password; !got.Equal(want) {
				t.Errorf("password=%v, want=%v", got, want)
			}
		})
	}
}
