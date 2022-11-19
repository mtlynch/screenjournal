package metadata

import (
	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/ryanbradynd05/go-tmdb"
)

type (
	Finder interface {
		Search(query string) (MovieSearchResults, error)
		GetMovieInfo(id screenjournal.TmdbID) (screenjournal.Movie, error)
	}

	MovieSearchResult struct {
		TmdbID      screenjournal.TmdbID
		Title       string
		ReleaseDate string
		PosterPath  string
	}

	MovieSearchResults struct {
		Matches      []MovieSearchResult
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

func (f tmdbFinder) GetMovieInfo(id screenjournal.TmdbID) (screenjournal.Movie, error) {
	m, err := f.tmdbAPI.GetMovieInfo(int(id.Int32()), map[string]string{})
	if err != nil {
		return screenjournal.Movie{}, err
	}

	title, err := parse.MediaTitle(m.Title)
	if err != nil {
		return screenjournal.Movie{}, err
	}

	// TODO: Return a TMDB-specific type.
	return screenjournal.Movie{
		TmdbID:     id,
		ImdbID:     screenjournal.ImdbID(m.ImdbID), // TODO: Actually parse this
		Title:      title,
		PosterPath: m.PosterPath, // TODO: Actually parse this
	}, nil
}
