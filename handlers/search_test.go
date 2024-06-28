package handlers_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kylelemons/godebug/diff"

	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
	"github.com/mtlynch/screenjournal/v2/metadata"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

func TestSearchGet(t *testing.T) {
	for _, tt := range []struct {
		description  string
		url          string
		sessionToken string
		sessions     []mockSessionEntry
		status       int
		response     string
	}{
		{
			description:  "returns empty result on no matches",
			url:          "/api/search?query=matchesnothing555",
			sessionToken: "abc123",
			sessions: []mockSessionEntry{
				{
					token: "abc123",
					session: sessions.Session{
						Username: screenjournal.Username("user123"),
					},
				},
			},
			status: http.StatusOK,
			response: `
<ul class="dropdown-menu show" aria-labelledby="search-box">
  <li>No matches</li>
</ul>
`,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			store := test_sqlite.New()

			// Populate datastore with dummy users.
			for _, s := range tt.sessions {
				store.InsertUser(screenjournal.User{
					Username: s.session.Username,
				})
			}

			authenticator := auth.New(store)
			sessionManager := newMockSessionManager(tt.sessions)
			metadataFinder := NewMockMetadataFinder([]metadata.MovieInfo{})
			s := handlers.New(authenticator, nilAnnouncer, &sessionManager, store, metadataFinder)

			req, err := http.NewRequest("GET", tt.url, nil)
			if err != nil {
				t.Fatal(err)
			}
			req.AddCookie(&http.Cookie{
				Name:  mockSessionTokenName,
				Value: tt.sessionToken,
			})

			w := httptest.NewRecorder()
			s.Router().ServeHTTP(w, req)

			if got, want := w.Code, tt.status; got != want {
				t.Fatalf("httpStatus=%v, want=%v", got, want)
			}

			if tt.status != http.StatusOK {
				return
			}

			response, err := io.ReadAll(w.Body)
			if err != nil {
				t.Fatalf("failed to read server response: %v", err)
			}

			got := string(response)
			want := strings.TrimPrefix(tt.response, "\n")
			if delta := diff.Diff(got, want); delta != "" {
				t.Errorf("diff in search response\nGot:\n%s\nWant:\n%s\nDiff:%s", got, want, delta)
			}
		})
	}
}
