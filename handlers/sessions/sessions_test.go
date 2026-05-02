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
	store := test_sqlite.New()
	requireTls := false
	manager := sessions.NewManager(store, requireTls)
	// Create a handler that loads the session and writes the user ID in the
	// response.
	loadUserHandler := manager.LoadUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeBody := func(body string) {
			if _, err := w.Write([]byte(body)); err != nil {
				panic(err)
			}
		}

		loadedUserID, err := manager.UserIDFromContext(r.Context())
		if err != nil {
			writeBody("no session found")
			return
		}
		writeBody("loaded user " + loadedUserID.String())
	}))
	userID, err := simple_sessions.NewUserID("dummyuserID")
	if err != nil {
		t.Fatalf("failed to create user ID: %v", err)
	}

	var sessionCookie *http.Cookie
	{
		// Start a session and capture the cookie that the manager issues.
		rec := httptest.NewRecorder()
		if err := manager.LogIn(context.Background(), rec, userID); err != nil {
			t.Fatalf("LogIn err=%v, want=%v", err, nil)
		}

		cookies := rec.Result().Cookies()
		if got, want := len(cookies), 1; got != want {
			t.Fatalf("len(cookies)=%d, want=%d", got, want)
		}
		sessionCookie = cookies[0]
	}

	{
		// Send a second request with the session cookie from the login response.
		// This exercises the normal request path where LoadUser reads the cookie
		// and loads the stored session before the handler runs.
		rec := httptest.NewRecorder()
		loadUserHandler.ServeHTTP(rec, requestWithCookie(sessionCookie))
		// Confirm that the authenticated request completed successfully.
		if got, want := rec.Code, http.StatusOK; got != want {
			t.Fatalf("rec.Code=%d, want=%d", got, want)
		}
		// Confirm that the response contains the original user ID, which shows that
		// the manager loaded the stored session for this second request.
		if got, want := rec.Body.String(), "loaded user "+userID.String(); got != want {
			t.Errorf("rec.Body=%q, want=%q", got, want)
		}
	}

	{
		// Log out through the authenticated request and delete the session.
		rec := httptest.NewRecorder()
		var err error
		manager.LoadUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err = manager.LogOut(r.Context(), w)
		})).ServeHTTP(rec, requestWithCookie(sessionCookie))
		if got, want := err, error(nil); got != want {
			t.Fatalf("err=%v, want=%v", got, want)
		}
	}

	{
		// Reuse the same cookie and confirm that logout removed the session.
		rec := httptest.NewRecorder()
		loadUserHandler.ServeHTTP(rec, requestWithCookie(sessionCookie))
		if got, want := rec.Body.String(), "no session found"; got != want {
			t.Fatalf("rec.Body=%q, want=%q", got, want)
		}
	}
}

func requestWithCookie(cookie *http.Cookie) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)
	return req
}
