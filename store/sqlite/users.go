package sqlite

import (
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/store"
)

func (db DB) CountUsers() (uint, error) {
	var c uint
	if err := db.ctx.QueryRow(`SELECT COUNT(*)	AS user_count FROM users`).Scan(&c); err != nil {
		return 0, err
	}
	return c, nil
}

func (db DB) ReadUsers() ([]screenjournal.User, error) {
	rows, err := db.ctx.Query(`
	SELECT
		username,
		is_admin,
		email,
		password_hash
	FROM
		users`)
	if err != nil {
		if err == sql.ErrNoRows {
			return []screenjournal.User{}, nil
		}
		return []screenjournal.User{}, err
	}

	users := []screenjournal.User{}
	for rows.Next() {
		user, err := userFromRow(rows)
		if err != nil {
			return []screenjournal.User{}, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (db DB) ReadUser(username screenjournal.Username) (screenjournal.User, error) {
	row := db.ctx.QueryRow(`
	SELECT
		username,
		is_admin,
		email,
		password_hash
	FROM
		users
	WHERE
		username = ?`, username.String())
	user, err := userFromRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return screenjournal.User{}, screenjournal.ErrUserNotFound
		}
		return screenjournal.User{}, err
	}

	return user, nil
}

func userFromRow(row rowScanner) (screenjournal.User, error) {
	var username string
	var isAdmin bool
	var email string
	var passwordHashEncoded string
	if err := row.Scan(&username, &isAdmin, &email, &passwordHashEncoded); err != nil {
		return screenjournal.User{}, err
	}
	return screenjournal.User{
		IsAdmin:      isAdmin,
		Username:     screenjournal.Username(username),
		Email:        screenjournal.Email(email),
		PasswordHash: decodePasswordHash(passwordHashEncoded),
	}, nil
}

func (db DB) InsertUser(user screenjournal.User) error {
	log.Printf("inserting new user %s, isAdmin=%v", user.Username.String(), user.IsAdmin)

	now := time.Now()

	if _, err := db.ctx.Exec(`
	INSERT INTO
		users
	(
		username,
		is_admin,
		email,
		password_hash,
		created_time,
		last_modified_time
	)
	VALUES (
		?, ?, ?, ?, ?, ?
	)
	`,
		user.Username.String(),
		user.IsAdmin,
		user.Email.String(),
		encodePasswordHash(user.PasswordHash),
		formatTime(now),
		formatTime(now)); err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if errors.Is(sqliteErr.Code, sqlite3.ErrConstraint) {
				if strings.HasSuffix(err.Error(), "users.email") {
					return store.ErrEmailAssociatedWithAnotherAccount
				}
				if strings.HasSuffix(err.Error(), "users.username") {
					return store.ErrUsernameNotAvailable
				}
			}
		}
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
