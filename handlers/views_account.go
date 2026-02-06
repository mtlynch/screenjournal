package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
)

func (s Server) accountPasswordResetGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").ParseFS(
			templatesFS,
			append(baseTemplates, "templates/pages/account-change-password.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		// If a token is provided, validate it before rendering the page.
		token, err := parse.PasswordResetToken(r.URL.Query().Get("token"))
		if err != nil {
			http.Error(w, "Invalid password reset token", http.StatusUnauthorized)
			return
		}

		// Verify token exists and hasn't expired.
		passwordResetEntry, err := s.getDB(r).ReadPasswordResetEntry(token)
		if err != nil {
			http.Error(w, "Invalid or expired password reset token", http.StatusUnauthorized)
			return
		}

		if passwordResetEntry.IsExpired() {
			// Clean up expired token.
			if err := s.getDB(r).DeletePasswordResetEntry(token); err != nil {
				log.Printf("failed to delete expired password reset token %s: %v", token, err)
			}
			http.Error(w, "Password reset token has expired", http.StatusUnauthorized)
			return
		}

		if err := t.Execute(w, struct {
			commonProps
			Token         string
			FormTargetURL string
			CancelURL     string
		}{
			commonProps:   makeCommonProps(r.Context()),
			Token:         token.String(),
			FormTargetURL: fmt.Sprintf("/account/password-reset?token=%s", token),
			CancelURL:     "/login",
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) accountChangePasswordGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").ParseFS(
			templatesFS,
			append(baseTemplates, "templates/pages/account-change-password.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		if err := t.Execute(w, struct {
			commonProps
			Token         string
			FormTargetURL string
			CancelURL     string
		}{
			commonProps:   makeCommonProps(r.Context()),
			Token:         "",
			FormTargetURL: "/account/password",
			CancelURL:     "/account/security",
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) accountNotificationsGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").ParseFS(
			templatesFS,
			append(baseTemplates, "templates/pages/account-notifications.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		prefs, err := s.getDB(r).ReadNotificationPreferences(mustGetUsernameFromContext(r.Context()))
		if err != nil {
			log.Printf("failed to read notification preferences: %v", err)
			http.Error(w, fmt.Sprintf("failed to read notification preferences: %v", err), http.StatusInternalServerError)
			return
		}

		if err := t.Execute(w, struct {
			commonProps
			ReceivesReviewNotices     bool
			ReceivesAllCommentNotices bool
		}{
			commonProps:               makeCommonProps(r.Context()),
			ReceivesReviewNotices:     prefs.NewReviews,
			ReceivesAllCommentNotices: prefs.AllNewComments,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) accountSecurityGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").ParseFS(
			templatesFS,
			append(baseTemplates, "templates/pages/account-security.html")...))

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
