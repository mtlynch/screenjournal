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

func (db DB) ReadUser(username screenjournal.Username) (screenjournal.User, error) {
	var email string
	var passwordHashEncoded string
	if err := db.ctx.QueryRow(`
	SELECT
		email,
		password_hash
	FROM
		users
	WHERE
		username = ?`, username.String()).Scan(&email, &passwordHashEncoded); err != nil {
		return screenjournal.User{}, err
	}

	// TODO: read isadmin

	return screenjournal.User{
		Username:     screenjournal.Username(username),
		Email:        screenjournal.Email(email),
		PasswordHash: decodePasswordHash(passwordHashEncoded),
	}, nil
}

func (db DB) InsertUser(user screenjournal.User) error {
	log.Printf("inserting new user %s, isAdmin=%v", user.Username.String(), user.IsAdmin)

	now := time.Now()

	// TODO: Set isadmin

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
		encodePasswordHash(user.PasswordHash),
		formatTime(now),
		formatTime(now)); err != nil {
		return err
	}

	return nil
}

func encodePasswordHash(ph screenjournal.PasswordHash) string {
	return ph.String()
}

func decodePasswordHash(s string) screenjournal.PasswordHash {
	b := []byte(s)
	return screenjournal.NewPasswordHashFromBytes(b)
}
