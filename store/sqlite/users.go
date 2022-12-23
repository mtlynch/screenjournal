package sqlite

import (
	"log"
	"time"

	"github.com/mtlynch/screenjournal/v2"
)

func (db DB) CountUsers() (uint, error) {
	var c uint
	if err := db.ctx.QueryRow(`SELECT COUNT(*)	AS user_count FROM users`).Scan(&c); err != nil {
		return 0, err
	}
	return c, nil
}

func (db DB) InsertUser(user screenjournal.User, passwordHash screenjournal.PasswordHash) error {
	log.Printf("inserting new user %s, isAdmin=%v", user.Username.String(), user.IsAdmin)

	now := time.Now()

	if _, err := db.ctx.Exec(`
	INSERT INTO
		users
	(
		username,
		email,
		password_hash,
		created_time,
		last_modified_time
	)
	VALUES (
		?, ?, ?, ?, ?
	)
	`,
		user.Username.String(),
		user.Email.String(),
		passwordHash.String(),
		formatTime(now),
		formatTime(now)); err != nil {
		return err
	}

	return nil
}
