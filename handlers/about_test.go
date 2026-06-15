package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

func TestAboutPageIncludesVersionMetadata(t *testing.T) {
	sessionManager := newMockSessionManager(nil)
	server := handlers.New(handlers.ServerParams{
		SessionManager: &sessionManager,
		Store:          test_sqlite.New(),
	})
	req := httptest.NewRequest(http.MethodGet, "/about", nil)
	res := httptest.NewRecorder()

	server.Router().ServeHTTP(res, req)

	if got, want := res.Code, http.StatusOK; got != want {
		t.Fatalf("status=%d, want=%d", got, want)
	}

	body := res.Body.String()
	if !regexp.MustCompile(`<h1[^>]*>\s*About ScreenJournal\s*</h1>`).MatchString(body) {
		t.Fatalf("about heading not found")
	}

	versionTimeMatch := regexp.MustCompile(
		`id="version-time-local" datetime="([^"]+)"`,
	).FindStringSubmatch(body)
	if len(versionTimeMatch) == 0 {
		return
	}

	versionTime := versionTimeMatch[1]
	if _, err := time.Parse(time.RFC3339, versionTime); err != nil {
		t.Fatalf("version time %q not RFC3339: %v", versionTime, err)
	}

	if !regexp.MustCompile(`Version last modified at`).MatchString(body) {
		t.Errorf("version timestamp label not found")
	}
}
