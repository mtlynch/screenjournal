package test_sqlite

import (
	"log"
	"os"
	"testing"

	"github.com/ncruces/go-sqlite3/vfs/memdb"

	"github.com/mtlynch/screenjournal/v2/store/sqlite"
)

func New(t testing.TB) sqlite.Store {
	t.Helper()

	// Suppress log output, as the migration logs are too noisy during tests.
	defer quietLogs()()
	const optimizeForLitestream = false
	// Use memdb instead of shared-cache in-memory DBs:
	// https://github.com/ncruces/go-sqlite3/discussions/97#discussioncomment-9744524.
	db := sqlite.MustOpen(memdb.TestDB(t))
	t.Cleanup(func() {
		t.Helper()
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close db: %v", err)
		}
	})
	return sqlite.New(db, optimizeForLitestream)
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
