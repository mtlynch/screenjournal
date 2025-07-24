package handlers

import (
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
)

type Store interface {
	ReadReview(screenjournal.ReviewID) (screenjournal.Review, error)
	ReadReviews(...store.ReadReviewsOption) ([]screenjournal.Review, error)
	InsertReview(screenjournal.Review) (screenjournal.ReviewID, error)
	UpdateReview(screenjournal.Review) error
	DeleteReview(screenjournal.ReviewID) error
	ReadComments(screenjournal.ReviewID) ([]screenjournal.ReviewComment, error)
	ReadComment(screenjournal.CommentID) (screenjournal.ReviewComment, error)
	InsertComment(screenjournal.ReviewComment) (screenjournal.CommentID, error)
	UpdateComment(screenjournal.ReviewComment) error
	DeleteComment(screenjournal.CommentID) error
	ReadMovie(screenjournal.MovieID) (screenjournal.Movie, error)
	ReadMovieByTmdbID(screenjournal.TmdbID) (screenjournal.Movie, error)
	InsertMovie(screenjournal.Movie) (screenjournal.MovieID, error)
	UpdateMovie(screenjournal.Movie) error
	ReadTvShow(screenjournal.TvShowID) (screenjournal.TvShow, error)
	ReadTvShowByTmdbID(screenjournal.TmdbID) (screenjournal.TvShow, error)
	InsertTvShow(screenjournal.TvShow) (screenjournal.TvShowID, error)
	UpdateTvShow(screenjournal.TvShow) error
	CountUsers() (uint, error)
	ReadUser(screenjournal.Username) (screenjournal.User, error)
	ReadUsersPublicMeta() ([]screenjournal.UserPublicMeta, error)
	InsertUser(screenjournal.User) error
	UpdateUserPassword(screenjournal.Username, screenjournal.PasswordHash) error
	InsertSignupInvitation(screenjournal.SignupInvitation) error
	ReadSignupInvitation(screenjournal.InviteCode) (screenjournal.SignupInvitation, error)
	ReadSignupInvitations() ([]screenjournal.SignupInvitation, error)
	DeleteSignupInvitation(screenjournal.InviteCode) error
	ReadReviewSubscribers() ([]screenjournal.EmailSubscriber, error)
	ReadCommentSubscribers() ([]screenjournal.EmailSubscriber, error)
	ReadNotificationPreferences(screenjournal.Username) (screenjournal.NotificationPreferences, error)
	UpdateNotificationPreferences(screenjournal.Username, screenjournal.NotificationPreferences) error
	InsertPasswordResetRequest(screenjournal.PasswordResetRequest) error
	ReadPasswordResetRequest(screenjournal.PasswordResetToken) (screenjournal.PasswordResetRequest, error)
	ReadPasswordResetRequests() ([]screenjournal.PasswordResetRequest, error)
	DeletePasswordResetRequest(screenjournal.PasswordResetToken) error
	DeleteExpiredPasswordResetRequests() error
}
