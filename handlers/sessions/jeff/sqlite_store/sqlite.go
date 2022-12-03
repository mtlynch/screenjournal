package sqlite_store

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// Store satisfies the jeff.Storage interface
type Store struct {
	db *sql.DB
}

const sqliteDatetimeFormat = "2006-01-02 15:04:05"

// New initializes a new sqlite Storage for jeff
func New(db *sql.DB) (*Store, error) {
	return NewWithCleanupInterval(db, 5*time.Minute)
}

// NewWithCleanupInterval returns a new SQLite3Store instance. The cleanupInterval
// parameter controls how frequently expired session data is removed by the
// background cleanup goroutine. Setting it to 0 prevents the cleanup goroutine
// from running (i.e. expired sessions will not be removed).
func NewWithCleanupInterval(db *sql.DB, cleanupInterval time.Duration) (*Store, error) {
	tableName := "sessions" // TODO: Make this an option to New
	if _, err := db.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			key TEXT PRIMARY KEY,
			value BLOB,
			expires_at TEXT NOT NULL
		)`, tableName)); err != nil {
		return nil, err
	}

	s := &Store{db: db}
	if cleanupInterval > 0 {
		go s.startCleanup(cleanupInterval)
	}
	return s, nil
}

// Store satisfies the jeff.Store.Store method
func (s *Store) Store(_ context.Context, key, value []byte, exp time.Time) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO
				sessions
		(
				key,
				value,
				expires_at
		)
		VALUES
		(
				?,
				?,
				?
		)`,
		string(key), value, exp.Format(sqliteDatetimeFormat))
	if err != nil {
		return err
	}
	return nil
}

// Fetch satisfies the jeff.Store.Fetch method
func (s *Store) Fetch(ctx context.Context, key []byte) ([]byte, error) {
	var value []byte
	if err := s.db.QueryRow(`
	SELECT
		value
	FROM
		sessions
	WHERE
		key = ? AND
		expires_at > datetime('now', 'localtime')`, string(key)).Scan(&value); err != nil {
		// Not found sessions must return nil value, nil error
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return value, nil
}

// Delete satisfies the jeff.Store.Delete method
func (s *Store) Delete(ctx context.Context, key []byte) error {
	log.Printf("deleting session key %v from SQLite store", string(key))
	_, err := s.db.Exec(`DELETE FROM sessions WHERE key = ?`, string(key))
	return err
}

func (p *Store) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		err := p.deleteExpired()
		if err != nil {
			log.Println(err)
		}
	}
}

func (p *Store) deleteExpired() error {
	_, err := p.db.Exec(`
		DELETE FROM
			sessions
		WHERE
			expires_at < datetime('now', 'localtime')`)
	return err
}
