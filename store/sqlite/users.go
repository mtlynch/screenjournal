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

func (s Store) ReadUsersPublicMeta() ([]screenjournal.UserPublicMeta, error) {
	rows, err := s.ctx.Query(`
	SELECT
    u.username,
    u.created_time,
    COUNT(r.id) as review_count
	FROM
		users u
	LEFT JOIN
		reviews r ON u.username = r.review_owner
	GROUP BY
		u.username,
		u.created_time
	ORDER BY
		u.created_time
`)
	if err != nil {
		if err == sql.ErrNoRows {
			return []screenjournal.UserPublicMeta{}, nil
		}
		return []screenjournal.UserPublicMeta{}, err
	}

	users := []screenjournal.UserPublicMeta{}
	for rows.Next() {
		var username string
		var joinTimeRaw string
		var reviewCount uint
		if err := rows.Scan(&username, &joinTimeRaw, &reviewCount); err != nil {
			return []screenjournal.UserPublicMeta{}, err
		}
		joinTime, err := parseDatetime(joinTimeRaw)
		if err != nil {
			return []screenjournal.UserPublicMeta{}, err
		}
		users = append(users, screenjournal.UserPublicMeta{
			Username:    screenjournal.Username(username),
			JoinDate:    joinTime,
			ReviewCount: reviewCount,
		})
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
		username = :username`, sql.Named("username", username.String()))
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
		:username, :is_admin, :email, :password_hash, :created_time, :last_modified_time
	)
	`,
		sql.Named("username", user.Username.String()),
		sql.Named("is_admin", user.IsAdmin),
		sql.Named("email", user.Email.String()),
		sql.Named("password_hash", encodePasswordHash(user.PasswordHash)),
		sql.Named("created_time", formatTime(now)),
		sql.Named("last_modified_time", formatTime(now))); err != nil {
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
		:username, 1, 1, 1
	)
	`,
		sql.Named("username", user.Username.String())); err != nil {
		return err
	}

	return tx.Commit()
}

func (s Store) UpdateUserPassword(username screenjournal.Username, newPasswordHash screenjournal.PasswordHash) error {
	log.Printf("updating password user %s", username.String())

	if _, err := s.ctx.Exec(`
	UPDATE users
	SET
		password_hash = :password_hash
	WHERE
		username = :username`,
		sql.Named("password_hash", encodePasswordHash(newPasswordHash.Bytes())),
		sql.Named("username", username.String())); err != nil {
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
