//go:build dev

package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

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
		{
			Username:     screenjournal.Username("userB"),
			PasswordHash: screenjournal.NewPasswordHash(screenjournal.Password("password456")),
			IsAdmin:      false,
			Email:        screenjournal.Email("userB@example.com"),
		},
	}
	movies := []screenjournal.Movie{
		{
			ID:    screenjournal.MovieID(1),
			Title: screenjournal.MediaTitle("The Waterboy"),
		},
	}
	reviews := []screenjournal.Review{
		{
			ID:     screenjournal.ReviewID(1),
			Owner:  screenjournal.Username("userA"),
			Rating: screenjournal.Rating(5),
			Movie: screenjournal.Movie{
				ID: screenjournal.MovieID(1),
			},
			Watched: screenjournal.WatchDate(time.Date(2020, time.October, 5, 20, 18, 55, 0, time.Local)),
			Blurb:   screenjournal.Blurb("I love water!"),
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		for _, u := range users {
			if err := s.store.InsertUser(u); err != nil {
				http.Error(w, fmt.Sprintf("Failed to insert user: %v", err), http.StatusInternalServerError)
				return
			}
		}
		for _, movie := range movies {
			if _, err := s.store.InsertMovie(movie); err != nil {
				http.Error(w, fmt.Sprintf("Failed to insert movie: %v", err), http.StatusInternalServerError)
				return
			}
		}
		for _, review := range reviews {
			if _, err := s.store.InsertReview(review); err != nil {
				http.Error(w, fmt.Sprintf("Failed to insert review: %v", err), http.StatusInternalServerError)
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
