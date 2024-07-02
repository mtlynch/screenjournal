package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type accountChangePasswordPutRequest struct {
	OldPassword     screenjournal.Password
	NewPasswordHash screenjournal.PasswordHash
}

func (s Server) accountChangePasswordPut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parsed, err := parseAccountChangePasswordPutRequest(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to change password: %v", err), http.StatusBadRequest)
			return
		}

		username := mustGetUsernameFromContext(r.Context())
		if err := s.getAuthenticator(r).Authenticate(username, parsed.OldPassword); err != nil {
			log.Printf("password change failed for user %s: %v", username, err)
			http.Error(w, "Failed to change password: current password is incorrect", http.StatusUnauthorized)
			return
		}

		user, err := s.getDB(r).ReadUser(username)
		if err != nil {
			http.Error(w, "Failed to change password: couldn't read user information", http.StatusInternalServerError)
			return
		}

		if err := s.getDB(r).UpdateUserPassword(user.Username, parsed.NewPasswordHash); err != nil {
			http.Error(w, "Failed to change password: couldn't save new password", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprint(w, "Password updated"); err != nil {
			log.Printf("failed to write response: %v", err)
		}
	}
}

func parseAccountChangePasswordPutRequest(r *http.Request) (accountChangePasswordPutRequest, error) {
	if err := r.ParseForm(); err != nil {
		log.Printf("failed to decode password PUT request: %v", err)
		return accountChangePasswordPutRequest{}, err
	}

	oldPassword, err := parse.Password(r.PostFormValue("old-password"))
	if err != nil {
		return accountChangePasswordPutRequest{}, err
	}

	newPassword, err := parse.Password(r.PostFormValue("password"))
	if err != nil {
		return accountChangePasswordPutRequest{}, err
	}

	newPasswordConfirm, err := parse.Password(r.PostFormValue("password-confirm"))
	if err != nil {
		return accountChangePasswordPutRequest{}, err
	}

	if oldPassword.Equal(newPassword) {
		return accountChangePasswordPutRequest{}, errors.New("old password is the same as the new password")
	}

	if !newPassword.Equal(newPasswordConfirm) {
		return accountChangePasswordPutRequest{}, errors.New("password confirmation does not match")
	}

	newPasswordHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return accountChangePasswordPutRequest{}, err
	}

	return accountChangePasswordPutRequest{
		OldPassword:     oldPassword,
		NewPasswordHash: newPasswordHash,
	}, nil
}

type accountNotificationsPutRequest struct {
	NewReviews  bool
	AllComments bool
}

func (s Server) accountNotificationsPut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := parseAccountNotificationsPutRequest(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		username := mustGetUsernameFromContext(r.Context())
		if err = s.getDB(r).UpdateNotificationPreferences(username, screenjournal.NotificationPreferences{
			NewReviews:     req.NewReviews,
			AllNewComments: req.AllComments,
		}); err != nil {
			log.Printf("failed to save notification preferences: %v", err)
			http.Error(w, fmt.Sprintf("Failed to save notification preferences: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprint(w, "Changes saved"); err != nil {
			log.Printf("failed to write response: %v", err)
		}
	}
}

func parseAccountNotificationsPutRequest(r *http.Request) (accountNotificationsPutRequest, error) {
	if err := r.ParseForm(); err != nil {
		log.Printf("failed to decode review POST request: %v", err)
		return accountNotificationsPutRequest{}, err
	}

	return accountNotificationsPutRequest{
		NewReviews:  parse.CheckboxToBool(r.PostFormValue("new-reviews")),
		AllComments: parse.CheckboxToBool(r.PostFormValue("all-comments")),
	}, nil
}
