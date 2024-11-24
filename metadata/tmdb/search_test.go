package tmdb_test

import (
	"net/url"
	"testing"

	tmdbWrapper "github.com/ryanbradynd05/go-tmdb"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
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

func (m *mockTmdbAPI) GetTvInfo(id int, options map[string]string) (*tmdbWrapper.TV, error) {
	return nil, nil
}

func (m *mockTmdbAPI) GetTvExternalIds(id int, options map[string]string) (*tmdbWrapper.TvExternalIds, error) {
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
		{
			description: "handles empty results",
			query:       screenjournal.SearchQuery("nonexistent show"),
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
				}{},
			},
			want: []metadata.SearchResult{},
			err:  nil,
		},
		{
			description: "handles multiple results",
			query:       screenjournal.SearchQuery("popular shows"),
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
						Name:         "Show One",
						FirstAirDate: "2020-01-01",
						PosterPath:   "/show-one.jpg",
					},
					{
						ID:           67890,
						Name:         "Show Two",
						FirstAirDate: "2021-02-02",
						PosterPath:   "/show-two.jpg",
					},
				},
			},
			want: []metadata.SearchResult{
				{
					TmdbID:      screenjournal.TmdbID(12345),
					Title:       screenjournal.MediaTitle("Show One"),
					ReleaseDate: mustParseReleaseDate("2020-01-01"),
					PosterPath:  mustParseURL("/show-one.jpg"),
				},
				{
					TmdbID:      screenjournal.TmdbID(67890),
					Title:       screenjournal.MediaTitle("Show Two"),
					ReleaseDate: mustParseReleaseDate("2021-02-02"),
					PosterPath:  mustParseURL("/show-two.jpg"),
				},
			},
			err: nil,
		},
		{
			description: "ignores shows with no poster",
			query:       screenjournal.SearchQuery("no poster show"),
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
						Name:         "No Poster Show",
						FirstAirDate: "2022-03-03",
						PosterPath:   "",
					},
				},
			},
			want: []metadata.SearchResult{},
			err:  nil,
		},
		{
			description: "ignores shows with missing release date",
			query:       screenjournal.SearchQuery("breaking bad"),
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
						ID:           62428,
						Name:         "Breaking Bad",
						FirstAirDate: "",
						PosterPath:   "/breaking-bad-season1.jpg",
					},
				},
			},
			want: []metadata.SearchResult{},
			err:  nil,
		},
		{
			description: "ignores shows with invalid release date format",
			query:       screenjournal.SearchQuery("stranger things"),
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
						ID:           66732,
						Name:         "Stranger Things",
						FirstAirDate: "not-a-date",
						PosterPath:   "/stranger-things-s4.jpg",
					},
				},
			},
			want: []metadata.SearchResult{},
			err:  tmdb.ErrInvalidReleaseDate,
		},
		{
			description: "fails when a show has an invalid poster path",
			query:       screenjournal.SearchQuery("the office"),
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
						ID:           2316,
						Name:         "The Office",
						FirstAirDate: "2005-03-24",
						PosterPath:   ":\\invalid-office-poster",
					},
				},
			},
			want: []metadata.SearchResult{},
			err:  parse.ErrInvalidPosterPath,
		},
		{
			description: "fails when the result contains an empty title",
			query:       screenjournal.SearchQuery("friends"),
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
						ID:           1668,
						Name:         "",
						FirstAirDate: "1994-09-22",
						PosterPath:   "/friends-complete-series.jpg",
					},
				},
			},
			want: []metadata.SearchResult{},
			err:  parse.ErrInvalidMediaTitle,
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
					t.Errorf("result[%d].tmdbID=%v, want=%v", i, got, want)
				}
				if got, want := got[i].Title, tt.want[i].Title; got != want {
					t.Errorf("result[%d].title=%v, want=%v", i, got, want)
				}
				if got, want := got[i].ReleaseDate, tt.want[i].ReleaseDate; got != want {
					t.Errorf("result[%d].releaseDate=%v, want=%v", i, got, want)
				}
				if got, want := got[i].PosterPath.String(), tt.want[i].PosterPath.String(); got != want {
					t.Errorf("result[%d].posterPath=%v, want=%v", i, got, want)
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
