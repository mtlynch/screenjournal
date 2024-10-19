package handlers_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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

var mockSearchableData = []metadata.MovieInfo{
	{
		TmdbID:      screenjournal.TmdbID(1),
		Title:       screenjournal.MediaTitle("The Waterboy"),
		ReleaseDate: mustParseReleaseDate("1998-11-06"),
		PosterPath: url.URL{
			Path: "/the-waterboy.jpg",
		},
	},
	{
		TmdbID:      screenjournal.TmdbID(2),
		Title:       screenjournal.MediaTitle("Waterboys"),
		ReleaseDate: mustParseReleaseDate("2011-12-05"),
		PosterPath: url.URL{
			Path: "/waterboys.jpg",
		},
	},
}

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
			description:  "matches search results",
			url:          "/api/search?mediaType=movie&query=waterbo",
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
<ul class="py-0 my-0 list-unstyled border border-success">
	<li>
			<a href="/reviews/new/write?tmdbId=1"
				><img src="https://image.tmdb.org/t/p/w92/the-waterboy.jpg" /><span class="mx-3"
					>The Waterboy (1998)</span
				></a
			>
		</li>
	<li>
			<a href="/reviews/new/write?tmdbId=2"
				><img src="https://image.tmdb.org/t/p/w92/waterboys.jpg" /><span class="mx-3"
					>Waterboys (2011)</span
				></a
			>
		</li>
</ul>
`,
		},
		{
			description:  "returns empty result on no matches",
			url:          "/api/search?mediaType=movie&query=matchesnothing555",
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
<ul class="py-0 my-0 list-unstyled border border-success">
  <li>No matches</li>
</ul>
`,
		},
		{
			description:  "returns empty string when query is too short",
			url:          "/api/search?mediaType=movie&query=a",
			sessionToken: "abc123",
			sessions: []mockSessionEntry{
				{
					token: "abc123",
					session: sessions.Session{
						Username: screenjournal.Username("user123"),
					},
				},
			},
			status:   http.StatusUnprocessableEntity,
			response: "",
		},
		{
			description:  "prevents an unauthenticated user from searching",
			url:          "/api/search?mediaType=movie&query=waterbo",
			sessionToken: "",
			sessions: []mockSessionEntry{
				{
					token: "abc123",
					session: sessions.Session{
						Username: screenjournal.Username("user123"),
					},
				},
			},
			status:   http.StatusUnauthorized,
			response: "You must log in to perform searches",
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
			metadataFinder := NewMockMetadataFinder(mockSearchableData)
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

			got := removeBlankLines(string(response))
			want := formatExpectedResponse(tt.response)
			if delta := diff.Diff(got, want); delta != "" {
				t.Errorf("diff in search response\nGot:\n%s\nWant:\n%s\nDiff:%s", got, want, delta)
			}
		})
	}
}

func formatExpectedResponse(raw string) string {
	response := strings.TrimPrefix(raw, "\n")
	response = strings.ReplaceAll(response, "\t", "  ")
	response = removeBlankLines(response)
	return response
}

func removeBlankLines(input string) string {
	lines := strings.Split(input, "\n")
	var result []string

	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}
