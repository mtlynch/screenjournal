package handlers

import (
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func (s Server) indexGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			ParseFS(
				templatesFS,
				append(baseTemplates, "templates/pages/index.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		// Redirect logged in users to the reviews index instead of the landing
		// page.
		if isAuthenticated(r.Context()) {
			http.Redirect(w, r, "/reviews", http.StatusTemporaryRedirect)
			return
		}

		if err := t.Execute(w, struct {
			commonProps
		}{
			commonProps: makeCommonProps(r.Context()),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) aboutGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			ParseFS(
				templatesFS,
				append(baseTemplates, "templates/pages/about.html")...))
	return func(w http.ResponseWriter, r *http.Request) {
		if err := t.Execute(w, struct {
			commonProps
		}{
			commonProps: makeCommonProps(r.Context()),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) logInGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			ParseFS(
				templatesFS,
				append(baseTemplates, "templates/pages/login.html")...))
	return func(w http.ResponseWriter, r *http.Request) {
		if err := t.Execute(w, struct {
			commonProps
		}{
			commonProps: makeCommonProps(r.Context()),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) signUpGet() http.HandlerFunc {
	noInviteTemplate := template.Must(
		template.New("base.html").
			ParseFS(
				templatesFS,
				append(baseTemplates, "templates/pages/sign-up.html")...))
	byInviteTemplate := template.Must(
		template.New("base.html").
			ParseFS(
				templatesFS,
				append(baseTemplates, "templates/pages/sign-up-by-invitation.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		inviteCode, err := inviteCodeFromQueryParams(r)
		if err != nil {
			log.Printf("invalid invite code: %v", err)
			http.Error(w, "Invalid invite code", http.StatusBadRequest)
			return
		}

		var invite screenjournal.SignupInvitation
		if !inviteCode.Empty() {
			invite, err = s.getDB(r).ReadSignupInvitation(inviteCode)
			if err != nil {
				log.Printf("invalid invite code: %v", err)
				http.Error(w, "Invalid invite code", http.StatusUnauthorized)
				return
			}
		}

		uc, err := s.getDB(r).CountUsers()
		if err != nil {
			log.Printf("failed to count users: %v", err)
			http.Error(w, "Failed to load signup template", http.StatusInternalServerError)
			return
		}

		var t *template.Template
		if uc > 0 && invite.Empty() {
			t = byInviteTemplate
		} else {
			t = noInviteTemplate
		}

		var suggestedUsername string
		if !invite.Empty() {
			nonSuggestedCharsPattern := regexp.MustCompile(`(?i)[^a-z0-9]`)
			firstPart := strings.SplitN(invite.Invitee.String(), " ", 2)[0]
			suggestedUsername = nonSuggestedCharsPattern.ReplaceAllString(strings.ToLower(firstPart), "")
		}

		if err := t.Execute(w, struct {
			commonProps
			Invitee           screenjournal.Invitee
			SuggestedUsername string
		}{
			commonProps:       makeCommonProps(r.Context()),
			Invitee:           invite.Invitee,
			SuggestedUsername: suggestedUsername,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
