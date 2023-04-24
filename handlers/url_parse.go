package handlers

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
)

var ErrMoveIDNotProvided = errors.New("no movie ID in query parameters")

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
