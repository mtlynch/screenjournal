package handlers

import (
	"fmt"
	"net/http"
)

type searchMatch struct {
	ID          int    `json:"tmdbId"`
	Title       string `json:"title"`
	ReleaseDate string `json:"releaseDate"`
	PosterURL   string `json:"posterUrl"`
}

func (s Server) searchGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		res, err := s.metadataFinder.Search(query)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to query metadata: %v", err), http.StatusInternalServerError)
		}

		matches := make([]searchMatch, len(res.Matches))
		for i, m := range res.Matches {
			matches[i].ID = m.ID.Int()
			matches[i].Title = m.Title
			matches[i].ReleaseDate = m.ReleaseDate
			matches[i].PosterURL = m.PosterURL
		}

		respondJSON(w, struct {
			Matches []searchMatch `json:"matches"`
		}{
			Matches: matches,
		})
	}
}
