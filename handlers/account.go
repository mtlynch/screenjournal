package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

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
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		username := mustGetUsernameFromContext(r.Context())
		if err := s.getAuthenticator(r).Authenticate(username, req.OldPassword); err != nil {
			log.Printf("password change failed for user %s: %v", username, err)
			http.Error(w, "Old password is incorrect", http.StatusUnauthorized)
			return
		}

		user, err := s.getDB(r).ReadUser(username)
		if err != nil {
			http.Error(w, "Failed to read user information", http.StatusInternalServerError)
			return
		}

		user.PasswordHash = req.NewPasswordHash

		if err := s.getDB(r).UpdateUser(user); err != nil {
			http.Error(w, "Failed to save new password", http.StatusInternalServerError)
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

	newPasswordHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return accountChangePasswordPostRequest{}, err
	}

	return accountChangePasswordPostRequest{
		OldPassword:     oldPassword,
		NewPasswordHash: newPasswordHash,
	}, nil
}

type accountNotificationsPostRequest struct {
	NewReviews  bool
	AllComments bool
}

func (s Server) accountNotificationsPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := notificationPreferencesFromRequest(r)
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

func notificationPreferencesFromRequest(r *http.Request) (accountNotificationsPostRequest, error) {
	var payload struct {
		NewReviews  bool `json:"isSubscribedToNewReviews"`
		AllComments bool `json:"isSubscribedToAllComments"`
	}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Printf("failed to decode JSON request: %v", err)
		return accountNotificationsPostRequest{}, err
	}

	return accountNotificationsPostRequest{
		NewReviews:  payload.NewReviews,
		AllComments: payload.AllComments,
	}, nil
}
