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

func (d DB) ReadSignupInvitations() ([]screenjournal.SignupInvitation, error) {
	rows, err := d.ctx.Query(`
		SELECT
			invitee,
			code
		FROM
			invites`)
	if err != nil {
		return []screenjournal.SignupInvitation{}, err
	}

	invites := []screenjournal.SignupInvitation{}
	for rows.Next() {
		var inviteeRaw string
		var inviteCodeRaw string
		if err := rows.Scan(&inviteeRaw, &inviteCodeRaw); err != nil {
			return []screenjournal.SignupInvitation{}, err
		}

		invites = append(invites, screenjournal.SignupInvitation{
			Invitee:    screenjournal.Invitee(inviteeRaw),
			InviteCode: screenjournal.InviteCode(inviteCodeRaw),
		})
	}

	return invites, nil
}

func (d DB) DeleteSignupInvitation(code screenjournal.InviteCode) error {
	_, err := d.ctx.Exec(`DELETE FROM invites WHERE code = ?`, code.String())
	if err != nil {
		return err
	}
	return nil
}
