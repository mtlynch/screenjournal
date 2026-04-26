package sqlite

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
)

func (s Store) InsertPasswordResetEntry(request screenjournal.PasswordResetEntry) error {
	log.Printf("inserting new password reset token for user %s", request.Username)

	if _, err := s.exec(`
	INSERT OR REPLACE INTO
		password_reset_tokens
	(
		username,
		token,
		expires_at
	)
	VALUES (
		:username, :token, :expires_at
	)
	`,
		sql.Named("username", request.Username),
		sql.Named("token", request.Token.String()),
		sql.Named("expires_at", formatTime(request.ExpiresAt))); err != nil {
		return err
	}

	return nil
}

func (s Store) ReadPasswordResetEntry(token screenjournal.PasswordResetToken) (screenjournal.PasswordResetEntry, error) {
	var username string
	var expiresAtRaw string
	if err := s.queryRow(`
		SELECT
			username,
			expires_at
		FROM
			password_reset_tokens
		WHERE
			token = :token`, sql.Named("token", token.String())).Scan(&username, &expiresAtRaw); err != nil {
		return screenjournal.PasswordResetEntry{}, err
	}

	expiresAt, err := parseDatetime(expiresAtRaw)
	if err != nil {
		return screenjournal.PasswordResetEntry{}, err
	}

	return screenjournal.PasswordResetEntry{
		Username:  screenjournal.Username(username),
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

func (s Store) ReadLatestPasswordResetEntryForUser(username screenjournal.Username) (screenjournal.PasswordResetEntry, error) {
	var tokenRaw string
	var expiresAtRaw string
	if err := s.queryRow(`
		SELECT
			token,
			expires_at
		FROM
			password_reset_tokens
		WHERE
			username = :username
		ORDER BY
			expires_at DESC
		LIMIT 1`, sql.Named("username", username.String())).
		Scan(&tokenRaw, &expiresAtRaw); err != nil {
		return screenjournal.PasswordResetEntry{}, err
	}

	expiresAt, err := parseDatetime(expiresAtRaw)
	if err != nil {
		return screenjournal.PasswordResetEntry{}, err
	}

	return screenjournal.PasswordResetEntry{
		Username:  username,
		Token:     screenjournal.NewPasswordResetTokenFromString(tokenRaw),
		ExpiresAt: expiresAt,
	}, nil
}

func (s Store) UsePasswordResetEntry(
	username screenjournal.Username,
	token screenjournal.PasswordResetToken,
	newPasswordHash screenjournal.PasswordHash,
	now time.Time,
) error {
	tx, err := s.beginTx()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback password reset transaction: %v", err)
		}
	}()

	var entryUsernameRaw string
	var expiresAtRaw string
	if err := tx.QueryRow(`
		SELECT
			username,
			expires_at
		FROM
			password_reset_tokens
		WHERE
			token = :token`, sql.Named("token", token.String())).Scan(&entryUsernameRaw, &expiresAtRaw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.ErrInvalidPasswordResetToken
		}
		return err
	}

	entryUsername := screenjournal.Username(entryUsernameRaw)
	if !entryUsername.Equal(username) {
		return store.ErrInvalidPasswordResetToken
	}

	expiresAt, err := parseDatetime(expiresAtRaw)
	if err != nil {
		return err
	}
	if now.After(expiresAt) {
		if _, err := tx.Exec(`
			DELETE FROM
				password_reset_tokens
			WHERE
				token = :token`,
			sql.Named("token", token.String())); err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		return store.ErrExpiredPasswordResetToken
	}

	updateResult, err := tx.Exec(`
		UPDATE users
		SET
			password_hash = :password_hash
		WHERE
			username = :username`,
		sql.Named("password_hash", encodePasswordHash(newPasswordHash)),
		sql.Named("username", entryUsername.String()))
	if err != nil {
		return err
	}
	rowsUpdated, err := updateResult.RowsAffected()
	if err != nil {
		return err
	}
	if rowsUpdated != 1 {
		return store.ErrUserNotFound
	}

	deleteResult, err := tx.Exec(`
		DELETE FROM
			password_reset_tokens
		WHERE
			token = :token`,
		sql.Named("token", token.String()))
	if err != nil {
		return err
	}
	rowsDeleted, err := deleteResult.RowsAffected()
	if err != nil {
		return err
	}
	if rowsDeleted != 1 {
		return store.ErrInvalidPasswordResetToken
	}

	return tx.Commit()
}
