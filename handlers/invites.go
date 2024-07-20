package handlers

import (
	"html/template"
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type invitesPostRequest struct {
	Invitee screenjournal.Invitee
}

func (s Server) invitesPost() http.HandlerFunc {
	t := template.Must(template.ParseFS(templatesFS, "templates/fragments/invite-row.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := parseInvitesPostRequest(r)
		if err != nil {
			log.Printf("failed to parse invites POST: %v", err)
			http.Error(w, "Invalid invite creation", http.StatusBadRequest)
			return
		}

		invitation := screenjournal.SignupInvitation{
			Invitee:    req.Invitee,
			InviteCode: screenjournal.NewInviteCode(),
		}
		if err := s.getDB(r).InsertSignupInvitation(invitation); err != nil {
			log.Printf("failed to add new signup invite %+v: %v", invitation, err)
			http.Error(w, "Failed to store new signup invite", http.StatusInternalServerError)
			return
		}

		if err := t.Execute(w, struct {
			Invitee    screenjournal.Invitee
			InviteCode screenjournal.InviteCode
		}{
			Invitee:    invitation.Invitee,
			InviteCode: invitation.InviteCode,
		}); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("failed to render invite row template: %v", err)
			return
		}
	}

}

func parseInvitesPostRequest(r *http.Request) (invitesPostRequest, error) {
	if err := r.ParseForm(); err != nil {
		log.Printf("failed to decode comment POST request: %v", err)
		return invitesPostRequest{}, err
	}

	invitee, err := parse.Invitee(r.PostFormValue("invitee"))
	if err != nil {
		return invitesPostRequest{}, err
	}

	return invitesPostRequest{
		Invitee: invitee,
	}, nil
}
