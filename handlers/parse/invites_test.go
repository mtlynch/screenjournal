package parse_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
)

func TestInvitee(t *testing.T) {
	for _, tt := range []struct {
		description string
		input       string
		inviteee    screenjournal.Invitee
		err         error
	}{
		{
			"valid invitee is valid",
			"Joe",
			screenjournal.Invitee("Joe"),
			nil,
		},
		{
			"invitee with accented characters is valid",
			"Joe Peña",
			screenjournal.Invitee("Joe Peña"),
			nil,
		},
		{
			"invitee with dots and dashes is valid",
			"J.P. Morgendorfer-Hoogenblast",
			screenjournal.Invitee("J.P. Morgendorfer-Hoogenblast"),
			nil,
		},
		{
			"invitee that's exactly the maximum length is valid",
			strings.Repeat("A", 80),
			screenjournal.Invitee(strings.Repeat("A", 80)),
			nil,
		},
		{
			"empty string is invalid",
			"",
			screenjournal.Invitee(""),
			parse.ErrInviteeInvalid,
		},
		{
			"invitee with more than the maximum length is invalid",
			strings.Repeat("A", 81),
			screenjournal.Invitee(""),
			parse.ErrInviteeInvalid,
		},
		{
			"invitee with newlines is invalid",
			"Joe\nSmith",
			screenjournal.Invitee(""),
			parse.ErrInviteeInvalid,
		},
		{
			"invitee with script tags is invalid",
			"Joe<script>Smith",
			screenjournal.Invitee(""),
			parse.ErrInviteeInvalid,
		},
		{
			"invitee with emoji characters is invalid",
			"M😊rk",
			screenjournal.Invitee(""),
			parse.ErrInviteeInvalid,
		},
	} {
		t.Run(fmt.Sprintf("%s [%s]", tt.description, tt.input), func(t *testing.T) {
			code, err := parse.Invitee(tt.input)

			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := code, tt.inviteee; !got.Equal(want) {
				t.Errorf("inviteee=%v, want=%v", got, want)
			}
		})
	}
}

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
			"invite code with invalid character is invalid",
			"abc45#",
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
