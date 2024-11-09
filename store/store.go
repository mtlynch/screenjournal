package store

import (
	"errors"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type (
	reviewFilters struct {
		Username *screenjournal.Username
		MovieID  *screenjournal.MovieID
		TvShowID *screenjournal.TvShowID
	}

	ReadReviewsParams struct {
		Filters reviewFilters
		Order   *screenjournal.SortOrder
	}

	ReadReviewsOption func(*ReadReviewsParams)
)

var (
	ErrMovieNotFound                     = errors.New("could not find movie")
	ErrTvShowNotFound                    = errors.New("could not find TV show")
	ErrCommentNotFound                   = errors.New("could not find comment")
	ErrReviewNotFound                    = errors.New("could not find review")
	ErrUserNotFound                      = errors.New("could not find user")
	ErrUsernameNotAvailable              = errors.New("username is not available")
	ErrEmailAssociatedWithAnotherAccount = errors.New("email address is associated with another account")
)

func FilterReviewsByUsername(u screenjournal.Username) func(*ReadReviewsParams) {
	return func(p *ReadReviewsParams) {
		p.Filters.Username = &u
	}
}

func FilterReviewsByMovieID(id screenjournal.MovieID) func(*ReadReviewsParams) {
	return func(p *ReadReviewsParams) {
		p.Filters.MovieID = &id
	}
}

func FilterReviewsByTvShowID(id screenjournal.TvShowID) func(*ReadReviewsParams) {
	return func(p *ReadReviewsParams) {
		p.Filters.TvShowID = &id
	}
}

func SortReviews(order screenjournal.SortOrder) func(*ReadReviewsParams) {
	return func(p *ReadReviewsParams) {
		p.Order = &order
	}
}
