package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/store"
)

type userPutRequest struct {
	Email        screenjournal.Email
	Username     screenjournal.Username
	PasswordHash screenjournal.PasswordHash
	InviteCode   screenjournal.InviteCode
}

func (s Server) usersPut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := newUserFromRequest(r)
		if err != nil {
			log.Printf("invalid new user form: %v", err)
			http.Error(w, fmt.Sprintf("Signup failed: %v", err), http.StatusBadRequest)
			return
		}

		c, err := s.getDB(r).CountUsers()
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

		if c >= 1 {
			if _, err := s.getDB(r).ReadSignupInvitation(req.InviteCode); err != nil {
				log.Printf("invalid invite code: %v", err)
				http.Error(w, "Invalid invite code", http.StatusForbidden)
				return
			}
		}

		if err := s.getDB(r).InsertUser(user); err != nil {
			if err == store.ErrEmailAssociatedWithAnotherAccount {
				http.Error(w, "Failed to add new user", http.StatusConflict)
			} else if err == store.ErrUsernameNotAvailable {
				http.Error(w, "Username is not avilable", http.StatusConflict)
			}
			log.Printf("failed to add new user: %v", err)
			http.Error(w, "Failed to add new user", http.StatusInternalServerError)
			return
		}

		if err := s.sessionManager.CreateSession(w, r, user); err != nil {
			log.Printf("failed to create session for new user %+v: %v", user, err)
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}

		if !req.InviteCode.Empty() {
			if err := s.getDB(r).DeleteSignupInvitation(req.InviteCode); err != nil {
				log.Printf("failed to delete used signup invitation code: %v", err)
			}
		}

	}
}

func newUserFromRequest(r *http.Request) (userPutRequest, error) {
	username, err := usernameFromRequestPath(r)
	if err != nil {
		return userPutRequest{}, err
	}

	var payload struct {
		Email      string `json:"email"`
		Password   string `json:"password"`
		InviteCode string `json:"inviteCode"`
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

	var inviteCode screenjournal.InviteCode
	if payload.InviteCode != "" {
		if inviteCode, err = parse.InviteCode(payload.InviteCode); err != nil {
			return userPutRequest{}, err
		}
	}

	hash, err := auth.NewPasswordHash(plaintextPassword.String())
	if err != nil {
		return userPutRequest{}, err
	}

	return userPutRequest{
		Email:        email,
		Username:     username,
		PasswordHash: screenjournal.PasswordHash(hash.Bytes()),
		InviteCode:   inviteCode,
	}, nil
}
