package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type (
	searchGetRequest struct {
		Query screenjournal.SearchQuery
	}

	searchMatch struct {
		TmdbID      int32
		Title       string
		ReleaseYear int
		PosterURL   string
	}
)

func (s Server) searchGet() http.HandlerFunc {
	t := template.Must(template.ParseFS(templatesFS, "templates/fragments/search-results.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := parseSearchGetRequest(r)
		if err != nil {
			log.Printf("failed to parse search query request: %v", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		res, err := s.metadataFinder.SearchMovies(req.Query)
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

func parseSearchGetRequest(r *http.Request) (searchGetRequest, error) {
	q, err := searchQueryFromQueryParams(r)
	if err != nil {
		return searchGetRequest{}, err
	}

	return searchGetRequest{
		Query: q,
	}, nil
}
