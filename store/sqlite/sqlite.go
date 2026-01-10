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
		Scan(...interface{}) error
	}
)

func New(path string, optimizeForLitestream bool) Store {
	ctx, err := OpenDB(path)
	if err != nil {
		log.Fatalln(err)
	}
	return NewFromDB(ctx, optimizeForLitestream)
}

func OpenDB(path string) (*sql.DB, error) {
	log.Printf("reading DB from %s", path)
	ctx, err := driver.Open(path)
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

func NewFromDB(ctx *sql.DB, optimizeForLitestream bool) Store {
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
