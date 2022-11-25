package handlers_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-test/deep"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/metadata"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

type mockAuthenticator struct{}

func (ma mockAuthenticator) StartSession(w http.ResponseWriter, r *http.Request) {}

func (ma mockAuthenticator) ClearSession(w http.ResponseWriter) {}

func (ma mockAuthenticator) Authenticate(r *http.Request) bool {
	return true
}

type mockMetadataFinder struct {
	db map[screenjournal.TmdbID]metadata.MovieInfo
}

func (mf mockMetadataFinder) Search(query string) (metadata.MovieSearchResults, error) {
	return metadata.MovieSearchResults{}, nil
}

func (mf mockMetadataFinder) GetMovieInfo(id screenjournal.TmdbID) (metadata.MovieInfo, error) {
	var m metadata.MovieInfo
	var ok bool
	if m, ok = mf.db[id]; !ok {
		return metadata.MovieInfo{}, fmt.Errorf("could not find movie with id %d in mock DB", id.Int32())
	}
	return m, nil
}

func TestReviewsPostAcceptsValidRequest(t *testing.T) {
	for _, tt := range []struct {
		description     string
		payload         string
		localMovies     []screenjournal.Movie
		remoteMovieInfo []metadata.MovieInfo
		expected        screenjournal.Review
	}{
		{
			description: "valid request with all fields populated and movie information is in local DB",
			payload: `{
					"tmdbId": 38,
					"rating": 10,
					"watched":"2022-10-28T00:00:00-04:00",
					"blurb": "It's my favorite movie!"
				}`,
			localMovies: []screenjournal.Movie{
				{
					TmdbID: screenjournal.TmdbID(38),
					Title:  screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
				},
			},
			expected: screenjournal.Review{
				MovieID: screenjournal.MovieID(1),
				Owner:   screenjournal.Username("mike"),
				Title:   "Eternal Sunshine of the Spotless Mind",
				Rating:  screenjournal.Rating(10),
				Watched: mustParseWatchDate("2022-10-28T00:00:00-04:00"),
				Blurb:   screenjournal.Blurb("It's my favorite movie!"),
			},
		},
		{
			description: "valid request without a blurb",
			payload: `{
					"tmdbId": 14577,
					"rating": 9,
					"watched":"2022-10-21T00:00:00-04:00",
					"blurb": ""
				}`,
			localMovies: []screenjournal.Movie{
				{
					TmdbID: screenjournal.TmdbID(14577),
					Title:  screenjournal.MediaTitle("Dirty Work"),
				},
			},
			expected: screenjournal.Review{
				MovieID: screenjournal.MovieID(1),
				Owner:   screenjournal.Username("mike"),
				Title:   "Dirty Work",
				Rating:  screenjournal.Rating(9),
				Watched: mustParseWatchDate("2022-10-21T00:00:00-04:00"),
				Blurb:   screenjournal.Blurb(""),
			},
		},
		// TODO: Test when movie isn't local but it's in metadata finder
		// TODO: Test when movie isn't in local or metadata finder
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()

			for _, movie := range tt.localMovies {
				if _, err := dataStore.InsertMovie(movie); err != nil {
					panic(err)
				}
			}

			s := handlers.New(mockAuthenticator{}, dataStore, mockMetadataFinder{})

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
				t.Fatalf("%s: failed to retrieve review from datastore: %v", tt.description, err)
			}

			found := false
			for _, r := range rr {
				if reviewContentsEqual(r, tt.expected) {
					found = true
				}
			}
			if !found {
				t.Fatalf("did not find expected review: %s", tt.expected.Title)
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

			s := handlers.New(mockAuthenticator{}, dataStore, mockMetadataFinder{})

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

func TestReviewsPutAcceptsValidRequest(t *testing.T) {
	for _, tt := range []struct {
		description  string
		localMovies  []screenjournal.Movie
		priorReviews []screenjournal.Review
		route        string
		payload      string
		expected     screenjournal.Review
	}{
		{
			description: "valid request with all fields populated",
			localMovies: []screenjournal.Movie{
				{
					TmdbID: screenjournal.TmdbID(38),
					Title:  screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
				},
				{
					TmdbID: screenjournal.TmdbID(14577),
					Title:  screenjournal.MediaTitle("Dirty Work"),
				},
			},
			priorReviews: []screenjournal.Review{
				{
					ID:      screenjournal.ReviewID(1),
					MovieID: screenjournal.MovieID(1),
					Title:   "Eternal Sunshine of the Spotless Mind",
					Rating:  screenjournal.Rating(10),
					Watched: mustParseWatchDate("2022-10-28T00:00:00-04:00"),
					Blurb:   screenjournal.Blurb("It's my favorite movie!"),
				},
			},
			route: "/api/reviews/1",
			payload: `{
					"rating": 8,
					"watched":"2022-10-30T00:00:00-04:00",
					"blurb": "It's a pretty good movie!"
				}`,
			expected: screenjournal.Review{
				MovieID: screenjournal.MovieID(1),
				Title:   "Eternal Sunshine of the Spotless Mind",
				Rating:  screenjournal.Rating(8),
				Watched: mustParseWatchDate("2022-10-30T00:00:00-04:00"),
				Blurb:   screenjournal.Blurb("It's a pretty good movie!"),
			},
		},
		{
			description: "valid request without a blurb",
			localMovies: []screenjournal.Movie{
				{
					TmdbID: screenjournal.TmdbID(38),
					Title:  screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
				},
				{
					TmdbID: screenjournal.TmdbID(14577),
					Title:  screenjournal.MediaTitle("Dirty Work"),
				},
			},
			priorReviews: []screenjournal.Review{
				{
					ID:      screenjournal.ReviewID(1),
					MovieID: screenjournal.MovieID(2),
					Title:   "Dirty Work",
					Rating:  screenjournal.Rating(9),
					Watched: mustParseWatchDate("2022-10-21T00:00:00-04:00"),
					Blurb:   screenjournal.Blurb("Love Norm McDonald!"),
				},
			},
			route: "/api/reviews/1",
			payload: `{
					"rating": 5,
					"watched":"2022-10-28T00:00:00-04:00",
					"blurb": ""
				}`,
			expected: screenjournal.Review{
				MovieID: screenjournal.MovieID(2),
				Title:   "Dirty Work",
				Rating:  screenjournal.Rating(5),
				Watched: mustParseWatchDate("2022-10-28T00:00:00-04:00"),
				Blurb:   screenjournal.Blurb(""),
			},
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()

			for _, movie := range tt.localMovies {
				if _, err := dataStore.InsertMovie(movie); err != nil {
					panic(err)
				}
			}

			for _, r := range tt.priorReviews {
				if err := dataStore.InsertReview(r); err != nil {
					panic(err)
				}
			}

			s := handlers.New(mockAuthenticator{}, dataStore, mockMetadataFinder{})

			req, err := http.NewRequest("PUT", tt.route, strings.NewReader(tt.payload))
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
				t.Fatalf("%s: failed to retrieve review from datastore: %v", tt.description, err)
			}

			if got, want := len(rr), 1; got != want {
				t.Fatalf("unexpected review count: got %v, want %v", got, want)
			}

			if !reviewContentsEqual(rr[0], tt.expected) {
				t.Error(deep.Equal(rr[0], tt.expected))
			}
		})
	}
}

func TestReviewsPutRejectsInvalidRequest(t *testing.T) {
	for _, tt := range []struct {
		description  string
		localMovies  []screenjournal.Movie
		priorReviews []screenjournal.Review
		route        string
		payload      string
		status       int
	}{
		{
			description: "rejects request with review ID of zero",
			localMovies: []screenjournal.Movie{
				{
					TmdbID: screenjournal.TmdbID(38),
					Title:  screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
				},
			},
			priorReviews: []screenjournal.Review{
				{
					MovieID: screenjournal.MovieID(1),
					Title:   "Eternal Sunshine of the Spotless Mind",
					Rating:  screenjournal.Rating(10),
					Watched: mustParseWatchDate("2022-10-28T00:00:00-04:00"),
					Blurb:   screenjournal.Blurb("It's my favorite movie!"),
				},
			},
			route: "/api/reviews/0",
			payload: `{
					"rating": 8,
					"watched":"2022-10-30T00:00:00-04:00",
					"blurb": "It's a pretty good movie!"
				}`,
			status: http.StatusBadRequest,
		},
		{
			description: "rejects request with non-existent review ID",
			localMovies: []screenjournal.Movie{
				{
					TmdbID: screenjournal.TmdbID(38),
					Title:  screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
				},
			},
			priorReviews: []screenjournal.Review{
				{
					MovieID: screenjournal.MovieID(1),
					Title:   "Eternal Sunshine of the Spotless Mind",
					Rating:  screenjournal.Rating(10),
					Watched: mustParseWatchDate("2022-10-28T00:00:00-04:00"),
					Blurb:   screenjournal.Blurb("It's my favorite movie!"),
				},
			},
			route: "/api/reviews/9876",
			payload: `{
					"rating": 8,
					"watched":"2022-10-30T00:00:00-04:00",
					"blurb": "It's a pretty good movie!"
				}`,
			status: http.StatusNotFound,
		},
		{
			description: "rejects request with malformed JSON",
			localMovies: []screenjournal.Movie{
				{
					TmdbID: screenjournal.TmdbID(38),
					Title:  screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
				},
			},
			priorReviews: []screenjournal.Review{
				{
					MovieID: screenjournal.MovieID(1),
					Title:   "Eternal Sunshine of the Spotless Mind",
					Rating:  screenjournal.Rating(10),
					Watched: mustParseWatchDate("2022-10-28T00:00:00-04:00"),
					Blurb:   screenjournal.Blurb("It's my favorite movie!"),
				},
			},
			route: "/api/reviews/1",
			payload: `{
					"rating": 8,
					"watched":"2022-10-30T00:00:00-04:00",
					"blurb": "no JSON ending brace!"`,
			status: http.StatusBadRequest,
		},
		{
			description: "rejects request with missing rating field",
			localMovies: []screenjournal.Movie{
				{
					TmdbID: screenjournal.TmdbID(38),
					Title:  screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
				},
			},
			priorReviews: []screenjournal.Review{
				{
					MovieID: screenjournal.MovieID(1),
					Title:   "Eternal Sunshine of the Spotless Mind",
					Rating:  screenjournal.Rating(10),
					Watched: mustParseWatchDate("2022-10-28T00:00:00-04:00"),
					Blurb:   screenjournal.Blurb("It's my favorite movie!"),
				},
			},
			route: "/api/reviews/1",
			payload: `{
					"watched":"2022-10-30T00:00:00-04:00",
					"blurb": "It's a pretty good movie!"
				}`,
			status: http.StatusBadRequest,
		},
		{
			description: "rejects request with missing watched field",
			localMovies: []screenjournal.Movie{
				{
					TmdbID: screenjournal.TmdbID(38),
					Title:  screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
				},
			},
			priorReviews: []screenjournal.Review{
				{
					MovieID: screenjournal.MovieID(1),
					Title:   "Eternal Sunshine of the Spotless Mind",
					Rating:  screenjournal.Rating(10),
					Watched: mustParseWatchDate("2022-10-28T00:00:00-04:00"),
					Blurb:   screenjournal.Blurb("It's my favorite movie!"),
				},
			},
			route: "/api/reviews/1",
			payload: `{
					"rating": 8,
					"blurb": "It's a pretty good movie!"
				}`,
			status: http.StatusBadRequest,
		},
		{
			description: "rejects request with numeric blurb field",
			localMovies: []screenjournal.Movie{
				{
					TmdbID: screenjournal.TmdbID(38),
					Title:  screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
				},
			},
			priorReviews: []screenjournal.Review{
				{
					MovieID: screenjournal.MovieID(1),
					Title:   "Eternal Sunshine of the Spotless Mind",
					Rating:  screenjournal.Rating(10),
					Watched: mustParseWatchDate("2022-10-28T00:00:00-04:00"),
					Blurb:   screenjournal.Blurb("It's my favorite movie!"),
				},
			},
			route: "/api/reviews/1",
			payload: `{
					"rating": 8,
					"watched":"2022-10-30T00:00:00-04:00",
					"blurb": 6
				}`,
			status: http.StatusBadRequest,
		},
		{
			description: "rejects request with script tag in blurb field",
			localMovies: []screenjournal.Movie{
				{
					TmdbID: screenjournal.TmdbID(38),
					Title:  screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
				},
			},
			priorReviews: []screenjournal.Review{
				{
					MovieID: screenjournal.MovieID(1),
					Title:   "Eternal Sunshine of the Spotless Mind",
					Rating:  screenjournal.Rating(10),
					Watched: mustParseWatchDate("2022-10-28T00:00:00-04:00"),
					Blurb:   screenjournal.Blurb("It's my favorite movie!"),
				},
			},
			route: "/api/reviews/1",
			payload: `{
					"rating": 8,
					"watched":"2022-10-30T00:00:00-04:00",
					"blurb": "Nothing evil going on here...<script>alert(1)</script>"
				}`,
			status: http.StatusBadRequest,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()

			for _, movie := range tt.localMovies {
				if _, err := dataStore.InsertMovie(movie); err != nil {
					panic(err)
				}
			}

			for _, r := range tt.priorReviews {
				if err := dataStore.InsertReview(r); err != nil {
					panic(err)
				}
			}

			s := handlers.New(mockAuthenticator{}, dataStore, mockMetadataFinder{})

			req, err := http.NewRequest("PUT", tt.route, strings.NewReader(tt.payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Add("Content-Type", "text/json")

			w := httptest.NewRecorder()
			s.Router().ServeHTTP(w, req)

			if got, want := w.Code, tt.status; got != want {
				t.Fatalf("%s PUT returned wrong status: got=%v, want=%v", tt.route, got, want)
			}
		})
	}
}

func reviewContentsEqual(a, b screenjournal.Review) bool {
	a.ID, b.ID = screenjournal.ReviewID(0), screenjournal.ReviewID(0)
	a.Created, b.Created = time.Time{}, time.Time{}
	a.Modified, b.Modified = time.Time{}, time.Time{}

	return reflect.DeepEqual(a, b)

}

func mustParseWatchDate(s string) screenjournal.WatchDate {
	wd, err := parse.WatchDate(s)
	if err != nil {
		log.Fatalf("failed to parse watch date: %s", s)
	}
	return wd
}
