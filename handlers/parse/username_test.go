package parse_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
)

func TestUsername(t *testing.T) {
	for _, tt := range []struct {
		description string
		in          string
		username    screenjournal.Username
		err         error
	}{
		{
			"regular username is valid",
			"jerry.seinfeld",
			screenjournal.Username("jerry.seinfeld"),
			nil,
		},
		{
			"empty string is invalid",
			"",
			screenjournal.Username(""),
			parse.ErrInvalidUsername,
		},
		{
			"single character username is invalid",
			"q",
			screenjournal.Username(""),
			parse.ErrInvalidUsername,
		},
		{
			"username with exactly 80 characters is valid",
			strings.Repeat("A", 80),
			screenjournal.Username(strings.Repeat("A", 80)),
			nil,
		},
		{
			"username with more than 80 characters is invalid",
			strings.Repeat("A", 81),
			screenjournal.Username(""),
			parse.ErrInvalidUsername,
		},
		{
			"username with spaces is invalid",
			"jerry seinfeld",
			screenjournal.Username(""),
			parse.ErrInvalidUsername,
		},
		{
			"username with underscore is invalid",
			"jerry_seinfeld",
			screenjournal.Username(""),
			parse.ErrInvalidUsername,
		},
		{
			"username with dash is invalid",
			"jerry-seinfeld",
			screenjournal.Username(""),
			parse.ErrInvalidUsername,
		},
		{
			"'undefined' as a username is invalid",
			"undefined",
			screenjournal.Username(""),
			parse.ErrInvalidUsername,
		},
		{
			"'null' as a username is invalid",
			"null",
			screenjournal.Username(""),
			parse.ErrInvalidUsername,
		},
		{
			"'root' as a username is invalid",
			"root",
			screenjournal.Username(""),
			parse.ErrInvalidUsername,
		},
		{
			"'admin' as a username is invalid",
			"admin",
			screenjournal.Username(""),
			parse.ErrInvalidUsername,
		},
		{
			"'add' as a username is invalid",
			"add",
			screenjournal.Username(""),
			parse.ErrInvalidUsername,
		},
		{
			"'delete' as a username is invalid",
			"delete",
			screenjournal.Username(""),
			parse.ErrInvalidUsername,
		},
		{
			"'edit' as a username is invalid",
			"edit",
			screenjournal.Username(""),
			parse.ErrInvalidUsername,
		},
		{
			"'copy' as a username is invalid",
			"copy",
			screenjournal.Username(""),
			parse.ErrInvalidUsername,
		},
	} {
		t.Run(fmt.Sprintf("%s [%s]", tt.description, tt.in), func(t *testing.T) {
			username, err := parse.Username(tt.in)

			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := username, tt.username; got != want {
				t.Errorf("username=%v, want=%v", got, want)
			}
		})
	}
}
