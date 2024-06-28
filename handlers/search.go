package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type searchMatch struct {
	TmdbID      int32
	Title       string
	ReleaseYear int
	PosterURL   string
}

func (s Server) searchGet() http.HandlerFunc {
	t := template.Must(template.ParseFS(templatesFS, "templates/fragments/search-results.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		if len(query) < 2 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		res, err := s.metadataFinder.Search(query)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to query metadata: %v", err), http.StatusInternalServerError)
		}

		const limit = 7 // Arbitrary limit to show 7 results max
		matches := []searchMatch{}
		for i, m := range res {
			if i >= limit {
				break
			}
			matches = append(matches, searchMatch{
				TmdbID:      m.TmdbID.Int32(),
				Title:       m.Title.String(),
				ReleaseYear: m.ReleaseDate.Year(),
				PosterURL:   "https://image.tmdb.org/t/p/w92" + m.PosterPath.Path,
			})
		}

		if err := t.Execute(w, struct {
			Results []searchMatch
		}{
			Results: matches,
		}); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("failed to render search results template: %v", err)
			return
		}
	}
}
