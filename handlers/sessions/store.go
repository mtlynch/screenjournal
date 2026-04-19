package sessions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	simple_sessions "codeberg.org/mtlynch/simpleauth/v3/sessions"
)

const (
	sqliteDatetimeFormat = "2006-01-02 15:04:05"
	sessionLifetime      = 30 * 24 * time.Hour
)

type (
	store struct {
		db *sql.DB
	}

	storedSession struct {
		UserID    string    `json:"userID"`
		CreatedAt time.Time `json:"createdAt"`
	}
)

func (s store) CreateSession(ctx context.Context, session simple_sessions.Session) error {
	b, err := json.Marshal(storedSession{
		UserID:    session.UserID.String(),
		CreatedAt: session.CreatedAt,
	})
	if err != nil {
		return err
	}

	if _, err := s.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO auth_sessions
		(
			session_id,
			session_data,
			expires_at
		)
		VALUES
		(
			?,
			?,
			?
		)`,
		session.ID.String(),
		b,
		formatSessionExpiration(session.CreatedAt),
	); err != nil {
		return err
	}
	return nil
}

func (s store) ReadSession(ctx context.Context, id simple_sessions.ID) (simple_sessions.Session, error) {
	var b []byte
	if err := s.db.QueryRowContext(ctx, `
		SELECT
			session_data
		FROM
			auth_sessions
		WHERE
			session_id = ?
			AND expires_at > datetime('now', 'localtime')`,
		id.String(),
	).Scan(&b); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return simple_sessions.Session{}, simple_sessions.ErrNoSessionFound
		}
		return simple_sessions.Session{}, err
	}

	var stored storedSession
	if err := json.Unmarshal(b, &stored); err != nil {
		return simple_sessions.Session{}, err
	}
	userID, err := simple_sessions.NewUserID(stored.UserID)
	if err != nil {
		return simple_sessions.Session{}, fmt.Errorf("parse stored user ID: %w", err)
	}
	return simple_sessions.Session{
		ID:        id,
		UserID:    userID,
		CreatedAt: stored.CreatedAt,
	}, nil
}

func (s store) DeleteSession(ctx context.Context, id simple_sessions.ID) error {
	if _, err := s.db.ExecContext(ctx, `
		DELETE FROM
			auth_sessions
		WHERE
			session_id = ?`,
		id.String(),
	); err != nil {
		return err
	}
	return nil
}

func formatSessionExpiration(createdAt time.Time) string {
	return createdAt.Add(sessionLifetime).Format(sqliteDatetimeFormat)
}
