package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
)

type reactionPostRequest struct {
	ReviewID screenjournal.ReviewID
	Emoji    screenjournal.ReactionEmoji
}

type reactionForTemplate struct {
	ID          screenjournal.ReactionID
	Emoji       screenjournal.ReactionEmoji
	Owner       screenjournal.Username
	UserCanEdit bool
	Created     time.Time
}

// convertReactionsForTemplate converts reactions from the database format to
// the template format, setting UserCanEdit based on ownership and admin status.
func convertReactionsForTemplate(reactions []screenjournal.ReviewReaction, loggedInUsername screenjournal.Username, isAdminUser bool) []reactionForTemplate {
	reactionsForTemplate := make([]reactionForTemplate, len(reactions))
	for i, reaction := range reactions {
		reactionsForTemplate[i] = reactionForTemplate{
			ID:          reaction.ID,
			Emoji:       reaction.Emoji,
			Owner:       reaction.Owner,
			UserCanEdit: loggedInUsername.Equal(reaction.Owner) || isAdminUser,
			Created:     reaction.Created,
		}
	}
	return reactionsForTemplate
}

func (s Server) reactionsPost() http.HandlerFunc {
	t := template.Must(template.New("reviews-for-single-media-entry.html").
		Funcs(moviePageFns).
		ParseFS(templatesFS, "templates/pages/reviews-for-single-media-entry.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := parseReactionPostRequest(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		review, err := s.getDB(r).ReadReview(req.ReviewID)
		if err == store.ErrReviewNotFound {
			http.Error(w, "Review not found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("Failed to read review: %v", err)
			http.Error(w, fmt.Sprintf("Failed to read review: %v", err), http.StatusInternalServerError)
			return
		}

		rr := screenjournal.ReviewReaction{
			Review: review,
			Owner:  mustGetUsernameFromContext(r.Context()),
			Emoji:  req.Emoji,
		}

		rr.ID, err = s.getDB(r).InsertReaction(rr)
		if err != nil {
			log.Printf("failed to save reaction: %v", err)
			http.Error(w, fmt.Sprintf("Failed to save reaction: %v", err), http.StatusInternalServerError)
			return
		}

		// Reload all reactions for this review to render the updated section.
		reactions, err := s.getDB(r).ReadReactions(req.ReviewID)
		if err != nil {
			log.Printf("failed to read reactions: %v", err)
			http.Error(w, "Failed to read reactions", http.StatusInternalServerError)
			return
		}

		loggedInUsername := mustGetUsernameFromContext(r.Context())
		isAdminUser := isAdmin(r.Context())

		// Convert reactions to template format.
		reactionsForTemplate := convertReactionsForTemplate(reactions, loggedInUsername, isAdminUser)

		if err := t.ExecuteTemplate(w, "reactions-section", struct {
			ReviewID         screenjournal.ReviewID
			Reactions        []reactionForTemplate
			UserHasReacted   bool
			AvailableEmojis  []screenjournal.ReactionEmoji
			LoggedInUsername screenjournal.Username
		}{
			ReviewID:         req.ReviewID,
			Reactions:        reactionsForTemplate,
			UserHasReacted:   true,
			AvailableEmojis:  screenjournal.AllowedReactionEmojis(),
			LoggedInUsername: loggedInUsername,
		}); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("failed to render reactions section: %v", err)
			return
		}
	}
}

func (s Server) reactionsDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid, err := reactionIDFromRequestPath(r)
		if err != nil {
			http.Error(w, "Invalid reaction ID", http.StatusBadRequest)
			return
		}

		rr, err := s.getDB(r).ReadReaction(rid)
		if err == store.ErrReactionNotFound {
			http.Error(w, "Reaction not found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("failed to read reaction: %v", err)
			http.Error(w, fmt.Sprintf("Failed to read reaction: %v", err), http.StatusInternalServerError)
			return
		}

		loggedInUsername := mustGetUsernameFromContext(r.Context())
		if !loggedInUsername.Equal(rr.Owner) && !isAdmin(r.Context()) {
			http.Error(w, "Can't delete another user's reaction", http.StatusForbidden)
			return
		}

		if err := s.getDB(r).DeleteReaction(rid); err != nil {
			log.Printf("failed to delete reaction id=%v: %v", rid, err)
			http.Error(w, "Failed to delete reaction", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func parseReactionPostRequest(r *http.Request) (reactionPostRequest, error) {
	if err := r.ParseForm(); err != nil {
		log.Printf("failed to decode reaction POST request: %v", err)
		return reactionPostRequest{}, err
	}

	rid, err := parse.ReviewIDFromString(r.PostFormValue("review-id"))
	if err != nil {
		return reactionPostRequest{}, err
	}

	emoji, err := parse.ReactionEmoji(r.PostFormValue("emoji"))
	if err != nil {
		return reactionPostRequest{}, err
	}

	return reactionPostRequest{
		ReviewID: rid,
		Emoji:    emoji,
	}, nil
}

func reactionIDFromRequestPath(r *http.Request) (screenjournal.ReactionID, error) {
	return parse.ReactionID(r.PathValue("reactionID"))
}
