package handlers_test

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
	"github.com/mtlynch/screenjournal/v2/metadata"
	"github.com/mtlynch/screenjournal/v2/random"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

type contextKey struct {
	name string
}

var contextKeySession = &contextKey{"session"}

var nilAnnouncer handlers.Announcer

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
	session sessions.Session
}

type mockSessionManager struct {
	sessions map[string]sessions.Session
}

const mockSessionTokenName = "mock-session-token"

func newMockSessionManager(mockSessions []mockSessionEntry) mockSessionManager {
	sessions := make(map[string]sessions.Session, len(mockSessions))
	for _, ms := range mockSessions {
		sessions[ms.token] = ms.session
	}
	return mockSessionManager{
		sessions: sessions,
	}
}

func (sm *mockSessionManager) CreateSession(w http.ResponseWriter, ctx context.Context, username screenjournal.Username, isAdmin bool) error {
	token := random.String(10, []rune("abcdefghijklmnopqrstuvwxyz0123456789"))
	sm.sessions[token] = sessions.Session{
		Username: username,
		IsAdmin:  isAdmin,
	}
	http.SetCookie(w, &http.Cookie{
		Name:  mockSessionTokenName,
		Value: token,
	})
	return nil
}

func (sm mockSessionManager) SessionFromContext(ctx context.Context) (sessions.Session, error) {
	token, ok := ctx.Value(contextKeySession).(string)
	if !ok {
		return sessions.Session{}, errors.New("dummy no session in context")
	}
	return sm.SessionFromToken(token)
}

func (sm mockSessionManager) SessionFromToken(token string) (sessions.Session, error) {
	session, ok := sm.sessions[token]
	if !ok {
		return sessions.Session{}, fmt.Errorf("mock session manager: no session associated with token %s", token)
	}

	return session, nil
}

func (sm mockSessionManager) EndSession(context.Context, http.ResponseWriter) {}

func (sm mockSessionManager) WrapRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token, err := r.Cookie(mockSessionTokenName); err == nil {
			r = r.WithContext(context.WithValue(r.Context(), contextKeySession, token.Value))
		}
		next.ServeHTTP(w, r)
	})
}

type mockMetadataFinder struct {
	movies  []metadata.MovieInfo
	tvShows []metadata.TvShowInfo
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
				ReleaseDate: v.ReleaseDate,
				PosterPath:  v.PosterPath,
			})
		}
	}
	return matches, nil
}

func (mf mockMetadataFinder) GetMovieInfo(id screenjournal.TmdbID) (metadata.MovieInfo, error) {
	for _, m := range mf.movies {
		if id.Equal(m.TmdbID) {
			return m, nil
		}
	}
	return metadata.MovieInfo{}, fmt.Errorf("could not find movie with id %d in mock DB", id.Int32())
}

func (mf mockMetadataFinder) GetTvShowInfo(id screenjournal.TmdbID) (metadata.TvShowInfo, error) {
	for _, t := range mf.tvShows {
		if id.Equal(t.TmdbID) {
			return t, nil
		}
	}
	return metadata.TvShowInfo{}, fmt.Errorf("could not find TV show with id %d in mock DB", id.Int32())
}

func NewMockMetadataFinder(movies []metadata.MovieInfo, tvShows []metadata.TvShowInfo) mockMetadataFinder {
	moviesCopy := make([]metadata.MovieInfo, len(movies))
	copy(moviesCopy, movies)

	tvShowsCopy := make([]metadata.TvShowInfo, len(tvShows))
	copy(tvShowsCopy, tvShows)

	return mockMetadataFinder{
		movies:  moviesCopy,
		tvShows: tvShowsCopy,
	}
}

func TestReviewsPostAcceptsValidRequest(t *testing.T) {
	for _, tt := range []struct {
		description     string
		payload         string
		sessionToken    string
		localMovies     []screenjournal.Movie
		remoteMovieInfo []metadata.MovieInfo
		sessions        []mockSessionEntry
		expected        screenjournal.Review
	}{
		{
			description:  "valid request with all fields populated and movie information is in local DB",
			payload:      "tmdb-id=38&rating=5&watch-date=2022-10-28&blurb=It's%20my%20favorite%20movie!",
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
				{
					token: "abc123",
					session: sessions.Session{
						Username: screenjournal.Username("dummyadmin"),
						IsAdmin:  true,
					},
				},
			},
			expected: screenjournal.Review{
				Owner:   screenjournal.Username("dummyadmin"),
				Rating:  screenjournal.Rating(5),
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
			payload:      "tmdb-id=14577&rating=4&watch-date=2022-10-21&blurb=",
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
				{
					token: "abc123",
					session: sessions.Session{
						Username: screenjournal.Username("dummyadmin"),
						IsAdmin:  true,
					},
				},
			},
			expected: screenjournal.Review{
				Owner:   screenjournal.Username("dummyadmin"),
				Rating:  screenjournal.Rating(4),
				Watched: mustParseWatchDate("2022-10-21"),
				Blurb:   screenjournal.Blurb(""),
				Movie: screenjournal.Movie{
					ID:          screenjournal.MovieID(1),
					TmdbID:      screenjournal.TmdbID(14577),
					ImdbID:      screenjournal.ImdbID("tt0120654"),
					Title:       "Dirty Work",
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("1998-06-12")),
				},
				Comments: []screenjournal.ReviewComment{},
			},
		},
		{
			description:  "valid request but we have to query metadata finder for movie info",
			payload:      "tmdb-id=38&rating=5&watch-date=2022-10-28&blurb=It's%20my%20favorite%20movie!",
			sessionToken: "abc123",
			remoteMovieInfo: []metadata.MovieInfo{
				{
					TmdbID:      screenjournal.TmdbID(38),
					ImdbID:      screenjournal.ImdbID("tt0338013"),
					Title:       screenjournal.MediaTitle("Eternal Sunshine of the Spotless Mind"),
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
				},
			},
			sessions: []mockSessionEntry{
				{
					token: "abc123",
					session: sessions.Session{
						Username: screenjournal.Username("dummyadmin"),
						IsAdmin:  true,
					},
				},
			},
			expected: screenjournal.Review{
				Owner:   screenjournal.Username("dummyadmin"),
				Rating:  screenjournal.Rating(5),
				Watched: mustParseWatchDate("2022-10-28"),
				Blurb:   screenjournal.Blurb("It's my favorite movie!"),
				Movie: screenjournal.Movie{
					ID:          screenjournal.MovieID(1),
					TmdbID:      screenjournal.TmdbID(38),
					ImdbID:      screenjournal.ImdbID("tt0338013"),
					Title:       "Eternal Sunshine of the Spotless Mind",
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
				},
				Comments: []screenjournal.ReviewComment{},
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

			announcer := mockAnnouncer{}

			sessionManager := newMockSessionManager(tt.sessions)

			remoteTvShowInfo := []metadata.TvShowInfo{}

			s := handlers.New(nilAuthenticator, &announcer, &sessionManager, dataStore, NewMockMetadataFinder(tt.remoteMovieInfo, remoteTvShowInfo))

			req, err := http.NewRequest("POST", "/reviews", strings.NewReader(tt.payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.AddCookie(&http.Cookie{
				Name:  mockSessionTokenName,
				Value: tt.sessionToken,
			})

			w := httptest.NewRecorder()
			s.Router().ServeHTTP(w, req)

			if got, want := w.Code, http.StatusSeeOther; got != want {
				t.Fatalf("httpStatus=%v, want=%v", got, want)
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
				t.Errorf("db did not contain expected review of %s", want.Movie.Title)
				t.Errorf("got=%#v, want=%#v, diff=%s", got, want, deep.Equal(got, want))
			}

			if got, want := len(announcer.announcedReviews), 1; got != want {
				t.Fatalf("reviewCountAnnounced=%d, want=%d", got, want)
			}

			clearUnpredictableReviewProperties(&announcer.announcedReviews[0])
			if got, want := announcer.announcedReviews[0], tt.expected; !reflect.DeepEqual(got, want) {
				t.Errorf("announcer did not announce expected review of %s", want.Movie.Title)
				t.Errorf("got=%#v, want=%#v, diff=%s", got, want, deep.Equal(got, want))
			}
		})
	}
}

func TestReviewsPostRejectsInvalidRequest(t *testing.T) {
	for _, tt := range []struct {
		description  string
		payload      string
		sessionToken string
		sessions     []mockSessionEntry
	}{
		{
			description:  "empty string",
			payload:      "",
			sessionToken: "abc123",
			sessions: []mockSessionEntry{
				{
					token: "abc123",
					session: sessions.Session{
						Username: screenjournal.Username("userA"),
					},
				},
			},
		},
		{
			description:  "empty payload",
			payload:      "tmdb-id=&rating=&watch-date=&blurb=",
			sessionToken: "abc123",
			sessions: []mockSessionEntry{
				{
					token: "abc123",
					session: sessions.Session{
						Username: screenjournal.Username("userA"),
					},
				},
			},
		},
		{
			description:  "invalid tmdb field (non-number)",
			payload:      "tmdb-id=banana&rating=5&watch-date=2022-10-28&blurb=It's%20my%20favorite%20movie!",
			sessionToken: "abc123",
			sessions: []mockSessionEntry{
				{
					token: "abc123",
					session: sessions.Session{
						Username: screenjournal.Username("userA"),
					},
				},
			},
		},
		{
			description:  "invalid rating field (non-number)",
			payload:      "tmdb-id=banana&rating=banana&watch-date=2022-10-28&blurb=It's%20my%20favorite%20movie!",
			sessionToken: "abc123",
			sessions: []mockSessionEntry{
				{
					token: "abc123",
					session: sessions.Session{
						Username: screenjournal.Username("userA"),
					},
				},
			},
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()

			announcer := mockAnnouncer{}

			sessionManager := newMockSessionManager(tt.sessions)
			s := handlers.New(nilAuthenticator, &announcer, &sessionManager, dataStore, mockMetadataFinder{})

			req, err := http.NewRequest("POST", "/reviews", strings.NewReader(tt.payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.AddCookie(&http.Cookie{
				Name:  mockSessionTokenName,
				Value: tt.sessionToken,
			})

			w := httptest.NewRecorder()
			s.Router().ServeHTTP(w, req)

			if got, want := w.Code, http.StatusBadRequest; got != want {
				t.Fatalf("/reviews POST returned wrong status: got=%v, want=%v", got, want)
			}

			if got, want := len(announcer.announcedReviews), 0; got != want {
				t.Errorf("announcedReviews=%d, want=%d", got, want)
			}
		})
	}
}

func TestReviewsPutAcceptsValidRequest(t *testing.T) {
	for _, tt := range []struct {
		description  string
		localMovies  []screenjournal.Movie
		priorReviews []screenjournal.Review
		sessions     []mockSessionEntry
		route        string
		payload      string
		sessionToken string
		expected     screenjournal.Review
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
					Rating:  screenjournal.Rating(5),
					Watched: mustParseWatchDate("2022-10-28"),
					Blurb:   screenjournal.Blurb("It's my favorite movie!"),
					Movie: screenjournal.Movie{
						ID:          screenjournal.MovieID(1),
						TmdbID:      screenjournal.TmdbID(38),
						ImdbID:      screenjournal.ImdbID("tt0338013"),
						Title:       "Eternal Sunshine of the Spotless Mind",
						ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
					},
				},
			},
			sessions: []mockSessionEntry{
				{
					token: "abc123",
					session: sessions.Session{
						Username: screenjournal.Username("userA"),
					},
				},
			},
			route:        "/reviews/1",
			payload:      "rating=4&watch-date=2022-10-30&blurb=It's%20a%20pretty%20good%20movie!",
			sessionToken: "abc123",
			expected: screenjournal.Review{
				Owner:   screenjournal.Username("userA"),
				Rating:  screenjournal.Rating(4),
				Watched: mustParseWatchDate("2022-10-30"),
				Blurb:   screenjournal.Blurb("It's a pretty good movie!"),
				Movie: screenjournal.Movie{
					ID:          screenjournal.MovieID(1),
					TmdbID:      screenjournal.TmdbID(38),
					ImdbID:      screenjournal.ImdbID("tt0338013"),
					Title:       "Eternal Sunshine of the Spotless Mind",
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
					Rating:  screenjournal.Rating(4),
					Watched: mustParseWatchDate("2022-10-21"),
					Blurb:   screenjournal.Blurb("Love Norm McDonald!"),
					Movie: screenjournal.Movie{
						ID:          screenjournal.MovieID(2),
						TmdbID:      screenjournal.TmdbID(14577),
						ImdbID:      screenjournal.ImdbID("tt0120654"),
						Title:       "Dirty Work",
						ReleaseDate: screenjournal.ReleaseDate(mustParseDate("1998-06-12")),
					},
				},
			},
			sessions: []mockSessionEntry{
				{
					token: "abc123",
					session: sessions.Session{
						Username: screenjournal.Username("userA"),
					},
				},
			},
			route:        "/reviews/1",
			payload:      "rating=3&watch-date=2022-10-28&blurb=",
			sessionToken: "abc123",
			expected: screenjournal.Review{
				Owner:   screenjournal.Username("userA"),
				Rating:  screenjournal.Rating(3),
				Watched: mustParseWatchDate("2022-10-28"),
				Blurb:   screenjournal.Blurb(""),
				Movie: screenjournal.Movie{
					ID:          screenjournal.MovieID(2),
					TmdbID:      screenjournal.TmdbID(14577),
					ImdbID:      screenjournal.ImdbID("tt0120654"),
					Title:       "Dirty Work",
					ReleaseDate: screenjournal.ReleaseDate(mustParseDate("1998-06-12")),
				},
				Comments: []screenjournal.ReviewComment{},
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
				if _, err := dataStore.InsertReview(r); err != nil {
					panic(err)
				}
			}

			sessionManager := newMockSessionManager(tt.sessions)
			s := handlers.New(nilAuthenticator, nilAnnouncer, &sessionManager, dataStore, mockMetadataFinder{})

			req, err := http.NewRequest("PUT", tt.route, strings.NewReader(tt.payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.AddCookie(&http.Cookie{
				Name:  mockSessionTokenName,
				Value: tt.sessionToken,
			})

			w := httptest.NewRecorder()
			s.Router().ServeHTTP(w, req)

			if got, want := w.Code, http.StatusSeeOther; got != want {
				t.Fatalf("handler returned wrong status code: got %v want %v", got, want)
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

func TestReviewsPutRejectsInvalidRequest(t *testing.T) {
	for _, tt := range []struct {
		description  string
		localMovies  []screenjournal.Movie
		priorReviews []screenjournal.Review
		sessions     []mockSessionEntry
		route        string
		payload      string
		sessionToken string
		status       int
	}{
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
					Rating:  screenjournal.Rating(5),
					Watched: mustParseWatchDate("2022-10-28"),
					Blurb:   screenjournal.Blurb("It's my favorite movie!"),
					Movie: screenjournal.Movie{
						ID:          screenjournal.MovieID(1),
						TmdbID:      screenjournal.TmdbID(38),
						ImdbID:      screenjournal.ImdbID("tt0338013"),
						Title:       "Eternal Sunshine of the Spotless Mind",
						ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
					},
				},
			},
			sessions: []mockSessionEntry{
				{
					token: "abc123",
					session: sessions.Session{
						Username: screenjournal.Username("userA"),
					},
				},
			},
			route:        "/reviews/0",
			payload:      "rating=4&watch-date=2022-10-30&blurb=It's%20a%20pretty%20good%20movie!",
			sessionToken: "abc123",
			status:       http.StatusBadRequest,
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
					Rating:  screenjournal.Rating(5),
					Watched: mustParseWatchDate("2022-10-28"),
					Blurb:   screenjournal.Blurb("It's my favorite movie!"),
					Movie: screenjournal.Movie{
						ID:          screenjournal.MovieID(1),
						TmdbID:      screenjournal.TmdbID(38),
						ImdbID:      screenjournal.ImdbID("tt0338013"),
						Title:       "Eternal Sunshine of the Spotless Mind",
						ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
					},
				},
			},
			sessions: []mockSessionEntry{
				{
					token: "abc123",
					session: sessions.Session{
						Username: screenjournal.Username("userA"),
					},
				},
			},
			route:        "/reviews/9876",
			payload:      "rating=4&watch-date=2022-10-30&blurb=It's%20a%20pretty%20good%20movie!",
			sessionToken: "abc123",
			status:       http.StatusNotFound,
		},
		{
			description: "rejects request with missing rating field",
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
					Rating:  screenjournal.Rating(5),
					Watched: mustParseWatchDate("2022-10-28"),
					Blurb:   screenjournal.Blurb("It's my favorite movie!"),
					Movie: screenjournal.Movie{
						ID:          screenjournal.MovieID(1),
						TmdbID:      screenjournal.TmdbID(38),
						ImdbID:      screenjournal.ImdbID("tt0338013"),
						Title:       "Eternal Sunshine of the Spotless Mind",
						ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
					},
				},
			},
			sessions: []mockSessionEntry{
				{
					token: "abc123",
					session: sessions.Session{
						Username: screenjournal.Username("userA"),
					},
				},
			},
			route:        "/reviews/1",
			payload:      "watch-date=2022-10-30&blurb=It's%20a%20pretty%20good%20movie!",
			sessionToken: "abc123",
			status:       http.StatusBadRequest,
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
					Rating:  screenjournal.Rating(5),
					Watched: mustParseWatchDate("2022-10-28"),
					Blurb:   screenjournal.Blurb("It's my favorite movie!"),
					Movie: screenjournal.Movie{
						ID:          screenjournal.MovieID(1),
						TmdbID:      screenjournal.TmdbID(38),
						ImdbID:      screenjournal.ImdbID("tt0338013"),
						Title:       "Eternal Sunshine of the Spotless Mind",
						ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
					},
				},
			},
			sessions: []mockSessionEntry{
				{
					token: "abc123",
					session: sessions.Session{
						Username: screenjournal.Username("userA"),
					},
				},
			},
			route:        "/reviews/1",
			payload:      "rating=4&blurb=It's%20a%20pretty%20good%20movie!",
			sessionToken: "abc123",
			status:       http.StatusBadRequest,
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
					Rating:  screenjournal.Rating(5),
					Watched: mustParseWatchDate("2022-10-28"),
					Blurb:   screenjournal.Blurb("It's my favorite movie!"),
					Movie: screenjournal.Movie{
						ID:          screenjournal.MovieID(1),
						TmdbID:      screenjournal.TmdbID(38),
						ImdbID:      screenjournal.ImdbID("tt0338013"),
						Title:       "Eternal Sunshine of the Spotless Mind",
						ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
					},
				},
			},
			sessions: []mockSessionEntry{
				{
					token: "abc123",
					session: sessions.Session{
						Username: screenjournal.Username("userA"),
					},
				},
			},
			route:        "/reviews/1",
			payload:      "rating=4&watch-date=2022-10-30&blurb=Nothing%20evil%20going%20on%20here...%3Cscript%3Ealert(1)%3C%2Fscript%3E",
			sessionToken: "abc123",
			status:       http.StatusBadRequest,
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
					Rating:  screenjournal.Rating(5),
					Watched: mustParseWatchDate("2022-10-28"),
					Blurb:   screenjournal.Blurb("It's my favorite movie!"),
					Movie: screenjournal.Movie{
						ID:          screenjournal.MovieID(1),
						TmdbID:      screenjournal.TmdbID(38),
						ImdbID:      screenjournal.ImdbID("tt0338013"),
						Title:       "Eternal Sunshine of the Spotless Mind",
						ReleaseDate: screenjournal.ReleaseDate(mustParseDate("2004-03-19")),
					},
				},
			},
			sessions: []mockSessionEntry{
				{
					token: "abc123",
					session: sessions.Session{
						Username: screenjournal.Username("userA"),
					},
				},
				{
					token: "def456",
					session: sessions.Session{
						Username: screenjournal.Username("userB"),
					},
				},
			},
			route:        "/reviews/1",
			payload:      "rating=4&watch-date=2022-10-30&blurb=I'm%20overwriting%20userA's%20review!",
			sessionToken: "def456",
			status:       http.StatusForbidden,
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
				if _, err := dataStore.InsertReview(r); err != nil {
					panic(err)
				}
			}

			sessionManager := newMockSessionManager(tt.sessions)

			s := handlers.New(nilAuthenticator, nilAnnouncer, &sessionManager, dataStore, mockMetadataFinder{})

			req, err := http.NewRequest("PUT", tt.route, strings.NewReader(tt.payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.AddCookie(&http.Cookie{
				Name:  mockSessionTokenName,
				Value: tt.sessionToken,
			})

			w := httptest.NewRecorder()
			s.Router().ServeHTTP(w, req)

			if got, want := w.Code, tt.status; got != want {
				t.Fatalf("%s PUT returned wrong status: got=%v, want=%v", tt.route, got, want)
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
