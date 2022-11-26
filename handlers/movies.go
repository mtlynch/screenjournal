package handlers

import (
	"fmt"
	"log"
	"net/http"
)

func (s Server) repopulateMoviesGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("repopulating movies metadata")

		rr, err := s.store.ReadReviews()
		if err != nil {
			log.Printf("failed to read reviews: %v", err)
			http.Error(w, fmt.Sprintf("failed to read reviews: %v", err), http.StatusInternalServerError)
			return
		}

		log.Printf("read data from %d reviews", len(rr))

		for _, rev := range rr {
			log.Printf("updating movie %s with latest metadata", rev.Movie.Title)

			movieInfo, err := s.metadataFinder.GetMovieInfo(rev.Movie.TmdbID)
			if err != nil {
				log.Printf("failed to get metadata for %s (tmdb ID=%v): %v", movieInfo.Title, movieInfo.TmdbID, err)
				http.Error(w, fmt.Sprintf("Failed to retrieve metadata: %v", err), http.StatusInternalServerError)
				return
			}

			rev.Movie.Title = movieInfo.Title

			if err := s.store.UpdateMovie(rev.Movie); err != nil {
				log.Printf("failed to get metadata for %s (tmdb ID=%v): %v", rev.Movie.Title, rev.Movie.TmdbID, err)
				http.Error(w, fmt.Sprintf("Failed to save updated movie metadata: %v", err), http.StatusInternalServerError)
				return
			}
		}
	}
}
