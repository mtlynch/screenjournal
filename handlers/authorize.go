package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
)

// readOwnedReview reads a review and verifies ownership. Returns false if it
// wrote an error response.
func (s Server) readOwnedReview(w http.ResponseWriter, r *http.Request, id screenjournal.ReviewID) (screenjournal.Review, bool) {
	review, err := s.getDB(r).ReadReview(id)
	if err == store.ErrReviewNotFound {
		http.Error(w, "Review not found", http.StatusNotFound)
		return screenjournal.Review{}, false
	} else if err != nil {
		log.Printf("failed to read review: %v", err)
		http.Error(w, fmt.Sprintf("Failed to read review: %v", err), http.StatusInternalServerError)
		return screenjournal.Review{}, false
	}
	if !mustGetUsernameFromContext(r.Context()).Equal(review.Owner) {
		http.Error(w, "You can't modify another user's review", http.StatusForbidden)
		return screenjournal.Review{}, false
	}
	return review, true
}

// readOwnedComment reads a comment and verifies ownership. Returns false if it
// wrote an error response.
func (s Server) readOwnedComment(w http.ResponseWriter, r *http.Request, id screenjournal.CommentID) (screenjournal.ReviewComment, bool) {
	rc, err := s.getDB(r).ReadComment(id)
	if err == store.ErrCommentNotFound {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return screenjournal.ReviewComment{}, false
	} else if err != nil {
		log.Printf("failed to read comment: %v", err)
		http.Error(w, fmt.Sprintf("Failed to read comment: %v", err), http.StatusInternalServerError)
		return screenjournal.ReviewComment{}, false
	}
	if !mustGetUsernameFromContext(r.Context()).Equal(rc.Owner) {
		http.Error(w, "Can't modify another user's comment", http.StatusForbidden)
		return screenjournal.ReviewComment{}, false
	}
	return rc, true
}

// readOwnedReaction reads a reaction and verifies ownership or admin status.
// Returns false if it wrote an error response.
func (s Server) readOwnedReaction(w http.ResponseWriter, r *http.Request, id screenjournal.ReactionID) (screenjournal.ReviewReaction, bool) {
	rr, err := s.getDB(r).ReadReaction(id)
	if err == store.ErrReactionNotFound {
		http.Error(w, "Reaction not found", http.StatusNotFound)
		return screenjournal.ReviewReaction{}, false
	} else if err != nil {
		log.Printf("failed to read reaction: %v", err)
		http.Error(w, fmt.Sprintf("Failed to read reaction: %v", err), http.StatusInternalServerError)
		return screenjournal.ReviewReaction{}, false
	}
	if !mustGetUsernameFromContext(r.Context()).Equal(rr.Owner) && !isAdmin(r.Context()) {
		http.Error(w, "Can't delete another user's reaction", http.StatusForbidden)
		return screenjournal.ReviewReaction{}, false
	}
	return rr, true
}
