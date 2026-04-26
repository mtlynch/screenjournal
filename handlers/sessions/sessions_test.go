package sessions_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

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

	loadRec := httptest.NewRecorder()
	manager.LoadUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loadedUserID, err := manager.UserIDFromContext(r.Context())
		if err != nil {
			t.Fatalf("UserIDFromContext err=%v, want=%v", err, nil)
		}
		if _, err := w.Write([]byte("loaded user " + loadedUserID.String())); err != nil {
			t.Fatalf("Write err=%v, want=%v", err, nil)
		}
	})).ServeHTTP(loadRec, requestWithCookie(cookies[0]))
	if got, want := loadRec.Code, http.StatusOK; got != want {
		t.Fatalf("loadRec.Code=%d, want=%d", got, want)
	}
	if got, want := loadRec.Body.String(), "loaded user "+userID.String(); got != want {
		t.Errorf("loadRec.Body=%q, want=%q", got, want)
	}
	endRec := httptest.NewRecorder()
	var endErr error
	manager.LoadUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		endErr = manager.LogOut(r.Context(), w)
	})).ServeHTTP(endRec, requestWithCookie(cookies[0]))
	if got, want := endErr, error(nil); got != want {
		t.Fatalf("endErr=%v, want=%v", got, want)
	}

	var postLogoutErr error
	manager.LoadUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, postLogoutErr = manager.UserIDFromContext(r.Context())
	})).ServeHTTP(httptest.NewRecorder(), requestWithCookie(cookies[0]))
	if got, want := postLogoutErr, simple_sessions.ErrNoSessionFound; got != want {
		t.Fatalf("postLogoutErr=%v, want=%v", got, want)
	}
}

func requestWithCookie(cookie *http.Cookie) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)
	return req
}
