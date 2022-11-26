package metadata

import (
	"github.com/mtlynch/screenjournal/v2"
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
		TmdbID screenjournal.TmdbID
		ImdbID screenjournal.ImdbID
		Title  screenjournal.MediaTitle
	}
)

func MovieFromMovieInfo(mi MovieInfo) screenjournal.Movie {
	return screenjournal.Movie{
		TmdbID: mi.TmdbID,
		ImdbID: mi.ImdbID,
		Title:  mi.Title,
	}
}
