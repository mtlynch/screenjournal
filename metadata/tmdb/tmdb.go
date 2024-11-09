package tmdb

import (
	tmdbWrapper "github.com/ryanbradynd05/go-tmdb"
)

type tmdbAPI interface {
	GetMovieInfo(int, map[string]string) (*tmdbWrapper.Movie, error)
	GetTvInfo(int, map[string]string) (*tmdbWrapper.TV, error)
	SearchMovie(query string, options map[string]string) (*tmdbWrapper.MovieSearchResults, error)
	SearchTv(query string, options map[string]string) (*tmdbWrapper.TvSearchResults, error)
}

type Finder struct {
	tmdbAPI tmdbAPI
}

func New(apiKey string) (Finder, error) {
	tmbdAPI := tmdbWrapper.Init(tmdbWrapper.Config{
		APIKey: apiKey,
	})
	return NewWithAPI(tmbdAPI), nil
}

func NewWithAPI(api tmdbAPI) Finder {
	return Finder{
		tmdbAPI: api,
	}
}
