package sqlite

import (
	"database/sql"
	"log"
	"time"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func (s Store) InsertPasswordResetRequest(request screenjournal.PasswordResetRequest) error {
	log.Printf("inserting new password reset token for user %s", request.Username)

	if _, err := s.ctx.Exec(`
	INSERT INTO
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

func (s Store) ReadPasswordResetRequest(token screenjournal.PasswordResetToken) (screenjournal.PasswordResetRequest, error) {
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
		return screenjournal.PasswordResetRequest{}, err
	}

	expiresAt, err := parseDatetime(expiresAtRaw)
	if err != nil {
		return screenjournal.PasswordResetRequest{}, err
	}

	return screenjournal.PasswordResetRequest{
		Username:  screenjournal.Username(username),
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

func (s Store) ReadPasswordResetRequests() ([]screenjournal.PasswordResetRequest, error) {
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
		return []screenjournal.PasswordResetRequest{}, err
	}
	defer func() {
		if err := rows.Close(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to close rows after reading password reset tokens: %v", err)
		}
	}()

	requests := []screenjournal.PasswordResetRequest{}
	for rows.Next() {
		var usernameRaw string
		var tokenRaw string
		var expiresAtRaw string
		if err := rows.Scan(&usernameRaw, &tokenRaw, &expiresAtRaw); err != nil {
			return []screenjournal.PasswordResetRequest{}, err
		}

		expiresAt, err := parseDatetime(expiresAtRaw)
		if err != nil {
			return []screenjournal.PasswordResetRequest{}, err
		}

		requests = append(requests, screenjournal.PasswordResetRequest{
			Username:  screenjournal.Username(usernameRaw),
			Token:     screenjournal.PasswordResetToken(tokenRaw),
			ExpiresAt: expiresAt,
		})
	}

	return requests, nil
}

func (s Store) DeletePasswordResetRequest(token screenjournal.PasswordResetToken) error {
	log.Printf("deleting password reset token: %s", token)
	_, err := s.ctx.Exec(`DELETE FROM password_reset_tokens WHERE token = :token`, sql.Named("token", token.String()))
	if err != nil {
		return err
	}
	return nil
}

func (s Store) DeleteExpiredPasswordResetRequests() error {
	log.Printf("deleting expired password reset tokens")
	now := time.Now()
	_, err := s.ctx.Exec(`DELETE FROM password_reset_tokens WHERE expires_at < :now`, sql.Named("now", formatTime(now)))
	if err != nil {
		return err
	}
	return nil
}
