package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2"
)

func (s Server) usersPut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := usernameFromRequestPath(r)
		if err != nil {
			log.Printf("invalid username: %v", err)
			http.Error(w, fmt.Sprintf("Invalid username: %v", err), http.StatusBadRequest)
			return
		}

		user := screenjournal.User{
			IsAdmin:  true,
			Username: username,
		}

		if err := s.sessionManager.CreateSession(w, r, user); err != nil {
			http.Error(w, fmt.Sprintf("Failed to create session: %v", err), http.StatusInternalServerError)
			return
		}
	}
}
