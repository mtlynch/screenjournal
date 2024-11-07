package tmdb_test

import (
	"log"
	"net/url"
	"testing"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/metadata"
	"github.com/mtlynch/screenjournal/v2/metadata/tmdb"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type mockTmdbAPI struct {
	searchTvResponse tmdbSearchTvResponse
}

func (m *mockTmdbAPI) SearchTv(query string, options map[string]string) (tmdbSearchTvResponse, error) {
	return m.searchTvResponse, nil
}

type tmdbSearchTvResponse struct {
	Results []tmdbTvResult `json:"results"`
}

type tmdbTvResult struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	FirstAirDate string `json:"first_air_date"`
	PosterPath   string `json:"poster_path"`
}

func mustParseReleaseDate(s string) screenjournal.ReleaseDate {
	d, err := tmdb.ParseReleaseDate(s)
	if err != nil {
		log.Fatalf("failed to parse release date: %s", s)
	}
	return d
}

func TestSearchTvShows(t *testing.T) {
	for _, tt := range []struct {
		description string
		query       screenjournal.SearchQuery
		mockResults tmdbSearchTvResponse
		want        []metadata.SearchResult
		err         error
	}{
		{
			description: "handles empty results",
			query:       screenjournal.SearchQuery("nonexistent show"),
			mockResults: tmdbSearchTvResponse{
				Results: []tmdbTvResult{},
			},
			want: []metadata.SearchResult{},
			err:  nil,
		},
		{
			description: "skips results with missing poster path",
			query:       screenjournal.SearchQuery("test show"),
			mockResults: tmdbSearchTvResponse{
				Results: []tmdbTvResult{
					{
						ID:           12345,
						Name:         "Test Show",
						FirstAirDate: "2023-01-15",
						PosterPath:   "",
					},
				},
			},
			want: []metadata.SearchResult{},
			err:  nil,
		},
		{
			description: "skips results with missing air date",
			query:       screenjournal.SearchQuery("test show"),
			mockResults: tmdbSearchTvResponse{
				Results: []tmdbTvResult{
					{
						ID:           12345,
						Name:         "Test Show",
						FirstAirDate: "",
						PosterPath:   "/poster.jpg",
					},
				},
			},
			want: []metadata.SearchResult{},
			err:  nil,
		},
		{
			description: "processes valid TV show result",
			query:       screenjournal.SearchQuery("valid show"),
			mockResults: tmdbSearchTvResponse{
				Results: []tmdbTvResult{
					{
						ID:           54321,
						Name:         "Valid Show",
						FirstAirDate: "2023-06-20",
						PosterPath:   "/valid-show.jpg",
					},
				},
			},
			want: []metadata.SearchResult{
				{
					TmdbID:      screenjournal.TmdbID(54321),
					Title:       screenjournal.MediaTitle("Valid Show"),
					ReleaseDate: mustParseReleaseDate("2023-06-20"),
					PosterPath: url.URL{
						Path: "/valid-show.jpg",
					},
				},
			},
			err: nil,
		},
		{
			description: "handles invalid TMDB ID",
			query:       screenjournal.SearchQuery("bad show"),
			mockResults: tmdbSearchTvResponse{
				Results: []tmdbTvResult{
					{
						ID:           -1,
						Name:         "Bad Show",
						FirstAirDate: "2023-06-20",
						PosterPath:   "/bad-show.jpg",
					},
				},
			},
			want: []metadata.SearchResult{},
			err:  parse.ErrInvalidTmdbID,
		},
		{
			description: "handles invalid release date",
			query:       screenjournal.SearchQuery("bad date show"),
			mockResults: tmdbSearchTvResponse{
				Results: []tmdbTvResult{
					{
						ID:           12345,
						Name:         "Bad Date Show",
						FirstAirDate: "invalid-date",
						PosterPath:   "/bad-date-show.jpg",
					},
				},
			},
			want: []metadata.SearchResult{},
			err:  tmdb.ErrInvalidReleaseDate,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			mockAPI := &mockTmdbAPI{
				searchTvResponse: tt.mockResults,
			}
			f := tmdb.NewWithAPI(mockAPI)

			got, err := f.SearchTvShows(tt.query)
			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if err != nil {
				return
			}

			if got, want := len(got), len(tt.want); got != want {
				t.Fatalf("result count=%d, want=%d", got, want)
			}

			for i := range got {
				if got, want := got[i].TmdbID, tt.want[i].TmdbID; got != want {
					t.Errorf("TmdbID=%v, want=%v", got, want)
				}
				if got, want := got[i].Title, tt.want[i].Title; got != want {
					t.Errorf("Title=%v, want=%v", got, want)
				}
				if got, want := got[i].ReleaseDate, tt.want[i].ReleaseDate; got != want {
					t.Errorf("ReleaseDate=%v, want=%v", got, want)
				}
				if got, want := got[i].PosterPath.String(), tt.want[i].PosterPath.String(); got != want {
					t.Errorf("PosterPath=%v, want=%v", got, want)
				}
			}
		})
	}
}
