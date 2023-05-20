//go:build dev

package handlers

import (
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/random"
	"github.com/mtlynch/screenjournal/v2/store"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

// addDevRoutes adds debug routes that we only use during development or e2e
// tests.
func (s *Server) addDevRoutes() {
	s.router.Use(assignSessionDB)
	s.router.HandleFunc("/api/debug/db/populate-dummy-data", s.populateDummyData()).Methods(http.MethodGet)
	s.router.HandleFunc("/api/debug/db/per-session", dbPerSessionPost()).Methods(http.MethodPost)
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
	reviews := []screenjournal.Review{
		{
			ID:     screenjournal.ReviewID(1),
			Owner:  screenjournal.Username("userB"),
			Rating: screenjournal.Rating(5),
			Movie: screenjournal.Movie{
				ID:    screenjournal.MovieID(1),
				Title: screenjournal.MediaTitle("The Waterboy"),
			},
			Watched: mustParseWatchDate("2020-10-05T20:18:55-04:00"),
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
			Rating: screenjournal.Rating(2),
			Movie: screenjournal.Movie{
				ID:    screenjournal.MovieID(2),
				Title: screenjournal.MediaTitle("Billy Madison"),
			},
			Watched: mustParseWatchDate("2023-02-05T20:18:55-04:00"),
			Blurb:   screenjournal.Blurb("A staggering lack of water."),
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
		for _, review := range reviews {
			if _, err := s.getDB(r).InsertReview(review); err != nil {
				panic(err)
			}

			for _, c := range review.Comments {
				c.Review = review
				if _, err := s.getDB(r).InsertComment(c); err != nil {
					panic(err)
				}
			}
		}
	}
}

const dbTokenCookieName = "db-token"

type (
	dbToken string

	dbSettings struct {
		isolateBySession bool
		tokenToDB        map[dbToken]store.Store
		lock             sync.RWMutex
	}
)

var sharedDBSettings = dbSettings{
	tokenToDB: map[dbToken]store.Store{},
}

func (dbs *dbSettings) IsSessionIsolationEnabled() bool {
	dbs.lock.RLock()
	dbs.lock.RUnlock()
	return dbs.isolateBySession
}

func (dbs *dbSettings) EnableSessionIsolation() {
	dbs.lock.Lock()
	dbs.isolateBySession = true
	dbs.lock.Unlock()
	log.Print("per-session database = on")
}

func (dbs *dbSettings) GetDB(token dbToken) store.Store {
	dbs.lock.RLock()
	defer dbs.lock.RUnlock()
	return dbs.tokenToDB[token]
}

func (dbs *dbSettings) SaveDB(token dbToken, db store.Store) {
	dbs.lock.Lock()
	defer dbs.lock.Unlock()
	dbs.tokenToDB[token] = db
}

func (s Server) getDB(r *http.Request) store.Store {
	if !sharedDBSettings.IsSessionIsolationEnabled() {
		return s.store
	}
	c, err := r.Cookie(dbTokenCookieName)
	if err != nil {
		panic(err)
	}
	return sharedDBSettings.GetDB(dbToken(c.Value))
}

func (s Server) getAuthenticator(r *http.Request) auth.Authenticator {
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
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		log.Fatalf("failed to parse release date: %s", s)
	}
	return screenjournal.ReleaseDate(d)
}

func mustParseWatchDate(s string) screenjournal.WatchDate {
	wd, err := parse.WatchDate(s)
	if err != nil {
		log.Fatalf("failed to parse watch date: %s", s)
	}
	return wd
}

// assignSessionDB provisions a session-specific database if per-session
// databases are enabled. If per-session databases are not enabled (the default)
// this is a no-op.
func assignSessionDB(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if sharedDBSettings.IsSessionIsolationEnabled() {
			if _, err := r.Cookie(dbTokenCookieName); err != nil {
				token := dbToken(random.String(30, []rune("abcdefghijkmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")))
				log.Printf("provisioning a new private database with token %s", token)
				createDBCookie(token, w)
				sharedDBSettings.SaveDB(token, test_sqlite.New())
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
	h, err := auth.NewPasswordHash(plaintext)
	if err != nil {
		panic(err)
	}
	return screenjournal.PasswordHash(h.Bytes())
}
