package parse_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
)

func TestPassword(t *testing.T) {
	for _, tt := range []struct {
		description string
		plaintext   string
		hash        screenjournal.PasswordHash
		err         error
	}{
		{
			"parses valid password",
			"dummy-p@ssword",
			screenjournal.NewPasswordHash([]byte("dummy-p@ssword")),
			nil,
		},
		{
			"empty string is invalid",
			"",
			screenjournal.PasswordHash{},
			parse.ErrPasswordTooShort,
		},
		{
			"single character password is invalid",
			"q",
			screenjournal.PasswordHash{},
			parse.ErrPasswordTooShort,
		},
		{
			"password with exactly 8 characters is valid",
			strings.Repeat("A", 8),
			screenjournal.NewPasswordHash([]byte(strings.Repeat("A", 8))),
			nil,
		},
		{
			"password with exactly 7 characters is invalid",
			strings.Repeat("A", 7),
			screenjournal.PasswordHash{},
			parse.ErrPasswordTooShort,
		},
		{
			"password with exactly 40 characters is valid",
			strings.Repeat("A", 40),
			screenjournal.NewPasswordHash([]byte(strings.Repeat("A", 40))),
			nil,
		},
		{
			"password with exactly 41 characters is invalid",
			strings.Repeat("A", 41),
			screenjournal.PasswordHash{},
			parse.ErrPasswordTooLong,
		},
		{
			"password with emoji characters is invalid",
			"passw😊rd123",
			screenjournal.PasswordHash{},
			parse.ErrPasswordHasInvalidCharacters,
		},
	} {
		t.Run(fmt.Sprintf("%s [%s]", tt.description, tt.plaintext), func(t *testing.T) {
			hash, err := parse.Password(tt.plaintext)

			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			// Make sure hash matches plaintext if it's a valid hash.
			if err != nil {
				return
			}
			var nilErr error
			if got, want := hash.MatchesPlaintext(tt.plaintext), nilErr; got != want {
				t.Errorf("matchErr=%v, want=%v", got, want)
			}
		})
	}
}
