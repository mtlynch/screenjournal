package metadata

import (
	"github.com/mtlynch/screenjournal/v2"
	"github.com/ryanbradynd05/go-tmdb"
)

type (
	Finder interface {
		Search(query string) (MovieSearchResults, error)
		GetMovieInfo(id screenjournal.TmdbID) (Movie, error)
	}

	MovieStub struct {
		ID          screenjournal.TmdbID
		Title       string
		ReleaseDate string
		PosterURL   string
	}

	MovieSearchResults struct {
		Matches      []MovieStub
		Page         int
		TotalPages   int
		TotalResults int
	}

	Movie struct {
		Title     string
		PosterURL string
	}

	tmdbFinder struct {
		tmdbAPI *tmdb.TMDb
	}
)

func New(apiKey string) (Finder, error) {
	tmbdAPI := tmdb.Init(tmdb.Config{
		APIKey: apiKey,
	})
	return tmdbFinder{
		tmdbAPI: tmbdAPI,
	}, nil
}

func (f tmdbFinder) GetMovieInfo(id screenjournal.TmdbID) (Movie, error) {
	m, err := f.tmdbAPI.GetMovieInfo(id.Int(), map[string]string{})
	if err != nil {
		return Movie{}, err
	}

	return Movie{
		Title: m.Title,
	}, nil
}
