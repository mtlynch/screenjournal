package sqlite

import (
	"log"

	"github.com/mtlynch/screenjournal/v2"
)

func (d DB) InsertSignupInvitation(invite screenjournal.SignupInvitation) error {
	log.Printf("inserting new signup invite code for %s: %v", invite.Invitee, invite.InviteCode)

	if _, err := d.ctx.Exec(`
	INSERT INTO
		invites
	(
		invitee,
		code
	)
	VALUES (
		?, ?
	)
	`,
		invite.Invitee, invite.InviteCode); err != nil {
		return err
	}

	return nil

}
