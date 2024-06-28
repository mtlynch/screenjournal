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
			http.Error(w, fmt.Sprintf("failed to query metadata: %v", err), http.StatusInternalServerError)
		}

		matches := make([]searchMatch, len(res))
		for i, m := range res {
			matches[i].TmdbID = m.TmdbID.Int32()
			matches[i].Title = m.Title.String()
			matches[i].ReleaseDate = m.ReleaseDate.Time().Format("2006-01-02")
			matches[i].PosterURL = "https://image.tmdb.org/t/p/w92" + m.PosterPath.Path
		}

		respondJSON(w, struct {
			Matches []searchMatch `json:"matches"`
		}{
			Matches: matches,
		})
	}
}
