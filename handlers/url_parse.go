package handlers

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

var ErrMoveIDNotProvided = errors.New("no movie ID in query parameters")
var ErrSortOrderNotProvided = errors.New("no sort order in query parameters")

func movieIDFromRequestPath(r *http.Request) (screenjournal.MovieID, error) {
	return parse.MovieIDFromString(mux.Vars(r)["movieID"])
}

func movieIDFromQueryParams(r *http.Request) (screenjournal.MovieID, error) {
	raw := r.URL.Query().Get("movieId")
	if raw == "" {
		return screenjournal.MovieID(0), ErrMoveIDNotProvided
	}

	return parse.MovieIDFromString(raw)
}

func reviewIDFromRequestPath(r *http.Request) (screenjournal.ReviewID, error) {
	return parse.ReviewIDFromString(mux.Vars(r)["reviewID"])
}

func commentIDFromRequestPath(r *http.Request) (screenjournal.CommentID, error) {
	return parse.CommentID(mux.Vars(r)["commentID"])
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
