package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
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
		req, err := parseReviewPostRequest(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			log.Printf("couldn't parse review POST request: %v", err)
			return
		}

		review := screenjournal.Review{
			Owner:        mustGetUsernameFromContext(r.Context()),
			TvShowSeason: req.TvShowSeason,
			Rating:       req.Rating,
			Watched:      req.WatchDate,
			Blurb:        req.Blurb,
			Comments:     []screenjournal.ReviewComment{},
		}

		if req.MediaType == screenjournal.MediaTypeMovie {
			review.Movie, err = s.moviefromTmdbID(s.getDB(r), req.TmdbID)
			if err == store.ErrMovieNotFound {
				http.Error(w, fmt.Sprintf("Could not find movie with TMDB ID: %v", req.TmdbID), http.StatusNotFound)
				return
			} else if err != nil {
				log.Printf("failed to get local media ID for movie with TMDB ID %v: %v", req.TmdbID, err)
				http.Error(w, fmt.Sprintf("Failed to look up movie with TMDB ID: %v: %v", req.TmdbID, err), http.StatusInternalServerError)
				return
			}
		} else if req.MediaType == screenjournal.MediaTypeTvShow {
			review.TvShow, err = s.tvShowfromTmdbID(s.getDB(r), req.TmdbID)
			if err == store.ErrTvShowNotFound {
				http.Error(w, fmt.Sprintf("Could not find tv show with TMDB ID: %v", req.TmdbID), http.StatusNotFound)
				return
			} else if err != nil {
				log.Printf("failed to get local media ID for TV show with TMDB ID %v: %v", req.TmdbID, err)
				http.Error(w, fmt.Sprintf("Failed to look up TV show with TMDB ID: %v: %v", req.TmdbID, err), http.StatusInternalServerError)
				return
			}
		}

		review.ID, err = s.getDB(r).InsertReview(review)
		if err != nil {
			log.Printf("failed to save review: %v", err)
			http.Error(w, fmt.Sprintf("Failed to save review: %v", err), http.StatusInternalServerError)
			return
		}

		s.announcer.AnnounceNewReview(review)

		if review.MediaType() == screenjournal.MediaTypeMovie {
			http.Redirect(w, r, fmt.Sprintf("/movies/%d#review%d", review.Movie.ID.Int64(), review.ID.UInt64()), http.StatusSeeOther)
		} else {
			http.Redirect(w, r, fmt.Sprintf("/tv-shows/%d?season=%d#review%d", review.TvShow.ID.Int64(), review.TvShowSeason.UInt8(), review.ID.UInt64()), http.StatusSeeOther)
		}

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

		var newRoute string
		if review.MediaType() == screenjournal.MediaTypeMovie {
			newRoute = fmt.Sprintf("/movies/%d", review.Movie.ID.Int64())
		} else {
			newRoute = fmt.Sprintf("/tv-shows/%d?season=%d", review.TvShow.ID.Int64(), review.TvShowSeason.UInt8())
		}
		http.Redirect(w, r, newRoute, http.StatusSeeOther)
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

func (s Server) tvShowfromTmdbID(db Store, tmdbID screenjournal.TmdbID) (screenjournal.TvShow, error) {
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

func (s Server) updateTvShowDetailsInStore(db Store, tvShow screenjournal.TvShow) error {
	tvShowUpdated, err := s.metadataFinder.GetTvShow(tvShow.TmdbID)
	if err != nil {
		return err
	}

	if err := db.UpdateTvShow(tvShowUpdated); err != nil {
		return err
	}

	return nil
}
