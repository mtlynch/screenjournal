package handlers_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

type mockAuthenticator struct{}

func (ma mockAuthenticator) StartSession(w http.ResponseWriter, r *http.Request) {}

func (ma mockAuthenticator) ClearSession(w http.ResponseWriter) {}

func (ma mockAuthenticator) Authenticate(r *http.Request) bool {
	return true
}

func TestReviewsPostAcceptsValidRequest(t *testing.T) {
	for _, tt := range []struct {
		description string
		payload     string
		expected    screenjournal.Review
	}{
		{
			description: "valid request with all fields populated",
			payload: `{
					"title": "Eternal Sunshine of the Spotless Mind",
					"rating": 10,
					"watched":"2022-10-28T00:00:00-04:00",
					"blurb": "It's my favorite movie!"
				}`,
			expected: screenjournal.Review{
				Title:   "Eternal Sunshine of the Spotless Mind",
				Rating:  screenjournal.Rating(10),
				Watched: mustParseWatchDate("2022-10-28T00:00:00-04:00"),
				Blurb:   screenjournal.Blurb("It's my favorite movie!"),
			},
		},
		{
			description: "valid request without a blurb",
			payload: `{
					"title": "Dirty Work",
					"rating": 9,
					"watched":"2022-10-21T00:00:00-04:00",
					"blurb": ""
				}`,
			expected: screenjournal.Review{
				Title:   "Dirty Work",
				Rating:  screenjournal.Rating(9),
				Watched: mustParseWatchDate("2022-10-21T00:00:00-04:00"),
				Blurb:   screenjournal.Blurb(""),
			},
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()

			s := handlers.New(mockAuthenticator{}, dataStore)

			req, err := http.NewRequest("POST", "/api/reviews", strings.NewReader(tt.payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Add("Content-Type", "text/json")

			w := httptest.NewRecorder()
			s.Router().ServeHTTP(w, req)

			if status := w.Code; status != http.StatusOK {
				t.Fatalf("%s: handler returned wrong status code: got %v want %v",
					tt.description, status, http.StatusOK)
			}

			rr, err := dataStore.ReadReviews()
			if err != nil {
				t.Fatalf("%s: failed to retrieve guest link from datastore: %v", tt.description, err)
			}

			found := false
			for _, r := range rr {
				if r.Title == tt.expected.Title &&
					r.Rating == tt.expected.Rating &&
					r.Blurb == tt.expected.Blurb &&
					r.Watched.Time().Equal(tt.expected.Watched.Time()) {
					found = true
				}
			}
			if !found {
				t.Fatalf("Did not find expected review: %s", tt.expected.Title)
			}
		})
	}
}

func TestReviewsPostRejectsInvalidRequest(t *testing.T) {
	for _, tt := range []struct {
		description string
		payload     string
	}{
		{
			description: "empty string",
			payload:     "",
		},
		{
			description: "empty payload",
			payload:     "{}",
		},
		{
			description: "invalid title field (non-string)",
			payload: `{
					"title": 5,
					"rating": 10,
					"watched":"2022-10-28T00:00:00-04:00",
					"blurb": "It's my favorite movie!"
				}`,
		},
		{
			description: "invalid title field (too long)",
			payload: fmt.Sprintf(`{
					"title": "%s",
					"rating": 10,
					"watched":"2022-10-28T00:00:00-04:00",
					"blurb": "It's my favorite movie!"
				}`, strings.Repeat("A", parse.MediaTitleMaxLength+1)),
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()

			s := handlers.New(mockAuthenticator{}, dataStore)

			req, err := http.NewRequest("POST", "/api/reviews", strings.NewReader(tt.payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Add("Content-Type", "text/json")

			w := httptest.NewRecorder()
			s.Router().ServeHTTP(w, req)

			if got, want := w.Code, http.StatusBadRequest; got != want {
				t.Fatalf("/api/reviews POST returned wrong status: got=%v, want=%v", got, want)
			}
		})
	}
}

func mustParseWatchDate(s string) screenjournal.WatchDate {
	wd, err := parse.WatchDate(s)
	if err != nil {
		log.Fatalf("failed to parse watch date: %s", s)
	}
	return wd
}
