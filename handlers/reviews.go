package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
	"github.com/mtlynch/screenjournal/v2/store/sqlite"
)

type reviewPostRequest struct {
	MediaType    screenjournal.MediaType
	TmdbID       screenjournal.TmdbID
	TvShowSeason screenjournal.TvShowSeason
	Rating       screenjournal.Rating
	WatchDate    screenjournal.WatchDate
	Blurb        screenjournal.Blurb
}

type reviewPutRequest struct {
	Rating  screenjournal.Rating
	Blurb   screenjournal.Blurb
	Watched screenjournal.WatchDate
}

func (s Server) reviewsPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := parseReviewPostRequest(r, true)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			log.Printf("couldn't parse review POST request: %v", err)
			return
		}

		review, ok := s.reviewFromPostRequest(w, req, mustGetUsernameFromContext(r.Context()), false)
		if !ok {
			return
		}

		review.ID, err = s.store.InsertReview(review)
		if err != nil {
			log.Printf("failed to save review: %v", err)
			http.Error(w, fmt.Sprintf("Failed to save review: %v", err), http.StatusInternalServerError)
			return
		}

		s.announcer.AnnounceNewReview(review)

		http.Redirect(w, r, publishedReviewRoute(review), http.StatusSeeOther)
	}
}

func (s Server) reviewsPut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := reviewIDFromRequestPath(r)
		if err != nil {
			http.Error(w, "Invalid review ID", http.StatusBadRequest)
			return
		}

		review, ok := s.loadOwnedReview(w, r, id)
		if !ok {
			return
		}

		parsedRequest, err := parseReviewPutRequest(r, true)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		review.Rating = parsedRequest.Rating
		review.Blurb = parsedRequest.Blurb
		review.Watched = parsedRequest.Watched
		wasDraft := review.IsDraft
		review.IsDraft = false

		if err := s.store.UpdateReview(review); err != nil {
			log.Printf("failed to update review: %v", err)
			http.Error(w, fmt.Sprintf("Failed to update review: %v", err), http.StatusInternalServerError)
			return
		}

		if wasDraft && !review.IsDraft {
			s.announcer.AnnounceNewReview(review)
		}

		http.Redirect(w, r, publishedReviewRoute(review), http.StatusSeeOther)
	}
}

func (s Server) reviewsDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := reviewIDFromRequestPath(r)
		if err != nil {
			http.Error(w, "Invalid review ID", http.StatusBadRequest)
			return
		}

		review, err := s.store.ReadReview(id)
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

		if err := s.store.DeleteReview(id); err != nil {
			log.Printf("failed to delete review: %v", err)
			http.Error(w, fmt.Sprintf("Failed to delete review: %v", err), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/reviews", http.StatusSeeOther)
	}
}

// parseReviewPostRequest parses the fields shared by published reviews and
// drafts. When requireWatchDate is false (drafts), an empty watch date is
// allowed so that incomplete drafts can still be saved.
func parseReviewPostRequest(r *http.Request, requireWatchDate bool) (reviewPostRequest, error) {
	if err := r.ParseForm(); err != nil {
		log.Printf("failed to decode review POST request: %v", err)
		return reviewPostRequest{}, err
	}

	parsed := reviewPostRequest{}
	var err error

	if parsed.MediaType, err = parse.MediaType(r.PostFormValue("media-type")); err != nil {
		return reviewPostRequest{}, err
	}

	if parsed.TmdbID, err = parse.TmdbIDFromString(r.PostFormValue("tmdb-id")); err != nil {
		return reviewPostRequest{}, err
	}

	if parsed.MediaType == screenjournal.MediaTypeTvShow {
		if parsed.TvShowSeason, err = parse.TvShowSeason(r.PostFormValue("season")); err != nil {
			return reviewPostRequest{}, err
		}
	}

	if parsed.Rating, err = parse.RatingFromString(r.PostFormValue("rating")); err != nil {
		return reviewPostRequest{}, err
	}

	if parsed.WatchDate, err = parseWatchDate(r, requireWatchDate); err != nil {
		return reviewPostRequest{}, err
	}

	if parsed.Blurb, err = parse.Blurb(r.PostFormValue("blurb")); err != nil {
		return reviewPostRequest{}, err
	}

	return parsed, nil
}

// parseReviewPutRequest parses an update to an existing review or draft. When
// requireWatchDate is false (drafts), an empty watch date is allowed.
func parseReviewPutRequest(r *http.Request, requireWatchDate bool) (reviewPutRequest, error) {
	if err := r.ParseForm(); err != nil {
		log.Printf("failed to decode review PUT request: %v", err)
		return reviewPutRequest{}, err
	}

	parsed := reviewPutRequest{}
	var err error
	if parsed.Rating, err = parse.RatingFromString(r.PostFormValue("rating")); err != nil {
		return reviewPutRequest{}, err
	}

	if parsed.Watched, err = parseWatchDate(r, requireWatchDate); err != nil {
		return reviewPutRequest{}, err
	}

	if parsed.Blurb, err = parse.Blurb(r.PostFormValue("blurb")); err != nil {
		return reviewPutRequest{}, err
	}

	return parsed, nil
}

// parseWatchDate parses the watch-date form field. When the field is optional
// (drafts) and empty, it returns the zero WatchDate without error.
func parseWatchDate(r *http.Request, required bool) (screenjournal.WatchDate, error) {
	raw := r.PostFormValue("watch-date")
	if !required && raw == "" {
		return screenjournal.WatchDate{}, nil
	}
	return parse.WatchDate(raw)
}

// formBool reports whether the named form field holds a truthy value.
func formBool(r *http.Request, field string) bool {
	v := r.PostFormValue(field)
	return v == "true" || v == "1"
}

// loadOwnedReview reads the review at id and verifies the logged-in user owns
// it. On any failure it writes the appropriate error response and returns
// ok=false.
func (s Server) loadOwnedReview(w http.ResponseWriter, r *http.Request, id screenjournal.ReviewID) (screenjournal.Review, bool) {
	review, err := s.store.ReadReview(id)
	if err == store.ErrReviewNotFound {
		http.Error(w, "Review not found", http.StatusNotFound)
		return screenjournal.Review{}, false
	} else if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read review: %v", err), http.StatusInternalServerError)
		return screenjournal.Review{}, false
	}

	if !review.Owner.Equal(mustGetUsernameFromContext(r.Context())) {
		http.Error(w, "You can't edit another user's review", http.StatusForbidden)
		return screenjournal.Review{}, false
	}

	return review, true
}

// reviewFromPostRequest builds a Review from a parsed POST request, resolving
// its movie or TV show from the TMDB ID. On any failure it writes the
// appropriate error response and returns ok=false.
func (s Server) reviewFromPostRequest(w http.ResponseWriter, req reviewPostRequest, owner screenjournal.Username, isDraft bool) (screenjournal.Review, bool) {
	review := screenjournal.Review{
		Owner:        owner,
		TvShowSeason: req.TvShowSeason,
		Rating:       req.Rating,
		Watched:      req.WatchDate,
		Blurb:        req.Blurb,
		IsDraft:      isDraft,
		Comments:     []screenjournal.ReviewComment{},
	}

	var err error
	if req.MediaType == screenjournal.MediaTypeMovie {
		review.Movie, err = s.moviefromTmdbID(s.store, req.TmdbID)
		if err == store.ErrMovieNotFound {
			http.Error(w, fmt.Sprintf("Could not find movie with TMDB ID: %v", req.TmdbID), http.StatusNotFound)
			return screenjournal.Review{}, false
		} else if err != nil {
			log.Printf("failed to get local media ID for movie with TMDB ID %v: %v", req.TmdbID, err)
			http.Error(w, fmt.Sprintf("Failed to look up movie with TMDB ID: %v: %v", req.TmdbID, err), http.StatusInternalServerError)
			return screenjournal.Review{}, false
		}
	} else if req.MediaType == screenjournal.MediaTypeTvShow {
		review.TvShow, err = s.tvShowfromTmdbID(s.store, req.TmdbID)
		if err == store.ErrTvShowNotFound {
			http.Error(w, fmt.Sprintf("Could not find tv show with TMDB ID: %v", req.TmdbID), http.StatusNotFound)
			return screenjournal.Review{}, false
		} else if err != nil {
			log.Printf("failed to get local media ID for TV show with TMDB ID %v: %v", req.TmdbID, err)
			http.Error(w, fmt.Sprintf("Failed to look up TV show with TMDB ID: %v: %v", req.TmdbID, err), http.StatusInternalServerError)
			return screenjournal.Review{}, false
		}
	}

	return review, true
}

// publishedReviewRoute returns the URL of a published review, anchored to the
// review on its movie or TV show page.
func publishedReviewRoute(review screenjournal.Review) string {
	if review.MediaType() == screenjournal.MediaTypeMovie {
		return fmt.Sprintf("/movies/%d#review%d", review.Movie.ID.Int64(), review.ID.UInt64())
	}
	return fmt.Sprintf("/tv-shows/%d?season=%d#review%d", review.TvShow.ID.Int64(), review.TvShowSeason.UInt8(), review.ID.UInt64())
}

func (s Server) moviefromTmdbID(db sqlite.Store, tmdbID screenjournal.TmdbID) (screenjournal.Movie, error) {
	movie, err := db.ReadMovieByTmdbID(tmdbID)
	if err != nil && err != store.ErrMovieNotFound {
		return screenjournal.Movie{}, err
	} else if err == nil {
		return movie, nil
	}

	movie, err = s.metadataFinder.GetMovie(tmdbID)
	if err != nil {
		return screenjournal.Movie{}, err
	}

	movie.ID, err = db.InsertMovie(movie)
	if err != nil {
		return screenjournal.Movie{}, err
	}

	return movie, nil
}

func (s Server) tvShowfromTmdbID(db sqlite.Store, tmdbID screenjournal.TmdbID) (screenjournal.TvShow, error) {
	tvShow, err := db.ReadTvShowByTmdbID(tmdbID)
	if err != nil && err != store.ErrTvShowNotFound {
		return screenjournal.TvShow{}, err
	} else if err == nil {
		if err := s.updateTvShowDetailsInStore(db, tvShow); err != nil {
			log.Printf("failed to refresh TV show details: %v", err)
		}
		return tvShow, nil
	}

	tvShow, err = s.metadataFinder.GetTvShow(tmdbID)
	if err != nil {
		return screenjournal.TvShow{}, err
	}

	tvShow.ID, err = db.InsertTvShow(tvShow)
	if err != nil {
		return screenjournal.TvShow{}, err
	}

	return tvShow, nil
}

func (s Server) updateTvShowDetailsInStore(db sqlite.Store, tvShow screenjournal.TvShow) error {
	tvShowUpdated, err := s.metadataFinder.GetTvShow(tvShow.TmdbID)
	if err != nil {
		return err
	}

	if err := db.UpdateTvShow(tvShowUpdated); err != nil {
		return err
	}

	return nil
}
