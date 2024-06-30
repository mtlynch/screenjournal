package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/metadata"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
)

type reviewPostRequest struct {
	TmdbID    screenjournal.TmdbID
	Rating    screenjournal.Rating
	WatchDate screenjournal.WatchDate
	Blurb     screenjournal.Blurb
}

type reviewPutRequest struct {
	Rating  screenjournal.Rating
	Blurb   screenjournal.Blurb
	Watched screenjournal.WatchDate
}

func (s Server) reviewsPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := parseReviewPostRequest(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			log.Printf("couldn't parse review POST request: %v", err)
			return
		}

		review := screenjournal.Review{
			Owner:    mustGetUsernameFromContext(r.Context()),
			Rating:   req.Rating,
			Watched:  req.WatchDate,
			Blurb:    req.Blurb,
			Comments: []screenjournal.ReviewComment{},
		}

		review.Movie, err = s.moviefromTmdbID(s.getDB(r), req.TmdbID)
		if err == store.ErrMovieNotFound {
			http.Error(w, fmt.Sprintf("Could not find movie with TMDB ID: %v", req.TmdbID), http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("failed to get local media ID for TMDB ID %v: %v", req.TmdbID, err)
			http.Error(w, fmt.Sprintf("Failed to look up TMDB ID: %v: %v", req.TmdbID, err), http.StatusInternalServerError)
			return
		}

		review.ID, err = s.getDB(r).InsertReview(review)
		if err != nil {
			log.Printf("failed to save review: %v", err)
			http.Error(w, fmt.Sprintf("Failed to save review: %v", err), http.StatusInternalServerError)
			return
		}

		s.announcer.AnnounceNewReview(review)

		http.Redirect(w, r, "/reviews", http.StatusSeeOther)
	}
}

func (s Server) reviewsPut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := reviewIDFromRequestPath(r)
		if err != nil {
			http.Error(w, "Invalid review ID", http.StatusBadRequest)
			return
		}

		review, err := s.getDB(r).ReadReview(id)
		if err == store.ErrReviewNotFound {
			http.Error(w, "Review not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read review: %v", err), http.StatusInternalServerError)
			return
		}

		loggedInUsername := mustGetUsernameFromContext(r.Context())
		if !review.Owner.Equal(loggedInUsername) {
			http.Error(w, "You can't edit another user's review", http.StatusForbidden)
			return
		}

		parsedRequest, err := parseReviewPutRequest(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		review.Rating = parsedRequest.Rating
		review.Blurb = parsedRequest.Blurb
		review.Watched = parsedRequest.Watched

		if err := s.getDB(r).UpdateReview(review); err != nil {
			log.Printf("failed to update review: %v", err)
			http.Error(w, fmt.Sprintf("Failed to update review: %v", err), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/reviews", http.StatusSeeOther)
	}
}

func (s Server) reviewsDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := reviewIDFromRequestPath(r)
		if err != nil {
			http.Error(w, "Invalid review ID", http.StatusBadRequest)
			return
		}

		review, err := s.getDB(r).ReadReview(id)
		if err == store.ErrReviewNotFound {
			http.Error(w, "Review not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read review: %v", err), http.StatusInternalServerError)
			return
		}

		loggedInUsername := mustGetUsernameFromContext(r.Context())
		if !review.Owner.Equal(loggedInUsername) {
			http.Error(w, "You can't delete another user's review", http.StatusForbidden)
			return
		}

		if err := s.getDB(r).DeleteReview(id); err != nil {
			log.Printf("failed to delete review: %v", err)
			http.Error(w, fmt.Sprintf("Failed to delete review: %v", err), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/reviews", http.StatusSeeOther)
	}
}

func parseReviewPostRequest(r *http.Request) (reviewPostRequest, error) {
	if err := r.ParseForm(); err != nil {
		log.Printf("failed to decode review POST request: %v", err)
		return reviewPostRequest{}, err
	}

	parsed := reviewPostRequest{}
	var err error

	if parsed.TmdbID, err = parse.TmdbIDFromString(r.PostFormValue("tmdb-id")); err != nil {
		return reviewPostRequest{}, err
	}

	if parsed.Rating, err = parse.RatingFromString(r.PostFormValue("rating")); err != nil {
		return reviewPostRequest{}, err
	}

	if parsed.WatchDate, err = parse.WatchDate(r.PostFormValue("watch-date")); err != nil {
		return reviewPostRequest{}, err
	}

	if parsed.Blurb, err = parse.Blurb(r.PostFormValue("blurb")); err != nil {
		return reviewPostRequest{}, err
	}

	return parsed, nil
}

func parseReviewPutRequest(r *http.Request) (reviewPutRequest, error) {
	if err := r.ParseForm(); err != nil {
		log.Printf("failed to decode review PUT request: %v", err)
		return reviewPutRequest{}, err
	}

	parsed := reviewPutRequest{}
	var err error
	if parsed.Rating, err = parse.RatingFromString(r.PostFormValue("rating")); err != nil {
		return reviewPutRequest{}, err
	}

	if parsed.Watched, err = parse.WatchDate(r.PostFormValue("watch-date")); err != nil {
		return reviewPutRequest{}, err
	}

	if parsed.Blurb, err = parse.Blurb(r.PostFormValue("blurb")); err != nil {
		return reviewPutRequest{}, err
	}

	return parsed, nil
}

func (s Server) moviefromTmdbID(db Store, tmdbID screenjournal.TmdbID) (screenjournal.Movie, error) {
	movie, err := db.ReadMovieByTmdbID(tmdbID)
	if err != nil && err != store.ErrMovieNotFound {
		return screenjournal.Movie{}, err
	} else if err == nil {
		return movie, nil
	}

	mi, err := s.metadataFinder.GetMovieInfo(tmdbID)
	if err != nil {
		return screenjournal.Movie{}, err
	}

	movie = metadata.MovieFromMovieInfo(mi)
	movie.ID, err = db.InsertMovie(movie)
	if err != nil {
		return screenjournal.Movie{}, err
	}

	return movie, nil
}
