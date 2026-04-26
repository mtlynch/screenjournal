//go:build dev

package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/gorilla/mux"
	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
	"github.com/mtlynch/screenjournal/v2/metadata/tmdb"
	"github.com/mtlynch/screenjournal/v2/random"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store/sqlite"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

// initDev sets up dev-mode state before routes are created.
// It adds the assignSessionDB middleware which must run BEFORE
// populateAuthenticationContext to ensure the test database is
// available when looking up user information.
//
// It also replaces the session manager so that sessions are stored in
// the same per-session database as application data (users, reviews,
// etc.), eliminating the shared-DB/per-session-DB mismatch.
func (s *Server) initDev() {
	s.router.Use(assignSessionDB)
	if s.rawDB != nil {
		s.store = sqlite.New(func(ctx context.Context) *sql.DB {
			if !sharedDBSettings.IsSessionIsolationEnabled() {
				return s.rawDB
			}
			if token, ok := ctx.Value(dbTokenContextKey{}).(dbToken); ok {
				if sess := sharedDBSettings.GetDB(token); sess.db != nil {
					return sess.db
				}
			}
			return s.rawDB
		}, false)
		s.sessionManager = sessions.NewManager(s.store, false)
	}
}

// addDevRoutes adds debug routes that we only use during development or e2e
// tests.
func (s *Server) addDevRoutes() {
	s.router.HandleFunc("/api/debug/db/populate-dummy-data", s.populateDummyData()).Methods(http.MethodGet)
	s.router.HandleFunc("/api/debug/db/per-session", dbPerSessionPost()).Methods(http.MethodPost)
	s.router.HandleFunc("/api/debug/password-reset-token/{username}", s.debugPasswordResetTokenGet()).Methods(http.MethodGet)
}

func (s Server) populateDummyData() http.HandlerFunc {
	users := []screenjournal.User{
		{
			Username:     screenjournal.Username("dummyadmin"),
			PasswordHash: mustCreatePasswordHash("dummypass"),
			IsAdmin:      true,
			Email:        screenjournal.Email("dummyadmin@example.com"),
		},
		{
			Username:     screenjournal.Username("userA"),
			PasswordHash: mustCreatePasswordHash("password123"),
			IsAdmin:      false,
			Email:        screenjournal.Email("userA@example.com"),
		},
		{
			Username:     screenjournal.Username("userB"),
			PasswordHash: mustCreatePasswordHash("password456"),
			IsAdmin:      false,
			Email:        screenjournal.Email("userB@example.com"),
		},
	}
	movies := []screenjournal.Movie{
		{
			ID:          screenjournal.MovieID(1),
			TmdbID:      screenjournal.TmdbID(10663),
			ImdbID:      screenjournal.ImdbID("tt0120484"),
			Title:       screenjournal.MediaTitle("The Waterboy"),
			ReleaseDate: mustParseReleaseDate("1998-11-06"),
			PosterPath: url.URL{
				Path: "/miT42qWYC4D0n2mXNzJ9VfhheWW.jpg",
			},
		},
		{
			ID:          screenjournal.MovieID(2),
			TmdbID:      screenjournal.TmdbID(11017),
			ImdbID:      screenjournal.ImdbID("tt0112508"),
			Title:       screenjournal.MediaTitle("Billy Madison"),
			ReleaseDate: mustParseReleaseDate("1995-02-10"),
			PosterPath: url.URL{
				Path: "/iwk9pWR6MwTInEQc8Vw5vGHjeQ0.jpg",
			},
		},
	}
	tvShows := []screenjournal.TvShow{
		{
			ID:      screenjournal.TvShowID(1),
			TmdbID:  screenjournal.TmdbID(1400),
			ImdbID:  screenjournal.ImdbID("tt0098904"),
			Title:   screenjournal.MediaTitle("Seinfeld"),
			AirDate: mustParseReleaseDate("1989-07-05"),
			PosterPath: url.URL{
				Path: "/aCw8ONfyz3AhngVQa1E2Ss4KSUQ.jpg",
			},
		},
	}
	reviews := []screenjournal.Review{
		{
			ID:     screenjournal.ReviewID(1),
			Owner:  screenjournal.Username("userB"),
			Rating: screenjournal.NewRating(10),
			Movie: screenjournal.Movie{
				ID:    screenjournal.MovieID(1),
				Title: screenjournal.MediaTitle("The Waterboy"),
			},
			Watched: mustParseWatchDate("2020-10-05"),
			Blurb:   screenjournal.Blurb("I love water!"),
			Comments: []screenjournal.ReviewComment{
				{
					ID:          screenjournal.CommentID(1),
					Owner:       screenjournal.Username("userA"),
					CommentText: screenjournal.CommentText("You sure do!"),
				},
			},
		},
		{
			ID:     screenjournal.ReviewID(2),
			Owner:  screenjournal.Username("userB"),
			Rating: screenjournal.NewRating(3),
			Movie: screenjournal.Movie{
				ID:    screenjournal.MovieID(2),
				Title: screenjournal.MediaTitle("Billy Madison"),
			},
			Watched: mustParseWatchDate("2023-02-05"),
			Blurb:   screenjournal.Blurb("A staggering lack of water."),
		},
		{
			ID:     screenjournal.ReviewID(3),
			Owner:  screenjournal.Username("userA"),
			Rating: screenjournal.NewRating(10),
			TvShow: screenjournal.TvShow{
				ID:    screenjournal.TvShowID(1),
				Title: screenjournal.MediaTitle("Seinfeld"),
			},
			TvShowSeason: screenjournal.TvShowSeason(1),
			Watched:      mustParseWatchDate("2024-11-04"),
			Blurb:        screenjournal.Blurb("I see what the fuss is about!"),
		},
		{
			ID:     screenjournal.ReviewID(3),
			Owner:  screenjournal.Username("userB"),
			Rating: screenjournal.NewRating(9),
			TvShow: screenjournal.TvShow{
				ID:    screenjournal.TvShowID(1),
				Title: screenjournal.MediaTitle("Seinfeld"),
			},
			TvShowSeason: screenjournal.TvShowSeason(2),
			Watched:      mustParseWatchDate("2024-11-05"),
			Blurb:        screenjournal.Blurb("Loving this second season!"),
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		for _, u := range users {
			if err := s.getDB(r).InsertUser(u); err != nil {
				panic(err)
			}
		}
		for _, movie := range movies {
			if _, err := s.getDB(r).InsertMovie(movie); err != nil {
				panic(err)
			}
		}
		for _, tvShow := range tvShows {
			if _, err := s.getDB(r).InsertTvShow(tvShow); err != nil {
				panic(err)
			}
		}
		for _, review := range reviews {
			reviewID, err := s.getDB(r).InsertReview(review)
			if err != nil {
				panic(err)
			}
			review.ID = reviewID

			for _, c := range review.Comments {
				c.Review = review
				if _, err := s.getDB(r).InsertComment(c); err != nil {
					panic(err)
				}
			}

			if review.Owner.Equal(screenjournal.Username("userB")) && review.Movie.Title == "The Waterboy" {
				reaction := screenjournal.ReviewReaction{
					Review: review,
					Owner:  screenjournal.Username("dummyadmin"),
					Emoji:  screenjournal.NewReactionEmoji("🥞"),
				}
				if _, err := s.getDB(r).InsertReaction(reaction); err != nil {
					panic(err)
				}
			}
		}
	}
}

const dbTokenCookieName = "db-token"

type (
	dbToken string

	sessionDB struct {
		db *sql.DB
	}

	dbSettings struct {
		isolateBySession bool
		tokenToDB        map[dbToken]sessionDB
		lock             sync.RWMutex
	}
)

var sharedDBSettings = dbSettings{
	tokenToDB: map[dbToken]sessionDB{},
}

func (dbs *dbSettings) IsSessionIsolationEnabled() bool {
	dbs.lock.RLock()
	defer dbs.lock.RUnlock()
	return dbs.isolateBySession
}

func (dbs *dbSettings) EnableSessionIsolation() {
	dbs.lock.Lock()
	dbs.isolateBySession = true
	dbs.lock.Unlock()
	log.Print("per-session database = on")
}

func (dbs *dbSettings) GetDB(token dbToken) sessionDB {
	dbs.lock.RLock()
	defer dbs.lock.RUnlock()
	return dbs.tokenToDB[token]
}

func (dbs *dbSettings) SaveDB(token dbToken, sdb sessionDB) {
	dbs.lock.Lock()
	defer dbs.lock.Unlock()
	dbs.tokenToDB[token] = sdb
}

func (s Server) getDB(r *http.Request) sqlite.Store {
	return s.store.WithContext(r.Context())
}

func (s Server) getAuthenticator(r *http.Request) Authenticator {
	if !sharedDBSettings.IsSessionIsolationEnabled() {
		return s.authenticator
	}
	return auth.New(s.getDB(r))
}

func dbPerSessionPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sharedDBSettings.EnableSessionIsolation()
	}
}

func mustParseReleaseDate(s string) screenjournal.ReleaseDate {
	d, err := tmdb.ParseReleaseDate(s)
	if err != nil {
		log.Fatalf("failed to parse release date: %s", s)
	}
	return d
}

func mustParseWatchDate(s string) screenjournal.WatchDate {
	wd, err := parse.WatchDate(s)
	if err != nil {
		log.Fatalf("failed to parse watch date: %s", s)
	}
	return wd
}

type dbTokenContextKey struct{}

// assignSessionDB provisions a session-specific database if per-session
// databases are enabled. If per-session databases are not enabled (the default)
// this is a no-op.
func assignSessionDB(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if sharedDBSettings.IsSessionIsolationEnabled() {
			c, err := r.Cookie(dbTokenCookieName)
			if err != nil {
				token := dbToken(random.String(30, []rune("abcdefghijkmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")))
				log.Printf("provisioning a new private database with token %s", token)
				createDBCookie(token, w)
				db, _ := test_sqlite.New()
				sharedDBSettings.SaveDB(token, sessionDB{db: db})
				ctx := context.WithValue(r.Context(), dbTokenContextKey{}, token)
				r = r.WithContext(ctx)
			} else {
				ctx := context.WithValue(r.Context(), dbTokenContextKey{}, dbToken(c.Value))
				r = r.WithContext(ctx)
			}
		}
		h.ServeHTTP(w, r)
	})
}

func createDBCookie(token dbToken, w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:  dbTokenCookieName,
		Value: string(token),
		Path:  "/",
	})
}

func mustCreatePasswordHash(plaintext string) screenjournal.PasswordHash {
	h, err := auth.HashPassword(screenjournal.Password(plaintext))
	if err != nil {
		panic(err)
	}
	return screenjournal.PasswordHash(h.Bytes())
}

func (s Server) debugPasswordResetTokenGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		username, err := parse.Username(vars["username"])
		if err != nil {
			http.Error(w, "Invalid username", http.StatusBadRequest)
			return
		}

		entry, err := s.getDB(r).ReadLatestPasswordResetEntryForUser(username)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "No password reset token found for user", http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf("Failed to read password reset entry: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"token": entry.Token.String()}); err != nil {
			log.Printf("failed to encode password reset token response: %v", err)
		}
	}
}
