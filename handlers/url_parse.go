package handlers

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
)

func reviewIDFromRequestPath(r *http.Request) (screenjournal.ReviewID, error) {
	return parse.ReviewIDFromString(mux.Vars(r)["reviewID"])
}

func usernameFromRequestPath(r *http.Request) (screenjournal.Username, error) {
	// TODO: Verify this is a real user that exists in the DB.
	return parse.Username(mux.Vars(r)["username"])
}
