package sqlite

import (
	"log"
	"time"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func (d DB) InsertSignupInvitation(invite screenjournal.SignupInvitation) error {
	log.Printf("inserting new signup invite code for %s: %v", invite.Invitee, invite.InviteCode)

	now := time.Now()

	if _, err := d.ctx.Exec(`
	INSERT INTO
		invites
	(
		invitee,
		code,
		created_time
	)
	VALUES (
		?, ?, ?
	)
	`,
		invite.Invitee, invite.InviteCode, formatTime(now)); err != nil {
		return err
	}

	return nil
}

func (d DB) ReadSignupInvitation(code screenjournal.InviteCode) (screenjournal.SignupInvitation, error) {
	var invitee string
	if err := d.ctx.QueryRow(`
		SELECT
			invitee
		FROM
			invites
		WHERE
			code = ?`, code).Scan(&invitee); err != nil {
		return screenjournal.SignupInvitation{}, err
	}

	return screenjournal.SignupInvitation{
		Invitee:    screenjournal.Invitee(invitee),
		InviteCode: code,
	}, nil
}

func (d DB) ReadSignupInvitations() ([]screenjournal.SignupInvitation, error) {
	rows, err := d.ctx.Query(`
		SELECT
			invitee,
			code
		FROM
			invites
		ORDER BY
			created_time DESC`)
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
	log.Printf("deleting signup code: %s", code)
	_, err := d.ctx.Exec(`DELETE FROM invites WHERE code = ?`, code.String())
	if err != nil {
		return err
	}
	return nil
}
