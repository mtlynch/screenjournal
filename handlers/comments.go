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

type commentPostRequest struct {
	ReviewID    screenjournal.ReviewID
	CommentID   *screenjournal.CommentID
	CommentText screenjournal.CommentText
}

type commentPutRequest struct {
	CommentID   screenjournal.CommentID
	CommentText screenjournal.CommentText
}

func (s Server) commentsAddGet() http.HandlerFunc {
	t := template.Must(template.New("movies-view.html").
		Funcs(moviePageFns).
		ParseFS(templatesFS, "templates/pages/movies-view.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		reviewID, err := reviewIDFromQueryParams(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		if err := t.ExecuteTemplate(w, "add-comment-button", struct {
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

func (s Server) commentsEditGet() http.HandlerFunc {
	t := template.Must(template.ParseFS(templatesFS, "templates/fragments/comments-edit.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		reviewID, err := reviewIDFromQueryParams(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			log.Printf("error=%v", err)
			return
		}

		var commentID *screenjournal.CommentID
		var commentText *screenjournal.CommentText
		if id, err := commentIDFromQueryParams(r); err == nil {
			rc, err := s.getDB(r).ReadComment(id)
			if err == store.ErrCommentNotFound {
				http.Error(w, "Comment not found", http.StatusNotFound)
				return
			} else if err != nil {
				log.Printf("failed to read comment: %v", err)
				http.Error(w, fmt.Sprintf("Failed to read comment: %v", err), http.StatusInternalServerError)
				return
			}
			commentID = &rc.ID
			commentText = &rc.CommentText
		}

		var cID screenjournal.CommentID
		if commentID != nil {
			cID = *commentID
		}
		var cText screenjournal.CommentText
		if commentText != nil {
			cText = *commentText
		}

		if err := t.Execute(w, struct {
			ReviewID    screenjournal.ReviewID
			CommentID   screenjournal.CommentID
			CommentText screenjournal.CommentText
		}{
			ReviewID:    reviewID,
			CommentID:   cID,
			CommentText: cText,
		}); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("error=%v", err)
			return
		}
	}
}

func (s Server) commentsGet() http.HandlerFunc {
	t := template.Must(template.New("movies-view.html").
		Funcs(moviePageFns).
		ParseFS(templatesFS, "templates/pages/movies-view.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := commentIDFromRequestPath(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		rc, err := s.getDB(r).ReadComment(id)
		if err == store.ErrCommentNotFound {
			http.Error(w, "Comment not found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("failed to read comment: %v", err)
			http.Error(w, fmt.Sprintf("Failed to read comment: %v", err), http.StatusInternalServerError)
			return
		}

		if err := t.ExecuteTemplate(w, "comment", struct {
			Comment          screenjournal.ReviewComment
			LoggedInUsername screenjournal.Username
		}{
			Comment:          rc,
			LoggedInUsername: mustGetUsernameFromContext(r.Context()),
		}); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("failed to render: %v", err) // TODO: Better error
			return
		}
	}
}

func (s Server) commentsPost() http.HandlerFunc {
	t := template.Must(template.New("movies-view.html").
		Funcs(moviePageFns).
		ParseFS(templatesFS, "templates/pages/movies-view.html"))
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

		rc := screenjournal.ReviewComment{
			Review:      review,
			Owner:       mustGetUsernameFromContext(r.Context()),
			CommentText: req.CommentText,
		}

		rc.ID, err = s.getDB(r).InsertComment(rc)
		if err != nil {
			log.Printf("failed to save comment: %v", err)
			http.Error(w, fmt.Sprintf("Failed to save comment: %v", err), http.StatusInternalServerError)
			return
		}
		// Hack so that when we render the comment in the response, it has roughly
		// the correct creation time.
		rc.Created = time.Now()

		if err := t.ExecuteTemplate(w, "comment", struct {
			Comment          screenjournal.ReviewComment
			LoggedInUsername screenjournal.Username
		}{
			Comment:          rc,
			LoggedInUsername: mustGetUsernameFromContext(r.Context()),
		}); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("error=%v", err)
			return
		}

		s.announcer.AnnounceNewComment(rc)
	}
}

func (s Server) commentsPut() http.HandlerFunc {
	t := template.Must(template.New("movies-view.html").
		Funcs(moviePageFns).
		ParseFS(templatesFS, "templates/pages/movies-view.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := parseCommentPutRequest(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			log.Printf("invalid comments PUT request: %v", err)
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

		if err := t.ExecuteTemplate(w, "comment", struct {
			Comment          screenjournal.ReviewComment
			LoggedInUsername screenjournal.Username
		}{
			Comment:          rc,
			LoggedInUsername: mustGetUsernameFromContext(r.Context()),
		}); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("failed to render: %v", err) // TODO: Better error
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

		w.WriteHeader(http.StatusNoContent)
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

	var pCid *screenjournal.CommentID
	if cid, err := commentIDFromQueryParams(r); err == nil {
		pCid = &cid
	}

	comment, err := parse.CommentText(r.PostFormValue("comment"))
	if err != nil {
		return commentPostRequest{}, err
	}

	return commentPostRequest{
		ReviewID:    rid,
		CommentID:   pCid,
		CommentText: comment,
	}, nil
}

func parseCommentPutRequest(r *http.Request) (commentPutRequest, error) {
	cid, err := commentIDFromRequestPath(r)
	if err != nil {
		return commentPutRequest{}, err
	}

	comment, err := parse.CommentText(r.PostFormValue("comment"))
	if err != nil {
		return commentPutRequest{}, err
	}

	return commentPutRequest{
		CommentID:   cid,
		CommentText: comment,
	}, nil
}
