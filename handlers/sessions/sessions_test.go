package sessions_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	simple_sessions "codeberg.org/mtlynch/simpleauth/v3/sessions"

	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
	"github.com/mtlynch/screenjournal/v2/store/sqlite"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

func TestManagerStoresSessionsInSQLite(t *testing.T) {
	db, store := test_sqlite.New()
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close db: %v", err)
		}
	})
	manager := sessions.NewManager(func(_ context.Context) sqlite.Store { return store }, false)
	userID, err := simple_sessions.NewUserID("dummyuserID")
	if err != nil {
		t.Fatalf("failed to create user ID: %v", err)
	}

	// Start a session and capture the cookie that the manager issues.
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

	// Send a second request with the session cookie from the login response.
	// This exercises the normal request path where LoadUser reads the cookie and
	// loads the stored session before the handler runs.
	loadRec := httptest.NewRecorder()
	manager.LoadUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the user ID that LoadUser attached to the request context.
		loadedUserID, err := manager.UserIDFromContext(r.Context())
		if err != nil {
			t.Fatalf("UserIDFromContext err=%v, want=%v", err, nil)
		}
		// Echo the loaded user ID so the test can verify that the manager
		// restored the same session it created during login.
		if _, err := w.Write([]byte("loaded user " + loadedUserID.String())); err != nil {
			t.Fatalf("Write err=%v, want=%v", err, nil)
		}
	})).ServeHTTP(loadRec, requestWithCookie(cookies[0]))
	// Confirm that the authenticated request completed successfully.
	if got, want := loadRec.Code, http.StatusOK; got != want {
		t.Fatalf("loadRec.Code=%d, want=%d", got, want)
	}
	// Confirm that the response contains the original user ID, which shows that
	// the manager loaded the stored session for this second request.
	if got, want := loadRec.Body.String(), "loaded user "+userID.String(); got != want {
		t.Errorf("loadRec.Body=%q, want=%q", got, want)
	}

	// Log out through the authenticated request and delete the session.
	endRec := httptest.NewRecorder()
	var endErr error
	manager.LoadUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		endErr = manager.LogOut(r.Context(), w)
	})).ServeHTTP(endRec, requestWithCookie(cookies[0]))
	if got, want := endErr, error(nil); got != want {
		t.Fatalf("endErr=%v, want=%v", got, want)
	}

	// Reuse the same cookie and confirm that logout removed the session.
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
