//go:build dev

package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	simple_sessions "codeberg.org/mtlynch/simpleauth/v3/sessions"

	"github.com/mtlynch/screenjournal/v2/handlers"
	appsessions "github.com/mtlynch/screenjournal/v2/handlers/sessions"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

func TestAuthenticatedRequestPanicsForStalePerSessionDatabaseToken(t *testing.T) {
	dataStore := test_sqlite.New()
	sessionManager := appsessions.NewManager(dataStore, false)
	userID, err := simple_sessions.NewUserID(userA.Username.String())
	if err != nil {
		t.Fatalf("NewUserID err=%v, want=%v", err, nil)
	}

	loginRec := httptest.NewRecorder()
	if err := sessionManager.LogIn(context.Background(), loginRec, userID); err != nil {
		t.Fatalf("LogIn err=%v, want=%v", err, nil)
	}

	var authCookie *http.Cookie
	for _, cookie := range loginRec.Result().Cookies() {
		if cookie.Name == "token" {
			authCookie = cookie
			break
		}
	}
	if authCookie == nil {
		t.Fatal("authCookie=nil, want session cookie")
	}

	s := handlers.New(handlers.ServerParams{
		SessionManager: sessionManager,
		Store:          dataStore,
	})

	enableIsolationReq := httptest.NewRequest(
		http.MethodPost,
		"/api/debug/db/per-session",
		nil,
	)
	enableIsolationRec := httptest.NewRecorder()
	s.Router().ServeHTTP(enableIsolationRec, enableIsolationReq)
	if got, want := enableIsolationRec.Code, http.StatusOK; got != want {
		t.Fatalf("enableIsolationRec.Code=%d, want=%d", got, want)
	}

	req := httptest.NewRequest(http.MethodGet, "/account/security", nil)
	req.AddCookie(authCookie)
	req.AddCookie(&http.Cookie{
		Name:  "db-token",
		Value: "stale-token",
	})

	defer func() {
		if got, want := recover() == nil, false; got != want {
			t.Fatal("recover()=nil, want panic")
		}
	}()

	s.Router().ServeHTTP(httptest.NewRecorder(), req)
}
