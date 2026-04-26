package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	simple_sessions "codeberg.org/mtlynch/simpleauth/v3/sessions"
)

const sessionDatetimeFormat = "2006-01-02 15:04:05"

type SessionStore struct {
	dbFunc func(context.Context) *sql.DB
}

func NewSessionStore(dbFunc func(context.Context) *sql.DB) SessionStore {
	return SessionStore{dbFunc: dbFunc}
}

func (s SessionStore) CreateSession(
	ctx context.Context,
	session simple_sessions.Session,
) error {
	if _, err := s.dbFunc(ctx).ExecContext(ctx, `
		INSERT OR REPLACE INTO auth_sessions
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
		sql.Named("created_at", formatSessionTime(session.CreatedAt)),
		sql.Named("expires_at", formatSessionTime(session.ExpiresAt)),
	); err != nil {
		return err
	}
	return nil
}

func (s SessionStore) ReadSession(
	ctx context.Context,
	id simple_sessions.ID,
) (simple_sessions.Session, error) {
	var userIDRaw string
	var createdAtRaw string
	var expiresAtRaw string
	if err := s.dbFunc(ctx).QueryRowContext(ctx, `
		SELECT
			user_id,
			created_at,
			expires_at
		FROM
			auth_sessions
		WHERE
			session_id = :session_id`,
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
	createdAt, err := parseSessionTime(createdAtRaw)
	if err != nil {
		return simple_sessions.Session{}, fmt.Errorf(
			"parse stored created at: %w",
			err,
		)
	}

	// We don't filter out expired sessions at this layer because simpleauth is
	// responsible for handling session expiration.
	expiresAt, err := parseSessionTime(expiresAtRaw)
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

func (s SessionStore) DeleteSession(
	ctx context.Context,
	id simple_sessions.ID,
) error {
	if _, err := s.dbFunc(ctx).ExecContext(ctx, `
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

func formatSessionTime(t time.Time) string {
	return t.Format(sessionDatetimeFormat)
}

func parseSessionTime(raw string) (time.Time, error) {
	return time.Parse(sessionDatetimeFormat, raw)
}
