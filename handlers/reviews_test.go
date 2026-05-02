package handlers_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-test/deep"

	simple_sessions "codeberg.org/mtlynch/simpleauth/v3/sessions"

	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/metadata"
	"github.com/mtlynch/screenjournal/v2/random"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

type contextKey struct {
	name string
}

var contextKeySession = &contextKey{"session"}

var nilAuthenticator auth.Authenticator

type mockAnnouncer struct {
	announcedReviews  []screenjournal.Review
	announcedComments []screenjournal.ReviewComment
}

func (a *mockAnnouncer) AnnounceNewReview(r screenjournal.Review) {
	a.announcedReviews = append(a.announcedReviews, r)
}

func (a *mockAnnouncer) AnnounceNewComment(rc screenjournal.ReviewComment) {
	a.announcedComments = append(a.announcedComments, rc)
}

type mockSessionEntry struct {
	token   string
	session mockSession
}

type mockSession struct {
	Username screenjournal.Username
}

type mockSessionManager struct {
	sessions map[string]mockSession
}

const mockSessionTokenName = "mock-session-token"

func newMockSessionManager(mockSessions []mockSessionEntry) mockSessionManager {
	sessions := make(map[string]mockSession, len(mockSessions))
	for _, ms := range mockSessions {
		sessions[ms.token] = ms.session
	}
	return mockSessionManager{
		sessions: sessions,
	}
}

func (sm *mockSessionManager) LogIn(ctx context.Context, w http.ResponseWriter, userID simple_sessions.UserID) error {
	_ = ctx
	token := random.String(10, []rune("abcdefghijklmnopqrstuvwxyz0123456789"))
	sm.sessions[token] = mockSession{
		Username: screenjournal.Username(userID.String()),
	}
	http.SetCookie(w, &http.Cookie{
		Name:  mockSessionTokenName,
		Value: token,
	})
	return nil
}

func (sm mockSessionManager) UserIDFromContext(ctx context.Context) (simple_sessions.UserID, error) {
	token, ok := ctx.Value(contextKeySession).(string)
	if !ok {
		return simple_sessions.UserID{}, simple_sessions.ErrNoSessionFound
	}
	session, err := sm.SessionFromToken(token)
	if err != nil {
		return simple_sessions.UserID{}, err
	}
	return simple_sessions.NewUserID(session.Username.String())
}

func (sm mockSessionManager) SessionFromToken(token string) (mockSession, error) {
	session, ok := sm.sessions[token]
	if !ok {
		return mockSession{}, fmt.Errorf("mock session manager: no session associated with token %s", token)
	}

	return session, nil
}

func (sm mockSessionManager) LogOut(context.Context, http.ResponseWriter) error {
	return nil
}

func (sm mockSessionManager) LoadUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token, err := r.Cookie(mockSessionTokenName); err == nil {
			r = r.WithContext(context.WithValue(r.Context(), contextKeySession, token.Value))
		}
		next.ServeHTTP(w, r)
	})
}

type mockUserStore interface {
	InsertUser(screenjournal.User) error
}

func insertMockUsers(t *testing.T, userStore mockUserStore, users []screenjournal.User) {
	t.Helper()
	for _, user := range users {
		if err := userStore.InsertUser(user); err != nil {
			t.Fatalf("failed to insert mock user %+v: %v", user, err)
		}
	}
}

func insertMockUsersForSessions(
	t *testing.T,
	userStore mockUserStore,
	sessions []mockSessionEntry,
	adminUsernames ...screenjournal.Username,
) {
	t.Helper()
	admins := make(map[screenjournal.Username]struct{}, len(adminUsernames))
	for _, username := range adminUsernames {
		admins[username] = struct{}{}
	}

	users := make([]screenjournal.User, 0, len(sessions))
	for _, s := range sessions {
		if _, ok := admins[s.session.Username]; ok {
			users = append(users, newMockAdminUser(s.session.Username))
			continue
		}
		users = append(users, newMockUser(s.session.Username))
	}
	insertMockUsers(t, userStore, users)
}

func newMockSessionEntry(token string, username screenjournal.Username) mockSessionEntry {
	return mockSessionEntry{
		token: token,
		session: mockSession{
			Username: username,
		},
	}
}

func newMockUser(username screenjournal.Username) screenjournal.User {
	return screenjournal.User{
		Username:     username,
		Email:        screenjournal.Email(username.String() + "@example.com"),
		PasswordHash: screenjournal.PasswordHash("dummy-password-hash"),
	}
}

func newMockAdminUser(username screenjournal.Username) screenjournal.User {
	user := newMockUser(username)
	user.IsAdmin = true
	return user
}

type mockMetadataFinder struct {
	movies  []screenjournal.Movie
	tvShows []screenjournal.TvShow
}

func (mf mockMetadataFinder) SearchMovies(query screenjournal.SearchQuery) ([]metadata.SearchResult, error) {
	matches := []metadata.SearchResult{}
	for _, v := range mf.movies {
		if strings.Contains(strings.ToLower(v.Title.String()), strings.ToLower(query.String())) {
			matches = append(matches, metadata.SearchResult{
				TmdbID:      v.TmdbID,
				Title:       v.Title,
				ReleaseDate: v.ReleaseDate,
				PosterPath:  v.PosterPath,
			})
		}
	}
	return matches, nil
}

func (mf mockMetadataFinder) SearchTvShows(query screenjournal.SearchQuery) ([]metadata.SearchResult, error) {
	matches := []metadata.SearchResult{}

	for _, v := range mf.tvShows {
		if strings.Contains(strings.ToLower(v.Title.String()), strings.ToLower(query.String())) {
			matches = append(matches, metadata.SearchResult{
				TmdbID:      v.TmdbID,
				Title:       v.Title,
				ReleaseDate: v.AirDate,
				PosterPath:  v.PosterPath,
			})
		}
	}
	return matches, nil
}

func (mf mockMetadataFinder) GetMovie(id screenjournal.TmdbID) (screenjournal.Movie, error) {
	for _, m := range mf.movies {
		if id.Equal(m.TmdbID) {
			return m, nil
		}
	}
	return screenjournal.Movie{}, fmt.Errorf("could not find movie with id %d in mock DB", id.Int32())
}

func (mf mockMetadataFinder) GetTvShow(id screenjournal.TmdbID) (screenjournal.TvShow, error) {
	for _, t := range mf.tvShows {
		if id.Equal(t.TmdbID) {
			return t, nil
		}
	}
	return screenjournal.TvShow{}, fmt.Errorf("could not find TV show with id %d in mock DB", id.Int32())
}

func NewMockMetadataFinder(movies []screenjournal.Movie, tvShows []screenjournal.TvShow) mockMetadataFinder {
	moviesCopy := make([]screenjournal.Movie, len(movies))
	copy(moviesCopy, movies)

	tvShowsCopy := make([]screenjournal.TvShow, len(tvShows))
	copy(tvShowsCopy, tvShows)

	return mockMetadataFinder{
		movies:  moviesCopy,
		tvShows: tvShowsCopy,
	}
}

func TestReviewsPost(t *testing.T) {
	for _, tt := range []struct {
		description     string
		payload         string
		sessionToken    string
		localMovies     []screenjournal.Movie
		remoteMovieInfo []screenjournal.Movie
		sessions        []mockSessionEntry
		expectedStatus  int
		expected        screenjournal.Review
	}{
		{
			description:  "valid request with all fields populated and movie information is in local DB",
			payload:      "media-type=movie&tmdb-id=38&rating=5&watch-date=2022-10-28&blurb=It's%20my%20favorite%20movie!",
			sessionToken: "abc123",
			localMovies: []screenjournal.Movie{
				{
					TmdbID:      screenjournal.TmdbID(38),
					ImdbID:      screenjournal.ImdbID("tt0338013"),
					Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
				},
			},
			sessions: []mockSessionEntry{
				newMockSessionEntry("abc123", screenjournal.Username("dummyadmin")),
			},
			expectedStatus: http.StatusSeeOther,
			expected: screenjournal.Review{
				Owner:   screenjournal.Username("dummyadmin"),
				Rating:  screenjournal.NewRating(5),
				Watched: mustParseWatchDate("2022-10-28"),
				Blurb:   screenjournal.Blurb("It's my favorite movie!"),
				Movie: screenjournal.Movie{
					ID:          screenjournal.MovieID(1),
					TmdbID:      screenjournal.TmdbID(38),
					ImdbID:      screenjournal.ImdbID("tt0338013"),
					Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
				},
				Comments: []screenjournal.ReviewComment{},
			},
		},
		{
			description:  "valid request without a blurb",
			payload:      "media-type=movie&tmdb-id=14577&rating=4&watch-date=2022-10-21&blurb=",
			sessionToken: "abc123",
			localMovies: []screenjournal.Movie{
				{
					ID:          screenjournal.MovieID(1),
					TmdbID:      screenjournal.TmdbID(14577),
					ImdbID:      screenjournal.ImdbID("tt0120654"),
					Title:       screenjournal.MediaTitle("Dirty Work"),
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("1998-06-12")),
				},
			},
			sessions: []mockSessionEntry{
				newMockSessionEntry("abc123", screenjournal.Username("dummyadmin")),
			},
			expectedStatus: http.StatusSeeOther,
			expected: screenjournal.Review{
				Owner:   screenjournal.Username("dummyadmin"),
				Rating:  screenjournal.NewRating(4),
				Watched: mustParseWatchDate("2022-10-21"),
				Blurb:   screenjournal.Blurb(""),
				Movie: screenjournal.Movie{
					ID:          screenjournal.MovieID(1),
					TmdbID:      screenjournal.TmdbID(14577),
					ImdbID:      screenjournal.ImdbID("tt0120654"),
					Title:       screenjournal.MediaTitle("Dirty Work"),
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("1998-06-12")),
				},
				Comments: []screenjournal.ReviewComment{},
			},
		},
		{
			description:  "valid request but we have to query metadata finder for movie info",
			payload:      "media-type=movie&tmdb-id=38&rating=5&watch-date=2022-10-28&blurb=It's%20my%20favorite%20movie!",
			sessionToken: "abc123",
			remoteMovieInfo: []screenjournal.Movie{
				{
					TmdbID:      screenjournal.TmdbID(38),
					ImdbID:      screenjournal.ImdbID("tt0338013"),
					Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
				},
			},
			sessions: []mockSessionEntry{
				newMockSessionEntry("abc123", screenjournal.Username("dummyadmin")),
			},
			expectedStatus: http.StatusSeeOther,
			expected: screenjournal.Review{
				Owner:   screenjournal.Username("dummyadmin"),
				Rating:  screenjournal.NewRating(5),
				Watched: mustParseWatchDate("2022-10-28"),
				Blurb:   screenjournal.Blurb("It's my favorite movie!"),
				Movie: screenjournal.Movie{
					ID:          screenjournal.MovieID(1),
					TmdbID:      screenjournal.TmdbID(38),
					ImdbID:      screenjournal.ImdbID("tt0338013"),
					Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
				},
				Comments: []screenjournal.ReviewComment{},
			},
		},
		{
			description:  "rejects empty string payload",
			payload:      "",
			sessionToken: "abc123",
			sessions: []mockSessionEntry{
				newMockSessionEntry("abc123", screenjournal.Username("userA")),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			description:  "rejects empty form fields",
			payload:      "tmdb-id=&rating=&watch-date=&blurb=",
			sessionToken: "abc123",
			sessions: []mockSessionEntry{
				newMockSessionEntry("abc123", screenjournal.Username("userA")),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			description:  "rejects invalid tmdb field (non-number)",
			payload:      "tmdb-id=banana&rating=5&watch-date=2022-10-28&blurb=It's%20my%20favorite%20movie!",
			sessionToken: "abc123",
			sessions: []mockSessionEntry{
				newMockSessionEntry("abc123", screenjournal.Username("userA")),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			description:  "rejects invalid rating field (non-number)",
			payload:      "tmdb-id=banana&rating=banana&watch-date=2022-10-28&blurb=It's%20my%20favorite%20movie!",
			sessionToken: "abc123",
			sessions: []mockSessionEntry{
				newMockSessionEntry("abc123", screenjournal.Username("userA")),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			description:  "accepts request with missing rating field",
			payload:      "media-type=movie&tmdb-id=38&watch-date=2022-10-28&blurb=It's%20my%20favorite%20movie!",
			sessionToken: "abc123",
			localMovies: []screenjournal.Movie{
				{
					TmdbID:      screenjournal.TmdbID(38),
					ImdbID:      screenjournal.ImdbID("tt0338013"),
					Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
				},
			},
			sessions: []mockSessionEntry{
				newMockSessionEntry("abc123", screenjournal.Username("userA")),
			},
			expectedStatus: http.StatusSeeOther,
			expected: screenjournal.Review{
				Owner:   screenjournal.Username("userA"),
				Rating:  screenjournal.Rating{},
				Watched: mustParseWatchDate("2022-10-28"),
				Blurb:   screenjournal.Blurb("It's my favorite movie!"),
				Movie: screenjournal.Movie{
					ID:          screenjournal.MovieID(1),
					TmdbID:      screenjournal.TmdbID(38),
					ImdbID:      screenjournal.ImdbID("tt0338013"),
					Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
				},
				Comments: []screenjournal.ReviewComment{},
			},
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()

			insertMockUsersForSessions(
				t,
				dataStore,
				tt.sessions,
				screenjournal.Username("dummyadmin"),
			)

			for _, movie := range tt.localMovies {
				if _, err := dataStore.InsertMovie(movie); err != nil {
					t.Fatalf("failed to insert mock movie %+v: %v", movie, err)
				}
			}

			announcer := mockAnnouncer{}
			sessionManager := newMockSessionManager(tt.sessions)
			s := handlers.New(handlers.ServerParams{
				Authenticator:  nilAuthenticator,
				Announcer:      &announcer,
				SessionManager: &sessionManager,
				Store:          dataStore,
				MetadataFinder: NewMockMetadataFinder(tt.remoteMovieInfo, nil),
			})

			req, err := http.NewRequest("POST", "/reviews", strings.NewReader(tt.payload))
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

			if got, want := res.StatusCode, tt.expectedStatus; got != want {
				t.Fatalf("httpStatus=%v, want=%v", got, want)
			}

			if tt.expectedStatus != http.StatusSeeOther {
				return
			}

			rr, err := dataStore.ReadReviews()
			if err != nil {
				t.Fatalf("failed to retrieve review from datastore: %v", err)
			}

			if got, want := len(rr), 1; got != want {
				t.Fatalf("reviewCountInStore=%d, want=%d", got, want)
			}

			clearUnpredictableReviewProperties(&rr[0])
			if got, want := rr[0], tt.expected; !reflect.DeepEqual(got, want) {
				t.Errorf("got=%#v, want=%#v, diff=%s", got, want, deep.Equal(got, want))
			}

			if got, want := len(announcer.announcedReviews), 1; got != want {
				t.Fatalf("reviewCountAnnounced=%d, want=%d", got, want)
			}

			clearUnpredictableReviewProperties(&announcer.announcedReviews[0])
			if got, want := announcer.announcedReviews[0], tt.expected; !reflect.DeepEqual(got, want) {
				t.Errorf("got=%#v, want=%#v, diff=%s", got, want, deep.Equal(got, want))
			}
		})
	}
}
func TestReviewsPut(t *testing.T) {
	for _, tt := range []struct {
		description    string
		localMovies    []screenjournal.Movie
		priorReviews   []screenjournal.Review
		sessions       []mockSessionEntry
		route          string
		payload        string
		sessionToken   string
		expectedStatus int
		expected       screenjournal.Review
	}{
		{
			description: "valid request with all fields populated",
			localMovies: []screenjournal.Movie{
				{
					TmdbID:      screenjournal.TmdbID(38),
					ImdbID:      screenjournal.ImdbID("tt0338013"),
					Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
				},
				{
					TmdbID:      screenjournal.TmdbID(14577),
					ImdbID:      screenjournal.ImdbID("tt0120654"),
					Title:       screenjournal.MediaTitle("Dirty Work"),
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("1998-06-12")),
				},
			},
			priorReviews: []screenjournal.Review{
				{
					Owner:   screenjournal.Username("userA"),
					Rating:  screenjournal.NewRating(5),
					Watched: mustParseWatchDate("2022-10-28"),
					Blurb:   screenjournal.Blurb("It's my favorite movie!"),
					Movie: screenjournal.Movie{
						ID:          screenjournal.MovieID(1),
						TmdbID:      screenjournal.TmdbID(38),
						ImdbID:      screenjournal.ImdbID("tt0338013"),
						Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
						ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
					},
				},
			},
			sessions: []mockSessionEntry{
				newMockSessionEntry("abc123", screenjournal.Username("userA")),
			},
			route:          "/reviews/1",
			payload:        "rating=4&watch-date=2022-10-30&blurb=It's%20a%20pretty%20good%20movie!",
			sessionToken:   "abc123",
			expectedStatus: http.StatusSeeOther,
			expected: screenjournal.Review{
				Owner:   screenjournal.Username("userA"),
				Rating:  screenjournal.NewRating(4),
				Watched: mustParseWatchDate("2022-10-30"),
				Blurb:   screenjournal.Blurb("It's a pretty good movie!"),
				Movie: screenjournal.Movie{
					ID:          screenjournal.MovieID(1),
					TmdbID:      screenjournal.TmdbID(38),
					ImdbID:      screenjournal.ImdbID("tt0338013"),
					Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
				},
				Comments: []screenjournal.ReviewComment{},
			},
		},
		{
			description: "valid request with an empty blurb",
			localMovies: []screenjournal.Movie{
				{
					TmdbID:      screenjournal.TmdbID(38),
					ImdbID:      screenjournal.ImdbID("tt0338013"),
					Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
				},
				{
					TmdbID:      screenjournal.TmdbID(14577),
					ImdbID:      screenjournal.ImdbID("tt0120654"),
					Title:       screenjournal.MediaTitle("Dirty Work"),
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("1998-06-12")),
				},
			},
			priorReviews: []screenjournal.Review{
				{
					Owner:   screenjournal.Username("userA"),
					Rating:  screenjournal.NewRating(4),
					Watched: mustParseWatchDate("2022-10-21"),
					Blurb:   screenjournal.Blurb("Love Norm McDonald!"),
					Movie: screenjournal.Movie{
						ID:          screenjournal.MovieID(2),
						TmdbID:      screenjournal.TmdbID(14577),
						ImdbID:      screenjournal.ImdbID("tt0120654"),
						Title:       screenjournal.MediaTitle("Dirty Work"),
						ReleaseDate: screenjournal.ReleaseDate(mustParseDate("1998-06-12")),
					},
				},
			},
			sessions: []mockSessionEntry{
				newMockSessionEntry("abc123", screenjournal.Username("userA")),
			},
			route:          "/reviews/1",
			payload:        "rating=3&watch-date=2022-10-28&blurb=",
			sessionToken:   "abc123",
			expectedStatus: http.StatusSeeOther,
			expected: screenjournal.Review{
				Owner:   screenjournal.Username("userA"),
				Rating:  screenjournal.NewRating(3),
				Watched: mustParseWatchDate("2022-10-28"),
				Blurb:   screenjournal.Blurb(""),
				Movie: screenjournal.Movie{
					ID:          screenjournal.MovieID(2),
					TmdbID:      screenjournal.TmdbID(14577),
					ImdbID:      screenjournal.ImdbID("tt0120654"),
					Title:       screenjournal.MediaTitle("Dirty Work"),
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("1998-06-12")),
				},
				Comments: []screenjournal.ReviewComment{},
			},
		},
		{
			description: "rejects request with review ID of zero",
			localMovies: []screenjournal.Movie{
				{
					TmdbID:      screenjournal.TmdbID(38),
					ImdbID:      screenjournal.ImdbID("tt0338013"),
					Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
				},
			},
			priorReviews: []screenjournal.Review{
				{
					ID:      screenjournal.ReviewID(1),
					Owner:   screenjournal.Username("userA"),
					Rating:  screenjournal.NewRating(5),
					Watched: mustParseWatchDate("2022-10-28"),
					Blurb:   screenjournal.Blurb("It's my favorite movie!"),
					Movie: screenjournal.Movie{
						ID:          screenjournal.MovieID(1),
						TmdbID:      screenjournal.TmdbID(38),
						ImdbID:      screenjournal.ImdbID("tt0338013"),
						Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
						ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
					},
				},
			},
			sessions: []mockSessionEntry{
				newMockSessionEntry("abc123", screenjournal.Username("userA")),
			},
			route:          "/reviews/0",
			payload:        "rating=4&watch-date=2022-10-30&blurb=It's%20a%20pretty%20good%20movie!",
			sessionToken:   "abc123",
			expectedStatus: http.StatusBadRequest,
		},
		{
			description: "rejects request with non-existent review ID",
			localMovies: []screenjournal.Movie{
				{
					TmdbID:      screenjournal.TmdbID(38),
					ImdbID:      screenjournal.ImdbID("tt0338013"),
					Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
				},
			},
			priorReviews: []screenjournal.Review{
				{
					Owner:   screenjournal.Username("userA"),
					Rating:  screenjournal.NewRating(5),
					Watched: mustParseWatchDate("2022-10-28"),
					Blurb:   screenjournal.Blurb("It's my favorite movie!"),
					Movie: screenjournal.Movie{
						ID:          screenjournal.MovieID(1),
						TmdbID:      screenjournal.TmdbID(38),
						ImdbID:      screenjournal.ImdbID("tt0338013"),
						Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
						ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
					},
				},
			},
			sessions: []mockSessionEntry{
				newMockSessionEntry("abc123", screenjournal.Username("userA")),
			},
			route:          "/reviews/9876",
			payload:        "rating=4&watch-date=2022-10-30&blurb=It's%20a%20pretty%20good%20movie!",
			sessionToken:   "abc123",
			expectedStatus: http.StatusNotFound,
		},
		{
			description: "accepts request with missing rating field",
			localMovies: []screenjournal.Movie{
				{
					TmdbID:      screenjournal.TmdbID(38),
					ImdbID:      screenjournal.ImdbID("tt0338013"),
					Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
				},
			},
			priorReviews: []screenjournal.Review{
				{
					ID:      screenjournal.ReviewID(1),
					Owner:   screenjournal.Username("userA"),
					Rating:  screenjournal.NewRating(5),
					Watched: mustParseWatchDate("2022-10-28"),
					Blurb:   screenjournal.Blurb("It's my favorite movie!"),
					Movie: screenjournal.Movie{
						ID:          screenjournal.MovieID(1),
						TmdbID:      screenjournal.TmdbID(38),
						ImdbID:      screenjournal.ImdbID("tt0338013"),
						Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
						ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
					},
				},
			},
			sessions: []mockSessionEntry{
				newMockSessionEntry("abc123", screenjournal.Username("userA")),
			},
			route:          "/reviews/1",
			payload:        "watch-date=2022-10-30&blurb=It's%20a%20pretty%20good%20movie!",
			sessionToken:   "abc123",
			expectedStatus: http.StatusSeeOther,
			expected: screenjournal.Review{
				Owner:   screenjournal.Username("userA"),
				Rating:  screenjournal.Rating{},
				Watched: mustParseWatchDate("2022-10-30"),
				Blurb:   screenjournal.Blurb("It's a pretty good movie!"),
				Movie: screenjournal.Movie{
					ID:          screenjournal.MovieID(1),
					TmdbID:      screenjournal.TmdbID(38),
					ImdbID:      screenjournal.ImdbID("tt0338013"),
					Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
				},
				Comments: []screenjournal.ReviewComment{},
			},
		},
		{
			description: "rejects request with missing watched field",
			localMovies: []screenjournal.Movie{
				{
					TmdbID:      screenjournal.TmdbID(38),
					ImdbID:      screenjournal.ImdbID("tt0338013"),
					Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
				},
			},
			priorReviews: []screenjournal.Review{
				{
					ID:      screenjournal.ReviewID(1),
					Owner:   screenjournal.Username("userA"),
					Rating:  screenjournal.NewRating(5),
					Watched: mustParseWatchDate("2022-10-28"),
					Blurb:   screenjournal.Blurb("It's my favorite movie!"),
					Movie: screenjournal.Movie{
						ID:          screenjournal.MovieID(1),
						TmdbID:      screenjournal.TmdbID(38),
						ImdbID:      screenjournal.ImdbID("tt0338013"),
						Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
						ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
					},
				},
			},
			sessions: []mockSessionEntry{
				newMockSessionEntry("abc123", screenjournal.Username("userA")),
			},
			route:          "/reviews/1",
			payload:        "rating=4&blurb=It's%20a%20pretty%20good%20movie!",
			sessionToken:   "abc123",
			expectedStatus: http.StatusBadRequest,
		},
		{
			description: "rejects request with script tag in blurb field",
			localMovies: []screenjournal.Movie{
				{
					TmdbID:      screenjournal.TmdbID(38),
					ImdbID:      screenjournal.ImdbID("tt0338013"),
					Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
				},
			},
			priorReviews: []screenjournal.Review{
				{
					ID:      screenjournal.ReviewID(1),
					Owner:   screenjournal.Username("userA"),
					Rating:  screenjournal.NewRating(5),
					Watched: mustParseWatchDate("2022-10-28"),
					Blurb:   screenjournal.Blurb("It's my favorite movie!"),
					Movie: screenjournal.Movie{
						ID:          screenjournal.MovieID(1),
						TmdbID:      screenjournal.TmdbID(38),
						ImdbID:      screenjournal.ImdbID("tt0338013"),
						Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
						ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
					},
				},
			},
			sessions: []mockSessionEntry{
				newMockSessionEntry("abc123", screenjournal.Username("userA")),
			},
			route:          "/reviews/1",
			payload:        "rating=4&watch-date=2022-10-30&blurb=Nothing%20evil%20going%20on%20here...%3Cscript%3Ealert(1)%3C%2Fscript%3E",
			sessionToken:   "abc123",
			expectedStatus: http.StatusBadRequest,
		},
		{
			description: "prevents a user from overwriting another user's review",
			localMovies: []screenjournal.Movie{
				{
					TmdbID:      screenjournal.TmdbID(38),
					ImdbID:      screenjournal.ImdbID("tt0338013"),
					Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
				},
			},
			priorReviews: []screenjournal.Review{
				{
					ID:      screenjournal.ReviewID(1),
					Owner:   screenjournal.Username("userA"),
					Rating:  screenjournal.NewRating(5),
					Watched: mustParseWatchDate("2022-10-28"),
					Blurb:   screenjournal.Blurb("It's my favorite movie!"),
					Movie: screenjournal.Movie{
						ID:          screenjournal.MovieID(1),
						TmdbID:      screenjournal.TmdbID(38),
						ImdbID:      screenjournal.ImdbID("tt0338013"),
						Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
						ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
					},
				},
			},
			sessions: []mockSessionEntry{
				newMockSessionEntry("abc123", screenjournal.Username("userA")),
				newMockSessionEntry("def456", screenjournal.Username("userB")),
			},
			route:          "/reviews/1",
			payload:        "rating=4&watch-date=2022-10-30&blurb=I'm%20overwriting%20userA's%20review!",
			sessionToken:   "def456",
			expectedStatus: http.StatusForbidden,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()

			insertMockUsersForSessions(t, dataStore, tt.sessions)

			for _, movie := range tt.localMovies {
				if _, err := dataStore.InsertMovie(movie); err != nil {
					t.Fatalf("failed to insert mock movie %+v: %v", movie, err)
				}
			}

			for _, r := range tt.priorReviews {
				if _, err := dataStore.InsertReview(r); err != nil {
					t.Fatalf("failed to insert mock review %+v: %v", r, err)
				}
			}

			sessionManager := newMockSessionManager(tt.sessions)
			s := handlers.New(handlers.ServerParams{
				Authenticator:  nilAuthenticator,
				SessionManager: &sessionManager,
				Store:          dataStore,
				MetadataFinder: mockMetadataFinder{},
			})

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

			if got, want := res.StatusCode, tt.expectedStatus; got != want {
				t.Fatalf("handler returned wrong status code: got %v want %v", got, want)
			}

			if tt.expectedStatus != http.StatusSeeOther {
				return
			}

			rr, err := dataStore.ReadReviews()
			if err != nil {
				t.Fatalf("failed to retrieve review from datastore: %v", err)
			}

			if got, want := len(rr), 1; got != want {
				t.Fatalf("unexpected review count: got %v, want %v", got, want)
			}

			clearUnpredictableReviewProperties(&rr[0])
			if got, want := rr[0], tt.expected; !reflect.DeepEqual(got, want) {
				t.Errorf("unexpected reviews, got=%+v, want=%+v, diff=%s", got, want, deep.Equal(got, want))
			}
		})
	}
}

func clearUnpredictableReviewProperties(r *screenjournal.Review) {
	r.ID = screenjournal.ReviewID(0)
	r.Created = time.Time{}
	r.Modified = time.Time{}
}

func mustParseWatchDate(s string) screenjournal.WatchDate {
	wd, err := parse.WatchDate(s)
	if err != nil {
		log.Fatalf("failed to parse watch date: %s", s)
	}
	return wd
}

func mustParseDate(s string) time.Time {
	d, err := time.Parse(time.DateOnly, s)
	if err != nil {
		log.Fatalf("failed to parse watch date: %s", s)
	}
	return d
}
