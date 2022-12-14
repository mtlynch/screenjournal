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
	return parse.Username(mux.Vars(r)["username"])
}

func inviteCodeFromQueryParams(r *http.Request) (screenjournal.InviteCode, error) {
	raw := r.URL.Query().Get("invite")
	if raw == "" {
		return screenjournal.InviteCode(""), nil
	}
	return parse.InviteCode(raw)
}
