package parse_test

import (
	"testing"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
)

func TestEmailAddress(t *testing.T) {
	for _, tt := range []struct {
		explanation    string
		email          string
		validExpected  bool
		parsedExpected screenjournal.Email
	}{
		{
			"well-formed email address is valid",
			"hello@example.com",
			true,
			"hello@example.com",
		},
		{
			"email with name data is valid",
			"Barry Gibbs <bg@example.com>",
			true,
			"bg@example.com",
		},
		{
			"email with angle brackets is valid",
			"<hello@example.com>",
			true,
			"hello@example.com",
		},
		{
			"empty string is invalid",
			"",
			false,
			"",
		},
		{
			"email without @ is invalid",
			"hello[at]example.com",
			false,
			"",
		},
	} {
		t.Run(tt.explanation, func(t *testing.T) {
			parsed, err := parse.Email(tt.email)
			if got, want := (err == nil), tt.validExpected; got != want {
				t.Fatalf("valid=%v, want=%v", got, want)
			}
			if got, want := parsed, tt.parsedExpected; got != want {
				t.Errorf("email=%v, want=%v", got, want)
			}
		})
	}
}
