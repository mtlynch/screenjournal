package sessions_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

func TestManagerStoresSessionsInSQLite(t *testing.T) {
	db := test_sqlite.NewDB(t)
	manager := sessions.NewManager(db, false)

	rec := httptest.NewRecorder()
	if err := manager.CreateSession(
		rec,
		context.Background(),
		screenjournal.Username("mavis"),
	); err != nil {
		t.Fatalf("CreateSession err=%v, want=%v", err, nil)
	}

	cookies := rec.Result().Cookies()
	if got, want := len(cookies), 1; got != want {
		t.Fatalf("len(cookies)=%d, want=%d", got, want)
	}

	var loaded sessions.Session
	var loadErr error
	manager.WrapRequest(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loaded, loadErr = manager.SessionFromContext(r.Context())
	})).ServeHTTP(httptest.NewRecorder(), requestWithCookie(cookies[0]))

	if got, want := loadErr, error(nil); got != want {
		t.Fatalf("loadErr=%v, want=%v", got, want)
	}
	if got, want := loaded.Username, screenjournal.Username("mavis"); got != want {
		t.Errorf("loaded.Username=%v, want=%v", got, want)
	}
	endRec := httptest.NewRecorder()
	var endErr error
	manager.WrapRequest(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		endErr = manager.EndSession(r.Context(), w)
	})).ServeHTTP(endRec, requestWithCookie(cookies[0]))
	if got, want := endErr, error(nil); got != want {
		t.Fatalf("endErr=%v, want=%v", got, want)
	}

	manager.WrapRequest(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, loadErr = manager.SessionFromContext(r.Context())
	})).ServeHTTP(httptest.NewRecorder(), requestWithCookie(cookies[0]))
	if got, want := loadErr, sessions.ErrNoSessionFound; got != want {
		t.Fatalf("loadErr=%v, want=%v", got, want)
	}
}

func requestWithCookie(cookie *http.Cookie) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)
	return req
}
