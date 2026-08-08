package sqlite

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/ncruces/go-sqlite3"
	"github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

const (
	timeFormat = time.RFC3339
)

type (
	Store struct {
		db *sql.DB
	}

	rowScanner interface {
		Scan(...any) error
	}
)

func (s Store) Close() error {
	return s.db.Close()
}

func MustOpen(path string, optimizeForLitestream bool) *sql.DB {
	log.Printf("reading DB from %s", path)
	ctx, err := driver.Open(path, newConnInitializer(optimizeForLitestream))
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	return ctx
}

func New(db *sql.DB) Store {
	// journal_mode lives in the database file header, so it persists across
	// connections and only needs to run once.
	if _, err := db.Exec(`PRAGMA journal_mode = WAL`); err != nil {
		log.Fatalf("failed to enable WAL mode: %v", err)
	}

	store := Store{db: db}
	store.applyMigrations()

	return store
}

// newConnInitializer returns the callback that the driver runs on every new
// connection. SQLite scopes these pragmas to a single connection, so setting
// them through a single db.Exec would leave every other connection that
// database/sql opens on SQLite's defaults.
func newConnInitializer(optimizeForLitestream bool) func(*sqlite3.Conn) error {
	pragmas := []string{
		// Keep large temporary b-trees off the heap.
		`PRAGMA temp_store = FILE`,
		`PRAGMA foreign_keys = ON`,
	}
	if optimizeForLitestream {
		// Apply Litestream recommendations: https://litestream.io/tips/
		//
		// Litestream owns checkpointing, so SQLite must not checkpoint on its
		// own and discard WAL frames that Litestream has not replicated yet.
		//
		// busy_timeout is deliberately absent. The driver already applies a
		// one-minute timeout to every connection, which is stricter than
		// Litestream's suggested five seconds.
		pragmas = append(pragmas,
			`PRAGMA synchronous = NORMAL`,
			`PRAGMA wal_autocheckpoint = 0`,
		)
	}

	return func(conn *sqlite3.Conn) error {
		for _, pragma := range pragmas {
			if err := conn.Exec(pragma); err != nil {
				return fmt.Errorf("failed to apply %s: %w", pragma, err)
			}
		}
		return nil
	}
}

func parseDatetime(s string) (time.Time, error) {
	return time.Parse(timeFormat, s)
}

func formatTime(t time.Time) string {
	return t.Format(timeFormat)
}

func formatWatchDate(w screenjournal.WatchDate) string {
	return formatTime(w.Time())
}

func formatReleaseDate(rd screenjournal.ReleaseDate) string {
	return formatTime(rd.Time())
}
