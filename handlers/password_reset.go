package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type passwordResetAdminGetRequest struct {
	Users                []screenjournal.UserPublicMeta
	PasswordResetEntries []screenjournal.PasswordResetEntry
}

type passwordResetAdminPostRequest struct {
	Username screenjournal.Username
}

func (s Server) passwordResetAdminPost() http.HandlerFunc {
	t := template.Must(
		template.New("password-reset-row.html").
			Funcs(template.FuncMap{
				"formatTime": func(t time.Time) string {
					return t.Format("Jan 2, 2006 3:04 PM")
				},
			}).
			ParseFS(templatesFS, "templates/fragments/password-reset-row.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := parsePasswordResetAdminPostRequest(r)
		if err != nil {
			log.Printf("failed to parse password reset admin POST: %v", err)
			http.Error(w, "Invalid password reset creation", http.StatusBadRequest)
			return
		}

		// Check if user exists
		_, err = s.getDB(r).ReadUser(req.Username)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "User not found", http.StatusBadRequest)
				return
			}
			log.Printf("failed to look up user %s: %v", req.Username, err)
			http.Error(w, "Failed to verify user", http.StatusInternalServerError)
			return
		}

		passwordResetEntry := screenjournal.PasswordResetEntry{
			Username:  req.Username,
			Token:     screenjournal.NewPasswordResetToken(),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
		}

		if err := s.getDB(r).InsertPasswordResetEntry(passwordResetEntry); err != nil {
			log.Printf("failed to insert password reset entry %+v: %v", passwordResetEntry, err)
			http.Error(w, "Failed to create password reset entry", http.StatusInternalServerError)
			return
		}

		if err := t.Execute(w, struct {
			Username  screenjournal.Username
			Token     screenjournal.PasswordResetToken
			ExpiresAt time.Time
		}{
			Username:  passwordResetEntry.Username,
			Token:     passwordResetEntry.Token,
			ExpiresAt: passwordResetEntry.ExpiresAt,
		}); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("failed to render password reset row template: %v", err)
			return
		}
	}
}

func (s Server) passwordResetAdminDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		token := screenjournal.PasswordResetToken(vars["token"])

		if err := s.getDB(r).DeletePasswordResetEntry(token); err != nil {
			log.Printf("failed to delete password reset token %s: %v", token, err)
			http.Error(w, "Failed to delete password reset token", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (s Server) accountPasswordResetPut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := screenjournal.PasswordResetToken(r.URL.Query().Get("token"))
		if token.Empty() {
			http.Error(w, "Missing password reset token", http.StatusBadRequest)
			return
		}

		// Verify token exists and hasn't expired
		passwordResetEntry, err := s.getDB(r).ReadPasswordResetEntry(token)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Invalid or expired password reset token", http.StatusBadRequest)
				return
			}
			log.Printf("failed to read password reset entry for token %s: %v", token, err)
			http.Error(w, "Failed to verify password reset token", http.StatusInternalServerError)
			return
		}

		if passwordResetEntry.IsExpired() {
			// Clean up expired token
			if err := s.getDB(r).DeletePasswordResetEntry(token); err != nil {
				log.Printf("failed to delete expired password reset token %s: %v", token, err)
			}
			http.Error(w, "Password reset token has expired", http.StatusBadRequest)
			return
		}

		// Parse new password
		newPassword, err := parse.Password(r.PostFormValue("password"))
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid password: %v", err), http.StatusBadRequest)
			return
		}

		newPasswordHash, err := auth.HashPassword(newPassword)
		if err != nil {
			log.Printf("failed to hash password: %v", err)
			http.Error(w, "Failed to process new password", http.StatusInternalServerError)
			return
		}

		// Update user password
		if err := s.getDB(r).UpdateUserPassword(passwordResetEntry.Username, newPasswordHash); err != nil {
			log.Printf("failed to update password for %v: %v", passwordResetEntry.Username, err)
			http.Error(w, "Failed to update password", http.StatusInternalServerError)
			return
		}

		// Delete the used token
		if err := s.getDB(r).DeletePasswordResetEntry(token); err != nil {
			log.Printf("failed to delete used password reset token %s: %v", token, err)
			// Don't fail the request since password was updated successfully
		}

		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprint(w, "Password updated successfully"); err != nil {
			log.Printf("failed to write response: %v", err)
		}
	}
}

func parsePasswordResetAdminPostRequest(r *http.Request) (passwordResetAdminPostRequest, error) {
	if err := r.ParseForm(); err != nil {
		return passwordResetAdminPostRequest{}, err
	}

	username, err := parse.Username(r.PostFormValue("username"))
	if err != nil {
		return passwordResetAdminPostRequest{}, err
	}

	return passwordResetAdminPostRequest{
		Username: username,
	}, nil
}
