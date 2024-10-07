package handlers

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type (
	searchGetRequest struct {
		Query     string // TODO: Replace with better type
		MediaType screenjournal.MediaType
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
			log.Printf("failed to parse search query: %v", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		res, err := s.metadataFinder.Search(req.Query)
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
	query := r.URL.Query().Get("query")
	if len(query) < 2 {
		return searchGetRequest{}, errors.New("invalid search query")
	}

	mediaType, err := parse.MediaType(r.URL.Query().Get("media-type"))
	if err != nil {
		return searchGetRequest{}, err
	}

	return searchGetRequest{
		Query:     query,
		MediaType: screenjournal.MediaType(mediaType),
	}, nil
}
