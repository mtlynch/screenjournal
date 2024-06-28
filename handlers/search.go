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

		const limit = 7
		max := func() int {
			if limit < len(res) {
				return limit
			}
			return len(res)
		}()
		matches := make([]searchMatch, max)
		for i, m := range res {
			if i >= max {
				break
			}
			matches[i].TmdbID = m.TmdbID.Int32()
			matches[i].Title = m.Title.String()
			matches[i].ReleaseYear = m.ReleaseDate.Year()
			matches[i].PosterURL = "https://image.tmdb.org/t/p/w92" + m.PosterPath.Path
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
