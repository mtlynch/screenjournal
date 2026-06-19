package test_sqlite

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/ncruces/go-sqlite3/vfs/memdb"

	"github.com/mtlynch/screenjournal/v2/random"
	"github.com/mtlynch/screenjournal/v2/store/sqlite"
)

const optimizeForLitestream = false

func New() sqlite.Store {
	_, store := newDBAndStore()
	return store
}

// NewDB returns an ephemeral SQLite database with all migrations applied.
func NewDB(t testing.TB) *sql.DB {
	t.Helper()
	db, _ := newDBAndStore()
	t.Cleanup(func() {
		t.Helper()
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close db: %v", err)
		}
	})
	return db
}

func newDBAndStore() (*sql.DB, sqlite.Store) {
	// Suppress log output, as the migration logs are too noisy during tests.
	defer quietLogs()()
	db := sqlite.MustOpen(ephemeralDbURI())
	// The ncruces/go-sqlite3 WASM driver does not implement cache=shared for
	// in-memory databases: each new connection gets an empty database instead
	// of sharing the named one. Limiting to one connection ensures every
	// caller uses the same underlying database.
	db.SetMaxOpenConns(1)
	store := sqlite.New(db, optimizeForLitestream)
	return db, store
}

func ephemeralDbURI() string {
	name := random.String(
		10,
		[]rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"))
	memdb.Create("/"+name, nil)

	// Use memdb instead of shared-cache in-memory DBs:
	// https://github.com/ncruces/go-sqlite3/discussions/97#discussioncomment-9744524
	return fmt.Sprintf("file:/%s?vfs=memdb", name)
}

// quietLogs suppresses log output during a function execution.
func quietLogs() func() {
	devNull, _ := os.Open(os.DevNull)
	stdout := os.Stdout
	stderr := os.Stderr
	os.Stdout = devNull
	os.Stderr = devNull
	log.SetOutput(devNull)
	return func() {
		defer func() {
			if err := devNull.Close(); err != nil {
				log.Printf("failed to close handle to /dev/null")
				return
			}
		}()
		os.Stdout = stdout
		os.Stderr = stderr
		log.SetOutput(os.Stderr)
	}
}
