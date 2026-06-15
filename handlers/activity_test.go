package handlers_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store/sqlite"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

func TestActivityPageOrdersItemsAndGroupsByDate(t *testing.T) {
	db := test_sqlite.NewDB(t)
	dataStore := sqlite.New(db, false)
	insertMockUsers(t, dataStore, []screenjournal.User{
		{
			Username:     screenjournal.Username("mike"),
			Email:        screenjournal.Email("mike@example.com"),
			PasswordHash: screenjournal.PasswordHash("dummy-password-hash"),
		},
		{
			Username:     screenjournal.Username("jamie"),
			Email:        screenjournal.Email("jamie@example.com"),
			PasswordHash: screenjournal.PasswordHash("dummy-password-hash"),
		},
		{
			Username:     screenjournal.Username("joe"),
			Email:        screenjournal.Email("joe@example.com"),
			PasswordHash: screenjournal.PasswordHash("dummy-password-hash"),
		},
	})

	movieID, err := dataStore.InsertMovie(screenjournal.Movie{
		Title:  screenjournal.MediaTitle("Poker Face"),
		ImdbID: screenjournal.ImdbID("tt1234567"),
	})
	if err != nil {
		t.Fatalf("failed to insert movie: %v", err)
	}
	reviewID, err := dataStore.InsertReview(screenjournal.Review{
		Owner:   screenjournal.Username("mike"),
		Rating:  screenjournal.NewRating(7),
		Movie:   screenjournal.Movie{ID: movieID, Title: screenjournal.MediaTitle("Poker Face")},
		Watched: screenjournal.WatchDate(time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)),
	})
	if err != nil {
		t.Fatalf("failed to insert review: %v", err)
	}
	setReviewCreatedTime(t, db, reviewID, time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC))
	review := screenjournal.Review{
		ID:    reviewID,
		Owner: screenjournal.Username("mike"),
		Movie: screenjournal.Movie{ID: movieID, Title: screenjournal.MediaTitle("Poker Face")},
	}
	commentID, err := dataStore.InsertComment(screenjournal.ReviewComment{
		Review:      review,
		Owner:       screenjournal.Username("jamie"),
		CommentText: screenjournal.CommentText("Nice review."),
	})
	if err != nil {
		t.Fatalf("failed to insert comment: %v", err)
	}
	setCommentCreatedTime(t, db, commentID, time.Date(2025, 1, 1, 11, 0, 0, 0, time.UTC))
	reactionID, err := dataStore.InsertReaction(screenjournal.ReviewReaction{
		Review: review,
		Owner:  screenjournal.Username("joe"),
		Emoji:  screenjournal.NewReactionEmoji("🥞"),
	})
	if err != nil {
		t.Fatalf("failed to insert reaction: %v", err)
	}
	setReactionCreatedTime(t, db, reactionID, time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC))

	sessionManager := newMockSessionManager([]mockSessionEntry{
		{
			token: "session-token",
			session: mockSession{
				Username: screenjournal.Username("mike"),
			},
		},
	})
	server := handlers.New(handlers.ServerParams{
		SessionManager: &sessionManager,
		Store:          dataStore,
	})
	req := httptest.NewRequest(http.MethodGet, "/activity", nil)
	req.AddCookie(&http.Cookie{Name: mockSessionTokenName, Value: "session-token"})
	rec := httptest.NewRecorder()

	server.Router().ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusOK; got != want {
		t.Fatalf("status=%d, want=%d", got, want)
	}
	body := rec.Body.String()
	if got, want := body, "Jan 1, 2025"; !strings.Contains(got, want) {
		t.Fatalf("body does not contain %q", want)
	}
	reactionIndex := strings.Index(body, "reacted to")
	commentIndex := strings.Index(body, "replied to")
	reviewIndex := strings.Index(body, "reviewed")
	if reactionIndex == -1 || commentIndex == -1 || reviewIndex == -1 {
		t.Fatalf("body missing activity entries: %s", body)
	}
	if !(reactionIndex < commentIndex && commentIndex < reviewIndex) {
		t.Errorf(
			"activity order reaction=%d, comment=%d, review=%d; want reaction before comment before review",
			reactionIndex,
			commentIndex,
			reviewIndex,
		)
	}
}

func TestActivityPageUsesTvShowSeasonLinks(t *testing.T) {
	db := test_sqlite.NewDB(t)
	dataStore := sqlite.New(db, false)
	insertMockUsers(t, dataStore, []screenjournal.User{
		{
			Username:     screenjournal.Username("dave"),
			Email:        screenjournal.Email("dave@example.com"),
			PasswordHash: screenjournal.PasswordHash("dummy-password-hash"),
		},
		{
			Username:     screenjournal.Username("jamie"),
			Email:        screenjournal.Email("jamie@example.com"),
			PasswordHash: screenjournal.PasswordHash("dummy-password-hash"),
		},
	})

	tvShowID, err := dataStore.InsertTvShow(screenjournal.TvShow{
		Title:  screenjournal.MediaTitle("Batman Forever"),
		ImdbID: screenjournal.ImdbID("tt2345678"),
	})
	if err != nil {
		t.Fatalf("failed to insert TV show: %v", err)
	}
	reviewID, err := dataStore.InsertReview(screenjournal.Review{
		Owner:        screenjournal.Username("dave"),
		Rating:       screenjournal.NewRating(5),
		TvShow:       screenjournal.TvShow{ID: tvShowID, Title: screenjournal.MediaTitle("Batman Forever")},
		TvShowSeason: screenjournal.TvShowSeason(2),
		Watched:      screenjournal.WatchDate(time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC)),
	})
	if err != nil {
		t.Fatalf("failed to insert review: %v", err)
	}
	review := screenjournal.Review{
		ID:           reviewID,
		Owner:        screenjournal.Username("dave"),
		TvShow:       screenjournal.TvShow{ID: tvShowID, Title: screenjournal.MediaTitle("Batman Forever")},
		TvShowSeason: screenjournal.TvShowSeason(2),
	}
	commentID, err := dataStore.InsertComment(screenjournal.ReviewComment{
		Review:      review,
		Owner:       screenjournal.Username("jamie"),
		CommentText: screenjournal.CommentText("Nice review."),
	})
	if err != nil {
		t.Fatalf("failed to insert comment: %v", err)
	}

	sessionManager := newMockSessionManager([]mockSessionEntry{
		{
			token: "session-token",
			session: mockSession{
				Username: screenjournal.Username("dave"),
			},
		},
	})
	server := handlers.New(handlers.ServerParams{
		SessionManager: &sessionManager,
		Store:          dataStore,
	})
	req := httptest.NewRequest(http.MethodGet, "/activity", nil)
	req.AddCookie(&http.Cookie{Name: mockSessionTokenName, Value: "session-token"})
	rec := httptest.NewRecorder()

	server.Router().ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusOK; got != want {
		t.Fatalf("status=%d, want=%d", got, want)
	}
	body := rec.Body.String()
	if got, want := body, "/tv-shows/"+tvShowID.String()+"?season=2#comment"+commentID.String(); !strings.Contains(got, want) {
		t.Errorf("body does not contain %q", want)
	}
	if got, want := body, "/tv-shows/"+tvShowID.String()+"?season=2#review"+reviewID.String(); !strings.Contains(got, want) {
		t.Errorf("body does not contain %q", want)
	}
}

func setReviewCreatedTime(t *testing.T, db *sql.DB, id screenjournal.ReviewID, created time.Time) {
	t.Helper()
	setCreatedTime(t, db, "reviews", id.String(), created)
}

func setCommentCreatedTime(t *testing.T, db *sql.DB, id screenjournal.CommentID, created time.Time) {
	t.Helper()
	setCreatedTime(t, db, "review_comments", id.String(), created)
}

func setReactionCreatedTime(t *testing.T, db *sql.DB, id screenjournal.ReactionID, created time.Time) {
	t.Helper()
	setCreatedTime(t, db, "review_reactions", id.String(), created)
}

func setCreatedTime(t *testing.T, db *sql.DB, table string, id string, created time.Time) {
	t.Helper()
	if _, err := db.Exec(
		"UPDATE "+table+" SET created_time = :created_time WHERE id = :id",
		sql.Named("created_time", created.Format(time.RFC3339)),
		sql.Named("id", id)); err != nil {
		t.Fatalf("failed to set created time: %v", err)
	}
}
