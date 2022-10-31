package sqlite

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/store"
)

const (
	timeFormat = time.RFC3339
)

type (
	db struct {
		ctx *sql.DB
	}
)

func New(path string, optimizeForLitestream bool) store.Store {
	log.Printf("reading DB from %s", path)
	ctx, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatalln(err)
	}

	if _, err := ctx.Exec(`
		PRAGMA temp_store = FILE;
		PRAGMA journal_mode = WAL;
		`); err != nil {
		log.Fatalf("failed to set pragmas: %v", err)
	}

	d := &db{ctx: ctx}
	if optimizeForLitestream {
		d.optimizeForLitestream()
	}

	d.applyMigrations()

	return d
}

func (db db) ReadReviews() ([]screenjournal.Review, error) {
	rows, err := db.ctx.Query(`
	SELECT
		id,
		review_owner,
		title,
		rating,
		blurb,
		watched_date,
    created_time,
    last_modified_time
	FROM
		reviews`)
	if err != nil {
		return []screenjournal.Review{}, err
	}

	reviews := []screenjournal.Review{}
	for rows.Next() {
		var id int
		var owner string
		var title string
		var rating screenjournal.Rating
		var blurb string
		var watchedDateRaw string
		var createdTimeRaw string
		var lastModifiedTimeRaw string

		if err := rows.Scan(&id, &owner, &title, &rating, &blurb, &watchedDateRaw, &createdTimeRaw, &lastModifiedTimeRaw); err != nil {
			return []screenjournal.Review{}, err
		}

		wd, err := parseDatetime(watchedDateRaw)
		if err != nil {
			return []screenjournal.Review{}, err
		}

		ct, err := parseDatetime(createdTimeRaw)
		if err != nil {
			return []screenjournal.Review{}, err
		}

		lmt, err := parseDatetime(lastModifiedTimeRaw)
		if err != nil {
			return []screenjournal.Review{}, err
		}

		reviews = append(reviews, screenjournal.Review{
			ID:       screenjournal.ReviewID(id),
			Owner:    screenjournal.Username(owner),
			Title:    screenjournal.MediaTitle(title),
			Rating:   rating,
			Blurb:    screenjournal.Blurb(blurb),
			Watched:  screenjournal.WatchDate(wd),
			Created:  ct,
			Modified: lmt,
		})
	}

	return reviews, nil
}

func (d db) InsertReview(r screenjournal.Review) error {
	log.Printf("inserting new review of %s: %v", r.Title, r.Rating.UInt8())

	now := time.Now()

	if _, err := d.ctx.Exec(`
	INSERT INTO
		reviews
	(
		review_owner,
		title,
		rating,
		blurb,
		watched_date,
		created_time,
		last_modified_time
	)
	VALUES (
		?, ?, ?, ?, ?, ?, ?
	)
	`,
		r.Owner,
		r.Title,
		r.Rating,
		r.Blurb,
		formatWatchDate(r.Watched),
		formatTime(now),
		formatTime(now)); err != nil {
		return err
	}

	return nil
}

func parseDatetime(s string) (time.Time, error) {
	return time.Parse(timeFormat, s)
}

func formatTime(t time.Time) string {
	return t.Format(timeFormat)
}

func formatWatchDate(w screenjournal.WatchDate) string {
	return formatTime(time.Time(w))
}
