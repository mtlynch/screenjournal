package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

type reactionsTestData struct {
	sessions struct {
		userA mockSessionEntry
		userB mockSessionEntry
		admin mockSessionEntry
	}
	movies struct {
		theWaterBoy screenjournal.Movie
	}
	reviews struct {
		userBTheWaterBoy screenjournal.Review
	}
}

func makeReactionsTestData() reactionsTestData {
	td := reactionsTestData{}
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
	td.sessions.admin = mockSessionEntry{
		token: "admin789",
		session: sessions.Session{
			Username: screenjournal.Username("admin"),
			IsAdmin:  true,
		},
	}
	td.movies.theWaterBoy = screenjournal.Movie{
		ID:          screenjournal.MovieID(1),
		Title:       screenjournal.MediaTitle("The Waterboy"),
		ReleaseDate: mustParseReleaseDate("1998-11-06"),
		ImdbID:      screenjournal.ImdbID("tt1234556"),
	}
	td.reviews.userBTheWaterBoy = screenjournal.Review{
		ID:      screenjournal.ReviewID(1),
		Owner:   td.sessions.userB.session.Username,
		Rating:  screenjournal.NewRating(5),
		Movie:   td.movies.theWaterBoy,
		Watched: mustParseWatchDate("2020-10-05"),
		Blurb:   screenjournal.Blurb("I love water!"),
	}
	return td
}

func TestReactionsPost(t *testing.T) {
	for _, tt := range []struct {
		description       string
		payload           string
		sessionToken      string
		sessions          []mockSessionEntry
		movies            []screenjournal.Movie
		reviews           []screenjournal.Review
		status            int
		expectedReactions []screenjournal.ReviewReaction
	}{
		{
			description:  "allows user to add a reaction to an existing review",
			payload:      "review-id=1&emoji=👍",
			sessionToken: makeReactionsTestData().sessions.userA.token,
			sessions: []mockSessionEntry{
				makeReactionsTestData().sessions.userA,
				makeReactionsTestData().sessions.userB,
			},
			movies: []screenjournal.Movie{
				makeReactionsTestData().movies.theWaterBoy,
			},
			reviews: []screenjournal.Review{
				makeReactionsTestData().reviews.userBTheWaterBoy,
			},
			status: http.StatusOK,
			expectedReactions: []screenjournal.ReviewReaction{
				{
					ID:    screenjournal.ReactionID(1),
					Owner: makeReactionsTestData().sessions.userA.session.Username,
					Emoji: screenjournal.NewReactionEmoji("👍"),
					Review: screenjournal.Review{
						ID: screenjournal.ReviewID(1),
					},
				},
			},
		},
		{
			description:  "allows user to add pancakes emoji reaction",
			payload:      "review-id=1&emoji=🥞",
			sessionToken: makeReactionsTestData().sessions.userA.token,
			sessions: []mockSessionEntry{
				makeReactionsTestData().sessions.userA,
				makeReactionsTestData().sessions.userB,
			},
			movies: []screenjournal.Movie{
				makeReactionsTestData().movies.theWaterBoy,
			},
			reviews: []screenjournal.Review{
				makeReactionsTestData().reviews.userBTheWaterBoy,
			},
			status: http.StatusOK,
			expectedReactions: []screenjournal.ReviewReaction{
				{
					ID:    screenjournal.ReactionID(1),
					Owner: makeReactionsTestData().sessions.userA.session.Username,
					Emoji: screenjournal.NewReactionEmoji("🥞"),
					Review: screenjournal.Review{
						ID: screenjournal.ReviewID(1),
					},
				},
			},
		},
		{
			description:  "rejects request with invalid emoji",
			payload:      "review-id=1&emoji=❤️",
			sessionToken: makeReactionsTestData().sessions.userA.token,
			sessions: []mockSessionEntry{
				makeReactionsTestData().sessions.userA,
				makeReactionsTestData().sessions.userB,
			},
			movies: []screenjournal.Movie{
				makeReactionsTestData().movies.theWaterBoy,
			},
			reviews: []screenjournal.Review{
				makeReactionsTestData().reviews.userBTheWaterBoy,
			},
			status: http.StatusBadRequest,
		},
		{
			description:  "rejects request with invalid review ID",
			payload:      "review-id=0&emoji=👍",
			sessionToken: makeReactionsTestData().sessions.userA.token,
			sessions: []mockSessionEntry{
				makeReactionsTestData().sessions.userA,
				makeReactionsTestData().sessions.userB,
			},
			movies: []screenjournal.Movie{
				makeReactionsTestData().movies.theWaterBoy,
			},
			reviews: []screenjournal.Review{
				makeReactionsTestData().reviews.userBTheWaterBoy,
			},
			status: http.StatusBadRequest,
		},
		{
			description:  "returns 404 if user attempts to react to non-existent review",
			payload:      "review-id=999&emoji=👍",
			sessionToken: makeReactionsTestData().sessions.userA.token,
			sessions: []mockSessionEntry{
				makeReactionsTestData().sessions.userA,
				makeReactionsTestData().sessions.userB,
			},
			movies: []screenjournal.Movie{
				makeReactionsTestData().movies.theWaterBoy,
			},
			reviews: []screenjournal.Review{
				makeReactionsTestData().reviews.userBTheWaterBoy,
			},
			status:            http.StatusNotFound,
			expectedReactions: []screenjournal.ReviewReaction{},
		},
		{
			description:  "rejects request if user is not authenticated",
			payload:      "review-id=1&emoji=👍",
			sessionToken: "dummy-invalid-token",
			sessions: []mockSessionEntry{
				makeReactionsTestData().sessions.userA,
				makeReactionsTestData().sessions.userB,
			},
			status: http.StatusUnauthorized,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()

			for _, s := range tt.sessions {
				mockUser := screenjournal.User{
					Username:     s.session.Username,
					Email:        screenjournal.Email(s.session.Username.String() + "@example.com"),
					PasswordHash: screenjournal.PasswordHash("dummy-password-hash"),
				}
				if err := dataStore.InsertUser(mockUser); err != nil {
					t.Fatalf("failed to insert mock user: %+v: %v", mockUser, err)
				}
			}
			for _, movie := range tt.movies {
				if _, err := dataStore.InsertMovie(movie); err != nil {
					t.Fatalf("failed to insert mock movie: %+v: %v", movie, err)
				}
			}
			for _, review := range tt.reviews {
				if _, err := dataStore.InsertReview(review); err != nil {
					t.Fatalf("failed to insert mock review: %+v: %v", review, err)
				}
			}

			authenticator := auth.New(dataStore)
			sessionManager := newMockSessionManager(tt.sessions)
			s := handlers.New(authenticator, nilAnnouncer, &sessionManager, dataStore, nilMetadataFinder)

			req, err := http.NewRequest("POST", "/reactions", strings.NewReader(tt.payload))
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

			reactions, err := dataStore.ReadReactions(screenjournal.ReviewID(1))
			if err != nil {
				t.Fatalf("failed to read reactions from datastore: %v", err)
			}
			if got, want := reactions, tt.expectedReactions; !reviewReactionsEqual(got, want) {
				t.Errorf("reactions=%+v, want=%+v", got, want)
			}
		})
	}
}

func TestReactionsDelete(t *testing.T) {
	for _, tt := range []struct {
		description       string
		route             string
		sessionToken      string
		sessions          []mockSessionEntry
		movies            []screenjournal.Movie
		reviews           []screenjournal.Review
		reactions         []screenjournal.ReviewReaction
		status            int
		expectedReactions []screenjournal.ReviewReaction
	}{
		{
			description:  "allows a user to delete their own reaction",
			route:        "/reactions/1",
			sessionToken: makeReactionsTestData().sessions.userA.token,
			sessions: []mockSessionEntry{
				makeReactionsTestData().sessions.userA,
				makeReactionsTestData().sessions.userB,
			},
			movies: []screenjournal.Movie{
				makeReactionsTestData().movies.theWaterBoy,
			},
			reviews: []screenjournal.Review{
				makeReactionsTestData().reviews.userBTheWaterBoy,
			},
			reactions: []screenjournal.ReviewReaction{
				{
					ID:    screenjournal.ReactionID(1),
					Owner: makeReactionsTestData().sessions.userA.session.Username,
					Emoji: screenjournal.NewReactionEmoji("👍"),
					Review: screenjournal.Review{
						ID: screenjournal.ReviewID(1),
					},
				},
			},
			status:            http.StatusNoContent,
			expectedReactions: []screenjournal.ReviewReaction{},
		},
		{
			description:  "allows an admin to delete another user's reaction",
			route:        "/reactions/1",
			sessionToken: makeReactionsTestData().sessions.admin.token,
			sessions: []mockSessionEntry{
				makeReactionsTestData().sessions.userA,
				makeReactionsTestData().sessions.userB,
				makeReactionsTestData().sessions.admin,
			},
			movies: []screenjournal.Movie{
				makeReactionsTestData().movies.theWaterBoy,
			},
			reviews: []screenjournal.Review{
				makeReactionsTestData().reviews.userBTheWaterBoy,
			},
			reactions: []screenjournal.ReviewReaction{
				{
					ID:    screenjournal.ReactionID(1),
					Owner: makeReactionsTestData().sessions.userA.session.Username,
					Emoji: screenjournal.NewReactionEmoji("👍"),
					Review: screenjournal.Review{
						ID: screenjournal.ReviewID(1),
					},
				},
			},
			status:            http.StatusNoContent,
			expectedReactions: []screenjournal.ReviewReaction{},
		},
		{
			description:  "prevents a non-admin user from deleting another user's reaction",
			route:        "/reactions/1",
			sessionToken: makeReactionsTestData().sessions.userB.token,
			sessions: []mockSessionEntry{
				makeReactionsTestData().sessions.userA,
				makeReactionsTestData().sessions.userB,
			},
			movies: []screenjournal.Movie{
				makeReactionsTestData().movies.theWaterBoy,
			},
			reviews: []screenjournal.Review{
				makeReactionsTestData().reviews.userBTheWaterBoy,
			},
			reactions: []screenjournal.ReviewReaction{
				{
					ID:    screenjournal.ReactionID(1),
					Owner: makeReactionsTestData().sessions.userA.session.Username,
					Emoji: screenjournal.NewReactionEmoji("👍"),
					Review: screenjournal.Review{
						ID: screenjournal.ReviewID(1),
					},
				},
			},
			status: http.StatusForbidden,
		},
		{
			description:  "prevents deleting a non-existent reaction",
			route:        "/reactions/999",
			sessionToken: makeReactionsTestData().sessions.userA.token,
			sessions: []mockSessionEntry{
				makeReactionsTestData().sessions.userA,
				makeReactionsTestData().sessions.userB,
			},
			movies: []screenjournal.Movie{
				makeReactionsTestData().movies.theWaterBoy,
			},
			reviews: []screenjournal.Review{
				makeReactionsTestData().reviews.userBTheWaterBoy,
			},
			reactions: []screenjournal.ReviewReaction{
				{
					ID:    screenjournal.ReactionID(1),
					Owner: makeReactionsTestData().sessions.userA.session.Username,
					Emoji: screenjournal.NewReactionEmoji("👍"),
					Review: screenjournal.Review{
						ID: screenjournal.ReviewID(1),
					},
				},
			},
			status: http.StatusNotFound,
		},
		{
			description:  "prevents deleting with invalid reaction ID",
			route:        "/reactions/banana",
			sessionToken: makeReactionsTestData().sessions.userA.token,
			sessions: []mockSessionEntry{
				makeReactionsTestData().sessions.userA,
				makeReactionsTestData().sessions.userB,
			},
			movies: []screenjournal.Movie{
				makeReactionsTestData().movies.theWaterBoy,
			},
			reviews: []screenjournal.Review{
				makeReactionsTestData().reviews.userBTheWaterBoy,
			},
			reactions: []screenjournal.ReviewReaction{},
			status:    http.StatusBadRequest,
		},
		{
			description:  "prevents unauthenticated user from deleting any reaction",
			route:        "/reactions/1",
			sessionToken: "",
			sessions: []mockSessionEntry{
				makeReactionsTestData().sessions.userA,
				makeReactionsTestData().sessions.userB,
			},
			movies: []screenjournal.Movie{
				makeReactionsTestData().movies.theWaterBoy,
			},
			reviews: []screenjournal.Review{
				makeReactionsTestData().reviews.userBTheWaterBoy,
			},
			reactions: []screenjournal.ReviewReaction{
				{
					ID:    screenjournal.ReactionID(1),
					Owner: makeReactionsTestData().sessions.userA.session.Username,
					Emoji: screenjournal.NewReactionEmoji("👍"),
					Review: screenjournal.Review{
						ID: screenjournal.ReviewID(1),
					},
				},
			},
			status: http.StatusUnauthorized,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()

			for _, s := range tt.sessions {
				mockUser := screenjournal.User{
					Username:     s.session.Username,
					Email:        screenjournal.Email(s.session.Username + "@example.com"),
					PasswordHash: screenjournal.PasswordHash("dummy-password-hash"),
				}
				if err := dataStore.InsertUser(mockUser); err != nil {
					t.Fatalf("failed to create mock user %+v: %v", mockUser, err)
				}
			}
			for _, movie := range tt.movies {
				if _, err := dataStore.InsertMovie(movie); err != nil {
					t.Fatalf("failed to insert mock movie: %+v: %v", movie, err)
				}
			}
			for _, review := range tt.reviews {
				if _, err := dataStore.InsertReview(review); err != nil {
					t.Fatalf("failed to insert mock review: %+v: %v", review, err)
				}
			}
			for _, reaction := range tt.reactions {
				if _, err := dataStore.InsertReaction(reaction); err != nil {
					t.Fatalf("failed to insert mock reaction: %+v: %v", reaction, err)
				}
			}

			authenticator := auth.New(dataStore)
			sessionManager := newMockSessionManager(tt.sessions)
			s := handlers.New(authenticator, nilAnnouncer, &sessionManager, dataStore, nilMetadataFinder)

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

			if tt.status != http.StatusNoContent {
				return
			}

			reactions, err := dataStore.ReadReactions(screenjournal.ReviewID(1))
			if err != nil {
				t.Fatalf("failed to read reactions from datastore: %v", err)
			}
			if got, want := reactions, tt.expectedReactions; !reviewReactionsEqual(got, want) {
				t.Errorf("reactions=%+v, want=%+v", got, want)
			}
		})
	}
}

func reviewReactionsEqual(a, b []screenjournal.ReviewReaction) bool {
	if len(a) != len(b) {
		return false
	}

	// Clear timestamps from the reactions because we don't want to compare based
	// on times.
	for i := range a {
		a[i].Created = time.Time{}
		b[i].Created = time.Time{}
		// Clear the full review object, just compare IDs.
		a[i].Review = screenjournal.Review{ID: a[i].Review.ID}
		b[i].Review = screenjournal.Review{ID: b[i].Review.ID}
	}

	return reflect.DeepEqual(a, b)
}
