package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2"
)

type accountNotificationsPostRequest struct {
	NewReviews bool
}

func (s Server) accountNotificationsPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := notificationPreferencesFromRequest(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		prefs := screenjournal.NotificationPreferences{
			NewReviews: req.NewReviews,
		}

		user := mustGetUserFromContext(r.Context())
		err = s.getDB(r).UpdateNotificationPreferences(user.Username, prefs)
		if err != nil {
			log.Printf("failed to save notification preferences: %v", err)
			http.Error(w, fmt.Sprintf("Failed to save notification preferences: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func notificationPreferencesFromRequest(r *http.Request) (accountNotificationsPostRequest, error) {
	var payload struct {
		NewReviews bool `json:"isSubscribedToNewReviews"`
	}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Printf("failed to decode JSON request: %v", err)
		return accountNotificationsPostRequest{}, err
	}

	return accountNotificationsPostRequest{
		NewReviews: payload.NewReviews,
	}, nil
}
