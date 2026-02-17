package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/passwordreset"
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

		// Check if user exists.
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
		token, err := parse.PasswordResetToken(vars["token"])
		if err != nil {
			http.Error(w, "Invalid password reset token", http.StatusBadRequest)
			return
		}

		if err := s.getDB(r).DeletePasswordResetEntry(token); err != nil {
			log.Printf(
				"failed to delete password reset token %s: %v",
				passwordResetTokenPrefix(token),
				err,
			)
			http.Error(w, "Failed to delete password reset token", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (s Server) accountPasswordResetPut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		passwordResetter := s.passwordResetterForRequest(r)
		if passwordResetter == nil {
			http.Error(w, "Password resets are not available on this server", http.StatusServiceUnavailable)
			return
		}

		username, err := parse.Username(r.URL.Query().Get("username"))
		if err != nil {
			http.Error(w, "Invalid username", http.StatusBadRequest)
			return
		}

		token, err := parse.PasswordResetToken(r.URL.Query().Get("token"))
		if err != nil {
			http.Error(w, "Invalid password reset token", http.StatusBadRequest)
			return
		}

		// Parse new password.
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

		if err := passwordResetter.Reset(username, token, newPasswordHash); err != nil {
			switch {
			case errors.Is(err, passwordreset.ErrTooManyResetAttempts):
				http.Error(w, err.Error(), http.StatusTooManyRequests)
			case errors.Is(err, passwordreset.ErrInvalidResetToken),
				errors.Is(err, passwordreset.ErrExpiredResetToken):
				http.Error(w, err.Error(), http.StatusBadRequest)
			default:
				log.Printf("failed to reset password for user %s: %v", username, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Read user data to get admin status for session creation.
		user, err := s.getDB(r).ReadUser(username)
		if err != nil {
			log.Printf("failed to read user data for session creation %s: %v", username, err)
			http.Error(w, "Failed to read user information", http.StatusInternalServerError)
			return
		}

		// Create session to automatically log in the user.
		if err := s.sessionManager.CreateSession(w, r.Context(), user.Username, user.IsAdmin); err != nil {
			log.Printf("failed to create session for user %s after password reset: %v", user.Username.String(), err)
			http.Error(w, "Password updated but failed to log in", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprint(w, "Password updated successfully"); err != nil {
			log.Printf("failed to write response: %v", err)
		}
	}
}

func (s Server) resetPasswordGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			ParseFS(
				templatesFS,
				append(baseTemplates, "templates/pages/reset-password.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		if err := t.Execute(w, struct {
			commonProps
			Submitted   bool
			Unavailable bool
		}{
			commonProps: makeCommonProps(r.Context()),
			Unavailable: s.passwordResetter == nil,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) resetPasswordPost() http.HandlerFunc {
	pageTemplate := template.Must(
		template.New("base.html").
			ParseFS(
				templatesFS,
				append(baseTemplates, "templates/pages/reset-password.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		passwordResetter := s.passwordResetterForRequest(r)
		if passwordResetter == nil {
			http.Error(w, "Password resets are not available on this server", http.StatusServiceUnavailable)
			return
		}

		// Always render the same success page regardless of outcome.
		renderSuccess := func() {
			if err := pageTemplate.Execute(w, struct {
				commonProps
				Submitted   bool
				Unavailable bool
			}{
				commonProps: makeCommonProps(r.Context()),
				Submitted:   true,
			}); err != nil {
				log.Printf("failed to render reset password success page: %v", err)
			}
		}

		emailAddr, err := parse.Email(r.FormValue("email"))
		if err != nil {
			renderSuccess()
			return
		}

		if err := passwordResetter.SendEmail(emailAddr); err != nil {
			log.Printf("failed to process password reset email: %v", err)
			http.Error(w, "Failed to process password reset request", http.StatusInternalServerError)
			return
		}

		renderSuccess()
	}
}

func (s Server) passwordResetterForRequest(r *http.Request) PasswordResetter {
	if s.passwordResetter == nil {
		return nil
	}

	resetter := s.passwordResetter
	if typedResetter, ok := resetter.(passwordreset.Resetter); ok {
		return typedResetter.WithStore(s.getDB(r))
	}

	return resetter
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

func passwordResetTokenPrefix(token screenjournal.PasswordResetToken) string {
	tokenRaw := token.String()
	const tokenPreviewLength = 6
	if len(tokenRaw) <= tokenPreviewLength {
		return tokenRaw
	}
	return tokenRaw[:tokenPreviewLength] + "..."
}
