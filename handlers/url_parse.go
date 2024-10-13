package handlers

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

var (
	ErrMovieIDNotProvided   = errors.New("no movie ID in query parameters")
	ErrTvShowIDNotProvided  = errors.New("no TV show ID in query parameters")
	ErrTmdbIDNotProvided    = errors.New("no TMDB ID in query parameters")
	ErrReviewIDNotProvided  = errors.New("no review ID in query parameters")
	ErrSortOrderNotProvided = errors.New("no sort order in query parameters")
	ErrCommentIDNotProvided = errors.New("no comment ID in query parameters")
)

func movieIDFromRequestPath(r *http.Request) (screenjournal.MovieID, error) {
	return parse.MovieIDFromString(mux.Vars(r)["movieID"])
}

func movieIDFromQueryParams(r *http.Request) (screenjournal.MovieID, error) {
	raw := r.URL.Query().Get("movieId")
	if raw == "" {
		return screenjournal.MovieID(0), ErrMovieIDNotProvided
	}

	return parse.MovieIDFromString(raw)
}

func tvShowIDFromQueryParams(r *http.Request) (screenjournal.TvShowID, error) {
	raw := r.URL.Query().Get("tvShowId")
	if raw == "" {
		return screenjournal.TvShowID(0), ErrTvShowIDNotProvided
	}

	return parse.TvShowIDFromString(raw)
}

func tmdbIDFromQueryParams(r *http.Request) (screenjournal.TmdbID, error) {
	raw := r.URL.Query().Get("tmdbId")
	if raw == "" {
		return screenjournal.TmdbID(0), ErrTmdbIDNotProvided
	}

	return parse.TmdbIDFromString(raw)
}

func reviewIDFromRequestPath(r *http.Request) (screenjournal.ReviewID, error) {
	return parse.ReviewIDFromString(mux.Vars(r)["reviewID"])
}

func reviewIDFromQueryParams(r *http.Request) (screenjournal.ReviewID, error) {
	raw := r.URL.Query().Get("reviewId")
	if raw == "" {
		return screenjournal.ReviewID(0), ErrReviewIDNotProvided
	}

	return parse.ReviewIDFromString(raw)
}

func commentIDFromRequestPath(r *http.Request) (screenjournal.CommentID, error) {
	return parse.CommentID(mux.Vars(r)["commentID"])
}

func commentIDFromQueryParams(r *http.Request) (screenjournal.CommentID, error) {
	raw := r.URL.Query().Get("commentId")
	if raw == "" {
		return screenjournal.CommentID(0), ErrCommentIDNotProvided
	}

	return parse.CommentID(raw)
}

func usernameFromRequestPath(r *http.Request) (screenjournal.Username, error) {
	return parse.Username(mux.Vars(r)["username"])
}

func inviteCodeFromQueryParams(r *http.Request) (screenjournal.InviteCode, error) {
	raw := r.URL.Query().Get("invite")
	if raw == "" {
		return screenjournal.InviteCode(""), nil
	}
	return parse.InviteCode(raw)
}

func sortOrderFromQueryParams(r *http.Request) (screenjournal.SortOrder, error) {
	raw := r.URL.Query().Get("sort-by")
	if raw == "" {
		return screenjournal.SortOrder(""), ErrSortOrderNotProvided
	}

	if raw == string(screenjournal.ByRating) {
		return screenjournal.ByRating, nil
	} else if raw == string(screenjournal.ByWatchDate) {
		return screenjournal.ByWatchDate, nil
	}
	return screenjournal.SortOrder(""), errors.New("unrecognized sort order")
}
