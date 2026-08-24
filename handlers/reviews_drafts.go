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
		req, err := parseReviewPostRequest(r, false)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}
		saveDraftIntent := formBool(r, "save-draft")

		loggedInUsername := mustGetUsernameFromContext(r.Context())
		review, ok := s.reviewFromPostRequest(w, req, loggedInUsername, true)
		if !ok {
			return
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
			if err := s.store.UpdateReview(review); err != nil {
				log.Printf("failed to update draft: %v", err)
				http.Error(w, fmt.Sprintf("Failed to save draft: %v", err), http.StatusInternalServerError)
				return
			}
			statusCode = http.StatusOK
		} else {
			review.ID, err = s.store.InsertReview(review)
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

	drafts, err := s.store.ReadReviews(queryOptions...)
	if err != nil {
		return nil, err
	}
	if len(drafts) == 0 {
		return nil, nil
	}
	// A partial unique index (see migration 015) guarantees at most one draft
	// per (owner, media, season), so the first match is the only match. This
	// dedup is load-bearing: it's what keeps the editor's autosave from creating
	// duplicate drafts when its initial "create" requests race.
	return &drafts[0], nil
}

func (s Server) reviewsDraftsPut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := reviewIDFromRequestPath(r)
		if err != nil {
			http.Error(w, "Invalid review ID", http.StatusBadRequest)
			return
		}

		review, ok := s.loadOwnedReview(w, r, id)
		if !ok {
			return
		}

		if !review.IsDraft {
			http.Error(w, "Review already published", http.StatusBadRequest)
			return
		}

		parsedRequest, err := parseReviewPutRequest(r, false)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		review.Rating = parsedRequest.Rating
		review.Blurb = parsedRequest.Blurb
		review.Watched = parsedRequest.Watched

		if err := s.store.UpdateReview(review); err != nil {
			log.Printf("failed to update draft: %v", err)
			http.Error(w, fmt.Sprintf("Failed to update draft: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
