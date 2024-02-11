package handlers_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/metadata"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

func TestSearch(t *testing.T) {
	movies := []metadata.MovieInfo{
		{
			TmdbID:      screenjournal.TmdbID(38),
			ImdbID:      screenjournal.ImdbID("tt0338013"),
			Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
			ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
		},
	}
	for _, tt := range []struct {
		description       string
		sessionToken      string
		searchQuery       string
		metadataFinderErr error
		statusExpected    int
		expectedResponse  string
	}{
		{
			description:    "returns matching search results",
			sessionToken:   "abc123",
			searchQuery:    "spo",
			statusExpected: http.StatusOK,
			expectedResponse: `{
				"matches": [
					{
						"tmdbId": 38,
						"title" "Eternal Sunshine of the Spotless Mind",
						"releaseDate": "TODO",
						"posterUrl": ""https://image.tmdb.org/t/p/w92/TODO"
					}
				]
			}`,
		},
		{
			description:    "prohibits search if user is not authenticated",
			sessionToken:   "",
			searchQuery:    "spo",
			statusExpected: http.StatusUnauthorized,
		},
		{
			description:    "prohibits search if user has an invalid session token",
			sessionToken:   "invalid-token",
			searchQuery:    "spo",
			statusExpected: http.StatusUnauthorized,
		},
		{
			description:       "returns internal server error if metadata finder fails",
			sessionToken:      "abc123",
			searchQuery:       "spo",
			metadataFinderErr: errors.New("dummy error"),
			statusExpected:    http.StatusInternalServerError,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()

			// TODO: Connect the session manager
			sessionManager := newMockSessionManager([]mockSessionEntry{})

			// TODO: Connect metadata finder with metadataFinderErr

			s := handlers.New(nilAuthenticator, nilAnnouncer, &sessionManager, dataStore, NewMockMetadataFinder(movies))

			req, err := http.NewRequest("GET", "/api/search", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.URL.RawQuery = fmt.Sprintf("query=%s", tt.searchQuery)
			req.Header.Add("Accept", "text/json")
			req.AddCookie(&http.Cookie{
				Name: mockSessionTokenName,
			})

			w := httptest.NewRecorder()
			s.Router().ServeHTTP(w, req)

			if got, want := w.Code, tt.statusExpected; got != want {
				t.Fatalf("httpStatus=%v, want=%v", got, want)
			}

			if w.Code != http.StatusOK {
				return
			}
		})
	}
}
