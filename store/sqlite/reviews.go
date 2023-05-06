package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
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

	review, err := reviewFromRow(row)
	if err != nil {
		return screenjournal.Review{}, err
	}

	review.Movie, err = db.ReadMovie(review.Movie.ID)
	if err != nil {
		return screenjournal.Review{}, err
	}

	return review, nil
}

func (db DB) ReadReviews(filters store.ReviewFilters) ([]screenjournal.Review, error) {
	whereClauses := []string{}
	queryArgs := []any{}

	if filters.Username != nil {
		whereClauses = append(whereClauses, "review_owner = ?")
		queryArgs = append(queryArgs, filters.Username.String())
	}
	if filters.MovieID != nil {
		whereClauses = append(whereClauses, "movie_id = ?")
		queryArgs = append(queryArgs, filters.MovieID.Int64())
	}

	query := `
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
		reviews`
	if len(queryArgs) > 0 {
		query += fmt.Sprintf("\n\tWHERE\n\t\t%s", strings.Join(whereClauses, "\n\t\t"))
	}
	query += `
	ORDER BY
		watched_date DESC,
		created_time DESC`

	rows, err := db.ctx.Query(query, queryArgs...)
	if err != nil {
		return []screenjournal.Review{}, err
	}

	reviews := []screenjournal.Review{}
	for rows.Next() {
		review, err := reviewFromRow(rows)
		if err != nil {
			return []screenjournal.Review{}, err
		}

		if review.Movie, err = db.ReadMovie(review.Movie.ID); err != nil {
			return []screenjournal.Review{}, err
		}

		if review.Comments, err = db.ReadComments(review.ID); err != nil {
			return []screenjournal.Review{}, err
		}

		reviews = append(reviews, review)
	}

	return reviews, nil
}

func (d DB) InsertReview(r screenjournal.Review) (screenjournal.ReviewID, error) {
	log.Printf("inserting new review of movie ID %v: %v", r.Movie.ID, r.Rating.UInt8())

	now := time.Now()

	res, err := d.ctx.Exec(`
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
		r.Movie.ID,
		r.Rating,
		r.Blurb,
		formatWatchDate(r.Watched),
		formatTime(now),
		formatTime(now))
	if err != nil {
		return screenjournal.ReviewID(0), err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return screenjournal.ReviewID(0), err
	}

	return screenjournal.ReviewID(lastID), nil
}

func (d DB) UpdateReview(r screenjournal.Review) error {
	log.Printf("updating review of movie ID %v: %v", r.Movie.ID, r.Rating.UInt8())

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
	var rating screenjournal.Rating
	var blurb string
	var watchedDateRaw string
	var createdTimeRaw string
	var lastModifiedTimeRaw string

	err := row.Scan(&id, &owner, &movie_id, &rating, &blurb, &watchedDateRaw, &createdTimeRaw, &lastModifiedTimeRaw)
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
		Rating:   rating,
		Blurb:    screenjournal.Blurb(blurb),
		Watched:  screenjournal.WatchDate(wd),
		Created:  ct,
		Modified: lmt,
		Movie: screenjournal.Movie{
			ID: screenjournal.MovieID(movie_id),
		},
	}, nil
}
