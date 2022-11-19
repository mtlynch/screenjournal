package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/store"
)

func (s Server) reviewsPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rev, err := newReviewFromRequest(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		m, err := s.metadataFinder.GetMovieInfo(rev.TmdbID)
		if err != nil {
			log.Printf("failed to get metadata for ID (%v): %v", rev.TmdbID, err)
			http.Error(w, fmt.Sprintf("Failed to retrieve metadata information for ID: %v", rev.TmdbID), http.StatusFailedDependency)
			return
		}

		rev.Title, err = parse.MediaTitle(m.Title)
		if err != nil {
			log.Printf("failed to parse media title (%s) from metadata: %v", m.Title, err)
			http.Error(w, fmt.Sprintf("Failed to retrieve metadata information for ID: %v", rev.TmdbID), http.StatusFailedDependency)
			return
		}

		if err := s.store.InsertReview(rev); err != nil {
			log.Printf("failed to save review: %v", err)
			http.Error(w, fmt.Sprintf("Failed to save review: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) reviewsPut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := reviewIDFromRequestPath(r)
		if err != nil {
			http.Error(w, "Invalid review ID", http.StatusBadRequest)
			return
		}

		review, err := s.store.ReadReview(id)
		if err == store.ErrReviewNotFound {
			http.Error(w, "Review not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read review: %v", err), http.StatusInternalServerError)
			return
		}

		if err := updateReviewFromRequest(r, &review); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		if err := s.store.UpdateReview(review); err != nil {
			log.Printf("failed to update review: %v", err)
			http.Error(w, fmt.Sprintf("Failed to update review: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func newReviewFromRequest(r *http.Request) (screenjournal.Review, error) {
	var payload struct {
		TmdbID  int    `json:"tmdbId"`
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

	tmdbID, err := parse.TmdbID(payload.TmdbID)
	if err != nil {
		return screenjournal.Review{}, err
	}
	rating, err := parse.Rating(payload.Rating)
	if err != nil {
		return screenjournal.Review{}, err
	}
	watchDate, err := parse.WatchDate(payload.Watched)
	if err != nil {
		return screenjournal.Review{}, err
	}
	blurb, err := parse.Blurb(payload.Blurb)
	if err != nil {
		return screenjournal.Review{}, err
	}

	return screenjournal.Review{
		Owner:   owner,
		TmdbID:  tmdbID,
		Rating:  rating,
		Blurb:   blurb,
		Watched: watchDate,
	}, nil
}

func updateReviewFromRequest(r *http.Request, review *screenjournal.Review) error {
	var payload struct {
		Rating  int    `json:"rating"`
		Watched string `json:"watched"`
		Blurb   string `json:"blurb"`
	}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Printf("failed to decode JSON request: %v", err)
		return err
	}

	rating, err := parse.Rating(payload.Rating)
	if err != nil {
		return err
	}
	review.Rating = rating

	watched, err := parse.WatchDate(payload.Watched)
	if err != nil {
		return err
	}
	review.Watched = watched

	blurb, err := parse.Blurb(payload.Blurb)
	if err != nil {
		return err
	}
	review.Blurb = blurb

	return nil
}
