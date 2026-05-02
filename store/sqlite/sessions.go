package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	simple_sessions "codeberg.org/mtlynch/simpleauth/v3/sessions"
)

func (s Store) CreateSession(
	ctx context.Context,
	session simple_sessions.Session,
) error {
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO auth_sessions
		(
			session_id,
			user_id,
			created_at,
			expires_at
		)
		VALUES
		(
			:session_id,
			:user_id,
			:created_at,
			:expires_at
		)`,
		sql.Named("session_id", session.ID.String()),
		sql.Named("user_id", session.UserID.String()),
		sql.Named("created_at", formatTime(session.CreatedAt)),
		sql.Named("expires_at", formatTime(session.ExpiresAt)),
	); err != nil {
		return err
	}
	return nil
}

func (s Store) ReadSession(
	ctx context.Context,
	id simple_sessions.ID,
) (simple_sessions.Session, error) {
	var userIDRaw string
	var createdAtRaw string
	var expiresAtRaw string
	if err := s.db.QueryRowContext(ctx, `
		SELECT
			user_id,
			created_at,
			expires_at
		FROM
			auth_sessions
		WHERE
			session_id = :session_id AND
			datetime(expires_at) > datetime('now')`,
		sql.Named("session_id", id.String()),
	).Scan(&userIDRaw, &createdAtRaw, &expiresAtRaw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return simple_sessions.Session{}, simple_sessions.ErrNoSessionFound
		}
		return simple_sessions.Session{}, err
	}

	userID, err := simple_sessions.NewUserID(userIDRaw)
	if err != nil {
		return simple_sessions.Session{}, fmt.Errorf("parse stored user ID: %w", err)
	}
	createdAt, err := parseDatetime(createdAtRaw)
	if err != nil {
		return simple_sessions.Session{}, fmt.Errorf(
			"parse stored created at: %w",
			err,
		)
	}

	expiresAt, err := parseDatetime(expiresAtRaw)
	if err != nil {
		return simple_sessions.Session{}, fmt.Errorf(
			"parse stored expires at: %w",
			err,
		)
	}
	return simple_sessions.Session{
		ID:        id,
		UserID:    userID,
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
	}, nil
}

func (s Store) DeleteSession(
	ctx context.Context,
	id simple_sessions.ID,
) error {
	if _, err := s.db.ExecContext(ctx, `
		DELETE FROM
			auth_sessions
		WHERE
			session_id = :session_id`,
		sql.Named("session_id", id.String()),
	); err != nil {
		return err
	}
	return nil
}

func (s Store) DeleteExpiredSessions(ctx context.Context, now time.Time) error {
	log.Printf("cleaning old sessions from store")

	if _, err := s.db.ExecContext(ctx, `
		DELETE FROM
			auth_sessions
		WHERE
			datetime(expires_at) <= datetime(:now)`,
		sql.Named("now", formatTime(now)),
	); err != nil {
		return err
	}
	return nil
}
