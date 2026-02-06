package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
)

func (s Server) isOwnerOrAdmin(r *http.Request, owner screenjournal.Username) bool {
	return mustGetUsernameFromContext(r.Context()).Equal(owner) || isAdmin(r.Context())
}

func (s Server) readReviewOrWriteError(w http.ResponseWriter, r *http.Request, id screenjournal.ReviewID) (screenjournal.Review, bool) {
	review, err := s.getDB(r).ReadReview(id)
	if err == store.ErrReviewNotFound {
		http.Error(w, "Review not found", http.StatusNotFound)
		return screenjournal.Review{}, false
	} else if err != nil {
		log.Printf("failed to read review: %v", err)
		http.Error(w, fmt.Sprintf("Failed to read review: %v", err), http.StatusInternalServerError)
		return screenjournal.Review{}, false
	}
	return review, true
}

func (s Server) readCommentOrWriteError(w http.ResponseWriter, r *http.Request, id screenjournal.CommentID) (screenjournal.ReviewComment, bool) {
	rc, err := s.getDB(r).ReadComment(id)
	if err == store.ErrCommentNotFound {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return screenjournal.ReviewComment{}, false
	} else if err != nil {
		log.Printf("failed to read comment: %v", err)
		http.Error(w, fmt.Sprintf("Failed to read comment: %v", err), http.StatusInternalServerError)
		return screenjournal.ReviewComment{}, false
	}
	return rc, true
}

func (s Server) readReactionOrWriteError(w http.ResponseWriter, r *http.Request, id screenjournal.ReactionID) (screenjournal.ReviewReaction, bool) {
	rr, err := s.getDB(r).ReadReaction(id)
	if err == store.ErrReactionNotFound {
		http.Error(w, "Reaction not found", http.StatusNotFound)
		return screenjournal.ReviewReaction{}, false
	} else if err != nil {
		log.Printf("failed to read reaction: %v", err)
		http.Error(w, fmt.Sprintf("Failed to read reaction: %v", err), http.StatusInternalServerError)
		return screenjournal.ReviewReaction{}, false
	}
	return rr, true
}

func (s Server) updateReview(w http.ResponseWriter, r *http.Request, id screenjournal.ReviewID) (screenjournal.Review, bool) {
	review, ok := s.readReviewOrWriteError(w, r, id)
	if !ok {
		return screenjournal.Review{}, false
	}
	if !s.isOwnerOrAdmin(r, review.Owner) {
		http.Error(w, "You can't edit another user's review", http.StatusForbidden)
		return screenjournal.Review{}, false
	}

	parsedRequest, err := parseReviewPutRequest(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return screenjournal.Review{}, false
	}

	review.Rating = parsedRequest.Rating
	review.Blurb = parsedRequest.Blurb
	review.Watched = parsedRequest.Watched

	if err := s.getDB(r).UpdateReview(review); err != nil {
		log.Printf("failed to update review: %v", err)
		http.Error(w, fmt.Sprintf("Failed to update review: %v", err), http.StatusInternalServerError)
		return screenjournal.Review{}, false
	}

	return review, true
}

func (s Server) deleteReview(w http.ResponseWriter, r *http.Request, id screenjournal.ReviewID) bool {
	review, ok := s.readReviewOrWriteError(w, r, id)
	if !ok {
		return false
	}
	if !s.isOwnerOrAdmin(r, review.Owner) {
		http.Error(w, "You can't delete another user's review", http.StatusForbidden)
		return false
	}

	if err := s.getDB(r).DeleteReview(id); err != nil {
		log.Printf("failed to delete review: %v", err)
		http.Error(w, fmt.Sprintf("Failed to delete review: %v", err), http.StatusInternalServerError)
		return false
	}

	return true
}

func (s Server) updateComment(w http.ResponseWriter, r *http.Request, id screenjournal.CommentID) (screenjournal.ReviewComment, bool) {
	rc, ok := s.readCommentOrWriteError(w, r, id)
	if !ok {
		return screenjournal.ReviewComment{}, false
	}
	if !s.isOwnerOrAdmin(r, rc.Owner) {
		http.Error(w, "Can't edit another user's comment", http.StatusForbidden)
		return screenjournal.ReviewComment{}, false
	}

	parsedRequest, err := parseCommentPutRequest(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		log.Printf("invalid comment PUT request: %v", err)
		return screenjournal.ReviewComment{}, false
	}

	rc.CommentText = parsedRequest.CommentText
	if err := s.getDB(r).UpdateComment(rc); err != nil {
		log.Printf("failed to update comment: %v", err)
		http.Error(w, fmt.Sprintf("Failed to update comment: %v", err), http.StatusInternalServerError)
		return screenjournal.ReviewComment{}, false
	}

	return rc, true
}

func (s Server) deleteComment(w http.ResponseWriter, r *http.Request, id screenjournal.CommentID) bool {
	rc, ok := s.readCommentOrWriteError(w, r, id)
	if !ok {
		return false
	}
	if !s.isOwnerOrAdmin(r, rc.Owner) {
		http.Error(w, "Can't delete another user's comment", http.StatusForbidden)
		return false
	}

	if err := s.getDB(r).DeleteComment(id); err != nil {
		log.Printf("failed to delete comment id=%v: %v", id, err)
		http.Error(w, "Failed to delete comment: %v", http.StatusInternalServerError)
		return false
	}

	return true
}

func (s Server) deleteReaction(w http.ResponseWriter, r *http.Request, id screenjournal.ReactionID) bool {
	rr, ok := s.readReactionOrWriteError(w, r, id)
	if !ok {
		return false
	}
	if !s.isOwnerOrAdmin(r, rr.Owner) {
		http.Error(w, "Can't delete another user's reaction", http.StatusForbidden)
		return false
	}

	if err := s.getDB(r).DeleteReaction(id); err != nil {
		log.Printf("failed to delete reaction id=%v: %v", id, err)
		http.Error(w, "Failed to delete reaction", http.StatusInternalServerError)
		return false
	}

	return true
}
