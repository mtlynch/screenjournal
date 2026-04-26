package test_sqlite

import (
	"fmt"
	"log"
	"os"

	"github.com/ncruces/go-sqlite3/vfs/memdb"

	"github.com/mtlynch/screenjournal/v2/random"
	"github.com/mtlynch/screenjournal/v2/store/sqlite"
)

func New() sqlite.Store {
	// Suppress log output, as the migration logs are too noisy during tests.
	defer quietLogs()()
	const optimizeForLitestream = false
	return sqlite.New(sqlite.MustOpen(ephemeralDbURI()), optimizeForLitestream)
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
