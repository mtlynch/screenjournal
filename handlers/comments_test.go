package handlers_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/auth/simple"
	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
	"github.com/mtlynch/screenjournal/v2/metadata"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

type commentsTestData struct {
	users struct {
		userA screenjournal.User
		userB screenjournal.User
	}
	sessions struct {
		userA mockSession
	}
	movies struct {
		theWaterBoy screenjournal.Movie
	}
	reviews struct {
		userBTheWaterBoy screenjournal.Review
	}
}

func makeCommentsTestData() commentsTestData {
	td := commentsTestData{}
	td.users.userA = screenjournal.User{
		Username: screenjournal.Username("userA"),
	}
	td.sessions.userA = mockSession{
		token: "abc123",
		session: sessions.Session{
			User: td.users.userA,
		},
	}
	td.users.userB = screenjournal.User{
		Username: screenjournal.Username("userB"),
	}
	td.movies.theWaterBoy = screenjournal.Movie{
		ID:          screenjournal.MovieID(1),
		Title:       screenjournal.MediaTitle("The Waterboy"),
		ReleaseDate: mustParseReleaseDate("1998-11-06"),
	}
	td.reviews.userBTheWaterBoy = screenjournal.Review{
		ID:      screenjournal.ReviewID(1),
		Owner:   td.users.userB.Username,
		Rating:  screenjournal.Rating(5),
		Movie:   td.movies.theWaterBoy,
		Watched: mustParseWatchDate("2020-10-05T20:18:55-04:00"),
		Blurb:   screenjournal.Blurb("I love water!"),
	}
	return td
}

func TestCommentsPost(t *testing.T) {
	for _, tt := range []struct {
		description      string
		payload          string
		sessionToken     string
		sessions         []mockSession
		movies           []screenjournal.Movie
		reviews          []screenjournal.Review
		status           int
		expectedComments []screenjournal.ReviewComment
	}{
		{
			description: "allows user to comment on an existing review",
			payload: `{
					"reviewId": 1,
					"comment": "Good insights!"
				}`,
			sessionToken: makeCommentsTestData().sessions.userA.token,
			sessions: []mockSession{
				makeCommentsTestData().sessions.userA,
			},
			movies: []screenjournal.Movie{
				makeCommentsTestData().movies.theWaterBoy,
			},
			reviews: []screenjournal.Review{
				makeCommentsTestData().reviews.userBTheWaterBoy,
			},
			status: http.StatusOK,
			expectedComments: []screenjournal.ReviewComment{
				{
					ID:          screenjournal.CommentID(1),
					Owner:       makeCommentsTestData().users.userA.Username,
					CommentText: screenjournal.CommentText("Good insights!"),
					Review:      makeCommentsTestData().reviews.userBTheWaterBoy,
				},
			},
		},
		{
			description:  "rejects an invalid JSON request",
			payload:      `{banana`,
			sessionToken: makeCommentsTestData().sessions.userA.token,
			sessions: []mockSession{
				makeCommentsTestData().sessions.userA,
			},
			movies: []screenjournal.Movie{
				makeCommentsTestData().movies.theWaterBoy,
			},
			reviews: []screenjournal.Review{
				makeCommentsTestData().reviews.userBTheWaterBoy,
			},
			status: http.StatusBadRequest,
		},
		{
			description: "rejects a comment with invalid content",
			payload: `{
					"reviewId": 1,
					"comment": "<script>alert(1)</script>"
				}`,
			sessionToken: makeCommentsTestData().sessions.userA.token,
			sessions: []mockSession{
				makeCommentsTestData().sessions.userA,
			},
			movies: []screenjournal.Movie{
				makeCommentsTestData().movies.theWaterBoy,
			},
			reviews: []screenjournal.Review{
				makeCommentsTestData().reviews.userBTheWaterBoy,
			},
			status: http.StatusBadRequest,
		},
		{
			description: "rejects a comment with invalid review ID",
			payload: `{
					"reviewId": 0,
					"comment": "Good insights!"
				}`,
			sessionToken: makeCommentsTestData().sessions.userA.token,
			sessions: []mockSession{
				makeCommentsTestData().sessions.userA,
			},
			movies: []screenjournal.Movie{
				makeCommentsTestData().movies.theWaterBoy,
			},
			reviews: []screenjournal.Review{
				makeCommentsTestData().reviews.userBTheWaterBoy,
			},
			status: http.StatusBadRequest,
		},
		{
			description: "returns 404 if user attempts to comment on non-existent review",
			payload: `{
					"reviewId": 999,
					"comment": "Good insights!"
				}`,
			sessionToken: makeCommentsTestData().sessions.userA.token,
			sessions: []mockSession{
				makeCommentsTestData().sessions.userA,
			},
			movies: []screenjournal.Movie{
				makeCommentsTestData().movies.theWaterBoy,
			},
			reviews: []screenjournal.Review{
				makeCommentsTestData().reviews.userBTheWaterBoy,
			},
			status:           http.StatusNotFound,
			expectedComments: []screenjournal.ReviewComment{},
		},
		{
			description: "rejects comment update if user is not authenticated",
			payload: `{
					"reviewId": 1,
					"comment": "I haven't logged in, but I'm commenting anyway!"
				}`,
			sessionToken: "dummy-invalid-token",
			sessions: []mockSession{
				makeCommentsTestData().sessions.userA,
			},
			status: http.StatusUnauthorized,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			store := test_sqlite.New()

			// Populate datastore with dummy users.
			for _, s := range tt.sessions {
				store.InsertUser(s.session.User)
			}

			for _, movie := range tt.movies {
				if _, err := store.InsertMovie(movie); err != nil {
					panic(err)
				}
			}
			for _, review := range tt.reviews {
				if _, err := store.InsertReview(review); err != nil {
					panic(err)
				}
			}

			authenticator := simple.New(store)
			var nilMetadataFinder metadata.Finder
			sessionManager := newMockSessionManager(tt.sessions)

			s := handlers.New(authenticator, nilAnnouncer, &sessionManager, store, nilMetadataFinder)

			req, err := http.NewRequest("POST", "/api/comments", strings.NewReader(tt.payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Add("Content-Type", "text/json")
			req.AddCookie(&http.Cookie{
				Name:  "token",
				Value: tt.sessionToken,
			})

			w := httptest.NewRecorder()
			s.Router().ServeHTTP(w, req)

			if got, want := w.Code, tt.status; got != want {
				t.Fatalf("httpStatus=%v, want=%v", got, want)
			}

			if tt.status != http.StatusOK {
				return
			}

			comments, err := store.ReadComments(screenjournal.ReviewID(1))
			if err != nil {
				t.Fatalf("failed to read comments from datastore: %v", err)
			}
			if got, want := comments, tt.expectedComments; !reviewCommentsEqual(got, want) {
				t.Errorf("comments=%+v, got=%+v", got, want)
			}
		})
	}
}

func mustParseReleaseDate(s string) screenjournal.ReleaseDate {
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		log.Fatalf("failed to parse release date: %s", s)
	}
	return screenjournal.ReleaseDate(d)
}

func reviewCommentsEqual(a, b []screenjournal.ReviewComment) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		a[i].Created, b[i].Created = time.Time{}, time.Time{}
		a[i].Modified, b[i].Modified = time.Time{}, time.Time{}
		a[i].Review.Created, b[i].Review.Created = time.Time{}, time.Time{}
		a[i].Review.Modified, b[i].Review.Modified = time.Time{}, time.Time{}
	}

	return reflect.DeepEqual(a, b)
}
