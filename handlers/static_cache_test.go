package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

func TestStaticFilesUseCacheControlAndETag(t *testing.T) {
	server := newStaticFileTestServer()
	responseRecorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/css/screenjournal.css",
		nil,
	)

	server.Router().ServeHTTP(responseRecorder, request)

	if got, want := responseRecorder.Code, http.StatusOK; got != want {
		t.Fatalf("status=%d, want=%d", got, want)
	}
	if got, want := responseRecorder.Header().Get("Cache-Control"), "public, max-age=1800"; got != want {
		t.Errorf("Cache-Control=%q, want=%q", got, want)
	}
	if got := responseRecorder.Header().Get("ETag"); got == "" {
		t.Errorf("ETag header is empty")
	}
}

func TestStaticFilesReturnNotModifiedForMatchingETag(t *testing.T) {
	server := newStaticFileTestServer()

	firstResponseRecorder := httptest.NewRecorder()
	firstRequest := httptest.NewRequest(
		http.MethodGet,
		"/css/screenjournal.css",
		nil,
	)
	server.Router().ServeHTTP(firstResponseRecorder, firstRequest)

	if got, want := firstResponseRecorder.Code, http.StatusOK; got != want {
		t.Fatalf("initial status=%d, want=%d", got, want)
	}

	cachedETag := firstResponseRecorder.Header().Get("ETag")
	if cachedETag == "" {
		t.Fatalf("initial ETag header is empty")
	}

	secondResponseRecorder := httptest.NewRecorder()
	secondRequest := httptest.NewRequest(
		http.MethodGet,
		"/css/screenjournal.css",
		nil,
	)
	secondRequest.Header.Set("If-None-Match", cachedETag)

	server.Router().ServeHTTP(secondResponseRecorder, secondRequest)

	if got, want := secondResponseRecorder.Code, http.StatusNotModified; got != want {
		t.Fatalf("status=%d, want=%d", got, want)
	}
	if got, want := secondResponseRecorder.Header().Get("Cache-Control"), "public, max-age=1800"; got != want {
		t.Errorf("Cache-Control=%q, want=%q", got, want)
	}
	if got, want := secondResponseRecorder.Header().Get("ETag"), cachedETag; got != want {
		t.Errorf("ETag=%q, want=%q", got, want)
	}
	if got, want := secondResponseRecorder.Body.Len(), 0; got != want {
		t.Errorf("body length=%d, want=%d", got, want)
	}
}

func TestStaticFilesServeFontawesomeWebfontRanges(t *testing.T) {
	server := newStaticFileTestServer()
	responseRecorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/third-party/fontawesome@6.2.0/webfonts/fa-solid-900.ttf",
		nil,
	)
	request.Header.Set("Range", "bytes=0-1")

	server.Router().ServeHTTP(responseRecorder, request)

	if got, want := responseRecorder.Code, http.StatusPartialContent; got != want {
		t.Fatalf(
			"status=%d, want=%d, body=%q",
			got,
			want,
			responseRecorder.Body.String(),
		)
	}
	if got, want := responseRecorder.Body.Len(), 2; got != want {
		t.Errorf("body length=%d, want=%d", got, want)
	}
}

func newStaticFileTestServer() handlers.Server {
	sessionManager := newMockSessionManager(nil)
	return handlers.New(handlers.ServerParams{
		SessionManager: &sessionManager,
		Store:          test_sqlite.New(),
	})
}
