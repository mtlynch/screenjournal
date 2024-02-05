package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
)

func (s Store) ReadReview(id screenjournal.ReviewID) (screenjournal.Review, error) {
	row := s.ctx.QueryRow(`
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

	review.Movie, err = s.ReadMovie(review.Movie.ID)
	if err != nil {
		return screenjournal.Review{}, err
	}

	return review, nil
}

func (s Store) ReadReviews(opts ...store.ReadReviewsOption) ([]screenjournal.Review, error) {
	params := store.ReadReviewsParams{}
	for _, o := range opts {
		o(&params)
	}
	whereClauses := []string{}
	queryArgs := []any{}

	if params.Filters.Username != nil {
		whereClauses = append(whereClauses, "review_owner = ?")
		queryArgs = append(queryArgs, params.Filters.Username.String())
	}
	if params.Filters.MovieID != nil {
		whereClauses = append(whereClauses, "movie_id = ?")
		queryArgs = append(queryArgs, params.Filters.MovieID.Int64())
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
	query += "\nORDER BY"
	if params.Order != nil && *params.Order == screenjournal.ByRating {
		query += "		rating DESC,\n"
	} else {
		query += "		watched_date DESC,\n"
	}
	query += "		created_time DESC\n"

	rows, err := s.ctx.Query(query, queryArgs...)
	if err != nil {
		return []screenjournal.Review{}, err
	}

	reviews := []screenjournal.Review{}
	for rows.Next() {
		review, err := reviewFromRow(rows)
		if err != nil {
			return []screenjournal.Review{}, err
		}

		if review.Movie, err = s.ReadMovie(review.Movie.ID); err != nil {
			return []screenjournal.Review{}, err
		}

		if review.Comments, err = s.ReadComments(review.ID); err != nil {
			return []screenjournal.Review{}, err
		}

		reviews = append(reviews, review)
	}

	return reviews, nil
}

func (s Store) InsertReview(r screenjournal.Review) (screenjournal.ReviewID, error) {
	log.Printf("inserting new review of movie ID %v: %v", r.Movie.ID, r.Rating.UInt8())

	now := time.Now()

	res, err := s.ctx.Exec(`
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

func (s Store) UpdateReview(r screenjournal.Review) error {
	log.Printf("updating review of movie ID %v: %v", r.Movie.ID, r.Rating.UInt8())

	if r.ID.IsZero() {
		return errors.New("invalid review ID")
	}

	now := time.Now()

	if _, err := s.ctx.Exec(`
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

func (s Store) DeleteReview(id screenjournal.ReviewID) error {
	log.Printf("deleting review of movie ID %v", id)

	tx, err := s.ctx.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM reviews WHERE id = ?`, id.UInt64()); err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM review_comments WHERE review_id = ?`, id.UInt64()); err != nil {
		return err
	}

	return tx.Commit()
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
