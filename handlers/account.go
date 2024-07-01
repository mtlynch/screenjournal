package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type accountChangePasswordPostRequest struct {
	OldPassword     screenjournal.Password
	NewPasswordHash screenjournal.PasswordHash
}

func (s Server) accountChangePasswordPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := changePasswordFromRequest(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to change password: %v", err), http.StatusBadRequest)
			return
		}

		username := mustGetUsernameFromContext(r.Context())
		if err := s.getAuthenticator(r).Authenticate(username, req.OldPassword); err != nil {
			log.Printf("password change failed for user %s: %v", username, err)
			http.Error(w, "Failed to change password: current password is incorrect", http.StatusUnauthorized)
			return
		}

		user, err := s.getDB(r).ReadUser(username)
		if err != nil {
			http.Error(w, "Failed to change password: couldn't read user information", http.StatusInternalServerError)
			return
		}

		if err := s.getDB(r).UpdateUserPassword(user.Username, req.NewPasswordHash); err != nil {
			http.Error(w, "Failed to change password: couldn't save new password", http.StatusInternalServerError)
			return
		}
	}
}

func changePasswordFromRequest(r *http.Request) (accountChangePasswordPostRequest, error) {
	var payload struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Printf("failed to decode JSON request: %v", err)
		return accountChangePasswordPostRequest{}, err
	}

	oldPassword, err := parse.Password(payload.OldPassword)
	if err != nil {
		return accountChangePasswordPostRequest{}, err
	}

	newPassword, err := parse.Password(payload.NewPassword)
	if err != nil {
		return accountChangePasswordPostRequest{}, err
	}

	if oldPassword.Equal(newPassword) {
		return accountChangePasswordPostRequest{}, errors.New("old password is the same as the new password")
	}

	newPasswordHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return accountChangePasswordPostRequest{}, err
	}

	return accountChangePasswordPostRequest{
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
		time.Sleep(3 * time.Second) // DEBUG
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
