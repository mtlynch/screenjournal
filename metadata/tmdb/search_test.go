package tmdb_test

import (
	"net/url"
	"testing"

	tmdbWrapper "github.com/ryanbradynd05/go-tmdb"

	"github.com/mtlynch/screenjournal/v2/metadata"
	"github.com/mtlynch/screenjournal/v2/metadata/tmdb"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type mockTmdbAPI struct {
	searchTvResponse *tmdbWrapper.TvSearchResults
}

func (m *mockTmdbAPI) GetMovieInfo(id int, options map[string]string) (*tmdbWrapper.Movie, error) {
	return nil, nil
}

func (m *mockTmdbAPI) SearchMovie(query string, options map[string]string) (*tmdbWrapper.MovieSearchResults, error) {
	return nil, nil
}

func (m *mockTmdbAPI) SearchTv(query string, options map[string]string) (*tmdbWrapper.TvSearchResults, error) {
	return m.searchTvResponse, nil
}

func TestSearchTvShows(t *testing.T) {
	for _, tt := range []struct {
		description string
		query       screenjournal.SearchQuery
		mockResults tmdbWrapper.TvSearchResults
		want        []metadata.SearchResult
		err         error
	}{
		{
			description: "processes valid TV show result",
			query:       screenjournal.SearchQuery("valid show"),
			mockResults: tmdbWrapper.TvSearchResults{
				Results: []struct {
					BackdropPath  string `json:"backdrop_path"`
					ID            int
					OriginalName  string   `json:"original_name"`
					FirstAirDate  string   `json:"first_air_date"`
					OriginCountry []string `json:"origin_country"`
					PosterPath    string   `json:"poster_path"`
					Popularity    float32
					Name          string
					VoteAverage   float32 `json:"vote_average"`
					VoteCount     uint32  `json:"vote_count"`
				}{
					{
						ID:           12345,
						Name:         "Black Mirror",
						FirstAirDate: "2009-01-05",
						PosterPath:   "/black-mirror.jpg",
					},
				},
			},
			want: []metadata.SearchResult{
				{
					TmdbID:      screenjournal.TmdbID(12345),
					Title:       screenjournal.MediaTitle("Black Mirror"),
					ReleaseDate: mustParseReleaseDate("2009-01-05"),
					PosterPath:  mustParseURL("/black-mirror.jpg"),
				},
			},
			err: nil,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			mockAPI := &mockTmdbAPI{
				searchTvResponse: &tt.mockResults,
			}
			finder := tmdb.NewWithAPI(mockAPI)

			got, err := finder.SearchTvShows(tt.query)
			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}

			if got, want := len(got), len(tt.want); got != want {
				t.Fatalf("result count=%d, want=%d", got, want)
			}

			for i := range got {
				if got, want := got[i].TmdbID, tt.want[i].TmdbID; got != want {
					t.Errorf("tmdbID=%v, want=%v", got, want)
				}
				if got, want := got[i].Title, tt.want[i].Title; got != want {
					t.Errorf("title=%v, want=%v", got, want)
				}
				if got, want := got[i].ReleaseDate, tt.want[i].ReleaseDate; got != want {
					t.Errorf("releaseDate=%v, want=%v", got, want)
				}
				if got, want := got[i].PosterPath.String(), tt.want[i].PosterPath.String(); got != want {
					t.Errorf("posterPath=%v, want=%v", got, want)
				}
			}
		})
	}
}

func mustParseURL(s string) url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return *u
}

func mustParseReleaseDate(s string) screenjournal.ReleaseDate {
	rd, err := tmdb.ParseReleaseDate(s)
	if err != nil {
		panic(err)
	}
	return rd
}
