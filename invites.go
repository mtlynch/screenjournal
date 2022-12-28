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

func (i Invitee) String() string {
	return string(i)
}

func (i Invitee) Empty() bool {
	return i.String() == ""
}

func NewInviteCode() InviteCode {
	// Alphanumeric characters, with commonly-confused characters removed.
	charset := []rune("ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789")
	return InviteCode(random.String(6, charset))
}

func (ic InviteCode) String() string {
	return string(ic)
}

func (ic InviteCode) Empty() bool {
	return ic.String() == ""
}

func (si SignupInvitation) Empty() bool {
	return si.Invitee.Empty()
}
