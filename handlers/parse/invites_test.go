package parse_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
)

func TestInviteCode(t *testing.T) {
	for _, tt := range []struct {
		description string
		input       string
		code        screenjournal.InviteCode
		err         error
	}{
		{
			"valid invite code is valid",
			"abc456",
			screenjournal.InviteCode("abc456"),
			nil,
		},
		{
			"empty string is invalid",
			"",
			screenjournal.InviteCode(""),
			parse.ErrInviteCodeInvalid,
		},
		{
			"single character invite code is invalid",
			"q",
			screenjournal.InviteCode(""),
			parse.ErrInviteCodeInvalid,
		},
		{
			"invite code with less than the code length is invalid",
			strings.Repeat("A", screenjournal.InviteCodeLength-1),
			screenjournal.InviteCode(""),
			parse.ErrInviteCodeInvalid,
		},
		{
			"invite code with more than the code length is invalid",
			strings.Repeat("A", screenjournal.InviteCodeLength+1),
			screenjournal.InviteCode(""),
			parse.ErrInviteCodeInvalid,
		},
		{
			"invite code with emoji characters is invalid",
			"23456😊",
			screenjournal.InviteCode(""),
			parse.ErrInviteCodeInvalid,
		},
	} {
		t.Run(fmt.Sprintf("%s [%s]", tt.description, tt.input), func(t *testing.T) {
			code, err := parse.InviteCode(tt.input)

			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := code, tt.code; !got.Equal(want) {
				t.Errorf("code=%v, want=%v", got, want)
			}
		})
	}
}
