package sessions_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	simple_sessions "codeberg.org/mtlynch/simpleauth/v3/sessions"

	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

func TestManagerStoresSessionsInSQLite(t *testing.T) {
	db := test_sqlite.NewDB(t)
	manager := sessions.NewManager(db, false)
	userID, err := simple_sessions.NewUserID("mavis")
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
	if got, want := loaded.String(), "mavis"; got != want {
		t.Errorf("loaded=%v, want=%v", got, want)
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
