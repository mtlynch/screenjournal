package handlers_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-test/deep"

	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
	"github.com/mtlynch/screenjournal/v2/metadata/tmdb"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

type commentsTestData struct {
	sessions struct {
		userA mockSessionEntry
		userB mockSessionEntry
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
	td.sessions.userA = mockSessionEntry{
		token: "abc123",
		session: sessions.Session{
			Username: screenjournal.Username("userA"),
		},
	}
	td.sessions.userB = mockSessionEntry{
		token: "def456",
		session: sessions.Session{
			Username: screenjournal.Username("userB"),
		},
	}
	td.movies.theWaterBoy = screenjournal.Movie{
		ID:          screenjournal.MovieID(1),
		Title:       screenjournal.MediaTitle("The Waterboy"),
		ReleaseDate: mustParseReleaseDate("1998-11-06"),
	}
	td.reviews.userBTheWaterBoy = screenjournal.Review{
		ID:      screenjournal.ReviewID(1),
		Owner:   td.sessions.userB.session.Username,
		Rating:  screenjournal.Rating(5),
		Movie:   td.movies.theWaterBoy,
		Watched: mustParseWatchDate("2020-10-05"),
		Blurb:   screenjournal.Blurb("I love water!"),
	}
	return td
}

func TestCommentsPost(t *testing.T) {
	for _, tt := range []struct {
		description      string
		payload          string
		sessionToken     string
		sessions         []mockSessionEntry
		movies           []screenjournal.Movie
		reviews          []screenjournal.Review
		status           int
		expectedComments []screenjournal.ReviewComment
	}{
		{
			description:  "allows user to comment on an existing review",
			payload:      "review-id=1&comment=Good+insights!",
			sessionToken: makeCommentsTestData().sessions.userA.token,
			sessions: []mockSessionEntry{
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
					Owner:       makeCommentsTestData().sessions.userA.session.Username,
					CommentText: screenjournal.CommentText("Good insights!"),
					Review:      makeCommentsTestData().reviews.userBTheWaterBoy,
				},
			},
		},
		{
			description:  "trims leading and trailing whitespace from a review comment",
			payload:      "review-id=1&comment=%0AYes%2C%20but%20can%20you%20strip%20my%20whitespace%3F%0A",
			sessionToken: makeCommentsTestData().sessions.userA.token,
			sessions: []mockSessionEntry{
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
					Owner:       makeCommentsTestData().sessions.userA.session.Username,
					CommentText: screenjournal.CommentText("Yes, but can you strip my whitespace?"),
					Review:      makeCommentsTestData().reviews.userBTheWaterBoy,
				},
			},
		},
		{
			description:  "rejects a request without a comment field",
			payload:      `banana=true`,
			sessionToken: makeCommentsTestData().sessions.userA.token,
			sessions: []mockSessionEntry{
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
			description:  "rejects a comment with invalid content",
			payload:      "review-id=1&comment=%3Cscript%3Ealert(1)%3C%2Fscript%3E",
			sessionToken: makeCommentsTestData().sessions.userA.token,
			sessions: []mockSessionEntry{
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
			description:  "rejects a comment with invalid review ID",
			payload:      "review-id=0&comment=Good+insights!",
			sessionToken: makeCommentsTestData().sessions.userA.token,
			sessions: []mockSessionEntry{
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
			description:  "returns 404 if user attempts to comment on non-existent review",
			payload:      "review-id=999&comment=Good+insights!",
			sessionToken: makeCommentsTestData().sessions.userA.token,
			sessions: []mockSessionEntry{
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
			description:  "rejects comment update if user is not authenticated",
			payload:      "review-id=1&comment=I+haven't+logged+in+but+I'm+commenting+anyway",
			sessionToken: "dummy-invalid-token",
			sessions: []mockSessionEntry{
				makeCommentsTestData().sessions.userA,
			},
			status: http.StatusUnauthorized,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			store := test_sqlite.New()

			for _, s := range tt.sessions {
				if err := store.InsertUser(screenjournal.User{
					Username: s.session.Username,
				}); err != nil {
					panic(err)
				}
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

			announcer := mockAnnouncer{}
			authenticator := auth.New(store)
			sessionManager := newMockSessionManager(tt.sessions)
			s := handlers.New(authenticator, &announcer, &sessionManager, store, nilMetadataFinder)

			req, err := http.NewRequest("POST", "/api/comments", strings.NewReader(tt.payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.AddCookie(&http.Cookie{
				Name:  mockSessionTokenName,
				Value: tt.sessionToken,
			})

			rec := httptest.NewRecorder()
			s.Router().ServeHTTP(rec, req)
			res := rec.Result()

			if got, want := res.StatusCode, tt.status; got != want {
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

			if got, want := len(announcer.announcedComments), 1; got != want {
				t.Fatalf("commentCountAnnounced=%d, want=%d", got, want)
			}

			clearUnpredictableCommentProperties(&announcer.announcedComments[0])
			clearUnpredictableCommentProperties(&tt.expectedComments[0])
			if !reflect.DeepEqual(announcer.announcedComments, tt.expectedComments) {
				t.Errorf("did not find expected announced comments: %v", deep.Equal(announcer.announcedComments, tt.expectedComments))
			}
		})
	}
}

func TestCommentsPut(t *testing.T) {
	for _, tt := range []struct {
		description      string
		route            string
		payload          string
		sessionToken     string
		sessions         []mockSessionEntry
		comments         []screenjournal.ReviewComment
		status           int
		expectedComments []screenjournal.ReviewComment
	}{
		{
			description:  "allows a user to update their own comment",
			route:        "/api/comments/1",
			payload:      "comment=So-so%20insights",
			sessionToken: makeCommentsTestData().sessions.userA.token,
			sessions: []mockSessionEntry{
				makeCommentsTestData().sessions.userA,
			},
			comments: []screenjournal.ReviewComment{
				{
					ID:          screenjournal.CommentID(1),
					Owner:       makeCommentsTestData().sessions.userA.session.Username,
					CommentText: screenjournal.CommentText("Good insights!"),
					Review:      makeCommentsTestData().reviews.userBTheWaterBoy,
				},
			},
			status: http.StatusOK,
			expectedComments: []screenjournal.ReviewComment{
				{
					ID:          screenjournal.CommentID(1),
					Owner:       makeCommentsTestData().sessions.userA.session.Username,
					CommentText: screenjournal.CommentText("So-so insights"),
					Review:      makeCommentsTestData().reviews.userBTheWaterBoy,
				},
			},
		},
		{
			description:  "prevents a user from updating a non-existent comment",
			route:        "/api/comments/999",
			payload:      "comment=So-so%20insights",
			sessionToken: makeCommentsTestData().sessions.userA.token,
			sessions: []mockSessionEntry{
				makeCommentsTestData().sessions.userA,
			},
			comments: []screenjournal.ReviewComment{
				{
					ID:          screenjournal.CommentID(1),
					Owner:       makeCommentsTestData().sessions.userA.session.Username,
					CommentText: screenjournal.CommentText("Good insights!"),
					Review:      makeCommentsTestData().reviews.userBTheWaterBoy,
				},
			},
			status: http.StatusNotFound,
		},
		{
			description:  "prevents a user from updating with missing comment field",
			route:        "/api/comments/1",
			payload:      "review-id=1",
			sessionToken: makeCommentsTestData().sessions.userA.token,
			sessions: []mockSessionEntry{
				makeCommentsTestData().sessions.userA,
			},
			comments: []screenjournal.ReviewComment{
				{
					ID:          screenjournal.CommentID(1),
					Owner:       makeCommentsTestData().sessions.userA.session.Username,
					CommentText: screenjournal.CommentText("Good insights!"),
					Review:      makeCommentsTestData().reviews.userBTheWaterBoy,
				},
			},
			status: http.StatusBadRequest,
		},
		{
			description:  "prevents a user from updating an invalid comment ID",
			route:        "/api/comments/banana",
			payload:      "comment=So-so%20insights",
			sessionToken: makeCommentsTestData().sessions.userA.token,
			sessions: []mockSessionEntry{
				makeCommentsTestData().sessions.userA,
			},
			comments: []screenjournal.ReviewComment{
				{
					ID:          screenjournal.CommentID(1),
					Owner:       makeCommentsTestData().sessions.userA.session.Username,
					CommentText: screenjournal.CommentText("Good insights!"),
					Review:      makeCommentsTestData().reviews.userBTheWaterBoy,
				},
			},
			status: http.StatusBadRequest,
		},
		{
			description:  "prevents a user from updating their comment with invalid content",
			route:        "/api/comments/1",
			payload:      "comment=%3Cscript%3Ealert(1)%3C%2Fscript%3E",
			sessionToken: makeCommentsTestData().sessions.userA.token,
			sessions: []mockSessionEntry{
				makeCommentsTestData().sessions.userA,
			},
			comments: []screenjournal.ReviewComment{
				{
					ID:          screenjournal.CommentID(1),
					Owner:       makeCommentsTestData().sessions.userA.session.Username,
					CommentText: screenjournal.CommentText("Good insights!"),
					Review:      makeCommentsTestData().reviews.userBTheWaterBoy,
				},
			},
			status: http.StatusBadRequest,
		},
		{
			description:  "prevents a user from updating someone else's comment",
			route:        "/api/comments/1",
			payload:      "comment=I%20overwrote%20your%20comment!",
			sessionToken: makeCommentsTestData().sessions.userB.token,
			sessions: []mockSessionEntry{
				makeCommentsTestData().sessions.userB,
			},
			comments: []screenjournal.ReviewComment{
				{
					ID:          screenjournal.CommentID(1),
					Owner:       makeCommentsTestData().sessions.userA.session.Username,
					CommentText: screenjournal.CommentText("Good insights!"),
					Review:      makeCommentsTestData().reviews.userBTheWaterBoy,
				},
			},
			status: http.StatusForbidden,
		},
		{
			description:  "prevents an unauthenticated user from updating any comment",
			route:        "/api/comments/1",
			payload:      "comment=I%20overwrote%20your%20comment!",
			sessionToken: "",
			sessions: []mockSessionEntry{
				makeCommentsTestData().sessions.userB,
			},
			comments: []screenjournal.ReviewComment{
				{
					ID:          screenjournal.CommentID(1),
					Owner:       makeCommentsTestData().sessions.userA.session.Username,
					CommentText: screenjournal.CommentText("Good insights!"),
					Review:      makeCommentsTestData().reviews.userBTheWaterBoy,
				},
			},
			status: http.StatusUnauthorized,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			store := test_sqlite.New()

			// Populate datastore with dummy users.
			for _, s := range tt.sessions {
				if err := store.InsertUser(screenjournal.User{
					Username: s.session.Username,
				}); err != nil {
					panic(err)
				}
			}

			if _, err := store.InsertMovie(makeCommentsTestData().movies.theWaterBoy); err != nil {
				panic(err)
			}

			if _, err := store.InsertReview(makeCommentsTestData().reviews.userBTheWaterBoy); err != nil {
				panic(err)
			}
			for _, comment := range tt.comments {
				if _, err := store.InsertComment(comment); err != nil {
					panic(err)
				}
			}

			authenticator := auth.New(store)
			sessionManager := newMockSessionManager(tt.sessions)
			s := handlers.New(authenticator, nilAnnouncer, &sessionManager, store, nilMetadataFinder)

			req, err := http.NewRequest("PUT", tt.route, strings.NewReader(tt.payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.AddCookie(&http.Cookie{
				Name:  mockSessionTokenName,
				Value: tt.sessionToken,
			})

			rec := httptest.NewRecorder()
			s.Router().ServeHTTP(rec, req)
			res := rec.Result()

			if got, want := res.StatusCode, tt.status; got != want {
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

func TestCommentsDelete(t *testing.T) {
	for _, tt := range []struct {
		description      string
		route            string
		sessionToken     string
		sessions         []mockSessionEntry
		comments         []screenjournal.ReviewComment
		status           int
		expectedComments []screenjournal.ReviewComment
	}{
		{
			description:  "allows a user to delete their own comment",
			route:        "/api/comments/1",
			sessionToken: makeCommentsTestData().sessions.userA.token,
			sessions: []mockSessionEntry{
				makeCommentsTestData().sessions.userA,
			},
			comments: []screenjournal.ReviewComment{
				{
					ID:          screenjournal.CommentID(1),
					Owner:       makeCommentsTestData().sessions.userA.session.Username,
					CommentText: screenjournal.CommentText("Good insights!"),
					Review:      makeCommentsTestData().reviews.userBTheWaterBoy,
				},
			},
			status:           http.StatusNoContent,
			expectedComments: []screenjournal.ReviewComment{},
		},
		{
			description:  "prevents a user from deleting a non-existent comment",
			route:        "/api/comments/999",
			sessionToken: makeCommentsTestData().sessions.userA.token,
			sessions: []mockSessionEntry{
				makeCommentsTestData().sessions.userA,
			},
			comments: []screenjournal.ReviewComment{
				{
					ID:          screenjournal.CommentID(1),
					Owner:       makeCommentsTestData().sessions.userA.session.Username,
					CommentText: screenjournal.CommentText("Good insights!"),
					Review:      makeCommentsTestData().reviews.userBTheWaterBoy,
				},
			},
			status: http.StatusNotFound,
		},
		{
			description:  "prevents a user from deleting an invalid comment ID",
			route:        "/api/comments/banana",
			sessionToken: makeCommentsTestData().sessions.userA.token,
			sessions: []mockSessionEntry{
				makeCommentsTestData().sessions.userA,
			},
			comments: []screenjournal.ReviewComment{
				{
					ID:          screenjournal.CommentID(1),
					Owner:       makeCommentsTestData().sessions.userA.session.Username,
					CommentText: screenjournal.CommentText("Good insights!"),
					Review:      makeCommentsTestData().reviews.userBTheWaterBoy,
				},
			},
			status: http.StatusBadRequest,
		},
		{
			description:  "prevents a user from deleting someone else's comment",
			route:        "/api/comments/1",
			sessionToken: makeCommentsTestData().sessions.userB.token,
			sessions: []mockSessionEntry{
				makeCommentsTestData().sessions.userB,
			},
			comments: []screenjournal.ReviewComment{
				{
					ID:          screenjournal.CommentID(1),
					Owner:       makeCommentsTestData().sessions.userA.session.Username,
					CommentText: screenjournal.CommentText("Good insights!"),
					Review:      makeCommentsTestData().reviews.userBTheWaterBoy,
				},
			},
			status: http.StatusForbidden,
		},
		{
			description:  "prevents an unauthenticated user from deleting any comment",
			route:        "/api/comments/1",
			sessionToken: "",
			sessions: []mockSessionEntry{
				makeCommentsTestData().sessions.userB,
			},
			comments: []screenjournal.ReviewComment{
				{
					ID:          screenjournal.CommentID(1),
					Owner:       makeCommentsTestData().sessions.userA.session.Username,
					CommentText: screenjournal.CommentText("Good insights!"),
					Review:      makeCommentsTestData().reviews.userBTheWaterBoy,
				},
			},
			status: http.StatusUnauthorized,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			store := test_sqlite.New()

			for _, s := range tt.sessions {
				if err := store.InsertUser(screenjournal.User{
					Username: s.session.Username,
				}); err != nil {
					panic(err)
				}
			}
			if _, err := store.InsertMovie(makeCommentsTestData().movies.theWaterBoy); err != nil {
				panic(err)
			}
			if _, err := store.InsertReview(makeCommentsTestData().reviews.userBTheWaterBoy); err != nil {
				panic(err)
			}
			for _, comment := range tt.comments {
				if _, err := store.InsertComment(comment); err != nil {
					panic(err)
				}
			}

			authenticator := auth.New(store)
			sessionManager := newMockSessionManager(tt.sessions)
			s := handlers.New(authenticator, nilAnnouncer, &sessionManager, store, nilMetadataFinder)

			req, err := http.NewRequest("DELETE", tt.route, strings.NewReader(""))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Add("Content-Type", "text/json")
			req.AddCookie(&http.Cookie{
				Name:  mockSessionTokenName,
				Value: tt.sessionToken,
			})

			rec := httptest.NewRecorder()
			s.Router().ServeHTTP(rec, req)
			res := rec.Result()

			if got, want := res.StatusCode, tt.status; got != want {
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
	d, err := tmdb.ParseReleaseDate(s)
	if err != nil {
		log.Fatalf("failed to parse release date: %s", s)
	}
	return d
}

func reviewCommentsEqual(a, b []screenjournal.ReviewComment) bool {
	if len(a) != len(b) {
		return false
	}

	// Clear timestamps from the comments because we don't want to compare based
	// on times.
	for i := range a {
		a[i].Created, b[i].Created = time.Time{}, time.Time{}
		a[i].Modified, b[i].Modified = time.Time{}, time.Time{}
		a[i].Review.Created, b[i].Review.Created = time.Time{}, time.Time{}
		a[i].Review.Modified, b[i].Review.Modified = time.Time{}, time.Time{}
	}

	return reflect.DeepEqual(a, b)
}

func clearUnpredictableCommentProperties(c *screenjournal.ReviewComment) {
	c.Created = time.Time{}
	c.Modified = time.Time{}
	clearUnpredictableReviewProperties(&c.Review)
}
