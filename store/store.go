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
		ReadUser(screenjournal.Username) (screenjournal.User, error)
		InsertUser(screenjournal.User) error
		InsertSignupInvitation(screenjournal.SignupInvitation) error
		ReadSignupInvitation(screenjournal.InviteCode) (screenjournal.SignupInvitation, error)
		ReadSignupInvitations() ([]screenjournal.SignupInvitation, error)
		DeleteSignupInvitation(screenjournal.InviteCode) error
	}
)

var (
	ErrMovieNotFound                     = errors.New("could not find movie")
	ErrReviewNotFound                    = errors.New("could not find review")
	ErrUsernameNotAvailable              = errors.New("username is not available")
	ErrEmailAssociatedWithAnotherAccount = errors.New("email address is associated with another account")
)
