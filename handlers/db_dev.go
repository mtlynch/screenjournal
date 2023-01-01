//go:build dev

package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/store/sqlite"
)

// addDevRoutes adds debug routes that we only use during development or e2e
// tests.
func (s *Server) addDevRoutes() {
	s.router.HandleFunc("/api/debug/db/populate-dummy-data", s.populateDummyData()).Methods(http.MethodGet)
	s.router.HandleFunc("/api/debug/db/wipe", s.wipeDB()).Methods(http.MethodGet)
}

func (s Server) populateDummyData() http.HandlerFunc {
	users := []screenjournal.User{
		{
			Username:     screenjournal.Username("dummyadmin"),
			PasswordHash: screenjournal.NewPasswordHash(screenjournal.Password("dummypass")),
			IsAdmin:      true,
			Email:        screenjournal.Email("dummyadmin@example.com"),
		},
		{
			Username:     screenjournal.Username("userA"),
			PasswordHash: screenjournal.NewPasswordHash(screenjournal.Password("password123")),
			IsAdmin:      false,
			Email:        screenjournal.Email("userA@example.com"),
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		for _, u := range users {
			if err := s.store.InsertUser(u); err != nil {
				http.Error(w, fmt.Sprintf("Failed to insert user: %v", err), http.StatusInternalServerError)
				return
			}
		}
	}
}

// wipeDB wipes the database back to a freshly initialized state.
func (s Server) wipeDB() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sqlStore, ok := s.store.(*sqlite.DB)
		if !ok {
			log.Fatalf("store is not SQLite, can't wipe database")
		}
		sqlStore.Clear()
	}
}
