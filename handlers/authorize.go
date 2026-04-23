package handlers

import (
	"errors"
	"net/http"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

var errForbidden = errors.New("forbidden")

func (s Server) isOwnerOrAdmin(r *http.Request, owner screenjournal.Username) bool {
	return mustGetUsernameFromContext(r.Context()).Equal(owner) || isAdmin(r.Context())
}

func (s Server) readReview(r *http.Request, id screenjournal.ReviewID) (screenjournal.Review, error) {
	return s.getDB(r).ReadReview(id)
}

func (s Server) readComment(r *http.Request, id screenjournal.CommentID) (screenjournal.ReviewComment, error) {
	return s.getDB(r).ReadComment(id)
}

func (s Server) readReaction(r *http.Request, id screenjournal.ReactionID) (screenjournal.ReviewReaction, error) {
	return s.getDB(r).ReadReaction(id)
}

func (s Server) updateReview(
	r *http.Request,
	id screenjournal.ReviewID,
	updated reviewPutRequest,
) (screenjournal.Review, error) {
	review, err := s.readReview(r, id)
	if err != nil {
		return screenjournal.Review{}, err
	}
	if !s.isOwnerOrAdmin(r, review.Owner) {
		return screenjournal.Review{}, errForbidden
	}

	review.Rating = updated.Rating
	review.Blurb = updated.Blurb
	review.Watched = updated.Watched

	if err := s.getDB(r).UpdateReview(review); err != nil {
		return screenjournal.Review{}, err
	}

	return review, nil
}

func (s Server) deleteReview(r *http.Request, id screenjournal.ReviewID) error {
	review, err := s.readReview(r, id)
	if err != nil {
		return err
	}
	if !s.isOwnerOrAdmin(r, review.Owner) {
		return errForbidden
	}

	if err := s.getDB(r).DeleteReview(id); err != nil {
		return err
	}

	return nil
}

func (s Server) updateComment(
	r *http.Request,
	id screenjournal.CommentID,
	commentText screenjournal.CommentText,
) (screenjournal.ReviewComment, error) {
	rc, err := s.readComment(r, id)
	if err != nil {
		return screenjournal.ReviewComment{}, err
	}
	if !s.isOwnerOrAdmin(r, rc.Owner) {
		return screenjournal.ReviewComment{}, errForbidden
	}

	rc.CommentText = commentText
	if err := s.getDB(r).UpdateComment(rc); err != nil {
		return screenjournal.ReviewComment{}, err
	}

	return rc, nil
}

func (s Server) deleteComment(r *http.Request, id screenjournal.CommentID) error {
	rc, err := s.readComment(r, id)
	if err != nil {
		return err
	}
	if !s.isOwnerOrAdmin(r, rc.Owner) {
		return errForbidden
	}

	if err := s.getDB(r).DeleteComment(id); err != nil {
		return err
	}

	return nil
}

func (s Server) deleteReaction(r *http.Request, id screenjournal.ReactionID) error {
	rr, err := s.readReaction(r, id)
	if err != nil {
		return err
	}
	if !s.isOwnerOrAdmin(r, rr.Owner) {
		return errForbidden
	}

	if err := s.getDB(r).DeleteReaction(id); err != nil {
		return err
	}

	return nil
}
