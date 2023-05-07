package store

import (
	"errors"

	"github.com/mtlynch/screenjournal/v2"
)

type (
	reviewFilters struct {
		Username *screenjournal.Username
		MovieID  *screenjournal.MovieID
	}

	ReadReviewsParams struct {
		Filters reviewFilters
		Order   *screenjournal.SortOrder
	}

	ReadReviewsOption func(*ReadReviewsParams)

	Store interface {
		ReadReview(screenjournal.ReviewID) (screenjournal.Review, error)
		ReadReviews(...ReadReviewsOption) ([]screenjournal.Review, error)
		InsertReview(screenjournal.Review) (screenjournal.ReviewID, error)
		UpdateReview(screenjournal.Review) error
		ReadComments(screenjournal.ReviewID) ([]screenjournal.ReviewComment, error)
		ReadComment(screenjournal.CommentID) (screenjournal.ReviewComment, error)
		InsertComment(screenjournal.ReviewComment) (screenjournal.CommentID, error)
		UpdateComment(screenjournal.ReviewComment) error
		DeleteComment(screenjournal.CommentID) error
		ReadMovie(screenjournal.MovieID) (screenjournal.Movie, error)
		ReadMovieByTmdbID(screenjournal.TmdbID) (screenjournal.Movie, error)
		InsertMovie(screenjournal.Movie) (screenjournal.MovieID, error)
		UpdateMovie(screenjournal.Movie) error
		CountUsers() (uint, error)
		ReadUser(screenjournal.Username) (screenjournal.User, error)
		ReadUsers() ([]screenjournal.User, error)
		InsertUser(screenjournal.User) error
		InsertSignupInvitation(screenjournal.SignupInvitation) error
		ReadSignupInvitation(screenjournal.InviteCode) (screenjournal.SignupInvitation, error)
		ReadSignupInvitations() ([]screenjournal.SignupInvitation, error)
		DeleteSignupInvitation(screenjournal.InviteCode) error
		ReadReviewSubscribers() ([]screenjournal.EmailSubscriber, error)
		ReadNotificationPreferences(screenjournal.Username) (screenjournal.NotificationPreferences, error)
		UpdateNotificationPreferences(screenjournal.Username, screenjournal.NotificationPreferences) error
	}
)

var (
	ErrMovieNotFound                     = errors.New("could not find movie")
	ErrCommentNotFound                   = errors.New("could not find comment")
	ErrReviewNotFound                    = errors.New("could not find review")
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

func SortReviews(order screenjournal.SortOrder) func(*ReadReviewsParams) {
	return func(p *ReadReviewsParams) {
		p.Order = &order
	}
}
