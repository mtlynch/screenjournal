package sessions_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	simple_sessions "codeberg.org/mtlynch/simpleauth/v3/sessions"

	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

func TestManagerStoresSessionsInSQLite(t *testing.T) {
	db := test_sqlite.NewDB(t)
	manager := sessions.NewManager(func(_ context.Context) *sql.DB { return db }, false)
	userID, err := simple_sessions.NewUserID("dummyuserID")
	if err != nil {
		t.Fatalf("failed to create user ID: %v", err)
	}

	rec := httptest.NewRecorder()
	if err := manager.LogIn(
		context.Background(),
		rec,
		userID,
	); err != nil {
		t.Fatalf("LogIn err=%v, want=%v", err, nil)
	}

	cookies := rec.Result().Cookies()
	if got, want := len(cookies), 1; got != want {
		t.Fatalf("len(cookies)=%d, want=%d", got, want)
	}

	var loaded simple_sessions.UserID
	var loadErr error
	manager.LoadUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loaded, loadErr = manager.UserIDFromContext(r.Context())
	})).ServeHTTP(httptest.NewRecorder(), requestWithCookie(cookies[0]))

	if got, want := loadErr, error(nil); got != want {
		t.Fatalf("loadErr=%v, want=%v", got, want)
	}
	if got, want := loaded.String(), userID.String(); got != want {
		t.Errorf("loaded=%v, want=%v", got, want)
	}
	var userIDRaw string
	var createdAtRaw string
	var expiresAtRaw string
	if err := db.QueryRow(`
		SELECT
			user_id,
			created_at,
			expires_at
		FROM
			auth_sessions
		WHERE
			session_id = :session_id`,
		sql.Named("session_id", cookies[0].Value)).
		Scan(&userIDRaw, &createdAtRaw, &expiresAtRaw); err != nil {
		t.Fatalf("failed to read stored session: %v", err)
	}
	if got, want := userIDRaw, userID.String(); got != want {
		t.Errorf("userID=%s, want=%s", got, want)
	}
	createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtRaw)
	if err != nil {
		t.Fatalf("createdAt=%q failed to parse: %v", createdAtRaw, err)
	}
	expiresAt, err := time.Parse("2006-01-02 15:04:05", expiresAtRaw)
	if err != nil {
		t.Fatalf("expiresAt=%q failed to parse: %v", expiresAtRaw, err)
	}
	if got, want := expiresAt.Sub(createdAt), 30*24*time.Hour; got != want {
		t.Errorf("session lifetime=%v, want=%v", got, want)
	}
	endRec := httptest.NewRecorder()
	var endErr error
	manager.LoadUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		endErr = manager.LogOut(r.Context(), w)
	})).ServeHTTP(endRec, requestWithCookie(cookies[0]))
	if got, want := endErr, error(nil); got != want {
		t.Fatalf("endErr=%v, want=%v", got, want)
	}

	manager.LoadUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, loadErr = manager.UserIDFromContext(r.Context())
	})).ServeHTTP(httptest.NewRecorder(), requestWithCookie(cookies[0]))
	if got, want := loadErr, simple_sessions.ErrNoSessionFound; got != want {
		t.Fatalf("loadErr=%v, want=%v", got, want)
	}
}

func requestWithCookie(cookie *http.Cookie) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)
	return req
}
