package store

import (
	"errors"

	"github.com/mtlynch/screenjournal/v2"
)

type (
	ReviewFilters struct {
		Username screenjournal.Username
	}

	Store interface {
		ReadReview(screenjournal.ReviewID) (screenjournal.Review, error)
		ReadReviews(ReviewFilters) ([]screenjournal.Review, error)
		InsertReview(screenjournal.Review) error
		UpdateReview(screenjournal.Review) error
		ReadMovie(screenjournal.MovieID) (screenjournal.Movie, error)
		ReadMovieByTmdbID(screenjournal.TmdbID) (screenjournal.Movie, error)
		InsertMovie(screenjournal.Movie) (screenjournal.MovieID, error)
		UpdateMovie(screenjournal.Movie) error
		CountUsers() (uint, error)
		InsertUser(screenjournal.User, screenjournal.PasswordHash) error
	}
)

var (
	ErrMovieNotFound  = errors.New("could not find movie")
	ErrReviewNotFound = errors.New("could not find review")
)
