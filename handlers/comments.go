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
	ReviewID    screenjournal.ReviewID
	CommentText screenjournal.CommentText
}

func (s Server) commentsPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := newCommentFromRequest(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		review, err := s.getDB(r).ReadReview(req.ReviewID)
		if err == store.ErrReviewNotFound {
			http.Error(w, "Review not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read review: %v", err), http.StatusInternalServerError)
			return
		}

		rc := screenjournal.ReviewComment{
			Review:      review,
			Owner:       mustGetUserFromContext(r.Context()).Username,
			CommentText: req.CommentText,
		}

		rc.ID, err = s.getDB(r).InsertComment(rc)
		if err != nil {
			log.Printf("failed to save comment: %v", err)
			http.Error(w, fmt.Sprintf("Failed to save comment: %v", err), http.StatusInternalServerError)
			return
		}

		respondJSON(w, struct {
			ID uint64 `json:"id"`
		}{
			ID: rc.ID.UInt64(),
		})
	}
}

func newCommentFromRequest(r *http.Request) (commentPostRequest, error) {
	var payload struct {
		ReviewID uint64 `json:"reviewId"`
		Comment  string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("failed to decode JSON request: %v", err)
		return commentPostRequest{}, err
	}

	rid, err := parse.ReviewID(payload.ReviewID)
	if err != nil {
		return commentPostRequest{}, err
	}

	comment, err := parse.CommentText(payload.Comment)
	if err != nil {
		return commentPostRequest{}, err
	}

	return commentPostRequest{
		ReviewID:    rid,
		CommentText: comment,
	}, nil
}
