package metadata

import "github.com/ryanbradynd05/go-tmdb"

type (
	Finder interface {
		Search(query string) (MovieSearchResults, error)
	}

	MovieStub struct {
		ID          int
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
