package handlers

import "net/http"

func (s Server) invitesPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not implemented", http.StatusNotImplemented)
	}
}
