package store

import (
	"errors"

	"github.com/mtlynch/screenjournal/v2"
)

type (
	ReviewFilters struct {
		Username *screenjournal.Username
		MovieID  *screenjournal.MovieID
	}

	Store interface {
		ReadReview(screenjournal.ReviewID) (screenjournal.Review, error)
		ReadReviews(ReviewFilters) ([]screenjournal.Review, error)
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
		ReadNotificationSubscribers() ([]screenjournal.User, error)
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
