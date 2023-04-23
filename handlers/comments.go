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

type commentPostRequest struct {
	ReviewID screenjournal.ReviewID
	Comment  screenjournal.Comment
}

func (s Server) commentsPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid, err := reviewIDFromRequestPath(r)
		if err != nil {
			http.Error(w, "Invalid review ID", http.StatusBadRequest)
			return
		}

		review, err := s.store.ReadReview(rid)
		if err == store.ErrReviewNotFound {
			http.Error(w, "Review not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read review: %v", err), http.StatusInternalServerError)
			return
		}

		req, err := newCommentFromRequest(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		rc := screenjournal.ReviewComment{
			Review:  review,
			Owner:   mustGetUserFromContext(r.Context()).Username,
			Comment: req.Comment,
		}

		rc.ID, err = s.store.InsertComment(rc)
		if err != nil {
			log.Printf("failed to save comment: %v", err)
			http.Error(w, fmt.Sprintf("Failed to save comment: %v", err), http.StatusInternalServerError)
			return
		}

		log.Printf("created new comment with ID=%v", rc.ID)

		respondJSON(w, struct {
			ID uint64 `json:"id"`
		}{
			ID: rc.ID.UInt64(),
		})

		// TODO: Announce comment
	}
}

func newCommentFromRequest(r *http.Request) (commentPostRequest, error) {
	var payload struct {
		Comment string `json:"comment"`
	}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Printf("failed to decode JSON request: %v", err)
		return commentPostRequest{}, err
	}

	comment, err := parse.Comment(payload.Comment)
	if err != nil {
		return commentPostRequest{}, err
	}

	return commentPostRequest{
		Comment: comment,
	}, nil
}
