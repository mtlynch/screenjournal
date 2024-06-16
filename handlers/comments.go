package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
)

type commentPostRequest struct {
	ReviewID    screenjournal.ReviewID
	CommentText screenjournal.CommentText
}

type commentPutRequest struct {
	CommentID   screenjournal.CommentID
	CommentText screenjournal.CommentText
}

func (s Server) commentsPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := parseCommentPostRequest(r)
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

		now := time.Now()
		rc := screenjournal.ReviewComment{
			Review:      review,
			Owner:       mustGetUsernameFromContext(r.Context()),
			CommentText: req.CommentText,
			Created:     now,
			Modified:    now,
		}

		rc.ID, err = s.getDB(r).InsertComment(rc)
		if err != nil {
			log.Printf("failed to save comment: %v", err)
			http.Error(w, fmt.Sprintf("Failed to save comment: %v", err), http.StatusInternalServerError)
			return
		}

		funcMap := template.FuncMap{
			"formatCommentTime":   formatIso8601Datetime,
			"relativeCommentDate": relativeCommentDate,
			"isLoggedInUser": func(u screenjournal.Username) bool {
				return u.Equal(mustGetUsernameFromContext(r.Context()))
			},
			"splitByNewline": func(s string) []string {
				return strings.Split(s, "\n")
			},
		}

		t, err := template.New("view.html").Funcs(funcMap).ParseFS(templatesFS, "templates/fragments/comments/view.html")
		if err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("error=%s", err)
			return
		}

		if err := t.Execute(w, rc); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("error=%s", err)
			return
		}

		s.announcer.AnnounceNewComment(rc)
	}
}

func (s Server) commentsAddGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reviewID, err := reviewIDFromQueryParams(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		t, err := template.ParseFS(templatesFS, "templates/fragments/comments/add.html")
		if err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("error=%s", err)
			return
		}

		if err := t.Execute(w, struct {
			ID screenjournal.ReviewID
		}{
			ID: reviewID,
		}); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("failed to get add form: %v", err)
			return
		}
	}
}

func (s Server) commentsNewGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reviewID, err := reviewIDFromQueryParams(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		t, err := template.ParseFS(templatesFS, "templates/fragments/comments/edit.html")
		if err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("error=%s", err)
			return
		}

		if err := t.Execute(w, struct {
			ID screenjournal.ReviewID
		}{
			ID: reviewID,
		}); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("error=%s", err)
			return
		}
	}
}

func (s Server) commentsPut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := parseCommentPutRequest(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		rc, err := s.getDB(r).ReadComment(req.CommentID)
		if err == store.ErrCommentNotFound {
			http.Error(w, "Comment not found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("failed to read comment: %v", err)
			http.Error(w, fmt.Sprintf("Failed to read comment: %v", err), http.StatusInternalServerError)
			return
		}

		if !mustGetUsernameFromContext(r.Context()).Equal(rc.Owner) {
			http.Error(w, "Can't edit another user's comment", http.StatusForbidden)
			return
		}

		rc.CommentText = req.CommentText
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

		rc, err := s.getDB(r).ReadComment(cid)
		if err == store.ErrCommentNotFound {
			http.Error(w, "Comment not found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("failed to read comment: %v", err)
			http.Error(w, fmt.Sprintf("Failed to read comment: %v", err), http.StatusInternalServerError)
			return
		}

		if !mustGetUsernameFromContext(r.Context()).Equal(rc.Owner) {
			http.Error(w, "Can't delete another user's comment", http.StatusForbidden)
			return
		}

		if err := s.getDB(r).DeleteComment(cid); err != nil {
			log.Printf("failed to delete comment id=%v: %v", cid, err)
			http.Error(w, "Failed to delete comment: %v", http.StatusInternalServerError)
			return
		}
	}
}

func parseCommentPostRequest(r *http.Request) (commentPostRequest, error) {
	if err := r.ParseForm(); err != nil {
		log.Printf("failed to decode comment POST request: %v", err)
		return commentPostRequest{}, err
	}

	rid, err := parse.ReviewIDFromString(r.PostFormValue("review-id"))
	if err != nil {
		return commentPostRequest{}, err
	}

	comment, err := parse.CommentText(r.PostFormValue("comment"))
	if err != nil {
		return commentPostRequest{}, err
	}

	return commentPostRequest{
		ReviewID:    rid,
		CommentText: comment,
	}, nil
}

func parseCommentPutRequest(r *http.Request) (commentPutRequest, error) {
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

	comment, err := parse.CommentText(payload.Comment)
	if err != nil {
		return commentPutRequest{}, err
	}

	return commentPutRequest{
		CommentID:   cid,
		CommentText: comment,
	}, nil
}
