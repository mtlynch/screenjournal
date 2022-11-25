package sqlite

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/store"
)

func (db DB) ReadReview(id screenjournal.ReviewID) (screenjournal.Review, error) {
	row := db.ctx.QueryRow(`
	SELECT
		id,
		review_owner,
		movie_id,
		rating,
		blurb,
		watched_date,
		created_time,
		last_modified_time
	FROM
		reviews
	WHERE
		reviews.id = ?`, id)

	return reviewFromRow(row)
}

func (db DB) ReadReviews() ([]screenjournal.Review, error) {
	rows, err := db.ctx.Query(`
	SELECT
		id,
		review_owner,
		movie_id,
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
		review, err := reviewFromRow(rows)
		if err != nil {
			return []screenjournal.Review{}, err
		}

		reviews = append(reviews, review)
	}

	return reviews, nil
}

func (d DB) InsertReview(r screenjournal.Review) error {
	log.Printf("inserting new review of media ID %v: %v", r.MovieID, r.Rating.UInt8())

	now := time.Now()

	if _, err := d.ctx.Exec(`
	INSERT INTO
		reviews
	(
		review_owner,
		movie_id,
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
		r.MovieID,
		r.Rating,
		r.Blurb,
		formatWatchDate(r.Watched),
		formatTime(now),
		formatTime(now)); err != nil {
		return err
	}

	return nil
}

func (d DB) UpdateReview(r screenjournal.Review) error {
	log.Printf("updating review of media ID %v: %v", r.MovieID, r.Rating.UInt8())

	if r.ID.IsZero() {
		return errors.New("invalid review ID")
	}

	now := time.Now()

	if _, err := d.ctx.Exec(`
	UPDATE reviews
	SET
		rating = ?,
		blurb = ?,
		watched_date = ?,
		last_modified_time = ?
	WHERE
		id = ?`,
		r.Rating,
		r.Blurb,
		formatWatchDate(r.Watched),
		formatTime(now),
		r.ID.UInt64()); err != nil {
		return err
	}

	return nil
}

func reviewFromRow(row rowScanner) (screenjournal.Review, error) {
	var id int
	var owner string
	var movie_id int
	var title string
	var rating screenjournal.Rating
	var blurb string
	var watchedDateRaw string
	var createdTimeRaw string
	var lastModifiedTimeRaw string

	err := row.Scan(&id, &owner, &movie_id, &title, &rating, &blurb, &watchedDateRaw, &createdTimeRaw, &lastModifiedTimeRaw)
	if err == sql.ErrNoRows {
		return screenjournal.Review{}, store.ErrReviewNotFound
	} else if err != nil {
		return screenjournal.Review{}, err
	}

	wd, err := parseDatetime(watchedDateRaw)
	if err != nil {
		return screenjournal.Review{}, err
	}

	ct, err := parseDatetime(createdTimeRaw)
	if err != nil {
		return screenjournal.Review{}, err
	}

	lmt, err := parseDatetime(lastModifiedTimeRaw)
	if err != nil {
		return screenjournal.Review{}, err
	}

	return screenjournal.Review{
		ID:       screenjournal.ReviewID(id),
		Owner:    screenjournal.Username(owner),
		MovieID:  screenjournal.MovieID(movie_id),
		Rating:   rating,
		Blurb:    screenjournal.Blurb(blurb),
		Watched:  screenjournal.WatchDate(wd),
		Created:  ct,
		Modified: lmt,
	}, nil
}
