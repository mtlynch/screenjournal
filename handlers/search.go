package handlers

import (
	"fmt"
	"net/http"
)

type searchMatch struct {
	TmdbID      int32  `json:"tmdbId"`
	Title       string `json:"title"`
	ReleaseDate string `json:"releaseDate"`
	PosterURL   string `json:"posterUrl"`
}

func (s Server) searchGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		res, err := s.metadataFinder.Search(query)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to query metadata: %v", err), http.StatusInternalServerError)
			return
		}

		matches := make([]searchMatch, len(res.Matches))
		for i, m := range res.Matches {
			matches[i].TmdbID = m.TmdbID.Int32()
			matches[i].Title = m.Title
			matches[i].ReleaseDate = m.ReleaseDate
			matches[i].PosterURL = "https://image.tmdb.org/t/p/w92" + m.PosterPath
		}

		respondJSON(w, struct {
			Matches []searchMatch `json:"matches"`
		}{
			Matches: matches,
		})
	}
}
