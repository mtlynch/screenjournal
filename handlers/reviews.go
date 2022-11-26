package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/metadata/tmdb"
	"github.com/mtlynch/screenjournal/v2/store"
)

type reviewPostRequest struct {
	Review screenjournal.Review
	TmdbID screenjournal.TmdbID
}

func (s Server) reviewsPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner, ok := userFromContext(r.Context())
		if !ok {
			http.Error(w, "Must be logged in to add a review", http.StatusForbidden)
			return
		}

		req, err := newReviewFromRequest(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		req.Review.Owner = owner.Username

		req.Review.Movie, err = s.moviefromTmdbID(req.TmdbID)
		if err == store.ErrMovieNotFound {
			http.Error(w, fmt.Sprintf("Could not find movie with TMDB ID: %v", req.TmdbID), http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("failed to get local media ID for TMDB ID %v: %v", req.TmdbID, err)
			http.Error(w, fmt.Sprintf("Failed to look up TMDB ID: %v: %v", req.TmdbID, err), http.StatusInternalServerError)
			return
		}

		if err := s.store.InsertReview(req.Review); err != nil {
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

func newReviewFromRequest(r *http.Request) (reviewPostRequest, error) {
	var payload struct {
		TmdbID  int    `json:"tmdbId"`
		Rating  int    `json:"rating"`
		Watched string `json:"watched"`
		Blurb   string `json:"blurb"`
	}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Printf("failed to decode JSON request: %v", err)
		return reviewPostRequest{}, err
	}

	tmdbID, err := tmdb.ParseTmdbID(payload.TmdbID)
	if err != nil {
		return reviewPostRequest{}, err
	}
	rating, err := parse.Rating(payload.Rating)
	if err != nil {
		return reviewPostRequest{}, err
	}
	watchDate, err := parse.WatchDate(payload.Watched)
	if err != nil {
		return reviewPostRequest{}, err
	}
	blurb, err := parse.Blurb(payload.Blurb)
	if err != nil {
		return reviewPostRequest{}, err
	}

	return reviewPostRequest{
		Review: screenjournal.Review{
			Rating:  rating,
			Blurb:   blurb,
			Watched: watchDate,
		},
		TmdbID: tmdbID,
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

func (s Server) moviefromTmdbID(tmdbID screenjournal.TmdbID) (screenjournal.Movie, error) {
	movie, err := s.store.ReadMovieByTmdbID(tmdbID)
	if err != nil && err != store.ErrMovieNotFound {
		return screenjournal.Movie{}, err
	} else if err == nil {
		return movie, nil
	}

	mi, err := s.metadataFinder.GetMovieInfo(tmdbID)
	if err != nil {
		return screenjournal.Movie{}, err
	}

	movie = screenjournal.Movie{
		TmdbID: mi.TmdbID,
		ImdbID: mi.ImdbID,
		Title:  mi.Title,
	}

	movie.ID, err = s.store.InsertMovie(movie)
	if err != nil {
		return screenjournal.Movie{}, err
	}

	return movie, nil
}
