package handlers

import (
	"errors"
	"net/http"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

var errForbidden = errors.New("forbidden")

type accessService struct {
	db           AccessStore
	actor        screenjournal.Username
	actorIsAdmin bool
}

func (s Server) access(r *http.Request) accessService {
	return accessService{
		db:           s.getAccessDB(r),
		actor:        mustGetUsernameFromContext(r.Context()),
		actorIsAdmin: isAdmin(r.Context()),
	}
}

func (a accessService) isOwnerOrAdmin(owner screenjournal.Username) bool {
	return a.actor.Equal(owner) || a.actorIsAdmin
}

func (a accessService) readReview(id screenjournal.ReviewID) (screenjournal.Review, error) {
	return a.db.ReadReview(id)
}

func (a accessService) readComment(id screenjournal.CommentID) (screenjournal.ReviewComment, error) {
	return a.db.ReadComment(id)
}

func (a accessService) readReaction(id screenjournal.ReactionID) (screenjournal.ReviewReaction, error) {
	return a.db.ReadReaction(id)
}

func (a accessService) updateReview(
	id screenjournal.ReviewID,
	updated reviewPutRequest,
) (screenjournal.Review, error) {
	review, err := a.readReview(id)
	if err != nil {
		return screenjournal.Review{}, err
	}
	if !a.isOwnerOrAdmin(review.Owner) {
		return screenjournal.Review{}, errForbidden
	}

	review.Rating = updated.Rating
	review.Blurb = updated.Blurb
	review.Watched = updated.Watched

	if err := a.db.UpdateReview(review); err != nil {
		return screenjournal.Review{}, err
	}

	return review, nil
}

func (a accessService) deleteReview(id screenjournal.ReviewID) error {
	review, err := a.readReview(id)
	if err != nil {
		return err
	}
	if !a.isOwnerOrAdmin(review.Owner) {
		return errForbidden
	}

	if err := a.db.DeleteReview(id); err != nil {
		return err
	}

	return nil
}

func (a accessService) updateComment(
	id screenjournal.CommentID,
	commentText screenjournal.CommentText,
) (screenjournal.ReviewComment, error) {
	rc, err := a.readComment(id)
	if err != nil {
		return screenjournal.ReviewComment{}, err
	}
	if !a.isOwnerOrAdmin(rc.Owner) {
		return screenjournal.ReviewComment{}, errForbidden
	}

	rc.CommentText = commentText
	if err := a.db.UpdateComment(rc); err != nil {
		return screenjournal.ReviewComment{}, err
	}

	return rc, nil
}

func (a accessService) deleteComment(id screenjournal.CommentID) error {
	rc, err := a.readComment(id)
	if err != nil {
		return err
	}
	if !a.isOwnerOrAdmin(rc.Owner) {
		return errForbidden
	}

	if err := a.db.DeleteComment(id); err != nil {
		return err
	}

	return nil
}

func (a accessService) deleteReaction(id screenjournal.ReactionID) error {
	rr, err := a.readReaction(id)
	if err != nil {
		return err
	}
	if !a.isOwnerOrAdmin(rr.Owner) {
		return errForbidden
	}

	if err := a.db.DeleteReaction(id); err != nil {
		return err
	}

	return nil
}
