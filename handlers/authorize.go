package handlers

import (
	"errors"
	"net/http"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

var errForbidden = errors.New("forbidden")

type dbService struct {
	Store
	request *http.Request
}

func (d dbService) isOwnerOrAdmin(owner screenjournal.Username) bool {
	return mustGetUsernameFromContext(d.request.Context()).Equal(owner) ||
		isAdmin(d.request.Context())
}

func (d dbService) readReview(id screenjournal.ReviewID) (screenjournal.Review, error) {
	return d.ReadReview(id)
}

func (d dbService) readComment(id screenjournal.CommentID) (screenjournal.ReviewComment, error) {
	return d.ReadComment(id)
}

func (d dbService) readReaction(id screenjournal.ReactionID) (screenjournal.ReviewReaction, error) {
	return d.ReadReaction(id)
}

func (d dbService) updateReview(id screenjournal.ReviewID, updated reviewPutRequest) (screenjournal.Review, error) {
	review, err := d.readReview(id)
	if err != nil {
		return screenjournal.Review{}, err
	}
	if !d.isOwnerOrAdmin(review.Owner) {
		return screenjournal.Review{}, errForbidden
	}

	review.Rating = updated.Rating
	review.Blurb = updated.Blurb
	review.Watched = updated.Watched

	if err := d.UpdateReview(review); err != nil {
		return screenjournal.Review{}, err
	}

	return review, nil
}

func (d dbService) deleteReview(id screenjournal.ReviewID) error {
	review, err := d.readReview(id)
	if err != nil {
		return err
	}
	if !d.isOwnerOrAdmin(review.Owner) {
		return errForbidden
	}

	if err := d.DeleteReview(id); err != nil {
		return err
	}

	return nil
}

func (d dbService) updateComment(id screenjournal.CommentID, commentText screenjournal.CommentText) (screenjournal.ReviewComment, error) {
	rc, err := d.readComment(id)
	if err != nil {
		return screenjournal.ReviewComment{}, err
	}
	if !d.isOwnerOrAdmin(rc.Owner) {
		return screenjournal.ReviewComment{}, errForbidden
	}

	rc.CommentText = commentText
	if err := d.UpdateComment(rc); err != nil {
		return screenjournal.ReviewComment{}, err
	}

	return rc, nil
}

func (d dbService) deleteComment(id screenjournal.CommentID) error {
	rc, err := d.readComment(id)
	if err != nil {
		return err
	}
	if !d.isOwnerOrAdmin(rc.Owner) {
		return errForbidden
	}

	if err := d.DeleteComment(id); err != nil {
		return err
	}

	return nil
}

func (d dbService) deleteReaction(id screenjournal.ReactionID) error {
	rr, err := d.readReaction(id)
	if err != nil {
		return err
	}
	if !d.isOwnerOrAdmin(rr.Owner) {
		return errForbidden
	}

	if err := d.DeleteReaction(id); err != nil {
		return err
	}

	return nil
}
