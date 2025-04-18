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
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

var mockSearchableMovies = []screenjournal.Movie{
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

var mockSearchableTvShows = []screenjournal.TvShow{
	{
		TmdbID:  screenjournal.TmdbID(3),
		Title:   screenjournal.MediaTitle("Party Down"),
		AirDate: mustParseReleaseDate("2009-01-01"),
		PosterPath: url.URL{
			Path: "/party-down.jpg",
		},
	},
	{
		TmdbID:  screenjournal.TmdbID(4),
		Title:   screenjournal.MediaTitle("Party Down South"),
		AirDate: mustParseReleaseDate("2014-05-09"),
		PosterPath: url.URL{
			Path: "/party-down-south.jpg",
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
			description:  "returns valid results for valid movie query",
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
			<a
				href="/reviews/new/write?tmdbId=1&mediaType=movie"
				><img src="https://image.tmdb.org/t/p/w92/the-waterboy.jpg" /><span class="mx-3"
					>The Waterboy (1998)</span
				></a
			>
		</li>
		<li>
			<a
				href="/reviews/new/write?tmdbId=2&mediaType=movie"
				><img src="https://image.tmdb.org/t/p/w92/waterboys.jpg" /><span class="mx-3"
					>Waterboys (2011)</span
				></a
			>
		</li>
</ul>
`,
		},
		{
			description:  "returns valid results for valid TV query",
			url:          "/api/search?mediaType=tv-show&query=party%20down",
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
			<a
				href="/reviews/new/tv/pick-season?tmdbId=3"
				><img src="https://image.tmdb.org/t/p/w92/party-down.jpg" /><span class="mx-3"
					>Party Down (2009)</span
				></a
			>
		</li>
		<li>
			<a
				href="/reviews/new/tv/pick-season?tmdbId=4"
				><img src="https://image.tmdb.org/t/p/w92/party-down-south.jpg" /><span class="mx-3"
					>Party Down South (2014)</span
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
			dataStore := test_sqlite.New()

			// Populate datastore with dummy users.
			for _, s := range tt.sessions {
				dataStore.InsertUser(screenjournal.User{
					Username: s.session.Username,
				})
			}

			authenticator := auth.New(dataStore)
			sessionManager := newMockSessionManager(tt.sessions)
			metadataFinder := NewMockMetadataFinder(mockSearchableMovies, mockSearchableTvShows)
			s := handlers.New(authenticator, nilAnnouncer, &sessionManager, dataStore, metadataFinder)

			req, err := http.NewRequest("GET", tt.url, nil)
			if err != nil {
				t.Fatal(err)
			}
			req.AddCookie(&http.Cookie{
				Name:  mockSessionTokenName,
				Value: tt.sessionToken,
			})

			rec := httptest.NewRecorder()
			s.Router().ServeHTTP(rec, req)
			res := rec.Result()

			if got, want := res.StatusCode, tt.status; got != want {
				t.Fatalf("httpStatus=%v, want=%v", got, want)
			}

			if tt.status != http.StatusOK {
				return
			}

			response, err := io.ReadAll(res.Body)
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
