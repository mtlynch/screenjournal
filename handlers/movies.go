package handlers

import (
	"fmt"
	"log"
	"net/http"
)

func (s Server) repopulateMoviesGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("repopulating movies metadata")

		rr, err := s.getDB(r).ReadReviews()
		if err != nil {
			log.Printf("failed to read reviews: %v", err)
			http.Error(w, fmt.Sprintf("failed to read reviews: %v", err), http.StatusInternalServerError)
			return
		}

		log.Printf("read movie data from %d reviews", len(rr))

		// We could parallelize this, but it's a maintenance function we use rarely,
		// so we're keeping it simple for now.
		for _, rev := range rr {
			movie, err := s.metadataFinder.GetMovie(rev.Movie.TmdbID)
			if err != nil {
				log.Printf("failed to get metadata for %s (tmdb ID=%v): %v", rev.Movie.Title, rev.Movie.TmdbID, err)
				http.Error(w, fmt.Sprintf("Failed to retrieve metadata: %v", err), http.StatusInternalServerError)
				return
			}

			// Update movie with latest metadata.
			movie.ID = rev.Movie.ID

			if err := s.getDB(r).UpdateMovie(movie); err != nil {
				log.Printf("failed to update metadata for %s (tmdb ID=%v): %v", rev.Movie.Title, rev.Movie.TmdbID, err)
				http.Error(w, fmt.Sprintf("Failed to save updated movie metadata: %v", err), http.StatusInternalServerError)
				return
			}
		}
	}
}
