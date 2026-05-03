package handlers

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/passwordreset"
)

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
		user, err := s.store.ReadUser(username)
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
		serverSupportsPasswordResets := s.passwordResetterForRequest(r) != nil
		if err := t.Execute(w, struct {
			commonProps
			Submitted                    bool
			ServerSupportsPasswordResets bool
		}{
			commonProps:                  makeCommonProps(r.Context()),
			ServerSupportsPasswordResets: serverSupportsPasswordResets,
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

		// Always render the same success page regardless of account lookup
		// outcome.
		renderSuccess := func() {
			if err := pageTemplate.Execute(w, struct {
				commonProps
				Submitted                    bool
				ServerSupportsPasswordResets bool
			}{
				commonProps:                  makeCommonProps(r.Context()),
				Submitted:                    true,
				ServerSupportsPasswordResets: true,
			}); err != nil {
				log.Printf("failed to render reset password success page: %v", err)
			}
		}

		emailAddr, err := parse.Email(r.FormValue("email"))
		if err != nil {
			http.Error(w, "Invalid email address", http.StatusBadRequest)
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
		return typedResetter.WithStore(s.store)
	}

	return resetter
}
