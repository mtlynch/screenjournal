package tmdb_test

import (
	"testing"

	tmdbWrapper "github.com/ryanbradynd05/go-tmdb"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/metadata"
	"github.com/mtlynch/screenjournal/v2/metadata/tmdb"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func TestGetTvShowInfo(t *testing.T) {
	for _, tt := range []struct {
		description string
		id          screenjournal.TmdbID
		mockResult  tmdbWrapper.TV
		want        metadata.TvShowInfo
		err         error
	}{
		{
			description: "processes valid TV show info",
			id:          screenjournal.TmdbID(12345),
			mockResult: tmdbWrapper.TV{
				Name:         "Seinfeld",
				FirstAirDate: "1989-07-05",
				PosterPath:   "/seinfeld-poster.jpg",
				ExternalIDs: &tmdbWrapper.TvExternalIds{
					ImdbID: "tt0098904",
				},
			},
			want: metadata.TvShowInfo{
				TmdbID:      screenjournal.TmdbID(12345),
				Title:       screenjournal.MediaTitle("Seinfeld"),
				ReleaseDate: mustParseReleaseDate("1989-07-05"),
				PosterPath:  mustParseURL("/seinfeld-poster.jpg"),
				ImdbID:      mustParseImdbID("tt0098904"),
			},
			err: nil,
		},
		{
			description: "handles show with missing IMDB ID",
			id:          screenjournal.TmdbID(67890),
			mockResult: tmdbWrapper.TV{
				Name:         "The Mandalorian",
				FirstAirDate: "2019-11-12",
				PosterPath:   "/mandalorian-poster.jpg",
			},
			want: metadata.TvShowInfo{
				TmdbID:      screenjournal.TmdbID(67890),
				Title:       screenjournal.MediaTitle("The Mandalorian"),
				ReleaseDate: mustParseReleaseDate("2019-11-12"),
				PosterPath:  mustParseURL("/mandalorian-poster.jpg"),
			},
			err: nil,
		},
		{
			description: "fails when show has empty title",
			id:          screenjournal.TmdbID(11111),
			mockResult: tmdbWrapper.TV{
				Name:         "",
				FirstAirDate: "2020-01-01",
				PosterPath:   "/show-poster.jpg",
			},
			want: metadata.TvShowInfo{},
			err:  parse.ErrInvalidMediaTitle,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			mockAPI := &mockTmdbAPI{
				getTvInfoResponse: &tt.mockResult,
			}
			finder := tmdb.NewWithAPI(mockAPI)

			got, err := finder.GetTvShowInfo(tt.id)
			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}

			if err != nil {
				return
			}

			if got, want := got.TmdbID, tt.want.TmdbID; got != want {
				t.Errorf("tmdbID=%v, want=%v", got, want)
			}
			if got, want := got.Title, tt.want.Title; got != want {
				t.Errorf("title=%v, want=%v", got, want)
			}
			if got, want := got.ReleaseDate, tt.want.ReleaseDate; got != want {
				t.Errorf("releaseDate=%v, want=%v", got, want)
			}
			if got, want := got.PosterPath.String(), tt.want.PosterPath.String(); got != want {
				t.Errorf("posterPath=%v, want=%v", got, want)
			}
			if got, want := got.ImdbID, tt.want.ImdbID; got != want {
				t.Errorf("imdbID=%v, want=%v", got, want)
			}
		})
	}
}

func mustParseImdbID(s string) screenjournal.ImdbID {
	id, err := tmdb.ParseImdbID(s)
	if err != nil {
		panic(err)
	}
	return id
}
