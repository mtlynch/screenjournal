package tmdb

import (
	"github.com/ryanbradynd05/go-tmdb"
)

type Finder struct {
	tmdbAPI *tmdb.TMDb
}

func New(apiKey string) (Finder, error) {
	tmbdAPI := tmdb.Init(tmdb.Config{
		APIKey: apiKey,
	})
	return Finder{
		tmdbAPI: tmbdAPI,
	}, nil
}
