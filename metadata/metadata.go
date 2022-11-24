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
		TmdbID       screenjournal.TmdbID
		ImdbID       screenjournal.ImdbID
		Title        screenjournal.MediaTitle
		ReleaseDate  screenjournal.ReleaseDate
		PosterPath   screenjournal.ImagePath
		BackdropPath screenjournal.ImagePath
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

	info.ReleaseDate, err = parse.ReleaseDate(m.ReleaseDate)
	if err != nil {
		return MovieInfo{}, err
	}

	if len(m.ImdbID) > 0 {
		imdbID, err := parse.ImdbID(m.ImdbID)
		if err != nil {
			log.Printf("failed to parse IMDB ID (%s) from TMDB ID %v: %v", m.ImdbID, id, err)
		} else {
			info.ImdbID = imdbID
		}
	}

	if len(m.PosterPath) > 0 {
		posterPath, err := parse.ImagePath(m.PosterPath)
		if err != nil {
			log.Printf("failed to parse poster path (%s) from TMDB ID %v: %v", m.PosterPath, id, err)
		} else {
			info.PosterPath = posterPath
		}
	}

	if len(m.BackdropPath) > 0 {
		backdropPath, err := parse.ImagePath(m.BackdropPath)
		if err != nil {
			log.Printf("failed to parse backdrop path (%s) from TMDB ID %v: %v", m.BackdropPath, id, err)
		} else {
			info.BackdropPath = backdropPath
		}
	}

	return info, nil
}
