package metadata

import (
	"log"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/ryanbradynd05/go-tmdb"
)

type (
	Finder interface {
		Search(query string) (MovieSearchResults, error)
		GetMovieInfo(id screenjournal.TmdbID) (MovieInfo, error)
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

	MovieInfo struct {
		TmdbID     screenjournal.TmdbID
		ImdbID     screenjournal.ImdbID
		Title      screenjournal.MediaTitle
		PosterPath string
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

func (f tmdbFinder) GetMovieInfo(id screenjournal.TmdbID) (MovieInfo, error) {
	m, err := f.tmdbAPI.GetMovieInfo(int(id.Int32()), map[string]string{})
	if err != nil {
		return MovieInfo{}, err
	}

	info := MovieInfo{
		TmdbID: id,
	}

	info.Title, err = parse.MediaTitle(m.Title)
	if err != nil {
		return MovieInfo{}, err
	}

	imdbID, err := parse.ImdbID(m.ImdbID)
	if err != nil {
		log.Printf("failed to parse IMDB ID (%s) from TMDB ID (%v): %v", m.ImdbID, id, err)
	} else {
		info.ImdbID = imdbID
	}

	// TODO: Actually parse this
	info.PosterPath = m.PosterPath

	return info, nil
}
