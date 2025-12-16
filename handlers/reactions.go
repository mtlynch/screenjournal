package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
)

type reactionPostRequest struct {
	ReviewID screenjournal.ReviewID
	Emoji    screenjournal.ReactionEmoji
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
		// Hack so that when we render the reaction in the response, it has roughly
		// the correct creation time.
		rr.Created = time.Now()

		// Reload all reactions for this review to render the updated section.
		reactions, err := s.getDB(r).ReadReactions(req.ReviewID)
		if err != nil {
			log.Printf("failed to read reactions: %v", err)
			http.Error(w, "Failed to read reactions", http.StatusInternalServerError)
			return
		}

		loggedInUsername := mustGetUsernameFromContext(r.Context())

		// Check if the user has reacted (they just did, so this should be true).
		userHasReacted := false
		for _, reaction := range reactions {
			if reaction.Owner.Equal(loggedInUsername) {
				userHasReacted = true
				break
			}
		}

		if err := t.ExecuteTemplate(w, "reactions-section", struct {
			ReviewID         screenjournal.ReviewID
			Reactions        []screenjournal.ReviewReaction
			UserHasReacted   bool
			AvailableEmojis  []screenjournal.ReactionEmoji
			LoggedInUsername screenjournal.Username
			IsAdmin          bool
		}{
			ReviewID:         req.ReviewID,
			Reactions:        reactions,
			UserHasReacted:   userHasReacted,
			AvailableEmojis:  screenjournal.AllowedReactionEmojis(),
			LoggedInUsername: loggedInUsername,
			IsAdmin:          isAdmin(r.Context()),
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
	return parse.ReactionID(mux.Vars(r)["reactionID"])
}
