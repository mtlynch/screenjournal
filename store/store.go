package store

import (
	"errors"

	"github.com/mtlynch/screenjournal/v2"
)

type Store interface {
	ReadReview(screenjournal.ReviewID) (screenjournal.Review, error)
	ReadReviews() ([]screenjournal.Review, error)
	InsertReview(screenjournal.Review) error
	UpdateReview(screenjournal.Review) error
	ReadMovie(screenjournal.MovieID) (screenjournal.Movie, error)
	ReadMovieByTmdbID(screenjournal.TmdbID) (screenjournal.Movie, error)
	InsertMovie(screenjournal.Movie) (screenjournal.MovieID, error)
}

var (
	ErrMovieNotFound  = errors.New("could not find movie")
	ErrReviewNotFound = errors.New("could not find review")
	ErrTmdbIDNotFound = errors.New("could not find media with matching TMDB ID")
)
