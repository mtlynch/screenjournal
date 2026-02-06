package handlers

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func (s Server) invitesGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").ParseFS(
			templatesFS,
			append(
				baseTemplates,
				"templates/fragments/invite-row.html",
				"templates/pages/invites.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		invites, err := s.getDB(r).ReadSignupInvitations()
		if err != nil {
			log.Printf("failed to read signup invitations: %v", err)
			http.Error(w, "Failed to read signup invitations", http.StatusInternalServerError)
			return
		}
		if err := t.Execute(w, struct {
			commonProps
			Invites []screenjournal.SignupInvitation
		}{
			commonProps: makeCommonProps(r.Context()),
			Invites:     invites,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) passwordResetAdminGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			Funcs(template.FuncMap{
				"formatTime": func(t time.Time) string {
					return t.Format("Jan 2, 2006 3:04 PM")
				},
			}).
			ParseFS(
				templatesFS,
				append(baseTemplates, "templates/pages/admin-reset-password.html", "templates/fragments/password-reset-row.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		allUsers, err := s.getDB(r).ReadUsersPublicMeta()
		if err != nil {
			log.Printf("failed to read users: %v", err)
			http.Error(w, "Failed to load users", http.StatusInternalServerError)
			return
		}

		// Filter the current user from the list.
		currentUsername := mustGetUsernameFromContext(r.Context())
		var users []screenjournal.UserPublicMeta
		for _, user := range allUsers {
			if !user.Username.Equal(currentUsername) {
				users = append(users, user)
			}
		}

		// Clean up expired tokens before displaying.
		if err := s.getDB(r).DeleteExpiredPasswordResetEntries(); err != nil {
			log.Printf("failed to clean up expired password reset tokens: %v", err)
		}

		passwordResetEntries, err := s.getDB(r).ReadPasswordResetEntries()
		if err != nil {
			log.Printf("failed to read password reset requests: %v", err)
			http.Error(w, "Failed to load password reset requests", http.StatusInternalServerError)
			return
		}

		if err := t.Execute(w, struct {
			commonProps
			passwordResetAdminGetRequest
		}{
			commonProps: makeCommonProps(r.Context()),
			passwordResetAdminGetRequest: passwordResetAdminGetRequest{
				Users:                users,
				PasswordResetEntries: passwordResetEntries,
			},
		}); err != nil {
			log.Printf("failed to render admin reset password template: %v", err)
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
		}
	}
}

func (s Server) usersGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			Funcs(reviewPageFns).
			ParseFS(
				templatesFS,
				append(baseTemplates, "templates/pages/users.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		users, err := s.getDB(r).ReadUsersPublicMeta()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := t.Execute(w, struct {
			commonProps
			Users []screenjournal.UserPublicMeta
		}{
			commonProps: makeCommonProps(r.Context()),
			Users:       users,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
