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
	}
}
