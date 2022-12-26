package test_sqlite

import (
	"fmt"
	"log"
	"os"

	"github.com/mtlynch/screenjournal/v2/random"
	"github.com/mtlynch/screenjournal/v2/store"
	"github.com/mtlynch/screenjournal/v2/store/sqlite"
)

func New() store.Store {
	// Suppress log output, as the migration logs are too noisy during tests.
	defer quietLogs()()
	const optimizeForLitestream = false
	return sqlite.New(ephemeralDbURI(), optimizeForLitestream)
}

func ephemeralDbURI() string {
	name := random.String(
		10,
		[]rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"))
	return fmt.Sprintf("file:%s?mode=memory&cache=shared", name)
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
