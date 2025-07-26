package sqlite

import (
	"database/sql"
	"log"
	"time"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func (s Store) InsertPasswordResetEntry(request screenjournal.PasswordResetEntry) error {
	log.Printf("inserting new password reset token for user %s", request.Username)

	if _, err := s.ctx.Exec(`
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
		sql.Named("token", request.Token),
		sql.Named("expires_at", formatTime(request.ExpiresAt))); err != nil {
		return err
	}

	return nil
}

func (s Store) ReadPasswordResetEntry(token screenjournal.PasswordResetToken) (screenjournal.PasswordResetEntry, error) {
	var username string
	var expiresAtRaw string
	if err := s.ctx.QueryRow(`
		SELECT
			username,
			expires_at
		FROM
			password_reset_tokens
		WHERE
			token = :token`, sql.Named("token", token)).Scan(&username, &expiresAtRaw); err != nil {
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

func (s Store) ReadPasswordResetEntries() ([]screenjournal.PasswordResetEntry, error) {
	rows, err := s.ctx.Query(`
		SELECT
			username,
			token,
			expires_at
		FROM
			password_reset_tokens
		ORDER BY
			expires_at DESC`)
	if err != nil {
		return []screenjournal.PasswordResetEntry{}, err
	}
	defer func() {
		if err := rows.Close(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to close rows after reading password reset tokens: %v", err)
		}
	}()

	requests := []screenjournal.PasswordResetEntry{}
	for rows.Next() {
		var usernameRaw string
		var tokenRaw string
		var expiresAtRaw string
		if err := rows.Scan(&usernameRaw, &tokenRaw, &expiresAtRaw); err != nil {
			return []screenjournal.PasswordResetEntry{}, err
		}

		expiresAt, err := parseDatetime(expiresAtRaw)
		if err != nil {
			return []screenjournal.PasswordResetEntry{}, err
		}

		requests = append(requests, screenjournal.PasswordResetEntry{
			Username:  screenjournal.Username(usernameRaw),
			Token:     screenjournal.PasswordResetToken(tokenRaw),
			ExpiresAt: expiresAt,
		})
	}

	return requests, nil
}

func (s Store) DeletePasswordResetEntry(token screenjournal.PasswordResetToken) error {
	log.Printf("deleting password reset token: %s", token)
	if _, err := s.ctx.Exec(`DELETE FROM password_reset_tokens WHERE token = :token`, sql.Named("token", token.String())); err != nil {
		return err
	}
	return nil
}

func (s Store) DeleteExpiredPasswordResetEntries() error {
	log.Printf("deleting expired password reset tokens")
	now := time.Now()
	if _, err := s.ctx.Exec(`DELETE FROM password_reset_tokens WHERE expires_at < :now`, sql.Named("now", formatTime(now))); err != nil {
		return err
	}
	return nil
}
