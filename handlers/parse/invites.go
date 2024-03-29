package parse

import (
	"errors"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

var (
	ErrInviteeInvalid    = errors.New("invalid invitee name")
	ErrInviteCodeInvalid = errors.New("invalid invite code")

	validInviteCodeChars map[rune]bool
)

func init() {
	validInviteCodeChars = make(map[rune]bool)
	for _, c := range screenjournal.InviteCodeCharset {
		validInviteCodeChars[c] = true
	}
}

func Invitee(raw string) (screenjournal.Invitee, error) {
	if !screenjournal.InviteePattern.MatchString(raw) {
		return screenjournal.Invitee(""), ErrInviteeInvalid
	}
	return screenjournal.Invitee(raw), nil
}

func InviteCode(raw string) (screenjournal.InviteCode, error) {
	if len(raw) != screenjournal.InviteCodeLength {
		return screenjournal.InviteCode(""), ErrInviteCodeInvalid
	}
	for _, c := range raw {
		if !validInviteCodeChars[c] {
			return screenjournal.InviteCode(""), ErrInviteCodeInvalid
		}
	}
	return screenjournal.InviteCode(raw), nil
}
