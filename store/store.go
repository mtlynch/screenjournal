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
}

var ErrReviewNotFound = errors.New("could not find review")
