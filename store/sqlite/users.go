package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
)

func (s Store) CountUsers() (uint, error) {
	var c uint
	if err := s.ctx.QueryRow(`SELECT COUNT(*)	AS user_count FROM users`).Scan(&c); err != nil {
		return 0, err
	}
	return c, nil
}

func (s Store) ReadUsers() ([]screenjournal.User, error) {
	rows, err := s.ctx.Query(`
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

func (s Store) ReadUser(username screenjournal.Username) (screenjournal.User, error) {
	row := s.ctx.QueryRow(`
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
			return screenjournal.User{}, store.ErrUserNotFound
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

func (s Store) InsertUser(user screenjournal.User) error {
	log.Printf("inserting new user %s, isAdmin=%v", user.Username.String(), user.IsAdmin)

	now := time.Now()

	tx, err := s.ctx.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`
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

	if _, err := tx.Exec(`
	INSERT INTO
		notification_preferences
	(
		username,
		new_reviews,
		all_new_comments,
		comments_on_my_reviews
	)
	VALUES (
		?, 1, 1, 1
	)
	`,
		user.Username.String()); err != nil {
		return err
	}

	return tx.Commit()
}

func (s Store) UpdateUserPassword(username screenjournal.Username, newPasswordHash screenjournal.PasswordHash) error {
	log.Printf("updating password user %s", username.String())

	if _, err := s.ctx.Exec(`
	UPDATE users
	SET
		password_hash = ?
	WHERE
		username = ?`, encodePasswordHash(newPasswordHash.Bytes()), username.String()); err != nil {
		return err
	}

	return nil
}

func encodePasswordHash(ph screenjournal.PasswordHash) string {
	return string(ph.Bytes())
}

func decodePasswordHash(s string) screenjournal.PasswordHash {
	return screenjournal.PasswordHash([]byte(s))
}
