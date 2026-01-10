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
		tv_show_id,
		tv_show_season,
		rating,
		blurb,
		watched_date,
		created_time,
		last_modified_time
	FROM
		reviews
	WHERE
		reviews.id = :id`, sql.Named("id", id))

	review, err := reviewFromRow(row)
	if err != nil {
		return screenjournal.Review{}, err
	}

	if !review.Movie.ID.IsZero() {
		review.Movie, err = s.ReadMovie(review.Movie.ID)
		if err != nil {
			return screenjournal.Review{}, err
		}
	} else {
		review.TvShow, err = s.ReadTvShow(review.TvShow.ID)
		if err != nil {
			return screenjournal.Review{}, err
		}
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
		whereClauses = append(whereClauses, "review_owner = :username")
		queryArgs = append(queryArgs, sql.Named("username", params.Filters.Username.String()))
	}
	if params.Filters.MovieID != nil {
		whereClauses = append(whereClauses, "movie_id = :movie_id")
		queryArgs = append(queryArgs, sql.Named("movie_id", params.Filters.MovieID.Int64()))
	}
	if params.Filters.TvShowID != nil {
		whereClauses = append(whereClauses, "tv_show_id = :tv_show_id")
		queryArgs = append(queryArgs, sql.Named("tv_show_id", params.Filters.TvShowID.Int64()))
	}
	if params.Filters.TvShowSeason != nil {
		whereClauses = append(whereClauses, "tv_show_season = :tv_show_season")
		queryArgs = append(queryArgs, sql.Named("tv_show_season", params.Filters.TvShowSeason.UInt8()))
	}

	query := `
	SELECT
		id,
		review_owner,
		movie_id,
		tv_show_id,
		tv_show_season,
		rating,
		blurb,
		watched_date,
		created_time,
		last_modified_time
	FROM
		reviews`
	if len(queryArgs) > 0 {
		query += fmt.Sprintf("\n\tWHERE\n\t\t%s", strings.Join(whereClauses, " AND\n\t\t"))
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
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close review rows: %v", err)
		}
	}()

	reviews := []screenjournal.Review{}
	for rows.Next() {
		review, err := reviewFromRow(rows)
		if err != nil {
			return []screenjournal.Review{}, err
		}
		reviews = append(reviews, review)
	}
	if err := rows.Err(); err != nil {
		return []screenjournal.Review{}, err
	}

	// Populate the fields once the first SQL query is complete.
	for i := range reviews {
		review := reviews[i]
		if !review.Movie.ID.IsZero() {
			if review.Movie, err = s.ReadMovie(review.Movie.ID); err != nil {
				return []screenjournal.Review{}, err
			}
		} else {
			if review.TvShow, err = s.ReadTvShow(review.TvShow.ID); err != nil {
				return []screenjournal.Review{}, err
			}
		}

		if review.Comments, err = s.ReadComments(review.ID); err != nil {
			return []screenjournal.Review{}, err
		}
		reviews[i] = review
	}

	return reviews, nil
}

func (s Store) InsertReview(r screenjournal.Review) (screenjournal.ReviewID, error) {
	if r.MediaType() == screenjournal.MediaTypeMovie {
		log.Printf("inserting new review of movie ID %v: %s", r.Movie.ID, r.Rating)
	} else {
		log.Printf("inserting new review of TV show ID %v: %s", r.TvShow.ID, r.Rating)
	}

	now := time.Now()

	var movieID *screenjournal.MovieID
	var tvShowID *screenjournal.TvShowID
	var tvShowSeason *screenjournal.TvShowSeason
	if !r.Movie.ID.IsZero() {
		movieID = &r.Movie.ID
	} else {
		tvShowID = &r.TvShow.ID
		tvShowSeason = &r.TvShowSeason
	}

	res, err := s.ctx.Exec(`
	INSERT INTO
		reviews
	(
		review_owner,
		movie_id,
		tv_show_id,
		tv_show_season,
		rating,
		blurb,
		watched_date,
		created_time,
		last_modified_time
	)
	VALUES (
		:owner, :movie_id, :tv_show_id, :tv_show_season, :rating, :blurb, :watched_date, :created_time, :last_modified_time
	)
	`,
		sql.Named("owner", r.Owner),
		sql.Named("movie_id", movieID),
		sql.Named("tv_show_id", tvShowID),
		sql.Named("tv_show_season", tvShowSeason),
		sql.Named("rating", r.Rating.Value),
		sql.Named("blurb", r.Blurb),
		sql.Named("watched_date", formatWatchDate(r.Watched)),
		sql.Named("created_time", formatTime(now)),
		sql.Named("last_modified_time", formatTime(now)))
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
	if r.MediaType() == screenjournal.MediaTypeMovie {
		log.Printf("updating review of movie ID %v: %s", r.Movie.ID, r.Rating)
	} else {
		log.Printf("updating review of TV show ID %v: %s", r.TvShow.ID, r.Rating)
	}

	if r.ID.IsZero() {
		return errors.New("invalid review ID")
	}

	now := time.Now()

	if _, err := s.ctx.Exec(`
	UPDATE reviews
	SET
		rating = :rating,
		blurb = :blurb,
		watched_date = :watched_date,
		last_modified_time = :last_modified_time
	WHERE
		id = :id`,
		sql.Named("rating", r.Rating.Value),
		sql.Named("blurb", r.Blurb),
		sql.Named("watched_date", formatWatchDate(r.Watched)),
		sql.Named("last_modified_time", formatTime(now)),
		sql.Named("id", r.ID.UInt64())); err != nil {
		return err
	}

	return nil
}

func (s Store) DeleteReview(id screenjournal.ReviewID) error {
	log.Printf("deleting review and commments for review ID %v", id)

	tx, err := s.ctx.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback delete review: %v", err)
		}
	}()

	if _, err := tx.Exec(`DELETE FROM reviews WHERE id = :id`, sql.Named("id", id.UInt64())); err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM review_comments WHERE review_id = :review_id`, sql.Named("review_id", id.UInt64())); err != nil {
		return err
	}

	return tx.Commit()
}

func reviewFromRow(row rowScanner) (screenjournal.Review, error) {
	var id int
	var owner string
	var movieIDRaw *int
	var tvShowIDRaw *int
	var tvShowSeason *int
	var ratingRaw *uint8
	var blurb string
	var watchedDateRaw string
	var createdTimeRaw string
	var lastModifiedTimeRaw string

	err := row.Scan(&id, &owner, &movieIDRaw, &tvShowIDRaw, &tvShowSeason, &ratingRaw, &blurb, &watchedDateRaw, &createdTimeRaw, &lastModifiedTimeRaw)
	if err == sql.ErrNoRows {
		return screenjournal.Review{}, store.ErrReviewNotFound
	} else if err != nil {
		log.Printf("failed to read review from SQL row: %v", err)
		return screenjournal.Review{}, err
	}

	var movieID screenjournal.MovieID
	if movieIDRaw != nil {
		movieID = screenjournal.MovieID(*movieIDRaw)
	}

	var tvShowID screenjournal.TvShowID
	if tvShowIDRaw != nil {
		tvShowID = screenjournal.TvShowID(*tvShowIDRaw)
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

	var season screenjournal.TvShowSeason
	if tvShowSeason != nil {
		season = screenjournal.TvShowSeason(*tvShowSeason)
	}

	var rating screenjournal.Rating
	if ratingRaw != nil {
		rating = screenjournal.NewRating(*ratingRaw)
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
			ID: screenjournal.MovieID(movieID),
		},
		TvShow: screenjournal.TvShow{
			ID: screenjournal.TvShowID(tvShowID),
		},
		TvShowSeason: season,
	}, nil
}
