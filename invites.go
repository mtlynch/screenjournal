package screenjournal

import "github.com/mtlynch/screenjournal/v2/random"

type (
	Invitee    string
	InviteCode string

	SignupInvitation struct {
		Invitee    Invitee
		InviteCode InviteCode
	}
)

const InviteCodeLength = 6

// InviteCodeCharset contains the allowed characters for an invite code. It
// includes alphanumeric characters with commonly-confused characters removed.
var InviteCodeCharset = []rune("ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789")

func (i Invitee) String() string {
	return string(i)
}

func (i Invitee) Empty() bool {
	return i.String() == ""
}

func NewInviteCode() InviteCode {
	return InviteCode(random.String(InviteCodeLength, InviteCodeCharset))
}

func (ic InviteCode) String() string {
	return string(ic)
}

func (ic InviteCode) Empty() bool {
	return ic.String() == ""
}

func (ic InviteCode) Equal(other InviteCode) bool {
	return ic.String() == other.String()
}

func (si SignupInvitation) Empty() bool {
	return si.Invitee.Empty()
}
