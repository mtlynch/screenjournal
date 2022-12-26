package screenjournal

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

func (si SignupInvitation) Empty() bool {
	return si.Invitee.Empty()
}
