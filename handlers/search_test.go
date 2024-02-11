package handlers_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/go-test/deep"
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
			PosterPath:  url.URL{}, // TODO
		},
	}
	authenticatedToken := "dummy-auth-token"
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
			sessionToken:   authenticatedToken,
			searchQuery:    "spo",
			statusExpected: http.StatusOK,
			expectedResponse: `{
				"matches": [
					{
						"tmdbId": 38,
						"title": "Eternal Sunshine of the Spotless Mind",
						"releaseDate": "2004-03-19",
						"posterUrl": "https://image.tmdb.org/t/p/w92/TODO"
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
			sessionToken:      authenticatedToken,
			searchQuery:       "spo",
			metadataFinderErr: errors.New("dummy error"),
			statusExpected:    http.StatusInternalServerError,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()

			sessionManager := newMockSessionManager([]mockSessionEntry{
				{
					token: authenticatedToken,
					session: handlers.Session{
						Username: screenjournal.Username("dummyuser"),
					},
				},
			})

			// TODO: It feels like MetadataFinder is the wrong interface, and it
			// should be returning screenjournal-native types.
			metadataFinder := NewMockMetadataFinder(movies, tt.metadataFinderErr)

			s := handlers.New(nilAuthenticator, nilAnnouncer, &sessionManager, dataStore, metadataFinder)

			req, err := http.NewRequest("GET", "/api/search", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.URL.RawQuery = fmt.Sprintf("query=%s", tt.searchQuery)
			req.Header.Add("Accept", "text/json")
			req.AddCookie(&http.Cookie{
				Name:  mockSessionTokenName,
				Value: tt.sessionToken,
			})

			w := httptest.NewRecorder()
			s.Router().ServeHTTP(w, req)

			if got, want := w.Code, tt.statusExpected; got != want {
				t.Fatalf("httpStatus=%v, want=%v", got, want)
			}

			if w.Code != http.StatusOK {
				return
			}

			bodyBytes, err := io.ReadAll(w.Body)
			if err != nil {
				t.Fatalf("couldn't read HTTP response body: %v", err)
			}

			if got, want := string(bodyBytes), tt.expectedResponse; !jsonEqual(got, want, t) {
				t.Errorf("response doesn't match expected: %s", strings.Join(deep.Equal(got, want), "\n"))
			}
		})
	}
}

func jsonEqual(a, b string, t *testing.T) bool {
	var aj, bj interface{}
	if err := json.NewDecoder(strings.NewReader(a)).Decode(&aj); err != nil {
		t.Fatalf("error parsing JSON: %s, error: %v", a, err)
	}
	if err := json.NewDecoder(strings.NewReader(b)).Decode(&bj); err != nil {
		t.Fatalf("error parsing JSON: %s, error: %v", b, err)
	}
	return reflect.DeepEqual(aj, bj)
}
