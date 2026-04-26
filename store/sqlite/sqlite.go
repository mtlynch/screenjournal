package sqlite

import (
	"context"
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
		dbFunc func(context.Context) *sql.DB
		ctx    context.Context
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

func New(dbFunc func(context.Context) *sql.DB, optimizeForLitestream bool) Store {
	db := dbFunc(context.Background())
	if _, err := db.Exec(`
		PRAGMA temp_store = FILE;
		PRAGMA journal_mode = WAL;
		`); err != nil {
		log.Fatalf("failed to set pragmas: %v", err)
	}

	store := Store{
		dbFunc: dbFunc,
		ctx:    context.Background(),
	}
	if optimizeForLitestream {
		store.optimizeForLitestream()
	}

	store.applyMigrations()

	return store
}

func (s Store) WithContext(ctx context.Context) Store {
	s.ctx = ctx
	return s
}

func (s Store) db() *sql.DB {
	return s.dbFunc(s.ctx)
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
