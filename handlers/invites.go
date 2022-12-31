package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
)

type invitesPostRequest struct {
	Invitee screenjournal.Invitee
}

func (s Server) invitesPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := parseInvitesPostRequest(r)
		if err != nil {
			log.Printf("failed to parse invites POST: %v", err)
			http.Error(w, "Invalid invite creation", http.StatusBadRequest)
			return
		}

		if err := s.store.InsertSignupInvitation(screenjournal.SignupInvitation{
			Invitee:    req.Invitee,
			InviteCode: screenjournal.NewInviteCode(),
		}); err != nil {
			log.Printf("failed to add new signup invite code: %v", err)
			http.Error(w, "Failed to store new signup invite", http.StatusInternalServerError)
			return
		}
	}
}

func parseInvitesPostRequest(r *http.Request) (invitesPostRequest, error) {
	var payload struct {
		Invitee string `json:"invitee"`
	}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Printf("failed to decode JSON request: %v", err)
		return invitesPostRequest{}, err
	}

	invitee, err := parse.Invitee(payload.Invitee)
	if err != nil {
		return invitesPostRequest{}, err
	}

	return invitesPostRequest{
		Invitee: invitee,
	}, nil
}
