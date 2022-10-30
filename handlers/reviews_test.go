package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers"
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
			description: "valid request",
			payload: `{
					"title": "Eternal Sunshine of the Spotless Mind",
					"rating": 10,
					"watched":"2022-10-28T00:00:00-04:00",
					"blurb": "It's my favorite movie!"
				}`,
			expected: screenjournal.Review{
				Title:  "Eternal Sunshine of the Spotless Mind",
				Rating: screenjournal.Rating(10),
				// TODO: Watched
				Blurb: screenjournal.Blurb("It's my favorite movie!"),
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
				if r.Title == tt.expected.Title && r.Rating == tt.expected.Rating && r.Blurb == tt.expected.Blurb {
					found = true
				}
			}
			if !found {
				t.Fatalf("Did not find expected review: %s", tt.expected.Title)
			}
		})
	}
}
