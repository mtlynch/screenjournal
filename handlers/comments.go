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

type commentPutRequest struct {
	CommentID screenjournal.CommentID
	Comment   screenjournal.Comment
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
			Review:  review,
			Owner:   mustGetUserFromContext(r.Context()).Username,
			Comment: req.Comment,
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

		// TODO: Announce comment
	}
}

func (s Server) commentsPut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := commentFromPutRequest(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		rc, err := s.getDB(r).ReadComment(req.CommentID)
		if err == store.ErrCommentNotFound {
			http.Error(w, "Comment not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read comment: %v", err), http.StatusInternalServerError)
			return
		}

		if mustGetUserFromContext(r.Context()).Username != rc.Owner {
			http.Error(w, "Can't edit another user's comment", http.StatusForbidden)
			return
		}

		rc.Comment = req.Comment
		if err := s.getDB(r).UpdateComment(rc); err != nil {
			log.Printf("failed to update comment: %v", err)
			http.Error(w, fmt.Sprintf("Failed to update comment: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) commentsDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cid, err := commentIDFromRequestPath(r)
		if err != nil {
			http.Error(w, "Invalid comment ID", http.StatusBadRequest)
			return
		}

		if err := s.getDB(r).DeleteComment(cid); err != nil {
			log.Printf("failed to delete comment id=%v: %v", cid, err)
			http.Error(w, "Failed to delete comment: %v", http.StatusInternalServerError)
			return
		}
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

	comment, err := parse.Comment(payload.Comment)
	if err != nil {
		return commentPostRequest{}, err
	}

	return commentPostRequest{
		ReviewID: rid,
		Comment:  comment,
	}, nil
}

func commentFromPutRequest(r *http.Request) (commentPutRequest, error) {
	cid, err := commentIDFromRequestPath(r)
	if err != nil {
		return commentPutRequest{}, err
	}

	var payload struct {
		Comment string `json:"comment"`
	}
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("failed to decode JSON request: %v", err)
		return commentPutRequest{}, err
	}

	comment, err := parse.Comment(payload.Comment)
	if err != nil {
		return commentPutRequest{}, err
	}

	return commentPutRequest{
		CommentID: cid,
		Comment:   comment,
	}, nil
}
