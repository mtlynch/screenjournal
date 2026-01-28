package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
)

type reviewDraftResponse struct {
	ReviewID screenjournal.ReviewID `json:"reviewId"`
}

func (s Server) reviewsDraftsPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := parseReviewPostRequest(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}
		saveDraftIntent := r.PostFormValue("save-draft") == "true" ||
			r.PostFormValue("save-draft") == "1"

		loggedInUsername := mustGetUsernameFromContext(r.Context())
		review := screenjournal.Review{
			Owner:        loggedInUsername,
			TvShowSeason: req.TvShowSeason,
			Rating:       req.Rating,
			Watched:      req.WatchDate,
			Blurb:        req.Blurb,
			IsDraft:      true,
			Comments:     []screenjournal.ReviewComment{},
		}

		if req.MediaType == screenjournal.MediaTypeMovie {
			review.Movie, err = s.moviefromTmdbID(s.getDB(r), req.TmdbID)
			if err == store.ErrMovieNotFound {
				http.Error(w, fmt.Sprintf("Could not find movie with TMDB ID: %v", req.TmdbID), http.StatusNotFound)
				return
			} else if err != nil {
				log.Printf("failed to get local media ID for movie with TMDB ID %v: %v", req.TmdbID, err)
				http.Error(w, fmt.Sprintf("Failed to look up movie with TMDB ID: %v: %v", req.TmdbID, err), http.StatusInternalServerError)
				return
			}
		} else if req.MediaType == screenjournal.MediaTypeTvShow {
			review.TvShow, err = s.tvShowfromTmdbID(s.getDB(r), req.TmdbID)
			if err == store.ErrTvShowNotFound {
				http.Error(w, fmt.Sprintf("Could not find tv show with TMDB ID: %v", req.TmdbID), http.StatusNotFound)
				return
			} else if err != nil {
				log.Printf("failed to get local media ID for TV show with TMDB ID %v: %v", req.TmdbID, err)
				http.Error(w, fmt.Sprintf("Failed to look up TV show with TMDB ID: %v: %v", req.TmdbID, err), http.StatusInternalServerError)
				return
			}
		}

		existingDraft, err := s.findExistingDraft(r, loggedInUsername, review)
		if err != nil {
			log.Printf("failed to find existing draft: %v", err)
			http.Error(w, fmt.Sprintf("Failed to save draft: %v", err), http.StatusInternalServerError)
			return
		}
		statusCode := http.StatusCreated
		if existingDraft != nil {
			review.ID = existingDraft.ID
			if err := s.getDB(r).UpdateReview(review); err != nil {
				log.Printf("failed to update draft: %v", err)
				http.Error(w, fmt.Sprintf("Failed to save draft: %v", err), http.StatusInternalServerError)
				return
			}
			statusCode = http.StatusOK
		} else {
			review.ID, err = s.getDB(r).InsertReview(review)
			if err != nil {
				log.Printf("failed to save draft: %v", err)
				http.Error(w, fmt.Sprintf("Failed to save draft: %v", err), http.StatusInternalServerError)
				return
			}
		}

		if saveDraftIntent {
			http.Redirect(w, r, "/reviews/drafts", http.StatusSeeOther)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if err := json.NewEncoder(w).Encode(reviewDraftResponse{ReviewID: review.ID}); err != nil {
			log.Printf("failed to encode draft response: %v", err)
		}
	}
}

func (s Server) findExistingDraft(r *http.Request, owner screenjournal.Username, draft screenjournal.Review) (*screenjournal.Review, error) {
	queryOptions := []store.ReadReviewsOption{
		store.FilterReviewsByUsername(owner),
		store.FilterReviewsByDraftStatus(true),
	}
	if draft.MediaType() == screenjournal.MediaTypeMovie {
		queryOptions = append(queryOptions, store.FilterReviewsByMovieID(draft.Movie.ID))
	} else {
		queryOptions = append(queryOptions, store.FilterReviewsByTvShowID(draft.TvShow.ID))
		queryOptions = append(queryOptions, store.FilterReviewsByTvShowSeason(draft.TvShowSeason))
	}

	drafts, err := s.getDB(r).ReadReviews(queryOptions...)
	if err != nil {
		return nil, err
	}
	if len(drafts) == 0 {
		return nil, nil
	}

	latest := drafts[0]
	for _, candidate := range drafts[1:] {
		if candidate.Modified.After(latest.Modified) {
			latest = candidate
		}
	}
	return &latest, nil
}

func (s Server) reviewsDraftsPut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := reviewIDFromRequestPath(r)
		if err != nil {
			http.Error(w, "Invalid review ID", http.StatusBadRequest)
			return
		}

		review, err := s.getDB(r).ReadReview(id)
		if err == store.ErrReviewNotFound {
			http.Error(w, "Draft not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read draft: %v", err), http.StatusInternalServerError)
			return
		}

		loggedInUsername := mustGetUsernameFromContext(r.Context())
		if !review.Owner.Equal(loggedInUsername) {
			http.Error(w, "You can't edit another user's draft", http.StatusForbidden)
			return
		}

		if !review.IsDraft {
			http.Error(w, "Review already published", http.StatusBadRequest)
			return
		}

		parsedRequest, err := parseReviewPutRequest(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		review.Rating = parsedRequest.Rating
		review.Blurb = parsedRequest.Blurb
		review.Watched = parsedRequest.Watched

		if err := s.getDB(r).UpdateReview(review); err != nil {
			log.Printf("failed to update draft: %v", err)
			http.Error(w, fmt.Sprintf("Failed to update draft: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
