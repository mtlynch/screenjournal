package tmdb

import (
	"github.com/mtlynch/screenjournal/v2/metadata"
	"github.com/ryanbradynd05/go-tmdb"
)

type tmdbFinder struct {
	tmdbAPI *tmdb.TMDb
}

func New(apiKey string) (metadata.Finder, error) {
	tmbdAPI := tmdb.Init(tmdb.Config{
		APIKey: apiKey,
	})
	return tmdbFinder{
		tmdbAPI: tmbdAPI,
	}, nil
}
