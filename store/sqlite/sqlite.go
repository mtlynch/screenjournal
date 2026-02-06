package sqlite

import (
	"database/sql"
	"log"
	"time"

	"github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

const (
	timeFormat = time.RFC3339
)

type (
	Store struct {
		ctx *sql.DB
	}

	rowScanner interface {
		Scan(...any) error
	}
)

func MustOpen(path string) *sql.DB {
	log.Printf("reading DB from %s", path)
	ctx, err := driver.Open(path)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	return ctx
}

func New(ctx *sql.DB, optimizeForLitestream bool) Store {
	if _, err := ctx.Exec(`
		PRAGMA temp_store = FILE;
		PRAGMA journal_mode = WAL;
		`); err != nil {
		log.Fatalf("failed to set pragmas: %v", err)
	}

	store := Store{ctx: ctx}
	if optimizeForLitestream {
		store.optimizeForLitestream()
	}

	store.applyMigrations()

	return store
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
