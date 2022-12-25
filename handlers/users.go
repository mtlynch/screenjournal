package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
)

type userPutRequest struct {
	Email        screenjournal.Email
	Username     screenjournal.Username
	PasswordHash screenjournal.PasswordHash
}

func (s Server) usersPut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := newUserFromRequest(r)
		if err != nil {
			log.Printf("invalid new user form: %v", err)
			http.Error(w, fmt.Sprintf("Signup failed: %v", err), http.StatusBadRequest)
			return
		}

		c, err := s.store.CountUsers()
		if err != nil {
			log.Printf("failed to query user count: %v", err)
			http.Error(w, "Failed to query user count", http.StatusInternalServerError)
			return
		}

		user := screenjournal.User{
			IsAdmin:      c == 0, // First user is automatically admin
			Email:        req.Email,
			Username:     req.Username,
			PasswordHash: req.PasswordHash,
		}

		if err := s.store.InsertUser(user); err != nil {
			log.Printf("failed to add new user: %v", err)
			http.Error(w, "Failed to add new user", http.StatusInternalServerError)
			return
		}

		if err := s.sessionManager.CreateSession(w, r, user); err != nil {
			log.Printf("failed to create session for new user %+v: %v", user, err)
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}
	}
}

func newUserFromRequest(r *http.Request) (userPutRequest, error) {
	username, err := usernameFromRequestPath(r)
	if err != nil {
		return userPutRequest{}, err
	}

	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("failed to decode JSON request: %v", err)
		return userPutRequest{}, err
	}

	email, err := parse.Email(payload.Email)
	if err != nil {
		return userPutRequest{}, err
	}

	plaintextPassword, err := parse.Password(payload.Password)
	if err != nil {
		return userPutRequest{}, err
	}

	return userPutRequest{
		Email:        email,
		Username:     username,
		PasswordHash: screenjournal.NewPasswordHash(plaintextPassword),
	}, nil
}