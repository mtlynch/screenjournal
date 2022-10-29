package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
)

func (s Server) reviewsPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rev, err := reviewFromRequest(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		// TODO: Prevent user from reviewing the same movie twice

		now := time.Now()
		rev.Created = now
		rev.Modified = now

		if err := s.store.InsertReview(rev); err != nil {
			log.Printf("failed to save review: %v", err)
			http.Error(w, fmt.Sprintf("Failed to save review: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func reviewFromRequest(r *http.Request) (screenjournal.Review, error) {
	var payload struct {
		Title   string `json:"title"`
		Rating  int    `json:"rating"`
		Watched string `json:"watched"`
		Blurb   string `json:"blurb"`
	}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Printf("failed to decode JSON request: %v", err)
		return screenjournal.Review{}, err
	}

	// TODO: Support reviews by other users
	owner := screenjournal.Username("mike")

	// TODO: Actually parse these
	title := screenjournal.MediaTitle(payload.Title)
	rating := screenjournal.Rating(payload.Rating)
	watchDate, err := parse.WatchDate(payload.Watched)
	if err != nil {
		return screenjournal.Review{}, err
	}
	blurb := screenjournal.Blurb(payload.Blurb)

	return screenjournal.Review{
		Owner:   owner,
		Title:   title,
		Rating:  rating,
		Blurb:   blurb,
		Watched: watchDate,
	}, nil
}
